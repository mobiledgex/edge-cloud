package dmecommon

import (
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
)

// Constants for Debug
const (
	RequestAppInstLatency = "request-appinst-latency"
)

// Distances from Device to AppInst in km
var GpsLocationDists = []int{
	1,
	2,
	5,
	10,
	50,
	100,
}

var GpsLocationQuadrants = []string{
	"Northeast",
	"Southeast",
	"Southwest",
	"Northwest",
}

type GpsLocationStats struct {
	Stats     []*GpsLocationStat
	StartTime time.Time
}

// TODO: RENAME THESE STRUCTS FOR MORE CLARITY
type GpsLocationStat struct {
	GpsLocation      *dme.Loc
	SessionCookieKey CookieKey
	Timestamp        dme.Timestamp
	Carrier          string
}

func NewGpsLocationStats() *GpsLocationStats {
	g := new(GpsLocationStats)
	g.Stats = make([]*GpsLocationStat, 0)
	g.StartTime = time.Now()
	return g
}

// Filled in by DME. Used to process samples into trends/stats
type AppInstLatencyData struct {
	Samples          []*dme.Sample
	Latency          *dme.Latency
	SessionCookieKey CookieKey // SessionCookie to identify unique clients for EdgeEvents
	GpsLocation      *dme.Loc
	DataNetworkType  string
	Carrier          string
	DeviceOs         string
}

type LatencyStats struct {
	LatencyMetric  grpcstats.LatencyMetric // buckets for counts
	RollingLatency *RollingLatency         // General stats: Avg, StdDev, Min, Max, NumClients
}

func NewLatencyStats(latencyBuckets []time.Duration) *LatencyStats {
	l := new(LatencyStats)
	grpcstats.InitLatencyMetric(&l.LatencyMetric, latencyBuckets)
	l.RollingLatency = NewRollingLatency()
	return l
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

func NewAppInstLatencyStats(latencyBuckets []time.Duration) *AppInstLatencyStats {
	a := new(AppInstLatencyStats)
	a.latencyBuckets = latencyBuckets
	a.LatencyTotal = NewLatencyStats(latencyBuckets)
	a.LatencyPerLoc = make(map[int]map[string]*LatencyStats)
	for _, dist := range GpsLocationDists {
		a.LatencyPerLoc[dist] = make(map[string]*LatencyStats)
	}
	a.LatencyPerCarrier = make(map[string]*LatencyStats)
	a.LatencyPerNetDataType = make(map[string]*LatencyStats)
	a.LatencyPerDeviceOs = make(map[string]*LatencyStats)
	a.StartTime = time.Now()
	return a
}

func (a *AppInstLatencyStats) Update(data *AppInstLatencyData) {
	latencyAvg := time.Duration(data.Latency.Avg) * time.Millisecond
	// Update LatencyTotal
	a.LatencyTotal.LatencyMetric.AddLatency(latencyAvg)
	a.LatencyTotal.RollingLatency.UpdateRollingLatency(data.Samples, data.SessionCookieKey)
	// Update LatencyPer Maps
	a.UpdateLatencyMaps(a.LatencyPerCarrier, data, data.Carrier, latencyAvg)
	a.UpdateLatencyMaps(a.LatencyPerNetDataType, data, data.DataNetworkType, latencyAvg)
	a.UpdateLatencyMaps(a.LatencyPerDeviceOs, data, data.DeviceOs, latencyAvg)
	// TODO: Algorithmically determine location "section"
	// Figure out distance and figure out orientation
	a.UpdateLatencyMaps(a.LatencyPerLoc[1], data, "Northeast", latencyAvg)
}

// TODO: SHOULD ROLLINGLATENCY AND BUCKETS BE PER SAMPLE OR PER BATCH (ie. avg for a batch)???
func (a *AppInstLatencyStats) UpdateLatencyMaps(latencyMap map[string]*LatencyStats, data *AppInstLatencyData, key string, sample time.Duration) {
	if key == "" {
		key = "unknown"
	}
	val, ok := latencyMap[key]
	if ok {
		val.LatencyMetric.AddLatency(sample)
		val.RollingLatency.UpdateRollingLatency(data.Samples, data.SessionCookieKey)
	} else {
		val = NewLatencyStats(a.latencyBuckets)
		val.LatencyMetric.AddLatency(sample)
		val.RollingLatency.UpdateRollingLatency(data.Samples, data.SessionCookieKey)
	}
	latencyMap[key] = val
}
