package dmecommon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// 1 km is considered near enough to call 2 locations equivalent
const VeryCloseDistanceKm = 1

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
	CloudletState    dme.CloudletState
	MaintenanceState dme.MaintenanceState
	// Health state of the appInst
	AppInstHealth dme.HealthCheck
	TrackedState  edgeproto.TrackedState
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
	AutoProvPolicies   map[string]*AutoProvPolicy
	Deployment         string
	// Non mapped AppPorts from App definition (used for AppOfficialFqdnReply)
	Ports []dme.AppPort
}

type DmeCloudlet struct {
	// No need for a mutex - protected under DmeApps mutex
	CloudletKey      edgeproto.CloudletKey
	State            dme.CloudletState
	MaintenanceState dme.MaintenanceState
	GpsLocation      dme.Loc
}

type AutoProvPolicy struct {
	Name              string
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
	OperatorCodes              edgeproto.OperatorCodeCache
}

type ClientToken struct {
	Location dme.Loc
	AppKey   edgeproto.AppKey
}

var DmeAppTbl *DmeApps
var Settings edgeproto.Settings

// Stats are collected per App per Cloudlet and per method name (verifylocation, etc).
type StatKey struct {
	AppKey        edgeproto.AppKey
	CloudletFound edgeproto.CloudletKey
	CellId        uint32
	Method        string
	// Only for EdgeEvents (to identify AppInsts)
	ClusterKey     edgeproto.ClusterKey
	ClusterInstOrg string
}

type EdgeEventStatKey struct {
	AppInstKey edgeproto.AppInstKey
	Metric     string
}

// This is for passing the carrier/cloudlet in the context
type StatKeyContextType string

var StatKeyContextKey = StatKeyContextType("statKey")

var EdgeEventsCookieExpiration = flag.Duration("edgeEventsCookieExpiration", time.Minute*10, "Edge Events Cookie expiration time")

// EdgeEventsHandler implementation (loaded from Plugin)
var EEHandler EdgeEventsHandler

func SetupMatchEngine(eehandler EdgeEventsHandler) {
	DmeAppTbl = new(DmeApps)
	DmeAppTbl.Apps = make(map[edgeproto.AppKey]*DmeApp)
	DmeAppTbl.Cloudlets = make(map[edgeproto.CloudletKey]*DmeCloudlet)
	DmeAppTbl.AutoProvPolicies = make(map[edgeproto.PolicyKey]*AutoProvPolicy)
	DmeAppTbl.FreeReservableClusterInsts.Init()
	edgeproto.InitOperatorCodeCache(&DmeAppTbl.OperatorCodes)
	EEHandler = eehandler
}

// AppInst state is a superset of the cloudlet state and appInst state
// Returns if this AppInstance is usable or not
func IsAppInstUsable(appInst *DmeAppInst) bool {
	if appInst == nil {
		return false
	}
	if appInst.TrackedState != edgeproto.TrackedState_READY {
		return false
	}
	if appInst.MaintenanceState == dme.MaintenanceState_UNDER_MAINTENANCE {
		return false
	}
	if appInst.CloudletState == dme.CloudletState_CLOUDLET_STATE_READY {
		return appInst.AppInstHealth == dme.HealthCheck_HEALTH_CHECK_OK || appInst.AppInstHealth == dme.HealthCheck_HEALTH_CHECK_UNKNOWN
	}
	return false
}

// TODO: Have protoc auto-generate Equal functions.
func cloudletKeyEqual(key1 *edgeproto.CloudletKey, key2 *edgeproto.CloudletKey) bool {
	return key1.GetKeyString() == key2.GetKeyString()
}

func AddApp(ctx context.Context, in *edgeproto.App) {
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.Apps[in.Key]
	if !ok {
		// Key doesn't exists
		app = new(DmeApp)
		app.Carriers = make(map[string]*DmeAppInsts)
		app.AutoProvPolicies = make(map[string]*AutoProvPolicy)
		app.AppKey = in.Key
		tbl.Apps[in.Key] = app
		log.SpanLog(ctx, log.DebugLevelDmedb, "Adding app",
			"key", in.Key,
			"package", in.AndroidPackageName,
			"OfficialFqdn", in.OfficialFqdn)
	}
	app.Lock()
	defer app.Unlock()
	app.AuthPublicKey = in.AuthPublicKey
	app.AndroidPackageName = in.AndroidPackageName
	app.OfficialFqdn = in.OfficialFqdn
	app.Deployment = in.Deployment
	ports, _ := edgeproto.ParseAppPorts(in.AccessPorts)
	app.Ports = ports
	clearAutoProvStats := []string{}
	inAP := make(map[string]struct{})
	if in.AutoProvPolicy != "" {
		inAP[in.AutoProvPolicy] = struct{}{}
	}
	for _, name := range in.AutoProvPolicies {
		inAP[name] = struct{}{}
	}
	// remove policies
	for name, _ := range app.AutoProvPolicies {
		if _, found := inAP[name]; !found {
			delete(app.AutoProvPolicies, name)
			clearAutoProvStats = append(clearAutoProvStats, name)
		}
	}
	// add policies
	for name, _ := range inAP {
		_, found := app.AutoProvPolicies[name]
		if !found {
			ppKey := edgeproto.PolicyKey{
				Organization: in.Key.Organization,
				Name:         name,
			}
			if pp, found := tbl.AutoProvPolicies[ppKey]; found {
				app.AutoProvPolicies[name] = pp
			} else {
				log.SpanLog(ctx, log.DebugLevelDmedb, "AutoProvPolicy on App not found", "app", in.Key, "policy", name)
			}
		}
	}
	for _, name := range clearAutoProvStats {
		autoProvStats.Clear(&in.Key, name)
	}
}

