package dmecommon

import (
	"math"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

//These are the possible values returned by GDDT in the non-error case.
//Based on business logic (currently just defaults), we map these to
//a location range in KM
var (
	LocationUnknown                uint32 = 0
	LocationVerifiedNear           uint32 = 1
	LocationVerifiedMedium         uint32 = 2
	LocationVerifiedFar            uint32 = 3
	LocationMismatchSameCountry    uint32 = 4
	LocationMismatchOtherCountry   uint32 = 5
	LocationRoamingCountryMatch    uint32 = 6
	LocationRoamingCountryMismatch uint32 = 7
)

// InfiniteDistance is a value higher than the circumference of the globe,
// any cloudlet found would have to be closer than this.
var InfiniteDistance float64 = 100000

type LocationResult struct {
	DistanceRange        float64
	MatchEngineLocStatus dme.VerifyLocationReply_GPSLocationStatus
}

// it has been agreed that mappings between location result integer and distances
// in kilometers should be flexible.  These are the default mappings
var DefaultLocationRangeMap = map[uint32]LocationResult{
	LocationUnknown:                {-1, dme.VerifyLocationReply_LOC_UNKNOWN},                  // unknown = negative distance (unverified)
	LocationVerifiedNear:           {2, dme.VerifyLocationReply_LOC_VERIFIED},                  // within 2km
	LocationVerifiedMedium:         {10, dme.VerifyLocationReply_LOC_VERIFIED},                 // within 10km
	LocationVerifiedFar:            {100, dme.VerifyLocationReply_LOC_VERIFIED},                // within 100km
	LocationMismatchSameCountry:    {-1, dme.VerifyLocationReply_LOC_MISMATCH_SAME_COUNTRY},    // mismatch = negative distance (unverified)
	LocationMismatchOtherCountry:   {-1, dme.VerifyLocationReply_LOC_MISMATCH_OTHER_COUNTRY},   // mismatch, wrong country = negative distance (unverified)
	LocationRoamingCountryMatch:    {-1, dme.VerifyLocationReply_LOC_ROAMING_COUNTRY_MATCH},    // roamer in correct country = negative distance (unverified)
	LocationRoamingCountryMismatch: {-1, dme.VerifyLocationReply_LOC_ROAMING_COUNTRY_MISMATCH}, // roamer in wrong country = negative distance (unverified)

}

// GetDistanceAndStatusForLocationResult - Given a value returned by GDDT API GW, map that into a
// distance and DME return status.
func GetDistanceAndStatusForLocationResult(locationResult uint32) LocationResult {
	l, ok := DefaultLocationRangeMap[locationResult]
	if !ok {
		return DefaultLocationRangeMap[LocationUnknown]
	}
	return l
}

// GetLocationResultForDistance - given a distance, converts that into a location result.
// This function has no knowledge of country, it is not used when we have access to
// the real Location Verification API Gateway.
func GetLocationResultForDistance(distance float64) uint32 {
	closestDistance := float64(999999)
	rc := LocationMismatchSameCountry

	for l, m := range DefaultLocationRangeMap {
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

func todegs(rad float64) float64 {
	return rad * 180 / math.Pi
}

// Return 0 - 360
func todegsUnitCircle(rad float64) float64 {
	deg := todegs(rad)
	if deg < 0 {
		return math.Mod(deg, 360.0) + 360
	}
	return math.Mod(deg, 360.0)
}

// Use the ‘haversine’ formula to calculate the great-circle distance between two points
func DistanceBetween(loc1, loc2 dme.Loc) float64 {
	radiusofearth := 6371
	var diff_lat, diff_long float64
	var a, c, dist float64
	var lat1, long1, lat2, long2 float64

	lat1 = loc1.Latitude
	long1 = loc1.Longitude
	lat2 = loc2.Latitude
	long2 = loc2.Longitude

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
