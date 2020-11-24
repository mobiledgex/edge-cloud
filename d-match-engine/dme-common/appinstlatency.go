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
	Samples          []*dme.Sample
	Latency          *dme.Latency
	SessionCookieKey CookieKey // SessionCookie to identify unique clients for EdgeEvents
	GpsLocation      *dme.Loc
	DataNetworkType  string
	Carrier          string
	DeviceOs         string
}

// Wrapper struct that holds latency stats
type LatencyStats struct {
	LatencyMetric  grpcstats.LatencyMetric // buckets for counts
	RollingLatency *RollingLatency         // General stats: Avg, StdDev, Min, Max, NumClients
}

// Rolling avg, min, max, std dev, and number of clients
type RollingLatency struct {
	Latency          *dme.Latency
	UniqueClients    map[CookieKey]int // Maps unique client to number of occurences of that unique client
	NumUniqueClients uint64
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
func (r *RollingLatency) UpdateRollingLatency(samples []*dme.Sample, sessionCookieKey CookieKey) {
	// Previous statistics used to calculate rolling variance
	prevNumSamples := r.Latency.NumSamples
	prevAvg := r.Latency.Avg
	prevVariance := r.Latency.Variance
	// Update Min, Max, and Avg
	total := r.Latency.Avg * float64(r.Latency.NumSamples)
	for _, sample := range samples {
		if sample.Value < r.Latency.Min || r.Latency.Min == 0 {
			r.Latency.Min = sample.Value
		}
		if sample.Value > r.Latency.Max || r.Latency.Max == 0 {
			r.Latency.Max = sample.Value
		}
		total += sample.Value
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
	prevSumSquared := prevVariance * float64(prevNumSamples-1)
	newSumSquared := prevSumSquared
	for _, sample := range samples {
		newSumSquared += (sample.Value - prevAvg) * (sample.Value - r.Latency.Avg)
	}
	r.Latency.Variance = newSumSquared / float64(r.Latency.NumSamples-1)
	r.Latency.StdDev = math.Sqrt(r.Latency.Variance)
}

/*
// Remove samples that are older than time.Now() - duration and update RollingLatency statistics
func (r *RollingLatency) RemoveOldSamples() {
	// Previous statistics used to calculate rolling variance
	prevNumSamples := r.Latency.NumSamples
	prevAvg := r.Latency.Avg
	prevVariance := r.Latency.Variance
	prevSamples := r.Samples
	// Update Min, Max, and Avg
	total := r.Latency.Avg * float64(r.Latency.NumSamples)
	recalculateLatency := false
	for _, sample := range r.Samples {
		if time.Since(cloudcommon.TimestampToTime(*sample.Timestamp)) > r.Duration {
			// remove first element in slice
			r.Samples = r.Samples[1:]
			total = total - sample.Value
			r.Latency.NumSamples--
			r.RemoveUniqueClient(sample.SessionCookie)
		} else {
			break
		}
		if sample.Value == r.Latency.Min || sample.Value == r.Latency.Max {
			recalculateLatency = true
		}
	}
	if recalculateLatency {
		// If a removed sample was min or max, we have to iterate through entire list of samples to find new min/max
		latency := CalculateLatency(r.Samples)
		r.Latency = &latency
	} else {
		// Update rolling Latency without iterating through entire list of samples
		// Return empty latency if no samples
		if r.Latency.NumSamples == 0 {
			r.Latency = new(dme.Latency)
			return
		}
		r.Latency.Avg = total / float64(r.Latency.NumSamples)
		// Calulate Rolling variance and std dev (Using Welford's Algorithm) (Removing samples)
		// NewSumSquared = OldSumSquared - (sample - OldAverage)(sample - NewAverage)
		prevSumSquared := prevVariance * float64(prevNumSamples-1)
		newSumSquared := prevSumSquared
		for _, sample := range prevSamples {
			if time.Since(cloudcommon.TimestampToTime(*sample.Timestamp)) > r.Duration {
				newSumSquared -= (sample.Value - prevAvg) * (sample.Value - r.Latency.Avg)
			} else {
				break
			}
		}
		r.Latency.Variance = newSumSquared / float64(r.Latency.NumSamples-1)
		r.Latency.StdDev = math.Sqrt(r.Latency.Variance)
	}
}
*/

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

/*
// Remove client to map of UniqueClients, update NumUniqueClients
func (r *RollingLatency) RemoveUniqueClient(sessionCookie string) {
	if num, ok := r.UniqueClients[sessionCookie]; ok {
		if num == 1 {
			delete(r.UniqueClients, sessionCookie)
		} else {
			r.UniqueClients[sessionCookie] = num - 1
		}
	}
	r.NumUniqueClients = uint64(len(r.UniqueClients))
}
*/

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
	grpcstats.InitLatencyMetric(&l.LatencyMetric, latencyBuckets)
	l.RollingLatency = NewRollingLatency()
	return l
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

func (a *AppInstLatencyStats) Update(data *AppInstLatencyInfo) {
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
func (a *AppInstLatencyStats) UpdateLatencyMaps(latencyMap map[string]*LatencyStats, data *AppInstLatencyInfo, key string, sample time.Duration) {
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
