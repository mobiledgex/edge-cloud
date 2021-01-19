package dmecommon

import (
	"strconv"
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
	Key                EdgeEventStatKey
	AppInstLatencyInfo *AppInstLatencyInfo // Latency samples for EdgeEvents
	GpsLocationInfo    *GpsLocationInfo    // Gps Location update
	CustomStatInfo     *CustomStatInfo     // Custom stat update
}

type EdgeEventStat struct {
	AppInstLatencyStats *AppInstLatencyStats // Aggregated latency stats from persistent connection (resets every hour)
	GpsLocationStats    *GpsLocationStats    // Gps Locations update
	CustomStats         *CustomStats         // Application defined custom stats
	Mux                 sync.Mutex
	Changed             bool
}

type EdgeEventMapShard struct {
	edgeEventStatMap map[EdgeEventStatKey]*EdgeEventStat
	notify           bool
	mux              sync.Mutex
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
		e.shards[ii].edgeEventStatMap = make(map[EdgeEventStatKey]*EdgeEventStat)
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
	defer e.mux.Unlock()
	close(e.stop)
	e.waitGroup.Wait()
	e.stop = nil
}

func (e *EdgeEventStats) UpdateSettings(interval time.Duration) {
	if e.interval == interval {
		return
	}
	restart := false
	if e.stop != nil {
		e.Stop()
		restart = true
	}
	e.mux.Lock()
	e.interval = interval
	e.mux.Unlock()
	if restart {
		e.Start()
	}
}

func (e *EdgeEventStats) LookupEdgeEventStatCall(call *EdgeEventStatCall) (*EdgeEventStat, bool) {
	idx := util.GetShardIndex(call.Key, e.numShards)

	shard := &e.shards[idx]
	shard.mux.Lock()
	defer shard.mux.Unlock()
	stat, found := shard.edgeEventStatMap[call.Key]
	return stat, found
}

func (e *EdgeEventStats) RecordEdgeEventStatCall(call *EdgeEventStatCall) {
	idx := util.GetShardIndex(call.Key, e.numShards)

	shard := &e.shards[idx]
	shard.mux.Lock()
	stat, found := shard.edgeEventStatMap[call.Key]
	if !found {
		stat = &EdgeEventStat{}
		shard.edgeEventStatMap[call.Key] = stat
	}
	if call.Key.Metric == cloudcommon.AppInstLatencyMetric {
		if stat.AppInstLatencyStats == nil {
			stat.AppInstLatencyStats = NewAppInstLatencyStats(LatencyTimes)
		}
		stat.AppInstLatencyStats.Update(call.AppInstLatencyInfo)
	} else if call.Key.Metric == cloudcommon.GpsLocationMetric {
		if stat.GpsLocationStats == nil {
			stat.GpsLocationStats = NewGpsLocationStats()
		}
		stat.GpsLocationStats.Stats = append(stat.GpsLocationStats.Stats, call.GpsLocationInfo)
	} else if call.Key.Metric == cloudcommon.CustomMetric {
		if stat.CustomStats == nil {
			stat.CustomStats = NewCustomStats()
		}
		stat.CustomStats.Update(call.CustomStatInfo)
	}
	stat.Changed = true
	shard.mux.Unlock()
}

