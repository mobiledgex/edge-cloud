package dmecommon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/ratelimit"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// 1 km is considered near enough to call 2 locations equivalent
const VeryCloseDistanceKm = 1

// AppInst within a cloudlet
type DmeAppInst struct {
	// Virtual clusterInstKey
	virtualClusterInstKey edgeproto.VirtualClusterInstKey
	// Unique identifier key for the clusterInst
	clusterInstKey edgeproto.ClusterInstKey
	// URI to connect to app inst in this cloudlet
	Uri string
	// Ip to connect o app inst in this cloudlet (XXX why is this needed?)
	ip []byte
	// Location of the cloudlet site (lat, long?)
	Location dme.Loc
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

type DmeAppInstState struct {
	CloudletState    dme.CloudletState
	MaintenanceState dme.MaintenanceState
	AppInstHealth    dme.HealthCheck
}

type DmeAppInsts struct {
	Insts         map[edgeproto.VirtualClusterInstKey]*DmeAppInst
	AllianceInsts map[edgeproto.VirtualClusterInstKey]*DmeAppInst
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
	DefaultFlavor      string
	// Non mapped AppPorts from App definition (used for AppOfficialFqdnReply)
	Ports []dme.AppPort
}

type DmeCloudlet struct {
	// No need for a mutex - protected under DmeApps mutex
	CloudletKey      edgeproto.CloudletKey
	State            dme.CloudletState
	MaintenanceState dme.MaintenanceState
	GpsLocation      dme.Loc
	AllianceCarriers map[string]struct{}
	AppInstKeys      map[edgeproto.AppInstKey]struct{}
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

// EdgeEventsHandler implementation (loaded from Plugin)
var EEHandler EdgeEventsHandler

// RateLimitManager
var RateLimitMgr *ratelimit.RateLimitManager

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
	return AreStatesUsable(appInst.MaintenanceState, appInst.CloudletState, appInst.AppInstHealth)
}

// Checks dme proto states for an appinst or cloudlet (Maintenance, Cloudlet, and AppInstHealth states)
func AreStatesUsable(maintenanceState dme.MaintenanceState, cloudletState dme.CloudletState, appInstHealth dme.HealthCheck) bool {
	if maintenanceState == dme.MaintenanceState_UNDER_MAINTENANCE {
		return false
	}
	if cloudletState == dme.CloudletState_CLOUDLET_STATE_READY {
		return appInstHealth == dme.HealthCheck_HEALTH_CHECK_OK || appInstHealth == dme.HealthCheck_HEALTH_CHECK_UNKNOWN
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
	app.DefaultFlavor = in.DefaultFlavor.Name
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
	defer app.Unlock()
	insts, foundCarrier := app.Carriers[carrierName]
	if foundCarrier {
		log.SpanLog(ctx, log.DebugLevelDmedb, "carrier already exists", "carrierName", carrierName)
	} else {
		log.SpanLog(ctx, log.DebugLevelDmedb, "adding carrier for app", "carrierName", carrierName)
		insts = newDmeAppInsts()
		app.Carriers[carrierName] = insts
	}
	logMsg := "Updating app inst"
	cl, foundAppInst := insts.Insts[appInst.Key.ClusterInstKey]
	sendAvailableAppInst := false
	if !foundAppInst {
		sendAvailableAppInst = true // if this is a new appinst, send msg to clients that are closer that this appinst is available
		cl = new(DmeAppInst)
		cl.virtualClusterInstKey = appInst.Key.ClusterInstKey
		cl.clusterInstKey = *appInst.ClusterInstKey()
		insts.Insts[cl.virtualClusterInstKey] = cl
		logMsg = "Adding app inst"
	}
	// update existing app inst
	cl.Uri = appInst.Uri
	cl.Location = appInst.CloudletLoc
	cl.ports = appInst.MappedPorts
	cl.TrackedState = appInst.State
	// Check if AppInstHealth has changed
	if cl.AppInstHealth != appInst.HealthCheck && foundAppInst {
		cl.AppInstHealth = appInst.HealthCheck
		if cl.AppInstHealth == dme.HealthCheck_HEALTH_CHECK_OK {
			// send msg to clients that are not on this appinst but are closer to this appinst that it is available
			sendAvailableAppInst = true
		} else {
			// send msg to clients on this appinst that it is not available
			appinstState := &DmeAppInstState{
				AppInstHealth: cl.AppInstHealth,
			}
			go EEHandler.SendAppInstStateEdgeEvent(ctx, appinstState, appInst.Key, dme.ServerEdgeEvent_EVENT_APPINST_HEALTH)
		}
	}
	// add ref on cloudlet
	cloudlet, foundCloudlet := tbl.Cloudlets[appInst.Key.ClusterInstKey.CloudletKey]
	if !foundCloudlet {
		log.SpanLog(ctx, log.DebugLevelInfo, "AddAppInst: cloudlet not found", "key", appInst.Key)
		return
	}
	// track appinst key by cloudlet
	cloudlet.AppInstKeys[appInst.Key] = struct{}{}

	// Check if Cloudlet states have changed
	cl.CloudletState = cloudlet.State
	cl.MaintenanceState = cloudlet.MaintenanceState
	// add to alliance carriers
	for allianceCarrier, _ := range cloudlet.AllianceCarriers {
		addAppInstAlliance(ctx, app, cl, allianceCarrier)
	}
	if sendAvailableAppInst && IsAppInstUsable(cl) {
		go EEHandler.SendAvailableAppInst(ctx, app, appInst.Key, cl, carrierName)
	}

	log.SpanLog(ctx, log.DebugLevelDmedb, logMsg,
		"appName", app.AppKey.Name,
		"appVersion", app.AppKey.Version,
		"cloudletKey", appInst.Key.ClusterInstKey.CloudletKey,
		"uri", appInst.Uri,
		"latitude", appInst.CloudletLoc.Latitude,
		"longitude", appInst.CloudletLoc.Longitude,
		"healthState", appInst.HealthCheck)
}

func newDmeAppInsts() *DmeAppInsts {
	d := &DmeAppInsts{}
	d.Insts = make(map[edgeproto.VirtualClusterInstKey]*DmeAppInst)
	d.AllianceInsts = make(map[edgeproto.VirtualClusterInstKey]*DmeAppInst)
	return d
}

func addAppInstAlliance(ctx context.Context, app *DmeApp, appInst *DmeAppInst, allianceCarrier string) {
	log.SpanLog(ctx, log.DebugLevelDmedb, "addAppInstAlliance", "allianceCarrier", allianceCarrier, "app", app.AppKey, "clusterInst", appInst.virtualClusterInstKey)
	insts, found := app.Carriers[allianceCarrier]
	if !found {
		insts = newDmeAppInsts()
		app.Carriers[allianceCarrier] = insts
	}
	insts.AllianceInsts[appInst.virtualClusterInstKey] = appInst
}

func removeAppInstAlliance(ctx context.Context, app *DmeApp, appInstKey edgeproto.AppInstKey, allianceCarrier string) {
	log.SpanLog(ctx, log.DebugLevelDmedb, "removeAppInstAlliance", "allianceCarrier", allianceCarrier, "appInst", appInstKey)
	insts, found := app.Carriers[allianceCarrier]
	if !found {
		return
	}
	delete(insts.AllianceInsts, appInstKey.ClusterInstKey)
	if len(insts.Insts) == 0 && len(insts.AllianceInsts) == 0 {
		delete(app.Carriers, allianceCarrier)
	}
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
	if !ok {
		return
	}
	app.Lock()
	defer app.Unlock()
	if c, foundCarrier := app.Carriers[carrierName]; foundCarrier {
		if cl, foundAppInst := c.Insts[appInst.Key.ClusterInstKey]; foundAppInst {
			log.SpanLog(ctx, log.DebugLevelDmereq, "removing app inst", "appinst", cl, "removed appinst health", appInst.HealthCheck)
			cl.AppInstHealth = dme.HealthCheck_HEALTH_CHECK_FAIL_SERVER_FAIL

			// Remove AppInst from edgeevents plugin
			appinstState := &DmeAppInstState{
				AppInstHealth: cl.AppInstHealth,
			}
			go func(a *DmeAppInstState) {
				EEHandler.SendAppInstStateEdgeEvent(ctx, a, appInst.Key, dme.ServerEdgeEvent_EVENT_APPINST_HEALTH)
				EEHandler.RemoveAppInst(ctx, appInst.Key)
			}(appinstState)

			delete(app.Carriers[carrierName].Insts, appInst.Key.ClusterInstKey)
			log.SpanLog(ctx, log.DebugLevelDmedb, "Removing app inst",
				"appName", appkey.Name,
				"appVersion", appkey.Version,
				"latitude", cl.Location.Latitude,
				"longitude", cl.Location.Longitude)
		}
		if len(c.Insts) == 0 && len(c.AllianceInsts) == 0 {
			delete(tbl.Apps[appkey].Carriers, carrierName)
			log.SpanLog(ctx, log.DebugLevelDmedb, "Removing carrier for app",
				"carrier", carrierName,
				"appName", appkey.Name,
				"appVersion", appkey.Version)
		}
	}
	cloudlet, found := tbl.Cloudlets[appInst.Key.ClusterInstKey.CloudletKey]
	if !found {
		log.SpanLog(ctx, log.DebugLevelDmedb, "RemoveAppInst: cloudlet not found", "key", appInst.Key)
		return
	}
	// remove alliance references
	for allianceCarrier, _ := range cloudlet.AllianceCarriers {
		removeAppInstAlliance(ctx, app, appInst.Key, allianceCarrier)
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
				key.ClusterInstKey = inst.virtualClusterInstKey
				if _, foundAppInst := appInsts[key]; !foundAppInst {
					log.SpanLog(ctx, log.DebugLevelDmereq, "pruning app", "key", key)

					// Remove AppInst from edgeevents plugin
					appinstState := &DmeAppInstState{
						AppInstHealth: inst.AppInstHealth,
					}
					go func(a *DmeAppInstState) {
						EEHandler.SendAppInstStateEdgeEvent(ctx, a, key, dme.ServerEdgeEvent_EVENT_APPINST_HEALTH)
						EEHandler.RemoveAppInst(ctx, key)
					}(appinstState)

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
					appInstKey := edgeproto.AppInstKey{
						AppKey:         app.AppKey,
						ClusterInstKey: clusterInstKey,
					}

					appinstState := &DmeAppInstState{
						CloudletState: c.Insts[clusterInstKey].CloudletState,
					}
					go func(a *DmeAppInstState) {
						EEHandler.SendAppInstStateEdgeEvent(ctx, a, appInstKey, dme.ServerEdgeEvent_EVENT_CLOUDLET_STATE)
						EEHandler.RemoveCloudlet(ctx, clusterInstKey.CloudletKey)

					}(appinstState)
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
					appInstKey := edgeproto.AppInstKey{
						AppKey:         app.AppKey,
						ClusterInstKey: clusterInstKey,
					}

					appinstState := &DmeAppInstState{
						CloudletState: carr.Insts[clusterInstKey].CloudletState,
					}
					go func(a *DmeAppInstState) {
						EEHandler.SendAppInstStateEdgeEvent(ctx, a, appInstKey, dme.ServerEdgeEvent_EVENT_CLOUDLET_STATE)
						EEHandler.RemoveCloudlet(ctx, clusterInstKey.CloudletKey)
					}(appinstState)

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
		cloudlet.AllianceCarriers = make(map[string]struct{})
		cloudlet.AppInstKeys = make(map[edgeproto.AppInstKey]struct{})
		tbl.Cloudlets[in.Key] = cloudlet
	}
	sendAvailableAppInst := false
	// Check if CloudletMaintenance state has changed
	if cloudlet.MaintenanceState != in.MaintenanceState && foundCloudlet {
		cloudlet.MaintenanceState = in.MaintenanceState
		if cloudlet.MaintenanceState == dme.MaintenanceState_NORMAL_OPERATION {
			// send msg to clients that are not on this cloudlet but are closer to this cloudlet that it is available
			sendAvailableAppInst = true
		} else {
			// send msg to clients on this cloudlet that it is not available
			appinstState := &DmeAppInstState{
				MaintenanceState: cloudlet.MaintenanceState,
			}
			go EEHandler.SendCloudletMaintenanceStateEdgeEvent(ctx, appinstState, in.Key)
		}
	}
	cloudlet.GpsLocation = in.Location
	oldAllianceCarriers := cloudlet.AllianceCarriers
	newAllianceCarriers := map[string]struct{}{}
	cloudlet.AllianceCarriers = map[string]struct{}{}
	for _, org := range in.AllianceOrgs {
		cloudlet.AllianceCarriers[org] = struct{}{}
		if _, found := oldAllianceCarriers[org]; !found {
			newAllianceCarriers[org] = struct{}{}
		}
		// anything remaining in oldAllianceCarriers are those
		// that have been removed.
		delete(oldAllianceCarriers, org)
	}

	// do updates for all AppInsts on this Cloudlet
	for appInstKey, _ := range cloudlet.AppInstKeys {
		// get app
		app, _, appinst, err := getTableAppInstAndLock(appInstKey)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelDmedb, "SetInstStateFromCloudlet: appInst lookup failed", "key", appInstKey, "err", err)
			continue
		}
		appinst.MaintenanceState = in.MaintenanceState
		if sendAvailableAppInst && IsAppInstUsable(appinst) {
			go EEHandler.SendAvailableAppInst(ctx, app, appInstKey, appinst, carrier)
		}
		// add new alliance carriers
		for allianceCarrier, _ := range newAllianceCarriers {
			addAppInstAlliance(ctx, app, appinst, allianceCarrier)
		}
		// remove old alliance carriers
		for allianceCarrier, _ := range oldAllianceCarriers {
			removeAppInstAlliance(ctx, app, appInstKey, allianceCarrier)
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
	sendAvailableAppInst := false
	// Check if Cloudlet state has changed
	if cloudlet.State != info.State {
		cloudlet.State = info.State
		if cloudlet.State == dme.CloudletState_CLOUDLET_STATE_READY {
			// send msg to clients that are not on this cloudlet but are closer to this cloudlet that it is available
			sendAvailableAppInst = true
		} else {
			// send msg to clients on this cloudlet that it is not available
			appinstState := &DmeAppInstState{
				CloudletState: cloudlet.State,
			}
			go EEHandler.SendCloudletStateEdgeEvent(ctx, appinstState, info.Key)
		}
	}

	for appInstKey, _ := range cloudlet.AppInstKeys {
		app, _, appinst, err := getTableAppInstAndLock(appInstKey)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelDmedb, "SetInstStateFromCloudletInfo: appInst lookup failed", "key", appInstKey, "err", err)
			continue
		}
		appinst.CloudletState = info.State
		appinst.MaintenanceState = info.MaintenanceState
		if sendAvailableAppInst && IsAppInstUsable(appinst) {
			go EEHandler.SendAvailableAppInst(ctx, app, appInstKey, appinst, carrier)
		}
		app.Unlock()
	}
}

// must be called with table lock held. If found, app lock must be
// released by caller.
func getTableAppInstAndLock(appInstKey edgeproto.AppInstKey) (*DmeApp, *DmeAppInsts, *DmeAppInst, error) {
	app, ok := DmeAppTbl.Apps[appInstKey.AppKey]
	if !ok {
		return nil, nil, nil, fmt.Errorf("app not found")
	}
	app.Lock()
	carrier := appInstKey.ClusterInstKey.CloudletKey.Organization
	insts, ok := app.Carriers[carrier]
	if !ok {
		app.Unlock()
		return nil, nil, nil, fmt.Errorf("insts for carrier not found")
	}
	appinst, ok := insts.Insts[appInstKey.ClusterInstKey]
	if !ok {
		app.Unlock()
		return nil, nil, nil, fmt.Errorf("appinst not found")
	}
	return app, insts, appinst, nil
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
	loc           *dme.Loc
	reqCarrier    string
	results       []*foundAppInst
	appDeployment string
	appFlavor     string
	resultDist    float64
	resultLimit   int
}

type foundAppInst struct {
	distance       float64
	AppInst        *DmeAppInst
	appInstCarrier string
}

type policySearch struct {
	typ          string // just for logging
	distance     float64
	policy       *AutoProvPolicy
	cloudlet     *edgeproto.AutoProvCloudlet
	deployNowKey *edgeproto.ClusterInstKey
	freeInst     bool
	carrier      string
}

func (s *policySearch) log(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelDmereq, "search policySearch result", "potentialType", s.typ, "potentialPolicy", s.policy, "policySearch", s.cloudlet, "distance", s.distance, "freeInst", s.freeInst)
}

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

	return SearchAppInsts(ctx, carrierName, app, loc, app.Carriers, resultLimit)
}

