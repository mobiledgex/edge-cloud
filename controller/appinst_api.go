package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AppInstApi struct {
	all     *AllApis
	sync    *Sync
	store   edgeproto.AppInstStore
	cache   edgeproto.AppInstCache
	idStore edgeproto.AppInstIdStore
}

const RootLBSharedPortBegin int32 = 10000

var RequireAppInstPortConsistency = false

// Transition states indicate states in which the CRM is still busy.
var CreateAppInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_CREATING: struct{}{},
}
var UpdateAppInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_UPDATING: struct{}{},
}
var DeleteAppInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_DELETING: struct{}{},
}

func NewAppInstApi(sync *Sync, all *AllApis) *AppInstApi {
	appInstApi := AppInstApi{}
	appInstApi.all = all
	appInstApi.sync = sync
	appInstApi.store = edgeproto.NewAppInstStore(sync.store)
	edgeproto.InitAppInstCache(&appInstApi.cache)
	sync.RegisterCache(&appInstApi.cache)
	return &appInstApi
}

func (s *AppInstApi) Get(key *edgeproto.AppInstKey, val *edgeproto.AppInst) bool {
	return s.cache.Get(key, val)
}

func (s *AppInstApi) HasKey(key *edgeproto.AppInstKey) bool {
	return s.cache.HasKey(key)
}

func isAutoDeleteAppInstOk(callerOrg string, appInst *edgeproto.AppInst, app *edgeproto.App) bool {
	if appInst.Liveness == edgeproto.Liveness_LIVENESS_DYNAMIC || app.DelOpt == edgeproto.DeleteType_AUTO_DELETE {
		return true
	}
	if callerOrg == app.Key.Organization && appInst.Liveness == edgeproto.Liveness_LIVENESS_AUTOPROV {
		// Caller owns the App and AppInst. Allow them to automatically
		// delete auto-provisioned instances. Otherwise, this is
		// probably an operator trying to delete a cloudlet or common
		// ClusterInst, and should not be able to automatically delete
		// developer's instances.
		return true
	}
	return false
}

func (s *AppInstApi) deleteCloudletOk(stm concurrency.STM, refs *edgeproto.CloudletRefs, defaultClustKey *edgeproto.ClusterInstKey, dynInsts map[edgeproto.AppInstKey]struct{}) error {
	aiKeys := []*edgeproto.AppInstKey{}
	// Only need to check VM apps, as other AppInsts require ClusterInsts,
	// so ClusterInst check will apply.
	for _, aiRefKey := range refs.VmAppInsts {
		aiKey := edgeproto.AppInstKey{}
		aiKey.FromAppInstRefKey(&aiRefKey, &refs.Key)
		aiKeys = append(aiKeys, &aiKey)
	}
	// check any AppInsts on default cluster
	clustRefs := edgeproto.ClusterRefs{}
	if defaultClustKey != nil && s.all.clusterRefsApi.store.STMGet(stm, defaultClustKey, &clustRefs) {
		for _, aiRefKey := range clustRefs.Apps {
			aiKey := edgeproto.AppInstKey{}
			aiKey.FromClusterRefsAppInstKey(&aiRefKey, defaultClustKey)
			aiKeys = append(aiKeys, &aiKey)
		}
	}
	return s.cascadeDeleteOk(stm, refs.Key.Organization, "Cloudlet", aiKeys, dynInsts)
}

func (s *AppInstApi) cascadeDeleteOk(stm concurrency.STM, callerOrg, deleteTarget string, aiKeys []*edgeproto.AppInstKey, dynInsts map[edgeproto.AppInstKey]struct{}) error {
	for _, aiKey := range aiKeys {
		ai := edgeproto.AppInst{}
		if !s.store.STMGet(stm, aiKey, &ai) {
			continue
		}
		app := edgeproto.App{}
		if !s.all.appApi.store.STMGet(stm, &ai.Key.AppKey, &app) {
			continue
		}
		if isAutoDeleteAppInstOk(callerOrg, &ai, &app) {
			dynInsts[ai.Key] = struct{}{}
			continue
		}
		return fmt.Errorf("%s in use by AppInst %s", deleteTarget, ai.Key.GetKeyString())
	}
	return nil
}

func (s *AppInstApi) CheckCloudletAppinstsCompatibleWithTrustPolicy(ctx context.Context, ckey *edgeproto.CloudletKey, TrustPolicy *edgeproto.TrustPolicy) error {
	apps := make(map[edgeproto.AppKey]*edgeproto.App)
	s.all.appApi.GetAllApps(apps)
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, data := range s.cache.Objs {
		if !key.ClusterInstKey.CloudletKey.Matches(ckey) {
			continue
		}
		val := data.Obj
		app, found := apps[val.Key.AppKey]
		if !found {
			return fmt.Errorf("App not found: %s", val.Key.AppKey.String())
		}
		err := s.all.appApi.CheckAppCompatibleWithTrustPolicy(ctx, ckey, app, TrustPolicy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *AppInstApi) updateAppInstRevision(ctx context.Context, key *edgeproto.AppInstKey, revision string) error {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, key, &inst) {
			// got deleted in the meantime
			return nil
		}
		inst.Revision = revision
		log.SpanLog(ctx, log.DebugLevelApi, "AppInst revision updated", "key", key, "revision", revision)

		s.store.STMPut(stm, &inst)
		return nil
	})

	return err
}

func (s *AppInstApi) UsesClusterInst(callerOrg string, in *edgeproto.ClusterInstKey) bool {
	var app edgeproto.App
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		val := data.Obj
		if val.ClusterInstKey().Matches(in) && s.all.appApi.Get(&val.Key.AppKey, &app) {
			if !isAutoDeleteAppInstOk(callerOrg, val, &app) {
				return true
			}
		}
	}
	return false
}

func (s *AppInstApi) AutoDeleteAppInsts(ctx context.Context, dynInsts map[edgeproto.AppInstKey]struct{}, crmoverride edgeproto.CRMOverride, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	var err error
	log.SpanLog(ctx, log.DebugLevelApi, "Auto-deleting AppInsts")

	keys := []edgeproto.AppInstKey{}
	for key := range dynInsts {
		keys = append(keys, key)
	}
	// sort keys for stable iteration order, needed for testing
	sort.Slice(keys[:], func(i, j int) bool {
		return keys[i].GetKeyString() < keys[j].GetKeyString()
	})

	//Spin in case cluster was just created and apps are still in the creation process and cannot be deleted
	var spinTime time.Duration
	start := time.Now()
	for _, key := range keys {
		val := &edgeproto.AppInst{}
		if !s.cache.Get(&key, val) {
			continue
		}
		log.SpanLog(ctx, log.DebugLevelApi, "Auto-deleting AppInst ", "appinst", val.Key.AppKey.Name)
		cb.Send(&edgeproto.Result{Message: "Autodeleting AppInst " + val.Key.AppKey.Name})
		for {
			// ignore CRM errors when deleting dynamic apps as we will be deleting the cluster anyway
			cctx := DefCallContext()
			if crmoverride != edgeproto.CRMOverride_NO_OVERRIDE {
				cctx.SetOverride(&crmoverride)
			} else {
				crmo := edgeproto.CRMOverride_IGNORE_CRM_ERRORS
				cctx.SetOverride(&crmo)
			}
			// cloudlet ready check should already have been done
			cctx.SkipCloudletReadyCheck = true
			err = s.deleteAppInstInternal(cctx, val, cb)
			if err != nil && err.Error() == val.Key.NotFoundError().Error() {
				err = nil
				break
			}
			if err != nil && strings.Contains(err.Error(), ObjBusyDeletionMsg) {
				spinTime = time.Since(start)
				if spinTime > s.all.settingsApi.Get().DeleteAppInstTimeout.TimeDuration() {
					log.SpanLog(ctx, log.DebugLevelApi, "Timeout while waiting for App", "appName", val.Key.AppKey.Name)
					return err
				}
				log.SpanLog(ctx, log.DebugLevelApi, "AppInst busy, retrying in 0.5s...", "appName", val.Key.AppKey.Name)
				time.Sleep(500 * time.Millisecond)
			} else { //if its anything other than an appinst busy error, break out of the spin
				break
			}
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (s *AppInstApi) AutoDelete(ctx context.Context, appinsts []*edgeproto.AppInst) error {
	if len(appinsts) == 0 {
		return nil
	}
	// sort so order is deterministic for testing
	sort.Slice(appinsts, func(i, j int) bool {
		return appinsts[i].Key.GetKeyString() < appinsts[j].Key.GetKeyString()
	})

	failed := 0
	deleted := 0
	for _, val := range appinsts {
		log.SpanLog(ctx, log.DebugLevelApi, "Auto-delete AppInst for App", "AppInst", val.Key)
		stream := streamoutAppInst{}
		stream.ctx = ctx
		stream.debugLvl = log.DebugLevelApi
		err := s.DeleteAppInst(val, &stream)
		if err != nil && err.Error() != val.Key.NotFoundError().Error() {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to auto-delete AppInst", "AppInst", val.Key, "err", err)
			failed++
		} else {
			deleted++
		}
	}
	if failed > 0 {
		return fmt.Errorf("Auto-deleted %d AppInsts but failed to delete %d AppInsts for App", deleted, failed)
	}
	return nil
}

func (s *AppInstApi) UsesFlavor(key *edgeproto.FlavorKey) *edgeproto.AppInstKey {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for k, data := range s.cache.Objs {
		app := data.Obj
		if app.Flavor.Matches(key) {
			return &k
		}
	}
	return nil
}

func (s *AppInstApi) CreateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_CreateAppInstServer) error {
	return s.createAppInstInternal(DefCallContext(), in, cb)
}

func getProtocolBitMap(proto dme.LProto) (int32, error) {
	var bitmap int32
	switch proto {
	case dme.LProto_L_PROTO_TCP:
		bitmap = 1 //01
		break
	//put all "UDP" protocols below here
	case dme.LProto_L_PROTO_UDP:
		bitmap = 2 //10
		break
	default:
		return 0, errors.New("Unknown protocol in use for this app")
	}
	return bitmap, nil
}

func protocolInUse(protocolsToCheck int32, usedProtocols int32) bool {
	return (protocolsToCheck & usedProtocols) != 0
}

func addProtocol(protos int32, protocolToAdd int32) int32 {
	return protos | protocolToAdd
}

func removeProtocol(protos int32, protocolToRemove int32) int32 {
	return protos & (^protocolToRemove)
}

func (s *AppInstApi) startAppInstStream(ctx context.Context, key *edgeproto.AppInstKey, inCb edgeproto.AppInstApi_CreateAppInstServer) (*streamSend, edgeproto.AppInstApi_CreateAppInstServer, error) {
	streamSendObj, err := s.all.streamObjApi.startStream(ctx, key, inCb)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to start appinst stream", "err", err)
		return nil, inCb, err
	}
	return streamSendObj, &CbWrapper{
		streamSendObj: streamSendObj,
		GenericCb:     inCb,
	}, nil
}

func (s *AppInstApi) stopAppInstStream(ctx context.Context, key *edgeproto.AppInstKey, streamSendObj *streamSend, objErr error) {
	if err := s.all.streamObjApi.stopStream(ctx, key, streamSendObj, objErr); err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to stop appinst stream", "err", err)
	}
}

