package main

import (
	"fmt"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	locapi "github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
)

func VerifyClientLoc(mreq *dme.Match_Engine_Request, mreply *dme.Match_Engine_Loc_Verify, carrier string, peerIp string, locVerUrl string) error {
	var key carrierAppKey
	var found *carrierAppInst
	var app *carrierApp
	var distance, d float64
	var tbl *carrierApps

	tbl = carrierAppTbl
	key.carrierName = mreq.CarrierName
	key.appKey.DeveloperKey.Name = mreq.DevName
	key.appKey.Name = mreq.AppName
	key.appKey.Version = mreq.AppVers

	mreply.GpsLocationStatus = dme.Match_Engine_Loc_Verify_LOC_UNKNOWN
	mreply.GPS_Location_Accuracy_KM = -1

	if mreq.GpsLocation == nil {
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
		log.InfoLog("Could not find key in app table", "key", key)
		return fmt.Errorf("app not found")
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
		distance = 10000
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

func GetClientLoc(mreq *dme.Match_Engine_Request, mloc *dme.Match_Engine_Loc) {
	mloc.CarrierName = mreq.CarrierName
	mloc.Status = dme.Match_Engine_Loc_LOC_FOUND
	mloc.NetworkLocation = mreq.GpsLocation
}
