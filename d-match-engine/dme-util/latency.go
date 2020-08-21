package dmeutil

import (
	"fmt"
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

// Return Latency struct with Avg, Min, Max, StdDev, and NumSamples
func CalculateLatency(samples []float64) *dme.Latency {
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
	return latency
}

// Update rolling Avg, Min, Max, StdDev, and NumSamples in provided Latency struct
func UpdateRollingLatency(samples []float64, latency *dme.Latency) error {
	if latency == nil {
		return fmt.Errorf("Latency is unititialized")
	}
	// First samples
	if latency.NumSamples == 0 {
		*latency = *CalculateLatency(samples)
		return nil
	}
	// Previous statistics used to calculate rolling variance
	prevNumSamples := latency.NumSamples
	prevAvg := latency.Avg
	prevVariance := latency.Variance
	// Update Min, Max, and Avg
	total := latency.Avg * float64(latency.NumSamples)
	for _, sample := range samples {
		if sample < latency.Min {
			latency.Min = sample
		}
		if sample > latency.Max {
			latency.Max = sample
		}
		total += sample
		latency.NumSamples++
	}
	latency.Avg = total / float64(latency.NumSamples)
	// Calulate Rolling variance and std dev (Using Welford's Algorithm)
	// NewSumSquared = OldSumSquared + (sample - OldAverage)(sample - NewAverage)
	prevSumSquared := prevVariance * float64(prevNumSamples-1)
	newSumSquared := prevSumSquared
	for _, sample := range samples {
		newSumSquared += (sample - prevAvg) * (sample - latency.Avg)
	}
	latency.Variance = newSumSquared / float64(latency.NumSamples-1)
	latency.StdDev = math.Sqrt(latency.Variance)
	return nil
}
