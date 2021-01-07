package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AppInstApi struct {
	sync  *Sync
	store edgeproto.AppInstStore
	cache edgeproto.AppInstCache
}

const RootLBSharedPortBegin int32 = 10000

var RequireAppInstPortConsistency = false

var appInstApi = AppInstApi{}

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

func InitAppInstApi(sync *Sync) {
	appInstApi.sync = sync
	appInstApi.store = edgeproto.NewAppInstStore(sync.store)
	edgeproto.InitAppInstCache(&appInstApi.cache)
	sync.RegisterCache(&appInstApi.cache)
}

func (s *AppInstApi) Get(key *edgeproto.AppInstKey, val *edgeproto.AppInst) bool {
	return s.cache.Get(key, val)
}

func (s *AppInstApi) HasKey(key *edgeproto.AppInstKey) bool {
	return s.cache.HasKey(key)
}

func (s *AppInstApi) UsesCloudlet(in *edgeproto.CloudletKey, dynInsts map[edgeproto.AppInstKey]struct{}) bool {
	var app edgeproto.App
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	static := false
	for key, data := range s.cache.Objs {
		val := data.Obj
		if key.ClusterInstKey.CloudletKey.Matches(in) && appApi.Get(&val.Key.AppKey, &app) {
			if (val.Liveness == edgeproto.Liveness_LIVENESS_STATIC || val.Liveness == edgeproto.Liveness_LIVENESS_AUTOPROV) && (app.DelOpt == edgeproto.DeleteType_NO_AUTO_DELETE) {
				static = true
				//if can autodelete it then also add it to the dynInsts to be deleted later
			} else if (val.Liveness == edgeproto.Liveness_LIVENESS_DYNAMIC) || (app.DelOpt == edgeproto.DeleteType_AUTO_DELETE) {
				dynInsts[key] = struct{}{}
			}
		}
	}
	return static
}

func (s *AppInstApi) CheckCloudletAppinstsCompatibleWithTrustPolicy(ckey *edgeproto.CloudletKey, TrustPolicy *edgeproto.TrustPolicy) error {
	apps := make(map[edgeproto.AppKey]*edgeproto.App)
	appApi.GetAllApps(apps)
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
		err := CheckAppCompatibleWithTrustPolicy(app, TrustPolicy)
		if err != nil {
			return err
		}
	}
	return nil
}

// Checks if there is some action in progress by AppInst on the cloudlet
func (s *AppInstApi) UsingCloudlet(in *edgeproto.CloudletKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, data := range s.cache.Objs {
		val := data.Obj
		if key.ClusterInstKey.CloudletKey.Matches(in) {
			if edgeproto.IsTransientState(val.State) {
				return true
			}
		}
	}
	return false
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

func (s *AppInstApi) UsesApp(in *edgeproto.AppKey, dynInsts map[edgeproto.AppInstKey]struct{}) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	static := false
	for key, data := range s.cache.Objs {
		val := data.Obj
		if key.AppKey.Matches(in) {
			if val.Liveness == edgeproto.Liveness_LIVENESS_STATIC || val.Liveness == edgeproto.Liveness_LIVENESS_AUTOPROV {
				static = true
			} else if val.Liveness == edgeproto.Liveness_LIVENESS_DYNAMIC {
				dynInsts[key] = struct{}{}
			}
		}
	}
	return static
}

func (s *AppInstApi) UsesClusterInst(in *edgeproto.ClusterInstKey) bool {
	var app edgeproto.App
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, data := range s.cache.Objs {
		val := data.Obj
		if key.ClusterInstKey.Matches(in) && appApi.Get(&val.Key.AppKey, &app) {
			if val.Liveness == edgeproto.Liveness_LIVENESS_DYNAMIC {
				continue
			}
			log.DebugLog(log.DebugLevelApi, "AppInst found for clusterInst", "app", app.Key.Name, "autodelete", app.DelOpt.String())
			if app.DelOpt == edgeproto.DeleteType_NO_AUTO_DELETE {
				return true
			}
		}
	}
	return false
}

