package main

import (
	"fmt"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	locapi "github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
)

func VerifyClientLoc(mreq *dme.Match_Engine_Request, mreply *dme.Match_Engine_Loc_Verify, peerIp string, locVerUrl string) {
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

	mreply.GpsLocationStatus = 0

	tbl.RLock()
	app, ok := tbl.apps[key]
	if !ok {
		tbl.RUnlock()
		fmt.Printf("Couldn't find the key\n")
		return
	}

	// if the dme was started with a location verify API URL, use that.  At some point in future,
	// this will be the only supported way to verify the location
	if locVerUrl != "" {
		mreply.GpsLocationStatus = locapi.CallLocationVerifyAPI(locVerUrl, mreq.GpsLocation.Lat, mreq.GpsLocation.Long, peerIp)
	} else {
		distance = 10000
		log.DebugLog(log.DebugLevelDmereq, ">>>Verify Location",
			"appName", key.appKey.Name,
			"carrier", key.carrierName,
			"lat", mreq.GpsLocation.Lat,
			"long", mreq.GpsLocation.Long)
		for _, c := range app.insts {
			d = dmecommon.Distance_between(*mreq.GpsLocation, c.location)
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
		if found != nil {
			if distance < 2 {
				mreply.GpsLocationStatus = 1
			} else if distance < 10 {
				mreply.GpsLocationStatus = 2
			} else if distance < 100 {
				mreply.GpsLocationStatus = 3
			} else {
				mreply.GpsLocationStatus = 4
			}
			log.DebugLog(log.DebugLevelDmereq, "verified location at",
				"lat", found.location.Lat,
				"long", found.location.Long,
				"distance", distance,
				"status", mreply.GpsLocationStatus,
				"uri", found.uri)
		}
	}

	tbl.RUnlock()
}

func GetClientLoc(mreq *dme.Match_Engine_Request, mloc *dme.Match_Engine_Loc) {
	mloc.CarrierName = mreq.CarrierName
	mloc.Status = dme.Match_Engine_Loc_LOC_FOUND
	mloc.NetworkLocation = mreq.GpsLocation
}