func (s *StreamObjApi) StreamAppInst(key *edgeproto.AppInstKey, cb edgeproto.StreamObjApi_StreamAppInstServer) error {
	// populate the clusterinst developer from the app developer if not already present
	cloudcommon.SetAppInstKeyDefaults(key)
	return s.StreamMsgs(key, cb)
}

func (s *AppInstApi) checkForAppinstCollisions(ctx context.Context, key *edgeproto.AppInstKey) error {
	// To avoid name collisions in the CRM after sanitizing the app name, validate that there is not
	// another app running which will have the same name after sanitizing.   DNSSanitize is used here because
	// it is the most stringent special character replacement
	keyString := key.String()
	sanitizedKey := util.DNSSanitize(keyString)
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()

	for _, data := range s.cache.Objs {
		val := data.Obj
		existingKeyString := val.Key.String()
		existingSanitizedKey := util.DNSSanitize(existingKeyString)
		if sanitizedKey == existingSanitizedKey && keyString != existingKeyString {
			log.SpanLog(ctx, log.DebugLevelApi, "AppInst collision", "keyString", keyString, "existingKeyString", existingKeyString, "sanitizedKey", sanitizedKey)
			return fmt.Errorf("Cannot deploy AppInst due to DNS name collision with existing instance %s - %s", existingKeyString, sanitizedKey)
		}
	}
	return nil
}

type AutoClusterType int

const (
	NoAutoCluster AutoClusterType = iota
	ChooseAutoCluster
	ReservableAutoCluster
	MultiTenantAutoCluster
	// For backwards compatibility with old CRMs that don't understand
	// virtualized AppInst keys (where the real cluster name is
	// AppInst.RealClusterName), we create the old-style dedicated
	// auto-clusters instead of the new reservable auto-clusters
	DeprecatedAutoCluster
)

func (s *AppInstApi) checkPortOverlapDedicatedLB(appPorts []dme.AppPort, clusterInstKey *edgeproto.VirtualClusterInstKey) error {
	lookupKey := edgeproto.AppInst{Key: edgeproto.AppInstKey{ClusterInstKey: *clusterInstKey}}
	err := s.cache.Show(&lookupKey, func(obj *edgeproto.AppInst) error {
		if obj.State == edgeproto.TrackedState_DELETE_ERROR || edgeproto.IsTransientState(obj.State) {
			// ignore apps that are in failed, or transient state
			return nil
		}
		for ii := range appPorts {
			for jj := range obj.MappedPorts {
				if edgeproto.DoPortsOverlap(appPorts[ii], obj.MappedPorts[jj]) {
					if appPorts[ii].EndPort != appPorts[ii].InternalPort && appPorts[ii].EndPort != 0 {
						return fmt.Errorf("port range %d-%d overlaps with ports in use on the cluster", appPorts[ii].InternalPort, appPorts[ii].EndPort)
					}
					return fmt.Errorf("port %d is already in use on the cluster", appPorts[ii].InternalPort)
				}
			}
		}
		return nil
	})
	return err
}