func (s *AppInstApi) AutoDeleteAppInsts(key *edgeproto.ClusterInstKey, crmoverride edgeproto.CRMOverride, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	var app edgeproto.App
	var err error
	apps := make(map[edgeproto.AppInstKey]*edgeproto.AppInst)
	log.DebugLog(log.DebugLevelApi, "Auto-deleting AppInsts ", "cluster", key.ClusterKey.Name)
	s.cache.Mux.Lock()
	keys := []edgeproto.AppInstKey{}
	for k, data := range s.cache.Objs {
		val := data.Obj
		if k.ClusterInstKey.Matches(key) && appApi.Get(&val.Key.AppKey, &app) {
			if app.DelOpt == edgeproto.DeleteType_AUTO_DELETE || val.Liveness == edgeproto.Liveness_LIVENESS_DYNAMIC {
				apps[k] = val
				keys = append(keys, k)
			}
		}
	}
	s.cache.Mux.Unlock()

	// sort keys for stable iteration order, needed for testing
	sort.Slice(keys[:], func(i, j int) bool {
		return keys[i].GetKeyString() < keys[j].GetKeyString()
	})

	//Spin in case cluster was just created and apps are still in the creation process and cannot be deleted
	var spinTime time.Duration
	start := time.Now()
	for _, key := range keys {
		val := apps[key]
		log.DebugLog(log.DebugLevelApi, "Auto-deleting AppInst ", "appinst", val.Key.AppKey.Name)
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
			err = s.deleteAppInstInternal(cctx, val, cb)
			if err != nil && err.Error() == val.Key.NotFoundError().Error() {
				err = nil
				break
			}
			if err != nil && err.Error() == "AppInst busy, cannot delete" {
				spinTime = time.Since(start)
				if spinTime > settingsApi.Get().DeleteAppInstTimeout.TimeDuration() {
					log.DebugLog(log.DebugLevelApi, "Timeout while waiting for App", "appName", val.Key.AppKey.Name)
					return err
				}
				log.DebugLog(log.DebugLevelApi, "AppInst busy, retrying in 0.5s...", "appName", val.Key.AppKey.Name)
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

func (s *AppInstApi) AutoDelete(ctx context.Context, appinsts map[edgeproto.AppInstKey]*edgeproto.AppInst) error {
	failed := 0
	deleted := 0
	for key, val := range appinsts {
		log.SpanLog(ctx, log.DebugLevelApi, "Auto-delete AppInst for App", "AppInst", key)
		stream := streamoutAppInst{}
		stream.ctx = ctx
		stream.debugLvl = log.DebugLevelApi
		err := s.DeleteAppInst(val, &stream)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to auto-delete AppInst", "AppInst", key)
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

func (s *AppInstApi) UsesFlavor(key *edgeproto.FlavorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		app := data.Obj
		if app.Flavor.Matches(key) {
			return true
		}
	}
	return false
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

func (s *AppInstApi) setDefaultVMClusterKey(ctx context.Context, key *edgeproto.AppInstKey) {
	// If ClusterKey.Name already exists, then don't set
	// any default value for it
	if key.ClusterInstKey.ClusterKey.Name != "" {
		return
	}
	var app edgeproto.App
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !appApi.store.STMGet(stm, &key.AppKey, &app) {
			return key.AppKey.NotFoundError()
		}
		return nil
	})
	if err != nil {
		return
	}
	if app.Deployment == cloudcommon.DeploymentTypeVM {
		key.ClusterInstKey.ClusterKey.Name = cloudcommon.DefaultVMCluster
	}
}

func startAppInstStream(ctx context.Context, key *edgeproto.AppInstKey, inCb edgeproto.AppInstApi_CreateAppInstServer) (*streamSend, edgeproto.AppInstApi_CreateAppInstServer, error) {
	streamSendObj, err := streamObjApi.startStream(ctx, key, inCb)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to start appinst stream", "err", err)
		return nil, inCb, err
	}
	return streamSendObj, &CbWrapper{
		streamSendObj: streamSendObj,
		GenericCb:     inCb,
	}, nil
}

func stopAppInstStream(ctx context.Context, key *edgeproto.AppInstKey, streamSendObj *streamSend, objErr error) {
	if err := streamObjApi.stopStream(ctx, key, streamSendObj, objErr); err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "failed to stop appinst stream", "err", err)
	}
}

