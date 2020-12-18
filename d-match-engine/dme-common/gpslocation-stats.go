package dmecommon

import (
	"math"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Filled in by DME. Added to EdgeEventStatCall to update stats
type GpsLocationInfo struct {
	GpsLocation      *dme.Loc
	SessionCookieKey *CookieKey
	Timestamp        dme.Timestamp
	Carrier          string
	DeviceOs         string
	DeviceModel      string
}

type GpsLocationStats struct {
	Stats     []*GpsLocationInfo
	StartTime time.Time
}

// Distances from Device to AppInst in km
var DistsFromDeviceToAppInst = []int{
	0,
	1,
	5,
	10,
	50,
	100,
}

type GpsLocationBearing string

const (
	Northeast = "Northeast"
	Southeast = "Southeast"
	Southwest = "Southwest"
	Northwest = "Northwest"
)

func NewGpsLocationStats() *GpsLocationStats {
	g := new(GpsLocationStats)
	g.Stats = make([]*GpsLocationInfo, 0)
	g.StartTime = time.Now()
	return g
}

func RecordGpsLocationStatCall(loc *dme.Loc, appInstKey *edgeproto.AppInstKey, sessionCookieKey *CookieKey, carrier string, deviceInfo dme.DeviceInfo) {
	if EEStats == nil {
		return
	}
	call := EdgeEventStatCall{}
	call.Key.AppInstKey = *appInstKey
	call.Key.Metric = cloudcommon.GpsLocationMetric
	call.GpsLocationInfo = &GpsLocationInfo{
		GpsLocation:      loc,
		SessionCookieKey: sessionCookieKey,
		Timestamp:        cloudcommon.TimeToTimestamp(time.Now()),
		Carrier:          translateCarrierName(carrier),
		DeviceOs:         deviceInfo.DeviceOs,
		DeviceModel:      deviceInfo.DeviceModel,
	}
	EEStats.RecordEdgeEventStatCall(&call)
}

func GetDistanceBucketFromAppInst(appinst *DmeAppInst, loc dme.Loc) int {
	return GetDistanceBucket(appinst.location, loc)
}

func GetDistanceBucket(loc1, loc2 dme.Loc) int {
	dist := DistanceBetween(loc1, loc2)
	i := 0
	bucket := 0
	for i, bucket = range DistsFromDeviceToAppInst {
		if dist < float64(bucket) {
			i--
			break
		}
	}
	return DistsFromDeviceToAppInst[i]
}

func GetBearingFromAppInst(appinst *DmeAppInst, loc dme.Loc) GpsLocationBearing {
	return GetBearingFrom(appinst.location, loc)
}

// Gets Bearing in degrees from loc1 to loc2
// 0 degrees is North and move clockwise (ie. 90 degrees is East)
// Formula provided: https://www.movable-type.co.uk/scripts/latlong.html
func GetDegreesBearingFrom(loc1, loc2 dme.Loc) float64 {
	long1 := torads(loc1.Longitude)
	lat1 := torads(loc1.Latitude)
	long2 := torads(loc2.Longitude)
	lat2 := torads(loc2.Latitude)
	y := math.Sin(long2-long1) * math.Cos(lat2)
	x := (math.Cos(lat1) * math.Sin(lat2)) - (math.Sin(lat1) * math.Cos(lat2) * math.Cos(long2-long1))
	theta := math.Atan2(y, x)
	deg := todegsUnitCircle(theta)
	return deg
}

// Determines the orientation (NW, NE, SW, or SE) that loc2 is in relation to loc1
func GetBearingFrom(loc1, loc2 dme.Loc) GpsLocationBearing {
	deg := GetDegreesBearingFrom(loc1, loc2)
	// North == 0 degrees, bearings move clockwise
	if deg >= 0 && deg < 90 {
		return Northeast
	} else if deg >= 90 && deg < 180 {
		return Southeast
	} else if deg >= 180 && deg < 270 {
		return Southwest
	} else {
		return Northwest
	}
}
