package dmeutil

import (
	"math"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

// Constants for Debug
const (
	RequestAppInstLatency = "request-appinst-latency"
	ShowAppInstLatency    = "show-appinst-latency"
)

type RollingLatency struct {
	Latency          dme.Latency
	UniqueClients    map[string]struct{}
	NumUniqueClients uint64
}

// TODO: Store and correlate gps, time, and latency
type Sample struct {
	Sample          float64
	GpsLocation     dme.Loc
	Client          string // Client session cookie
	DataNetworkType string // eg. LTE, 5G, etc.
	Timestamp       dme.Timestamp
}

func NewRollingLatency() *RollingLatency {
	r := new(RollingLatency)
	r.UniqueClients = make(map[string]struct{})
	return r
}

// Update rolling Avg, Min, Max, StdDev, and NumSamples in provided Latency struct
func (r *RollingLatency) UpdateRollingLatency(samples []float64, sessionCookie string) {
	// Add client to UniqueClients map
	r.UpdateUniqueClients(sessionCookie)
	// First samples
	if r.Latency.NumSamples == 0 {
		r.Latency = CalculateLatency(samples)
		return
	}
	// Previous statistics used to calculate rolling variance
	prevNumSamples := r.Latency.NumSamples
	prevAvg := r.Latency.Avg
	prevVariance := r.Latency.Variance
	// Update Min, Max, and Avg
	total := r.Latency.Avg * float64(r.Latency.NumSamples)
	for _, sample := range samples {
		if sample < r.Latency.Min {
			r.Latency.Min = sample
		}
		if sample > r.Latency.Max {
			r.Latency.Max = sample
		}
		total += sample
		r.Latency.NumSamples++
	}
	r.Latency.Avg = total / float64(r.Latency.NumSamples)
	// Calulate Rolling variance and std dev (Using Welford's Algorithm)
	// NewSumSquared = OldSumSquared + (sample - OldAverage)(sample - NewAverage)
	prevSumSquared := prevVariance * float64(prevNumSamples-1)
	newSumSquared := prevSumSquared
	for _, sample := range samples {
		newSumSquared += (sample - prevAvg) * (sample - r.Latency.Avg)
	}
	r.Latency.Variance = newSumSquared / float64(r.Latency.NumSamples-1)
	r.Latency.StdDev = math.Sqrt(r.Latency.Variance)
}

func (r *RollingLatency) UpdateUniqueClients(sessionCookie string) {
	r.UniqueClients[sessionCookie] = struct{}{}
	r.NumUniqueClients = uint64(len(r.UniqueClients))
}

// Return Latency struct with Avg, Min, Max, StdDev, and NumSamples
func CalculateLatency(samples []float64) dme.Latency {
	// Create latency struct
	latency := new(dme.Latency)
	// calculate Min, Max, and Avg
	latency.NumSamples = uint64(len(samples))
	sum := 0.0
	for _, sample := range samples {
		sum += sample
		if latency.Min == 0.0 || sample < latency.Min {
			latency.Min = sample
		}
		if latency.Max == 0.0 || sample > latency.Max {
			latency.Max = sample
		}
	}
	latency.Avg = sum / float64(latency.NumSamples)
	// calculate StdDev
	diffSquared := 0.0
	for _, sample := range samples {
		diff := sample - latency.Avg
		diffSquared += diff * diff
	}
	latency.Variance = diffSquared / float64(latency.NumSamples-1)
	latency.StdDev = math.Sqrt(latency.Variance)
	ts := cloudcommon.TimeToTimestamp(time.Now())
	latency.Timestamp = &ts
	return *latency
}