func (s *StreamObjApi) StreamAppInst(key *edgeproto.AppInstKey, cb edgeproto.StreamObjApi_StreamAppInstServer) error {
	// populate the clusterinst developer from the app developer if not already present
	if key.ClusterInstKey.Organization == "" {
		key.ClusterInstKey.Organization = key.AppKey.Organization
	}
	appInstApi.setDefaultVMClusterKey(cb.Context(), key)
	if key.ClusterInstKey.ClusterKey.Name == "" {
		// if cluster name is still empty, fill it with
		// default vm cluster name
		key.ClusterInstKey.ClusterKey.Name = cloudcommon.DefaultVMCluster
	}
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

// createAppInstInternal is used to create dynamic app insts internally,
// bypassing static assignment.
func (s *AppInstApi) createAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, inCb edgeproto.AppInstApi_CreateAppInstServer) (reterr error) {
	var clusterInst edgeproto.ClusterInst
	ctx := inCb.Context()

	// populate the clusterinst developer from the app developer if not already present
	setClusterOrg := false
	if in.Key.ClusterInstKey.Organization == "" {
		in.Key.ClusterInstKey.Organization = in.Key.AppKey.Organization
		setClusterOrg = true
	}
	s.setDefaultVMClusterKey(ctx, &in.Key)

	appInstKey := in.Key

	// create stream once AppInstKey is formed correctly
	sendObj, cb, err := startAppInstStream(ctx, &appInstKey, inCb)
	if err == nil {
		defer func() {
			stopAppInstStream(ctx, &appInstKey, sendObj, reterr)
		}()
	}

	defer func() {
		if reterr == nil {
			RecordAppInstEvent(ctx, &in.Key, cloudcommon.CREATED, cloudcommon.InstanceUp)
		}
	}()

	if setClusterOrg {
		cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
	}

	if err := in.Key.ValidateKey(); err != nil {
		return err
	}

	if in.Liveness == edgeproto.Liveness_LIVENESS_UNKNOWN {
		in.Liveness = edgeproto.Liveness_LIVENESS_STATIC
	}
	cctx.SetOverride(&in.CrmOverride)

	var autocluster bool
	tenant := isTenantAppInst(&in.Key)
	// See if we need to create auto-cluster.
	// This also sets up the correct ClusterInstKey in "in".

	if in.Key.AppKey.Organization != in.Key.ClusterInstKey.Organization && in.Key.AppKey.Organization != cloudcommon.OrganizationMobiledgeX && !tenant {
		return fmt.Errorf("Developer name mismatch between App: %s and ClusterInst: %s", in.Key.AppKey.Organization, in.Key.ClusterInstKey.Organization)
	}
	err = s.checkForAppinstCollisions(ctx, &in.Key)
	if err != nil {
		return err
	}
	appDeploymentType := ""

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		autocluster = false
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
		// make sure cloudlet exists
		cloudlet := edgeproto.Cloudlet{}
		if !cloudletApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudlet) {
			return errors.New("Specified Cloudlet not found")
		}
		if err := checkCloudletReady(cctx, stm, &in.Key.ClusterInstKey.CloudletKey); err != nil {
			return err
		}
		cikey := &in.Key.ClusterInstKey
		// Explicit auto-cluster requirement
		if cikey.ClusterKey.Name == "" {
			return fmt.Errorf("No cluster name specified. Create one first or use \"%s\" as the name to automatically create a ClusterInst", ClusterAutoPrefix)
		}
		var app edgeproto.App
		if !appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return in.Key.AppKey.NotFoundError()
		}
		if app.DeletePrepare {
			return fmt.Errorf("Cannot create AppInst against App which is being deleted")
		}
		if cloudlet.TrustPolicy != "" && !app.Trusted {
			return fmt.Errorf("Cannot start non Trusted App on Trusted cloudlet")
		}
		if app.Deployment == cloudcommon.DeploymentTypeVM && in.AutoClusterIpAccess != edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
			return fmt.Errorf("Cannot specify AutoClusterIpAccess if deployment type is VM")
		}

		// Now that we have a cloudlet, and cloudletInfo, we can validate the flavor requested
		if in.Flavor.Name == "" {
			in.Flavor = app.DefaultFlavor
			if in.Flavor.Name == "" {
				return fmt.Errorf("No AppInst or App flavor specified")
			}
		}
		vmFlavor := edgeproto.Flavor{}
		if !flavorApi.store.STMGet(stm, &in.Flavor, &vmFlavor) {
			return in.Flavor.NotFoundError()
		}
		info := edgeproto.CloudletInfo{}
		if !cloudletInfoApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &info) {
			return fmt.Errorf("No resource information found for Cloudlet %s", in.Key.ClusterInstKey.CloudletKey)
		}
		vmspec, verr := resTagTableApi.GetVMSpec(ctx, stm, vmFlavor, cloudlet, info)
		if verr != nil {
			return verr
		}
		// if needed, master node flavor will be looked up from createClusterInst
		// save original in.Flavor.Name in that case
		in.VmFlavor = vmspec.FlavorName
		in.AvailabilityZone = vmspec.AvailabilityZone
		in.ExternalVolumeSize = vmspec.ExternalVolumeSize
		log.SpanLog(ctx, log.DebugLevelApi, "Selected AppInst Node Flavor", "vmspec", vmspec.FlavorName)

		if resTagTableApi.UsesGpu(ctx, stm, *vmspec.FlavorInfo, cloudlet) {
			in.OptRes = "gpu"
		}

		in.Revision = app.Revision
		appDeploymentType = app.Deployment
		if in.AutoClusterIpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED && app.AccessType == edgeproto.AccessType_ACCESS_TYPE_DIRECT {
			return fmt.Errorf("Cannot specify AutoClusterIpAccess as IP_ACCESS_SHARED if App access type is ACCESS_TYPE_DIRECT")
		}
		// Check if specified ClusterInst exists

		var clusterInst edgeproto.ClusterInst
		if !strings.HasPrefix(cikey.ClusterKey.Name, ClusterAutoPrefix) && cloudcommon.IsClusterInstReqd(&app) {
			found := clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, &clusterInst)
			if !found && in.Key.ClusterInstKey.Organization == "" {
				// developer may not be specified
				// in clusterinst.
				in.Key.ClusterInstKey.Organization = in.Key.AppKey.Organization
				found = clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, &clusterInst)
				if found {
					cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
				}
			}
			if !found {
				return errors.New("Specified ClusterInst not found")
			}
			if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_DIRECT && clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
				return fmt.Errorf("Direct Access App cannot be deployed on IP_ACCESS_SHARED ClusterInst")
			}
			// cluster inst exists so we're good.
		} else if cloudcommon.IsClusterInstReqd(&app) {
			// Auto-cluster
			if clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, nil) {
				// if it already exists, this means we just want to spawn more apps into it
			} else {
				// If this is an autodelete app, we should only allow those in existing cluster instances
				if app.DelOpt == edgeproto.DeleteType_AUTO_DELETE {
					return fmt.Errorf("Autodelete App %s requires an existing ClusterInst", app.Key.Name)
				}
				cikey.ClusterKey.Name = util.K8SSanitize(cikey.ClusterKey.Name)
				if cikey.Organization == "" {
					cikey.Organization = in.Key.AppKey.Organization
				}
				autocluster = true
			}
		}

		if in.SharedVolumeSize == 0 {
			in.SharedVolumeSize = app.DefaultSharedVolumeSize
		}
		if err := autoProvPolicyApi.appInstCheck(ctx, stm, cloudcommon.Create, &app, in); err != nil {
			return err
		}

		// Set new state to show autocluster clusterinst progress as part of
		// appinst progress
		in.State = edgeproto.TrackedState_CREATING_DEPENDENCIES
		in.Status = edgeproto.StatusInfo{}
		s.store.STMPut(stm, in)
		appInstRefsApi.addRef(stm, &in.Key)
		return nil
	})
	if err != nil {
		return err
	}

	defer func() {
		if reterr != nil {
			s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
				var curr edgeproto.AppInst
				if s.store.STMGet(stm, &in.Key, &curr) {
					// In case there is an error after CREATING_DEPENDENCIES state
					// is set, then delete AppInst obj directly as there is
					// no change done on CRM side
					if curr.State == edgeproto.TrackedState_CREATING_DEPENDENCIES {
						s.store.STMDel(stm, &in.Key)
						appInstRefsApi.removeRef(stm, &in.Key)
					}
				}
				return nil
			})
		}
	}()

	if autocluster {
		// auto-create cluster inst
		clusterInst.Key = in.Key.ClusterInstKey
		clusterInst.Auto = true
		log.SpanLog(ctx, log.DebugLevelApi,
			"Create auto-ClusterInst",
			"key", clusterInst.Key,
			"AppInst", in)

		clusterInst.Flavor.Name = in.Flavor.Name
		clusterInst.IpAccess = in.AutoClusterIpAccess
		clusterInst.Deployment = appDeploymentType
		clusterInst.SharedVolumeSize = in.SharedVolumeSize
		if appDeploymentType == cloudcommon.DeploymentTypeKubernetes ||
			appDeploymentType == cloudcommon.DeploymentTypeHelm {
			clusterInst.Deployment = cloudcommon.DeploymentTypeKubernetes
			clusterInst.NumMasters = 1
			clusterInst.NumNodes = 1 // TODO support 1 master, zero nodes
		}
		clusterInst.Liveness = edgeproto.Liveness_LIVENESS_DYNAMIC
		err := clusterInstApi.createClusterInstInternal(cctx, &clusterInst, cb)
		if err != nil {
			return err
		}
		defer func() {
			if reterr != nil && !cctx.Undo {
				cb.Send(&edgeproto.Result{Message: "Deleting auto-ClusterInst due to failure"})
				undoErr := clusterInstApi.deleteClusterInstInternal(cctx.WithUndo(), &clusterInst, cb)
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
		var cloudlet edgeproto.Cloudlet
		if !cloudletApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudlet) {
			return errors.New("Specified Cloudlet not found")
		}
		in.CloudletLoc = cloudlet.Location

		var app edgeproto.App
		if !appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return in.Key.AppKey.NotFoundError()
		}

		if in.Flavor.Name == "" {
			in.Flavor = app.DefaultFlavor
		}
		var clusterKey *edgeproto.ClusterKey
		ipaccess := edgeproto.IpAccess_IP_ACCESS_SHARED
		if cloudcommon.IsClusterInstReqd(&app) {
			clusterInst := edgeproto.ClusterInst{}
			if !clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, &clusterInst) {
				return errors.New("ClusterInst does not exist for App")
			}
			if clusterInst.State != edgeproto.TrackedState_READY {
				return fmt.Errorf("ClusterInst %s not ready", clusterInst.Key.GetKeyString())
			}
			if tenant {
				if !clusterInst.Reservable {
					return fmt.Errorf("ClusterInst is not reservable")
				}
				if clusterInst.ReservedBy != "" {
					// Only one AppInst (even from the same developer)
					// allowed per reservable ClusterInst
					return fmt.Errorf("Reservable MobiledgeX ClusterInst already in use")
				}
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
			if tenant {
				clusterInst.ReservedBy = in.Key.AppKey.Organization
				clusterInstApi.store.STMPut(stm, &clusterInst)
			}
		}

		cloudletRefs := edgeproto.CloudletRefs{}
		cloudletRefsChanged := false
		if !cloudletRefsApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudletRefs) {
			initCloudletRefs(&cloudletRefs, &in.Key.ClusterInstKey.CloudletKey)
		}

		ports, _ := edgeproto.ParseAppPorts(app.AccessPorts)
		if !cloudcommon.IsClusterInstReqd(&app) {
			in.Uri = cloudcommon.GetVMAppFQDN(&in.Key, &in.Key.ClusterInstKey.CloudletKey, *appDNSRoot)
			for ii, _ := range ports {
				ports[ii].PublicPort = ports[ii].InternalPort
			}
		} else if ipaccess == edgeproto.IpAccess_IP_ACCESS_SHARED && !app.InternalPorts {
			in.Uri = cloudcommon.GetRootLBFQDN(&in.Key.ClusterInstKey.CloudletKey, *appDNSRoot)
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
			if isIPAllocatedPerService(cloudlet.PlatformType, clusterInst.Key.CloudletKey.Organization) {
				//dedicated access in which each service gets a different ip
				in.Uri = cloudcommon.GetAppFQDN(&in.Key, &in.Key.ClusterInstKey.CloudletKey, clusterKey, *appDNSRoot)
				for ii, _ := range ports {
					ports[ii].PublicPort = ports[ii].InternalPort
				}
			} else {
				//dedicated access in which IP is that of the LB
				in.Uri = cloudcommon.GetDedicatedLBFQDN(&in.Key.ClusterInstKey.CloudletKey, clusterKey, *appDNSRoot)
				for ii, _ := range ports {
					ports[ii].PublicPort = ports[ii].InternalPort
				}
				// port range is validated on app create, but checked again here in case there were
				// pre-existing apps which violate the supported range
				err = validatePortRangeForAccessType(ports, app.AccessType, app.Deployment)
				if err != nil {
					return err
				}
			}
		}
		if len(ports) > 0 {
			in.MappedPorts = ports
			setPortFQDNPrefixes(in, &app)
		}

		// TODO: Make sure resources are available
		if cloudletRefsChanged {
			cloudletRefsApi.store.STMPut(stm, &cloudletRefs)
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
	err = appInstApi.cache.WaitForState(ctx, &in.Key, edgeproto.TrackedState_READY, CreateAppInstTransitions, edgeproto.TrackedState_CREATE_ERROR, settingsApi.Get().CreateAppInstTimeout.TimeDuration(), "Created AppInst successfully", cb.Send)
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
	return err
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

	s.setDefaultVMClusterKey(ctx, &key)
	if err := key.ValidateKey(); err != nil {
		return false, err
	}

	// create stream once AppInstKey is formed correctly
	sendObj, cb, err := startAppInstStream(ctx, &key, inCb)
	if err == nil {
		defer func() {
			stopAppInstStream(ctx, &key, sendObj, reterr)
		}()
	}

	var app edgeproto.App

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		var curr edgeproto.AppInst

		if !appApi.store.STMGet(stm, &key.AppKey, &app) {
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
			cloudletErr := checkCloudletReady(cctx, stm, &key.ClusterInstKey.CloudletKey)
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
		RecordAppInstEvent(ctx, &key, cloudcommon.UPDATE_START, cloudcommon.InstanceDown)

		defer func() {
			if reterr == nil {
				RecordAppInstEvent(ctx, &key, cloudcommon.UPDATE_COMPLETE, cloudcommon.InstanceUp)
			} else {
				RecordAppInstEvent(ctx, &key, cloudcommon.UPDATE_ERROR, cloudcommon.InstanceDown)
			}
		}()
		err = appInstApi.cache.WaitForState(cb.Context(), &key, edgeproto.TrackedState_READY, UpdateAppInstTransitions, edgeproto.TrackedState_UPDATE_ERROR, settingsApi.Get().UpdateAppInstTimeout.TimeDuration(), "", cb.Send)
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
		s.setDefaultVMClusterKey(ctx, &in.Key)
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
		log.DebugLog(log.DebugLevelApi, "no AppInsts matched", "key", in.Key)
		return in.Key.NotFoundError()
	}

	if !singleAppInst {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Updating: %d AppInsts", len(instances))})
	}

	for instkey, _ := range instances {
		go func(k edgeproto.AppInstKey) {
			log.DebugLog(log.DebugLevelApi, "updating AppInst", "key", k)
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
		log.DebugLog(log.DebugLevelApi, "instanceUpdateResult ", "key", k, "updated", result.revisionUpdated, "error", result.errString)
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
			if !appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
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

	s.setDefaultVMClusterKey(ctx, &in.Key)
	if err := in.Key.AppKey.ValidateKey(); err != nil {
		return err
	}

	appInstKey := in.Key
	// create stream once AppInstKey is formed correctly
	sendObj, cb, err := startAppInstStream(ctx, &appInstKey, inCb)
	if err == nil {
		defer func() {
			stopAppInstStream(ctx, &appInstKey, sendObj, reterr)
		}()
	}

	// get appinst info for flavor
	appInstInfo := edgeproto.AppInst{}
	if !appInstApi.cache.Get(&in.Key, &appInstInfo) {
		return in.Key.NotFoundError()
	}
	eventCtx := context.WithValue(ctx, in.Key, appInstInfo)
	defer func() {
		if reterr == nil {
			RecordAppInstEvent(eventCtx, &in.Key, cloudcommon.DELETED, cloudcommon.InstanceDown)
		}
	}()

	log.DebugLog(log.DebugLevelApi, "deleteAppInstInternal", "AppInst", in)
	// populate the clusterinst developer from the app developer if not already present
	if in.Key.ClusterInstKey.Organization == "" {
		in.Key.ClusterInstKey.Organization = in.Key.AppKey.Organization
		cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
	}

	// check if we are deleting an autocluster instance we need to set the key correctly.
	if strings.HasPrefix(in.Key.ClusterInstKey.ClusterKey.Name, ClusterAutoPrefix) && in.Key.ClusterInstKey.Organization == "" {
		in.Key.ClusterInstKey.Organization = in.Key.AppKey.Organization
	}
	clusterInstKey := edgeproto.ClusterInstKey{}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			if in.Key.ClusterInstKey.Organization == "" {
				// create still allows it be unset on input,
				// but will set it internally. But it is required
				// on delete just in case another ClusterInst
				// exists without it set. So be nice and
				// remind them they may have forgotten to
				// specify it.
				cb.Send(&edgeproto.Result{Message: "ClusterInstKey developer not specified, may need to specify it"})
			}
			// already deleted
			return in.Key.NotFoundError()
		}
		if err := in.Key.ClusterInstKey.ValidateKey(); err != nil {
			return err
		}
		if !cctx.Undo && in.State != edgeproto.TrackedState_READY && in.State != edgeproto.TrackedState_CREATE_ERROR && in.State != edgeproto.TrackedState_DELETE_ERROR &&
			in.State != edgeproto.TrackedState_DELETE_DONE && in.State != edgeproto.TrackedState_UPDATE_ERROR && !ignoreTransient(cctx, in.State) {
			log.SpanLog(ctx, log.DebugLevelApi, "AppInst busy, cannot delete", "state", in.State)
			return errors.New("AppInst busy, cannot delete")
		}
		if err := checkCloudletReady(cctx, stm, &in.Key.ClusterInstKey.CloudletKey); err != nil {
			return err
		}

		var cloudlet edgeproto.Cloudlet
		if !cloudletApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudlet) {
			return fmt.Errorf("For AppInst, %v", in.Key.ClusterInstKey.CloudletKey.NotFoundError())
		}
		app = edgeproto.App{}
		if !appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return fmt.Errorf("For AppInst, %v", in.Key.AppKey.NotFoundError())
		}
		clusterInstReqd := cloudcommon.IsClusterInstReqd(&app)
		clusterInst := edgeproto.ClusterInst{}
		if clusterInstReqd && !clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, &clusterInst) {
			return fmt.Errorf("For AppInst, %v", in.Key.ClusterInstKey.NotFoundError())
		}
		if err := autoProvPolicyApi.appInstCheck(ctx, stm, cloudcommon.Delete, &app, in); err != nil {
			return err
		}

		cloudletRefs := edgeproto.CloudletRefs{}
		cloudletRefsChanged := false
		hasRefs := cloudletRefsApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudletRefs)
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
		clusterInstKey = in.Key.ClusterInstKey
		if cloudletRefsChanged {
			cloudletRefsApi.store.STMPut(stm, &cloudletRefs)
		}
		if clusterInstReqd && clusterInst.ReservedBy != "" && clusterInst.ReservedBy == in.Key.AppKey.Organization {
			clusterInst.ReservedBy = ""
			clusterInstApi.store.STMPut(stm, &clusterInst)
		}

		// delete app inst
		if ignoreCRM(cctx) {
			// CRM state should be the same as before the
			// operation failed, so just need to clean up
			// controller state.
			s.store.STMDel(stm, &in.Key)
			appInstRefsApi.removeRef(stm, &in.Key)
		} else {
			in.State = edgeproto.TrackedState_DELETE_REQUESTED
			in.Status = edgeproto.StatusInfo{}
			s.store.STMPut(stm, in)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if ignoreCRM(cctx) {
		cb.Send(&edgeproto.Result{Message: "Deleted AppInst successfully"})
	} else {
		err = appInstApi.cache.WaitForState(ctx, &in.Key, edgeproto.TrackedState_NOT_PRESENT, DeleteAppInstTransitions, edgeproto.TrackedState_DELETE_ERROR, settingsApi.Get().DeleteAppInstTimeout.TimeDuration(), "Deleted AppInst successfully", cb.Send)
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
	}
	// delete clusterinst afterwards if it was auto-created and nobody is left using it
	clusterInst := edgeproto.ClusterInst{}
	if clusterInstApi.Get(&clusterInstKey, &clusterInst) && clusterInst.Auto && !appInstApi.UsesClusterInst(&clusterInstKey) {
		cb.Send(&edgeproto.Result{Message: "Deleting auto-ClusterInst"})
		autoerr := clusterInstApi.deleteClusterInstInternal(cctx, &clusterInst, cb)
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

func (s *AppInstApi) HealthCheckUpdate(ctx context.Context, in *edgeproto.AppInst, state edgeproto.HealthCheck) {
	log.DebugLog(log.DebugLevelApi, "Update AppInst Health Check", "key", in.Key, "state", state)
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
		if inst.HealthCheck == edgeproto.HealthCheck_HEALTH_CHECK_OK && state != edgeproto.HealthCheck_HEALTH_CHECK_OK {
			// healthy -> not healthy
			RecordAppInstEvent(ctx, &inst.Key, cloudcommon.HEALTH_CHECK_FAIL, cloudcommon.InstanceDown)
			nodeMgr.Event(ctx, "AppInst offline", in.Key.AppKey.Organization, in.Key.GetTags(), nil, "state", state.String())
		} else if inst.HealthCheck != edgeproto.HealthCheck_HEALTH_CHECK_OK && state == edgeproto.HealthCheck_HEALTH_CHECK_OK {
			// not healthy -> healthy
			RecordAppInstEvent(ctx, &inst.Key, cloudcommon.HEALTH_CHECK_OK, cloudcommon.InstanceUp)
			nodeMgr.Event(ctx, "AppInst online", in.Key.AppKey.Organization, in.Key.GetTags(), nil, "state", state.String())
		}
		inst.HealthCheck = state
		s.store.STMPut(stm, &inst)
		return nil
	})
}