func AddAppInst(ctx context.Context, appInst *edgeproto.AppInst) {
	carrierName := appInst.Key.ClusterInstKey.CloudletKey.Organization

	tbl := DmeAppTbl
	appkey := appInst.Key.AppKey
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.Apps[appkey]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelDmedb, "addAppInst: app not found", "key", appInst.Key)
		return
	}
	app.Lock()
	if _, foundCarrier := app.Carriers[carrierName]; foundCarrier {
		log.SpanLog(ctx, log.DebugLevelDmedb, "carrier already exists", "carrierName", carrierName)
	} else {
		log.SpanLog(ctx, log.DebugLevelDmedb, "adding carrier for app", "carrierName", carrierName)
		app.Carriers[carrierName] = new(DmeAppInsts)
		app.Carriers[carrierName].Insts = make(map[edgeproto.ClusterInstKey]*DmeAppInst)
	}
	logMsg := "Updating app inst"
	cl, foundAppInst := app.Carriers[carrierName].Insts[appInst.Key.ClusterInstKey]
	if !foundAppInst {
		cl = new(DmeAppInst)
		cl.clusterInstKey = appInst.Key.ClusterInstKey
		app.Carriers[carrierName].Insts[cl.clusterInstKey] = cl
		logMsg = "Adding app inst"
	}
	// update existing app inst
	cl.uri = appInst.Uri
	cl.location = appInst.CloudletLoc
	cl.ports = appInst.MappedPorts
	cl.TrackedState = appInst.State
	// Check if AppInstHealth has changed
	if cl.AppInstHealth != appInst.HealthCheck {
		EEHandler.SendAppInstStateEvent(ctx, cl, appInst.Key, dme.ServerEdgeEvent_EVENT_APPINST_HEALTH)
	}
	cl.AppInstHealth = appInst.HealthCheck
	// Check if Cloudlet states have changed
	if cloudlet, foundCloudlet := tbl.Cloudlets[appInst.Key.ClusterInstKey.CloudletKey]; foundCloudlet {
		if cl.CloudletState != cloudlet.State {
			EEHandler.SendAppInstStateEvent(ctx, cl, appInst.Key, dme.ServerEdgeEvent_EVENT_CLOUDLET_STATE)
		}
		if cl.MaintenanceState != cloudlet.MaintenanceState {
			EEHandler.SendAppInstStateEvent(ctx, cl, appInst.Key, dme.ServerEdgeEvent_EVENT_CLOUDLET_MAINTENANCE)
		}
		cl.CloudletState = cloudlet.State
		cl.MaintenanceState = cloudlet.MaintenanceState
	} else {
		cl.CloudletState = dme.CloudletState_CLOUDLET_STATE_UNKNOWN
		cl.MaintenanceState = dme.MaintenanceState_NORMAL_OPERATION
	}

	log.SpanLog(ctx, log.DebugLevelDmedb, logMsg,
		"appName", app.AppKey.Name,
		"appVersion", app.AppKey.Version,
		"cloudletKey", appInst.Key.ClusterInstKey.CloudletKey,
		"uri", appInst.Uri,
		"latitude", appInst.CloudletLoc.Latitude,
		"longitude", appInst.CloudletLoc.Longitude,
		"healthState", appInst.HealthCheck)
	app.Unlock()
}

func RemoveApp(ctx context.Context, in *edgeproto.App) {
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
		for name, _ := range app.AutoProvPolicies {
			autoProvStats.Clear(&in.Key, name)
		}
	}
}

func RemoveAppInst(ctx context.Context, appInst *edgeproto.AppInst) {
	var app *DmeApp
	var tbl *DmeApps

	tbl = DmeAppTbl
	appkey := appInst.Key.AppKey
	carrierName := appInst.Key.ClusterInstKey.CloudletKey.Organization
	tbl.Lock()
	defer tbl.Unlock()
	app, ok := tbl.Apps[appkey]
	if ok {
		app.Lock()
		if c, foundCarrier := app.Carriers[carrierName]; foundCarrier {
			if cl, foundAppInst := c.Insts[appInst.Key.ClusterInstKey]; foundAppInst {
				log.SpanLog(ctx, log.DebugLevelDmereq, "removing app inst", "appinst", cl, "removed appinst health", appInst.HealthCheck)
				cl.AppInstHealth = dme.HealthCheck_HEALTH_CHECK_FAIL_SERVER_FAIL
				// Remove AppInst from edgeevents plugin
				EEHandler.SendAppInstStateEvent(ctx, cl, appInst.Key, dme.ServerEdgeEvent_EVENT_APPINST_HEALTH)
				EEHandler.RemoveAppInstKey(ctx, appInst.Key)
				delete(app.Carriers[carrierName].Insts, appInst.Key.ClusterInstKey)
				log.SpanLog(ctx, log.DebugLevelDmedb, "Removing app inst",
					"appName", appkey.Name,
					"appVersion", appkey.Version,
					"latitude", cl.location.Latitude,
					"longitude", cl.location.Longitude)
			}
			if len(app.Carriers[carrierName].Insts) == 0 {
				delete(tbl.Apps[appkey].Carriers, carrierName)
				log.SpanLog(ctx, log.DebugLevelDmedb, "Removing carrier for app",
					"carrier", carrierName,
					"appName", appkey.Name,
					"appVersion", appkey.Version)
			}
		}
		app.Unlock()
	}
}

