package main

import (
	"net"
	"sync"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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
	id       uint64
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
		log.DebugLog(log.DebugLevelDmedb, "Adding app",
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
		log.DebugLog(log.DebugLevelDmedb, "Updating app inst",
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
		cNew.location = appInst.CloudletLoc
		cNew.id = appInst.Key.Id
		app.insts[cNew.cloudletKey] = cNew
		log.DebugLog(log.DebugLevelDmedb, "Adding app inst",
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
			log.DebugLog(log.DebugLevelDmedb, "Removing app inst",
				"appName", app.key.appKey.Name,
				"appVersion", app.key.appKey.Version,
				"carrier", c.carrierName,
				"lat", c.location.Lat,
				"long", c.location.Long)
		}
		if len(app.insts) == 0 {
			delete(tbl.apps, key)
			log.DebugLog(log.DebugLevelDmedb, "Removing app",
				"appName", app.key.appKey.Name,
				"appVersion", app.key.appKey.Version,
				"carrier", app.key.carrierName)
		}
		app.Unlock()
	}
	tbl.Unlock()
}

// pruneApps removes any data that was not sent by the controller.
func pruneApps(appInsts map[edgeproto.AppInstKey]struct{}) {
	var key edgeproto.AppInstKey
	var carrierKey carrierAppKey

	tbl := carrierAppTbl
	tbl.Lock()
	for _, app := range tbl.apps {
		app.Lock()
		for _, inst := range app.insts {
			key.AppKey = app.key.appKey
			key.CloudletKey = inst.cloudletKey
			key.Id = inst.id
			if _, found := appInsts[key]; !found {
				log.DebugLog(log.DebugLevelDmereq, "pruning app", "key", key)
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

// given the carrier app key, find closest cloudlet and
// give trhe distance
func findClosestForKey(key carrierAppKey, loc *dme.Loc) (*carrierAppInst, float64) {
	tbl := carrierAppTbl
	var found *carrierAppInst
	var distance, d float64

	distance = 10000

	tbl.RLock()

	app, ok := tbl.apps[key]
	if !ok {
		tbl.RUnlock()
		return nil, distance
	}

	log.DebugLog(log.DebugLevelDmereq, ">>>Find Closest",
		"appName", key.appKey.Name,
		"carrier", key.carrierName)
	for _, c := range app.insts {
		d = dmecommon.DistanceBetween(*loc, c.location)
		log.DebugLog(log.DebugLevelDmereq, "found cloudlet at",
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
		var ipaddr net.IP
		ipaddr = found.ip
		log.DebugLog(log.DebugLevelDmereq, "best cloudlet",
			"app", key.appKey.Name,
			"carrier", key.carrierName,
			"lat", found.location.Lat,
			"long", found.location.Long,
			"distance", distance,
			"uri", found.uri,
			"IP", ipaddr.String())
	}
	tbl.RUnlock()
	return found, distance
}

func findCloudlet(mreq *dme.Match_Engine_Request, mreply *dme.Match_Engine_Reply) {
	var key carrierAppKey
	var found *carrierAppInst
	publicCloudPadding := 100.0 // public clouds have to be this much closer in km

	key.carrierName = mreq.CarrierName
	key.appKey.DeveloperKey.Name = mreq.DevName
	key.appKey.Name = mreq.AppName
	key.appKey.Version = mreq.AppVers

	mreply.Status = dme.Match_Engine_Reply_FIND_NOTFOUND
	mreply.CloudletLocation = &dme.Loc{}

	log.DebugLog(log.DebugLevelDmereq, "findCloudlet", "carrier", key.carrierName, "app", key.appKey.Name, "developer", key.appKey.DeveloperKey.Name, "version", key.appKey.Version)

	var bestDistance float64 //just for logging
	c, carrierDistance := findClosestForKey(key, mreq.GpsLocation)
	searchPublicCloud := true

	if c != nil {
		found = c
		bestDistance = carrierDistance
		log.DebugLog(log.DebugLevelDmereq, "found carrier cloudlet", "uri", c.uri, "distance", carrierDistance)
		if carrierDistance <= publicCloudPadding {
			searchPublicCloud = false
		}
	}

	if searchPublicCloud {
		key.carrierName = "azure"
		a, azDistance := findClosestForKey(key, mreq.GpsLocation)
		key.carrierName = "gcp"
		g, gcpDistance := findClosestForKey(key, mreq.GpsLocation)

		azDistance += publicCloudPadding
		gcpDistance += publicCloudPadding
		log.DebugLog(log.DebugLevelDmereq, "public cloud padded distances", "azure", azDistance, "gcp", gcpDistance)

		if azDistance < gcpDistance && a != nil {
			if azDistance < carrierDistance {
				found = a
				bestDistance = azDistance
				log.DebugLog(log.DebugLevelDmereq, "found azure cloudlet", "uri", a.uri)
			}
		} else {
			if gcpDistance < azDistance && g != nil {
				if gcpDistance < carrierDistance {
					found = g
					bestDistance = gcpDistance
					log.DebugLog(log.DebugLevelDmereq, "found gcp cloudlet", "uri", g.uri)

				}
			}
		}
	}

	if found != nil {
		log.DebugLog(log.DebugLevelDmereq, "overall best cloudlet", "uri", found.uri, "distance", bestDistance)
		mreply.Status = dme.Match_Engine_Reply_FIND_FOUND
		mreply.Uri = found.uri
		mreply.ServiceIp = found.ip
		*mreply.CloudletLocation = found.location

	} else {
		mreply.Status = dme.Match_Engine_Reply_FIND_NOTFOUND
	}
}

func getCloudlets(mreq *dme.Match_Engine_Request, clist *dme.Match_Engine_Cloudlet_List) {
	var tbl *carrierApps
	tbl = carrierAppTbl
	foundCloudlets := make(map[edgeproto.CloudletKey]*dme.CloudletLocation)

	tbl.RLock()

	// find all the unique cloudlets, and the app instances for each.  the data is
	//stored as appinst->cloudlet and we need the opposite mapping.
	for _, a := range tbl.apps {
		//if the carrier name was provided, only look for cloudlets for that carrier, or for public cloudlets
		if mreq.CarrierName != "" && a.key.carrierName != "azure" && a.key.carrierName != "gcp" && mreq.CarrierName != a.key.carrierName {
			log.DebugLog(log.DebugLevelDmereq, "skipping cloudlet, mismatched carrier", "mreq.CarrierName", mreq.CarrierName, "app.CarrierName", a.key.carrierName)
			continue
		}
		//if the app name or version was provided, only look for cloudlets for that app
		if (mreq.AppName != "" && mreq.AppName != a.key.appKey.Name) ||
			(mreq.AppVers != "" && mreq.AppVers != a.key.appKey.Version) {
			continue
		}
		for _, i := range a.insts {
			cloc, exists := foundCloudlets[i.cloudletKey]
			if !exists {
				cloc = new(dme.CloudletLocation)
				d := dmecommon.DistanceBetween(*mreq.GpsLocation, i.location)
				cloc.GpsLocation = &i.location
				cloc.CarrierName = i.carrierName
				cloc.CloudletName = i.cloudletKey.Name
				cloc.Distance = d
			}
			ai := dme.Appinstance{}
			ai.Appname = a.key.appKey.Name
			ai.Appversion = a.key.appKey.Version
			ai.Uri = i.uri
			cloc.Appinstances = append(cloc.Appinstances, &ai)
			foundCloudlets[i.cloudletKey] = cloc
		}
	}
	for _, c := range foundCloudlets {
		clist.Cloudlets = append(clist.Cloudlets, c)
	}
	tbl.RUnlock()
}

func listAppinstTbl() {
	var app *carrierApp
	var inst *carrierAppInst
	var tbl *carrierApps

	tbl = carrierAppTbl
	tbl.RLock()
	for a := range tbl.apps {
		app = tbl.apps[a]
		log.DebugLog(log.DebugLevelDmedb, "app",
			"Name", app.key.appKey.Name,
			"Ver", app.key.appKey.Version,
			"Carrier", app.key.carrierName)
		for c := range app.insts {
			inst = app.insts[c]
			log.DebugLog(log.DebugLevelDmedb, "app",
				"Name", app.key.appKey.Name,
				"Ver", app.key.appKey.Version,
				"Carrier", app.key.carrierName,
				"Lat", inst.location.Lat,
				"Long", inst.location.Long)
		}
	}
	tbl.RUnlock()
}
