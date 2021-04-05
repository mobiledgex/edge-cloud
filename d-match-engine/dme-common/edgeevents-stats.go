package dmecommon

import (
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"golang.org/x/net/context"
)

var EEStats *EdgeEventStats

type EdgeEventStatCall struct {
	Metric          string           // Either cloudcommon.LatencyMetric, cloudcommon.DeviceMetric, cloudcommon.CustomMetric
	LatencyStatKey  LatencyStatKey   // Key needed if metric is cloudcommon.LatencyMetric
	LatencyStatInfo *LatencyStatInfo // Latency stat info if metric is cloudcommon.LatencyMetric
	DeviceStatKey   DeviceStatKey    // Key needed if metric is cloudcommon.DeviceStatKey
	CustomStatKey   CustomStatKey    // Key needed if metric is cloudcommon.CustomMetric
	CustomStatInfo  *CustomStatInfo  // Custom stat info if metric is cloudcommon.CustomMetric
}

type EdgeEventMapShard struct {
	latencyStatMap map[LatencyStatKey]*LatencyStat
	customStatMap  map[CustomStatKey]*CustomStat
	deviceStatMap  map[DeviceStatKey]*DeviceStat
	notify         bool
	mux            sync.Mutex
}

type EdgeEventStats struct {
	shards    []EdgeEventMapShard
	numShards uint
	mux       sync.Mutex
	interval  time.Duration
	send      func(ctx context.Context, metric *edgeproto.Metric) bool
	waitGroup sync.WaitGroup
	stop      chan struct{}
}

func NewEdgeEventStats(interval time.Duration, numShards uint, send func(ctx context.Context, metric *edgeproto.Metric) bool) *EdgeEventStats {
	e := EdgeEventStats{}
	e.shards = make([]EdgeEventMapShard, numShards, numShards)
	e.numShards = numShards
	for ii, _ := range e.shards {
		e.shards[ii].latencyStatMap = make(map[LatencyStatKey]*LatencyStat)
		e.shards[ii].deviceStatMap = make(map[DeviceStatKey]*DeviceStat)
		e.shards[ii].customStatMap = make(map[CustomStatKey]*CustomStat)
	}
	e.interval = interval
	e.send = send
	return &e
}

func (e *EdgeEventStats) Start() {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.stop != nil {
		return
	}
	e.stop = make(chan struct{})
	e.waitGroup.Add(1)
	go e.RunNotify()
}

func (e *EdgeEventStats) Stop() {
	e.mux.Lock()
	close(e.stop)
	e.mux.Unlock()
	e.waitGroup.Wait()
	e.mux.Lock()
	e.stop = nil
	e.mux.Unlock()
}

func (e *EdgeEventStats) UpdateSettings(newinterval time.Duration) {
	if e.interval == newinterval {
		return
	}

	restart := false
	if e.stop != nil {
		e.Stop()
		restart = true
	}
	e.mux.Lock()
	e.interval = newinterval
	e.mux.Unlock()
	if restart {
		e.Start()
	}
}

func (e *EdgeEventStats) RecordEdgeEventStatCall(call *EdgeEventStatCall) {
	if call.Metric == cloudcommon.LatencyMetric {
		key := call.LatencyStatKey
		emptyStatKey := LatencyStatKey{}
		if key == emptyStatKey {
			return
		}
		idx := util.GetShardIndex(key, e.numShards)

		shard := &e.shards[idx]
		shard.mux.Lock()
		defer shard.mux.Unlock()
		stat, found := shard.latencyStatMap[key]
		if !found {
			stat = NewLatencyStat(LatencyTimes)
		}
		stat.Update(call.LatencyStatInfo)
		shard.latencyStatMap[key] = stat
	} else if call.Metric == cloudcommon.DeviceMetric {
		key := call.DeviceStatKey
		emptyStatKey := DeviceStatKey{}
		if key == emptyStatKey {
			return
		}
		idx := util.GetShardIndex(key, e.numShards)

		shard := &e.shards[idx]
		shard.mux.Lock()
		defer shard.mux.Unlock()
		stat, found := shard.deviceStatMap[key]
		if !found {
			stat = NewDeviceStat()
		}
		stat.Update()
		shard.deviceStatMap[key] = stat
	} else if call.Metric == cloudcommon.CustomMetric {
		key := call.CustomStatKey
		emptyStatKey := CustomStatKey{}
		if key == emptyStatKey {
			return
		}
		idx := util.GetShardIndex(call.CustomStatKey, e.numShards)

		shard := &e.shards[idx]
		shard.mux.Lock()
		defer shard.mux.Unlock()
		stat, found := shard.customStatMap[key]
		if !found {
			stat = NewCustomStat()
		}
		stat.Update(call.CustomStatInfo)
		shard.customStatMap[key] = stat
	}
}