// createAppInstInternal is used to create dynamic app insts internally,
// bypassing static assignment.
func (s *AppInstApi) createAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, inCb edgeproto.AppInstApi_CreateAppInstServer) (reterr error) {
	var clusterInst edgeproto.ClusterInst
	ctx := inCb.Context()

	// To accommodate automatically created reservable ClusterInsts that can
	// have different names than the user's desired Cluster name, the
	// AppInst key's cluster name is a virtual or vanity name, and the
	// actual underlying real ClusterInst's name is in.RealClusterName.
	//
	// For backwards compatibility, if the RealClusterName is blank, then
	// RealClusterName is assumed to be the same as the Virtual cluster name.
	//
	// There are 3 possible configurations:
	// ClusterKey Name is not an autocluster, RealClusterName blank ->
	// ...may target Manual or Reservable, but not multi-tenant
	// ClusterKey Name is autocluster, RealClusterName specified ->
	// ...Real may target Manual, Reservable, or Multi-tenant
	// ClusterKey Name is autocluster, RealClusterName blank ->
	// ...controller will choose Reservable or Multi-tenant for Real.
	//
	// If left blank for an autocluster, then the system will choose
	// (or create) the target ClusterInst, and fill in the RealClusterName
	// with that chosen ClusterInst.
	autoClusterSpecified := false
	if strings.HasPrefix(in.Key.ClusterInstKey.ClusterKey.Name, cloudcommon.AutoClusterPrefix) {
		autoClusterSpecified = true
	}
	freeClusterInsts := []edgeproto.ClusterInstKey{}
	if in.RealClusterName == "" && autoClusterSpecified {
		// gather free reservable ClusterInsts for the target Cloudlet
		s.all.clusterInstApi.cache.Mux.Lock()
		for key, data := range s.all.clusterInstApi.cache.Objs {
			if !in.Key.ClusterInstKey.CloudletKey.Matches(&data.Obj.Key.CloudletKey) {
				// not the target Cloudlet
				continue
			}
			if data.Obj.Reservable && data.Obj.ReservedBy == "" {
				// free reservable ClusterInst - we will double-check in STM
				freeClusterInsts = append(freeClusterInsts, key)
			}
		}
		s.all.clusterInstApi.cache.Mux.Unlock()
	}
	if !autoClusterSpecified && in.RealClusterName != "" {
		if in.Key.ClusterInstKey.ClusterKey.Name == in.RealClusterName {
			in.RealClusterName = ""
		} else {
			return fmt.Errorf("Cannot specify real cluster name without %s cluster prefix, or must match cluster key name", cloudcommon.AutoClusterPrefix)
		}
	}

	// populate the clusterinst developer from the app developer if not already present
	setClusterOrg, setClusterName := cloudcommon.SetAppInstKeyDefaults(&in.Key)
	appInstKey := in.Key
	// create stream once AppInstKey is formed correctly
	sendObj, cb, err := s.startAppInstStream(ctx, &appInstKey, inCb)
	if err == nil {
		defer func() {
			s.stopAppInstStream(ctx, &appInstKey, sendObj, reterr)
		}()
	}

	if setClusterOrg {
		cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
	}
	if setClusterName {
		cb.Send(&edgeproto.Result{Message: "Setting ClusterInst name to default"})
	}
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}

	if in.Liveness == edgeproto.Liveness_LIVENESS_UNKNOWN {
		in.Liveness = edgeproto.Liveness_LIVENESS_STATIC
	}
	cctx.SetOverride(&in.CrmOverride)

	autocluster := false
	autoClusterType := NoAutoCluster
	sidecarApp := false
	err = s.checkForAppinstCollisions(ctx, &in.Key)
	if err != nil {
		return err
	}
	appDeploymentType := ""
	reservedAutoClusterId := -1
	var reservedClusterInstKey *edgeproto.ClusterInstKey
	realClusterName := in.RealClusterName
	var cloudletFeatures *platform.Features
	cloudletCompatibilityVersion := uint32(0)
	var cloudletPlatformType edgeproto.PlatformType
	var cloudletLoc dme.Loc

	defer func() {
		if reterr != nil {
			return
		}
		s.RecordAppInstEvent(ctx, &in.Key, cloudcommon.CREATED, cloudcommon.InstanceUp)
		if reservedClusterInstKey != nil {
			s.all.clusterInstApi.RecordClusterInstEvent(ctx, reservedClusterInstKey, cloudcommon.RESERVED, cloudcommon.InstanceUp)
		}
	}()

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		// reset modified state in case STM hits conflict and runs again
		autocluster = false
		autoClusterType = NoAutoCluster
		sidecarApp = false
		reservedAutoClusterId = -1
		reservedClusterInstKey = nil
		in.RealClusterName = realClusterName
		cloudletCompatibilityVersion = 0

		// lookup App so we can get flavor for reservable ClusterInst
		var app edgeproto.App
		if !s.all.appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return in.Key.AppKey.NotFoundError()
		}
		if app.DeletePrepare {
			return in.Key.AppKey.BeingDeletedError()
		}
		if cloudcommon.IsClusterInstReqd(&app) && in.Key.ClusterInstKey.ClusterKey.Name == cloudcommon.DefaultClust {
			return fmt.Errorf("Cannot use blank or default Cluster name when ClusterInst is required")
		}
		if in.Flavor.Name == "" {
			in.Flavor = app.DefaultFlavor
			if in.Flavor.Name == "" {
				return fmt.Errorf("No AppInst or App flavor specified")
			}
		}
		sidecarApp = cloudcommon.IsSideCarApp(&app)
		// make sure cloudlet exists so we don't create refs for missing cloudlet
		cloudlet := edgeproto.Cloudlet{}
		if !s.all.cloudletApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudlet) {
			return errors.New("Specified Cloudlet not found")
		}
		if cloudlet.DeletePrepare {
			return cloudlet.Key.BeingDeletedError()
		}
		cloudletPlatformType = cloudlet.PlatformType
		cloudletLoc = cloudlet.Location
		info := edgeproto.CloudletInfo{}
		if !s.all.cloudletInfoApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &info) {
			return fmt.Errorf("No resource information found for Cloudlet %s", in.Key.ClusterInstKey.CloudletKey)
		}
		cloudletCompatibilityVersion = info.CompatibilityVersion
		cloudletFeatures, err = GetCloudletFeatures(ctx, cloudlet.PlatformType)
		if err != nil {
			return fmt.Errorf("Failed to get features for platform: %s", err)
		}
		if in.DedicatedIp && !cloudletFeatures.SupportsAppInstDedicatedIP {
			return fmt.Errorf("Target cloudlet platform does not support a per-AppInst dedicated IP")
		}
		if s.store.STMGet(stm, &in.Key, in) {
			if !cctx.Undo && in.State != edgeproto.TrackedState_DELETE_ERROR && !ignoreTransient(cctx, in.State) {
				if in.State == edgeproto.TrackedState_CREATE_ERROR {
					cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
					cb.Send(&edgeproto.Result{Message: "Use DeleteAppInst to remove and try again"})
				}
				return in.Key.ExistsError()
			}
			in.Errors = nil
			// must reset Uri
			in.Uri = ""
		} else {
			err := in.Validate(edgeproto.AppInstAllFieldsMap)
			if err != nil {
				return err
			}
		}

		if err := s.all.cloudletInfoApi.checkCloudletReady(cctx, stm, &in.Key.ClusterInstKey.CloudletKey, cloudcommon.Create); err != nil {
			return err
		}

		if cloudletFeatures.IsSingleKubernetesCluster {
			// disable autocluster logic, since there's only one cluster
			autoClusterSpecified = false
			// Virtual cluster name is ignored (can be anything)
			// RealClusterName is default cluster if set.
			if in.RealClusterName != "" {
				// doesn't need to be specified, but if it is,
				// it better be the one and only cluster name.
				if in.RealClusterName != cloudcommon.DefaultClust {
					return fmt.Errorf("Invalid RealClusterName for single kubernetes cluster cloudlet, should be left blank")
				}
			}
			if in.Key.ClusterInstKey.ClusterKey.Name == cloudcommon.DefaultClust {
				in.RealClusterName = ""
			} else {
				in.RealClusterName = cloudcommon.DefaultClust
			}
			if cloudlet.SingleKubernetesClusterOwner != "" {
				// ST cluster
				if in.Key.ClusterInstKey.Organization != cloudlet.SingleKubernetesClusterOwner {
					return fmt.Errorf("ClusterInst organization must be set to %s", cloudlet.SingleKubernetesClusterOwner)
				}
				autoClusterType = NoAutoCluster
			} else {
				// MT cluster
				if !app.AllowServerless {
					return fmt.Errorf("Target cloudlet platform only supports serverless Apps")
				}
				if in.Key.ClusterInstKey.Organization != cloudcommon.OrganizationMobiledgeX {
					return fmt.Errorf("ClusterInst organization must be set to %s", cloudcommon.OrganizationMobiledgeX)
				}
				key := in.ClusterInstKey()
				clusterInst := edgeproto.ClusterInst{}
				if !s.all.clusterInstApi.store.STMGet(stm, key, &clusterInst) {
					return key.NotFoundError()
				}
				if clusterInst.DeletePrepare {
					return key.BeingDeletedError()
				}
				err := useMultiTenantClusterInst(stm, ctx, in, &app, sidecarApp, &clusterInst)
				if err != nil {
					return err
				}
				autoClusterType = MultiTenantAutoCluster
			}
		}

		if autoClusterSpecified {
			// we need to choose/find/create the best autocluster
			autoClusterType = ChooseAutoCluster
			if !cloudcommon.IsClusterInstReqd(&app) {
				return fmt.Errorf("No cluster required for App deployment type %s, cannot use cluster name %s which attempts to use or create a ClusterInst", app.Deployment, cloudcommon.AutoClusterPrefix)
			}
			if cloudletCompatibilityVersion < cloudcommon.CRMCompatibilityAutoReservableCluster {
				autoClusterType = DeprecatedAutoCluster
			}
			if autoClusterType != DeprecatedAutoCluster {
				// reservable and multi-tenant ClusterInsts are
				// always owned by the system
				if in.Key.ClusterInstKey.Organization != cloudcommon.OrganizationMobiledgeX {
					return fmt.Errorf("ClusterInst organization must be %s for autoclusters", cloudcommon.OrganizationMobiledgeX)
				}
			}
			if err := validateAutoDeployApp(stm, &app); err != nil {
				return err
			}
			if sidecarApp && in.RealClusterName == "" {
				return fmt.Errorf("Sidecar AppInst (AutoDelete App) must specify the RealClusterName field to deploy to the virtual cluster name %s, or specify the real cluster name in the key", in.Key.ClusterInstKey.ClusterKey.Name)
			}
		}
		if autoClusterType == ChooseAutoCluster && in.RealClusterName != "" {
			// Caller specified target ClusterInst
			if !sidecarApp {
				return fmt.Errorf("Cannot specify real cluster name with %s cluster name prefix", cloudcommon.AutoClusterPrefix)
			}
			key := in.ClusterInstKey()
			clusterInst := edgeproto.ClusterInst{}
			log.SpanLog(ctx, log.DebugLevelInfo, "look up specified ClusterInst", "name", in.RealClusterName)
			if !s.all.clusterInstApi.store.STMGet(stm, key, &clusterInst) {
				return fmt.Errorf("Specified real %s", key.NotFoundError())
			}
			if clusterInst.DeletePrepare {
				return key.BeingDeletedError()
			}
			if clusterInst.MultiTenant {
				// multi-tenant base cluster
				err := useMultiTenantClusterInst(stm, ctx, in, &app, sidecarApp, &clusterInst)
				if err != nil {
					return fmt.Errorf("Failed to use specified multi-tenant ClusterInst, %v", err)
				}
				autoClusterType = MultiTenantAutoCluster
			} else if clusterInst.Reservable {
				err := s.useReservableClusterInst(stm, ctx, in, &app, sidecarApp, &clusterInst)
				if err != nil {
					return fmt.Errorf("Failed to reserve specified reservable ClusterInst, %v", err)
				}
				autoClusterType = ReservableAutoCluster
				reservedClusterInstKey = key
			} else {
				return fmt.Errorf("Specified real cluster must be a multi-tenant or reservable ClusterInst")
			}
		}
		// Prefer multi-tenant autocluster over reservable autocluster.
		if autoClusterType == ChooseAutoCluster && app.AllowServerless {
			// if default multi-tenant cluster exists, target it
			key := in.ClusterInstKey()
			key.ClusterKey.Name = cloudcommon.DefaultMultiTenantCluster
			clusterInst := edgeproto.ClusterInst{}
			if s.all.clusterInstApi.store.STMGet(stm, key, &clusterInst) {
				if clusterInst.DeletePrepare {
					return key.BeingDeletedError()
				}
				err := useMultiTenantClusterInst(stm, ctx, in, &app, sidecarApp, &clusterInst)
				if err == nil {
					autoClusterType = MultiTenantAutoCluster
					in.RealClusterName = key.ClusterKey.Name
				}
			} else {
				err = key.NotFoundError()
			}
			log.SpanLog(ctx, log.DebugLevelInfo, "try default multi-tenant cluster check", "key", key, "err", err)
		}
		// Check for reservable cluster as the autocluster target.
		if autoClusterType == ChooseAutoCluster {
			// search for free reservable ClusterInst
			log.SpanLog(ctx, log.DebugLevelInfo, "reservable auto-cluster search", "key", in.Key)
			// search for free ClusterInst
			for _, key := range freeClusterInsts {
				cibuf := edgeproto.ClusterInst{}
				if !s.all.clusterInstApi.store.STMGet(stm, &key, &cibuf) {
					continue
				}
				if cibuf.DeletePrepare {
					return key.BeingDeletedError()
				}
				if s.useReservableClusterInst(stm, ctx, in, &app, sidecarApp, &cibuf) == nil {
					autoClusterType = ReservableAutoCluster
					reservedClusterInstKey = &key
					cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Chose reservable ClusterInst %s to deploy AppInst", cibuf.Key.ClusterKey.Name)})
					break
				}
			}
		}
		// Create reservable cluster if still no autocluster target
		if autoClusterType == ChooseAutoCluster {
			// No free reservable cluster found, create new one.
			cloudletKey := &in.Key.ClusterInstKey.CloudletKey
			refs := edgeproto.CloudletRefs{}
			if !s.all.cloudletRefsApi.store.STMGet(stm, cloudletKey, &refs) {
				initCloudletRefs(&refs, cloudletKey)
			}
			// find and reserve a free id
			id := 0
			for ; id < 64; id++ {
				mask := uint64(1) << id
				if refs.ReservedAutoClusterIds&mask != 0 {
					continue
				}
				refs.ReservedAutoClusterIds |= mask
				break
			}
			if id == 64 {
				return fmt.Errorf("Requested new reservable autocluster but maximum number reached")
			}
			s.all.cloudletRefsApi.store.STMPut(stm, &refs)
			reservedAutoClusterId = id
			autocluster = true
			autoClusterType = ReservableAutoCluster
			in.RealClusterName = fmt.Sprintf("%s%d", cloudcommon.ReservableClusterPrefix, reservedAutoClusterId)
			cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Creating new auto-cluster named %s to deploy AppInst", in.RealClusterName)})
			log.SpanLog(ctx, log.DebugLevelApi, "Creating new auto-cluster", "key", in.ClusterInstKey())
		}
		if autoClusterType == DeprecatedAutoCluster {
			// for backwards compatibility with old CRMs
			log.SpanLog(ctx, log.DebugLevelInfo, "backwards compatible dedicated auto-cluster", "key", in.Key)
			if in.RealClusterName != "" {
				return fmt.Errorf("Target cloudlet's CRM is an older version which does not support RealClusterName field; field must be left blank to create dedicated autocluster")
			}
			cibuf := edgeproto.ClusterInst{}
			if s.all.clusterInstApi.store.STMGet(stm, in.ClusterInstKey(), &cibuf) {
				// if it already exists, this means we just
				// want to spawn more apps into it
				if cibuf.DeletePrepare {
					return in.ClusterInstKey().BeingDeletedError()
				}
			} else {
				if sidecarApp {
					return fmt.Errorf("Sidecar App %s requires an existing ClusterInst", app.Key.Name)
				}
				if in.Key.AppKey.Organization != in.Key.ClusterInstKey.Organization {
					return fmt.Errorf("Developer name mismatch between App: %s and ClusterInst: %s", in.Key.AppKey.Organization, in.Key.ClusterInstKey.Organization)
				}
				autocluster = true
			}
		}
		if autoClusterType == NoAutoCluster && cloudcommon.IsClusterInstReqd(&app) {
			// Specified ClusterInst must exist
			var clusterInst edgeproto.ClusterInst
			if !s.all.clusterInstApi.store.STMGet(stm, in.ClusterInstKey(), &clusterInst) {
				return in.ClusterInstKey().NotFoundError()
			}
			if clusterInst.DeletePrepare {
				return in.ClusterInstKey().BeingDeletedError()
			}
			if clusterInst.MultiTenant && !sidecarApp {
				return fmt.Errorf("Cannot directly specify multi-tenant cluster, must use %s prefix on cluster name and specify real cluster name", cloudcommon.AutoClusterPrefix)
			}
			if clusterInst.Reservable {
				err := s.useReservableClusterInst(stm, ctx, in, &app, sidecarApp, &clusterInst)
				if err != nil {
					return fmt.Errorf("Failed to reserve specified reservable ClusterInst, %v", err)
				}
			}
			if !sidecarApp && !clusterInst.Reservable && in.Key.AppKey.Organization != in.Key.ClusterInstKey.Organization {
				return fmt.Errorf("Developer name mismatch between App: %s and ClusterInst: %s", in.Key.AppKey.Organization, in.Key.ClusterInstKey.Organization)
			}
			// cluster inst exists so we're good.
		}

		if cloudlet.TrustPolicy != "" {
			if !app.Trusted {
				return fmt.Errorf("Cannot start non trusted App on trusted cloudlet")
			}
			trustPolicy := edgeproto.TrustPolicy{}
			tpKey := edgeproto.PolicyKey{
				Name:         cloudlet.TrustPolicy,
				Organization: cloudlet.Key.Organization,
			}
			if !s.all.trustPolicyApi.store.STMGet(stm, &tpKey, &trustPolicy) {
				return errors.New("Trust Policy for cloudlet not found")
			}
			if trustPolicy.DeletePrepare {
				return tpKey.BeingDeletedError()
			}
			err = s.all.appApi.CheckAppCompatibleWithTrustPolicy(ctx, &cloudlet.Key, &app, &trustPolicy)
			if err != nil {
				return fmt.Errorf("App is not compatible with cloudlet trust policy: %v", err)
			}
		}

		// Since autoclusteripaccess is deprecated, set it to unknown
		in.AutoClusterIpAccess = edgeproto.IpAccess_IP_ACCESS_UNKNOWN

		err = validateImageTypeForPlatform(ctx, app.ImageType, cloudlet.PlatformType, cloudletFeatures)
		if err != nil {
			return err
		}

		// Now that we have a cloudlet, and cloudletInfo, we can validate the flavor requested
		vmFlavor := edgeproto.Flavor{}
		if !s.all.flavorApi.store.STMGet(stm, &in.Flavor, &vmFlavor) {
			return in.Flavor.NotFoundError()
		}
		if vmFlavor.DeletePrepare {
			return in.Flavor.BeingDeletedError()
		}
		if app.DeploymentManifest != "" {
			err = cloudcommon.IsValidDeploymentManifestForFlavor(app.Deployment, app.DeploymentManifest, &vmFlavor)
			if err != nil {
				return fmt.Errorf("Invalid deployment manifest, %v", err)
			}
		}

		vmspec, verr := s.all.resTagTableApi.GetVMSpec(ctx, stm, vmFlavor, cloudlet, info)
		if verr != nil {
			return verr
		}
		// if needed, master node flavor will be looked up from createClusterInst
		// save original in.Flavor.Name in that case
		in.VmFlavor = vmspec.FlavorName
		in.AvailabilityZone = vmspec.AvailabilityZone
		in.ExternalVolumeSize = vmspec.ExternalVolumeSize
		log.SpanLog(ctx, log.DebugLevelApi, "Selected AppInst Node Flavor", "vmspec", vmspec.FlavorName)
		in.OptRes = s.all.resTagTableApi.AddGpuResourceHintIfNeeded(ctx, stm, vmspec, cloudlet)
		in.Revision = app.Revision
		appDeploymentType = app.Deployment
		// there may be direct access apps still defined, disallow them from being instantiated.
		if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_DIRECT {
			return fmt.Errorf("Direct Access Apps are no longer supported, please re-create App as ACCESS_TYPE_LOAD_BALANCER")
		}

		if err := s.all.autoProvPolicyApi.appInstCheck(ctx, stm, cloudcommon.Create, &app, in); err != nil {
			return err
		}
		if in.Liveness == edgeproto.Liveness_LIVENESS_AUTOPROV && autoClusterType == DeprecatedAutoCluster {
			return fmt.Errorf("Target cloudlet running older version of CRM no longer supports auto-provisioning, please upgrade CRM")
		}

		if app.Deployment == cloudcommon.DeploymentTypeVM {
			refs := edgeproto.CloudletRefs{}
			if !s.all.cloudletRefsApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &refs) {
				initCloudletRefs(&refs, &in.Key.ClusterInstKey.CloudletKey)
			}
			err = s.all.clusterInstApi.validateResources(ctx, stm, nil, &app, in, &cloudlet, &info, &refs, GenResourceAlerts)
			if err != nil {
				return err
			}
			vmAppInstRefKey := edgeproto.AppInstRefKey{}
			vmAppInstRefKey.FromAppInstKey(&in.Key)
			refs.VmAppInsts = append(refs.VmAppInsts, vmAppInstRefKey)
			s.all.cloudletRefsApi.store.STMPut(stm, &refs)
		}
		// Iterate to get a unique id. The number of iterations must
		// be fairly low because the STM has a limit on the number of
		// keys it can manage.
		in.UniqueId = ""
		for ii := 0; ii < 10; ii++ {
			salt := ""
			if ii != 0 {
				salt = strconv.Itoa(ii)
			}
			id := cloudcommon.GetAppInstId(in, &app, salt)
			if s.idStore.STMHas(stm, id) {
				continue
			}
			in.UniqueId = id
			break
		}
		if in.UniqueId == "" {
			return fmt.Errorf("Unable to compute unique AppInstId, please change AppInst key values")
		}

		// Set new state to show autocluster clusterinst progress as part of
		// appinst progress
		in.State = edgeproto.TrackedState_CREATING_DEPENDENCIES
		in.Status = edgeproto.StatusInfo{}
		s.store.STMPut(stm, in)
		s.idStore.STMPut(stm, in.UniqueId)
		s.all.appInstRefsApi.addRef(stm, &in.Key)
		if cloudcommon.IsClusterInstReqd(&app) {
			s.all.clusterRefsApi.addRef(stm, in)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if reservedClusterInstKey != nil {
		clusterInstReservationEvent(ctx, cloudcommon.ReserveClusterEvent, in)
	}

	defer func() {
		if reterr == nil {
			return
		}
		// undo changes on error
		s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			var app edgeproto.App
			if !s.all.appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
				return in.Key.AppKey.NotFoundError()
			}
			refs := edgeproto.CloudletRefs{}
			refsFound := s.all.cloudletRefsApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &refs)
			refsChanged := false
			var curr edgeproto.AppInst
			if s.store.STMGet(stm, &in.Key, &curr) {
				// In case there is an error after CREATING_DEPENDENCIES state
				// is set, then delete AppInst obj directly as there is
				// no change done on CRM side
				if curr.State == edgeproto.TrackedState_CREATING_DEPENDENCIES {
					s.store.STMDel(stm, &in.Key)
					s.idStore.STMDel(stm, in.UniqueId)
					s.all.appInstRefsApi.removeRef(stm, &in.Key)
					if cloudcommon.IsClusterInstReqd(&app) {
						s.all.clusterRefsApi.removeRef(stm, in)
					}
					if app.Deployment == cloudcommon.DeploymentTypeVM {
						if refsFound {
							ii := 0
							for ; ii < len(refs.VmAppInsts); ii++ {
								aiKey := edgeproto.AppInstKey{}
								aiKey.FromAppInstRefKey(&refs.VmAppInsts[ii], &in.Key.ClusterInstKey.CloudletKey)
								if aiKey.Matches(&in.Key) {
									break
								}
							}
							if ii < len(refs.VmAppInsts) {
								// explicity zero out deleted item to
								// prevent memory leak
								a := refs.VmAppInsts
								copy(a[ii:], a[ii+1:])
								a[len(a)-1] = edgeproto.AppInstRefKey{}
								refs.VmAppInsts = a[:len(a)-1]
								refsChanged = true
							}
						}
					}
				}
			}
			// Cleanup reserved id on failure. Note that if we fail
			// after creating the auto-cluster, then deleting the
			// ClusterInst will cleanup the reserved id instead.
			if reservedAutoClusterId != -1 {
				if refsFound {
					mask := uint64(1) << reservedAutoClusterId
					refs.ReservedAutoClusterIds &^= mask
					refsChanged = true
				}
			}
			if refsFound && refsChanged {
				s.all.cloudletRefsApi.store.STMPut(stm, &refs)
			}
			// Remove reservation (if done) on failure.
			if reservedClusterInstKey != nil {
				cinst := edgeproto.ClusterInst{}
				if s.all.clusterInstApi.store.STMGet(stm, reservedClusterInstKey, &cinst) {
					cinst.ReservedBy = ""
					s.all.clusterInstApi.store.STMPut(stm, &cinst)
				}
			}
			return nil
		})
		if reservedClusterInstKey != nil {
			clusterInstReservationEvent(ctx, cloudcommon.FreeClusterEvent, in)
		}
	}()

	if autocluster {
		// auto-create cluster inst
		clusterInst.Key = *in.ClusterInstKey()
		clusterInst.Auto = true
		if autoClusterType == ReservableAutoCluster {
			clusterInst.Reservable = true
			clusterInst.ReservedBy = in.Key.AppKey.Organization
		}
		log.SpanLog(ctx, log.DebugLevelApi,
			"Create auto-ClusterInst",
			"key", clusterInst.Key,
			"AppInst", in)

		// To reduce the proliferation of different reservable ClusterInst
		// configurations, we restrict reservable ClusterInst configs.
		clusterInst.Flavor.Name = in.Flavor.Name
		// Prefer IP access shared, but some platforms (gcp, etc) only
		// support dedicated.
		clusterInst.IpAccess = edgeproto.IpAccess_IP_ACCESS_UNKNOWN
		clusterInst.Deployment = appDeploymentType
		if appDeploymentType == cloudcommon.DeploymentTypeKubernetes ||
			appDeploymentType == cloudcommon.DeploymentTypeHelm {
			clusterInst.Deployment = cloudcommon.DeploymentTypeKubernetes
			clusterInst.NumMasters = 1
			clusterInst.NumNodes = 1 // TODO support 1 master, zero nodes
			if cloudletPlatformType == edgeproto.PlatformType_PLATFORM_TYPE_K8S_BARE_METAL {
				// bare metal k8s clusters are virtual and have no nodes
				log.SpanLog(ctx, log.DebugLevelApi, "Setting num nodes to 0 for k8s baremetal virtual cluster")
				clusterInst.NumNodes = 0
			}
		}
		clusterInst.Liveness = edgeproto.Liveness_LIVENESS_DYNAMIC
		createStart := time.Now()
		cctxauto := cctx.WithAutoCluster()
		err := s.all.clusterInstApi.createClusterInstInternal(cctxauto, &clusterInst, cb)
		nodeMgr.TimedEvent(ctx, "AutoCluster create", in.Key.AppKey.Organization, node.EventType, in.Key.GetTags(), err, createStart, time.Now())
		clusterInstReservationEvent(ctx, cloudcommon.ReserveClusterEvent, in)
		if err != nil {
			return err
		}
		// disable the previous defer func for cleaning up the reserved id,
		// as the following defer func to cleanup the ClusterInst will
		// free it instead.
		reservedAutoClusterId = -1

		defer func() {
			if reterr != nil && !cctx.Undo {
				cb.Send(&edgeproto.Result{Message: "Deleting auto-ClusterInst due to failure"})
				undoErr := s.all.clusterInstApi.deleteClusterInstInternal(cctxauto.WithUndo().WithCRMUndo(), &clusterInst, cb)
				if undoErr != nil {
					log.SpanLog(ctx, log.DebugLevelApi,
						"Undo create auto-ClusterInst failed",
						"key", clusterInst.Key,
						"undoErr", undoErr)
				}
			}
		}()
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		// lookup already done, don't overwrite changes
		if s.store.STMGet(stm, &in.Key, in) {
			if in.State != edgeproto.TrackedState_CREATING_DEPENDENCIES {
				return in.Key.ExistsError()
			}
		} else {
			return fmt.Errorf("Unexpected error: AppInst %s was deleted", in.Key.GetKeyString())
		}

		// cache location of cloudlet in app inst
		in.CloudletLoc = cloudletLoc

		var app edgeproto.App
		if !s.all.appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return in.Key.AppKey.NotFoundError()
		}

		if in.Flavor.Name == "" {
			in.Flavor = app.DefaultFlavor
		}
		var clusterKey *edgeproto.ClusterKey
		ipaccess := edgeproto.IpAccess_IP_ACCESS_SHARED
		if cloudcommon.IsClusterInstReqd(&app) {
			clusterInst := edgeproto.ClusterInst{}
			if !s.all.clusterInstApi.store.STMGet(stm, in.ClusterInstKey(), &clusterInst) {
				return errors.New("ClusterInst does not exist for App")
			}
			if clusterInst.State != edgeproto.TrackedState_READY {
				return fmt.Errorf("ClusterInst %s not ready, it is %s", clusterInst.Key.GetKeyString(), clusterInst.State.String())
			}
			if !sidecarApp && clusterInst.Reservable && clusterInst.ReservedBy != in.Key.AppKey.Organization {
				return fmt.Errorf("ClusterInst reservation changed unexpectedly, expected %s but was %s", in.Key.AppKey.Organization, clusterInst.ReservedBy)
			}
			needDeployment := app.Deployment
			if app.Deployment == cloudcommon.DeploymentTypeHelm {
				needDeployment = cloudcommon.DeploymentTypeKubernetes
			}
			if clusterInst.Deployment != needDeployment {
				return fmt.Errorf("Cannot deploy %s App into %s ClusterInst", app.Deployment, clusterInst.Deployment)
			}
			ipaccess = clusterInst.IpAccess
			clusterKey = &clusterInst.Key.ClusterKey
		}

		cloudletRefs := edgeproto.CloudletRefs{}
		cloudletRefsChanged := false
		if !s.all.cloudletRefsApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudletRefs) {
			initCloudletRefs(&cloudletRefs, &in.Key.ClusterInstKey.CloudletKey)
		}

		ports, _ := edgeproto.ParseAppPorts(app.AccessPorts)
		if !cloudcommon.IsClusterInstReqd(&app) {
			in.Uri = cloudcommon.GetVMAppFQDN(&in.Key, &in.Key.ClusterInstKey.CloudletKey, *appDNSRoot)
			for ii := range ports {
				ports[ii].PublicPort = ports[ii].InternalPort
			}
		} else if in.DedicatedIp {
			// Per AppInst dedicated IP
			in.Uri = cloudcommon.GetAppFQDN(&in.Key, &in.Key.ClusterInstKey.CloudletKey, clusterKey, *appDNSRoot)
			for ii := range ports {
				ports[ii].PublicPort = ports[ii].InternalPort
			}
		} else if ipaccess == edgeproto.IpAccess_IP_ACCESS_SHARED && !app.InternalPorts {
			if cloudletCompatibilityVersion < cloudcommon.CRMCompatibilitySharedRootLBFQDN {
				// CRM has issued DNS entry only for old style FQDN.
				// This case can be removed once all CRMs have been
				// updated to current version.
				in.Uri = cloudcommon.GetRootLBFQDNOld(&in.Key.ClusterInstKey.CloudletKey, *appDNSRoot)
			} else {
				// current style FQDN
				in.Uri = cloudcommon.GetRootLBFQDN(&in.Key.ClusterInstKey.CloudletKey, *appDNSRoot)
			}
			if cloudletRefs.RootLbPorts == nil {
				cloudletRefs.RootLbPorts = make(map[int32]int32)
			}

			for ii, port := range ports {
				if port.EndPort != 0 {
					return fmt.Errorf("Shared IP access with port range not allowed")
				}
				// platos enabling layer ignores port mapping.
				// Attempt to use the internal port as the
				// external port so port remap is not required.
				protocolBits, err := getProtocolBitMap(ports[ii].Proto)
				if err != nil {
					return err
				}
				iport := ports[ii].InternalPort
				eport := int32(-1)
				if usedProtocols, found := cloudletRefs.RootLbPorts[iport]; !found || !protocolInUse(protocolBits, usedProtocols) {

					// rootLB has its own ports it uses
					// before any apps are even present.
					iport := ports[ii].InternalPort
					if iport != 22 && iport != cloudcommon.ProxyMetricsPort {
						eport = iport
					}
				}
				for p := RootLBSharedPortBegin; p < 65000 && eport == int32(-1); p++ {
					// each kubernetes service gets its own
					// nginx proxy that runs in the rootLB,
					// and http ports are also mapped to it,
					// so there is no shared L7 port + path.
					if usedProtocols, found := cloudletRefs.RootLbPorts[p]; found && protocolInUse(protocolBits, usedProtocols) {

						continue
					}
					eport = p
				}
				if eport == int32(-1) {
					return errors.New("no free external ports")
				}
				ports[ii].PublicPort = eport
				existingProtoBits, _ := cloudletRefs.RootLbPorts[eport]
				cloudletRefs.RootLbPorts[eport] = addProtocol(protocolBits, existingProtoBits)

				cloudletRefsChanged = true
			}
		} else {
			if isIPAllocatedPerService(ctx, cloudletPlatformType, cloudletFeatures, in.Key.ClusterInstKey.CloudletKey.Organization) {
				// dedicated access in which each service gets a different ip
				in.Uri = cloudcommon.GetAppFQDN(&in.Key, &in.Key.ClusterInstKey.CloudletKey, clusterKey, *appDNSRoot)
				for ii := range ports {
					ports[ii].PublicPort = ports[ii].InternalPort
				}
			} else {
				// we need to prevent overlapping ports on the dedicated rootLB
				if err = s.checkPortOverlapDedicatedLB(ports, &in.Key.ClusterInstKey); !cctx.Undo && err != nil {
					return err
				}
				// dedicated access in which IP is that of the LB
				in.Uri = cloudcommon.GetDedicatedLBFQDN(&in.Key.ClusterInstKey.CloudletKey, clusterKey, *appDNSRoot)
				for ii := range ports {
					ports[ii].PublicPort = ports[ii].InternalPort
				}
			}
		}
		if app.InternalPorts || len(ports) == 0 {
			// older CRMs require app URI regardless of external access to AppInst
			if cloudletCompatibilityVersion >= cloudcommon.CRMCompatibilitySharedRootLBFQDN {
				// no external access to AppInst, no need for URI
				in.Uri = ""
			}
		}
		if err := cloudcommon.CheckFQDNLengths("", in.Uri); err != nil {
			return err
		}
		if len(ports) > 0 {
			in.MappedPorts = ports
			if isIPAllocatedPerService(ctx, cloudletPlatformType, cloudletFeatures, in.Key.ClusterInstKey.CloudletKey.Organization) {
				setPortFQDNPrefixes(in, &app)
			}
		}

		// TODO: Make sure resources are available
		if cloudletRefsChanged {
			s.all.cloudletRefsApi.store.STMPut(stm, &cloudletRefs)
		}
		in.CreatedAt = cloudcommon.TimeToTimestamp(time.Now())

		if ignoreCRM(cctx) {
			in.State = edgeproto.TrackedState_READY
		} else {
			in.State = edgeproto.TrackedState_CREATE_REQUESTED
		}
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}
	if ignoreCRM(cctx) {
		cb.Send(&edgeproto.Result{Message: "Created AppInst successfully"})
		return nil
	}
	err = s.cache.WaitForState(ctx, &in.Key, edgeproto.TrackedState_READY,
		CreateAppInstTransitions, edgeproto.TrackedState_CREATE_ERROR,
		s.all.settingsApi.Get().CreateAppInstTimeout.TimeDuration(),
		"Created AppInst successfully", cb.Send,
		edgeproto.WithStreamObj(&s.all.streamObjApi.cache, &in.Key),
	)
	if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Create AppInst ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(ctx, in, edgeproto.TrackedState_READY)
		cb.Send(&edgeproto.Result{Message: "Created AppInst successfully"})
		err = nil
	}
	if err != nil {
		// XXX should probably track mod revision ID and only undo
		// if no other changes were made to appInst in the meantime.
		// crm failed or some other err, undo
		cb.Send(&edgeproto.Result{Message: "Deleting AppInst due to failure"})
		undoErr := s.deleteAppInstInternal(cctx.WithUndo(), in, cb)
		if undoErr != nil {
			log.InfoLog("Undo create AppInst", "undoErr", undoErr)
		}
	}
	if err == nil {
		s.updateCloudletResourcesMetric(ctx, in)
	}
	return err
}

