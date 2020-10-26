package dmecommon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"io"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
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
	CloudletState    edgeproto.CloudletState
	MaintenanceState edgeproto.MaintenanceState
	// Health state of the appInst
	AppInstHealth edgeproto.HealthCheck
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
	State            edgeproto.CloudletState
	MaintenanceState edgeproto.MaintenanceState
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
}

// This is for passing the carrier/cloudlet in the context
type StatKeyContextType string

var StatKeyContextKey = StatKeyContextType("statKey")

// Struct to monitor EdgeEvent receive and send goroutines
type EdgeEventPersistentMgr struct {
	Mux        util.Mutex
	Svr        *dme.MatchEngineApi_SendEdgeEventServer
	Err        error
	Terminated chan bool
	Msg        string
	Firstrcv   bool
}

// Edge Events Plugin functions
var AddClient func(ctx context.Context, key *edgeproto.AppInstKey, p *peer.Peer, mgr *EdgeEventPersistentMgr)
var RemoveClient func(ctx context.Context, key *edgeproto.AppInstKey, p *peer.Peer)
var MeasureLatency func(ctx context.Context, appInst *DmeAppInst, key *edgeproto.AppInstKey)
var UpdateAppInstState func(ctx context.Context, appInst *DmeAppInst, key *edgeproto.AppInstKey)
var UpdateClients func(ctx context.Context, serverEdgeEvent *dme.ServerEdgeEvent, mgr *EdgeEventPersistentMgr)

func SetupMatchEngine() {
	DmeAppTbl = new(DmeApps)
	DmeAppTbl.Apps = make(map[edgeproto.AppKey]*DmeApp)
	DmeAppTbl.Cloudlets = make(map[edgeproto.CloudletKey]*DmeCloudlet)
	DmeAppTbl.AutoProvPolicies = make(map[edgeproto.PolicyKey]*AutoProvPolicy)
	DmeAppTbl.FreeReservableClusterInsts.Init()
	edgeproto.InitOperatorCodeCache(&DmeAppTbl.OperatorCodes)
}

