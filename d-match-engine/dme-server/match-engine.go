package main

import (
	"math"
	"net"
	"sync"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
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
	// Ports and L7 Paths
	ports []dme.AppPort
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
		c.ports = appInst.MappedPorts
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
		cNew.ports = appInst.MappedPorts
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

// given the carrier app key, update the reply if we find a cloudlet closer
// than the max distance.  Return the distance and whether or not response was updated
func findClosestForKey(key carrierAppKey, loc *dme.Loc, maxDistance float64, mreply *dme.FindCloudletReply) (float64, bool) {
	tbl := carrierAppTbl
	var c, found *carrierAppInst
	var d float64
	var updated = false

	tbl.RLock()

	app, ok := tbl.apps[key]
	if !ok {
		tbl.RUnlock()
		return maxDistance, updated
	}

	log.DebugLog(log.DebugLevelDmereq, ">>>Find Closest",
		"appName", key.appKey.Name,
		"carrier", key.carrierName)
	for _, c = range app.insts {
		d = dmecommon.DistanceBetween(*loc, c.location)
		log.DebugLog(log.DebugLevelDmereq, "found cloudlet at",
			"lat", c.location.Lat,
			"long", c.location.Long,
			"maxDistance", maxDistance,
			"this-dist", d)
		if d < maxDistance {
			log.DebugLog(log.DebugLevelDmereq, "closer cloudlet", "uri", c.uri)
			updated = true
			maxDistance = d
			found = c
			mreply.FQDN = c.uri
			mreply.Status = dme.FindCloudletReply_FIND_FOUND
			*mreply.CloudletLocation = c.location
			mreply.Ports = copyPorts(c)
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
			"distance", maxDistance,
			"uri", found.uri,
			"IP", ipaddr.String())
	}
	tbl.RUnlock()
	return maxDistance, updated
}

func findCloudlet(ckey *dmecommon.CookieKey, mreq *dme.FindCloudletRequest, mreply *dme.FindCloudletReply) {
	var key carrierAppKey
	publicCloudPadding := 100.0 // public clouds have to be this much closer in km

	key.carrierName = mreq.CarrierName
	key.appKey.DeveloperKey.Name = ckey.DevName
	key.appKey.Name = ckey.AppName
	key.appKey.Version = ckey.AppVers
	mreply.Status = dme.FindCloudletReply_FIND_NOTFOUND
	mreply.CloudletLocation = &dme.Loc{}

	log.DebugLog(log.DebugLevelDmereq, "findCloudlet", "carrier", key.carrierName, "app", key.appKey.Name, "developer", key.appKey.DeveloperKey.Name, "version", key.appKey.Version)

	// first find carrier cloudlet
	bestDistance, updated := findClosestForKey(key, mreq.GpsLocation, dmecommon.InfiniteDistance, mreply)

	if updated {
		log.DebugLog(log.DebugLevelDmereq, "found carrier cloudlet", "FQDN", mreply.FQDN, "distance", bestDistance)
	}

	if updated && bestDistance > publicCloudPadding {
		paddedCarrierDistance := bestDistance - publicCloudPadding

		// look for an azure cloud closer than the carrier distance minus padding
		key.carrierName = cloudcommon.OperatorAzure
		azDistance, updated := findClosestForKey(key, mreq.GpsLocation, paddedCarrierDistance, mreply)
		if updated {
			log.DebugLog(log.DebugLevelDmereq, "found closer azure cloudlet", "FQDN", mreply.FQDN, "distance", azDistance)
			bestDistance = azDistance
		}

		// look for a gcp cloud closer than either the azure cloud or the carrier cloud
		key.carrierName = cloudcommon.OperatorGCP
		maxGCPDistance := math.Min(azDistance, paddedCarrierDistance)
		gcpDistance, updated := findClosestForKey(key, mreq.GpsLocation, maxGCPDistance, mreply)
		if updated {
			log.DebugLog(log.DebugLevelDmereq, "found closer gcp cloudlet", "FQDN", mreply.FQDN, "distance", gcpDistance)
			bestDistance = gcpDistance
		}
	}
	if mreply.Status == dme.FindCloudletReply_FIND_NOTFOUND {
		key.carrierName = cloudcommon.OperatorDeveloper
		// default cloudlet is at lat:0, long:0.  Look at any distance distance
		_, updated := findClosestForKey(key, mreq.GpsLocation, dmecommon.InfiniteDistance, mreply)
		if updated {
			log.DebugLog(log.DebugLevelDmereq, "found default operator cloudlet", "FQDN", mreply.FQDN)
			bestDistance = -1 //not used except in log
		} else {
			log.DebugLog(log.DebugLevelDmedb, "no default operator cloudlet for app", "appkey", key.appKey)
		}
	}

	if mreply.Status == dme.FindCloudletReply_FIND_FOUND {
		log.DebugLog(log.DebugLevelDmereq, "findCloudlet returning FIND_FOUND, overall best cloudlet", "FQDN", mreply.FQDN, "distance", bestDistance)
	} else {
		log.DebugLog(log.DebugLevelDmereq, "findCloudlet returning FIND_NOTFOUND")

	}
}

func isPublicCarrier(carriername string) bool {
	if carriername == cloudcommon.OperatorAzure ||
		carriername == cloudcommon.OperatorGCP ||
		carriername == cloudcommon.OperatorDeveloper {
		return true
	}
	return false
}

func getAppInstList(ckey *dmecommon.CookieKey, mreq *dme.AppInstListRequest, clist *dme.AppInstListReply) {
	var tbl *carrierApps
	tbl = carrierAppTbl
	foundCloudlets := make(map[edgeproto.CloudletKey]*dme.CloudletLocation)

	tbl.RLock()

	// find all the unique cloudlets, and the app instances for each.  the data is
	//stored as appinst->cloudlet and we need the opposite mapping.
	for _, a := range tbl.apps {
		//if the carrier name was provided, only look for cloudlets for that carrier, or for public cloudlets
		if mreq.CarrierName != "" && !isPublicCarrier(a.key.carrierName) && mreq.CarrierName != a.key.carrierName {
			log.DebugLog(log.DebugLevelDmereq, "skipping cloudlet, mismatched carrier", "mreq.CarrierName", mreq.CarrierName, "app.CarrierName", a.key.carrierName)
			continue
		}
		//if the app name or version was provided, only look for cloudlets for that app
		if (ckey.AppName != "" && ckey.AppName != a.key.appKey.Name) ||
			(ckey.AppVers != "" && ckey.AppVers != a.key.appKey.Version) {
			continue
		}
		for _, i := range a.insts {
			cloc, exists := foundCloudlets[i.cloudletKey]
			if !exists {
				cloc = new(dme.CloudletLocation)
				var d float64
				if mreq.CarrierName == cloudcommon.OperatorDeveloper {
					// there is no real distance as this is a fake cloudlet.
					d = dmecommon.InfiniteDistance
				} else {
					d = dmecommon.DistanceBetween(*mreq.GpsLocation, i.location)
				}
				cloc.GpsLocation = &i.location
				cloc.CarrierName = i.carrierName
				cloc.CloudletName = i.cloudletKey.Name
				cloc.Distance = d
			}
			ai := dme.Appinstance{}
			ai.Appname = a.key.appKey.Name
			ai.Appversion = a.key.appKey.Version
			ai.FQDN = i.uri
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

func copyPorts(cappInst *carrierAppInst) []*dme.AppPort {
	if cappInst.ports == nil || len(cappInst.ports) == 0 {
		return nil
	}
	ports := make([]*dme.AppPort, len(cappInst.ports))
	for ii, _ := range cappInst.ports {
		p := dme.AppPort{}
		p = cappInst.ports[ii]
		ports[ii] = &p
	}
	return ports
}
