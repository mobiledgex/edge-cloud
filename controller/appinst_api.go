package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/util"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AppInstApi struct {
	sync  *Sync
	store edgeproto.AppInstStore
	cache edgeproto.AppInstCache
}

const RootLBSharedPortBegin int32 = 10000

var appInstApi = AppInstApi{}

// TODO: these timeouts should be adjust based on target platform,
// as some platforms (azure, etc) may take much longer.
// These timeouts should be at least long enough for the controller and
// CRM to exchange an initial set of messages (i.e. 10 sec or so).
var CreateAppInstTimeout = 30 * time.Minute
var UpdateAppInstTimeout = 20 * time.Minute
var DeleteAppInstTimeout = 20 * time.Minute

// Transition states indicate states in which the CRM is still busy.
var CreateAppInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_Creating: struct{}{},
}
var UpdateAppInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_Updating: struct{}{},
}
var DeleteAppInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_Deleting: struct{}{},
}

func InitAppInstApi(sync *Sync) {
	appInstApi.sync = sync
	appInstApi.store = edgeproto.NewAppInstStore(sync.store)
	edgeproto.InitAppInstCache(&appInstApi.cache)
	sync.RegisterCache(&appInstApi.cache)
	if *shortTimeouts {
		CreateAppInstTimeout = 3 * time.Second
		UpdateAppInstTimeout = 2 * time.Second
		DeleteAppInstTimeout = 2 * time.Second
	}
}

func (s *AppInstApi) GetAllKeys(keys map[edgeproto.AppInstKey]struct{}) {
	s.cache.GetAllKeys(keys)
}

func (s *AppInstApi) Get(key *edgeproto.AppInstKey, val *edgeproto.AppInst) bool {
	return s.cache.Get(key, val)
}

func (s *AppInstApi) HasKey(key *edgeproto.AppInstKey) bool {
	return s.cache.HasKey(key)
}

func (s *AppInstApi) UsesCloudlet(in *edgeproto.CloudletKey, dynInsts map[edgeproto.AppInstKey]struct{}) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	static := false
	for key, val := range s.cache.Objs {
		if key.CloudletKey.Matches(in) {
			if val.Liveness == edgeproto.Liveness_LivenessStatic {
				static = true
			} else if val.Liveness == edgeproto.Liveness_LivenessDynamic {
				dynInsts[key] = struct{}{}
			}
		}
	}
	return static
}

func (s *AppInstApi) UsesApp(in *edgeproto.AppKey, dynInsts map[edgeproto.AppInstKey]struct{}) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	static := false
	for key, val := range s.cache.Objs {
		if key.AppKey.Matches(in) {
			if val.Liveness == edgeproto.Liveness_LivenessStatic {
				static = true
			} else if val.Liveness == edgeproto.Liveness_LivenessDynamic {
				dynInsts[key] = struct{}{}
			}
		}
	}
	return static
}

func (s *AppInstApi) UsesClusterInst(key *edgeproto.ClusterInstKey) bool {
	var app edgeproto.App
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, val := range s.cache.Objs {
		if val.ClusterInstKey.Matches(key) && appApi.Get(&val.Key.AppKey, &app) {
			log.DebugLog(log.DebugLevelApi, "AppInst found for clusterInst", "app", app.Key.Name,
				"autodelete", app.DelOpt.String())
			if app.DelOpt == edgeproto.DeleteType_NoAutoDelete {
				return true
			}
		}
	}
	return false
}