// pruneApps removes any data that was not sent by the controller.
func PruneApps(ctx context.Context, apps map[edgeproto.AppKey]struct{}) {
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
func PruneAppInsts(ctx context.Context, appInsts map[edgeproto.AppInstKey]struct{}) {
	var key edgeproto.AppInstKey

	log.SpanLog(ctx, log.DebugLevelDmereq, "pruneAppInsts called")

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
					log.SpanLog(ctx, log.DebugLevelDmereq, "pruning app", "key", key)
					// Remove AppInst from edgeevents plugin
					EEHandler.SendAppInstStateEvent(ctx, inst, key, dme.ServerEdgeEvent_EVENT_APPINST_HEALTH)
					EEHandler.RemoveAppInstKey(ctx, key)
					delete(carr.Insts, key.ClusterInstKey)
				}
			}
			if len(carr.Insts) == 0 {
				log.SpanLog(ctx, log.DebugLevelDmereq, "pruneAppInsts delete carriers")
				delete(app.Carriers, c)
			}
		}
		app.Unlock()
	}
}

func DeleteCloudletInfo(ctx context.Context, cloudletKey *edgeproto.CloudletKey) {
	log.SpanLog(ctx, log.DebugLevelDmereq, "DeleteCloudletInfo called")
	tbl := DmeAppTbl
	carrier := cloudletKey.Organization
	tbl.Lock()
	defer tbl.Unlock()

	// If there are still appInsts on that cloudlet they should be disabled
	for _, app := range tbl.Apps {
		app.Lock()
		if c, found := app.Carriers[carrier]; found {
			for clusterInstKey, _ := range c.Insts {
				if cloudletKeyEqual(&clusterInstKey.CloudletKey, cloudletKey) {
					c.Insts[clusterInstKey].CloudletState = dme.CloudletState_CLOUDLET_STATE_NOT_PRESENT
					c.Insts[clusterInstKey].MaintenanceState = dme.MaintenanceState_NORMAL_OPERATION
				}
			}
		}
		app.Unlock()
	}
	if _, found := tbl.Cloudlets[*cloudletKey]; found {
		log.SpanLog(ctx, log.DebugLevelDmereq, "delete cloudletInfo", "key", cloudletKey)
		delete(tbl.Cloudlets, *cloudletKey)
	}
}

// Remove any Cloudlets we track that no longer exist and reset the state for the AppInsts
func PruneCloudlets(ctx context.Context, cloudlets map[edgeproto.CloudletKey]struct{}) {
	log.SpanLog(ctx, log.DebugLevelDmereq, "PruneCloudlets called")
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()

	for _, app := range tbl.Apps {
		app.Lock()
		for _, carr := range app.Carriers {
			for clusterInstKey, _ := range carr.Insts {
				if _, found := cloudlets[clusterInstKey.CloudletKey]; !found {
					carr.Insts[clusterInstKey].CloudletState = dme.CloudletState_CLOUDLET_STATE_NOT_PRESENT
					carr.Insts[clusterInstKey].MaintenanceState = dme.MaintenanceState_NORMAL_OPERATION
				}
			}
		}
		app.Unlock()
	}
	for cloudlet, _ := range cloudlets {
		if _, foundInfo := tbl.Cloudlets[cloudlet]; !foundInfo {
			log.SpanLog(ctx, log.DebugLevelDmereq, "pruning cloudletInfo", "key", cloudlet)
			delete(tbl.Cloudlets, cloudlet)
		}
	}
}

