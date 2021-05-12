package dmecommon

import (
	"fmt"
	"math"
	"strconv"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

// These conversions are at equator
const kmPerDegLong = 111.32
const kmPerDegLat = 110.57

type LocationTileInfo struct {
	Quadrant   int
	LongIndex  int
	LatIndex   int
	TileLength int
}

func GetLocationTileFromGpsLocation(loc *dme.Loc, locationTileSideLengthKm int) string {
	// Validate location
	err := ValidateLocation(loc)
	if err != nil {
		return ""
	}
	// Create LocationTileInfo
	locationTileInfo := &LocationTileInfo{}
	locationTileInfo.TileLength = locationTileSideLengthKm
	if loc.Latitude >= 0 && loc.Longitude >= 0 {
		locationTileInfo.Quadrant = 1
	} else if loc.Latitude >= 0 && loc.Longitude < 0 {
		locationTileInfo.Quadrant = 2
	} else if loc.Latitude < 0 && loc.Longitude < 0 {
		locationTileInfo.Quadrant = 3
	} else {
		locationTileInfo.Quadrant = 4
	}
	// Convert lat, long to positive
	lat := math.Abs(loc.Latitude)
	long := math.Abs(loc.Longitude)
	// Calculate km away from (0, 0) for lat and long
	latKm := lat * kmPerDegLat
	longKm := long * kmPerDegLong
	// Calculate number of tiles to get to loc (round down)
	locationTileInfo.LatIndex = int(latKm / float64(locationTileSideLengthKm))
	locationTileInfo.LongIndex = int(longKm / float64(locationTileSideLengthKm))
	// Get the lower and upper gps location ranges
	locUnder, locOver, err := getGpsLocationRangeFromLocationInfo(locationTileInfo)
	if err != nil {
		return ""
	}
	// Append long,lat for both locUnder and locOver to string for metrics
	locUnderStr := fmt.Sprintf("%f,%f", locUnder.Longitude, locUnder.Latitude)
	locOverStr := fmt.Sprintf("%f,%f", locOver.Longitude, locOver.Latitude)
	// Append location tile side length for easy conversion
	sideLengthStr := strconv.Itoa(locationTileSideLengthKm)
	return locUnderStr + "_" + locOverStr + "_" + sideLengthStr
}

func getGpsLocationRangeFromLocationInfo(locationTileInfo *LocationTileInfo) (under *dme.Loc, over *dme.Loc, err error) {
	quadrant := locationTileInfo.Quadrant
	latIndex := locationTileInfo.LatIndex
	longIndex := locationTileInfo.LongIndex
	locationTileSideLengthKm := locationTileInfo.TileLength

	latKmUnder := latIndex * locationTileSideLengthKm
	latKmOver := (latIndex + 1) * locationTileSideLengthKm
	longKmUnder := longIndex * locationTileSideLengthKm
	longKmOver := (longIndex + 1) * locationTileSideLengthKm

	latUnder := float64(latKmUnder) / float64(kmPerDegLat)
	latOver := float64(latKmOver) / float64(kmPerDegLat)
	longUnder := float64(longKmUnder) / float64(kmPerDegLong)
	longOver := float64(longKmOver) / float64(kmPerDegLong)

	under = &dme.Loc{
		Latitude:  latUnder,
		Longitude: longUnder,
	}
	over = &dme.Loc{
		Latitude:  latOver,
		Longitude: longOver,
	}
	switch quadrant {
	case 1:
		// no-op
	case 2:
		under.Longitude *= -1
		over.Longitude *= -1
	case 3:
		under.Latitude *= -1
		under.Longitude *= -1
		over.Latitude *= -1
		over.Longitude *= -1
	case 4:
		under.Latitude *= -1
		over.Latitude *= -1
	default:
		return nil, nil, fmt.Errorf("Invalid quadrant for location tile")
	}
	return under, over, nil
}