func (s *AppInstApi) AutoDeleteAppInsts(key *edgeproto.ClusterInstKey, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	var app edgeproto.App
	apps := make(map[edgeproto.AppInstKey]*edgeproto.AppInst)
	log.DebugLog(log.DebugLevelApi, "Auto-deleting appinsts ", "cluster", key.ClusterKey.Name)
	s.cache.Mux.Lock()
	for k, val := range s.cache.Objs {
		if val.ClusterInstKey.Matches(key) && appApi.Get(&val.Key.AppKey, &app) {
			if app.DelOpt == edgeproto.DeleteType_AutoDelete {
				apps[k] = val
			}
		}
	}
	s.cache.Mux.Unlock()
	for _, val := range apps {
		log.DebugLog(log.DebugLevelApi, "Auto-deleting appinst ", "appinst", val.Key.AppKey.Name)
		cb.Send(&edgeproto.Result{Message: "Autodeleting appinst " + val.Key.AppKey.Name})
		if err := s.deleteAppInstInternal(DefCallContext(), val, cb); err != nil {
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
	in.Liveness = edgeproto.Liveness_LivenessStatic
	return s.createAppInstInternal(DefCallContext(), in, cb)
}

// createAppInstInternal is used to create dynamic app insts internally,
// bypassing static assignment.
func (s *AppInstApi) createAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, cb edgeproto.AppInstApi_CreateAppInstServer) (reterr error) {
	if in.Liveness == edgeproto.Liveness_LivenessUnknown {
		in.Liveness = edgeproto.Liveness_LivenessDynamic
	}
	cctx.SetOverride(&in.CrmOverride)
	if err := cloudletInfoApi.checkCloudletReady(&in.Key.CloudletKey); err != nil {
		return err
	}

	var autocluster bool
	// See if we need to create auto-cluster.
	// This also sets up the correct ClusterInstKey in "in".

	// indicates special default cloudlet maintained by the developer
	var defaultCloudlet bool

	if in.Key.CloudletKey == cloudcommon.DefaultCloudletKey {
		log.DebugLog(log.DebugLevelApi, "special default public cloud case", "appinst", in)
		defaultCloudlet = true
	}

	if !defaultCloudlet {
		err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
			autocluster = false
			if s.store.STMGet(stm, &in.Key, in) {
				if !cctx.Undo && in.State != edgeproto.TrackedState_DeleteError && !ignoreTransient(cctx, in.State) {
					if in.State == edgeproto.TrackedState_CreateError {
						cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
						cb.Send(&edgeproto.Result{Message: "Use DeleteAppInst to remove and try again"})
					}
					return objstore.ErrKVStoreKeyExists
				}
				in.Errors = nil
				if !defaultCloudlet {
					// must reset Uri
					in.Uri = ""
				}
			} else {
				err := in.Validate(edgeproto.AppInstAllFieldsMap)
				if err != nil {
					return err
				}
			}
			// make sure cloudlet exists
			if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, nil) {
				return errors.New("Specified cloudlet not found")
			}

			cikey := &in.ClusterInstKey
			if cikey.CloudletKey.Name != "" || cikey.CloudletKey.OperatorKey.Name != "" {
				// Make sure ClusterInst cloudlet key matches
				// AppInst's cloudlet key to prevent confusion
				// if user specifies both.
				if !in.Key.CloudletKey.Matches(&cikey.CloudletKey) {
					return errors.New("Specified ClusterInst cloudlet key does not match specified AppInst cloudlet key")
				}

			}
			in.ClusterInstKey.CloudletKey = in.Key.CloudletKey
			// Explicit auto-cluster requirement
			if cikey.ClusterKey.Name == "" {
				return fmt.Errorf("No cluster name specified. Create one first or use \"%s\" as the name to automatically create a ClusterInst", ClusterAutoPrefix)
			}
			// Check if specified ClusterInst exists
			if cikey.ClusterKey.Name != ClusterAutoPrefix {
				if !clusterInstApi.store.STMGet(stm, &in.ClusterInstKey, nil) {
					// developer may or may not be specified
					// in clusterinst.
					in.ClusterInstKey.Developer = in.Key.AppKey.DeveloperKey.Name
					if !clusterInstApi.store.STMGet(stm, &in.ClusterInstKey, nil) {
						return errors.New("Specified ClusterInst not found")
					}
				}
				// cluster inst exists so we're good.
				return nil
			}
			// Auto-cluster
			cikey.ClusterKey.Name = fmt.Sprintf("%s%s", ClusterAutoPrefix, in.Key.AppKey.Name)
			cikey.ClusterKey.Name = util.K8SSanitize(cikey.ClusterKey.Name)
			autocluster = true

			if in.Flavor.Name == "" {
				// find flavor from app
				var app edgeproto.App
				if !appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
					return edgeproto.ErrEdgeApiAppNotFound
				}
				in.Flavor = app.DefaultFlavor
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	if autocluster {
		// auto-create cluster inst
		clusterInst := edgeproto.ClusterInst{}
		clusterInst.Key = in.ClusterInstKey
		clusterInst.Auto = true
		log.DebugLog(log.DebugLevelApi,
			"Create auto-clusterinst",
			"key", clusterInst.Key,
			"appinst", in)
		clusterflavorKey, err := GetClusterFlavorForFlavor(&in.Flavor)
		if err != nil {
			return fmt.Errorf("error getting flavor for auto ClusterInst, %s", err.Error())
		}
		clusterInst.Flavor = *clusterflavorKey
		err = clusterInstApi.createClusterInstInternal(cctx, &clusterInst, cb)
		if err != nil {
			return err
		}
		defer func() {
			if reterr != nil && !cctx.Undo {
				cb.Send(&edgeproto.Result{Message: "Deleting auto-ClusterInst due to failure"})
				undoErr := clusterInstApi.deleteClusterInstInternal(cctx.WithUndo(), &clusterInst, cb)
				if undoErr != nil {
					log.DebugLog(log.DebugLevelApi,
						"Undo create auto-clusterinst failed",
						"key", clusterInst.Key,
						"undoErr", undoErr)
				}
			}
		}()
	}

	// these checks cannot be in ApplySTMWait funcs since
	// that func may run several times if the database changes while
	// it is running, and the stm func modifies in.Uri.
	if in.Uri == "" && defaultCloudlet {
		return errors.New("URI (Public FQDN) is required for default cloudlet")
	} else if in.Uri != "" && !defaultCloudlet {
		return fmt.Errorf("Cannot specify URI %s for non-default cloudlet", in.Uri)
	}

	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		buf := in
		if !defaultCloudlet {
			// lookup already done, don't overwrite changes
			buf = nil
		}
		if s.store.STMGet(stm, &in.Key, buf) {
			if !cctx.Undo && in.State != edgeproto.TrackedState_DeleteError {
				if in.State == edgeproto.TrackedState_CreateError {
					cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
					cb.Send(&edgeproto.Result{Message: "Use DeleteAppInst to remove and try again"})
				}
				return objstore.ErrKVStoreKeyExists
			}
			in.Errors = nil
		} else {
			err := in.Validate(edgeproto.AppInstAllFieldsMap)
			if err != nil {
				return err
			}
		}

		// cache location of cloudlet in app inst
		var cloudlet edgeproto.Cloudlet

		if !defaultCloudlet {
			if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudlet) {
				return errors.New("Specified cloudlet not found")
			}
			in.CloudletLoc = cloudlet.Location
		} else {
			in.CloudletLoc = dme.Loc{}
		}

		var app edgeproto.App
		if !appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return edgeproto.ErrEdgeApiAppNotFound
		}
		if in.Flavor.Name == "" {
			in.Flavor = app.DefaultFlavor
		}

		if !flavorApi.store.STMGet(stm, &in.Flavor, nil) {
			return fmt.Errorf("Flavor %s not found", in.Flavor.Name)
		}

		ipaccess := edgeproto.IpAccess_IpAccessDedicated
		if !defaultCloudlet {
			clusterInst := edgeproto.ClusterInst{}
			if !clusterInstApi.store.STMGet(stm, &in.ClusterInstKey, &clusterInst) {
				return errors.New("Cluster instance does not exist for app")
			}
			if clusterInst.State != edgeproto.TrackedState_Ready {
				return fmt.Errorf("ClusterInst %s not ready", clusterInst.Key.GetKeyString())
			}

			var info edgeproto.CloudletInfo
			if !cloudletInfoApi.store.STMGet(stm, &in.Key.CloudletKey, &info) {
				return errors.New("Info for cloudlet not found")
			}
			ipaccess = clusterInst.IpAccess
		}

		cloudletRefs := edgeproto.CloudletRefs{}
		cloudletRefsChanged := false
		if !cloudletRefsApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudletRefs) {
			initCloudletRefs(&cloudletRefs, &in.Key.CloudletKey)
		}

		ports, _ := edgeproto.ParseAppPorts(app.AccessPorts)
		if defaultCloudlet {
			// nothing to do
		} else if ipaccess == edgeproto.IpAccess_IpAccessShared {
			in.Uri = cloudcommon.GetRootLBFQDN(&in.Key.CloudletKey)
			if cloudletRefs.RootLbPorts == nil {
				cloudletRefs.RootLbPorts = make(map[int32]int32)
			}

			p := RootLBSharedPortBegin
			for ii, _ := range ports {
				// Samsung enabling layer ignores port mapping.
				// Attempt to use the internal port as the
				// external port so port remap is not required.
				eport := int32(-1)
				if _, found := cloudletRefs.RootLbPorts[ports[ii].InternalPort]; !found {
					// rootLB has its own ports it uses
					// before any apps are even present.
					iport := ports[ii].InternalPort
					if iport != 22 &&
						iport != 18888 &&
						iport != 18889 {
						eport = iport
					}
				}
				for ; p < 65000 && eport == int32(-1); p++ {
					// each kubernetes service gets its own
					// nginx proxy that runs in the rootLB,
					// and http ports are also mapped to it,
					// so there is no shared L7 port + path.
					if _, found := cloudletRefs.RootLbPorts[p]; found {
						continue
					}
					eport = p
				}
				if eport == int32(-1) {
					return errors.New("no free external ports")
				}
				ports[ii].PublicPort = eport
				cloudletRefs.RootLbPorts[eport] = 1
				cloudletRefsChanged = true
			}
		} else {
			in.Uri = cloudcommon.GetAppFQDN(&in.Key)
			for ii, _ := range ports {
				ports[ii].PublicPort = ports[ii].InternalPort
			}
		}
		if len(ports) > 0 {
			in.MappedPorts = ports
			if !defaultCloudlet {
				setPortFQDNPrefixes(in, &app)
			}
		}

		// TODO: Make sure resources are available
		if cloudletRefsChanged {
			cloudletRefsApi.store.STMPut(stm, &cloudletRefs)
		}
		in.CreatedAt = cloudcommon.TimeToTimestamp(time.Now())

		if ignoreCRM(cctx) || defaultCloudlet {
			in.State = edgeproto.TrackedState_Ready
		} else {
			in.State = edgeproto.TrackedState_CreateRequested
		}
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}
	if ignoreCRM(cctx) || defaultCloudlet {
		return nil
	}
	err = appInstApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.TrackedState_Ready, CreateAppInstTransitions, edgeproto.TrackedState_CreateError, CreateAppInstTimeout, "Created successfully", cb.Send)
	if err != nil && cctx.Override == edgeproto.CRMOverride_IgnoreCRMErrors {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Create AppInst ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(in, edgeproto.TrackedState_Ready)
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
			log.InfoLog("Undo create appinst", "undoErr", undoErr)
		}
	}
	return err
}