// Reset the cloudlet state for the AppInsts for any cloudlets that no longer exists
func PruneInstsCloudletState(ctx context.Context, cloudlets map[edgeproto.CloudletKey]struct{}) {
	log.SpanLog(ctx, log.DebugLevelDmereq, "PruneInstsCloudletState called")
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()

	for _, app := range tbl.Apps {
		app.Lock()
		for _, carr := range app.Carriers {
			for clusterInstKey, _ := range carr.Insts {
				if _, found := cloudlets[clusterInstKey.CloudletKey]; !found {
					carr.Insts[clusterInstKey].CloudletState = dme.CloudletState_CLOUDLET_STATE_NOT_PRESENT
				}
			}
		}
		app.Unlock()
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
	pp.Name = in.Key.Name
	pp.DeployClientCount = in.DeployClientCount
	pp.IntervalCount = in.DeployIntervalCount
	log.SpanLog(ctx, log.DebugLevelDmereq, "update AutoProvPolicy", "policy", in)
	// Organize potentional cloudlets by carrier
	pp.Cloudlets = make(map[string][]*edgeproto.AutoProvCloudlet)
	if in.Cloudlets != nil {
		for _, provCloudlet := range in.Cloudlets {
			list, found := pp.Cloudlets[provCloudlet.Key.Organization]
			if !found {
				list = make([]*edgeproto.AutoProvCloudlet, 0)
			}
			list = append(list, provCloudlet)
			pp.Cloudlets[provCloudlet.Key.Organization] = list
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
	for key, _ := range tbl.AutoProvPolicies {
		if _, found := keys[key]; !found {
			delete(tbl.AutoProvPolicies, key)
		}
	}
}

func (s *AutoProvPolicyHandler) Flush(ctx context.Context, notifyId int64) {}

// SetInstStateFromCloudlet - Sets the current maintenance state of the appInstances for the cloudlet
func SetInstStateFromCloudlet(ctx context.Context, in *edgeproto.Cloudlet) {
	log.SpanLog(ctx, log.DebugLevelDmereq, "SetInstStateFromCloudlet called", "cloudlet", in)
	carrier := in.Key.Organization
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	// Update the state in the cloudlet table
	cloudlet, foundCloudlet := tbl.Cloudlets[in.Key]
	if !foundCloudlet {
		cloudlet = new(DmeCloudlet)
		cloudlet.CloudletKey = in.Key
		tbl.Cloudlets[in.Key] = cloudlet
	}
	cloudlet.MaintenanceState = in.MaintenanceState
	cloudlet.GpsLocation = in.Location

	for _, app := range tbl.Apps {
		app.Lock()
		if c, found := app.Carriers[carrier]; found {
			for clusterInstKey, _ := range c.Insts {
				if cloudletKeyEqual(&clusterInstKey.CloudletKey, &in.Key) {
					c.Insts[clusterInstKey].MaintenanceState = in.MaintenanceState
				}
			}
		}
		app.Unlock()
	}
}

// SetInstStateFromCloudletInfo - Sets the current state of the appInstances for the cloudlet
// This gets called when a cloudlet goes offline, or comes back online
func SetInstStateFromCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) {
	log.SpanLog(ctx, log.DebugLevelDmereq, "SetInstStateFromCloudletInfo called", "cloudlet", info)
	carrier := info.Key.Organization
	tbl := DmeAppTbl
	tbl.Lock()
	defer tbl.Unlock()
	// Update the state in the cloudlet table
	cloudlet, foundCloudlet := tbl.Cloudlets[info.Key]
	if !foundCloudlet {
		// object should've been added on receipt of cloudlet object from controller
		return
	}
	cloudlet.State = info.State

	for _, app := range tbl.Apps {
		app.Lock()
		if c, found := app.Carriers[carrier]; found {
			for clusterInstKey, _ := range c.Insts {
				if cloudletKeyEqual(&clusterInstKey.CloudletKey, &info.Key) {
					c.Insts[clusterInstKey].CloudletState = info.State
					c.Insts[clusterInstKey].MaintenanceState = info.MaintenanceState
				}
			}
		}
		app.Unlock()
	}
}

// Given an AppInstKey, return the corresponding DmeCloudlet
func findDmeCloudlet(appInstKey *edgeproto.AppInstKey) DmeCloudlet {
	tbl := DmeAppTbl

	tbl.RLock()
	defer tbl.RUnlock()
	dmecloudlet, ok := tbl.Cloudlets[appInstKey.ClusterInstKey.CloudletKey]
	if !ok {
		return DmeCloudlet{}
	}
	return *dmecloudlet
}

// translateCarrierName translates carrier name (mcc+mnc) to
// mobiledgex operator name, otherwise returns original value
func translateCarrierName(carrierName string) string {
	tbl := DmeAppTbl

	cname := carrierName
	operCode := edgeproto.OperatorCode{}
	operCodeKey := edgeproto.OperatorCodeKey(carrierName)
	if tbl.OperatorCodes.Get(&operCodeKey, &operCode) {
		cname = operCode.Organization
	}
	return cname
}

type searchAppInst struct {
	loc                     *dme.Loc
	reqCarrier              string
	results                 []*foundAppInst
	appDeployment           string
	resultDist              float64
	potentialDist           float64
	potentialPolicy         *AutoProvPolicy
	potentialCloudlet       *edgeproto.AutoProvCloudlet
	potentialClusterInstKey *edgeproto.ClusterInstKey
	potentialCarrier        string
	resultLimit             int
}

type foundAppInst struct {
	distance       float64
	appInst        *DmeAppInst
	appInstCarrier string
}

// given the carrier, update the reply if we find a cloudlet closer
// than the max distance.  Return the distance and whether or not response was updated
func findBestForCarrier(ctx context.Context, carrierName string, key *edgeproto.AppKey, loc *dme.Loc, resultLimit int) []*foundAppInst {
	tbl := DmeAppTbl
	carrierName = translateCarrierName(carrierName)

	tbl.RLock()
	defer tbl.RUnlock()
	app, ok := tbl.Apps[*key]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelDmereq, "findBestForCarrier app not found", "key", *key)
		return nil
	}

	log.SpanLog(ctx, log.DebugLevelDmereq, "Find Closest", "appkey", key, "carrierName", carrierName, "loc", *loc)

	// Eventually when we have FindCloudlet policies, we should look it
	// up here and apply it to the search config.
	search := searchAppInst{
		loc:           loc,
		reqCarrier:    carrierName,
		appDeployment: app.Deployment,
		resultLimit:   resultLimit,
	}

	for cname, carrierData := range app.Carriers {
		search.searchAppInsts(ctx, cname, carrierData)
	}
	for ii, found := range search.results {
		var ipaddr net.IP
		ipaddr = found.appInst.ip
		log.SpanLog(ctx, log.DebugLevelDmereq, "best cloudlet",
			"rank", ii,
			"app", key.Name,
			"carrier", found.appInstCarrier,
			"latitude", found.appInst.location.Latitude,
			"longitude", found.appInst.location.Longitude,
			"distance", found.distance,
			"uri", found.appInst.uri,
			"IP", ipaddr.String())
	}

	if len(app.AutoProvPolicies) > 0 {
		log.SpanLog(ctx, log.DebugLevelDmereq, "search AutoProvPolicy", "found", len(search.results))
		search.resultDist = InfiniteDistance
		if len(search.results) > 0 {
			search.resultDist = search.results[0].distance
		}
		search.potentialDist = search.resultDist
		for _, policy := range app.AutoProvPolicies {
			for cname, list := range policy.Cloudlets {
				search.searchPotential(ctx, policy, cname, list)
			}
		}
		log.SpanLog(ctx, log.DebugLevelDmereq, "search AutoProvPolicy result", "potentialPolicy", search.potentialPolicy, "potentialClusterInst", search.potentialClusterInstKey, "autoProvStats", autoProvStats)
		if search.potentialClusterInstKey != nil && autoProvStats != nil {
			autoProvStats.Increment(ctx, &app.AppKey, search.potentialClusterInstKey, search.potentialPolicy)
			log.SpanLog(ctx, log.DebugLevelDmereq,
				"potential best cloudlet",
				"app", key.Name,
				"policy", search.potentialPolicy,
				"carrier", search.potentialCarrier,
				"latitude", search.potentialCloudlet.Loc.Latitude,
				"longitude", search.potentialCloudlet.Loc.Longitude,
				"distance", search.potentialDist)
		}
	}

	return search.results
}

