package main

import (
	"fmt"
	"net"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

func VerifyClientLoc(mreq *dme.Match_Engine_Request, mreply *dme.Match_Engine_Loc_Verify) {
	var key app_carrier_key
	var c, found *cloudlet
	var carrier *carrier_app_cloudlet
	var distance, d float64
	var tbl *carrier_app
	var ipaddr net.IP

	tbl = carrier_app_tbl	
	key.carrier_name = mreq.CarrierName
	key.app_key.DeveloperKey.Name = mreq.DevName
	key.app_key.Name = mreq.AppName
	key.app_key.Version = mreq.AppVers
	
	mreply.GpsLocationStatus = 0
	
	tbl.RLock()
	carrier, ok := tbl.apps[key]
	if (!ok) {
		tbl.RUnlock()
		fmt.Printf("Couldn't find the key\n");
		return
	}

	distance = 10000
	c = carrier.app_cloudlet_inst
	fmt.Printf(">>>Cloudlet for %s@%s\n", carrier.app_name, carrier.carrier_name)
	for ; c != nil; c = c.next {
		d = distance_between(*mreq.GpsLocation, c.location)
		fmt.Printf("Loc = %f/%f is at dist %f. ",
			c.location.Lat, c.location.Long, d);
		if (d < distance) {
			fmt.Printf("Repl. with new dist %f.", d)
			distance = d
			found = c
		}
		fmt.Printf("\n");
	}
	if (found != nil) {
		ipaddr = found.accessIp
		fmt.Printf("Found Loc = %f/%f with IP %s\n",
			found.location.Lat, found.location.Long, ipaddr.String());
		if (d < 2) {
			mreply.GpsLocationStatus = 1
		} else if (d < 10) {
			mreply.GpsLocationStatus = 2
		} else if (d < 100) {
			mreply.GpsLocationStatus = 3
		} else {
			mreply.GpsLocationStatus = 4
		}
	}
		
	tbl.RUnlock()
}

func GetClientLoc(mreq *dme.Match_Engine_Request, mloc *dme.Match_Engine_Loc) {
	mloc.CarrierName = mreq.CarrierName
	mloc.Status = 1
	mloc.NetworkLocation = mreq.GpsLocation
}