func (s *AppInstApi) updateAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, cb edgeproto.AppInstApi_UpdateAppInstServer) error {
	_, err := s.store.Update(in, s.sync.syncWait)
	return err
}

func (s *AppInstApi) UpdateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_UpdateAppInstServer) error {
	// don't allow updates to cached fields
	if in.Fields != nil {
		for _, field := range in.Fields {
			if strings.HasPrefix(field, edgeproto.AppInstFieldCloudletLoc) {
				return errors.New("Cannot specify cloudlet location fields as they are inherited from specified cloudlet")
			}
		}
	}
	return errors.New("Update app instance not supported yet")
}

func (s *AppInstApi) DeleteAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	return s.deleteAppInstInternal(DefCallContext(), in, cb)
}

func (s *AppInstApi) deleteAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	cctx.SetOverride(&in.CrmOverride)

	var defaultCloudlet bool

	log.DebugLog(log.DebugLevelApi, "deleteAppInstInternal", "appinst", in)

	if in.Key.CloudletKey == cloudcommon.DefaultCloudletKey {
		log.DebugLog(log.DebugLevelApi, "special public cloud case", "appinst", in)
		defaultCloudlet = true
	}

	if err := cloudletInfoApi.checkCloudletReady(&in.Key.CloudletKey); err != nil {
		return err
	}
	clusterInstKey := edgeproto.ClusterInstKey{}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			// already deleted
			return objstore.ErrKVStoreKeyNotFound
		}
		if !cctx.Undo && in.State != edgeproto.TrackedState_Ready && in.State != edgeproto.TrackedState_CreateError && !ignoreTransient(cctx, in.State) {
			if in.State == edgeproto.TrackedState_DeleteError {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous delete failed, %v", in.Errors)})
				cb.Send(&edgeproto.Result{Message: "Use CreateAppInst to rebuild, and try again"})
			}
			return errors.New("AppInst busy, cannot delete")
		}

		var cloudlet edgeproto.Cloudlet
		if !defaultCloudlet {
			if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudlet) {
				return errors.New("Specified cloudlet not found")
			}

			cloudletRefs := edgeproto.CloudletRefs{}
			cloudletRefsChanged := false
			if cloudletRefsApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudletRefs) {
				// shared root load balancer
				for ii, _ := range in.MappedPorts {
					p := in.MappedPorts[ii].PublicPort
					delete(cloudletRefs.RootLbPorts, p)
					cloudletRefsChanged = true
				}
			}
			clusterInstKey = in.ClusterInstKey
			if cloudletRefsChanged {
				cloudletRefsApi.store.STMPut(stm, &cloudletRefs)
			}
		}
		// delete app inst
		if ignoreCRM(cctx) || defaultCloudlet {
			// CRM state should be the same as before the
			// operation failed, so just need to clean up
			// controller state.
			s.store.STMDel(stm, &in.Key)
		} else {
			in.State = edgeproto.TrackedState_DeleteRequested
			s.store.STMPut(stm, in)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if ignoreCRM(cctx) || defaultCloudlet {
		return nil
	}
	err = appInstApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.TrackedState_NotPresent, DeleteAppInstTransitions, edgeproto.TrackedState_DeleteError, DeleteAppInstTimeout, "Deleted AppInst successfully", cb.Send)
	if err != nil && cctx.Override == edgeproto.CRMOverride_IgnoreCRMErrors {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Delete AppInst ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(in, edgeproto.TrackedState_NotPresent)
		cb.Send(&edgeproto.Result{Message: "Deleted AppInst successfully"})
		err = nil
	}
	if err != nil {
		// crm failed or some other err, undo
		cb.Send(&edgeproto.Result{Message: "Recreating AppInst due to failure"})
		undoErr := s.createAppInstInternal(cctx.WithUndo(), in, cb)
		if undoErr != nil {
			log.InfoLog("Undo delete appinst", "undoErr", undoErr)
		}
		return err
	} else {
		// delete clusterinst afterwards if it was auto-created
		clusterInst := edgeproto.ClusterInst{}
		if clusterInstApi.Get(&clusterInstKey, &clusterInst) && clusterInst.Auto {
			cb.Send(&edgeproto.Result{Message: "Deleting auto-cluster inst"})
			autoerr := clusterInstApi.deleteClusterInstInternal(cctx, &clusterInst, cb)
			if autoerr != nil {
				log.InfoLog("Failed to delete auto cluster inst",
					"clusterInst", clusterInst, "err", err)
			}
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

func (s *AppInstApi) UpdateFromInfo(in *edgeproto.AppInstInfo) {
	log.DebugLog(log.DebugLevelApi, "Update AppInst from info", "key", in.Key, "state", in.State)
	s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		if inst.State == in.State {
			// already in that state
			return nil
		}
		// please see state_transitions.md
		if !crmTransitionOk(inst.State, in.State) {
			log.DebugLog(log.DebugLevelApi, "Invalid state transition",
				"key", &in.Key, "cur", inst.State, "next", in.State)
			return nil
		}
		inst.State = in.State
		if in.State == edgeproto.TrackedState_CreateError || in.State == edgeproto.TrackedState_DeleteError || in.State == edgeproto.TrackedState_UpdateError {
			inst.Errors = in.Errors
		}
		s.store.STMPut(stm, &inst)
		return nil
	})
}

