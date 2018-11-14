package main

import (
	"fmt"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	locapi "github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func VerifyClientLoc(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply, carrier string, ckey *dmecommon.CookieKey, locVerUrl string, tokSrvUrl string) error {
	var key edgeproto.AppKey
	var found *dmeAppInst
	var app *dmeApp
	var distance, d float64
	var tbl *dmeApps

	tbl = dmeAppTbl
	key.DeveloperKey.Name = ckey.DevName
	key.Name = ckey.AppName
	key.Version = ckey.AppVers

	mreply.GpsLocationStatus = dme.VerifyLocationReply_LOC_UNKNOWN
	mreply.GPS_Location_Accuracy_KM = -1

	log.DebugLog(log.DebugLevelDmereq, "Received Verify Location",
		"appName", key.Name,
		"appVersion", key.Version,
		"devName", key.DeveloperKey.Name,
		"VerifyLocToken", mreq.VerifyLocToken,
		"GpsLocation", mreq.GpsLocation)

	if mreq.GpsLocation == nil || (mreq.GpsLocation.Lat == 0 && mreq.GpsLocation.Long == 0) {
		log.DebugLog(log.DebugLevelDmereq, "Invalid VerifyLocation request", "Error", "Missing GpsLocation")
		return fmt.Errorf("Missing GpsLocation")
	}

	tbl.RLock()
	defer tbl.RUnlock()

	app, ok := tbl.apps[key]
	if !ok {
		log.DebugLog(log.DebugLevelDmereq, "Could not find key in app table", "key", key)
		// return loc unknown
		return fmt.Errorf("app not found: %s", key)
	}

	//handling for each carrier may be different.  As of now there is only standalone and TDG
	switch carrier {
	case "tdg":
		fallthrough
	case "TDG":
		if mreq.VerifyLocToken == "" {
			return fmt.Errorf("verifyloc token required")
		}
		result := locapi.CallTDGLocationVerifyAPI(locVerUrl, mreq.GpsLocation.Lat, mreq.GpsLocation.Long, mreq.VerifyLocToken, tokSrvUrl)
		mreply.GpsLocationStatus = result.MatchEngineLocStatus
		mreply.GPS_Location_Accuracy_KM = result.DistanceRange
	default:
		// non-API based location uses cloudlets and so default and public cloudlets are not applicable
		carr, ok := app.carriers[mreq.CarrierName]
		if !ok {
			log.DebugLog(log.DebugLevelDmereq, "Could not find carrier for app", "appKey", key, "carrierName", mreq.CarrierName)
			return fmt.Errorf("carrier not found for app: %s", mreq.CarrierName)
		}
		distance = dmecommon.InfiniteDistance
		log.DebugLog(log.DebugLevelDmereq, ">>>Verify Location",
			"appName", key.Name,
			"lat", mreq.GpsLocation.Lat,
			"long", mreq.GpsLocation.Long)
		for _, c := range carr.insts {
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
	return nil
}

func GetClientLoc(mreq *dme.GetLocationRequest, reply *dme.GetLocationReply) {
	reply.CarrierName = mreq.CarrierName
	reply.Status = dme.GetLocationReply_LOC_FOUND
	reply.NetworkLocation = &dme.Loc{}
}