func (s *AppInstApi) useReservableClusterInst(stm concurrency.STM, ctx context.Context, in *edgeproto.AppInst, app *edgeproto.App, sidecarApp bool, cibuf *edgeproto.ClusterInst) error {
	if !cibuf.Reservable {
		return fmt.Errorf("ClusterInst not reservable")
	}
	if sidecarApp {
		// no restrictions, no reservation
		return nil
	}
	if in.Flavor.Name != cibuf.Flavor.Name {
		return fmt.Errorf("flavor mismatch between AppInst and reservable ClusterInst")
	}
	if cibuf.ReservedBy != "" {
		return fmt.Errorf("ClusterInst already reserved")
	}
	targetDeployment := app.Deployment
	if app.Deployment == cloudcommon.DeploymentTypeHelm {
		targetDeployment = cloudcommon.DeploymentTypeKubernetes
	}
	if targetDeployment != cibuf.Deployment {
		return fmt.Errorf("deployment type mismatch between App and reservable ClusterInst")
	}
	// reserve it
	log.SpanLog(ctx, log.DebugLevelApi, "reserving ClusterInst", "cluster", cibuf.Key.ClusterKey.Name, "AppInst", in.Key)
	cibuf.ReservedBy = in.Key.AppKey.Organization
	s.all.clusterInstApi.store.STMPut(stm, cibuf)
	in.RealClusterName = cibuf.Key.ClusterKey.Name
	return nil
}