// given the carrier, update the reply if we find a cloudlet closer
// than the max distance.  Return the distance and whether or not response was updated
func SearchAppInsts(ctx context.Context, carrierName string, app *DmeApp, loc *dme.Loc, carrierData map[string]*DmeAppInsts, resultLimit int) []*foundAppInst {
	// Eventually when we have FindCloudlet policies, we should look it
	// up here and apply it to the search config.
	search := searchAppInst{
		loc:           loc,
		reqCarrier:    carrierName,
		appDeployment: app.Deployment,
		appFlavor:     app.DefaultFlavor,
		resultLimit:   resultLimit,
	}

	for cname, carrierData := range app.Carriers {
		log.SpanLog(ctx, log.DebugLevelDmereq, "search carrier insts", "carrier", cname, "num insts", len(carrierData.Insts), "num alliance insts", len(carrierData.AllianceInsts))
		search.searchAppInsts(ctx, cname, carrierData.Insts)
		if carrierName != "" {
			// not searching all carriers, so include alliance
			search.searchAppInsts(ctx, cname, carrierData.AllianceInsts)
		}
	}
	key := &app.AppKey
	for ii, found := range search.results {
		var ipaddr net.IP
		ipaddr = found.AppInst.ip
		log.SpanLog(ctx, log.DebugLevelDmereq, "best cloudlet",
			"rank", ii,
			"app", key.Name,
			"carrier", found.appInstCarrier,
			"latitude", found.AppInst.Location.Latitude,
			"longitude", found.AppInst.Location.Longitude,
			"distance", found.distance,
			"uri", found.AppInst.Uri,
			"IP", ipaddr.String())
	}

	if len(app.AutoProvPolicies) > 0 {
		log.SpanLog(ctx, log.DebugLevelDmereq, "search AutoProvPolicy", "found", len(search.results))
		search.resultDist = InfiniteDistance
		if len(search.results) > 0 {
			search.resultDist = search.results[0].distance
		}

		policySearches := []*policySearch{}
		for _, policy := range app.AutoProvPolicies {
			if policy.DeployClientCount == 0 {
				continue
			}
			// Treat each policy as a single local redundancy group.
			// Within the group, prefer cloudlets that already have
			// a free reservable ClusterInst.
			pc := &policySearch{
				distance: search.resultDist,
				policy:   policy,
			}
			for cname, list := range policy.Cloudlets {
				search.searchPolicy(ctx, key, pc, cname, list)
			}
			pc.log(ctx)
			policySearches = append(policySearches, pc)
		}
		// Between policies, prefer the closest best cloudlet,
		// regardless of whether it targets an existing
		// ClusterInst or not.
		potential := search.getBestPotentialCloudlet(ctx, policySearches)

		log.SpanLog(ctx, log.DebugLevelDmereq, "search AutoProvPolicy result", "autoProvStats", autoProvStats, "num-potential-policies", len(policySearches))
		if potential != nil && potential.cloudlet != nil && autoProvStats != nil {
			autoProvStats.Increment(ctx, &app.AppKey, &potential.cloudlet.Key, potential.deployNowKey, potential.policy)
			log.SpanLog(ctx, log.DebugLevelDmereq,
				"potential best cloudlet",
				"app", key.Name,
				"policy", potential.policy,
				"freeInst", potential.freeInst,
				"carrier", potential.carrier,
				"latitude", potential.cloudlet.Loc.Latitude,
				"longitude", potential.cloudlet.Loc.Longitude,
				"distance", potential.distance)
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

func (s *searchAppInst) searchAppInsts(ctx context.Context, carrier string, appInsts map[edgeproto.VirtualClusterInstKey]*DmeAppInst) {
	if !s.searchCarrier(carrier) {
		return
	}
	for _, i := range appInsts {
		d := DistanceBetween(*s.loc, i.Location) + s.padDistance(carrier)
		usable := IsAppInstUsable(i)
		log.SpanLog(ctx, log.DebugLevelDmereq, "found cloudlet at",
			"carrier", carrier,
			"latitude", i.Location.Latitude,
			"longitude", i.Location.Longitude,
			"this-dist", d,
			"usable", usable)
		if !usable {
			continue
		}
		found := &foundAppInst{
			distance:       d,
			AppInst:        i,
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

func (s *searchAppInst) searchPolicy(ctx context.Context, key *edgeproto.AppKey, potential *policySearch, carrier string, list []*edgeproto.AutoProvCloudlet) {
	if !s.searchCarrier(carrier) {
		return
	}
	log.SpanLog(ctx, log.DebugLevelDmereq, "search AutoProvPolicy list", "policy", potential.policy.Name, "carrier", carrier, "list", list)

	for _, cl := range list {
		pd := DistanceBetween(*s.loc, cl.Loc) + s.padDistance(carrier)
		// This will intentionally skip auto-deployed
		// AppInsts, this should only find cloudlets
		// that could, but do not have the AppInst
		// auto-deployed.
		if pd >= s.resultDist {
			continue
		}
		freeInst := false
		// check for free reservable ClusterInst
		if DmeAppTbl.FreeReservableClusterInsts.GetForCloudlet(&cl.Key, s.appDeployment, s.appFlavor, cloudcommon.AppInstToClusterDeployment) != nil {
			freeInst = true
		}
		// controller will look for existing ClusterInst or create new one.
		cinstKey := &edgeproto.ClusterInstKey{
			CloudletKey: cl.Key,
			ClusterKey: edgeproto.ClusterKey{
				Name: cloudcommon.AutoProvClusterName,
			},
			Organization: cloudcommon.OrganizationMobiledgeX,
		}
		if potential.freeInst && !freeInst {
			// prefer free reservable ClusterInst over autocluster
			// within the same policy, regardless of distance
			continue
		}
		if potential.freeInst == freeInst {
			// compare by distance
			if pd > potential.distance {
				continue
			}
		}
		// new best
		potential.distance = pd
		potential.freeInst = freeInst
		potential.deployNowKey = cinstKey
		potential.cloudlet = cl
		potential.carrier = carrier
	}
}

func (s *searchAppInst) getBestPotentialCloudlet(ctx context.Context, list []*policySearch) *policySearch {
	// Prefer distance between cloudlet groups from different policies,
	// regardless of existing free reservable ClusterInst or not.
	var best *policySearch
	for _, pc := range list {
		if best == nil {
			best = pc
			continue
		}
		if pc.distance > best.distance {
			continue
		}
		if pc.distance == best.distance {
			// Multiple policies for the same Cloudlet.
			// Always prefer immediate provisioning policy.
			// We don't actually care which policy it is
			// for a non-immediate policy.
			if intervalCount(pc.policy) > intervalCount(best.policy) {
				continue
			}
			if pc.policy.DeployClientCount > best.policy.DeployClientCount {
				continue
			}
		}
		best = pc
	}
	return best
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

// Helper function that creates a FindCloudletReply from a DmeAppInst struct
func ConstructFindCloudletReplyFromDmeAppInst(ctx context.Context, appinst *DmeAppInst, clientloc *dme.Loc, mreply *dme.FindCloudletReply, edgeEventsCookieExpiration time.Duration) {
	mreply.Fqdn = appinst.Uri
	mreply.Status = dme.FindCloudletReply_FIND_FOUND
	mreply.CloudletLocation = &appinst.Location
	mreply.Ports = copyPorts(appinst.ports)
	// Create EdgeEventsCookieKey
	key := CreateEdgeEventsCookieKey(appinst, *clientloc)
	// Generate edgeEventsCookie
	ctx = NewEdgeEventsCookieContext(ctx, key)
	eecookie, _ := GenerateEdgeEventsCookie(key, ctx, edgeEventsCookieExpiration)
	mreply.EdgeEventsCookie = eecookie
}

func FindCloudlet(ctx context.Context, appkey *edgeproto.AppKey, carrier string, loc *dme.Loc, mreply *dme.FindCloudletReply, edgeEventsCookieExpiration time.Duration) error {
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
		ConstructFindCloudletReplyFromDmeAppInst(ctx, best.AppInst, loc, mreply, edgeEventsCookieExpiration)
		// Update Context variable if passed
		cloudlet := best.AppInst.clusterInstKey.CloudletKey.Name
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

func GetAppInstList(ctx context.Context, ckey *CookieKey, mreq *dme.AppInstListRequest, clist *dme.AppInstListReply, edgeEventsCookieExpiration time.Duration) {
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
		cloc, exists := foundCloudlets[found.AppInst.clusterInstKey.CloudletKey]
		if !exists {
			cloc = new(dme.CloudletLocation)

			cloc.GpsLocation = &found.AppInst.Location
			cloc.CarrierName = found.AppInst.clusterInstKey.CloudletKey.Organization
			cloc.CloudletName = found.AppInst.clusterInstKey.CloudletKey.Name
			cloc.Distance = found.distance
			foundCloudlets[found.AppInst.clusterInstKey.CloudletKey] = cloc
			clist.Cloudlets = append(clist.Cloudlets, cloc)
		}
		ai := dme.Appinstance{}
		ai.AppName = appkey.Name
		ai.AppVers = appkey.Version
		ai.OrgName = appkey.Organization
		ai.Fqdn = found.AppInst.Uri
		ai.Ports = copyPorts(found.AppInst.ports)

		// Create EdgeEventsCookieKey
		key := CreateEdgeEventsCookieKey(found.AppInst, *mreq.GpsLocation)
		// Generate edgeEventsCookie
		ctx = NewEdgeEventsCookieContext(ctx, key)
		eecookie, _ := GenerateEdgeEventsCookie(key, ctx, edgeEventsCookieExpiration)
		ai.EdgeEventsCookie = eecookie

		cloc.Appinstances = append(cloc.Appinstances, &ai)
	}
	clist.Status = dme.AppInstListReply_AI_SUCCESS
}

func StreamEdgeEvent(ctx context.Context, svr dme.MatchEngineApi_StreamEdgeEventServer, edgeEventsCookieExpiration time.Duration) (reterr error) {
	// Initialize vars used in persistent connection
	var appInstKey *edgeproto.AppInstKey
	var sessionCookie string
	var sessionCookieKey *CookieKey
	var edgeEventsCookieKey *EdgeEventsCookieKey
	// Initialize vars used for edgeevents stats
	var deviceInfoStatic *dme.DeviceInfoStatic
	var lastLocation *dme.Loc
	var lastCarrier string = ""
	var lastDeviceInfoDynamic *dme.DeviceInfoDynamic
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
	// Send first deviceinfo stats update
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
		lastLocation = &edgeEventsCookieKey.Location
		// Initialize deviceInfoStatic for stats
		if initMsg.DeviceInfoStatic != nil {
			deviceInfoStatic = initMsg.DeviceInfoStatic
		} else {
			deviceInfoStatic = &dme.DeviceInfoStatic{}
		}
		// Intialize last deviceInfoDynamic for stats
		lastDeviceInfoDynamic = initMsg.DeviceInfoDynamic
		if lastDeviceInfoDynamic != nil {
			lastCarrier = lastDeviceInfoDynamic.CarrierName
		}
		// Create AppInstKey from SessionCookie and EdgeEventsCookie
		appInstKey = &edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				Organization: sessionCookieKey.OrgName,
				Name:         sessionCookieKey.AppName,
				Version:      sessionCookieKey.AppVers,
			},
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
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
		// Send first deviceinfo stats for client
		deviceInfo := &DeviceInfo{
			DeviceInfoStatic:  deviceInfoStatic,
			DeviceInfoDynamic: lastDeviceInfoDynamic,
		}
		updateDeviceInfoStats(ctx, appInstKey, deviceInfo, lastLocation, "event init connection")
		// Add Client to edgeevents plugin
		EEHandler.AddClient(ctx, *appInstKey, *sessionCookieKey, *lastLocation, lastCarrier, sendFunc)
		// Remove Client from edgeevents plugin when StreamEdgeEvent exits
		defer EEHandler.RemoveClient(ctx, *appInstKey, *sessionCookieKey)
		// Send successful init response
		initServerEdgeEvent := new(dme.ServerEdgeEvent)
		initServerEdgeEvent.EventType = dme.ServerEdgeEvent_EVENT_INIT_CONNECTION
		EEHandler.SendEdgeEventToClient(ctx, initServerEdgeEvent, *appInstKey, *sessionCookieKey)
	} else {
		return fmt.Errorf("First message should have event type EVENT_INIT_CONNECTION")
	}

	// Initialize rate limiter so that we can handle all the incoming messages
	rateLimiter := ratelimit.NewTokenBucketLimiter(ratelimit.DefaultReqsPerSecondPerApi, int(ratelimit.DefaultTokenBucketSize))
	// Loop while persistent connection is up
loop:
	for {
		// Receive data from stream
		cupdate, err := svr.Recv()
		ctx = svr.Context()
		// Rate limit
		err = rateLimiter.Limit(ctx, nil)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelDmereq, "Limiting client messages", "err", err)
			sendErrorEventToClient(ctx, fmt.Sprintf("Limiting client messages. Most recent ClientEdgeEvent will not be processed: %v. Error is: %s", cupdate, err), *appInstKey, *sessionCookieKey)
			continue
		}
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
			err := ValidateLocation(cupdate.GpsLocation)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelDmereq, "Invalid EVENT_LATENCY_SAMPLES, invalid location", "err", err)
				sendErrorEventToClient(ctx, fmt.Sprintf("Invalid EVENT_LATENCY_SAMPLES, invalid location: %s", err), *appInstKey, *sessionCookieKey)
				continue
			}
			// Process latency samples and send results to client
			_, err = EEHandler.ProcessLatencySamples(ctx, *appInstKey, *sessionCookieKey, cupdate.Samples)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelDmereq, "ClientEdgeEvent latency unable to process latency samples", "err", err)
				sendErrorEventToClient(ctx, fmt.Sprintf("ClientEdgeEvent latency unable to process latency samples, error is: %s", err), *appInstKey, *sessionCookieKey)
				continue
			}
			// Latency stats update
			deviceInfoDynamic := cupdate.DeviceInfoDynamic
			deviceInfo := &DeviceInfo{
				DeviceInfoStatic:  deviceInfoStatic,
				DeviceInfoDynamic: deviceInfoDynamic,
			}
			latencyStatKey := GetLatencyStatKey(*appInstKey, deviceInfo, cupdate.GpsLocation, int(Settings.LocationTileSideLengthKm))
			latencyStatInfo := &LatencyStatInfo{
				Samples: cupdate.Samples,
			}
			edgeEventStatCall := &EdgeEventStatCall{
				Metric:          cloudcommon.LatencyMetric,
				LatencyStatKey:  latencyStatKey,
				LatencyStatInfo: latencyStatInfo,
			}
			EEStats.RecordEdgeEventStatCall(edgeEventStatCall)
		case dme.ClientEdgeEvent_EVENT_LOCATION_UPDATE:
			// Client updated gps location
			err := ValidateLocation(cupdate.GpsLocation)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelDmereq, "Invalid EVENT_LOCATION_UPDATE, invalid location", "err", err)
				sendErrorEventToClient(ctx, fmt.Sprintf("Invalid EVENT_LOCATION_UPDATE, invalid location: %s", err), *appInstKey, *sessionCookieKey)
				continue
			}
			// Deviceinfo stats update
			deviceInfoDynamic := cupdate.DeviceInfoDynamic
			deviceInfo := &DeviceInfo{
				DeviceInfoStatic:  deviceInfoStatic,
				DeviceInfoDynamic: deviceInfoDynamic,
			}
			// Update deviceinfo stats if DeviceInfoDynamic has changed or Location has moved to different tile
			if !isDeviceInfoDynamicEqual(deviceInfoDynamic, lastDeviceInfoDynamic) || !isLocationInSameTile(cupdate.GpsLocation, lastLocation) {
				updateDeviceInfoStats(ctx, appInstKey, deviceInfo, cupdate.GpsLocation, "event location update")
				lastDeviceInfoDynamic = deviceInfoDynamic
				if lastDeviceInfoDynamic != nil {
					if lastCarrier != lastDeviceInfoDynamic.CarrierName {
						EEHandler.UpdateClientCarrier(ctx, *appInstKey, *sessionCookieKey, lastDeviceInfoDynamic.CarrierName)
						lastCarrier = lastDeviceInfoDynamic.CarrierName
					}
				}
			}
			// Update last client location in plugin if different from lastLocation
			if cupdate.GpsLocation.Latitude != lastLocation.Latitude || cupdate.GpsLocation.Longitude != lastLocation.Longitude {
				EEHandler.UpdateClientLastLocation(ctx, *appInstKey, *sessionCookieKey, *cupdate.GpsLocation)
				lastLocation = cupdate.GpsLocation
			}
			// Check if there is a better cloudlet based on location update
			fcreply := new(dme.FindCloudletReply)
			err = FindCloudlet(ctx, &appInstKey.AppKey, lastCarrier, cupdate.GpsLocation, fcreply, edgeEventsCookieExpiration)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelDmereq, "Error trying to find closer cloudlet", "err", err)
				continue
			}
			if fcreply.Status != dme.FindCloudletReply_FIND_FOUND {
				log.SpanLog(ctx, log.DebugLevelDmereq, "Unable to find any cloudlets", "FindStatus", fcreply.Status)
				continue
			}
			newEECookieKey, err := VerifyEdgeEventsCookie(ctx, fcreply.EdgeEventsCookie)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelDmereq, "Error trying verify new cloudlet edgeeventscookie", "err", err)
				continue
			}
			// Check if new appinst is different from current
			if !IsTheSameCluster(newEECookieKey, edgeEventsCookieKey) {
				// Send new cloudlet update
				newCloudletEdgeEvent := new(dme.ServerEdgeEvent)
				newCloudletEdgeEvent.EventType = dme.ServerEdgeEvent_EVENT_CLOUDLET_UPDATE
				newCloudletEdgeEvent.NewCloudlet = fcreply
				EEHandler.SendEdgeEventToClient(ctx, newCloudletEdgeEvent, *appInstKey, *sessionCookieKey)
			}
		case dme.ClientEdgeEvent_EVENT_CUSTOM_EVENT:
			customStatKey := GetCustomStatKey(*appInstKey, cupdate.CustomEvent)
			customStatInfo := &CustomStatInfo{
				Samples: cupdate.Samples,
			}
			edgeEventStatCall := &EdgeEventStatCall{
				Metric:         cloudcommon.CustomMetric,
				CustomStatKey:  customStatKey,
				CustomStatInfo: customStatInfo,
			}
			EEStats.RecordEdgeEventStatCall(edgeEventStatCall)
		default:
			// Unknown client event
			log.SpanLog(ctx, log.DebugLevelDmereq, "Received unknown event type", "eventtype", cupdate.EventType)
			sendErrorEventToClient(ctx, fmt.Sprintf("Received unknown event type: %s", cupdate.EventType), *appInstKey, *sessionCookieKey)
		}
	}
	return reterr
}

