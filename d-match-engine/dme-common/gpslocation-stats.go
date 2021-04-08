package dmecommon

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

// These conversions are at equator
const kmPerDegLong = 111.32
const kmPerDegLat = 110.57

func GetLocationTileFromGpsLocation(loc *dme.Loc, locationTileSideLengthKm int) string {
	var quadrant string
	if loc == nil {
		return fmt.Sprintf("0-0,0-%s", strconv.Itoa(locationTileSideLengthKm))
	}
	if loc.Latitude >= 0 && loc.Longitude >= 0 {
		quadrant = "1"
	} else if loc.Latitude >= 0 && loc.Longitude < 0 {
		quadrant = "2"
	} else if loc.Latitude < 0 && loc.Longitude < 0 {
		quadrant = "3"
	} else {
		quadrant = "4"
	}
	// Convert lat, long to positive
	lat := math.Abs(loc.Latitude)
	long := math.Abs(loc.Longitude)
	// Calculate km away from (0, 0) for lat and long
	latKm := lat * kmPerDegLat
	longKm := long * kmPerDegLong
	// Calculate number of tiles to get to loc (round down)
	latIndex := int(latKm / float64(locationTileSideLengthKm))
	longIndex := int(longKm / float64(locationTileSideLengthKm))
	// Append lat long to string for metrics
	pair := strconv.Itoa(latIndex) + "," + strconv.Itoa(longIndex)
	// Append location tile side length for easy conversion
	sideLengthStr := strconv.Itoa(locationTileSideLengthKm)
	return quadrant + "-" + pair + "-" + sideLengthStr
}

func GetGpsLocationRangeFromLocationTile(locationTile string) (under *dme.Loc, over *dme.Loc, err error) {
	s := strings.Split(locationTile, "-")
	indeces := strings.Split(s[1], ",")
	quadrant := s[0]
	latIndex, err := strconv.Atoi(indeces[0])
	if err != nil {
		return nil, nil, err
	}
	longIndex, err := strconv.Atoi(indeces[1])
	if err != nil {
		return nil, nil, err
	}
	locationTileSideLengthKm, err := strconv.Atoi(s[2])
	if err != nil {
		return nil, nil, err
	}

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
	case "1":
		// no-op
	case "2":
		under.Longitude *= -1
		over.Longitude *= -1
	case "3":
		under.Latitude *= -1
		under.Longitude *= -1
		over.Latitude *= -1
		over.Longitude *= -1
	case "4":
		under.Latitude *= -1
		over.Latitude *= -1
	default:
		return nil, nil, fmt.Errorf("Invalid quadrant for location tile")
	}
	return under, over, nil
}
