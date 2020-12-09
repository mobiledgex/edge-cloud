package dmecommon

import (
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/stretchr/testify/require"
)

func TestGpsLocation(t *testing.T) {
	// San Jose
	sanjose := dme.Loc{
		Latitude:  37.3382,
		Longitude: -121.8863,
	}
	// San Francisco
	sanfran := dme.Loc{
		Latitude:  37.7749,
		Longitude: -122.4194,
	}
	// Palo Alto
	paloalto := dme.Loc{
		Latitude:  37.4419,
		Longitude: -122.1430,
	}
	// Anchorage
	anchorage := dme.Loc{
		Latitude:  61.2181,
		Longitude: -149.9003,
	}
	// Austin
	austin := dme.Loc{
		Latitude:  30.2672,
		Longitude: -97.7431,
	}
	// New York city
	newyork := dme.Loc{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}

	// Test Distance buckets
	dist1 := GetDistanceBucket(sanjose, sanfran)
	require.Equal(t, 50, dist1)

	dist2 := GetDistanceBucket(sanjose, paloalto)
	require.Equal(t, 10, dist2)

	dist3 := GetDistanceBucket(anchorage, austin)
	require.Equal(t, 100, dist3)

	// Test Bearings
	bearing1 := GetBearingFrom(sanjose, sanfran)
	require.Equal(t, "Northwest", string(bearing1))

	bearing2 := GetBearingFrom(sanfran, sanjose)
	require.Equal(t, "Southeast", string(bearing2))

	bearing3 := GetBearingFrom(newyork, austin)
	require.Equal(t, "Southwest", string(bearing3))

	bearing4 := GetBearingFrom(austin, newyork)
	require.Equal(t, "Northeast", string(bearing4))
}