func (s *AppInstApi) UpdateFromInfo(ctx context.Context, in *edgeproto.AppInstInfo) {
	log.DebugLog(log.DebugLevelApi, "Update AppInst from info", "key", in.Key, "state", in.State, "status", in.Status, "powerstate", in.PowerState)
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		if in.PowerState != edgeproto.PowerState_POWER_STATE_UNKNOWN {
			inst.PowerState = in.PowerState
		}
		// If AppInst is ready and state has not been set yet by HealthCheckUpdate, default to Ok.
		if in.State == edgeproto.TrackedState_READY &&
			inst.HealthCheck == edgeproto.HealthCheck_HEALTH_CHECK_UNKNOWN {
			inst.HealthCheck = edgeproto.HealthCheck_HEALTH_CHECK_OK
		}
		// update only diff of status msgs
		edgeproto.UpdateStatusDiff(&in.Status, &inst.Status)
		if inst.State == in.State {
			// already in that state
			if in.State == edgeproto.TrackedState_READY {
				// update runtime info
				inst.RuntimeInfo = in.RuntimeInfo
			}
			s.store.STMPut(stm, &inst)
			return nil
		}
		// please see state_transitions.md
		if !crmTransitionOk(inst.State, in.State) {
			log.DebugLog(log.DebugLevelApi, "Invalid state transition",
				"key", &in.Key, "cur", inst.State, "next", in.State)
			return nil
		}
		inst.State = in.State
		if in.State == edgeproto.TrackedState_CREATE_ERROR || in.State == edgeproto.TrackedState_DELETE_ERROR || in.State == edgeproto.TrackedState_UPDATE_ERROR {
			inst.Errors = in.Errors
		} else {
			inst.Errors = nil
		}
		inst.RuntimeInfo = in.RuntimeInfo
		s.store.STMPut(stm, &inst)
		return nil
	})
	if in.State == edgeproto.TrackedState_DELETE_DONE {
		s.DeleteFromInfo(ctx, in)
	}
}