func (s *searchAppInst) searchCarrier(carrier string) bool {
	if s.reqCarrier == "" {
		// search all carriers
		return true
	}
	// always allow public clouds
	if carrier == cloudcommon.OperatorAzure || carrier == cloudcommon.OperatorGCP || carrier == cloudcommon.OperatorAWS {
		return true
	}
	// later on we may have carrier groups or other logic,
	// but for now it's just 1-to-1.
	return s.reqCarrier == carrier
}

func (s *searchAppInst) padDistance(carrier string) float64 {
	if carrier == cloudcommon.OperatorAzure || carrier == cloudcommon.OperatorGCP || carrier == cloudcommon.OperatorAWS {
		if s.reqCarrier == "" || s.reqCarrier == carrier {
			return 0
		}
		// Carrier specified and it's not a public cloud carrier.
		// Assume it's cellular. Pad distance to favor cellular.
		return 100
	}
	return 0
}

func (s *searchAppInst) searchAppInsts(ctx context.Context, carrier string, appInsts *DmeAppInsts) {
	if !s.searchCarrier(carrier) {
		return
	}
	for _, i := range appInsts.Insts {
		d := DistanceBetween(*s.loc, i.location) + s.padDistance(carrier)
		usable := IsAppInstUsable(i)
		log.SpanLog(ctx, log.DebugLevelDmereq, "found cloudlet at",
			"carrier", carrier,
			"latitude", i.location.Latitude,
			"longitude", i.location.Longitude,
			"this-dist", d,
			"usable", usable)
		if !usable {
			continue
		}
		found := &foundAppInst{
			distance:       d,
			appInst:        i,
			appInstCarrier: carrier,
		}
		s.insertResult(found)
	}
}

func (s *searchAppInst) insertResult(found *foundAppInst) bool {
	inserted := false
	for ii, ai := range s.results {
		if s.veryClose(found, ai) {
			if rand.Float64() > .5 {
				continue
			}
		} else {
			if !s.less(found, ai) {
				continue
			}
		}
		// insert before
		if ii < len(s.results) {
			count := len(s.results)
			if count >= s.resultLimit {
				// drop last to prevent more than resultLimit
				count--
			}
			// shift out later entries (duplicates ii)
			s.results = append(s.results[:ii+1], s.results[ii:count]...)
		}
		// replace
		s.results[ii] = found
		inserted = true
		break
	}
	if !inserted && len(s.results) < s.resultLimit {
		s.results = append(s.results, found)
		inserted = true
	}
	return inserted
}

func (s *searchAppInst) less(f1, f2 *foundAppInst) bool {
	// For now we sort by distance, but in the future
	// we may sort by latency or some other metric.
	return f1.distance < f2.distance
}

func (s *searchAppInst) veryClose(f1, f2 *foundAppInst) bool {
	if f1.distance > f2.distance {
		return f1.distance-f2.distance < VeryCloseDistanceKm
	}
	return f2.distance-f1.distance < VeryCloseDistanceKm
}

func (s *searchAppInst) searchPotential(ctx context.Context, policy *AutoProvPolicy, carrier string, list []*edgeproto.AutoProvCloudlet) {
	if !s.searchCarrier(carrier) {
		return
	}
	if policy.DeployClientCount == 0 {
		return
	}
	log.SpanLog(ctx, log.DebugLevelDmereq, "search AutoProvPolicy list", "policy", policy.Name, "carrier", carrier, "list", list)

	for _, cl := range list {
		// make sure there's a free reservable ClusterInst
		// on the cloudlet. if cinsts exists there is at
		// least one free.
		cinstKey := DmeAppTbl.FreeReservableClusterInsts.GetForCloudlet(&cl.Key, s.appDeployment, cloudcommon.AppInstToClusterDeployment)
		if cinstKey == nil {
			continue
		}
		pd := DistanceBetween(*s.loc, cl.Loc) + s.padDistance(carrier)
		// This will intentionally skip auto-deployed
		// AppInsts, this should only find cloudlets
		// that could, but do not have the AppInst
		// auto-deployed.
		if pd >= s.resultDist || pd > s.potentialDist {
			continue
		}
		if pd == s.potentialDist && s.potentialPolicy != nil {
			// Multiple policies for the same Cloudlet.
			// Always prefer immediate provisioning policy.
			// We don't actually care which policy it is
			// for a non-immediate policy.
			if intervalCount(policy) > intervalCount(s.potentialPolicy) {
				continue
			}
			if policy.DeployClientCount > s.potentialPolicy.DeployClientCount {
				continue
			}
		}
		// new best
		s.potentialPolicy = policy
		s.potentialDist = pd
		s.potentialClusterInstKey = cinstKey
		s.potentialCloudlet = cl
		s.potentialCarrier = carrier
	}
}

func intervalCount(policy *AutoProvPolicy) uint32 {
	// normalize interval count because both 0 and 1
	// currently mean the same thing
	if policy.IntervalCount == 0 {
		return 1
	}
	return policy.IntervalCount
}

