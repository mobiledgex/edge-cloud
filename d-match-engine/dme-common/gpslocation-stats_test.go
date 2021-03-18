package dmecommon

import (
	"math"
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
	// Get Gps Location ranges for tile generated above
	under, over, err := GetGpsLocationRangeFromLocationTile(tile1Beacon, tl1)
	assert.Nil(t, err)
	// Check to make sure lat, long of beacon is within the location ranges generated above
	checkLocationRange(t, under, over, beacon)

	// Get location tile for Beacon with location tile length == 2km
	tile2Beacon := GetLocationTileFromGpsLocation(beacon, tl2)
	under, over, err = GetGpsLocationRangeFromLocationTile(tile2Beacon, tl2)
	assert.Nil(t, err)
	checkLocationRange(t, under, over, beacon)

	// Get location tile for Beacon with location tile length == 5km
	tile5Beacon := GetLocationTileFromGpsLocation(beacon, tl5)
	under, over, err = GetGpsLocationRangeFromLocationTile(tile5Beacon, tl5)
	assert.Nil(t, err)
	checkLocationRange(t, under, over, beacon)

	tile1SanJose := GetLocationTileFromGpsLocation(sanjose, tl1)
	under, over, err = GetGpsLocationRangeFromLocationTile(tile1SanJose, tl1)
	assert.Nil(t, err)
	checkLocationRange(t, under, over, sanjose)

	tile2SanJose := GetLocationTileFromGpsLocation(sanjose, tl2)
	under, over, err = GetGpsLocationRangeFromLocationTile(tile2SanJose, tl2)
	assert.Nil(t, err)
	checkLocationRange(t, under, over, sanjose)

	tile5SanJose := GetLocationTileFromGpsLocation(sanjose, tl5)
	under, over, err = GetGpsLocationRangeFromLocationTile(tile5SanJose, tl5)
	assert.Nil(t, err)
	checkLocationRange(t, under, over, sanjose)

	tile1Sydney := GetLocationTileFromGpsLocation(sydney, tl1)
	under, over, err = GetGpsLocationRangeFromLocationTile(tile1Sydney, tl1)
	assert.Nil(t, err)
	checkLocationRange(t, under, over, sydney)

	tile2Sydney := GetLocationTileFromGpsLocation(sydney, tl2)
	under, over, err = GetGpsLocationRangeFromLocationTile(tile2Sydney, tl2)
	assert.Nil(t, err)
	checkLocationRange(t, under, over, sydney)

	tile5Sydney := GetLocationTileFromGpsLocation(sydney, tl5)
	under, over, err = GetGpsLocationRangeFromLocationTile(tile5Sydney, tl5)
	assert.Nil(t, err)
	checkLocationRange(t, under, over, sydney)

	tile1Rio := GetLocationTileFromGpsLocation(rio, tl1)
	under, over, err = GetGpsLocationRangeFromLocationTile(tile1Rio, tl1)
	assert.Nil(t, err)
	checkLocationRange(t, under, over, rio)

	tile2Rio := GetLocationTileFromGpsLocation(rio, tl2)
	under, over, err = GetGpsLocationRangeFromLocationTile(tile2Rio, tl2)
	assert.Nil(t, err)
	checkLocationRange(t, under, over, rio)

	tile5Rio := GetLocationTileFromGpsLocation(rio, tl5)
	under, over, err = GetGpsLocationRangeFromLocationTile(tile5Rio, tl5)
	assert.Nil(t, err)
	checkLocationRange(t, under, over, rio)
}

func checkLocationRange(t *testing.T, under, over, actual *dme.Loc) {
	assert.True(t, math.Abs(actual.Latitude) < math.Abs(over.Latitude))
	assert.True(t, math.Abs(actual.Latitude) >= math.Abs(under.Latitude))

	assert.True(t, math.Abs(actual.Longitude) < math.Abs(over.Longitude))
	assert.True(t, math.Abs(actual.Longitude) >= math.Abs(under.Longitude))
}
