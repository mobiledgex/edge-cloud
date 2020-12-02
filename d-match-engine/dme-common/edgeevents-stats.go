package dmecommon

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
	"golang.org/x/net/context"
)

var EEStats *EdgeEventStats

type EdgeEventStatCall struct {
	Key                EdgeEventStatKey
	AppInstLatencyInfo *AppInstLatencyInfo // Latency samples for EdgeEvents
	GpsLocationInfo    *GpsLocationInfo    // Gps Location update
}

type EdgeEventStat struct {
	AppInstLatencyStats *AppInstLatencyStats // Aggregated latency stats from persistent connection (resets every hour)
	GpsLocationStats    *GpsLocationStats    // Gps Locations update
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
	e.stop = make(chan struct{})
	e.waitGroup.Add(1)
	go e.RunNotify()
}

func (e *EdgeEventStats) Stop() {
	close(e.stop)
	e.waitGroup.Wait()
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
	}
	stat.Changed = true
	shard.mux.Unlock()
}

func (e *EdgeEventStats) RunNotify() {
	done := false
	ctx := context.Background()
	for !done {
		select {
		case <-time.After(e.interval):
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
	// Latency per device os metrics
	for deviceos, latencystats := range stat.AppInstLatencyStats.LatencyPerDeviceOs {
		metrics = append(metrics, AppInstLatencyStatToMetric(ts, key, stat, latencystats, cloudcommon.LatencyPerDeviceOSMetric, map[string]string{"deviceos": deviceos}))
	}
	// Latency per location metrics
	for dist, dirmap := range stat.AppInstLatencyStats.LatencyPerLoc {
		for orientation, latencystats := range dirmap {
			metrics = append(metrics, AppInstLatencyStatToMetric(ts, key, stat, latencystats, cloudcommon.LatencyPerLocationMetric, map[string]string{"distance": strconv.Itoa(dist), "direction": orientation}))
		}
	}
	return metrics
}

func AppInstLatencyStatToMetric(ts *types.Timestamp, key *EdgeEventStatKey, stat *EdgeEventStat, latencyStats *LatencyStats, metricName string, indVarMap map[string]string) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Timestamp = *ts
	metric.Name = metricName
	metric.AddTag("dmecloudlet", MyCloudletKey.Name)
	metric.AddTag("dmecloudletorg", MyCloudletKey.Organization)
	// AppInst information
	metric.AddTag("app", key.AppInstKey.AppKey.Name)
	metric.AddTag("apporg", key.AppInstKey.AppKey.Organization)
	metric.AddTag("ver", key.AppInstKey.AppKey.Version)
	metric.AddTag("cloudlet", key.AppInstKey.ClusterInstKey.CloudletKey.Name)
	metric.AddTag("cloudletorg", key.AppInstKey.ClusterInstKey.CloudletKey.Organization)
	metric.AddTag("cluster", key.AppInstKey.ClusterInstKey.ClusterKey.Name)
	metric.AddTag("clusterorg", key.AppInstKey.ClusterInstKey.Organization)
	// Latency information
	for indepVarName, indepVar := range indVarMap {
		metric.AddTag(indepVarName, indepVar) // eg. "carrier" -> "TDG"
	}
	elapsed := ts.Seconds - int64(stat.AppInstLatencyStats.StartTime.Second())
	metric.AddIntVal("timeelapsed", uint64(elapsed))
	metric.AddIntVal("numclients", latencyStats.RollingLatency.NumUniqueClients)
	metric.AddDoubleVal("avg", latencyStats.RollingLatency.Latency.Avg)
	metric.AddDoubleVal("stddev", latencyStats.RollingLatency.Latency.StdDev)
	metric.AddDoubleVal("min", latencyStats.RollingLatency.Latency.Min)
	metric.AddDoubleVal("max", latencyStats.RollingLatency.Latency.Max)
	latencyStats.LatencyCounts.AddToMetric(&metric)
	return &metric
}

func GpsLocationStatToMetric(ts *types.Timestamp, key *EdgeEventStatKey, stat *EdgeEventStat, gpsInfo *GpsLocationInfo) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Timestamp = types.Timestamp(gpsInfo.Timestamp)
	metric.Name = cloudcommon.GpsLocationMetric
	metric.AddTag("dmecloudlet", MyCloudletKey.Name)
	metric.AddTag("dmecloudletorg", MyCloudletKey.Organization)
	// AppInst information
	metric.AddTag("app", key.AppInstKey.AppKey.Name)
	metric.AddTag("apporg", key.AppInstKey.AppKey.Organization)
	metric.AddTag("ver", key.AppInstKey.AppKey.Version)
	metric.AddTag("cloudlet", key.AppInstKey.ClusterInstKey.CloudletKey.Name)
	metric.AddTag("cloudletorg", key.AppInstKey.ClusterInstKey.CloudletKey.Organization)
	metric.AddTag("cluster", key.AppInstKey.ClusterInstKey.ClusterKey.Name)
	metric.AddTag("clusterorg", key.AppInstKey.ClusterInstKey.Organization)
	// GpsLocation information
	metric.AddTag("carrier", gpsInfo.Carrier)
	metric.AddTag("deviceos", gpsInfo.DeviceOs)
	metric.AddDoubleVal("longitude", gpsInfo.GpsLocation.Longitude)
	metric.AddDoubleVal("latitude", gpsInfo.GpsLocation.Latitude)
	k := gpsInfo.SessionCookieKey
	cookieKeyStr := k.OrgName + k.AppName + k.AppVers + k.UniqueIdType + k.UniqueId + strconv.Itoa(k.Kid)
	hash := md5.Sum([]byte(cookieKeyStr))
	sessCookie := hex.EncodeToString(hash[:])
	metric.AddTag("sessioncookie", sessCookie)
	return &metric
}
