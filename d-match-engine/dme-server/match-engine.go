package main

import (
	"math"
	"sync"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
)

// AppInst by carrier
type carrierAppInst struct {
	// Unique identifier key for the cloudlet
	cloudletKey edgeproto.CloudletKey
	//Carrier ID for carrier hosting the cloudlet
	carrierId uint64
	//Carrier Name for carrier hosting the cloudlet
	carrierName string
	// URI to connect to app inst in this cloudlet
	uri string
	// Ip to connect o app inst in this cloudlet (XXX why is this needed?)
	ip []byte
	// Location of the cloudlet site (lat, long?)
	location dme.Loc
}

type carrierAppKey struct {
	carrierName string // replace with edgeproto.OperatorKey?
	appKey      edgeproto.AppKey
}

// App by carrier
type carrierApp struct {
	sync.RWMutex
	carrierId uint64 // XXX: carrierId seems redundant with carrierName
	key       carrierAppKey
	insts     map[edgeproto.CloudletKey]*carrierAppInst
}

type carrierApps struct {
	sync.RWMutex
	apps map[carrierAppKey]*carrierApp
}

var carrierAppTbl *carrierApps

func setupMatchEngine() {
	carrierAppTbl = new(carrierApps)
	carrierAppTbl.apps = make(map[carrierAppKey]*carrierApp)
}

func torads(deg float64) float64 {
	return deg * math.Pi / 180
}

// Use the ‘haversine’ formula to calculate the great-circle distance between two points
func distance_between(loc1, loc2 dme.Loc) float64 {
	radiusofearth := 6371
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

	a = math.Sin(diff_lat/2)*math.Sin(diff_lat/2) + math.Sin(diff_long/2)*
		math.Sin(diff_long/2)*math.Cos(rad_lat1)*math.Cos(rad_lat2)
	c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	dist = c * float64(radiusofearth)

	return dist
}

// TODO: Have protoc auto-generate Equal functions.
func cloudletKeyEqual(key1 *edgeproto.CloudletKey, key2 *edgeproto.CloudletKey) bool {
	return key1.GetKeyString() == key2.GetKeyString()
}

func setCarrierAppKey(appInst *edgeproto.AppInst, key *carrierAppKey) {
	key.carrierName = appInst.Key.CloudletKey.OperatorKey.Name
	key.appKey = appInst.Key.AppKey
}

func addApp(appInst *edgeproto.AppInst) {
	var key carrierAppKey
	var cNew *carrierAppInst
	var app *carrierApp
	var tbl *carrierApps

	tbl = carrierAppTbl
	setCarrierAppKey(appInst, &key)

	tbl.Lock()
	_, ok := tbl.apps[key]
	if !ok {
		// Key doesn't exists
		app = new(carrierApp)
		app.insts = make(map[edgeproto.CloudletKey]*carrierAppInst)
		app.key = key
		tbl.apps[key] = app
		util.DebugLog(util.DebugLevelDmeDb, "Adding app",
			"appName", key.appKey.Name,
			"appVersion", key.appKey.Version,
			"carrier", key.carrierName)
	} else {
		app = tbl.apps[key]
	}

	app.Lock()
	if c, found := app.insts[appInst.Key.CloudletKey]; found {
		// update existing carrier app inst
		c.uri = appInst.Uri
		c.location = appInst.CloudletLoc
		util.DebugLog(util.DebugLevelDmeDb, "Updating app inst",
			"appName", app.key.appKey.Name,
			"appVersion", app.key.appKey.Version,
			"carrier", key.carrierName,
			"lat", appInst.CloudletLoc.Lat,
			"long", appInst.CloudletLoc.Long)
	} else {
		cNew = new(carrierAppInst)
		cNew.cloudletKey = appInst.Key.CloudletKey
		cNew.carrierName = key.carrierName
		cNew.uri = appInst.Uri
		cNew.ip = appInst.Ip
		cNew.location = appInst.CloudletLoc
		app.insts[cNew.cloudletKey] = cNew
		util.DebugLog(util.DebugLevelDmeDb, "Adding app inst",
			"appName", app.key.appKey.Name,
			"appVersion", app.key.appKey.Version,
			"carrier", cNew.carrierName,
			"lat", cNew.location.Lat,
			"long", cNew.location.Long)
	}
	app.Unlock()
	tbl.Unlock()
}

