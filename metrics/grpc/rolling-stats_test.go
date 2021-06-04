package grpcstats

import (
	"math"
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/stretchr/testify/require"
)

var errorThreshold = .00001

func TestRollingStatsCalculations(t *testing.T) {
	// Expected values confirmed using Average and Standard Deviation Calculators
	// Test CalculateLatency: min, max, avg, variance, stddev, numsamples calculation
	samples1 := make([]*dme.Sample, 0)
	list1 := []float64{3.1, 0, 2.3, 4.5, -4.5, 1.2, 13.4}
	for _, val := range list1 {
		s := &dme.Sample{
			Value: val,
		}
		samples1 = append(samples1, s)
	}
	latency1 := CalculateStatistics(samples1)
	require.Equal(t, 0.0, latency1.Min)
	require.Equal(t, 13.4, latency1.Max)
	require.True(t, math.Abs(4.083333333333333-latency1.Avg) < errorThreshold)
	require.True(t, math.Abs(23.221666666666664-latency1.Variance) < errorThreshold)
	require.True(t, math.Abs(4.818886455050239-latency1.StdDev) < errorThreshold)
	require.Equal(t, uint64(6), latency1.NumSamples)

	// Test UpdateRollingLatency: min, max, avg, variance, stddev, numsamples calculation
	r := NewRollingStatistics()
	list2 := []float64{.53, 14.2, 21.3, -4.5, 6.7, 8.8}
	// Try adding no elements, 0, and a negative number
	r.UpdateRollingStatistics()
	r.UpdateRollingStatistics(0)
	r.UpdateRollingStatistics(-3.2)
	// Add elements one by one
	for _, elem := range list2 {
		r.UpdateRollingStatistics(elem)
	}
	require.Equal(t, 0.0, r.Statistics.Min)
	require.Equal(t, 21.3, r.Statistics.Max)
	require.True(t, math.Abs(8.5883333333333-r.Statistics.Avg) < errorThreshold)
	require.True(t, math.Abs(67.076816666667-r.Statistics.Variance) < errorThreshold)
	require.True(t, math.Abs(8.1900437524269-r.Statistics.StdDev) < errorThreshold)
	require.Equal(t, uint64(6), r.Statistics.NumSamples)

	// Test UpdateRollingLatency: Adding Samples: rolling avg, min, max, stddev, variance, numsamples
	// Update latency2 with samples1
	// Add entire list1
	r.UpdateRollingStatistics(list1...)
	require.Equal(t, 0.0, r.Statistics.Min)
	require.Equal(t, 21.3, r.Statistics.Max)
	require.True(t, math.Abs(6.335833333333333-r.Statistics.Avg) < errorThreshold)
	require.True(t, math.Abs(46.579771969696964-r.Statistics.Variance) < errorThreshold)
	require.True(t, math.Abs(6.824937506651395-r.Statistics.StdDev) < errorThreshold)
	require.Equal(t, uint64(12), r.Statistics.NumSamples)

	// Test UpdateRollingLatency: Removing Samples: rolling avg, min, max, stddev, variance, numsamples
	list3 := []float64{.34, 33.21, 11.1, 4.2, 1.5, 0, -1.2}
	r.UpdateRollingStatistics(list3...)
	require.Equal(t, 0.0, r.Statistics.Min)
	require.Equal(t, 33.21, r.Statistics.Max)
	require.True(t, math.Abs(7.021111111111112-r.Statistics.Avg) < errorThreshold)
	require.True(t, math.Abs(79.58132810457516-r.Statistics.Variance) < errorThreshold)
	require.True(t, math.Abs(8.92083673791731-r.Statistics.StdDev) < errorThreshold)
	require.Equal(t, uint64(18), r.Statistics.NumSamples)
}