func (s *AppInstApi) DeleteFromInfo(ctx context.Context, in *edgeproto.AppInstInfo) {
	log.DebugLog(log.DebugLevelApi, "Delete AppInst from info", "key", in.Key, "state", in.State)
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		// please see state_transitions.md
		if inst.State != edgeproto.TrackedState_DELETING && inst.State != edgeproto.TrackedState_DELETE_REQUESTED &&
			inst.State != edgeproto.TrackedState_DELETE_DONE {
			log.DebugLog(log.DebugLevelApi, "Invalid state transition",
				"key", &in.Key, "cur", inst.State,
				"next", edgeproto.TrackedState_DELETE_DONE)
			return nil
		}
		s.store.STMDel(stm, &in.Key)
		appInstRefsApi.removeRef(stm, &in.Key)
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
			appInstRefsApi.removeRef(stm, &in.Key)
		} else {
			inst.State = newState
			inst.Errors = nil
			s.store.STMPut(stm, &inst)
		}
		return nil
	})
}

// public cloud k8s cluster allocates a separate IP per service.  This is a type of dedicated access
func isIPAllocatedPerService(platformType edgeproto.PlatformType, operator string) bool {
	log.DebugLog(log.DebugLevelApi, "isIPAllocatedPerService", "platformType", platformType, "operator", operator)

	if platformType == edgeproto.PlatformType_PLATFORM_TYPE_FAKE {
		// for a fake cloudlet used in testing, decide based on operator name
		return operator == cloudcommon.OperatorGCP || operator == cloudcommon.OperatorAzure || operator == cloudcommon.OperatorAWS
	}
	return platformType == edgeproto.PlatformType_PLATFORM_TYPE_AWS_EKS ||
		platformType == edgeproto.PlatformType_PLATFORM_TYPE_AZURE ||
		platformType == edgeproto.PlatformType_PLATFORM_TYPE_GCP
}

