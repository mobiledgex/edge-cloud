package main

import (
	"context"
	"errors"
	"fmt"
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
	for key, val := range s.cache.Objs {
		if key.ClusterInstKey.CloudletKey.Matches(in) && appApi.Get(&val.Key.AppKey, &app) {
			if (val.Liveness == edgeproto.Liveness_LIVENESS_STATIC) && (app.DelOpt == edgeproto.DeleteType_NO_AUTO_DELETE) {
				static = true
				//if can autodelete it then also add it to the dynInsts to be deleted later
			} else if (val.Liveness == edgeproto.Liveness_LIVENESS_DYNAMIC) || (app.DelOpt == edgeproto.DeleteType_AUTO_DELETE) {
				dynInsts[key] = struct{}{}
			}
		}
	}
	return static
}

// Checks if there is some action in progress by AppInst on the cloudlet
func (s *AppInstApi) UsingCloudlet(in *edgeproto.CloudletKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, val := range s.cache.Objs {
		if key.ClusterInstKey.CloudletKey.Matches(in) {
			if edgeproto.IsTransientState(val.State) {
				return true
			}
		}
	}
	return false
}

func (s *AppInstApi) updateAppInstRevision(ctx context.Context, key *edgeproto.AppInstKey, revision int32) error {
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
	for key, val := range s.cache.Objs {
		if key.AppKey.Matches(in) {
			if val.Liveness == edgeproto.Liveness_LIVENESS_STATIC {
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
	for key, val := range s.cache.Objs {
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

func (s *AppInstApi) AutoDeleteAppInsts(key *edgeproto.ClusterInstKey, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	var app edgeproto.App
	var err error
	apps := make(map[edgeproto.AppInstKey]*edgeproto.AppInst)
	log.DebugLog(log.DebugLevelApi, "Auto-deleting AppInsts ", "cluster", key.ClusterKey.Name)
	s.cache.Mux.Lock()
	for k, val := range s.cache.Objs {
		if k.ClusterInstKey.Matches(key) && appApi.Get(&val.Key.AppKey, &app) {
			if app.DelOpt == edgeproto.DeleteType_AUTO_DELETE || val.Liveness == edgeproto.Liveness_LIVENESS_DYNAMIC {
				apps[k] = val
			}
		}
	}
	s.cache.Mux.Unlock()

	//Spin in case cluster was just created and apps are still in the creation process and cannot be deleted
	var spinTime time.Duration
	start := time.Now()
	for _, val := range apps {
		log.DebugLog(log.DebugLevelApi, "Auto-deleting AppInst ", "appinst", val.Key.AppKey.Name)
		cb.Send(&edgeproto.Result{Message: "Autodeleting AppInst " + val.Key.AppKey.Name})
		for {
			// ignore CRM errors when deleting dynamic apps as we will be deleting the cluster anyway
			crmo := edgeproto.CRMOverride_IGNORE_CRM_ERRORS
			cctx := DefCallContext()
			cctx.SetOverride(&crmo)
			err = s.deleteAppInstInternal(cctx, val, cb)
			if err == nil {
				RecordAppInstEvent(cb.Context(), val, cloudcommon.DELETED, cloudcommon.InstanceDown)
			}
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

func (s *AppInstApi) UsesFlavor(key *edgeproto.FlavorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, app := range s.cache.Objs {
		if app.Flavor.Matches(key) {
			return true
		}
	}
	return false
}

func (s *AppInstApi) CreateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_CreateAppInstServer) error {
	in.Liveness = edgeproto.Liveness_LIVENESS_STATIC
	err := s.createAppInstInternal(DefCallContext(), in, cb)
	if err == nil {
		RecordAppInstEvent(cb.Context(), in, cloudcommon.CREATED, cloudcommon.InstanceUp)
	}
	return err
}

func getProtocolBitMap(proto dme.LProto) (int32, error) {
	var bitmap int32
	switch proto {
	//put all "TCP" protocols below here
	case dme.LProto_L_PROTO_HTTP:
		fallthrough
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
	if app.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW {
		key.ClusterInstKey.ClusterKey.Name = cloudcommon.DefaultVMCluster
	}
}

// createAppInstInternal is used to create dynamic app insts internally,
// bypassing static assignment.
func (s *AppInstApi) createAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, cb edgeproto.AppInstApi_CreateAppInstServer) (reterr error) {
	ctx := cb.Context()

	// populate the clusterinst developer from the app developer if not already present
	if in.Key.ClusterInstKey.Developer == "" {
		in.Key.ClusterInstKey.Developer = in.Key.AppKey.DeveloperKey.Name
		cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
	}
	s.setDefaultVMClusterKey(ctx, &in.Key)
	if err := in.Key.ValidateKey(); err != nil {
		return err
	}

	if in.Liveness == edgeproto.Liveness_LIVENESS_UNKNOWN {
		in.Liveness = edgeproto.Liveness_LIVENESS_DYNAMIC
	}
	cctx.SetOverride(&in.CrmOverride)
	if !ignoreCRM(cctx) {
		if err := cloudletInfoApi.checkCloudletReady(&in.Key.ClusterInstKey.CloudletKey); err != nil {
			return err
		}
	}

	var autocluster bool
	var tenant bool
	// See if we need to create auto-cluster.
	// This also sets up the correct ClusterInstKey in "in".

	if in.Key.AppKey.DeveloperKey.Name != in.Key.ClusterInstKey.Developer {
		// both are specified, make sure they match
		if in.Key.AppKey.DeveloperKey.Name == cloudcommon.DeveloperMobiledgeX {
			// mobiledgex apps on dev ClusterInst, like prometheus
		} else if in.Key.ClusterInstKey.Developer == cloudcommon.DeveloperMobiledgeX {
			// developer apps on reservable mobiledgex ClusterInst
			tenant = true
		} else {
			return fmt.Errorf("Developer name mismatch between App: %s and ClusterInst: %s", in.Key.AppKey.DeveloperKey.Name, in.Key.ClusterInstKey.Developer)
		}
	}
	appDeploymentType := ""

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
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
		if !cloudletApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, nil) {
			return errors.New("Specified Cloudlet not found")
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
		in.Revision = app.Revision
		appDeploymentType = app.Deployment
		// Check if specified ClusterInst exists
		if !strings.HasPrefix(cikey.ClusterKey.Name, ClusterAutoPrefix) && cloudcommon.IsClusterInstReqd(&app) {
			found := clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, nil)
			if !found && in.Key.ClusterInstKey.Developer == "" {
				// developer may not be specified
				// in clusterinst.
				in.Key.ClusterInstKey.Developer = in.Key.AppKey.DeveloperKey.Name
				found = clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, nil)
				if found {
					cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
				}
			}
			if !found {
				return errors.New("Specified ClusterInst not found")
			}
			// cluster inst exists so we're good.
			return nil
		}
		if cloudcommon.IsClusterInstReqd(&app) {
			// Auto-cluster
			if clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, nil) {
				// if it already exists, this means we just want to spawn more apps into it
				return nil
			}
			// If this is an autodelete app, we should only allow those in existing cluster instances
			if app.DelOpt == edgeproto.DeleteType_AUTO_DELETE {
				return fmt.Errorf("Autodelete App %s requires an existing ClusterInst", app.Key.Name)
			}
			cikey.ClusterKey.Name = util.K8SSanitize(cikey.ClusterKey.Name)
			if cikey.Developer == "" {
				cikey.Developer = in.Key.AppKey.DeveloperKey.Name
			}
			autocluster = true
		}

		if in.Flavor.Name == "" {
			// find flavor from app
			in.Flavor = app.DefaultFlavor
		}
		if in.SharedVolumeSize == 0 {
			in.SharedVolumeSize = app.DefaultSharedVolumeSize
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, in) {
			if !cctx.Undo && in.State != edgeproto.TrackedState_DELETE_ERROR && !ignoreTransient(cctx, in.State) {
				if in.State == edgeproto.TrackedState_CREATE_ERROR {
					cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
					cb.Send(&edgeproto.Result{Message: "Use DeleteAppInst to remove and try again"})
				}
				return in.Key.ExistsError()
			}
			in.Errors = nil
		} else {
			err := in.Validate(edgeproto.AppInstAllFieldsMap)
			if err != nil {
				return err
			}
		}
		// Set new state to show autocluster clusterinst progress as part of
		// appinst progress
		in.State = edgeproto.TrackedState_CREATING_DEPENDENCIES
		s.store.STMPut(stm, in)
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
					}
				}
				return nil
			})
		}
	}()

	if autocluster {
		// auto-create cluster inst
		clusterInst := edgeproto.ClusterInst{}
		clusterInst.Key = in.Key.ClusterInstKey
		clusterInst.Auto = true
		log.DebugLog(log.DebugLevelApi,
			"Create auto-ClusterInst",
			"key", clusterInst.Key,
			"AppInst", in)

		clusterInst.Flavor = in.Flavor
		clusterInst.IpAccess = in.AutoClusterIpAccess
		clusterInst.Deployment = appDeploymentType
		clusterInst.SharedVolumeSize = in.SharedVolumeSize
		if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes ||
			appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
			clusterInst.Deployment = cloudcommon.AppDeploymentTypeKubernetes
			clusterInst.NumMasters = 1
			clusterInst.NumNodes = 1 // TODO support 1 master, zero nodes
		}
		err := clusterInstApi.createClusterInstInternal(cctx, &clusterInst, cb)
		if err != nil {
			return err
		} else if clusterInst.State == edgeproto.TrackedState_READY {
			RecordClusterInstEvent(ctx, &clusterInst, cloudcommon.CREATED, cloudcommon.InstanceUp)
		}
		defer func() {
			if reterr != nil && !cctx.Undo {
				cb.Send(&edgeproto.Result{Message: "Deleting auto-ClusterInst due to failure"})
				undoErr := clusterInstApi.deleteClusterInstInternal(cctx.WithUndo(), &clusterInst, cb)
				if undoErr != nil {
					log.DebugLog(log.DebugLevelApi,
						"Undo create auto-ClusterInst failed",
						"key", clusterInst.Key,
						"undoErr", undoErr)
				} else {
					RecordClusterInstEvent(ctx, &clusterInst, cloudcommon.DELETED, cloudcommon.InstanceDown)
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

		if !flavorApi.store.STMGet(stm, &in.Flavor, nil) {
			return fmt.Errorf("Flavor %s not found", in.Flavor.Name)
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
			if app.Deployment == cloudcommon.AppDeploymentTypeHelm {
				needDeployment = cloudcommon.AppDeploymentTypeKubernetes
			}
			if clusterInst.Deployment != needDeployment {
				return fmt.Errorf("Cannot deploy %s App into %s ClusterInst", app.Deployment, clusterInst.Deployment)
			}
			ipaccess = clusterInst.IpAccess
			clusterKey = &clusterInst.Key.ClusterKey
			if tenant {
				clusterInst.ReservedBy = in.Key.AppKey.DeveloperKey.Name
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
			in.Uri = cloudcommon.GetVMAppFQDN(&in.Key, &in.Key.ClusterInstKey.CloudletKey)
			for ii, _ := range ports {
				ports[ii].PublicPort = ports[ii].InternalPort
			}
		} else if ipaccess == edgeproto.IpAccess_IP_ACCESS_SHARED && !app.InternalPorts {
			in.Uri = cloudcommon.GetRootLBFQDN(&in.Key.ClusterInstKey.CloudletKey)
			if cloudletRefs.RootLbPorts == nil {
				cloudletRefs.RootLbPorts = make(map[int32]int32)
			}

			for ii, port := range ports {
				if port.EndPort != 0 && ipaccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
					return fmt.Errorf("Shared IP access with port range not allowed")
				}
				if setL7Port(&ports[ii], &in.Key) {
					log.DebugLog(log.DebugLevelApi,
						"skip L7 port", "port", ports[ii])
					continue
				}
				// Samsung enabling layer ignores port mapping.
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
					if iport != 22 && iport != cloudcommon.RootLBL7Port && iport != cloudcommon.ProxyMetricsPort {
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
			if isIPAllocatedPerService(in.Key.ClusterInstKey.CloudletKey.OperatorKey.Name) {
				//dedicated access in which each service gets a different ip
				in.Uri = cloudcommon.GetAppFQDN(&in.Key, &in.Key.ClusterInstKey.CloudletKey, clusterKey)
				for ii, _ := range ports {
					// No rootLB to do L7 muxing, and each
					// service has it's own IP anyway so
					// no muxing is needed. Treat http as tcp.
					if ports[ii].Proto == dme.LProto_L_PROTO_HTTP {
						ports[ii].Proto = dme.LProto_L_PROTO_TCP
					}
					ports[ii].PublicPort = ports[ii].InternalPort
				}
			} else {
				//dedicated access in which IP is that of the LB
				in.Uri = cloudcommon.GetDedicatedLBFQDN(&in.Key.ClusterInstKey.CloudletKey, clusterKey)
				// Docker deployments do not need an L7 reverse
				// proxy (because they all bind to ports on the
				// same VM, so multiple containers trying to bind
				// to port 80 would fail, and if they're binding
				// to different ports, then they don't need the
				// L7 proxy).
				// Kubernetes deployments may want it, but so far
				// no devs are using L7, and if they did, they
				// probably would have an ingress controller
				// built into their k8s manifest. So for now,
				// do not support http on k8s either. If we see
				// demand/use cases for it we can add it in later.
				for ii, _ := range ports {
					if ports[ii].Proto == dme.LProto_L_PROTO_HTTP {
						ports[ii].Proto = dme.LProto_L_PROTO_TCP
					}
					ports[ii].PublicPort = ports[ii].InternalPort
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
func (s *AppInstApi) refreshAppInstInternal(cctx *CallContext, key edgeproto.AppInstKey, cb edgeproto.AppInstApi_RefreshAppInstServer, forceUpdate bool) (bool, error) {
	ctx := cb.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "refreshAppInstInternal", "key", key)

	updatedRevision := false
	crmUpdateRequired := false

	s.setDefaultVMClusterKey(ctx, &key)
	if err := key.ValidateKey(); err != nil {
		return false, err
	}

	cloudletErr := cloudletInfoApi.checkCloudletReady(&key.ClusterInstKey.CloudletKey)

	var app edgeproto.App

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
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
			if curr.Revision != app.Revision {
				crmUpdateRequired = true
				updatedRevision = true
			} else if forceUpdate {
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
			if crmUpdateRequired && cloudletErr != nil {
				return cloudletErr
			}
			curr.State = edgeproto.TrackedState_UPDATE_REQUESTED
		}
		s.store.STMPut(stm, &curr)
		return nil
	})

	if err != nil {
		return false, err
	}
	if crmUpdateRequired {
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
		// if UpdateMuliple flag is specified, then only the appkey must be present
		if err := in.Key.AppKey.ValidateKey(); err != nil {
			return err
		}
	} else {
		// populate the clusterinst developer from the app developer if not already present
		if in.Key.ClusterInstKey.Developer == "" {
			in.Key.ClusterInstKey.Developer = in.Key.AppKey.DeveloperKey.Name
			cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
		}

		// the whole key must be present
		s.setDefaultVMClusterKey(ctx, &in.Key)
		if err := in.Key.ValidateKey(); err != nil {
			return fmt.Errorf("cluster key needed without updatemultiple option: %v", err)
		}
	}

	s.cache.Mux.Lock()

	type updateResult struct {
		errString       string
		revisionUpdated bool
	}
	instanceUpdateResults := make(map[edgeproto.AppInstKey]chan updateResult)
	instances := make(map[edgeproto.AppInstKey]struct{})

	for k, val := range s.cache.Objs {
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

	cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Updating: %d AppInsts", len(instances))})

	for instkey, _ := range instances {
		go func(k edgeproto.AppInstKey) {
			log.DebugLog(log.DebugLevelApi, "updating AppInst", "key", k)
			RecordAppInstEvent(cb.Context(), in, cloudcommon.UPDATE_START, cloudcommon.InstanceDown)
			updated, err := s.refreshAppInstInternal(DefCallContext(), k, cb, in.ForceUpdate)
			if err == nil {
				instanceUpdateResults[k] <- updateResult{errString: "", revisionUpdated: updated}
				RecordAppInstEvent(cb.Context(), in, cloudcommon.UPDATE_COMPLETE, cloudcommon.InstanceUp)
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
			} else {
				numSkipped++
			}
		} else {
			numFailed++
		}
		// give some intermediate status
		if (numTotal%10 == 0) && numTotal != len(instances) {
			cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Processing: %d of %d AppInsts.  Updated: %d Skipped: %d Failed: %d", numTotal, len(instances), numUpdated, numSkipped, numFailed)})
		}
	}
	cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Completed: %d of %d AppInsts.  Updated: %d Skipped: %d Failed: %d", numTotal, len(instances), numUpdated, numSkipped, numFailed)})
	return nil
}

func (s *AppInstApi) UpdateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_UpdateAppInstServer) error {
	ctx := cb.Context()
	fmap := edgeproto.MakeFieldMap(in.Fields)
	err := in.Validate(fmap)
	if err != nil {
		return err
	}

	allowedFields := []string{}
	badFields := []string{}
	for _, field := range in.Fields {
		if field == edgeproto.AppInstFieldCrmOverride ||
			field == edgeproto.AppInstFieldKey ||
			in.IsKeyField(field) {
			continue
		} else if field == edgeproto.AppInstFieldConfigs || field == edgeproto.AppInstFieldConfigsKind || field == edgeproto.AppInstFieldConfigsConfig {
			// only thing modifiable right now is the "configs".
			allowedFields = append(allowedFields, field)
		} else {
			badFields = append(badFields, field)
		}
	}
	if len(badFields) > 0 {
		// cat all the bad field names and return error
		badstrs := []string{}
		for _, bad := range badFields {
			badstrs = append(badstrs, edgeproto.AppInstAllFieldsStringMap[bad])
		}
		return fmt.Errorf("specified fields %s cannot be modified", strings.Join(badstrs, ","))
	}
	in.Fields = allowedFields
	if len(allowedFields) == 0 {
		return fmt.Errorf("Nothing specified to modify")
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
	RecordAppInstEvent(cb.Context(), in, cloudcommon.UPDATE_START, cloudcommon.InstanceDown)
	_, err = s.refreshAppInstInternal(cctx, in.Key, cb, forceUpdate)
	if err != nil {
		RecordAppInstEvent(cb.Context(), in, cloudcommon.UPDATE_COMPLETE, cloudcommon.InstanceUp)
	}
	return err
}

func (s *AppInstApi) DeleteAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	err := s.deleteAppInstInternal(DefCallContext(), in, cb)
	if err == nil {
		RecordAppInstEvent(cb.Context(), in, cloudcommon.DELETED, cloudcommon.InstanceDown)
	}
	return err
}

func (s *AppInstApi) deleteAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	cctx.SetOverride(&in.CrmOverride)
	ctx := cb.Context()

	log.DebugLog(log.DebugLevelApi, "deleteAppInstInternal", "AppInst", in)
	// populate the clusterinst developer from the app developer if not already present
	if in.Key.ClusterInstKey.Developer == "" {
		in.Key.ClusterInstKey.Developer = in.Key.AppKey.DeveloperKey.Name
		cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
	}
	s.setDefaultVMClusterKey(ctx, &in.Key)
	if err := in.Key.AppKey.ValidateKey(); err != nil {
		return err
	}

	if !ignoreCRM(cctx) {
		if err := cloudletInfoApi.checkCloudletReady(&in.Key.ClusterInstKey.CloudletKey); err != nil {
			return err
		}
	}

	// check if we are deleting an autocluster instance we need to set the key correctly.
	if strings.HasPrefix(in.Key.ClusterInstKey.ClusterKey.Name, ClusterAutoPrefix) && in.Key.ClusterInstKey.Developer == "" {
		in.Key.ClusterInstKey.Developer = in.Key.AppKey.DeveloperKey.Name
	}
	clusterInstKey := edgeproto.ClusterInstKey{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			if in.Key.ClusterInstKey.Developer == "" {
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

		if !cctx.Undo && in.State != edgeproto.TrackedState_READY && in.State != edgeproto.TrackedState_CREATE_ERROR && in.State != edgeproto.TrackedState_DELETE_ERROR && !ignoreTransient(cctx, in.State) {
			log.SpanLog(ctx, log.DebugLevelApi, "AppInst busy, cannot delete", "state", in.State)
			return errors.New("AppInst busy, cannot delete")
		}

		var cloudlet edgeproto.Cloudlet

		if !cloudletApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudlet) {
			return errors.New("Specified Cloudlet not found")
		}

		cloudletRefs := edgeproto.CloudletRefs{}
		cloudletRefsChanged := false
		if cloudletRefsApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudletRefs) {
			// shared root load balancer
			for ii, _ := range in.MappedPorts {
				if in.MappedPorts[ii].Proto == dme.LProto_L_PROTO_HTTP {
					continue
				}

				p := in.MappedPorts[ii].PublicPort
				protocol, err := getProtocolBitMap(in.MappedPorts[ii].Proto)

				if err != nil {
					return err
				}
				protos, _ := cloudletRefs.RootLbPorts[p]
				if cloudletRefs.RootLbPorts != nil {
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
		clusterInst := edgeproto.ClusterInst{}
		if clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, &clusterInst) {
			if clusterInst.ReservedBy != "" && clusterInst.ReservedBy == in.Key.AppKey.DeveloperKey.Name {
				clusterInst.ReservedBy = ""
				clusterInstApi.store.STMPut(stm, &clusterInst)
			}
		}

		// delete app inst
		if ignoreCRM(cctx) {
			// CRM state should be the same as before the
			// operation failed, so just need to clean up
			// controller state.
			s.store.STMDel(stm, &in.Key)
		} else {
			in.State = edgeproto.TrackedState_DELETE_REQUESTED
			s.store.STMPut(stm, in)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if ignoreCRM(cctx) {
		cb.Send(&edgeproto.Result{Message: "Deleted AppInst successfully"})
		return nil
	}
	err = appInstApi.cache.WaitForState(ctx, &in.Key, edgeproto.TrackedState_NOT_PRESENT, DeleteAppInstTransitions, edgeproto.TrackedState_DELETE_ERROR, settingsApi.Get().DeleteAppInstTimeout.TimeDuration(), "Deleted AppInst successfully", cb.Send)
	if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Delete AppInst ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(ctx, in, edgeproto.TrackedState_NOT_PRESENT)
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
	// delete clusterinst afterwards if it was auto-created and nobody is left using it
	clusterInst := edgeproto.ClusterInst{}
	if clusterInstApi.Get(&clusterInstKey, &clusterInst) && clusterInst.Auto && !appInstApi.UsesClusterInst(&clusterInstKey) {
		cb.Send(&edgeproto.Result{Message: "Deleting auto-ClusterInst"})
		autoerr := clusterInstApi.deleteClusterInstInternal(cctx, &clusterInst, cb)
		if autoerr != nil {
			log.InfoLog("Failed to delete auto-ClusterInst",
				"clusterInst", clusterInst, "err", err)
		} else {
			RecordClusterInstEvent(ctx, &clusterInst, cloudcommon.DELETED, cloudcommon.InstanceDown)
		}
	}
	return err
}

func (s *AppInstApi) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInst) error {
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
			// got deleted in the meantime
			return nil
		}
		// healthy -> not healthy
		if inst.HealthCheck == edgeproto.HealthCheck_HEALTH_CHECK_OK && state != edgeproto.HealthCheck_HEALTH_CHECK_OK {
			RecordAppInstEvent(ctx, &inst, cloudcommon.HEALTH_CHECK_FAIL, cloudcommon.InstanceDown)
			// not healthy -> healthy
		} else if inst.HealthCheck != edgeproto.HealthCheck_HEALTH_CHECK_OK && state == edgeproto.HealthCheck_HEALTH_CHECK_OK {
			RecordAppInstEvent(ctx, &inst, cloudcommon.HEALTH_CHECK_OK, cloudcommon.InstanceUp)
		}
		inst.HealthCheck = state
		s.store.STMPut(stm, &inst)
		return nil
	})
}

func (s *AppInstApi) UpdateFromInfo(ctx context.Context, in *edgeproto.AppInstInfo) {
	log.DebugLog(log.DebugLevelApi, "Update AppInst from info", "key", in.Key, "state", in.State, "status", in.Status)
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		if inst.State == in.State {
			// already in that state
			if in.State == edgeproto.TrackedState_READY {
				// update runtime info
				inst.RuntimeInfo = in.RuntimeInfo
				inst.Status = in.Status
				s.store.STMPut(stm, &inst)
			} else if inst.Status != in.Status {
				// update status
				inst.Status = in.Status
				s.store.STMPut(stm, &inst)
			}
			return nil
		}
		// please see state_transitions.md
		if !crmTransitionOk(inst.State, in.State) {
			log.DebugLog(log.DebugLevelApi, "Invalid state transition",
				"key", &in.Key, "cur", inst.State, "next", in.State)
			return nil
		}
		inst.State = in.State
		inst.Status = in.Status
		if in.State == edgeproto.TrackedState_CREATE_ERROR || in.State == edgeproto.TrackedState_DELETE_ERROR || in.State == edgeproto.TrackedState_UPDATE_ERROR {
			inst.Errors = in.Errors
		} else {
			inst.Errors = nil
		}
		inst.RuntimeInfo = in.RuntimeInfo
		s.store.STMPut(stm, &inst)
		return nil
	})
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
		if inst.State != edgeproto.TrackedState_DELETING && inst.State != edgeproto.TrackedState_DELETE_REQUESTED {
			log.DebugLog(log.DebugLevelApi, "Invalid state transition",
				"key", &in.Key, "cur", inst.State,
				"next", edgeproto.TrackedState_NOT_PRESENT)
			return nil
		}
		s.store.STMDel(stm, &in.Key)
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
		if newState == edgeproto.TrackedState_NOT_PRESENT {
			s.store.STMDel(stm, &in.Key)
		} else {
			inst.State = newState
			inst.Errors = nil
			s.store.STMPut(stm, &inst)
		}
		return nil
	})
}

