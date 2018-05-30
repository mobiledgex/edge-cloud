package main

import (
	"fmt"
	"sync"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	//"github.com/mobiledgex/edge-cloud/util"
)

type cloudlet struct {
	// Unique identifier key
	id uint64
	//Carrier
	carrier string
	// IP to use to connect to and control cloudlet site
	accessIp []byte
	// Location of the cloudlet site (lat, long?)
	location edgeproto.Loc
	// Next cloudlet
	next *cloudlet
}

type app_carrier_key struct {
	carrier_id uint64
	app_id uint64
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
}


func add_app(app_inst *app, cloudlet_inst *cloudlet) {
	var key app_carrier_key
	var c, c_new *cloudlet
	var carrier *carrier_app_cloudlet
	var tbl *carrier_app

	tbl = carrier_app_tbl
	key.carrier_id = cloudlet_inst.id
	key.app_id = app_inst.id

	tbl.Lock()
	_, ok := tbl.apps[key]
	if (!ok) {
		// Key doesn't exists
		carrier = new(carrier_app_cloudlet)
		carrier.carrier_id = cloudlet_inst.id
		carrier.carrier_name = cloudlet_inst.carrier
		carrier.app_name = app_inst.name
		carrier.app_vers = app_inst.vers
		carrier.app_developer = app_inst.developer
		tbl.apps[key] = carrier
		fmt.Printf("Adding App %s/%s for Carrier = %s\n",
			carrier.app_name, carrier.app_vers, cloudlet_inst.carrier);
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
		carrier.app_name, carrier.app_vers, cloudlet_inst.carrier,
		cloudlet_inst.location.Lat, cloudlet_inst.location.Long);
	carrier.Unlock()
	tbl.Unlock()
}