func allocateIP(inst *edgeproto.ClusterInst, cloudlet *edgeproto.Cloudlet, platformType edgeproto.PlatformType, refs *edgeproto.CloudletRefs) error {

	if isIPAllocatedPerService(platformType, cloudlet.Key.Organization) {
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

func RecordAppInstEvent(ctx context.Context, appInstKey *edgeproto.AppInstKey, event cloudcommon.InstanceEvent, serverStatus string) {
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
	if !appApi.cache.Get(&appInstKey.AppKey, &app) {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Cannot find appdata for app", "app", appInstKey.AppKey)
		return
	}
	metric.AddStringVal("deployment", app.Deployment)

	// have to grab the appinst here because its now possible to create apps without a default flavor
	// on deletes, the appinst is passed into the context otherwise we wont be able to get it
	appInst, ok := ctx.Value(*appInstKey).(edgeproto.AppInst)
	if !ok {
		if !appInstApi.cache.Get(appInstKey, &appInst) {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Cannot find appdata for app", "app", appInstKey.AppKey)
			return
		}
	}
	if app.Deployment == cloudcommon.DeploymentTypeVM {
		metric.AddStringVal("flavor", appInst.Flavor.Name)
	}

	services.events.AddMetric(&metric)

	// check to see if it was autoprovisioned and they used a reserved clusterinst, then log the start and stop of the clusterinst as well
	if isTenantAppInst(appInstKey) && (event == cloudcommon.CREATED || event == cloudcommon.DELETED) {
		clusterEvent := cloudcommon.RESERVED
		if event == cloudcommon.DELETED {
			clusterEvent = cloudcommon.UNRESERVED
		}
		RecordClusterInstEvent(ctx, &appInstKey.ClusterInstKey, clusterEvent, serverStatus)
	}
}

func isTenantAppInst(appInstKey *edgeproto.AppInstKey) bool {
	return appInstKey.ClusterInstKey.Organization == cloudcommon.OrganizationMobiledgeX && appInstKey.AppKey.Organization != cloudcommon.OrganizationMobiledgeX
}