// AppInst state is a superset of the cloudlet state and appInst state
// Returns if this AppInstance is usable or not
func IsAppInstUsable(appInst *DmeAppInst) bool {
	if appInst == nil {
		return false
	}
	if appInst.MaintenanceState == edgeproto.MaintenanceState_CRM_UNDER_MAINTENANCE {
		return false
	}
	if appInst.CloudletState == edgeproto.CloudletState_CLOUDLET_STATE_READY {
		return appInst.AppInstHealth == edgeproto.HealthCheck_HEALTH_CHECK_OK || appInst.AppInstHealth == edgeproto.HealthCheck_HEALTH_CHECK_UNKNOWN
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

	appInstCheckChanged := false
	if cl.AppInstHealth != appInst.HealthCheck {
		appInstCheckChanged = true
	}
	cl.AppInstHealth = appInst.HealthCheck
	if cloudlet, foundCloudlet := tbl.Cloudlets[appInst.Key.ClusterInstKey.CloudletKey]; foundCloudlet {
		if cl.CloudletState != cloudlet.State || cl.MaintenanceState != cloudlet.MaintenanceState {
			appInstCheckChanged = true
		}
		cl.CloudletState = cloudlet.State
		cl.MaintenanceState = cloudlet.MaintenanceState
	} else {
		cl.CloudletState = edgeproto.CloudletState_CLOUDLET_STATE_UNKNOWN
		cl.MaintenanceState = edgeproto.MaintenanceState_NORMAL_OPERATION
	}

	// Check for changed in Cloudlet State and Health Check
	if appInst.MeasureLatency {
		log.SpanLog(ctx, log.DebugLevelDmereq, "AppInst update in DME. Request to Measure Latency")
		MeasureLatency(ctx, cl, &appInst.Key)
	} else if appInstCheckChanged {
		log.SpanLog(ctx, log.DebugLevelDmereq, "AppInst update in DME")
		UpdateAppInstState(ctx, cl, &appInst.Key)
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

// TODO: When an appinst or cloudlet is deleted, should send an edge event
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
				cl.AppInstHealth = edgeproto.HealthCheck_HEALTH_CHECK_FAIL_SERVER_FAIL
				UpdateAppInstState(ctx, cl, &appInst.Key)

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
				if cloudletKeyEqual(&clusterInstKey.CloudletKey, &info.Key) {
					c.Insts[clusterInstKey].CloudletState = edgeproto.CloudletState_CLOUDLET_STATE_NOT_PRESENT
					c.Insts[clusterInstKey].MaintenanceState = edgeproto.MaintenanceState_NORMAL_OPERATION
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
					carr.Insts[clusterInstKey].CloudletState = edgeproto.CloudletState_CLOUDLET_STATE_NOT_PRESENT
					carr.Insts[clusterInstKey].MaintenanceState = edgeproto.MaintenanceState_NORMAL_OPERATION
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
					carr.Insts[clusterInstKey].cloudletState = edgeproto.CloudletState_CLOUDLET_STATE_NOT_PRESENT
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

	for _, app := range tbl.Apps {
		app.Lock()
		if c, found := app.Carriers[carrier]; found {
			for clusterInstKey, _ := range c.Insts {
				if cloudletKeyEqual(&clusterInstKey.CloudletKey, &in.Key) {
					c.Insts[clusterInstKey].maintenanceState = in.MaintenanceState
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

func FindCloudlet(ctx context.Context, appkey *edgeproto.AppKey, carrier string, loc *dme.Loc, mreply *dme.FindCloudletReply, cookieExpiration *time.Duration) error {
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
		cookie, _ := GenerateEdgeEventsCookie(&key, ctx, cookieExpiration)
		mreply.EdgeEventsCookie = cookie

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

func SendEdgeEvent(ctx context.Context, svr *dme.MatchEngineApi_SendEdgeEventServer, cancel context.CancelFunc) error {
	mgr := new(EdgeEventPersistentMgr)
	mgr.Svr = svr
	mgr.Firstrcv = true
	var appInstKey *edgeproto.AppInstKey
	var p *peer.Peer

	// latency vars
	var latencies []float64
	numPings := 0

	for {
		//check if EdgeEvent persistent connection has been cancelled
		select {
		case <-ctx.Done():
			log.SpanLog(ctx, log.DebugLevelDmereq, "handle client edge event cancelled")
			mgr.Err = ctx.Err()
		case <-mgr.Terminated:
			if mgr.Err != nil {
				log.SpanLog(ctx, log.DebugLevelDmereq, "terminating persistent connection in mgr terminated", "error", mgr.Err)
				break
			}
			mgr.Err = ctx.Err()
			break
		default:
		}

		//receive data from stream
		cupdate, err := (*mgr.Svr).Recv()

		//check receive errors
		if err != nil && err != io.EOF {
			log.SpanLog(ctx, log.DebugLevelDmereq, "error on receive", "error", err)
			if strings.Contains(err.Error(), "rpc error") {
				mgr.Err = err
				mgr.Terminated <- true
				break
			}
		}
		log.SpanLog(ctx, log.DebugLevelDmereq, "Received Edge Event from client", "ClientEdgeEvent", cupdate)

		// Calculate RTT latency and send back to client if message is a pong
		if cupdate.Pong {
			numPings++
			currTimestamp := cloudcommon.TimeToTimestamp(time.Now())
			oldTimestamp := cupdate.Timestamp
			rtt := (currTimestamp.Nanos - oldTimestamp.Nanos) / 1000000
			log.SpanLog(ctx, log.DebugLevelDmereq, "received pong from client", "old timestamp", oldTimestamp.Nanos, "new timestamp", currTimestamp.Nanos, "rtt", rtt)
			latencies = append(latencies, float64(rtt))

			if numPings == 5 {
				serverEdgeEvent := new(dme.ServerEdgeEvent)
				serverEdgeEvent.LatencyRequested = true
				latency := new(dme.Latency)

				// calc avg, min, max
				sum := 0.0
				for _, val := range latencies {
					if val < latency.Min || latency.Min == 0.0 {
						latency.Min = val
					}
					if val > latency.Max || latency.Max == 0.0 {
						latency.Max = val
					}
					sum += val
				}
				latency.Avg = sum / 5

				// calc std dev
				squaredDiffs := 0.0
				for _, val := range latencies {
					diff := val - latency.Avg
					squaredDiffs += diff * diff
				}
				latency.StdDev = math.Sqrt(squaredDiffs / 5)

				serverEdgeEvent.Latency = latency
				UpdateClients(ctx, serverEdgeEvent, mgr)

				// reset
				numPings = 0
				latencies = latencies[:0]
			}
			continue
		}

		//check if client sent a request to terminate persistent connection
		if cupdate.Terminate == true {
			log.SpanLog(ctx, log.DebugLevelDmereq, "Client initiated termination of persistent connection")
			mgr.Terminated <- true
			break
		}

		// Add connection to Plugin hashmaps
		if mgr.Firstrcv {
			// Grab CookieKey
			key, err := VerifyCookie(ctx, cupdate.SessionCookie)
			log.SpanLog(ctx, log.DebugLevelDmereq, "VerifyCookie result", "key", key, "err", err)
			if err != nil {
				mgr.Err = grpc.Errorf(codes.Unauthenticated, err.Error())
				break
			}

			// Grab EdgeEventsCookieKey
			eekey, err := VerifyEdgeEventsCookie(ctx, cupdate.EeCookie)
			log.SpanLog(ctx, log.DebugLevelDmereq, "EdgeEvent VerifyEdgeEventsCookie result", "key", eekey, "err", err)
			if err != nil {
				mgr.Err = grpc.Errorf(codes.Unauthenticated, err.Error())
				break
			}

			appInstKey = &edgeproto.AppInstKey{
				AppKey: edgeproto.AppKey{
					Organization: key.OrgName,
					Name:         key.AppName,
					Version:      key.AppVers,
				},
				ClusterInstKey: edgeproto.ClusterInstKey{
					ClusterKey: edgeproto.ClusterKey{
						Name: eekey.ClusterName,
					},
					CloudletKey: edgeproto.CloudletKey{
						Organization: eekey.CloudletOrg,
						Name:         eekey.CloudletName,
					},
					Organization: eekey.ClusterOrg,
				},
			}

			p, ok := peer.FromContext(ctx)
			if !ok {
				mgr.Err = errors.New("SendEdgeEventServer unable to get peer context")
				break
			}
			log.SpanLog(ctx, log.DebugLevelDmereq, "got peer ip", "peer", p.Addr)

			AddClient(ctx, appInstKey, p, mgr)

			mgr.Firstrcv = false
		}
	}
	RemoveClient(ctx, appInstKey, p)

	return mgr.Err
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
}
