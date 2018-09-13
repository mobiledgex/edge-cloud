package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/util"
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
	appInstApi.cache.SetNotifyCb(notify.ServerMgrOne.UpdateAppInst)
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

func (s *AppInstApi) GetAppInstsForCloudlets(cloudlets map[edgeproto.CloudletKey]struct{}, appInsts map[edgeproto.AppInstKey]struct{}) {
	s.cache.GetAppInstsForCloudlets(cloudlets, appInsts)
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
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, val := range s.cache.Objs {
		if val.ClusterInstKey.Matches(key) {
			return true
		}
	}
	return false
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
	return s.createAppInstInternal(DefCallContext, in, cb)
}

// createAppInstInternal is used to create dynamic app insts internally,
// bypassing static assignment.
func (s *AppInstApi) createAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, cb edgeproto.AppInstApi_CreateAppInstServer) error {
	if in.Liveness == edgeproto.Liveness_LivenessUnknown {
		in.Liveness = edgeproto.Liveness_LivenessDynamic
	}

	var autocluster bool
	// See if we need to create auto-cluster.
	// This also sets up the correct ClusterInstKey in "in".
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		autocluster = false
		if s.store.STMGet(stm, &in.Key, in) {
			return objstore.ErrKVStoreKeyExists
		}
		// make sure cloudlet exists
		if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, nil) {
			return errors.New("Specified cloudlet not found")
		}

		// find cluster from app
		var app edgeproto.App
		if !appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return errors.New("Specified app not found")
		}

		// ClusterInstKey is derived from App's Cluster and
		// AppInst's Cloudlet. It is not specifiable by the user.
		// It is kept here is a shortcut to make looking up the
		// clusterinst object easier.
		in.ClusterInstKey.CloudletKey = in.Key.CloudletKey
		in.ClusterInstKey.ClusterKey = app.Cluster

		if !clusterInstApi.store.STMGet(stm, &in.ClusterInstKey, nil) {
			// cluster inst does not exist, see if we can create it.
			// because app cannot exist without cluster, cluster
			// should exist. But check anyway.
			if !clusterApi.store.STMGet(stm, &app.Cluster, nil) {
				return errors.New("Cluster does not exist for app")
			}
			autocluster = true
		}
		return nil
	})
	if err != nil {
		return err
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
		err := clusterInstApi.createClusterInstInternal(cctx, &clusterInst, cb)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil && !cctx.Undo {
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

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyExists
		}

		// cache location of cloudlet in app inst
		var cloudlet edgeproto.Cloudlet
		if !cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudlet) {
			return errors.New("Specified cloudlet not found")
		}
		in.CloudletLoc = cloudlet.Location

		// cache app path in app inst
		var app edgeproto.App
		if !appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return errors.New("Specified app not found")
		}
		in.ImagePath = app.ImagePath
		in.ImageType = app.ImageType
		in.Config = app.Config
		in.AccessLayer = app.AccessLayer
		if in.Flavor.Name == "" {
			in.Flavor = app.DefaultFlavor
		}

		if !flavorApi.store.STMGet(stm, &in.Flavor, nil) {
			return fmt.Errorf("Flavor %s not found", in.Flavor.Name)
		}

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

		cloudletRefs := edgeproto.CloudletRefs{}
		cloudletRefsChanged := false
		if !cloudletRefsApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudletRefs) {
			initCloudletRefs(&cloudletRefs, &in.Key.CloudletKey)
		}

		ports, _ := parseAppPorts(app.AccessPorts)

		// shared root load balancer
		// dedicated load balancer not supported yet
		in.Uri = cloudcommon.GetRootLBFQDN(&in.Key.CloudletKey)
		if len(ports) > 0 {
			if cloudletRefs.RootLbPorts == nil {
				cloudletRefs.RootLbPorts = make(map[int32]int32)
			}
			ii := 0
			p := RootLBSharedPortBegin
			for ; p < 65000 && ii < len(ports); p++ {
				if _, found := cloudletRefs.RootLbPorts[p]; found {
					continue
				}
				ports[ii].PublicPort = p
				cloudletRefs.RootLbPorts[p] = 1
				ii++
				cloudletRefsChanged = true
			}
		}
		in.MappedPath = util.DNSSanitize(app.Key.Name)
		if len(ports) > 0 {
			in.MappedPorts = ports
		}

		// TODO: Make sure resources are available
		if cloudletRefsChanged {
			cloudletRefsApi.store.STMPut(stm, &cloudletRefs)
		}

		in.State = edgeproto.TrackedState_CreateRequested
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}
	err = appInstApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.TrackedState_Ready, CreateAppInstTransitions, edgeproto.TrackedState_CreateError, CreateAppInstTimeout, "Created successfully", cb.Send)
	if err != nil && !cctx.Undo {
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
			if field == edgeproto.AppInstFieldImagePath {
				return errors.New("Cannot specify image path as it is inherited from specified app")
			} else if strings.HasPrefix(field, edgeproto.AppInstFieldCloudletLoc) {
				return errors.New("Cannot specify cloudlet location fields as they are inherited from specified cloudlet")
			}
		}
	}
	return errors.New("Update app instance not supported yet")
}

func (s *AppInstApi) DeleteAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	return s.deleteAppInstInternal(DefCallContext, in, cb)
}

func (s *AppInstApi) deleteAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	clusterInstKey := edgeproto.ClusterInstKey{}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			// already deleted
			return objstore.ErrKVStoreKeyNotFound
		}
		if in.State != edgeproto.TrackedState_Ready && !cctx.Undo {
			return errors.New("AppInst not ready, cannot delete")
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
		// delete app inst
		in.State = edgeproto.TrackedState_DeleteRequested
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}
	err = appInstApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.TrackedState_NotPresent, DeleteAppInstTransitions, edgeproto.TrackedState_DeleteError, DeleteAppInstTimeout, "Deleted AppInst successfully", cb.Send)
	if err != nil && !cctx.Undo {
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
