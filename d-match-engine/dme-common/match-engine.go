package dmecommon

import (
	"context"
	"math"
	"net"
	"strings"
	"sync"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// AppInst within a cloudlet
type DmeAppInst struct {
	// Unique identifier key for the clusterInst
	clusterInstKey edgeproto.ClusterInstKey
	// URI to connect to app inst in this cloudlet
	uri string
	// Ip to connect o app inst in this cloudlet (XXX why is this needed?)
	ip []byte
	// Location of the cloudlet site (lat, long?)
	location dme.Loc
	id       uint64
	// Ports and L7 Paths
	ports []dme.AppPort
	// State of the cloudlet - copy of the DmeCloudlet
	cloudletState edgeproto.CloudletState
	// Health state of the appInst
	appInstHealth edgeproto.HealthCheck
}

type DmeAppInsts struct {
	Insts map[edgeproto.ClusterInstKey]*DmeAppInst
}

type DmeApp struct {
	sync.RWMutex
	AppKey             edgeproto.AppKey
	Carriers           map[string]*DmeAppInsts
	AuthPublicKey      string
	AndroidPackageName string
	OfficialFqdn       string
	AutoProvPolicy     *AutoProvPolicy
}

type DmeCloudlet struct {
	// No need for a mutex - protected under DmeApps mutex
	CloudletKey edgeproto.CloudletKey
	State       edgeproto.CloudletState
}

type AutoProvPolicy struct {
	DeployClientCount uint32
	IntervalCount     uint32
	Cloudlets         map[string][]*edgeproto.AutoProvCloudlet // index is carrier
}

type DmeApps struct {
	sync.RWMutex
	Apps                       map[edgeproto.AppKey]*DmeApp
	Cloudlets                  map[edgeproto.CloudletKey]*DmeCloudlet
	AutoProvPolicies           map[edgeproto.PolicyKey]*AutoProvPolicy
	FreeReservableClusterInsts edgeproto.FreeReservableClusterInstCache
}

var DmeAppTbl *DmeApps

// Stats are collected per App per Cloudlet and per method name (verifylocation, etc).
type StatKey struct {
	AppKey        edgeproto.AppKey
	CloudletFound edgeproto.CloudletKey
	CellId        uint32
	Method        string
}

// This is for passing the carrier/cloudlet in the context
type StatKeyContextType string

var StatKeyContextKey = StatKeyContextType("statKey")

func SetupMatchEngine() {
	DmeAppTbl = new(DmeApps)
	DmeAppTbl.Apps = make(map[edgeproto.AppKey]*DmeApp)
	DmeAppTbl.Cloudlets = make(map[edgeproto.CloudletKey]*DmeCloudlet)
	DmeAppTbl.AutoProvPolicies = make(map[edgeproto.PolicyKey]*AutoProvPolicy)
	DmeAppTbl.FreeReservableClusterInsts.Init()
}

// AppInst state is a superset of the cloudlet state and appInst state
// Returns if this AppInstance is usable or not
func IsAppInstUsable(appInst *DmeAppInst) bool {
	if appInst == nil {
		return false
	}
	if appInst.cloudletState == edgeproto.CloudletState_CLOUDLET_STATE_READY {
		return appInst.appInstHealth == edgeproto.HealthCheck_HEALTH_CHECK_OK
	}
	return false
}

// TODO: Have protoc auto-generate Equal functions.
func cloudletKeyEqual(key1 *edgeproto.CloudletKey, key2 *edgeproto.CloudletKey) bool {
	return key1.GetKeyString() == key2.GetKeyString()
}

func AddApp(in *edgeproto.App) {
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.Apps[in.Key]
	if !ok {
		// Key doesn't exists
		app = new(DmeApp)
		app.Carriers = make(map[string]*DmeAppInsts)
		app.AppKey = in.Key
		tbl.Apps[in.Key] = app
		log.DebugLog(log.DebugLevelDmedb, "Adding app",
			"key", in.Key,
			"package", in.AndroidPackageName,
			"OfficialFqdn", in.OfficialFqdn)
	}
	app.Lock()
	defer app.Unlock()
	app.AuthPublicKey = in.AuthPublicKey
	app.AndroidPackageName = in.AndroidPackageName
	app.OfficialFqdn = in.OfficialFqdn
	clearAutoProvStats := false
	if app.AutoProvPolicy != nil && in.AutoProvPolicy == "" {
		clearAutoProvStats = true
	}
	app.AutoProvPolicy = nil
	if in.AutoProvPolicy != "" {
		ppKey := edgeproto.PolicyKey{
			Developer: in.Key.DeveloperKey.Name,
			Name:      in.AutoProvPolicy,
		}
		if pp, found := tbl.AutoProvPolicies[ppKey]; found {
			app.AutoProvPolicy = pp
		} else {
			log.DebugLog(log.DebugLevelDmedb, "AutoProvPolicy on App not found", "app", in.Key, "policy", app.AutoProvPolicy)
		}
	}
	if clearAutoProvStats {
		autoProvStats.Clear(&in.Key)
	}
}

func AddAppInst(appInst *edgeproto.AppInst) {
	var cNew *DmeAppInst

	carrierName := appInst.Key.ClusterInstKey.CloudletKey.OperatorKey.Name

	tbl := DmeAppTbl
	appkey := appInst.Key.AppKey
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.Apps[appkey]
	if !ok {
		log.DebugLog(log.DebugLevelDmedb, "addAppInst: app not found", "key", appInst.Key)
		return
	}
	app.Lock()
	if _, foundCarrier := app.Carriers[carrierName]; foundCarrier {
		log.DebugLog(log.DebugLevelDmedb, "carrier already exists", "carrierName", carrierName)
	} else {
		log.DebugLog(log.DebugLevelDmedb, "adding carrier for app", "carrierName", carrierName)
		app.Carriers[carrierName] = new(DmeAppInsts)
		app.Carriers[carrierName].Insts = make(map[edgeproto.ClusterInstKey]*DmeAppInst)
	}
	if cl, foundAppInst := app.Carriers[carrierName].Insts[appInst.Key.ClusterInstKey]; foundAppInst {
		// update existing app inst
		cl.uri = appInst.Uri
		cl.location = appInst.CloudletLoc
		cl.ports = appInst.MappedPorts
		cl.appInstHealth = appInst.HealthCheck
		if cloudlet, foundCloudlet := tbl.Cloudlets[appInst.Key.ClusterInstKey.CloudletKey]; foundCloudlet {
			cl.cloudletState = cloudlet.State
		} else {
			cl.cloudletState = edgeproto.CloudletState_CLOUDLET_STATE_UNKNOWN
		}
		log.DebugLog(log.DebugLevelDmedb, "Updating app inst",
			"appName", app.AppKey.Name,
			"appVersion", app.AppKey.Version,
			"latitude", appInst.CloudletLoc.Latitude,
			"longitude", appInst.CloudletLoc.Longitude,
			"state", appInst.State)
	} else {
		cNew = new(DmeAppInst)
		cNew.clusterInstKey = appInst.Key.ClusterInstKey
		cNew.uri = appInst.Uri
		cNew.location = appInst.CloudletLoc
		cNew.ports = appInst.MappedPorts
		cNew.appInstHealth = appInst.HealthCheck
		if cloudlet, foundCloudlet := tbl.Cloudlets[appInst.Key.ClusterInstKey.CloudletKey]; foundCloudlet {
			cNew.cloudletState = cloudlet.State
		} else {
			cNew.cloudletState = edgeproto.CloudletState_CLOUDLET_STATE_UNKNOWN
		}
		app.Carriers[carrierName].Insts[cNew.clusterInstKey] = cNew
		log.DebugLog(log.DebugLevelDmedb, "Adding app inst",
			"appName", app.AppKey.Name,
			"appVersion", app.AppKey.Version,
			"cloudletKey", appInst.Key.ClusterInstKey.CloudletKey,
			"uri", appInst.Uri,
			"latitude", cNew.location.Latitude,
			"longitude", cNew.location.Longitude,
			"HealthState", cNew.appInstHealth)
	}
	app.Unlock()
}

func RemoveApp(in *edgeproto.App) {
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.Apps[in.Key]
	if ok {
		app.Lock()
		delete(tbl.Apps, in.Key)
		app.Unlock()
	}
	// Clear auto-prov stats for App
	if autoProvStats != nil {
		autoProvStats.Clear(&in.Key)
	}
}

func RemoveAppInst(appInst *edgeproto.AppInst) {
	var app *DmeApp
	var tbl *DmeApps

	tbl = DmeAppTbl
	appkey := appInst.Key.AppKey
	carrierName := appInst.Key.ClusterInstKey.CloudletKey.OperatorKey.Name
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.Apps[appkey]
	if ok {
		app.Lock()
		if c, foundCarrier := app.Carriers[carrierName]; foundCarrier {
			if cl, foundAppInst := c.Insts[appInst.Key.ClusterInstKey]; foundAppInst {
				delete(app.Carriers[carrierName].Insts, appInst.Key.ClusterInstKey)
				log.DebugLog(log.DebugLevelDmedb, "Removing app inst",
					"appName", appkey.Name,
					"appVersion", appkey.Version,
					"latitude", cl.location.Latitude,
					"longitude", cl.location.Longitude)
			}
			if len(app.Carriers[carrierName].Insts) == 0 {
				delete(tbl.Apps[appkey].Carriers, carrierName)
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
func PruneApps(apps map[edgeproto.AppKey]struct{}) {
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	for key, app := range tbl.Apps {
		app.Lock()
		if _, found := apps[key]; !found {
			delete(tbl.Apps, key)
		}
		app.Unlock()
	}
	// Clear auto-prov stats for deleted Apps
	if autoProvStats != nil {
		autoProvStats.Prune(apps)
	}
}

// pruneApps removes any data that was not sent by the controller.
func PruneAppInsts(appInsts map[edgeproto.AppInstKey]struct{}) {
	var key edgeproto.AppInstKey

	log.DebugLog(log.DebugLevelDmereq, "pruneAppInsts called")

	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	for _, app := range tbl.Apps {
		app.Lock()
		for c, carr := range app.Carriers {
			for _, inst := range carr.Insts {
				key.AppKey = app.AppKey
				key.ClusterInstKey = inst.clusterInstKey
				if _, foundAppInst := appInsts[key]; !foundAppInst {
					log.DebugLog(log.DebugLevelDmereq, "pruning app", "key", key)
					delete(carr.Insts, key.ClusterInstKey)
				}
			}
			if len(carr.Insts) == 0 {
				log.DebugLog(log.DebugLevelDmereq, "pruneAppInsts delete carriers")
				delete(app.Carriers, c)
			}
		}
		app.Unlock()
	}
}

func DeleteCloudletInfo(info *edgeproto.CloudletInfo) {
	log.DebugLog(log.DebugLevelDmereq, "DeleteCloudletInfo called")
	tbl := DmeAppTbl
	carrier := info.Key.OperatorKey.Name
	tbl.Lock()
	defer tbl.Unlock()

	// If there are still appInsts on that cloudlet they should be disabled
	for _, app := range tbl.Apps {
		app.Lock()
		if c, found := app.Carriers[carrier]; found {
			for clusterInstKey, _ := range c.Insts {
				if cloudletKeyEqual(&clusterInstKey.CloudletKey, &info.Key) {
					c.Insts[clusterInstKey].cloudletState = edgeproto.CloudletState_CLOUDLET_STATE_NOT_PRESENT
				}
			}
		}
		app.Unlock()
	}
	if _, found := tbl.Cloudlets[info.Key]; found {
		log.DebugLog(log.DebugLevelDmereq, "delete cloudletInfo", "key", info.Key)
		delete(tbl.Cloudlets, info.Key)
	}
}

// Remove any Cloudlets we track that no longer exist and reset the state for the AppInsts
func PruneCloudlets(cloudlets map[edgeproto.CloudletKey]struct{}) {
	log.DebugLog(log.DebugLevelDmereq, "PruneCloudlets called")
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()

	for _, app := range tbl.Apps {
		app.Lock()
		for _, carr := range app.Carriers {
			for clusterInstKey, _ := range carr.Insts {
				if _, found := cloudlets[clusterInstKey.CloudletKey]; !found {
					carr.Insts[clusterInstKey].cloudletState = edgeproto.CloudletState_CLOUDLET_STATE_NOT_PRESENT
				}
			}
		}
		app.Unlock()
	}
	for cloudlet, _ := range cloudlets {
		if _, foundInfo := tbl.Cloudlets[cloudlet]; !foundInfo {
			log.DebugLog(log.DebugLevelDmereq, "pruning cloudletInfo", "key", cloudlet)
			delete(tbl.Cloudlets, cloudlet)
		}
	}
}

type AutoProvPolicyHandler struct{}

func (s *AutoProvPolicyHandler) Update(ctx context.Context, in *edgeproto.AutoProvPolicy, rev int64) {
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	pp, ok := tbl.AutoProvPolicies[in.Key]
	if !ok {
		pp = &AutoProvPolicy{}
		tbl.AutoProvPolicies[in.Key] = pp
	}
	pp.DeployClientCount = in.DeployClientCount
	pp.IntervalCount = in.DeployIntervalCount
	log.DebugLog(log.DebugLevelDmereq, "update AutoProvPolicy", "policy", in)
	// Organize potentional cloudlets by carrier
	pp.Cloudlets = make(map[string][]*edgeproto.AutoProvCloudlet)
	if in.Cloudlets != nil {
		for _, provCloudlet := range in.Cloudlets {
			list, found := pp.Cloudlets[provCloudlet.Key.OperatorKey.Name]
			if !found {
				list = make([]*edgeproto.AutoProvCloudlet, 0)
			}
			list = append(list, provCloudlet)
			pp.Cloudlets[provCloudlet.Key.OperatorKey.Name] = list
		}
	}
}

func (s *AutoProvPolicyHandler) Delete(ctx context.Context, in *edgeproto.AutoProvPolicy, rev int64) {
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	delete(tbl.AutoProvPolicies, in.Key)
}

func (s *AutoProvPolicyHandler) Prune(ctx context.Context, keys map[edgeproto.PolicyKey]struct{}) {
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	for key, _ := range keys {
		delete(tbl.AutoProvPolicies, key)
	}
}

func (s *AutoProvPolicyHandler) Flush(ctx context.Context, notifyId int64) {}

// SetInstStateForCloudlet - Sets the current state of the appInstances for the cloudlet
// This gets called when a cloudlet goes offline, or comes back online
func SetInstStateForCloudlet(info *edgeproto.CloudletInfo) {
	log.DebugLog(log.DebugLevelDmereq, "SetInstStateForCloudlet called", "cloudlet", info)
	carrier := info.Key.OperatorKey.Name
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	// Update the state in the cloudlet table
	if cloudlet, foundCloudlet := tbl.Cloudlets[info.Key]; foundCloudlet {
		cloudlet.State = info.State
	} else {
		cNew := new(DmeCloudlet)
		cNew.CloudletKey = info.Key
		cNew.State = info.State
		tbl.Cloudlets[info.Key] = cNew
	}
	for _, app := range tbl.Apps {
		app.Lock()
		if c, found := app.Carriers[carrier]; found {
			for clusterInstKey, _ := range c.Insts {
				if cloudletKeyEqual(&clusterInstKey.CloudletKey, &info.Key) {
					c.Insts[clusterInstKey].cloudletState = info.State

				}
			}
		}
		app.Unlock()
	}
}

// given the carrier, update the reply if we find a cloudlet closer
// than the max distance.  Return the distance and whether or not response was updated
func findClosestForCarrier(ctx context.Context, carrierName string, key edgeproto.AppKey, loc *dme.Loc, maxDistance float64, mreply *dme.FindCloudletReply) (float64, bool) {
	tbl := DmeAppTbl
	var d float64
	var updated = false
	var found *DmeAppInst
	var cloudlet string
	tbl.RLock()
	defer tbl.RUnlock()
	app, ok := tbl.Apps[key]
	if !ok {
		return maxDistance, updated
	}

	log.DebugLog(log.DebugLevelDmereq, "Find Closest", "appkey", key, "carrierName", carrierName)

	if c, carrierFound := app.Carriers[carrierName]; carrierFound {
		for _, i := range c.Insts {
			d = DistanceBetween(*loc, i.location)
			log.DebugLog(log.DebugLevelDmereq, "found cloudlet at",
				"latitude", i.location.Latitude,
				"longitude", i.location.Longitude,
				"maxDistance", maxDistance,
				"this-dist", d)
			if d < maxDistance && IsAppInstUsable(i) {
				log.DebugLog(log.DebugLevelDmereq, "closer cloudlet", "uri", i.uri)
				updated = true
				maxDistance = d
				found = i
				mreply.Fqdn = i.uri
				mreply.Status = dme.FindCloudletReply_FIND_FOUND
				*mreply.CloudletLocation = i.location
				mreply.Ports = copyPorts(i)
				cloudlet = i.clusterInstKey.CloudletKey.Name
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
			// Update Context variable if passed
			updateContextWithCloudletDetails(ctx, cloudlet, carrierName)
		}
	}
	if app.AutoProvPolicy != nil {
		list, found := app.AutoProvPolicy.Cloudlets[carrierName]
		log.DebugLog(log.DebugLevelDmereq, "search AutoProvPolicy", "carrier", carrierName, "found", found, "list", list)
		if found {
			potentialDist := maxDistance
			var potentialCloudlet *edgeproto.AutoProvCloudlet
			var potentialClusterInstKey *edgeproto.ClusterInstKey
			for _, cl := range list {
				// make sure there's a free reservable ClusterInst
				// on the cloudlet. if cinsts exists there is at
				// least one free.
				cinstKey := tbl.FreeReservableClusterInsts.GetForCloudlet(&cl.Key)
				if cinstKey == nil {
					continue
				}
				pd := DistanceBetween(*loc, cl.Loc)
				// This will intentionally skip auto-deployed
				// AppInsts, this should only find cloudlets
				// that could, but do not have the AppInst
				// auto-deployed.
				if pd < potentialDist {
					potentialDist = pd
					potentialClusterInstKey = cinstKey
					potentialCloudlet = cl
				}
			}
			log.DebugLog(log.DebugLevelDmereq, "search AutoProvPolicy result", "potentialClusterInst", potentialClusterInstKey, "autoProvStats", autoProvStats)
			if potentialClusterInstKey != nil && autoProvStats != nil {
				autoProvStats.Increment(ctx, &app.AppKey, potentialClusterInstKey, app.AutoProvPolicy.DeployClientCount, app.AutoProvPolicy.IntervalCount)
				log.DebugLog(log.DebugLevelDmereq, "potential best cloudlet",
					"app", key.Name,
					"carrier", carrierName,
					"latitude", potentialCloudlet.Loc.Latitude,
					"longitude", potentialCloudlet.Loc.Longitude,
					"distance", potentialDist)
			}
		}
	}
	return maxDistance, updated
}

// Helper function to populate the Stats key with Cloudlet data if it's passed in the context
func updateContextWithCloudletDetails(ctx context.Context, cloudlet, carrier string) {
	statKey, ok := ctx.Value(StatKeyContextKey).(*StatKey)
	if ok {
		statKey.CloudletFound.Name = cloudlet
		statKey.CloudletFound.OperatorKey.Name = carrier
	}
}

// returns true if if the requested app allows the registered app to
// access APIs on its behalf
func requestedAppPermitsRegisteredApp(requestedApp edgeproto.AppKey, registeredApp edgeproto.AppKey) bool {
	// if the 2 apps match, allow it.  It means the client requested the same app as was registered
	var tbl *DmeApps
	tbl = DmeAppTbl

	if requestedApp == registeredApp {
		return true
	}
	if !cloudcommon.IsPlatformApp(registeredApp.DeveloperKey.Name, registeredApp.Name) {
		return false
	}
	// now find the app and see if it permits platform apps
	tbl.Lock()
	defer tbl.Unlock()
	_, ok := tbl.Apps[requestedApp]
	return ok
}

func FindCloudlet(ctx context.Context, ckey *CookieKey, mreq *dme.FindCloudletRequest, mreply *dme.FindCloudletReply) error {
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
			return grpc.Errorf(codes.PermissionDenied, "Access to requested app: Devname: %s Appname: %s AppVers: %s not allowed for the registered app: Devname: %s Appname: %s Appvers: %s",
				mreq.DevName, mreq.AppName, mreq.AppVers, appkey.DeveloperKey.Name, appkey.Name, appkey.Version)
		}
		//update the appkey to use the requested key
		appkey = reqkey
	}

	// if the app itself is a platform app, it is not returned here
	if cloudcommon.IsPlatformApp(appkey.DeveloperKey.Name, appkey.Name) {
		return nil
	}

	log.DebugLog(log.DebugLevelDmereq, "findCloudlet", "carrier", mreq.CarrierName, "app", appkey.Name, "developer", appkey.DeveloperKey.Name, "version", appkey.Version)

	// first find carrier cloudlet
	bestDistance, updated := findClosestForCarrier(ctx, mreq.CarrierName, appkey, mreq.GpsLocation, InfiniteDistance, mreply)

	if updated {
		log.DebugLog(log.DebugLevelDmereq, "found carrier cloudlet", "Fqdn", mreply.Fqdn, "distance", bestDistance)
	}

	if bestDistance > publicCloudPadding {
		paddedCarrierDistance := bestDistance - publicCloudPadding

		// look for an azure cloud closer than the carrier distance minus padding
		azDistance, updated := findClosestForCarrier(ctx, cloudcommon.OperatorAzure, appkey, mreq.GpsLocation, paddedCarrierDistance, mreply)
		if updated {
			log.DebugLog(log.DebugLevelDmereq, "found closer azure cloudlet", "Fqdn", mreply.Fqdn, "distance", azDistance)
			bestDistance = azDistance
		}

		// look for a gcp cloud closer than either the azure cloud or the carrier cloud
		maxGCPDistance := math.Min(azDistance, paddedCarrierDistance)
		gcpDistance, updated := findClosestForCarrier(ctx, cloudcommon.OperatorGCP, appkey, mreq.GpsLocation, maxGCPDistance, mreply)
		if updated {
			log.DebugLog(log.DebugLevelDmereq, "found closer gcp cloudlet", "Fqdn", mreply.Fqdn, "distance", gcpDistance)
			bestDistance = gcpDistance
		}
	}

	if mreply.Status == dme.FindCloudletReply_FIND_FOUND {
		log.DebugLog(log.DebugLevelDmereq, "findCloudlet returning FIND_FOUND, overall best cloudlet", "Fqdn", mreply.Fqdn, "distance", bestDistance)
	} else {
		log.DebugLog(log.DebugLevelDmereq, "findCloudlet returning FIND_NOTFOUND")

	}
	return nil
}

func isPublicCarrier(carriername string) bool {
	if carriername == cloudcommon.OperatorAzure ||
		carriername == cloudcommon.OperatorGCP {
		return true
	}
	return false
}

func GetFqdnList(mreq *dme.FqdnListRequest, clist *dme.FqdnListReply) {
	var tbl *DmeApps
	tbl = DmeAppTbl
	tbl.RLock()
	defer tbl.RUnlock()
	for _, a := range tbl.Apps {
		// if the app itself is a platform app, it is not returned here
		if cloudcommon.IsPlatformApp(a.AppKey.DeveloperKey.Name, a.AppKey.Name) {
			continue
		}
		if a.OfficialFqdn != "" {
			fqdns := strings.Split(a.OfficialFqdn, ",")
			aq := dme.AppFqdn{
				AppName:            a.AppKey.Name,
				DevName:            a.AppKey.DeveloperKey.Name,
				AppVers:            a.AppKey.Version,
				Fqdns:              fqdns,
				AndroidPackageName: a.AndroidPackageName}
			clist.AppFqdns = append(clist.AppFqdns, &aq)
		}
	}
	clist.Status = dme.FqdnListReply_FL_SUCCESS
}

func GetAppInstList(ckey *CookieKey, mreq *dme.AppInstListRequest, clist *dme.AppInstListReply) {
	var tbl *DmeApps
	tbl = DmeAppTbl
	foundCloudlets := make(map[edgeproto.CloudletKey]*dme.CloudletLocation)

	tbl.RLock()
	defer tbl.RUnlock()

	// find all the unique cloudlets, and the app instances for each.  the data is
	//stored as appinst->cloudlet and we need the opposite mapping.
	for _, a := range tbl.Apps {

		//if the app name or version was provided, only look for cloudlets for that app
		if (ckey.AppName != "" && ckey.AppName != a.AppKey.Name) ||
			(ckey.AppVers != "" && ckey.AppVers != a.AppKey.Version) {
			continue
		}
		for cname, c := range a.Carriers {
			//if the carrier name was provided, only look for cloudlets for that carrier, or for public cloudlets
			if mreq.CarrierName != "" && !isPublicCarrier(cname) && mreq.CarrierName != cname {
				log.DebugLog(log.DebugLevelDmereq, "skipping cloudlet, mismatched carrier", "mreq.CarrierName", mreq.CarrierName, "i.cloudletKey.OperatorKey.Name", cname)
				continue
			}
			for _, i := range c.Insts {
				// skip disabled appInstances
				if !IsAppInstUsable(i) {
					continue
				}
				cloc, exists := foundCloudlets[i.clusterInstKey.CloudletKey]
				if !exists {
					cloc = new(dme.CloudletLocation)
					var d float64

					d = DistanceBetween(*mreq.GpsLocation, i.location)
					cloc.GpsLocation = &i.location
					cloc.CarrierName = i.clusterInstKey.CloudletKey.OperatorKey.Name
					cloc.CloudletName = i.clusterInstKey.CloudletKey.Name
					cloc.Distance = d
				}
				ai := dme.Appinstance{}
				ai.AppName = a.AppKey.Name
				ai.AppVers = a.AppKey.Version
				ai.Fqdn = i.uri
				ai.Ports = copyPorts(i)
				cloc.Appinstances = append(cloc.Appinstances, &ai)
				foundCloudlets[i.clusterInstKey.CloudletKey] = cloc
			}
		}
	}
	for _, c := range foundCloudlets {
		clist.Cloudlets = append(clist.Cloudlets, c)
	}
	clist.Status = dme.AppInstListReply_AI_SUCCESS
}

func ListAppinstTbl() {
	var app *DmeApp
	var inst *DmeAppInst
	var tbl *DmeApps

	tbl = DmeAppTbl
	tbl.RLock()
	defer tbl.RUnlock()

	for a := range tbl.Apps {
		app = tbl.Apps[a]
		log.DebugLog(log.DebugLevelDmedb, "app",
			"Name", app.AppKey.Name,
			"Ver", app.AppKey.Version)
		for cname, c := range app.Carriers {
			for _, inst = range c.Insts {
				log.DebugLog(log.DebugLevelDmedb, "app",
					"Name", app.AppKey.Name,
					"carrier", cname,
					"Ver", app.AppKey.Version,
					"Latitude", inst.location.Latitude,
					"Longitude", inst.location.Longitude)
			}
		}
	}
}

func copyPorts(cappInst *DmeAppInst) []*dme.AppPort {
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

func GetAuthPublicKey(devname string, appname string, appvers string) (string, error) {
	var key edgeproto.AppKey
	var tbl *DmeApps
	tbl = DmeAppTbl

	key.DeveloperKey.Name = devname
	key.Name = appname
	key.Version = appvers
	tbl.Lock()
	defer tbl.Unlock()

	app, ok := tbl.Apps[key]
	if ok {
		return app.AuthPublicKey, nil
	}
	return "", grpc.Errorf(codes.NotFound, "app not found")
}

func AppExists(devname string, appname string, appvers string) bool {
	var key edgeproto.AppKey
	key.DeveloperKey.Name = devname
	key.Name = appname
	key.Version = appvers

	var tbl *DmeApps
	tbl = DmeAppTbl
	tbl.RLock()
	defer tbl.RUnlock()
	_, ok := tbl.Apps[key]
	return ok
}
