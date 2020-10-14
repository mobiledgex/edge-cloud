package dmeutil

import (
	"math"
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/stretchr/testify/require"
)

var errorThreshold = .00001

func TestLatencyCalculations(t *testing.T) {
	// Exepected values confirmed using Average and Standard Deviation Calculators
	// Test min, max, avg, variance, stddev, numsamples calculation
	samples1 := []float64{3.1, 2.3, 4.5, 1.2, 13.4}
	latency1 := CalculateLatency(samples1)
	require.Equal(t, 1.2, latency1.Min)
	require.Equal(t, 13.4, latency1.Max)
	require.True(t, math.Abs(4.9-latency1.Avg) < errorThreshold)
	require.True(t, math.Abs(24.025-latency1.Variance) < errorThreshold)
	require.True(t, math.Abs(4.901530373261-latency1.StdDev) < errorThreshold)
	require.Equal(t, uint64(5), latency1.NumSamples)

	// Test min, max, avg, variance, stddev, numsamples calculation with UpdateRollingLatency
	samples2 := []float64{.53, 21.3, 14.2, 6.7, 8.8}
	latency2 := new(dme.Latency)
	UpdateRollingLatency(samples2, latency2)
	require.Equal(t, .53, latency2.Min)
	require.Equal(t, 21.3, latency2.Max)
	require.True(t, math.Abs(10.306-latency2.Avg) < errorThreshold)
	require.True(t, math.Abs(61.71818-latency2.Variance) < errorThreshold)
	require.True(t, math.Abs(7.8560919037394-latency2.StdDev) < errorThreshold)
	require.Equal(t, uint64(5), latency2.NumSamples)

	// Test rolling avg, min, max, stddev, variance, numsamples
	// Update latency2 with samples1
	UpdateRollingLatency(samples1, latency2)
	require.Equal(t, .53, latency2.Min)
	require.Equal(t, 21.3, latency2.Max)
	require.True(t, math.Abs(7.603-latency2.Avg) < errorThreshold)
	require.True(t, math.Abs(46.22609-latency2.Variance) < errorThreshold)
	require.True(t, math.Abs(6.7989771289511-latency2.StdDev) < errorThreshold)
	require.Equal(t, uint64(10), latency2.NumSamples)
}