func useMultiTenantClusterInst(stm concurrency.STM, ctx context.Context, in *edgeproto.AppInst, app *edgeproto.App, sidecarApp bool, cibuf *edgeproto.ClusterInst) error {
	if !cibuf.MultiTenant {
		return fmt.Errorf("ClusterInst not multi-tenant")
	}
	if sidecarApp {
		// no restrictions, no resource check
	}
	if !app.AllowServerless {
		return fmt.Errorf("App must allow serverless deployment to deploy to multi-tenant cluster %s", in.RealClusterName)
	}
	if app.Deployment != cloudcommon.DeploymentTypeKubernetes {
		return fmt.Errorf("Deployment type must be kubernetes for multi-tenant ClusterInst")
	}
	// TODO: check and reserve resources.
	// May need to trigger adding more nodes to multi-tenant
	// cluster if not enough resources.
	return nil
}

func (s *AppInstApi) updateCloudletResourcesMetric(ctx context.Context, in *edgeproto.AppInst) {
	var err error
	metrics := []*edgeproto.Metric{}
	skipMetric := true
	resErr := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		var app edgeproto.App
		if !s.all.appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return in.Key.AppKey.NotFoundError()
		}
		skipMetric = true
		if app.Deployment == cloudcommon.DeploymentTypeVM {
			metrics, err = s.all.clusterInstApi.getCloudletResourceMetric(ctx, stm, &in.Key.ClusterInstKey.CloudletKey)
			skipMetric = false
			return err
		}
		return nil
	})
	if !skipMetric {
		if resErr == nil {
			services.cloudletResourcesInfluxQ.AddMetric(metrics...)
		} else {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to generate cloudlet resource usage metric", "clusterInstKey", in.Key, "err", resErr)
		}
	}
	return
}

