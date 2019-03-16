package main

import (
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	locapi "github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

	mreply.GPS_Location_Status = dme.VerifyLocationReply_LOC_UNKNOWN
	mreply.GPS_Location_Accuracy_KM = -1

	log.DebugLog(log.DebugLevelDmereq, "Received Verify Location",
		"appName", key.Name,
		"appVersion", key.Version,
		"devName", key.DeveloperKey.Name,
		"VerifyLocToken", mreq.VerifyLocToken,
		"GPSLocation", mreq.GPSLocation)

	if mreq.GPSLocation == nil || (mreq.GPSLocation.Latitude == 0 && mreq.GPSLocation.Longitude == 0) {
		log.DebugLog(log.DebugLevelDmereq, "Invalid VerifyLocation request", "Error", "Missing GPSLocation")
		return grpc.Errorf(codes.InvalidArgument, "Missing GPS location")
	}

	if !util.IsLatitudeValid(mreq.GPSLocation.Latitude) || !util.IsLongitudeValid(mreq.GPSLocation.Longitude) {
		log.DebugLog(log.DebugLevelDmereq, "Invalid VerifyLocation GPSLocation", "lat", mreq.GPSLocation.Latitude, "long", mreq.GPSLocation.Longitude)
		return grpc.Errorf(codes.InvalidArgument, "Invalid GPSLocation")
	}

	tbl.RLock()
	defer tbl.RUnlock()

	app, ok := tbl.apps[key]
	if !ok {
		log.DebugLog(log.DebugLevelDmereq, "Could not find key in app table", "key", key)
		// return loc unknown
		return grpc.Errorf(codes.NotFound, "app not found: %s", key)
	}

	//handling for each carrier may be different.  As of now there is only standalone and TDG
	switch carrier {
	case "tdg":
		fallthrough
	case "TDG":
		if mreq.VerifyLocToken == "" {
			return grpc.Errorf(codes.InvalidArgument, "verifyloc token required")
		}
		result := locapi.CallTDGLocationVerifyAPI(locVerUrl, mreq.GPSLocation.Latitude, mreq.GPSLocation.Longitude, mreq.VerifyLocToken, tokSrvUrl)
		mreply.GPS_Location_Status = result.MatchEngineLocStatus
		mreply.GPS_Location_Accuracy_KM = result.DistanceRange
	default:
		// non-API based location uses cloudlets and so default and public cloudlets are not applicable
		carr, ok := app.carriers[mreq.CarrierName]
		if !ok {
			log.DebugLog(log.DebugLevelDmereq, "Could not find carrier for app", "appKey", key, "carrierName", mreq.CarrierName)
			return grpc.Errorf(codes.NotFound, "carrier not found for app: %s", mreq.CarrierName)
		}
		distance = dmecommon.InfiniteDistance
		log.DebugLog(log.DebugLevelDmereq, ">>>Verify Location",
			"appName", key.Name,
			"lat", mreq.GPSLocation.Latitude,
			"long", mreq.GPSLocation.Longitude)
		for _, c := range carr.insts {
			d = dmecommon.DistanceBetween(*mreq.GPSLocation, c.location)
			log.DebugLog(log.DebugLevelDmereq, "verify location at",
				"lat", c.location.Latitude,
				"long", c.location.Longitude,
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

		mreply.GPS_Location_Status = locval.MatchEngineLocStatus
		mreply.GPS_Location_Accuracy_KM = locval.DistanceRange

		log.DebugLog(log.DebugLevelDmereq, "verified location at",
			"lat", found.location.Latitude,
			"long", found.location.Longitude,
			"actual distance", distance,
			"distance range", mreply.GPS_Location_Accuracy_KM,
			"status", mreply.GPS_Location_Status,
			"uri", found.uri)

	}
	return nil
}

func GetClientLoc(mreq *dme.GetLocationRequest, reply *dme.GetLocationReply) {
	reply.CarrierName = mreq.CarrierName
	reply.Status = dme.GetLocationReply_LOC_FOUND
	reply.NetworkLocation = &dme.Loc{}
}
