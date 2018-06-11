package main

import (
	"fmt"
	"sync"
	"math"
	"net"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

type cloudlet struct {
	// Unique identifier key for the cloudlet
	id uint64
	//Carrier ID for carrier hosting the cloudlet
	carrierId uint64
	//Carrier Name for carrier hosting the cloudlet
	carrierName string
	// IP to use to connect to and control cloudlet site
	accessIp []byte
	// Location of the cloudlet site (lat, long?)
	location edgeproto.Loc
	// Next cloudlet
	next *cloudlet
}

type app_carrier_key struct {
	carrier_name string
	app_key edgeproto.AppKey
}

type carrier_app_cloudlet struct {
	sync.RWMutex
	carrier_id uint64
	carrier_name string
	app_name string
	app_vers string
	app_developer string
	app_cloudlet_inst *cloudlet
}

type carrier_app struct {
	sync.RWMutex
	apps map[app_carrier_key]*carrier_app_cloudlet
}

var carrier_app_tbl *carrier_app

func setup_match_engine() {
	carrier_app_tbl = new(carrier_app)
	carrier_app_tbl.apps = make(map[app_carrier_key]*carrier_app_cloudlet)

	populate_tbl()
	list_appinst_tbl()
}

func torads(deg float64) float64 {
  return deg * math.Pi / 180;
}

// Use the ‘haversine’ formula to calculate the great-circle distance between two points
func distance_between(loc1, loc2 edgeproto.Loc) float64 {
	radiusofearth := 6371;
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

	a = math.Sin(diff_lat/2) * math.Sin(diff_lat/2) + math.Sin(diff_long/2) *
		math.Sin(diff_long/2) * math.Cos(rad_lat1) * math.Cos(rad_lat2)
	c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1 - a))
	dist = c * float64(radiusofearth)
	
	return dist
}

func add_app(app_inst *app, cloudlet_inst *cloudlet) {
	var key app_carrier_key
	var c, c_new *cloudlet
	var carrier *carrier_app_cloudlet
	var tbl *carrier_app

	tbl = carrier_app_tbl
	key.carrier_name = cloudlet_inst.carrierName
	key.app_key.DeveloperKey.Name = app_inst.developer
	key.app_key.Name = app_inst.name
	key.app_key.Version = app_inst.vers

	tbl.Lock()
	_, ok := tbl.apps[key]
	if (!ok) {
		// Key doesn't exists
		carrier = new(carrier_app_cloudlet)
		carrier.carrier_id = cloudlet_inst.carrierId
		carrier.carrier_name = cloudlet_inst.carrierName
		carrier.app_name = app_inst.name
		carrier.app_vers = app_inst.vers
		carrier.app_developer = app_inst.developer
		tbl.apps[key] = carrier
		fmt.Printf("Adding App %s/%s for Carrier = %s\n",
			carrier.app_name, carrier.app_vers, cloudlet_inst.carrierName);
	} else {
		carrier = tbl.apps[key]
	}

	// Todo: Check for updates
	
	c_new = new (cloudlet)
	*c_new = *cloudlet_inst
	carrier.Lock()
	// check if the cloudlet already exists
	c = carrier.app_cloudlet_inst
	if (c != nil) {
		c_new.next = c
	}
	carrier.app_cloudlet_inst = c_new
	fmt.Printf("Adding App %s/%s for Carrier = %s, Loc = %f/%f\n",
		carrier.app_name, carrier.app_vers, cloudlet_inst.carrierName,
		cloudlet_inst.location.Lat, cloudlet_inst.location.Long);
	carrier.Unlock()
	tbl.Unlock()
}

func find_cloudlet(mreq *dme.Match_Engine_Request, mreply *dme.Match_Engine_Reply) {
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
	
	mreply.Status = false
	tbl.RLock()
	carrier, ok := tbl.apps[key]
	if (!ok) {
		tbl.RUnlock()
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
			mreply.ServiceIp = c.accessIp
			mreply.CloudletLocation = &c.location
		}
		fmt.Printf("\n");
	}
	if (found != nil) {
		ipaddr = found.accessIp
		fmt.Printf("Found Loc = %f/%f with IP %s\n",
			found.location.Lat, found.location.Long, ipaddr.String());
		mreply.Status = true
	}
	tbl.RUnlock()
}

// add function delete and find

func list_appinst_tbl() {
	var c *cloudlet
	var carrier *carrier_app_cloudlet
	var tbl *carrier_app

	tbl = carrier_app_tbl
	tbl.RLock()
	for a := range tbl.apps {
		carrier = tbl.apps[a]
		fmt.Printf(">> app = %s/%s info for carrier %s\n", carrier.app_name,
			carrier.app_vers, carrier.carrier_name);
		c = carrier.app_cloudlet_inst
		for ; c != nil; c = c.next {
			fmt.Printf("app = %s/%s info for carrier = %s, Loc = %f/%f\n",
				carrier.app_name, carrier.app_vers, carrier.carrier_name,
				c.location.Lat, c.location.Long);
		}
	}
	tbl.RUnlock()
}