func (s *AppInstApi) updateAppInstStore(ctx context.Context, in *edgeproto.AppInst) error {
	_, err := s.store.Update(ctx, in, s.sync.syncWait)
	return err
}

// refreshAppInstInternal returns true if the appinst updated, false otherwise.  False value with no error means no update was needed
func (s *AppInstApi) refreshAppInstInternal(cctx *CallContext, key edgeproto.AppInstKey, inCb edgeproto.AppInstApi_RefreshAppInstServer, forceUpdate bool) (retbool bool, reterr error) {
	ctx := inCb.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "refreshAppInstInternal", "key", key)

	updatedRevision := false
	crmUpdateRequired := false

	cloudcommon.SetAppInstKeyDefaults(&key)
	if err := key.ValidateKey(); err != nil {
		return false, err
	}

	// create stream once AppInstKey is formed correctly
	sendObj, cb, err := s.startAppInstStream(ctx, &key, inCb)
	if err == nil {
		defer func() {
			s.stopAppInstStream(ctx, &key, sendObj, reterr)
		}()
	}

	var app edgeproto.App

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		var curr edgeproto.AppInst

		if !s.all.appApi.store.STMGet(stm, &key.AppKey, &app) {
			return key.AppKey.NotFoundError()
		}
		if s.store.STMGet(stm, &key, &curr) {
			// allow UPDATE_ERROR state so updates can be retried
			if curr.State != edgeproto.TrackedState_READY && curr.State != edgeproto.TrackedState_UPDATE_ERROR {
				log.InfoLog("AppInst is not ready or update_error state for update", "state", curr.State)
				return fmt.Errorf("AppInst is not ready or update_error")
			}
			if curr.Revision != app.Revision || forceUpdate {
				crmUpdateRequired = true
				updatedRevision = true
			} else {
				return nil
			}
		} else {
			return key.NotFoundError()
		}
		if ignoreCRM(cctx) {
			crmUpdateRequired = false
		} else {
			// check cloudlet state before updating
			cloudletErr := s.all.cloudletInfoApi.checkCloudletReady(cctx, stm, &key.ClusterInstKey.CloudletKey, cloudcommon.Update)
			if crmUpdateRequired && cloudletErr != nil {
				return cloudletErr
			}
			curr.State = edgeproto.TrackedState_UPDATE_REQUESTED
		}
		curr.Status = edgeproto.StatusInfo{}
		s.store.STMPut(stm, &curr)
		return nil
	})

	if err != nil {
		return false, err
	}
	if crmUpdateRequired {
		s.RecordAppInstEvent(ctx, &key, cloudcommon.UPDATE_START, cloudcommon.InstanceDown)

		defer func() {
			if reterr == nil {
				s.RecordAppInstEvent(ctx, &key, cloudcommon.UPDATE_COMPLETE, cloudcommon.InstanceUp)
			} else {
				s.RecordAppInstEvent(ctx, &key, cloudcommon.UPDATE_ERROR, cloudcommon.InstanceDown)
			}
		}()
		err = s.cache.WaitForState(cb.Context(), &key, edgeproto.TrackedState_READY,
			UpdateAppInstTransitions, edgeproto.TrackedState_UPDATE_ERROR,
			s.all.settingsApi.Get().UpdateAppInstTimeout.TimeDuration(),
			"", cb.Send,
			edgeproto.WithStreamObj(&s.all.streamObjApi.cache, &key),
		)
	}
	if err != nil {
		return false, err
	} else {
		return updatedRevision, s.updateAppInstRevision(ctx, &key, app.Revision)
	}
}

func (s *AppInstApi) RefreshAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_RefreshAppInstServer) error {
	ctx := cb.Context()

	if in.UpdateMultiple {
		// if UpdateMultiple flag is specified, then only the appkey must be present
		if err := in.Key.AppKey.ValidateKey(); err != nil {
			return err
		}
	} else {
		// populate the clusterinst developer from the app developer if not already present
		if in.Key.ClusterInstKey.Organization == "" {
			in.Key.ClusterInstKey.Organization = in.Key.AppKey.Organization
			cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
		}

		// the whole key must be present
		cloudcommon.SetAppInstKeyDefaults(&in.Key)
		if err := in.Key.ValidateKey(); err != nil {
			return fmt.Errorf("cluster key needed without updatemultiple option: %v", err)
		}
	}

	singleAppInst := false
	if in.Key.ClusterInstKey.ValidateKey() == nil {
		// cluster inst specified
		singleAppInst = true
	}

	s.cache.Mux.Lock()

	type updateResult struct {
		errString       string
		revisionUpdated bool
	}
	instanceUpdateResults := make(map[edgeproto.AppInstKey]chan updateResult)
	instances := make(map[edgeproto.AppInstKey]struct{})

	for k, data := range s.cache.Objs {
		val := data.Obj
		// ignore forceupdate, Crmoverride updatemultiple for match
		val.ForceUpdate = in.ForceUpdate
		val.UpdateMultiple = in.UpdateMultiple
		val.CrmOverride = in.CrmOverride
		if !val.Matches(in, edgeproto.MatchFilter()) {
			continue
		}
		instances[k] = struct{}{}
		instanceUpdateResults[k] = make(chan updateResult)

	}
	s.cache.Mux.Unlock()

	if len(instances) == 0 {
		log.SpanLog(ctx, log.DebugLevelApi, "no AppInsts matched", "key", in.Key)
		return in.Key.NotFoundError()
	}

	if !singleAppInst {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Updating: %d AppInsts", len(instances))})
	}

	for instkey, _ := range instances {
		go func(k edgeproto.AppInstKey) {
			log.SpanLog(ctx, log.DebugLevelApi, "updating AppInst", "key", k)
			updated, err := s.refreshAppInstInternal(DefCallContext(), k, cb, in.ForceUpdate)
			if err == nil {
				instanceUpdateResults[k] <- updateResult{errString: "", revisionUpdated: updated}
			} else {
				instanceUpdateResults[k] <- updateResult{errString: err.Error(), revisionUpdated: updated}
			}
		}(instkey)
	}

	numUpdated := 0
	numFailed := 0
	numSkipped := 0
	numTotal := 0
	for k, r := range instanceUpdateResults {
		numTotal++
		result := <-r
		log.SpanLog(ctx, log.DebugLevelApi, "instanceUpdateResult ", "key", k, "updated", result.revisionUpdated, "error", result.errString)
		if result.errString == "" {
			if result.revisionUpdated {
				numUpdated++
				if singleAppInst {
					cb.Send(&edgeproto.Result{Message: "Successfully updated AppInst"})
				}
			} else {
				numSkipped++
				if singleAppInst {
					cb.Send(&edgeproto.Result{Message: "Skipped updating AppInst"})
				}
			}
		} else {
			numFailed++
			if singleAppInst {
				return fmt.Errorf("%s", result.errString)
			} else {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Failed for cluster (%s/%s), cloudlet (%s/%s): %s", k.ClusterInstKey.ClusterKey.Name, k.ClusterInstKey.Organization, k.ClusterInstKey.CloudletKey.Name, k.ClusterInstKey.CloudletKey.Organization, result.errString)})
			}
		}
		// give some intermediate status
		if (numTotal%10 == 0) && numTotal != len(instances) {
			cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Processing: %d of %d AppInsts.  Updated: %d Skipped: %d Failed: %d", numTotal, len(instances), numUpdated, numSkipped, numFailed)})
		}
	}
	if !singleAppInst {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Completed: %d of %d AppInsts.  Updated: %d Skipped: %d Failed: %d", numTotal, len(instances), numUpdated, numSkipped, numFailed)})
	}
	return nil
}

func (s *AppInstApi) UpdateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_UpdateAppInstServer) error {
	ctx := cb.Context()
	err := in.ValidateUpdateFields()
	if err != nil {
		return err
	}
	fmap := edgeproto.MakeFieldMap(in.Fields)
	err = in.Validate(fmap)
	if err != nil {
		return err
	}
	powerState := edgeproto.PowerState_POWER_STATE_UNKNOWN
	if _, found := fmap[edgeproto.AppInstFieldPowerState]; found {
		for _, field := range in.Fields {
			if field == edgeproto.AppInstFieldCrmOverride ||
				field == edgeproto.AppInstFieldKey ||
				field == edgeproto.AppInstFieldPowerState ||
				in.IsKeyField(field) {
				continue
			} else if _, ok := edgeproto.UpdateAppInstFieldsMap[field]; ok {
				return fmt.Errorf("If powerstate is to be updated, then no other fields can be modified")
			}
		}
		// Get the request state as user has specified action and not state
		powerState = edgeproto.GetNextPowerState(in.PowerState, edgeproto.RequestState)
		if powerState == edgeproto.PowerState_POWER_STATE_UNKNOWN {
			return fmt.Errorf("Invalid power state specified")
		}
	}

	cctx := DefCallContext()
	cctx.SetOverride(&in.CrmOverride)

	cur := edgeproto.AppInst{}
	changeCount := 0
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changeCount = cur.CopyInFields(in)
		if changeCount == 0 {
			// nothing changed
			return nil
		}
		if !ignoreCRM(cctx) && powerState != edgeproto.PowerState_POWER_STATE_UNKNOWN {
			var app edgeproto.App
			if !s.all.appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
				return in.Key.AppKey.NotFoundError()
			}
			if app.Deployment != cloudcommon.DeploymentTypeVM {
				return fmt.Errorf("Updating powerstate is only supported for VM deployment")
			}
			cur.PowerState = powerState
		}
		cur.UpdatedAt = cloudcommon.TimeToTimestamp(time.Now())
		s.store.STMPut(stm, &cur)
		return nil
	})
	if err != nil {
		return err
	}
	if changeCount == 0 {
		return nil
	}
	if ignoreCRM(cctx) {
		return nil
	}
	forceUpdate := true
	_, err = s.refreshAppInstInternal(cctx, in.Key, cb, forceUpdate)
	return err
}

func (s *AppInstApi) DeleteAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	return s.deleteAppInstInternal(DefCallContext(), in, cb)
}