func removeApp(appInst *edgeproto.AppInst) {
	var key carrierAppKey
	var app *carrierApp
	var tbl *carrierApps

	tbl = carrierAppTbl
	setCarrierAppKey(appInst, &key)

	tbl.Lock()
	app, ok := tbl.apps[key]
	if ok {
		app.Lock()
		if c, found := app.insts[appInst.Key.CloudletKey]; found {
			delete(app.insts, appInst.Key.CloudletKey)
			util.DebugLog(util.DebugLevelDmeDb, "Removing app inst",
				"appName", app.key.appKey.Name,
				"appVersion", app.key.appKey.Version,
				"carrier", c.carrierName,
				"lat", c.location.Lat,
				"long", c.location.Long)
		}
		if len(app.insts) == 0 {
			delete(tbl.apps, key)
			util.DebugLog(util.DebugLevelDmeDb, "Removing app",
				"appName", app.key.appKey.Name,
				"appVersion", app.key.appKey.Version,
				"carrier", app.key.carrierName)
		}
		app.Unlock()
	}
	tbl.Unlock()
}

// pruneApps removes any data that was not sent by the controller.
func pruneApps(allMaps *notify.AllMaps) {
	var key edgeproto.AppInstKey
	var carrierKey carrierAppKey

	tbl := carrierAppTbl
	tbl.Lock()
	for _, app := range tbl.apps {
		app.Lock()
		for _, inst := range app.insts {
			key.AppKey = app.key.appKey
			key.CloudletKey = inst.cloudletKey
			if _, found := allMaps.AppInsts[key]; !found {
				delete(app.insts, key.CloudletKey)
			}
		}
		if len(app.insts) == 0 {
			carrierKey.carrierName = key.CloudletKey.OperatorKey.Name
			carrierKey.appKey = key.AppKey
			delete(tbl.apps, carrierKey)
		}
		app.Unlock()
	}
	tbl.Unlock()
}

func findCloudlet(mreq *dme.Match_Engine_Request, mreply *dme.Match_Engine_Reply) {
	var key carrierAppKey
	var c, found *carrierAppInst
	var app *carrierApp
	var distance, d float64
	var tbl *carrierApps

	tbl = carrierAppTbl
	key.carrierName = mreq.CarrierName
	key.appKey.DeveloperKey.Name = mreq.DevName
	key.appKey.Name = mreq.AppName
	key.appKey.Version = mreq.AppVers

	mreply.Status = false
	mreply.CloudletLocation = &dme.Loc{}
	tbl.RLock()
	app, ok := tbl.apps[key]
	if !ok {
		tbl.RUnlock()
		return
	}

	distance = 10000
	util.DebugLog(util.DebugLevelDmeReq, ">>>Find Cloudlet",
		"appName", key.appKey.Name,
		"carrier", key.carrierName)
	for _, c = range app.insts {
		d = distance_between(*mreq.GpsLocation, c.location)
		util.DebugLog(util.DebugLevelDmeReq, "found cloudlet at",
			"lat", c.location.Lat,
			"long", c.location.Long,
			"distance", distance,
			"this-dist", d)
		if d < distance {
			distance = d
			found = c
			mreply.Uri = c.uri
			*mreply.CloudletLocation = c.location
		}
	}
	if found != nil {
		util.DebugLog(util.DebugLevelDmeReq, "best cloudlet at",
			"lat", found.location.Lat,
			"long", found.location.Long,
			"distance", distance,
			"uri", found.uri)
		mreply.Status = true
	}
	tbl.RUnlock()
}
