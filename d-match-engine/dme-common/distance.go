package dmecommon

import (
	"math"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

//These are the possible values returned by TDG in the non-error case.
//Based on business logic (currently just defaults), we map these to
//a location range in KM
var (
	LocationUnknown        uint32 = 0
	LocationVerifiedNear   uint32 = 1
	LocationVerifiedMedium uint32 = 2
	LocationVerifiedFar    uint32 = 3
	LocationMismatch       uint32 = 4
	LocationOtherCountry   uint32 = 5
)

type LocationResult struct {
	DistanceRange        float64
	MatchEngineLocStatus dme.Match_Engine_Loc_Verify_GPS_Location_Status
}

// it has been agreed that mappings between location result integer and distances
// in kilometers should be flexible.  These are the default mappings
var DefaultTDGLocationRangeMap = map[uint32]LocationResult{
	LocationUnknown:        {-1, dme.Match_Engine_Loc_Verify_LOC_UNKNOWN},   // unknown = negative distance (unverified)
	LocationVerifiedNear:   {2, dme.Match_Engine_Loc_Verify_LOC_VERIFIED},   // within 2km
	LocationVerifiedMedium: {10, dme.Match_Engine_Loc_Verify_LOC_VERIFIED},  // within 10km
	LocationVerifiedFar:    {100, dme.Match_Engine_Loc_Verify_LOC_VERIFIED}, //within 100km
	LocationMismatch:       {-1, dme.Match_Engine_Loc_Verify_LOC_MISMATCH},  // mismatch = negative distance (unverified)
	LocationOtherCountry:   {-1, dme.Match_Engine_Loc_Verify_LOC_MISMATCH},  // other-country = negative distance (unverified)
}

//Given a value returned by TDG API GW, map that into a distance and DME return status
func GetDistanceAndStatusForLocationResult(locationResult uint32) LocationResult {
	l, ok := DefaultTDGLocationRangeMap[locationResult]
	if !ok {
		return DefaultTDGLocationRangeMap[LocationMismatch]
	}
	return l
}

//returns gets the location range and verified distance.
func GetLocationResultForDistance(distance float64) uint32 {
	closestDistance := float64(999999)
	rc := LocationMismatch

	for l, m := range DefaultTDGLocationRangeMap {
		if m.DistanceRange >= 0 && m.DistanceRange < closestDistance && m.DistanceRange > distance {
			rc = l
			closestDistance = m.DistanceRange
		}
	}
	return rc
}

func torads(deg float64) float64 {
	return deg * math.Pi / 180
}

// Use the ‘haversine’ formula to calculate the great-circle distance between two points
func DistanceBetween(loc1, loc2 dme.Loc) float64 {
	radiusofearth := 6371
	var diff_lat, diff_long float64
	var a, c, dist float64
	var lat1, long1, lat2, long2 float64

	lat1 = loc1.Lat
	long1 = loc1.Long
	lat2 = loc2.Lat
	long2 = loc2.Long

	diff_lat = torads(lat2 - lat1)
	diff_long = torads(long2 - long1)

	rad_lat1 := torads(lat1)
	rad_lat2 := torads(lat2)

	a = math.Sin(diff_lat/2)*math.Sin(diff_lat/2) + math.Sin(diff_long/2)*
		math.Sin(diff_long/2)*math.Cos(rad_lat1)*math.Cos(rad_lat2)
	c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	dist = c * float64(radiusofearth)

	return dist
}