func (s *AppInstApi) deleteAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, inCb edgeproto.AppInstApi_DeleteAppInstServer) (reterr error) {
	cctx.SetOverride(&in.CrmOverride)
	ctx := inCb.Context()

	var app edgeproto.App
	var reservationFreed bool
	clusterInstKey := edgeproto.ClusterInstKey{}

	setClusterOrg, setClusterName := cloudcommon.SetAppInstKeyDefaults(&in.Key)
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}

	appInstKey := in.Key
	// create stream once AppInstKey is formed correctly
	sendObj, cb, err := s.startAppInstStream(ctx, &appInstKey, inCb)
	if err == nil {
		defer func() {
			s.stopAppInstStream(ctx, &appInstKey, sendObj, reterr)
		}()
	}

	// get appinst info for flavor
	appInstInfo := edgeproto.AppInst{}
	if !s.cache.Get(&in.Key, &appInstInfo) {
		return in.Key.NotFoundError()
	}
	eventCtx := context.WithValue(ctx, in.Key, appInstInfo)
	defer func() {
		if reterr != nil {
			return
		}
		s.RecordAppInstEvent(eventCtx, &in.Key, cloudcommon.DELETED, cloudcommon.InstanceDown)
		if reservationFreed {
			s.all.clusterInstApi.RecordClusterInstEvent(ctx, &clusterInstKey, cloudcommon.UNRESERVED, cloudcommon.InstanceDown)
		}
	}()

	log.SpanLog(ctx, log.DebugLevelApi, "deleteAppInstInternal", "AppInst", in)
	// populate the clusterinst developer from the app developer if not already present
	if setClusterOrg {
		cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
	}
	if setClusterName {
		cb.Send(&edgeproto.Result{Message: "Setting ClusterInst name to default"})
	}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		// clear change tracking vars in case STM is rerun due to conflict.
		reservationFreed = false

		if !s.store.STMGet(stm, &in.Key, in) {
			// already deleted
			return in.Key.NotFoundError()
		}
		if err := validateDeleteState(cctx, "AppInst", in.State, in.Errors, cb.Send); err != nil {
			return err
		}
		if err := s.all.cloudletInfoApi.checkCloudletReady(cctx, stm, &in.Key.ClusterInstKey.CloudletKey, cloudcommon.Delete); err != nil {
			return err
		}

		var cloudlet edgeproto.Cloudlet
		if !s.all.cloudletApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudlet) {
			return fmt.Errorf("For AppInst, %v", in.Key.ClusterInstKey.CloudletKey.NotFoundError())
		}
		app = edgeproto.App{}
		if !s.all.appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return fmt.Errorf("For AppInst, %v", in.Key.AppKey.NotFoundError())
		}
		clusterInstReqd := cloudcommon.IsClusterInstReqd(&app)
		clusterInst := edgeproto.ClusterInst{}
		if clusterInstReqd && !s.all.clusterInstApi.store.STMGet(stm, in.ClusterInstKey(), &clusterInst) {
			return fmt.Errorf("For AppInst, %v", in.Key.ClusterInstKey.NotFoundError())
		}
		if err := s.all.autoProvPolicyApi.appInstCheck(ctx, stm, cloudcommon.Delete, &app, in); err != nil {
			return err
		}

		cloudletRefs := edgeproto.CloudletRefs{}
		cloudletRefsChanged := false
		hasRefs := s.all.cloudletRefsApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudletRefs)
		if hasRefs && clusterInstReqd && clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED && !app.InternalPorts {
			// shared root load balancer
			log.SpanLog(ctx, log.DebugLevelApi, "refs", "AppInst", in)
			for ii, _ := range in.MappedPorts {

				p := in.MappedPorts[ii].PublicPort
				protocol, err := getProtocolBitMap(in.MappedPorts[ii].Proto)

				if err != nil {
					return err
				}
				protos, found := cloudletRefs.RootLbPorts[p]
				if RequireAppInstPortConsistency && !found {
					return fmt.Errorf("Port %d not found in cloudlet refs %v", p, cloudletRefs.RootLbPorts)
				}
				if cloudletRefs.RootLbPorts != nil {
					if RequireAppInstPortConsistency && !protocolInUse(protos, protocol) {
						return fmt.Errorf("Port %d proto %x not found in cloudlet refs %v", p, protocol, cloudletRefs.RootLbPorts)

					}
					cloudletRefs.RootLbPorts[p] = removeProtocol(protos, protocol)
					if cloudletRefs.RootLbPorts[p] == 0 {
						delete(cloudletRefs.RootLbPorts, p)
					}
				}
				cloudletRefsChanged = true
			}
		}
		if app.Deployment == cloudcommon.DeploymentTypeVM {
			ii := 0
			for ; ii < len(cloudletRefs.VmAppInsts); ii++ {
				aiKey := edgeproto.AppInstKey{}
				aiKey.FromAppInstRefKey(&cloudletRefs.VmAppInsts[ii], &in.Key.ClusterInstKey.CloudletKey)
				if aiKey.Matches(&in.Key) {
					break
				}
			}
			if ii < len(cloudletRefs.VmAppInsts) {
				// explicity zero out deleted item to
				// prevent memory leak
				a := cloudletRefs.VmAppInsts
				copy(a[ii:], a[ii+1:])
				a[len(a)-1] = edgeproto.AppInstRefKey{}
				cloudletRefs.VmAppInsts = a[:len(a)-1]
				cloudletRefsChanged = true
			}
		}
		clusterInstKey = *in.ClusterInstKey()
		if cloudletRefsChanged {
			s.all.cloudletRefsApi.store.STMPut(stm, &cloudletRefs)
		}
		if clusterInstReqd && clusterInst.ReservedBy != "" && clusterInst.ReservedBy == in.Key.AppKey.Organization {
			clusterInst.ReservedBy = ""
			clusterInst.ReservationEndedAt = cloudcommon.TimeToTimestamp(time.Now())
			s.all.clusterInstApi.store.STMPut(stm, &clusterInst)
			reservationFreed = true
		}

		// delete app inst
		if ignoreCRM(cctx) {
			// CRM state should be the same as before the
			// operation failed, so just need to clean up
			// controller state.
			s.store.STMDel(stm, &in.Key)
			s.idStore.STMDel(stm, in.UniqueId)
			s.all.appInstRefsApi.removeRef(stm, &in.Key)
			if cloudcommon.IsClusterInstReqd(&app) {
				s.all.clusterRefsApi.removeRef(stm, in)
			}
			// delete associated streamobj as well
			s.all.streamObjApi.store.STMDel(stm, &in.Key)
		} else {
			in.State = edgeproto.TrackedState_DELETE_REQUESTED
			in.Status = edgeproto.StatusInfo{}
			s.store.STMPut(stm, in)
			s.all.appInstRefsApi.addDeleteRequestedRef(stm, &in.Key)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if reservationFreed {
		clusterInstReservationEvent(ctx, cloudcommon.FreeClusterEvent, in)
	}
	// clear all alerts for this appInst
	s.all.alertApi.CleanupAppInstAlerts(ctx, &appInstKey)
	if ignoreCRM(cctx) {
		cb.Send(&edgeproto.Result{Message: "Deleted AppInst successfully"})
	} else {
		err = s.cache.WaitForState(ctx, &in.Key, edgeproto.TrackedState_NOT_PRESENT,
			DeleteAppInstTransitions, edgeproto.TrackedState_DELETE_ERROR,
			s.all.settingsApi.Get().DeleteAppInstTimeout.TimeDuration(),
			"Deleted AppInst successfully", cb.Send,
			edgeproto.WithStreamObj(&s.all.streamObjApi.cache, &in.Key),
		)
		if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
			cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Delete AppInst ignoring CRM failure: %s", err.Error())})
			s.ReplaceErrorState(ctx, in, edgeproto.TrackedState_DELETE_DONE)
			cb.Send(&edgeproto.Result{Message: "Deleted AppInst successfully"})
			err = nil
		}
		if err != nil {
			// crm failed or some other err, undo
			cb.Send(&edgeproto.Result{Message: "Recreating AppInst due to failure"})
			undoErr := s.createAppInstInternal(cctx.WithUndo(), in, cb)
			if undoErr != nil {
				log.InfoLog("Undo delete AppInst", "undoErr", undoErr)
			}
			return err
		}
		if err == nil {
			s.updateCloudletResourcesMetric(ctx, in)
		}
	}
	// delete clusterinst afterwards if it was auto-created and nobody is left using it
	// this is retained for old autoclusters that are not reservable,
	// and can be removed once no old autoclusters exist anymore.
	clusterInst := edgeproto.ClusterInst{}
	if s.all.clusterInstApi.Get(&clusterInstKey, &clusterInst) && clusterInst.Auto && !s.UsesClusterInst(in.Key.AppKey.Organization, &clusterInstKey) && !clusterInst.Reservable {
		cb.Send(&edgeproto.Result{Message: "Deleting auto-ClusterInst"})
		cctxauto := cctx.WithAutoCluster()
		autoerr := s.all.clusterInstApi.deleteClusterInstInternal(cctxauto, &clusterInst, cb)
		if autoerr != nil {
			log.InfoLog("Failed to delete auto-ClusterInst",
				"clusterInst", clusterInst, "err", autoerr)
		}
	}
	return err
}

func (s *AppInstApi) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInst) error {
		obj.Status = edgeproto.StatusInfo{}
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AppInstApi) HealthCheckUpdate(ctx context.Context, in *edgeproto.AppInst, state dme.HealthCheck) {
	log.SpanLog(ctx, log.DebugLevelApi, "Update AppInst Health Check", "key", in.Key, "state", state)
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			log.SpanLog(ctx, log.DebugLevelApi, "AppInst not found updating health check", "appinst", in)
			// got deleted in the meantime
			return nil
		}
		if inst.HealthCheck == state {
			log.SpanLog(ctx, log.DebugLevelApi, "AppInst state is already set", "appinst", inst, "state", state)
			// nothing to do
			return nil
		}
		if inst.HealthCheck == dme.HealthCheck_HEALTH_CHECK_OK && state != dme.HealthCheck_HEALTH_CHECK_OK {
			// healthy -> not healthy
			s.RecordAppInstEvent(ctx, &inst.Key, cloudcommon.HEALTH_CHECK_FAIL, cloudcommon.InstanceDown)
			nodeMgr.Event(ctx, "AppInst offline", in.Key.AppKey.Organization, in.Key.GetTags(), nil, "state", state.String())
		} else if inst.HealthCheck != dme.HealthCheck_HEALTH_CHECK_OK && state == dme.HealthCheck_HEALTH_CHECK_OK {
			// not healthy -> healthy
			s.RecordAppInstEvent(ctx, &inst.Key, cloudcommon.HEALTH_CHECK_OK, cloudcommon.InstanceUp)
			nodeMgr.Event(ctx, "AppInst online", in.Key.AppKey.Organization, in.Key.GetTags(), nil, "state", state.String())
		}
		inst.HealthCheck = state
		s.store.STMPut(stm, &inst)
		return nil
	})
}

