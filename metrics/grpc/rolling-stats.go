package grpcstats

import (
	"math"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

// Rolling avg, min, max, std dev, and number of clients
type RollingStatistics struct {
	Statistics dme.Statistics
}

func NewRollingStatistics() *RollingStatistics {
	r := new(RollingStatistics)
	r.Statistics = dme.Statistics{}
	return r
}

// Add new samples to RollingStatistics struct and update RollingLatency statistics
func (r *RollingStatistics) UpdateRollingStatistics(samples ...float64) {
	// return if no samples
	if len(samples) == 0 {
		return
	}
	// Previous statistics used to calculate rolling variance
	prevNumSamples := r.Statistics.NumSamples
	prevAvg := r.Statistics.Avg
	prevVariance := r.Statistics.Variance
	// Update Min, Max, and Avg
	total := r.Statistics.Avg * float64(r.Statistics.NumSamples)
	for _, sample := range samples {
		// Don't add negative numbers
		if sample < 0 {
			continue
		}
		if sample < r.Statistics.Min || r.Statistics.Min == 0 {
			r.Statistics.Min = sample
		}
		if sample > r.Statistics.Max || r.Statistics.Max == 0 {
			r.Statistics.Max = sample
		}
		total += sample
		r.Statistics.NumSamples++
	}
	// return if no valid samples
	if r.Statistics.NumSamples == 0 {
		return
	}
	r.Statistics.Avg = total / float64(r.Statistics.NumSamples)
	// Calulate Rolling variance and std dev (Using Welford's Algorithm)
	// NewSumSquared = OldSumSquared + (sample - OldAverage)(sample - NewAverage)
	unbiasedPrevNumSamples := prevNumSamples - 1
	prevSumSquared := prevVariance * float64(unbiasedPrevNumSamples)
	newSumSquared := prevSumSquared
	for _, sample := range samples {
		// Don't add negative numbers
		if sample < 0 {
			continue
		}
		newSumSquared += (sample - prevAvg) * (sample - r.Statistics.Avg)
	}
	unbiasedNumSamples := r.Statistics.NumSamples - 1
	if unbiasedNumSamples == 0 {
		unbiasedNumSamples = 1
	}
	r.Statistics.Variance = newSumSquared / float64(unbiasedNumSamples)
	r.Statistics.StdDev = math.Sqrt(r.Statistics.Variance)
}

// Utility function that returns Statistics struct with Avg, Min, Max, StdDev, and NumSamples
func CalculateStatistics(samples []*dme.Sample) dme.Statistics {
	// Create statistics struct
	statistics := new(dme.Statistics)
	ts := cloudcommon.TimeToTimestamp(time.Now())
	statistics.Timestamp = &ts
	// return if samples is nil
	if samples == nil {
		return *statistics
	}
	// return if no samples
	if len(samples) == 0 {
		return *statistics
	}
	// calculate Min, Max, and Avg
	sum := 0.0
	for _, sample := range samples {
		// Don't add negative numbers
		if sample.Value < 0 {
			continue
		}
		statistics.NumSamples++
		sum += sample.Value
		if statistics.Min == 0.0 || sample.Value < statistics.Min {
			statistics.Min = sample.Value
		}
		if statistics.Max == 0.0 || sample.Value > statistics.Max {
			statistics.Max = sample.Value
		}
	}
	// return if no valid samples
	if statistics.NumSamples == 0 {
		return *statistics
	}
	// calculate average
	statistics.Avg = sum / float64(statistics.NumSamples)
	// don't calculate variance and stddev if only one sample
	if statistics.NumSamples == 1 {
		return *statistics
	}
	// calculate StdDev
	diffSquared := 0.0
	for _, sample := range samples {
		// Don't add negative numbers
		if sample.Value < 0 {
			continue
		}
		diff := sample.Value - statistics.Avg
		diffSquared += diff * diff
	}
	statistics.Variance = diffSquared / float64(statistics.NumSamples-1)
	statistics.StdDev = math.Sqrt(statistics.Variance)
	return *statistics
}
