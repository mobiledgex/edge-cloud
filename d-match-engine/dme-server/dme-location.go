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
	key.carrier_id = mreq.Carrier
	key.app_key.DeveloperKey = mreq.dev_name
	key.app_key.Name = mreq.app_name
	key.app_key.Version = mreq.app_vers
	
	tbl.RLock()
	carrier, ok := tbl.apps[key]
	if (!ok) {
		// mreply.Status = false
		tbl.RUnlock()
		return
	}

	// mreply.Status = true
	distance = 10000
	c = carrier.app_cloudlet_inst
	fmt.Printf(">>>Cloudlet for %s@%s\n", carrier.app_name, carrier.carrier_name)
	for ; c != nil; c = c.next {
		//Todo: Check the usage of embedded pointer to Loc in proto files
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
	ipaddr = found.accessIp
	fmt.Printf("Found Loc = %f/%f with IP %s\n",
		found.location.Lat, found.location.Long, ipaddr.String());
	if (d < 50) {
		mreply.GpsLocationStatus = 1
	} else {
		mreply.GpsLocationStatus = 0
	}
	tbl.RUnlock()
}