// public cloud k8s cluster allocates a separate IP per service.  This is a type of dedicated access
func isIPAllocatedPerService(operator string) bool {
	return operator == cloudcommon.OperatorGCP || operator == cloudcommon.OperatorAzure
}

func allocateIP(inst *edgeproto.ClusterInst, cloudlet *edgeproto.Cloudlet, refs *edgeproto.CloudletRefs) error {
	if isIPAllocatedPerService(cloudlet.Key.OperatorKey.Name) {
		// public cloud implements dedicated access
		inst.IpAccess = edgeproto.IpAccess_IP_ACCESS_DEDICATED
		return nil
	}
	if inst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
		// shared, so no allocation needed
		return nil
	}
	if inst.IpAccess == edgeproto.IpAccess_IP_ACCESS_UNKNOWN {
		// set to shared, as CRM does not implement dedicated
		// ip assignment yet.
		inst.IpAccess = edgeproto.IpAccess_IP_ACCESS_SHARED
		return nil
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
	if app.Deployment == cloudcommon.AppDeploymentTypeKubernetes {
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

func setL7Port(port *dme.AppPort, key *edgeproto.AppInstKey) bool {
	if port.Proto != dme.LProto_L_PROTO_HTTP {
		return false
	}
	port.PublicPort = cloudcommon.RootLBL7Port
	port.PathPrefix = cloudcommon.GetL7Path(key, port.InternalPort)
	return true
}

func RecordAppInstEvent(ctx context.Context, app *edgeproto.AppInst, event cloudcommon.InstanceEvent, serverStatus string) {
	metric := edgeproto.Metric{}
	metric.Name = cloudcommon.AppInstEvent
	ts, _ := types.TimestampProto(time.Now())
	metric.Timestamp = *ts
	metric.AddTag("operator", app.Key.ClusterInstKey.CloudletKey.OperatorKey.Name)
	metric.AddTag("cloudlet", app.Key.ClusterInstKey.CloudletKey.Name)
	metric.AddTag("cluster", app.Key.ClusterInstKey.ClusterKey.Name)
	metric.AddTag("dev", app.Key.AppKey.DeveloperKey.Name)
	metric.AddTag("app", app.Key.AppKey.Name)
	metric.AddTag("version", app.Key.AppKey.Version)
	metric.AddStringVal("event", string(event))
	metric.AddStringVal("status", serverStatus)

	services.events.AddMetric(&metric)
}
