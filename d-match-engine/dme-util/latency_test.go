package dmeutil

import (
	"math"
	"testing"

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
	r := NewRollingLatency()
	samples2 := []float64{.53, 21.3, 14.2, 6.7, 8.8}
	r.UpdateRollingLatency(samples2, "123")
	require.Equal(t, .53, r.Latency.Min)
	require.Equal(t, 21.3, r.Latency.Max)
	require.True(t, math.Abs(10.306-r.Latency.Avg) < errorThreshold)
	require.True(t, math.Abs(61.71818-r.Latency.Variance) < errorThreshold)
	require.True(t, math.Abs(7.8560919037394-r.Latency.StdDev) < errorThreshold)
	require.Equal(t, uint64(5), r.Latency.NumSamples)
	require.Equal(t, uint64(1), r.NumUniqueClients)

	// Test rolling avg, min, max, stddev, variance, numsamples
	// Update latency2 with samples1
	r.UpdateRollingLatency(samples1, "234")
	require.Equal(t, .53, r.Latency.Min)
	require.Equal(t, 21.3, r.Latency.Max)
	require.True(t, math.Abs(7.603-r.Latency.Avg) < errorThreshold)
	require.True(t, math.Abs(46.22609-r.Latency.Variance) < errorThreshold)
	require.True(t, math.Abs(6.7989771289511-r.Latency.StdDev) < errorThreshold)
	require.Equal(t, uint64(10), r.Latency.NumSamples)
	require.Equal(t, uint64(2), r.NumUniqueClients)
}
