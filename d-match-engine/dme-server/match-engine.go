package main

import (
	"fmt"
	"math"
	"net"
	"sync"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// AppInst within a cloudlet
type dmeAppInst struct {
	// Unique identifier key for the cloudlet
	cloudletKey edgeproto.CloudletKey
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

type dmeAppInsts struct {
	insts map[edgeproto.CloudletKey]*dmeAppInst
}

type dmeApp struct {
	sync.RWMutex
	appKey              edgeproto.AppKey
	carriers            map[string]*dmeAppInsts
	authPublicKey       string
	androidPackageName  string
	permitsPlatformApps bool
}

type dmeApps struct {
	sync.RWMutex
	apps map[edgeproto.AppKey]*dmeApp
}

var dmeAppTbl *dmeApps

func setupMatchEngine() {
	dmeAppTbl = new(dmeApps)
	dmeAppTbl.apps = make(map[edgeproto.AppKey]*dmeApp)
}

// TODO: Have protoc auto-generate Equal functions.
func cloudletKeyEqual(key1 *edgeproto.CloudletKey, key2 *edgeproto.CloudletKey) bool {
	return key1.GetKeyString() == key2.GetKeyString()
}

func addApp(in *edgeproto.App) {
	tbl := dmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.apps[in.Key]
	if !ok {
		// Key doesn't exists
		app = new(dmeApp)
		app.carriers = make(map[string]*dmeAppInsts)
		app.appKey = in.Key
		tbl.apps[in.Key] = app
		log.DebugLog(log.DebugLevelDmedb, "Adding app",
			"key", in.Key,
			"package", in.AndroidPackageName,
			"PermitsPlatformApps", in.PermitsPlatformApps)
	}
	app.Lock()
	defer app.Unlock()
	app.authPublicKey = in.AuthPublicKey
	app.androidPackageName = in.AndroidPackageName
	app.permitsPlatformApps = in.PermitsPlatformApps
}

func addAppInst(appInst *edgeproto.AppInst) {
	var cNew *dmeAppInst

	carrierName := appInst.Key.CloudletKey.OperatorKey.Name

	tbl := dmeAppTbl
	appkey := appInst.Key.AppKey
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.apps[appkey]
	if !ok {
		log.DebugLog(log.DebugLevelDmedb, "addAppInst: app not found", "key", appInst.Key)
		return
	}
	app.Lock()
	if _, foundCarrier := app.carriers[carrierName]; foundCarrier {
		log.DebugLog(log.DebugLevelDmedb, "carrier already exists", "carrierName", carrierName)
	} else {
		log.DebugLog(log.DebugLevelDmedb, "adding carrier for app", "carrierName", carrierName)
		app.carriers[carrierName] = new(dmeAppInsts)
		app.carriers[carrierName].insts = make(map[edgeproto.CloudletKey]*dmeAppInst)
	}
	if cl, foundAppInst := app.carriers[carrierName].insts[appInst.Key.CloudletKey]; foundAppInst {
		// update existing app inst
		cl.uri = appInst.Uri
		cl.location = appInst.CloudletLoc
		cl.ports = appInst.MappedPorts
		log.DebugLog(log.DebugLevelDmedb, "Updating app inst",
			"appName", app.appKey.Name,
			"appVersion", app.appKey.Version,
			"latitude", appInst.CloudletLoc.Latitude,
			"longitude", appInst.CloudletLoc.Longitude)
	} else {
		cNew = new(dmeAppInst)
		cNew.cloudletKey = appInst.Key.CloudletKey
		cNew.uri = appInst.Uri
		cNew.location = appInst.CloudletLoc
		cNew.id = appInst.Key.Id
		cNew.ports = appInst.MappedPorts
		app.carriers[carrierName].insts[cNew.cloudletKey] = cNew
		log.DebugLog(log.DebugLevelDmedb, "Adding app inst",
			"appName", app.appKey.Name,
			"appVersion", app.appKey.Version,
			"cloudletKey", appInst.Key.CloudletKey,
			"uri", appInst.Uri,
			"latitude", cNew.location.Latitude,
			"longitude", cNew.location.Longitude)
	}
	app.Unlock()
}

func removeApp(in *edgeproto.App) {
	tbl := dmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.apps[in.Key]
	if ok {
		app.Lock()
		delete(tbl.apps, in.Key)
		app.Unlock()
	}
}

func removeAppInst(appInst *edgeproto.AppInst) {
	var app *dmeApp
	var tbl *dmeApps

	tbl = dmeAppTbl
	appkey := appInst.Key.AppKey
	carrierName := appInst.Key.CloudletKey.OperatorKey.Name
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.apps[appkey]
	if ok {
		app.Lock()
		if c, foundCarrier := app.carriers[carrierName]; foundCarrier {
			if cl, foundAppInst := c.insts[appInst.Key.CloudletKey]; foundAppInst {
				delete(app.carriers[carrierName].insts, appInst.Key.CloudletKey)
				log.DebugLog(log.DebugLevelDmedb, "Removing app inst",
					"appName", appkey.Name,
					"appVersion", appkey.Version,
					"latitude", cl.location.Latitude,
					"longitude", cl.location.Longitude)
			}
			if len(app.carriers[carrierName].insts) == 0 {
				delete(tbl.apps[appkey].carriers, carrierName)
				log.DebugLog(log.DebugLevelDmedb, "Removing carrier for app",
					"carrier", carrierName,
					"appName", appkey.Name,
					"appVersion", appkey.Version)
			}
		}
		app.Unlock()
	}
}

// pruneApps removes any data that was not sent by the controller.
func pruneApps(apps map[edgeproto.AppKey]struct{}) {
	tbl := dmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	for key, app := range tbl.apps {
		app.Lock()
		if _, found := apps[key]; !found {
			delete(tbl.apps, key)
		}
		app.Unlock()
	}
}

// pruneApps removes any data that was not sent by the controller.
func pruneAppInsts(appInsts map[edgeproto.AppInstKey]struct{}) {
	var key edgeproto.AppInstKey

	log.DebugLog(log.DebugLevelDmereq, "pruneAppInsts called")

	tbl := dmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	for _, app := range tbl.apps {
		app.Lock()
		for c, carr := range app.carriers {
			for _, inst := range carr.insts {
				key.AppKey = app.appKey
				key.CloudletKey = inst.cloudletKey
				key.Id = inst.id
				if _, foundAppInst := appInsts[key]; !foundAppInst {
					log.DebugLog(log.DebugLevelDmereq, "pruning app", "key", key)
					delete(carr.insts, key.CloudletKey)
				}
			}
			if len(carr.insts) == 0 {
				log.DebugLog(log.DebugLevelDmereq, "pruneAppInsts delete carriers")
				delete(app.carriers, c)
			}
		}
		app.Unlock()
	}
}

// given the carrier, update the reply if we find a cloudlet closer
// than the max distance.  Return the distance and whether or not response was updated
func findClosestForCarrier(carrierName string, key edgeproto.AppKey, loc *dme.Loc, maxDistance float64, mreply *dme.FindCloudletReply) (float64, bool) {
	tbl := dmeAppTbl
	var d float64
	var updated = false
	var found *dmeAppInst
	tbl.RLock()
	defer tbl.RUnlock()
	app, ok := tbl.apps[key]
	if !ok {
		return maxDistance, updated
	}

	log.DebugLog(log.DebugLevelDmereq, "Find Closest", "appkey", key, "carrierName", carrierName)

	if c, carrierFound := app.carriers[carrierName]; carrierFound {
		for _, i := range c.insts {
			d = dmecommon.DistanceBetween(*loc, i.location)
			log.DebugLog(log.DebugLevelDmereq, "found cloudlet at",
				"latitude", i.location.Latitude,
				"longitude", i.location.Longitude,
				"maxDistance", maxDistance,
				"this-dist", d)
			if d < maxDistance {
				log.DebugLog(log.DebugLevelDmereq, "closer cloudlet", "uri", i.uri)
				updated = true
				maxDistance = d
				found = i
				mreply.FQDN = i.uri
				mreply.Status = dme.FindCloudletReply_FIND_FOUND
				*mreply.CloudletLocation = i.location
				mreply.Ports = copyPorts(i)
			}
		}

		if found != nil {
			var ipaddr net.IP
			ipaddr = found.ip
			log.DebugLog(log.DebugLevelDmereq, "best cloudlet",
				"app", key.Name,
				"carrier", carrierName,
				"latitude", found.location.Latitude,
				"longitude", found.location.Longitude,
				"distance", maxDistance,
				"uri", found.uri,
				"IP", ipaddr.String())
		}
	}
	return maxDistance, updated
}

// returns true if if the requested app allows the registered app to
// access APIs on its behalf
func requestedAppPermitsRegisteredApp(requestedApp edgeproto.AppKey, registeredApp edgeproto.AppKey) bool {
	// if the 2 apps match, allow it.  It means the client requested the same app as was registered
	var tbl *dmeApps
	tbl = dmeAppTbl

	if requestedApp == registeredApp {
		return true
	}
	if !cloudcommon.IsPlatformApp(registeredApp.DeveloperKey.Name, registeredApp.Name) {
		return false
	}
	// now find the app and see if it permits platform apps
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.apps[requestedApp]
	return ok && app.permitsPlatformApps
}

func findCloudlet(ckey *dmecommon.CookieKey, mreq *dme.FindCloudletRequest, mreply *dme.FindCloudletReply) error {
	var appkey edgeproto.AppKey
	publicCloudPadding := 100.0 // public clouds have to be this much closer in km
	appkey.DeveloperKey.Name = ckey.DevName
	appkey.Name = ckey.AppName
	appkey.Version = ckey.AppVers
	mreply.Status = dme.FindCloudletReply_FIND_NOTFOUND
	mreply.CloudletLocation = &dme.Loc{}

	// specifying an app in the request is allowed for platform apps only,
	// and should require permission from the developer of the actual app
	if mreq.AppName != "" || mreq.DevName != "" || mreq.AppVers != "" {
		var reqkey edgeproto.AppKey
		reqkey.DeveloperKey.Name = mreq.DevName
		reqkey.Name = mreq.AppName
		reqkey.Version = mreq.AppVers
		if !requestedAppPermitsRegisteredApp(reqkey, appkey) {
			return fmt.Errorf("Access to requested app: Devname: %s Appname: %s AppVers: %s not allowed for the registered app: Devname: %s Appname: %s Appvers: %s",
				mreq.DevName, mreq.AppName, mreq.AppVers, appkey.DeveloperKey.Name, appkey.Name, appkey.Version)
		}
		//update the appkey to use the requested key
		appkey = reqkey
	}

	log.DebugLog(log.DebugLevelDmereq, "findCloudlet", "carrier", mreq.CarrierName, "app", appkey.Name, "developer", appkey.DeveloperKey.Name, "version", appkey.Version)

	// first find carrier cloudlet
	bestDistance, updated := findClosestForCarrier(mreq.CarrierName, appkey, mreq.GpsLocation, dmecommon.InfiniteDistance, mreply)

	if updated {
		log.DebugLog(log.DebugLevelDmereq, "found carrier cloudlet", "FQDN", mreply.FQDN, "distance", bestDistance)
	}

	if bestDistance > publicCloudPadding {
		paddedCarrierDistance := bestDistance - publicCloudPadding

		// look for an azure cloud closer than the carrier distance minus padding
		azDistance, updated := findClosestForCarrier(cloudcommon.OperatorAzure, appkey, mreq.GpsLocation, paddedCarrierDistance, mreply)
		if updated {
			log.DebugLog(log.DebugLevelDmereq, "found closer azure cloudlet", "FQDN", mreply.FQDN, "distance", azDistance)
			bestDistance = azDistance
		}

		// look for a gcp cloud closer than either the azure cloud or the carrier cloud
		maxGCPDistance := math.Min(azDistance, paddedCarrierDistance)
		gcpDistance, updated := findClosestForCarrier(cloudcommon.OperatorGCP, appkey, mreq.GpsLocation, maxGCPDistance, mreply)
		if updated {
			log.DebugLog(log.DebugLevelDmereq, "found closer gcp cloudlet", "FQDN", mreply.FQDN, "distance", gcpDistance)
			bestDistance = gcpDistance
		}
	}
	if mreply.Status == dme.FindCloudletReply_FIND_NOTFOUND {
		// default cloudlet is at lat:0, long:0.  Look at any distance distance
		_, updated := findClosestForCarrier(cloudcommon.OperatorDeveloper, appkey, mreq.GpsLocation, dmecommon.InfiniteDistance, mreply)
		if updated {
			log.DebugLog(log.DebugLevelDmereq, "found default operator cloudlet", "FQDN", mreply.FQDN)
			bestDistance = -1 //not used except in log
		} else {
			log.DebugLog(log.DebugLevelDmedb, "no default operator cloudlet for app", "appkey", appkey)
		}
	}

	if mreply.Status == dme.FindCloudletReply_FIND_FOUND {
		log.DebugLog(log.DebugLevelDmereq, "findCloudlet returning FIND_FOUND, overall best cloudlet", "FQDN", mreply.FQDN, "distance", bestDistance)
	} else {
		log.DebugLog(log.DebugLevelDmereq, "findCloudlet returning FIND_NOTFOUND")

	}
	return nil
}

func isPublicCarrier(carriername string) bool {
	if carriername == cloudcommon.OperatorAzure ||
		carriername == cloudcommon.OperatorGCP ||
		carriername == cloudcommon.OperatorDeveloper {
		return true
	}
	return false
}

func getFqdnList(mreq *dme.FqdnListRequest, clist *dme.FqdnListReply) {
	var tbl *dmeApps
	tbl = dmeAppTbl
	tbl.RLock()
	defer tbl.RUnlock()
	for _, a := range tbl.apps {
		// if the app it itself a platform app, it is not returned here
		if cloudcommon.IsPlatformApp(a.appKey.DeveloperKey.Name, a.appKey.Name) {
			continue
		}
		// if the app does not permit platform apps to access it, skip it
		if !a.permitsPlatformApps {
			log.DebugLog(log.DebugLevelDmereq, "skipping non permitted app for getFqdnList", "appkey", a.appKey)
			continue
		}

		c, defaultCarrierFound := a.carriers[cloudcommon.OperatorDeveloper]
		if defaultCarrierFound {
			for _, i := range c.insts {
				if i.cloudletKey == cloudcommon.DefaultCloudletKey {
					aq := dme.AppFqdn{AppName: a.appKey.Name, DevName: a.appKey.DeveloperKey.Name, AppVers: a.appKey.Version, FQDN: i.uri, AndroidPackageName: a.androidPackageName}
					clist.AppFqdns = append(clist.AppFqdns, &aq)
				}
			}
		}
	}
	clist.Status = dme.FqdnListReply_FL_SUCCESS
}

func getAppInstList(ckey *dmecommon.CookieKey, mreq *dme.AppInstListRequest, clist *dme.AppInstListReply) {
	var tbl *dmeApps
	tbl = dmeAppTbl
	foundCloudlets := make(map[edgeproto.CloudletKey]*dme.CloudletLocation)

	tbl.RLock()
	defer tbl.RUnlock()

	// find all the unique cloudlets, and the app instances for each.  the data is
	//stored as appinst->cloudlet and we need the opposite mapping.
	for _, a := range tbl.apps {

		//if the app name or version was provided, only look for cloudlets for that app
		if (ckey.AppName != "" && ckey.AppName != a.appKey.Name) ||
			(ckey.AppVers != "" && ckey.AppVers != a.appKey.Version) {
			continue
		}
		for cname, c := range a.carriers {
			//if the carrier name was provided, only look for cloudlets for that carrier, or for public cloudlets
			if mreq.CarrierName != "" && !isPublicCarrier(cname) && mreq.CarrierName != cname {
				log.DebugLog(log.DebugLevelDmereq, "skipping cloudlet, mismatched carrier", "mreq.CarrierName", mreq.CarrierName, "i.cloudletKey.OperatorKey.Name", cname)
				continue
			}
			for _, i := range c.insts {
				cloc, exists := foundCloudlets[i.cloudletKey]
				if !exists {
					cloc = new(dme.CloudletLocation)
					var d float64

					// do not return the default instance
					if i.cloudletKey == cloudcommon.DefaultCloudletKey {
						continue
					}
					d = dmecommon.DistanceBetween(*mreq.GpsLocation, i.location)
					cloc.GpsLocation = &i.location
					cloc.CarrierName = i.cloudletKey.OperatorKey.Name
					cloc.CloudletName = i.cloudletKey.Name
					cloc.Distance = d
				}
				ai := dme.Appinstance{}
				ai.AppName = a.appKey.Name
				ai.AppVers = a.appKey.Version
				ai.FQDN = i.uri
				ai.Ports = copyPorts(i)
				cloc.Appinstances = append(cloc.Appinstances, &ai)
				foundCloudlets[i.cloudletKey] = cloc
			}
		}
	}
	for _, c := range foundCloudlets {
		clist.Cloudlets = append(clist.Cloudlets, c)
	}
	clist.Status = dme.AppInstListReply_AI_SUCCESS
}

func listAppinstTbl() {
	var app *dmeApp
	var inst *dmeAppInst
	var tbl *dmeApps

	tbl = dmeAppTbl
	tbl.RLock()
	defer tbl.RUnlock()

	for a := range tbl.apps {
		app = tbl.apps[a]
		log.DebugLog(log.DebugLevelDmedb, "app",
			"Name", app.appKey.Name,
			"Ver", app.appKey.Version)
		for cname, c := range app.carriers {
			for _ = range c.insts {
				log.DebugLog(log.DebugLevelDmedb, "app",
					"Name", app.appKey.Name,
					"carrier", cname,
					"Ver", app.appKey.Version,
					"Latitude", inst.location.Latitude,
					"Longitude", inst.location.Longitude)
			}
		}
	}
}

func copyPorts(cappInst *dmeAppInst) []*dme.AppPort {
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
