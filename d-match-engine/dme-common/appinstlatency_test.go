package dmecommon

import (
	"math"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/stretchr/testify/require"
)

var errorThreshold = .00001

func TestLatencyCalculations(t *testing.T) {
	// Expected values confirmed using Average and Standard Deviation Calculators
	now := time.Now()

	// Test CalculateLatency: min, max, avg, variance, stddev, numsamples calculation
	samples1 := make([]*dme.Sample, 0)
	list1 := []float64{3.1, 2.3, 4.5, 1.2, 13.4}
	times1 := []time.Time{now.Add(-55 * time.Second), now.Add(-58 * time.Second), now.Add(-21 * time.Second), now.Add(-30 * time.Second), now.Add(-52 * time.Second)}
	for i, val := range list1 {
		ts := cloudcommon.TimeToTimestamp(times1[i])
		s := &dme.Sample{
			Value:     val,
			Timestamp: &ts,
		}
		samples1 = append(samples1, s)
	}
	latency1 := CalculateLatency(samples1)
	require.Equal(t, 1.2, latency1.Min)
	require.Equal(t, 13.4, latency1.Max)
	require.True(t, math.Abs(4.9-latency1.Avg) < errorThreshold)
	require.True(t, math.Abs(24.025-latency1.Variance) < errorThreshold)
	require.True(t, math.Abs(4.901530373261-latency1.StdDev) < errorThreshold)
	require.Equal(t, uint64(5), latency1.NumSamples)

	// Test UpdateRollingLatency: min, max, avg, variance, stddev, numsamples calculation
	now = time.Now()
	r := NewRollingLatency()
	list2 := []float64{.53, 14.2, 21.3, 6.7, 8.8}
	client1 := CookieKey{
		OrgName:  "testorg",
		AppName:  "testapp",
		AppVers:  "1",
		UniqueId: "123",
	}
	// Add elements one by one
	for _, elem := range list2 {
		r.UpdateRollingLatency(client1, elem)
	}
	require.Equal(t, .53, r.Latency.Min)
	require.Equal(t, 21.3, r.Latency.Max)
	require.True(t, math.Abs(10.306-r.Latency.Avg) < errorThreshold)
	require.True(t, math.Abs(61.71818-r.Latency.Variance) < errorThreshold)
	require.True(t, math.Abs(7.8560919037394-r.Latency.StdDev) < errorThreshold)
	require.Equal(t, uint64(5), r.Latency.NumSamples)
	require.Equal(t, uint64(1), r.NumUniqueClients)

	// Test UpdateRollingLatency: Adding Samples: rolling avg, min, max, stddev, variance, numsamples
	// Update latency2 with samples1
	client2 := CookieKey{
		OrgName:  "testorg",
		AppName:  "testapp",
		AppVers:  "1",
		UniqueId: "234",
	}
	// Add entire list1
	r.UpdateRollingLatency(client2, list1...)
	require.Equal(t, .53, r.Latency.Min)
	require.Equal(t, 21.3, r.Latency.Max)
	require.True(t, math.Abs(7.603-r.Latency.Avg) < errorThreshold)
	require.True(t, math.Abs(46.22609-r.Latency.Variance) < errorThreshold)
	require.True(t, math.Abs(6.7989771289511-r.Latency.StdDev) < errorThreshold)
	require.Equal(t, uint64(10), r.Latency.NumSamples)
	require.Equal(t, uint64(2), r.NumUniqueClients)

	// Test UpdateRollingLatency: Removing Samples: rolling avg, min, max, stddev, variance, numsamples
	time.Sleep(time.Second * 10)
	now = time.Now()
	list3 := []float64{.34, 33.21, 11.1, 4.2, 1.5}
	client3 := CookieKey{
		OrgName:  "testorg",
		AppName:  "testapp",
		AppVers:  "1",
		UniqueId: "345",
	}
	r.UpdateRollingLatency(client3, list3...)
	require.Equal(t, .34, r.Latency.Min)
	require.Equal(t, 33.21, r.Latency.Max)
	require.True(t, math.Abs(8.4253333333333-r.Latency.Avg) < errorThreshold)
	require.True(t, math.Abs(83.958355238095-r.Latency.Variance) < errorThreshold)
	require.True(t, math.Abs(9.1628792002348-r.Latency.StdDev) < errorThreshold)
	require.Equal(t, uint64(15), r.Latency.NumSamples)
	require.Equal(t, uint64(3), r.NumUniqueClients)
}