// Helper function to populate the Stats key with Cloudlet data if it's passed in the context
func updateContextWithCloudletDetails(ctx context.Context, cloudlet, carrier string) {
	statKey, ok := ctx.Value(StatKeyContextKey).(*StatKey)
	if ok {
		statKey.CloudletFound.Name = cloudlet
		statKey.CloudletFound.Organization = carrier
	}
}

func FindCloudlet(ctx context.Context, appkey *edgeproto.AppKey, carrier string, loc *dme.Loc, mreply *dme.FindCloudletReply, edgeEventsCookieExpiration *time.Duration) error {
	mreply.Status = dme.FindCloudletReply_FIND_NOTFOUND
	mreply.CloudletLocation = &dme.Loc{}

	// if the app itself is a platform app, it is not returned here
	if cloudcommon.IsPlatformApp(appkey.Organization, appkey.Name) {
		return nil
	}

	log.SpanLog(ctx, log.DebugLevelDmereq, "findCloudlet", "carrier", carrier, "app", appkey.Name, "developer", appkey.Organization, "version", appkey.Version)

	// first find carrier cloudlet
	list := findBestForCarrier(ctx, carrier, appkey, loc, 1)
	if len(list) > 0 {
		best := list[0]
		mreply.Fqdn = best.appInst.uri
		mreply.Status = dme.FindCloudletReply_FIND_FOUND
		*mreply.CloudletLocation = best.appInst.location
		mreply.Ports = copyPorts(best.appInst.ports)
		cloudlet := best.appInst.clusterInstKey.CloudletKey.Name

		key := EdgeEventsCookieKey{
			ClusterOrg:   best.appInst.clusterInstKey.Organization,
			ClusterName:  best.appInst.clusterInstKey.ClusterKey.Name,
			CloudletOrg:  best.appInst.clusterInstKey.CloudletKey.Organization,
			CloudletName: best.appInst.clusterInstKey.CloudletKey.Name,
		}
		eecookie, _ := GenerateEdgeEventsCookie(&key, ctx, edgeEventsCookieExpiration)
		mreply.EdgeEventsCookie = eecookie
		// Update Context variable if passed
		updateContextWithCloudletDetails(ctx, cloudlet, best.appInstCarrier)
		log.SpanLog(ctx, log.DebugLevelDmereq, "findCloudlet returning FIND_FOUND, overall best cloudlet", "Fqdn", mreply.Fqdn, "distance", best.distance)
	} else {
		log.SpanLog(ctx, log.DebugLevelDmereq, "findCloudlet returning FIND_NOTFOUND")
	}
	return nil
}

func GetFqdnList(mreq *dme.FqdnListRequest, clist *dme.FqdnListReply) {
	var tbl *DmeApps
	tbl = DmeAppTbl
	tbl.RLock()
	defer tbl.RUnlock()
	for _, a := range tbl.Apps {
		// if the app itself is a platform app, it is not returned here
		if cloudcommon.IsPlatformApp(a.AppKey.Organization, a.AppKey.Name) {
			continue
		}
		if a.OfficialFqdn != "" {
			fqdns := strings.Split(a.OfficialFqdn, ",")
			aq := dme.AppFqdn{
				AppName:            a.AppKey.Name,
				OrgName:            a.AppKey.Organization,
				AppVers:            a.AppKey.Version,
				Fqdns:              fqdns,
				AndroidPackageName: a.AndroidPackageName}
			clist.AppFqdns = append(clist.AppFqdns, &aq)
		}
	}
	clist.Status = dme.FqdnListReply_FL_SUCCESS
}

func GetClientDataFromToken(token string) (*ClientToken, error) {
	var clientToken ClientToken
	tokstr, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return &clientToken, fmt.Errorf("unable to decode token: %v", err)
	}
	err = json.Unmarshal([]byte(tokstr), &clientToken)
	if err != nil {
		return &clientToken, fmt.Errorf("unable to unmarshal token: %s - %v", tokstr, err)
	}
	return &clientToken, nil
}

func GetAppOfficialFqdn(ctx context.Context, ckey *CookieKey, mreq *dme.AppOfficialFqdnRequest, repl *dme.AppOfficialFqdnReply) {
	repl.Status = dme.AppOfficialFqdnReply_AOF_FAIL
	var tbl *DmeApps
	tbl = DmeAppTbl
	var appkey edgeproto.AppKey
	appkey.Organization = ckey.OrgName
	appkey.Name = ckey.AppName
	appkey.Version = ckey.AppVers
	tbl.RLock()
	defer tbl.RUnlock()
	app, ok := tbl.Apps[appkey]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelDmereq, "GetAppOfficialFqdn cannot find app", "appkey", appkey)
		return
	}
	repl.AppOfficialFqdn = tbl.Apps[appkey].OfficialFqdn
	if repl.AppOfficialFqdn == "" {
		log.SpanLog(ctx, log.DebugLevelDmereq, "GetAppOfficialFqdn FQDN is empty", "appkey", appkey)
		return
	}
	if mreq.GpsLocation == nil {
		log.SpanLog(ctx, log.DebugLevelDmereq, "missing location in request")
		return
	}
	var token ClientToken
	token.Location = *mreq.GpsLocation
	token.AppKey = appkey
	byt, err := json.Marshal(token)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelDmereq, "Unable to marshal GPS location", "mreq.GpsLocation", mreq.GpsLocation, "err", err)
		return
	}
	repl.ClientToken = base64.StdEncoding.EncodeToString(byt)

	repl.Status = dme.AppOfficialFqdnReply_AOF_SUCCESS
	repl.Ports = copyPorts(app.Ports)
}