func (e *EdgeEventStats) RunNotify() {
	done := false
	for !done {
		select {
		case <-time.After(time.Now().Truncate(e.interval).Add(e.interval).Sub(time.Now())):
			span := log.StartSpan(log.DebugLevelMetrics, "edgeevents-stats")
			ctx := log.ContextWithSpan(context.Background(), span)

			for ii, _ := range e.shards {
				ts, _ := types.TimestampProto(time.Now())
				e.shards[ii].mux.Lock()
				for key, stat := range e.shards[ii].latencyStatMap {
					if stat.Changed {
						metric := LatencyStatToMetric(ts, key, stat)
						e.send(ctx, metric)
						stat.ResetLatencyStat()
						e.shards[ii].latencyStatMap[key] = stat
					}
				}
				for key, stat := range e.shards[ii].deviceStatMap {
					if stat.Changed && stat.NumSessions > 0 {
						metric := DeviceStatToMetric(ts, key, stat)
						e.send(ctx, metric)
						e.shards[ii].deviceStatMap[key] = NewDeviceStat()
					}
				}
				for key, stat := range e.shards[ii].customStatMap {
					if stat.Changed {
						metric := CustomStatToMetric(ts, key, stat)
						e.send(ctx, metric)
						e.shards[ii].customStatMap[key] = NewCustomStat()
					}
				}
				e.shards[ii].mux.Unlock()
			}
		case <-e.stop:
			done = true
		}
	}
	e.waitGroup.Done()
}

func LatencyStatToMetric(ts *types.Timestamp, key LatencyStatKey, stat *LatencyStat) *edgeproto.Metric {
	metric := initMetric(cloudcommon.LatencyMetric, *ts, key.AppInstKey)
	// Add tags (independent variables)
	metric.AddTag("locationtile", key.LocationTile)
	metric.AddTag("devicecarrier", key.DeviceCarrier)
	metric.AddTag("datanetworktype", key.DataNetworkType)
	metric.AddTag("deviceos", key.DeviceOs)
	metric.AddTag("devicemodel", key.DeviceModel)
	metric.AddIntVal("signalstrength", key.SignalStrength)
	// Latency information
	metric.AddDoubleVal("avg", stat.RollingStatistics.Statistics.Avg)
	metric.AddDoubleVal("variance", stat.RollingStatistics.Statistics.Variance)
	metric.AddDoubleVal("stddev", stat.RollingStatistics.Statistics.StdDev)
	metric.AddDoubleVal("min", stat.RollingStatistics.Statistics.Min)
	metric.AddDoubleVal("max", stat.RollingStatistics.Statistics.Max)
	stat.LatencyCounts.AddToMetric(metric)
	// Additional latency information for calculations when downsampling/aggregating further
	metric.AddIntVal("numsamples", stat.RollingStatistics.Statistics.NumSamples)
	metric.AddDoubleVal("total", stat.RollingStatistics.Statistics.Avg*float64(stat.RollingStatistics.Statistics.NumSamples))
	return metric
}

func DeviceStatToMetric(ts *types.Timestamp, key DeviceStatKey, stat *DeviceStat) *edgeproto.Metric {
	metric := initMetric(cloudcommon.DeviceMetric, *ts, key.AppInstKey)
	// Add tags (independent variables)
	metric.AddTag("locationtile", key.LocationTile)
	metric.AddTag("devicecarrier", key.DeviceCarrier)
	metric.AddTag("datanetworktype", key.DataNetworkType)
	metric.AddTag("deviceos", key.DeviceOs)
	metric.AddTag("devicemodel", key.DeviceModel)
	metric.AddIntVal("signalstrength", key.SignalStrength)
	// Num session information
	metric.AddIntVal("numsessions", stat.NumSessions)
	return metric
}

func CustomStatToMetric(ts *types.Timestamp, key CustomStatKey, stat *CustomStat) *edgeproto.Metric {
	metric := initMetric(cloudcommon.CustomMetric, *ts, key.AppInstKey)
	// Custom Stats info
	metric.AddTag("statname", key.Name)
	metric.AddIntVal("count", stat.Count)
	metric.AddDoubleVal("avg", stat.RollingStatistics.Statistics.Avg)
	metric.AddDoubleVal("variance", stat.RollingStatistics.Statistics.Variance)
	metric.AddDoubleVal("stddev", stat.RollingStatistics.Statistics.StdDev)
	metric.AddDoubleVal("min", stat.RollingStatistics.Statistics.Min)
	metric.AddDoubleVal("max", stat.RollingStatistics.Statistics.Max)
	metric.AddIntVal("numsamples", stat.RollingStatistics.Statistics.NumSamples)
	return metric
}

// Helper function that adds in appinst info, metric name, metric timestamp, and dme cloudlet info
func initMetric(metricName string, ts types.Timestamp, appInstKey edgeproto.AppInstKey) *edgeproto.Metric {
	metric := &edgeproto.Metric{}
	metric.Timestamp = ts
	metric.Name = metricName
	metric.AddTag("dmecloudlet", MyCloudletKey.Name)
	metric.AddTag("dmecloudletorg", MyCloudletKey.Organization)
	// AppInst information
	metric.AddTag("app", appInstKey.AppKey.Name)
	metric.AddTag("apporg", appInstKey.AppKey.Organization)
	metric.AddTag("ver", appInstKey.AppKey.Version)
	metric.AddTag("cloudlet", appInstKey.ClusterInstKey.CloudletKey.Name)
	metric.AddTag("cloudletorg", appInstKey.ClusterInstKey.CloudletKey.Organization)
	metric.AddTag("cluster", appInstKey.ClusterInstKey.ClusterKey.Name)
	metric.AddTag("clusterorg", appInstKey.ClusterInstKey.Organization)
	return metric
}
