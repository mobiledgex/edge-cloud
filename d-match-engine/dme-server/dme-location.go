package main

import (
	"fmt"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	locapi "github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
)

func VerifyClientLoc(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply, carrier string, ckey *dmecommon.CookieKey, locVerUrl string) error {
	var key carrierAppKey
	var found *carrierAppInst
	var app *carrierApp
	var distance, d float64
	var tbl *carrierApps

	tbl = carrierAppTbl
	key.carrierName = mreq.CarrierName
	key.appKey.DeveloperKey.Name = ckey.DevName
	key.appKey.Name = ckey.AppName
	key.appKey.Version = ckey.AppVers

	mreply.GpsLocationStatus = dme.VerifyLocationReply_LOC_UNKNOWN
	mreply.GPS_Location_Accuracy_KM = -1

	if mreq.GpsLocation == nil {
		log.DebugLog(log.DebugLevelDmereq, "Invalid VerifyLocation request", "Error", "Missing GpsLocation")
		return fmt.Errorf("Missing GpsLocation")
	}

	log.DebugLog(log.DebugLevelDmereq, "Received Verify Location",
		"appName", key.appKey.Name,
		"appVersion", key.appKey.Version,
		"carrier", key.carrierName,
		"devName", key.appKey.DeveloperKey.Name,
		"lat", mreq.GpsLocation.Lat,
		"long", mreq.GpsLocation.Long)

	tbl.RLock()
	app, ok := tbl.apps[key]
	if !ok {
		tbl.RUnlock()
		log.DebugLog(log.DebugLevelDmereq, "Could not find key in app table", "key", key)
		// return loc unknown
		return nil
	}

	//handling for each carrier may be different.  As of now there is only standalone and TDG
	switch carrier {
	case "tdg":
		fallthrough
	case "TDG":
		result := locapi.CallTDGLocationVerifyAPI(locVerUrl, mreq.GpsLocation.Lat, mreq.GpsLocation.Long, mreq.VerifyLocToken)
		mreply.GpsLocationStatus = result.MatchEngineLocStatus
		mreply.GPS_Location_Accuracy_KM = result.DistanceRange
	default:
		distance = dmecommon.InfiniteDistance
		log.DebugLog(log.DebugLevelDmereq, ">>>Verify Location",
			"appName", key.appKey.Name,
			"carrier", key.carrierName,
			"lat", mreq.GpsLocation.Lat,
			"long", mreq.GpsLocation.Long)
		for _, c := range app.insts {
			d = dmecommon.DistanceBetween(*mreq.GpsLocation, c.location)
			log.DebugLog(log.DebugLevelDmereq, "verify location at",
				"lat", c.location.Lat,
				"long", c.location.Long,
				"distance", distance,
				"this-dist", d)
			if d < distance {
				distance = d
				found = c
			}
		}
		//here we want the verified range of the distance based on the distance rules
		//e.g. if the actual distance is 50KM we may return a range to say within 100KM
		locresult := dmecommon.GetLocationResultForDistance(distance)
		locval := dmecommon.GetDistanceAndStatusForLocationResult(locresult)

		mreply.GpsLocationStatus = locval.MatchEngineLocStatus
		mreply.GPS_Location_Accuracy_KM = locval.DistanceRange

		log.DebugLog(log.DebugLevelDmereq, "verified location at",
			"lat", found.location.Lat,
			"long", found.location.Long,
			"actual distance", distance,
			"distance range", mreply.GPS_Location_Accuracy_KM,
			"status", mreply.GpsLocationStatus,
			"uri", found.uri)

	}

	tbl.RUnlock()
	return nil
}

func GetClientLoc(mreq *dme.GetLocationRequest, reply *dme.GetLocationReply) {
	reply.CarrierName = mreq.CarrierName
	reply.Status = dme.GetLocationReply_LOC_FOUND
	reply.NetworkLocation = &dme.Loc{}
}