func GetAppInstList(ctx context.Context, ckey *CookieKey, mreq *dme.AppInstListRequest, clist *dme.AppInstListReply) {
	var appkey edgeproto.AppKey
	appkey.Organization = ckey.OrgName
	appkey.Name = ckey.AppName
	appkey.Version = ckey.AppVers

	foundCloudlets := make(map[edgeproto.CloudletKey]*dme.CloudletLocation)
	resultLimit := int(mreq.Limit)
	if resultLimit == 0 {
		resultLimit = 3
	}
	list := findBestForCarrier(ctx, mreq.CarrierName, &appkey, mreq.GpsLocation, resultLimit)

	// group by cloudlet but preserve order.
	// assumes all AppInsts on a Cloudlet are at the same distance.
	for _, found := range list {
		cloc, exists := foundCloudlets[found.appInst.clusterInstKey.CloudletKey]
		if !exists {
			cloc = new(dme.CloudletLocation)

			cloc.GpsLocation = &found.appInst.location
			cloc.CarrierName = found.appInst.clusterInstKey.CloudletKey.Organization
			cloc.CloudletName = found.appInst.clusterInstKey.CloudletKey.Name
			cloc.Distance = found.distance
			foundCloudlets[found.appInst.clusterInstKey.CloudletKey] = cloc
			clist.Cloudlets = append(clist.Cloudlets, cloc)
		}
		ai := dme.Appinstance{}
		ai.AppName = appkey.Name
		ai.AppVers = appkey.Version
		ai.OrgName = appkey.Organization
		ai.Fqdn = found.appInst.uri
		ai.Ports = copyPorts(found.appInst.ports)
		cloc.Appinstances = append(cloc.Appinstances, &ai)
	}
	clist.Status = dme.AppInstListReply_AI_SUCCESS
}

func StreamEdgeEvent(ctx context.Context, svr dme.MatchEngineApi_StreamEdgeEventServer) (reterr error) {
	// Initialize vars used in persistent connection
	var appInstKey *edgeproto.AppInstKey
	var sessionCookie string
	var sessionCookieKey *CookieKey
	var edgeEventsCookieKey *EdgeEventsCookieKey
	// Intialize send function to be passed to plugin functions
	sendFunc := func(event *dme.ServerEdgeEvent) {
		err := svr.Send(event)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelDmereq, "error sending event to client", "error", err, "eventType", event.EventType)
		}
	}
	// Receive first msg from stream
	initMsg, err := svr.Recv()
	if err != nil && err != io.EOF {
		return err
	}
	if initMsg == nil {
		return fmt.Errorf("Initial message is nil")
	}
	// On first message:
	// Verify Session Cookie and EdgeEvents Cookie
	// Then add connection, client, and appinst to Plugin hashmap
	if initMsg.EventType == dme.ClientEdgeEvent_EVENT_INIT_CONNECTION {
		// Verify session cookie
		sessionCookie = initMsg.SessionCookie
		sessionCookieKey, err = VerifyCookie(ctx, sessionCookie)
		log.SpanLog(ctx, log.DebugLevelDmereq, "EdgeEvent VerifyCookie result", "ckey", sessionCookieKey, "err", err)
		if err != nil {
			return grpc.Errorf(codes.Unauthenticated, err.Error())
		}
		ctx = NewCookieContext(ctx, sessionCookieKey)
		// Verify EdgeEventsCookieKey
		edgeEventsCookieKey, err = VerifyEdgeEventsCookie(ctx, initMsg.EdgeEventsCookie)
		log.SpanLog(ctx, log.DebugLevelDmereq, "EdgeEvent VerifyEdgeEventsCookie result", "key", edgeEventsCookieKey, "err", err)
		if err != nil {
			return grpc.Errorf(codes.Unauthenticated, err.Error())
		}
		// Create AppInstKey from SessionCookie and EdgeEventsCookie
		appInstKey = &edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				Organization: sessionCookieKey.OrgName,
				Name:         sessionCookieKey.AppName,
				Version:      sessionCookieKey.AppVers,
			},
			ClusterInstKey: edgeproto.ClusterInstKey{
				ClusterKey: edgeproto.ClusterKey{
					Name: edgeEventsCookieKey.ClusterName,
				},
				CloudletKey: edgeproto.CloudletKey{
					Organization: edgeEventsCookieKey.CloudletOrg,
					Name:         edgeEventsCookieKey.CloudletName,
				},
				Organization: edgeEventsCookieKey.ClusterOrg,
			},
		}
		// Add Client to edgeevents plugin
		EEHandler.AddClientKey(ctx, *appInstKey, *sessionCookieKey, sendFunc)
		// Send successful init response
		initServerEdgeEvent := new(dme.ServerEdgeEvent)
		initServerEdgeEvent.EventType = dme.ServerEdgeEvent_EVENT_INIT_CONNECTION
		EEHandler.SendEdgeEventToClient(ctx, initServerEdgeEvent, *appInstKey, *sessionCookieKey)
	} else {
		return fmt.Errorf("First message should have event type EVENT_INIT_CONNECTION")
	}

	// Loop while persistent connection is up
