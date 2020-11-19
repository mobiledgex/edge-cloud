package dmeutil

import (
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
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

type AppInstLatency struct {
	Latency         *dme.Latency
	SessionCookie   string // SessionCookie to identify unique clients for EdgeEvents
	GpsLocation     *dme.Loc
	DataNetworkType string
	Carrier         string
	DeviceOs        string
}

type AppInstLatencyStats struct {
	latencyBuckets        []time.Duration
	LatencyAvg            grpcstats.LatencyMetric
	LatencyStdDev         grpcstats.LatencyMetric
	LatencyPerLoc         map[string]grpcstats.LatencyMetric
	LatencyPerCarrier     map[string]grpcstats.LatencyMetric
	LatencyPerNetDataType map[string]grpcstats.LatencyMetric
	LatencyPerDeviceOs    map[string]grpcstats.LatencyMetric
	uniqueClients         map[string]int // Maps unique client to number of occurences of that unique client
	NumUniqueClients      uint64
	StartTime             dme.Timestamp // denotes when this struct began aggregating stats
}

func NewAppInstLatencyStats(latencyBuckets []time.Duration) *AppInstLatencyStats {
	a := new(AppInstLatencyStats)
	a.latencyBuckets = latencyBuckets
	grpcstats.InitLatencyMetric(&a.LatencyAvg, latencyBuckets)
	grpcstats.InitLatencyMetric(&a.LatencyStdDev, latencyBuckets)
	a.LatencyPerLoc = make(map[string]grpcstats.LatencyMetric)
	a.LatencyPerCarrier = make(map[string]grpcstats.LatencyMetric)
	a.LatencyPerNetDataType = make(map[string]grpcstats.LatencyMetric)
	a.LatencyPerDeviceOs = make(map[string]grpcstats.LatencyMetric)
	a.uniqueClients = make(map[string]int)
	a.StartTime = cloudcommon.TimeToTimestamp(time.Now())
	return a
}

func (a *AppInstLatencyStats) Update(appInstLatency *AppInstLatency) {
	latencyAvg := time.Duration(appInstLatency.Latency.Avg) * time.Millisecond
	latencyStdDev := time.Duration(appInstLatency.Latency.StdDev) * time.Millisecond
	// Update LatencyAvg and LatencyStdDev
	a.LatencyAvg.AddLatency(latencyAvg)
	a.LatencyStdDev.AddLatency(latencyStdDev)
	// Update LatencyPer Maps
	a.UpdateLatencyMaps(a.LatencyPerCarrier, appInstLatency.Carrier, latencyAvg)
	a.UpdateLatencyMaps(a.LatencyPerNetDataType, appInstLatency.DataNetworkType, latencyAvg)
	a.UpdateLatencyMaps(a.LatencyPerDeviceOs, appInstLatency.DeviceOs, latencyAvg)
	// TODO: Algorithmically determine location "section"
	a.UpdateLatencyMaps(a.LatencyPerLoc, "Northeast", latencyAvg)
	a.addUniqueClient(appInstLatency.SessionCookie)
}

func (a *AppInstLatencyStats) UpdateLatencyMaps(latencyMap map[string]grpcstats.LatencyMetric, key string, sample time.Duration) {
	val, ok := latencyMap[key]
	if ok {
		val.AddLatency(sample)
	} else {
		grpcstats.InitLatencyMetric(&val, a.latencyBuckets)
		val.AddLatency(sample)
	}
	latencyMap[key] = val
}

// Add client to map of UniqueClients, update NumUniqueClients
func (a *AppInstLatencyStats) addUniqueClient(sessionCookie string) {
	num, ok := a.uniqueClients[sessionCookie]
	if !ok {
		a.uniqueClients[sessionCookie] = 1
	} else {
		a.uniqueClients[sessionCookie] = num + 1
	}
	a.NumUniqueClients = uint64(len(a.uniqueClients))
}