func (s *AppInstApi) DeleteFromInfo(in *edgeproto.AppInstInfo) {
	log.DebugLog(log.DebugLevelApi, "Delete AppInst from info", "key", in.Key, "state", in.State)
	s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		// please see state_transitions.md
		if inst.State != edgeproto.TrackedState_Deleting && inst.State != edgeproto.TrackedState_DeleteRequested {
			log.DebugLog(log.DebugLevelApi, "Invalid state transition",
				"key", &in.Key, "cur", inst.State,
				"next", edgeproto.TrackedState_NotPresent)
			return nil
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
}

func (s *AppInstApi) ReplaceErrorState(in *edgeproto.AppInst, newState edgeproto.TrackedState) {
	s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		if inst.State != edgeproto.TrackedState_CreateError &&
			inst.State != edgeproto.TrackedState_DeleteError &&
			inst.State != edgeproto.TrackedState_UpdateError {
			return nil
		}
		if newState == edgeproto.TrackedState_NotPresent {
			s.store.STMDel(stm, &in.Key)
		} else {
			inst.State = newState
			inst.Errors = nil
			s.store.STMPut(stm, &inst)
		}
		return nil
	})
}

func allocateIP(inst *edgeproto.ClusterInst, cloudlet *edgeproto.Cloudlet, refs *edgeproto.CloudletRefs) error {
	if inst.Key.CloudletKey.OperatorKey.Name == cloudcommon.OperatorGCP || inst.Key.CloudletKey.OperatorKey.Name == cloudcommon.OperatorAzure {
		// public cloud implements dedicated access
		inst.IpAccess = edgeproto.IpAccess_IpAccessDedicated
		return nil
	}
	if inst.IpAccess == edgeproto.IpAccess_IpAccessShared {
		// shared, so no allocation needed
		return nil
	}
	if inst.IpAccess == edgeproto.IpAccess_IpAccessDedicatedOrShared {
		// set to shared, as CRM does not implement dedicated
		// ip assignment yet.
		inst.IpAccess = edgeproto.IpAccess_IpAccessShared
		return nil
	}

	// Allocate a dedicated IP
	if cloudlet.IpSupport == edgeproto.IpSupport_IpSupportStatic {
		// TODO:
		// parse cloudlet.StaticIps and refs.UsedStaticIps.
		// pick a free one, put it in refs.UsedStaticIps, and
		// set inst.AllocatedIp to the Ip.
		return errors.New("Static IPs not supported yet")
	} else if cloudlet.IpSupport == edgeproto.IpSupport_IpSupportDynamic {
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
	if inst.IpAccess == edgeproto.IpAccess_IpAccessShared {
		return
	}
	if cloudlet.IpSupport == edgeproto.IpSupport_IpSupportStatic {
		// TODO: free static ip in inst.AllocatedIp from refs.
	} else if cloudlet.IpSupport == edgeproto.IpSupport_IpSupportDynamic {
		refs.UsedDynamicIps--
		inst.AllocatedIp = ""
	}
}

func setPortFQDNPrefixes(in *edgeproto.AppInst, app *edgeproto.App) error {
	// For Kubernetes deployments, the CRM sets the
	// FQDN based on the service (load balancer) name
	// in the kubernetes deployment manifest.
	// The Controller needs to set a matching
	// FQDNPrefix on the ports so the DME can tell the
	// App Client the correct FQDN for a given port.
	if app.Deployment == cloudcommon.AppDeploymentTypeKubernetes {
		objs, _, err := cloudcommon.DecodeK8SYaml(app.DeploymentManifest)
		if err != nil {
			return fmt.Errorf("invalid kubernetes deployment yaml, %s", err.Error())
		}
		for ii, _ := range in.MappedPorts {
			err = setPortFQDNPrefix(&in.MappedPorts[ii], objs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func setPortFQDNPrefix(port *dme.AppPort, objs []runtime.Object) error {
	for _, obj := range objs {
		ksvc, ok := obj.(*v1.Service)
		if !ok {
			continue
		}
		for _, kp := range ksvc.Spec.Ports {
			if kp.TargetPort.IntValue() == int(port.InternalPort) {
				port.FQDNPrefix = cloudcommon.FQDNPrefix(ksvc.Name)
				return nil
			}
		}
	}
	return fmt.Errorf("no service for app port %d found in manifest", port.InternalPort)
}
