package dmecommon

import (
	"math"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
)

// Filled in by DME. Added to EdgeEventStatCall to update stats
type AppInstLatencyInfo struct {
	DmeAppInst       *DmeAppInst  // Information about appinst
	Latency          *dme.Latency // Latency avg, min, max, and std dev
	SessionCookieKey CookieKey    // SessionCookie to identify unique clients for EdgeEvents
	GpsLocation      *dme.Loc
	DataNetworkType  string
	Carrier          string
	DeviceOs         string
}

// Rolling avg, min, max, std dev, and number of clients
type RollingLatency struct {
	Latency          *dme.Latency
	UniqueClients    map[CookieKey]int // Maps unique client to number of occurences of that unique client
	NumUniqueClients uint64
}

// Wrapper struct that holds latency stats
type LatencyStats struct {
	LatencyCounts  grpcstats.LatencyMetric // buckets for counts
	RollingLatency *RollingLatency         // General stats: Avg, StdDev, Min, Max, NumClients
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

func NewRollingLatency() *RollingLatency {
	r := new(RollingLatency)
	r.UniqueClients = make(map[CookieKey]int)
	r.Latency = new(dme.Latency)
	return r
}

// Add new samples to RollingLatency struct and update RollingLatency statistics
func (r *RollingLatency) UpdateRollingLatency(sessionCookieKey CookieKey, samples ...float64) {
	// Previous statistics used to calculate rolling variance
	prevNumSamples := r.Latency.NumSamples
	prevAvg := r.Latency.Avg
	prevVariance := r.Latency.Variance
	// Update Min, Max, and Avg
	total := r.Latency.Avg * float64(r.Latency.NumSamples)
	for _, sample := range samples {
		if sample < r.Latency.Min || r.Latency.Min == 0 {
			r.Latency.Min = sample
		}
		if sample > r.Latency.Max || r.Latency.Max == 0 {
			r.Latency.Max = sample
		}
		total += sample
		r.Latency.NumSamples++
		// Add client to UniqueClients map
		r.AddUniqueClient(sessionCookieKey)
	}
	// Return empty latency if no samples
	if r.Latency.NumSamples == 0 {
		r.Latency = new(dme.Latency)
		return
	}
	r.Latency.Avg = total / float64(r.Latency.NumSamples)
	// Calulate Rolling variance and std dev (Using Welford's Algorithm)
	// NewSumSquared = OldSumSquared + (sample - OldAverage)(sample - NewAverage)
	unbiasedPrevNumSamples := prevNumSamples - 1
	prevSumSquared := prevVariance * float64(unbiasedPrevNumSamples)
	newSumSquared := prevSumSquared
	for _, sample := range samples {
		newSumSquared += (sample - prevAvg) * (sample - r.Latency.Avg)
	}
	unbiasedNumSamples := r.Latency.NumSamples - 1
	if unbiasedNumSamples == 0 {
		unbiasedNumSamples = 1
	}
	r.Latency.Variance = newSumSquared / float64(unbiasedNumSamples)
	r.Latency.StdDev = math.Sqrt(r.Latency.Variance)
}

// Add client to map of UniqueClients, update NumUniqueClients
func (r *RollingLatency) AddUniqueClient(sessionCookieKey CookieKey) {
	num, ok := r.UniqueClients[sessionCookieKey]
	if !ok {
		r.UniqueClients[sessionCookieKey] = 1
	} else {
		r.UniqueClients[sessionCookieKey] = num + 1
	}
	r.NumUniqueClients = uint64(len(r.UniqueClients))
}

// Utility function that returns Latency struct with Avg, Min, Max, StdDev, and NumSamples
func CalculateLatency(samples []*dme.Sample) dme.Latency {
	// Create latency struct
	latency := new(dme.Latency)
	latency.NumSamples = uint64(len(samples))
	if latency.NumSamples == 0 {
		return *latency
	}
	// calculate Min, Max, and Avg
	sum := 0.0
	for _, sample := range samples {
		sum += sample.Value
		if latency.Min == 0.0 || sample.Value < latency.Min {
			latency.Min = sample.Value
		}
		if latency.Max == 0.0 || sample.Value > latency.Max {
			latency.Max = sample.Value
		}
	}
	latency.Avg = sum / float64(latency.NumSamples)
	// calculate StdDev
	diffSquared := 0.0
	for _, sample := range samples {
		diff := sample.Value - latency.Avg
		diffSquared += diff * diff
	}
	latency.Variance = diffSquared / float64(latency.NumSamples-1)
	latency.StdDev = math.Sqrt(latency.Variance)
	ts := cloudcommon.TimeToTimestamp(time.Now())
	latency.Timestamp = &ts
	return *latency
}

func NewLatencyStats(latencyBuckets []time.Duration) *LatencyStats {
	l := new(LatencyStats)
	grpcstats.InitLatencyMetric(&l.LatencyCounts, latencyBuckets)
	l.RollingLatency = NewRollingLatency()
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

func (a *AppInstLatencyStats) Update(data *AppInstLatencyInfo) {
	// Update LatencyTotal
	a.LatencyTotal.LatencyCounts.AddLatency(time.Duration(data.Latency.Avg) * time.Millisecond)
	a.LatencyTotal.RollingLatency.UpdateRollingLatency(data.SessionCookieKey, data.Latency.Avg)
	// Update LatencyPer Maps
	a.UpdateLatencyMaps(a.LatencyPerCarrier, data, data.Carrier, data.Latency.Avg)
	a.UpdateLatencyMaps(a.LatencyPerNetDataType, data, data.DataNetworkType, data.Latency.Avg)
	a.UpdateLatencyMaps(a.LatencyPerDeviceOs, data, data.DeviceOs, data.Latency.Avg)
	if data.DmeAppInst != nil {
		// Figure out distance and figure out orientation
		distBucket := GetDistanceBucketFromAppInst(data.DmeAppInst, *data.GpsLocation)
		bearing := GetBearingFromAppInst(data.DmeAppInst, *data.GpsLocation)
		a.UpdateLatencyMaps(a.LatencyPerLoc[distBucket], data, string(bearing), data.Latency.Avg)
	}
}

func (a *AppInstLatencyStats) UpdateLatencyMaps(latencyMap map[string]*LatencyStats, data *AppInstLatencyInfo, key string, sample float64) {
	if key == "" {
		key = "unknown"
	}
	dsample := time.Duration(sample) * time.Millisecond
	val, ok := latencyMap[key]
	if ok {
		val.LatencyCounts.AddLatency(dsample)
		val.RollingLatency.UpdateRollingLatency(data.SessionCookieKey, sample)
	} else {
		val = NewLatencyStats(a.latencyBuckets)
		val.LatencyCounts.AddLatency(dsample)
		val.RollingLatency.UpdateRollingLatency(data.SessionCookieKey, sample)
	}
	latencyMap[key] = val
}