// helper function that creates and sends an EVENT_ERROR ServerEdgeEvent to corresponding client
func sendErrorEventToClient(ctx context.Context, msg string, appInstKey edgeproto.AppInstKey, cookieKey CookieKey) {
	errorEdgeEvent := new(dme.ServerEdgeEvent)
	errorEdgeEvent.EventType = dme.ServerEdgeEvent_EVENT_ERROR
	errorEdgeEvent.ErrorMsg = msg
	EEHandler.SendEdgeEventToClient(ctx, errorEdgeEvent, appInstKey, cookieKey)
}

// helper function that updates deviceinfo stats
func updateDeviceInfoStats(ctx context.Context, appInstKey *edgeproto.AppInstKey, deviceInfo *DeviceInfo, loc *dme.Loc, callerMethod string) {
	deviceStatKey := GetDeviceStatKey(*appInstKey, deviceInfo, loc, int(Settings.LocationTileSideLengthKm))
	edgeEventStatCall := &EdgeEventStatCall{
		Metric:        cloudcommon.DeviceMetric,
		DeviceStatKey: deviceStatKey,
	}
	log.SpanLog(ctx, log.DebugLevelDmereq, "Updating deviceinfo stats", "appinst", appInstKey, "callermethod", callerMethod)
	EEStats.RecordEdgeEventStatCall(edgeEventStatCall)
}