func (s *AppInstApi) UpdateFromInfo(ctx context.Context, in *edgeproto.AppInstInfo) {
	log.SpanLog(ctx, log.DebugLevelApi, "Update AppInst from info", "key", in.Key, "state", in.State, "status", in.Status, "powerstate", in.PowerState, "uri", in.Uri)

	// update only diff of status msgs
	s.all.streamObjApi.UpdateStatus(ctx, &in.Status, &in.Key)

	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		applyUpdate := false
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		if in.PowerState != edgeproto.PowerState_POWER_STATE_UNKNOWN &&
			inst.PowerState != in.PowerState {
			inst.PowerState = in.PowerState
			applyUpdate = true
		}
		// If AppInst is ready and state has not been set yet by HealthCheckUpdate, default to Ok.
		if in.State == edgeproto.TrackedState_READY &&
			inst.HealthCheck == dme.HealthCheck_HEALTH_CHECK_UNKNOWN {
			inst.HealthCheck = dme.HealthCheck_HEALTH_CHECK_OK
			applyUpdate = true
		}

		if in.Uri != "" && inst.Uri != in.Uri {
			inst.Uri = in.Uri
			applyUpdate = true
		}

		if inst.State == in.State {
			// already in that state
			if in.State == edgeproto.TrackedState_READY {
				// update runtime info
				if len(in.RuntimeInfo.ContainerIds) > 0 {
					inst.RuntimeInfo = in.RuntimeInfo
					applyUpdate = true
				}
			}
		} else {
			// please see state_transitions.md
			if !crmTransitionOk(inst.State, in.State) {
				log.SpanLog(ctx, log.DebugLevelApi, "Invalid state transition",
					"key", &in.Key, "cur", inst.State, "next", in.State)
				return nil
			}
			if inst.State == edgeproto.TrackedState_DELETE_REQUESTED && in.State != edgeproto.TrackedState_DELETE_REQUESTED {
				s.all.appInstRefsApi.removeDeleteRequestedRef(stm, &in.Key)
			}
			inst.State = in.State
			applyUpdate = true
		}
		if in.State == edgeproto.TrackedState_CREATE_ERROR || in.State == edgeproto.TrackedState_DELETE_ERROR || in.State == edgeproto.TrackedState_UPDATE_ERROR {
			inst.Errors = in.Errors
		} else {
			inst.Errors = nil
		}

		inst.RuntimeInfo = in.RuntimeInfo
		if applyUpdate {
			s.store.STMPut(stm, &inst)
		}
		return nil
	})
	if in.State == edgeproto.TrackedState_DELETE_DONE {
		s.DeleteFromInfo(ctx, in)
	}
}

func (s *AppInstApi) DeleteFromInfo(ctx context.Context, in *edgeproto.AppInstInfo) {
	log.SpanLog(ctx, log.DebugLevelApi, "Delete AppInst from info", "key", in.Key, "state", in.State)
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		// please see state_transitions.md
		if inst.State != edgeproto.TrackedState_DELETING && inst.State != edgeproto.TrackedState_DELETE_REQUESTED &&
			inst.State != edgeproto.TrackedState_DELETE_DONE {
			log.SpanLog(ctx, log.DebugLevelApi, "Invalid state transition",
				"key", &in.Key, "cur", inst.State,
				"next", edgeproto.TrackedState_DELETE_DONE)
			return nil
		}
		s.store.STMDel(stm, &in.Key)
		s.idStore.STMDel(stm, inst.UniqueId)
		s.all.appInstRefsApi.removeRef(stm, &in.Key)
		s.all.clusterRefsApi.removeRef(stm, &inst)

		// delete associated streamobj as well
		s.all.streamObjApi.store.STMDel(stm, &in.Key)
		return nil
	})
}

func (s *AppInstApi) ReplaceErrorState(ctx context.Context, in *edgeproto.AppInst, newState edgeproto.TrackedState) {
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		if inst.State != edgeproto.TrackedState_CREATE_ERROR &&
			inst.State != edgeproto.TrackedState_DELETE_ERROR &&
			inst.State != edgeproto.TrackedState_UPDATE_ERROR {
			return nil
		}
		if newState == edgeproto.TrackedState_DELETE_DONE {
			s.store.STMDel(stm, &in.Key)
			s.idStore.STMDel(stm, inst.UniqueId)
			s.all.appInstRefsApi.removeRef(stm, &in.Key)
			s.all.clusterRefsApi.removeRef(stm, &inst)
		} else {
			inst.State = newState
			inst.Errors = nil
			s.store.STMPut(stm, &inst)
		}
		return nil
	})
}

// public cloud k8s cluster allocates a separate IP per service.  This is a type of dedicated access
func isIPAllocatedPerService(ctx context.Context, platformType edgeproto.PlatformType, features *platform.Features, operator string) bool {
	log.SpanLog(ctx, log.DebugLevelApi, "isIPAllocatedPerService", "platformType", platformType, "operator", operator)

	if features.IsFake {
		// for a fake cloudlet used in testing, decide based on operator name
		return operator == cloudcommon.OperatorGCP || operator == cloudcommon.OperatorAzure || operator == cloudcommon.OperatorAWS
	}
	return features.IPAllocatedPerService
}

func validateImageTypeForPlatform(ctx context.Context, imageType edgeproto.ImageType, platformType edgeproto.PlatformType, features *platform.Features) error {
	log.SpanLog(ctx, log.DebugLevelApi, "validateImageTypeForPlatform", "imageType", imageType, "platformType", platformType)
	if imageType == edgeproto.ImageType_IMAGE_TYPE_OVF {
		if !features.SupportsImageTypeOVF {
			platName := edgeproto.PlatformType_name[int32(platformType)]
			return fmt.Errorf("image type %s is not valid for platform type: %s", imageType, platName)
		}
	}
	return nil
}

func allocateIP(ctx context.Context, inst *edgeproto.ClusterInst, cloudlet *edgeproto.Cloudlet, platformType edgeproto.PlatformType, features *platform.Features, refs *edgeproto.CloudletRefs) error {

	if isIPAllocatedPerService(ctx, platformType, features, cloudlet.Key.Organization) {
		// we don't track IPs in managed k8s clouds
		return nil
	}
	if inst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
		// shared, so no allocation needed
		return nil
	}
	if inst.IpAccess == edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
		// This should have been modified already before coming here, this is a bug if this is hit
		return fmt.Errorf("Unexpected IP_ACCESS_UNKNOWN ")
	}
	// Allocate a dedicated IP
	if cloudlet.IpSupport == edgeproto.IpSupport_IP_SUPPORT_STATIC {
		// TODO:
		// parse cloudlet.StaticIps and refs.UsedStaticIps.
		// pick a free one, put it in refs.UsedStaticIps, and
		// set inst.AllocatedIp to the Ip.
		return errors.New("Static IPs not supported yet")
	} else if cloudlet.IpSupport == edgeproto.IpSupport_IP_SUPPORT_DYNAMIC {
		// Note one dynamic IP is reserved for Global Reverse Proxy LB.
		if refs.UsedDynamicIps+1 >= cloudlet.NumDynamicIps {
			return errors.New("No more dynamic IPs left")
		}
		refs.UsedDynamicIps++
		inst.AllocatedIp = cloudcommon.AllocatedIpDynamic
		return nil
	}
	return errors.New("Invalid IpSupport type")
}

func freeIP(inst *edgeproto.ClusterInst, cloudlet *edgeproto.Cloudlet, refs *edgeproto.CloudletRefs) {
	if inst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
		return
	}
	if cloudlet.IpSupport == edgeproto.IpSupport_IP_SUPPORT_STATIC {
		// TODO: free static ip in inst.AllocatedIp from refs.
	} else if cloudlet.IpSupport == edgeproto.IpSupport_IP_SUPPORT_DYNAMIC {
		refs.UsedDynamicIps--
		inst.AllocatedIp = ""
	}
}

func setPortFQDNPrefixes(in *edgeproto.AppInst, app *edgeproto.App) error {
	// For Kubernetes deployments, the CRM sets the
	// Fqdn based on the service (load balancer) name
	// in the kubernetes deployment manifest.
	// The Controller needs to set a matching
	// FqdnPrefix on the ports so the DME can tell the
	// App Client the correct Fqdn for a given port.
	if app.Deployment == cloudcommon.DeploymentTypeKubernetes {
		objs, _, err := cloudcommon.DecodeK8SYaml(app.DeploymentManifest)
		if err != nil {
			return fmt.Errorf("invalid kubernetes deployment yaml, %s", err.Error())
		}
		for ii, _ := range in.MappedPorts {
			setPortFQDNPrefix(&in.MappedPorts[ii], objs)
			if err := cloudcommon.CheckFQDNLengths(in.MappedPorts[ii].FqdnPrefix, in.Uri); err != nil {
				return err
			}
		}
	}
	return nil
}

func setPortFQDNPrefix(port *dme.AppPort, objs []runtime.Object) {
	for _, obj := range objs {
		ksvc, ok := obj.(*v1.Service)
		if !ok {
			continue
		}
		for _, kp := range ksvc.Spec.Ports {
			lproto, err := edgeproto.LProtoStr(port.Proto)
			if err != nil {
				return
			}
			if lproto != strings.ToLower(string(kp.Protocol)) {
				continue
			}
			if kp.TargetPort.IntValue() == int(port.InternalPort) {
				port.FqdnPrefix = cloudcommon.FqdnPrefix(ksvc.Name)
				return
			}
		}
	}
}

func (s *AppInstApi) RecordAppInstEvent(ctx context.Context, appInstKey *edgeproto.AppInstKey, event cloudcommon.InstanceEvent, serverStatus string) {
	metric := edgeproto.Metric{}
	metric.Name = cloudcommon.AppInstEvent
	now := time.Now()
	ts, _ := types.TimestampProto(now)
	metric.Timestamp = *ts
	metric.AddStringVal("cloudletorg", appInstKey.ClusterInstKey.CloudletKey.Organization)
	metric.AddTag("cloudlet", appInstKey.ClusterInstKey.CloudletKey.Name)
	metric.AddTag("cluster", appInstKey.ClusterInstKey.ClusterKey.Name)
	metric.AddTag("clusterorg", appInstKey.ClusterInstKey.Organization)
	metric.AddTag("apporg", appInstKey.AppKey.Organization)
	metric.AddTag("app", appInstKey.AppKey.Name)
	metric.AddTag("ver", appInstKey.AppKey.Version)
	metric.AddStringVal("event", string(event))
	metric.AddStringVal("status", serverStatus)

	app := edgeproto.App{}
	if !s.all.appApi.cache.Get(&appInstKey.AppKey, &app) {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Cannot find appdata for app", "app", appInstKey.AppKey)
		return
	}
	metric.AddStringVal("deployment", app.Deployment)

	// have to grab the appinst here because its now possible to create apps without a default flavor
	// on deletes, the appinst is passed into the context otherwise we wont be able to get it
	appInst, ok := ctx.Value(*appInstKey).(edgeproto.AppInst)
	if !ok {
		if !s.cache.Get(appInstKey, &appInst) {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Cannot find appinst data for app", "app", appInstKey.AppKey)
			return
		}
	}
	if app.Deployment == cloudcommon.DeploymentTypeVM {
		metric.AddStringVal("flavor", appInst.Flavor.Name)
	}
	metric.AddStringVal("realcluster", appInst.GetRealClusterName())

	services.events.AddMetric(&metric)
}

func clusterInstReservationEvent(ctx context.Context, eventName string, appInst *edgeproto.AppInst) {
	realClusterName := appInst.RealClusterName
	if realClusterName == "" {
		realClusterName = appInst.Key.ClusterInstKey.ClusterKey.Name
	}
	nodeMgr.Event(ctx, eventName, appInst.Key.AppKey.Organization, appInst.Key.GetTags(), nil, "realcluster", realClusterName)
}