loop:
	for {
		// Receive data from stream
		cupdate, err := svr.Recv()
		ctx = svr.Context()
		// Check receive errors
		if err != nil && err != io.EOF {
			log.SpanLog(ctx, log.DebugLevelDmereq, "error on receive", "error", err)
			if strings.Contains(err.Error(), "rpc error") {
				reterr = err
				break
			}
		}
		log.SpanLog(ctx, log.DebugLevelDmereq, "Received Edge Event from client", "ClientEdgeEvent", cupdate, "context", ctx)
		if cupdate == nil {
			log.SpanLog(ctx, log.DebugLevelDmereq, "ClientEdgeEvent is nil. Ending connection")
			reterr = fmt.Errorf("ClientEdgeEvent is nil. Ending connection")
			break
		}
		// Handle Different Client events
		switch cupdate.EventType {
		case dme.ClientEdgeEvent_EVENT_TERMINATE_CONNECTION:
			// Client initiated termination
			log.SpanLog(ctx, log.DebugLevelDmereq, "Client initiated termination of persistent connection")
			break loop
		case dme.ClientEdgeEvent_EVENT_LATENCY_SAMPLES:
			// Client sent latency samples to be processed
			stats, err := EEHandler.ProcessLatencySamples(ctx, *appInstKey, *sessionCookieKey, cupdate.Samples)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelDmereq, "ClientEdgeEvent latency unable to process latency samples", "err", err)
				reterr = err
				break loop
			}
			deviceInfo := dme.DeviceInfo{}
			if cupdate.DeviceInfo != nil {
				deviceInfo = *cupdate.DeviceInfo
			}
			RecordAppInstLatencyStatCall(cupdate.GpsLocation, appInstKey, sessionCookieKey, edgeEventsCookieKey, stats, cupdate.CarrierName, deviceInfo)
		case dme.ClientEdgeEvent_EVENT_LOCATION_UPDATE:
			// Client updated gps location
			// Gps location stats update
			deviceInfo := dme.DeviceInfo{}
			if cupdate.DeviceInfo != nil {
				deviceInfo = *cupdate.DeviceInfo
			}
			RecordGpsLocationStatCall(cupdate.GpsLocation, appInstKey, sessionCookieKey, cupdate.CarrierName, deviceInfo)
			// Check if there is a better cloudlet based on location update
			fcreply := new(dme.FindCloudletReply)
			err = FindCloudlet(ctx, &appInstKey.AppKey, cupdate.CarrierName, cupdate.GpsLocation, fcreply, EdgeEventsCookieExpiration)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelDmereq, "Error trying to find closer cloudlet", "err", err)
				continue
			}
			newEECookieKey, err := VerifyEdgeEventsCookie(ctx, fcreply.EdgeEventsCookie)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelDmereq, "Error trying verify new cloudlet edgeeventscookie", "err", err)
				continue
			}
			// Check if new appinst is different from current
			if fcreply.Status == dme.FindCloudletReply_FIND_FOUND && !IsTheSameCluster(newEECookieKey, edgeEventsCookieKey) {
				// Send successful init response
				newCloudletEdgeEvent := new(dme.ServerEdgeEvent)
				newCloudletEdgeEvent.EventType = dme.ServerEdgeEvent_EVENT_CLOUDLET_UPDATE
				newCloudletEdgeEvent.NewCloudlet = fcreply
				EEHandler.SendEdgeEventToClient(ctx, newCloudletEdgeEvent, *appInstKey, *sessionCookieKey)
			}
		case dme.ClientEdgeEvent_EVENT_CUSTOM_EVENT:
			// Client sent custom stat to be stored
			RecordCustomStatCall(appInstKey, sessionCookieKey, cupdate.CustomEvent, cupdate.Samples)
		default:
			// Unknown client event
			log.SpanLog(ctx, log.DebugLevelDmereq, "Received unknown event type", "eventtype", cupdate.EventType)
		}
	}
	// Remove Client from edgeevents plugin
	EEHandler.RemoveClientKey(ctx, *appInstKey, *sessionCookieKey)
	return reterr
}

func ListAppinstTbl(ctx context.Context) {
	var app *DmeApp
	var inst *DmeAppInst
	var tbl *DmeApps

	tbl = DmeAppTbl
	tbl.RLock()
	defer tbl.RUnlock()

	for a := range tbl.Apps {
		app = tbl.Apps[a]
		log.SpanLog(ctx, log.DebugLevelDmedb, "app",
			"Name", app.AppKey.Name,
			"Ver", app.AppKey.Version)
		for cname, c := range app.Carriers {
			for _, inst = range c.Insts {
				log.SpanLog(ctx, log.DebugLevelDmedb, "app",
					"Name", app.AppKey.Name,
					"carrier", cname,
					"Ver", app.AppKey.Version,
					"Latitude", inst.location.Latitude,
					"Longitude", inst.location.Longitude)
			}
		}
	}
}

func copyPorts(cports []dme.AppPort) []*dme.AppPort {
	if cports == nil || len(cports) == 0 {
		return nil
	}
	ports := make([]*dme.AppPort, len(cports))
	for ii, _ := range cports {
		p := dme.AppPort{}
		p = cports[ii]
		ports[ii] = &p
	}
	return ports
}

func GetAuthPublicKey(orgname string, appname string, appvers string) (string, error) {
	var key edgeproto.AppKey
	var tbl *DmeApps
	tbl = DmeAppTbl

	key.Organization = orgname
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

func AppExists(orgname string, appname string, appvers string) bool {
	var key edgeproto.AppKey
	key.Organization = orgname
	key.Name = appname
	key.Version = appvers

	var tbl *DmeApps
	tbl = DmeAppTbl
	tbl.RLock()
	defer tbl.RUnlock()
	_, ok := tbl.Apps[key]
	return ok
}

func SettingsUpdated(ctx context.Context, old *edgeproto.Settings, new *edgeproto.Settings) {
	autoProvStats.UpdateSettings(new.AutoDeployIntervalSec)
	Stats.UpdateSettings(time.Duration(new.DmeApiMetricsCollectionInterval))
	EEStats.UpdateSettings(time.Duration(new.PersistentConnectionMetricsCollectionInterval))
}