func isDeviceInfoDynamicEqual(d1 *dme.DeviceInfoDynamic, d2 *dme.DeviceInfoDynamic) bool {
	if d1 == nil && d2 == nil {
		return true
	}
	if d1 == nil {
		return false
	}
	if d2 == nil {
		return false
	}
	if d1.DataNetworkType != d2.DataNetworkType || d1.SignalStrength != d2.SignalStrength {
		return false
	}
	return true
}

func isLocationInSameTile(l1 *dme.Loc, l2 *dme.Loc) bool {
	if l1 == nil && l2 == nil {
		return true
	}
	if l1 == nil {
		return false
	}
	if l2 == nil {
		return false
	}
	tileLength := int(Settings.LocationTileSideLengthKm)
	return GetLocationTileFromGpsLocation(l1, tileLength) == GetLocationTileFromGpsLocation(l2, tileLength)
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
					"Latitude", inst.Location.Latitude,
					"Longitude", inst.Location.Longitude)
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

func ValidateLocation(loc *dme.Loc) error {
	if loc == nil {
		return grpc.Errorf(codes.InvalidArgument, "Missing GpsLocation")
	}
	if !util.IsLatitudeValid(loc.Latitude) || !util.IsLongitudeValid(loc.Longitude) {
		return grpc.Errorf(codes.InvalidArgument, "Invalid GpsLocation")
	}
	return nil
}

func SettingsUpdated(ctx context.Context, old *edgeproto.Settings, new *edgeproto.Settings) {
	Settings = *new
	autoProvStats.UpdateSettings(new.AutoDeployIntervalSec)
	Stats.UpdateSettings(time.Duration(new.DmeApiMetricsCollectionInterval))
	clientsMap.UpdateClientTimeout(new.AppinstClientCleanupInterval)
	EEStats.UpdateSettings(time.Duration(new.EdgeEventsMetricsCollectionInterval))
	RateLimitMgr.UpdateDisableRateLimit(new.DisableRateLimit)
	RateLimitMgr.UpdateMaxTrackedIps(int(new.RateLimitMaxTrackedIps))
}
