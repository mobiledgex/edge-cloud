package grpcstats

import (
	"math"
	"testing"
	"time"

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
	require.Equal(t, 1.2, latency1.Min)
	require.Equal(t, 13.4, latency1.Max)
	require.True(t, math.Abs(4.9-latency1.Avg) < errorThreshold)
	require.True(t, math.Abs(24.025-latency1.Variance) < errorThreshold)
	require.True(t, math.Abs(4.901530373261-latency1.StdDev) < errorThreshold)
	require.Equal(t, uint64(5), latency1.NumSamples)

	// Test UpdateRollingLatency: min, max, avg, variance, stddev, numsamples calculation
	r := NewRollingStatistics()
	list2 := []float64{.53, 14.2, 21.3, 0, -4.5, 6.7, 8.8}
	// Try adding no elements, 0, and a negative number
	r.UpdateRollingStatistics()
	r.UpdateRollingStatistics(0)
	r.UpdateRollingStatistics(-3.2)
	// Add elements one by one
	for _, elem := range list2 {
		r.UpdateRollingStatistics(elem)
	}
	require.Equal(t, .53, r.Statistics.Min)
	require.Equal(t, 21.3, r.Statistics.Max)
	require.True(t, math.Abs(10.306-r.Statistics.Avg) < errorThreshold)
	require.True(t, math.Abs(61.71818-r.Statistics.Variance) < errorThreshold)
	require.True(t, math.Abs(7.8560919037394-r.Statistics.StdDev) < errorThreshold)
	require.Equal(t, uint64(5), r.Statistics.NumSamples)

	// Test UpdateRollingLatency: Adding Samples: rolling avg, min, max, stddev, variance, numsamples
	// Update latency2 with samples1
	// Add entire list1
	r.UpdateRollingStatistics(list1...)
	require.Equal(t, .53, r.Statistics.Min)
	require.Equal(t, 21.3, r.Statistics.Max)
	require.True(t, math.Abs(7.603-r.Statistics.Avg) < errorThreshold)
	require.True(t, math.Abs(46.22609-r.Statistics.Variance) < errorThreshold)
	require.True(t, math.Abs(6.7989771289511-r.Statistics.StdDev) < errorThreshold)
	require.Equal(t, uint64(10), r.Statistics.NumSamples)

	// Test UpdateRollingLatency: Removing Samples: rolling avg, min, max, stddev, variance, numsamples
	time.Sleep(time.Second * 10)
	list3 := []float64{.34, 33.21, 11.1, 4.2, 1.5, 0, -1.2}
	r.UpdateRollingStatistics(list3...)
	require.Equal(t, .34, r.Statistics.Min)
	require.Equal(t, 33.21, r.Statistics.Max)
	require.True(t, math.Abs(8.4253333333333-r.Statistics.Avg) < errorThreshold)
	require.True(t, math.Abs(83.958355238095-r.Statistics.Variance) < errorThreshold)
	require.True(t, math.Abs(9.1628792002348-r.Statistics.StdDev) < errorThreshold)
	require.Equal(t, uint64(15), r.Statistics.NumSamples)
}