func (e *EdgeEventStats) RunNotify() {
	done := false
	for !done {
		select {
		case <-time.After(time.Now().Truncate(e.interval).Add(e.interval).Sub(time.Now())):
			span := log.StartSpan(log.DebugLevelMetrics, "edgeevents-stats")
			ctx := log.ContextWithSpan(context.Background(), span)

			ts, _ := types.TimestampProto(time.Now())
			for ii, _ := range e.shards {
				e.shards[ii].mux.Lock()
				for key, stat := range e.shards[ii].edgeEventStatMap {
					if stat.Changed {
						switch key.Metric {
						case cloudcommon.AppInstLatencyMetric:
							metrics := getAppInstLatencyStatsToMetrics(ts, &key, stat)
							for _, metric := range metrics {
								e.send(ctx, metric)
							}
							stat.AppInstLatencyStats = nil
							stat.Changed = false
						case cloudcommon.GpsLocationMetric:
							for _, gpsStat := range stat.GpsLocationStats.Stats {
								e.send(ctx, GpsLocationStatToMetric(ts, &key, stat, gpsStat))
							}
							stat.GpsLocationStats = nil
							stat.Changed = false
						case cloudcommon.CustomMetric:
							cmetrics := getCustomStatsToMetrics(ts, &key, stat)
							for _, cmetric := range cmetrics {
								e.send(ctx, cmetric)
							}
							stat.CustomStats = nil
							stat.Changed = false
						default:
							continue
						}
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

// Compiles all of the AppInstLatencyStats fields into metrics, returns a slice of metrics
func getAppInstLatencyStatsToMetrics(ts *types.Timestamp, key *EdgeEventStatKey, stat *EdgeEventStat) []*edgeproto.Metric {
	metrics := make([]*edgeproto.Metric, 0)
	// Latency per Appinst (LatencyTotal) metric
	metrics = append(metrics, AppInstLatencyStatToMetric(ts, key, stat, stat.AppInstLatencyStats.LatencyTotal, cloudcommon.AppInstLatencyMetric, nil))
	// Latency per carrier metrics
	for carrier, latencystats := range stat.AppInstLatencyStats.LatencyPerCarrier {
		metrics = append(metrics, AppInstLatencyStatToMetric(ts, key, stat, latencystats, cloudcommon.LatencyPerCarrierMetric, map[string]string{"carrier": carrier}))
	}
	// Latency per data network type metrics
	for netdatatype, latencystats := range stat.AppInstLatencyStats.LatencyPerNetDataType {
		metrics = append(metrics, AppInstLatencyStatToMetric(ts, key, stat, latencystats, cloudcommon.LatencyPerDataNetworkMetric, map[string]string{"networkdatatype": netdatatype}))
	}
	// Latency per location metrics
	for dist, dirmap := range stat.AppInstLatencyStats.LatencyPerLoc {
		for orientation, latencystats := range dirmap {
			metrics = append(metrics, AppInstLatencyStatToMetric(ts, key, stat, latencystats, cloudcommon.LatencyPerLocationMetric, map[string]string{"distance": strconv.Itoa(dist), "direction": orientation}))
		}
	}
	return metrics
}

func getCustomStatsToMetrics(ts *types.Timestamp, key *EdgeEventStatKey, stat *EdgeEventStat) []*edgeproto.Metric {
	cmetrics := make([]*edgeproto.Metric, 0)
	for name, cstat := range stat.CustomStats.Stats {
		cmetrics = append(cmetrics, CustomStatToMetric(ts, key, stat.CustomStats, name, cstat))
	}
	return cmetrics
}

func AppInstLatencyStatToMetric(ts *types.Timestamp, key *EdgeEventStatKey, stat *EdgeEventStat, latencyStats *LatencyStats, metricName string, indVarMap map[string]string) *edgeproto.Metric {
	metric := initMetric(metricName, *ts, key.AppInstKey)
	// Latency information
	for indepVarName, indepVar := range indVarMap {
		metric.AddTag(indepVarName, indepVar) // eg. "carrier" -> "GDDT"
	}
	elapsed := ts.Seconds - int64(stat.AppInstLatencyStats.StartTime.Second())
	metric.AddIntVal("timeelapsed", uint64(elapsed))
	metric.AddIntVal("numclients", latencyStats.RollingStatistics.NumUniqueClients)
	metric.AddDoubleVal("avg", latencyStats.RollingStatistics.Statistics.Avg)
	metric.AddDoubleVal("stddev", latencyStats.RollingStatistics.Statistics.StdDev)
	metric.AddDoubleVal("min", latencyStats.RollingStatistics.Statistics.Min)
	metric.AddDoubleVal("max", latencyStats.RollingStatistics.Statistics.Max)
	latencyStats.LatencyCounts.AddToMetric(metric)
	return metric
}

func GpsLocationStatToMetric(ts *types.Timestamp, key *EdgeEventStatKey, stat *EdgeEventStat, gpsInfo *GpsLocationInfo) *edgeproto.Metric {
	metric := initMetric(cloudcommon.GpsLocationMetric, types.Timestamp(gpsInfo.Timestamp), key.AppInstKey)
	// GpsLocation information
	metric.AddTag("carrier", gpsInfo.Carrier)
	metric.AddTag("deviceos", gpsInfo.DeviceOs)
	metric.AddTag("devicemodel", gpsInfo.DeviceModel)
	metric.AddDoubleVal("longitude", gpsInfo.GpsLocation.Longitude)
	metric.AddDoubleVal("latitude", gpsInfo.GpsLocation.Latitude)
	metric.AddTag("uniqueid", gpsInfo.SessionCookieKey.UniqueId)
	return metric
}

func CustomStatToMetric(ts *types.Timestamp, key *EdgeEventStatKey, customStats *CustomStats, statName string, customStat *CustomStat) *edgeproto.Metric {
	metric := initMetric(cloudcommon.CustomMetric, *ts, key.AppInstKey)
	// Custom Stats info
	metric.AddTag("statname", statName)
	elapsed := ts.Seconds - int64(customStats.StartTime.Second())
	metric.AddIntVal("timeelapsed", uint64(elapsed))
	metric.AddIntVal("count", customStat.Count)
	metric.AddIntVal("numclients", customStat.RollingStatistics.NumUniqueClients)
	metric.AddDoubleVal("avg", customStat.RollingStatistics.Statistics.Avg)
	metric.AddDoubleVal("stddev", customStat.RollingStatistics.Statistics.StdDev)
	metric.AddDoubleVal("min", customStat.RollingStatistics.Statistics.Min)
	metric.AddDoubleVal("max", customStat.RollingStatistics.Statistics.Max)
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
