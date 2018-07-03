package dmecommon

import (
	"math"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

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
