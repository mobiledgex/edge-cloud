package dmecommon

import (
	"math"
	"strconv"
	"strings"
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/stretchr/testify/assert"
)

func TestGpsLocation(t *testing.T) {
	// Beacon (Quadrant I)
	beacon := &dme.Loc{
		Latitude:  52.5200,
		Longitude: 13.4050,
	}
	// San Jose (Quadrant II)
	sanjose := &dme.Loc{
		Latitude:  37.3382,
		Longitude: -121.8863,
	}
	// Sydney (Quadrant III)
	sydney := &dme.Loc{
		Latitude:  -33.8688,
		Longitude: 151.2093,
	}
	// Rio De Janeiro (Quadrant IV)
	rio := &dme.Loc{
		Latitude:  -22.9068,
		Longitude: -43.1729,
	}

	// tile length == 1 km
	tl1 := 1
	// tile length == 2 km
	tl2 := 2
	// tile length == 5 km
	tl5 := 5

	// Get location tile for Beacon with location tile length == 1km
	tile1Beacon := GetLocationTileFromGpsLocation(beacon, tl1)
	// Check to make sure lat, long of beacon is within the location ranges generated above
	checkLocationRange(t, tile1Beacon, beacon)

	// Get location tile for Beacon with location tile length == 2km
	tile2Beacon := GetLocationTileFromGpsLocation(beacon, tl2)
	checkLocationRange(t, tile2Beacon, beacon)

	// Get location tile for Beacon with location tile length == 5km
	tile5Beacon := GetLocationTileFromGpsLocation(beacon, tl5)
	checkLocationRange(t, tile5Beacon, beacon)

	tile1SanJose := GetLocationTileFromGpsLocation(sanjose, tl1)
	checkLocationRange(t, tile1SanJose, sanjose)

	tile2SanJose := GetLocationTileFromGpsLocation(sanjose, tl2)
	checkLocationRange(t, tile2SanJose, sanjose)

	tile5SanJose := GetLocationTileFromGpsLocation(sanjose, tl5)
	checkLocationRange(t, tile5SanJose, sanjose)

	tile1Sydney := GetLocationTileFromGpsLocation(sydney, tl1)
	checkLocationRange(t, tile1Sydney, sydney)

	tile2Sydney := GetLocationTileFromGpsLocation(sydney, tl2)
	checkLocationRange(t, tile2Sydney, sydney)

	tile5Sydney := GetLocationTileFromGpsLocation(sydney, tl5)
	checkLocationRange(t, tile5Sydney, sydney)

	tile1Rio := GetLocationTileFromGpsLocation(rio, tl1)
	checkLocationRange(t, tile1Rio, rio)

	tile2Rio := GetLocationTileFromGpsLocation(rio, tl2)
	checkLocationRange(t, tile2Rio, rio)

	tile5Rio := GetLocationTileFromGpsLocation(rio, tl5)
	checkLocationRange(t, tile5Rio, rio)
}

func checkLocationRange(t *testing.T, locationTile string, actual *dme.Loc) {
	s := strings.Split(locationTile, "_")
	// Pull out underLoc values
	underStr := s[0]
	underLongLat := strings.Split(underStr, ",")
	underLong, err := strconv.ParseFloat(underLongLat[0], 64)
	assert.Nil(t, err)
	underLat, err := strconv.ParseFloat(underLongLat[1], 64)
	assert.Nil(t, err)
	under := &dme.Loc{
		Longitude: underLong,
		Latitude:  underLat,
	}
	// Pull out overLoc values
	overStr := s[1]
	overLongLat := strings.Split(overStr, ",")
	overLong, err := strconv.ParseFloat(overLongLat[0], 64)
	assert.Nil(t, err)
	overLat, err := strconv.ParseFloat(overLongLat[1], 64)
	assert.Nil(t, err)
	over := &dme.Loc{
		Longitude: overLong,
		Latitude:  overLat,
	}

	assert.True(t, math.Abs(actual.Latitude) < math.Abs(over.Latitude))
	assert.True(t, math.Abs(actual.Latitude) >= math.Abs(under.Latitude))

	assert.True(t, math.Abs(actual.Longitude) < math.Abs(over.Longitude))
	assert.True(t, math.Abs(actual.Longitude) >= math.Abs(under.Longitude))
}
