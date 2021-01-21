package dmecommon

import (
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
)

// Filled in by DME. Added to EdgeEventStatCall to update stats
type AppInstLatencyInfo struct {
	Statistics       *dme.Statistics // Latency avg, min, max, and std dev
	SessionCookieKey *CookieKey      // SessionCookie to identify unique clients for EdgeEvents
	GpsLocation      *dme.Loc        // Client GPSLocation
	AppInstLocation  *dme.Loc        // AppInst GPSLocation
	DataNetworkType  string
	Carrier          string
}

// Wrapper struct that holds latency stats
type LatencyStats struct {
	LatencyCounts     grpcstats.LatencyMetric      // buckets for counts
	RollingStatistics *grpcstats.RollingStatistics // General stats: Avg, StdDev, Min, Max, NumClients
}

// Held in memory by ApiStat. Stats/trends per independent variable
// For example, this holds the statistics for Latency per carrier
type AppInstLatencyStats struct {
	latencyBuckets        []time.Duration
	LatencyTotal          *LatencyStats
	LatencyPerLoc         map[int]map[string]*LatencyStats
	LatencyPerCarrier     map[string]*LatencyStats
	LatencyPerNetDataType map[string]*LatencyStats
	LatencyPerDeviceOs    map[string]*LatencyStats
	StartTime             time.Time // denotes when this struct began aggregating stats
}

// Constants for Debug
const (
	RequestAppInstLatency = "request-appinst-latency"
)

func NewLatencyStats(latencyBuckets []time.Duration) *LatencyStats {
	l := new(LatencyStats)
	grpcstats.InitLatencyMetric(&l.LatencyCounts, latencyBuckets)
	l.RollingStatistics = grpcstats.NewRollingStatistics()
	return l
}

func NewAppInstLatencyStats(latencyBuckets []time.Duration) *AppInstLatencyStats {
	a := new(AppInstLatencyStats)
	a.latencyBuckets = latencyBuckets
	a.LatencyTotal = NewLatencyStats(latencyBuckets)
	a.LatencyPerLoc = make(map[int]map[string]*LatencyStats)
	for _, dist := range DistsFromDeviceToAppInst {
		a.LatencyPerLoc[dist] = make(map[string]*LatencyStats)
	}
	a.LatencyPerCarrier = make(map[string]*LatencyStats)
	a.LatencyPerNetDataType = make(map[string]*LatencyStats)
	a.LatencyPerDeviceOs = make(map[string]*LatencyStats)
	a.StartTime = time.Now()
	return a
}

func RecordAppInstLatencyStatCall(loc *dme.Loc, appInstKey *edgeproto.AppInstKey, sessionCookieKey *CookieKey, edgeEventsCookieKey *EdgeEventsCookieKey, stats *dme.Statistics, carrier string, deviceInfo dme.DeviceInfo) {
	if EEStats == nil {
		return
	}
	call := EdgeEventStatCall{}
	call.Key.AppInstKey = *appInstKey
	call.Key.Metric = cloudcommon.AppInstLatencyMetric // override method name
	dmecloudlet := findDmeCloudlet(appInstKey)
	call.AppInstLatencyInfo = &AppInstLatencyInfo{
		Statistics:       stats,
		SessionCookieKey: sessionCookieKey,
		Carrier:          translateCarrierName(carrier),
		GpsLocation:      loc,
		AppInstLocation:  &dmecloudlet.GpsLocation,
		DataNetworkType:  deviceInfo.DataNetworkType,
	}
	EEStats.RecordEdgeEventStatCall(&call)
}

func (a *AppInstLatencyStats) Update(data *AppInstLatencyInfo) {
	// Update LatencyTotal
	a.LatencyTotal.LatencyCounts.AddLatency(time.Duration(data.Statistics.Avg) * time.Millisecond)
	a.LatencyTotal.RollingStatistics.UpdateRollingStatistics(data.SessionCookieKey.UniqueId, data.Statistics.Avg)
	// Update LatencyPer Maps
	a.UpdateLatencyMaps(a.LatencyPerCarrier, data, data.Carrier, data.Statistics.Avg)
	a.UpdateLatencyMaps(a.LatencyPerNetDataType, data, data.DataNetworkType, data.Statistics.Avg)
	if data.AppInstLocation != nil {
		// Figure out distance and figure out orientation
		distBucket := GetDistanceBucket(*data.AppInstLocation, *data.GpsLocation)
		bearing := GetBearingFrom(*data.AppInstLocation, *data.GpsLocation)
		a.UpdateLatencyMaps(a.LatencyPerLoc[distBucket], data, string(bearing), data.Statistics.Avg)
	}
}

func (a *AppInstLatencyStats) UpdateLatencyMaps(latencyMap map[string]*LatencyStats, data *AppInstLatencyInfo, key string, sample float64) {
	dsample := time.Duration(sample) * time.Millisecond
	val, ok := latencyMap[key]
	if !ok {
		val = NewLatencyStats(a.latencyBuckets)
	}
	val.LatencyCounts.AddLatency(dsample)
	val.RollingStatistics.UpdateRollingStatistics(data.SessionCookieKey.UniqueId, sample)
	latencyMap[key] = val
}
