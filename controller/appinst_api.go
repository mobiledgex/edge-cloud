package main

import (
	"context"
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
var CreateAppInstTransitions = map[edgeproto.AppState]struct{}{
	edgeproto.AppState_AppStateBuilding: struct{}{},
}
var UpdateAppInstTransitions = map[edgeproto.AppState]struct{}{
	edgeproto.AppState_AppStateChanging: struct{}{},
}
var DeleteAppInstTransitions = map[edgeproto.AppState]struct{}{
	edgeproto.AppState_AppStateDeleting: struct{}{},
}

func InitAppInstApi(sync *Sync) {
	appInstApi.sync = sync
	appInstApi.store = edgeproto.NewAppInstStore(sync.store)
	edgeproto.InitAppInstCache(&appInstApi.cache)
	appInstApi.cache.SetNotifyCb(notify.ServerMgrOne.UpdateAppInst)
	sync.RegisterCache(&appInstApi.cache)
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

func (s *AppInstApi) CreateNoWait(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	in.Liveness = edgeproto.Liveness_LivenessStatic
	return s.createAppInstInternal(in)
}

// createAppInstInternal is used to create dynamic app insts internally,
// bypassing static assignment.
func (s *AppInstApi) createAppInstInternal(in *edgeproto.AppInst) (*edgeproto.Result, error) {
	if in.Liveness == edgeproto.Liveness_LivenessUnknown {
		in.Liveness = edgeproto.Liveness_LivenessDynamic
	}

	// make sure cluster inst exists.
	// This is a separate STM to avoid ordering issues between
	// auto-clusterinst create and appinst create in watch cb.
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if s.store.STMGet(stm, in.GetKey(), in) {
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
			// auto-create cluster inst
			clusterInst := edgeproto.ClusterInst{}
			clusterInst.Key = in.ClusterInstKey
			clusterInst.Auto = true
			err := clusterInstApi.createClusterInstInternal(stm,
				&clusterInst)
			if err != nil {
				return err
			}
			log.DebugLog(log.DebugLevelApi,
				"Create auto-clusterinst",
				"key", clusterInst.Key,
				"appinst", in)
			fmt.Printf("Creating auto cluster inst %v\n", clusterInst)
		}
		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}
	defer func() {
		if err != nil {
			s.deleteClusterInstAuto(&in.ClusterInstKey)
		}
	}()

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if s.store.STMGet(stm, in.GetKey(), nil) {
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
		fmt.Printf("Creating appinst with clusterinst %v/%v\n", clusterInst.Key, clusterInst)

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

		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AppInstApi) UpdateNoWait(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	// don't allow updates to cached fields
	if in.Fields != nil {
		for _, field := range in.Fields {
			if field == edgeproto.AppInstFieldImagePath {
				return &edgeproto.Result{}, errors.New("Cannot specify image path as it is inherited from specified app")
			} else if strings.HasPrefix(field, edgeproto.AppInstFieldCloudletLoc) {
				return &edgeproto.Result{}, errors.New("Cannot specify cloudlet location fields as they are inherited from specified cloudlet")
			}
		}
	}
	return s.store.Update(in, s.sync.syncWait)
}

func (s *AppInstApi) DeleteNoWait(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	clusterInstKey := edgeproto.ClusterInstKey{}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, in.GetKey(), in) {
			// already deleted
			return objstore.ErrKVStoreKeyNotFound
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
		s.store.STMDel(stm, in.GetKey())
		return nil
	})
	if err == nil {
		// delete clusterinst afterwards if it was auto-created
		s.deleteClusterInstAuto(&clusterInstKey)
	}
	return &edgeproto.Result{}, err
}

func (s *AppInstApi) deleteClusterInstAuto(key *edgeproto.ClusterInstKey) {
	clusterInst := edgeproto.ClusterInst{}
	if clusterInstApi.Get(key, &clusterInst) && clusterInst.Auto {
		_, err := clusterInstApi.deleteClusterInstInternal(&clusterInst)
		if err != nil {
			log.InfoLog("Failed to delete auto cluster inst",
				"clusterInst", key, "err", err)
		}
	}
}

func (s *AppInstApi) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AppInstApi) CreateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_CreateAppInstServer) error {
	_, err := s.CreateNoWait(cb.Context(), in)
	if err == objstore.ErrKVStoreKeyExists {
		// Already created, but check if not present in CRM.
		// If so, trigger update to try to get CRM back in sync.
		info := edgeproto.AppInstInfo{}
		if !appInstInfoApi.cache.Get(&in.Key, &info) {
			notify.ServerMgrOne.UpdateAppInst(&in.Key, in)
		}
	}
	if err != nil {
		return err
	}
	err = appInstInfoApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.AppState_AppStateReady, CreateAppInstTransitions, CreateAppInstTimeout, "Created successfully", cb.Send)
	if err != nil {
		// XXX should probably track mod revision ID and only undo
		// if no other changes were made to appInst in the meantime.
		// crm failed or some other err, undo
		_, undoErr := s.DeleteNoWait(cb.Context(), in)
		if undoErr != nil {
			log.InfoLog("Undo create appinst", "undoErr", undoErr)
		} else {
			undoErr = appInstInfoApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.AppState_AppStateNotPresent, DeleteAppInstTransitions, DeleteAppInstTimeout, "", nil)
			if undoErr != nil {
				log.InfoLog("Undo create appinst", "undoErr", undoErr)
			}
		}
		log.DebugLog(log.DebugLevelApi, "Create app inst failed",
			"in", in, "err", err)
	}
	return err
}

func (s *AppInstApi) DeleteAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	_, err := s.DeleteNoWait(cb.Context(), in)
	if err == objstore.ErrKVStoreKeyNotFound {
		// Already deleted, but check if still present in CRM.
		// If so, trigger update to try to get CRM back in sync.
		info := edgeproto.AppInstInfo{}
		if appInstInfoApi.cache.Get(&in.Key, &info) {
			notify.ServerMgrOne.UpdateAppInst(&in.Key, in)
		}
	}
	if err != nil {
		return err
	}
	err = appInstInfoApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.AppState_AppStateNotPresent, DeleteAppInstTransitions, DeleteAppInstTimeout, "Deleted successfully", cb.Send)
	if err != nil {
		// crm failed or some other err, undo
		_, undoErr := s.CreateNoWait(cb.Context(), in)
		if undoErr != nil {
			log.InfoLog("Undo delete appinst", "undoErr", undoErr)
		} else {
			undoErr = appInstInfoApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.AppState_AppStateReady, CreateAppInstTransitions, CreateAppInstTimeout, "", nil)
			if undoErr != nil {
				log.InfoLog("Undo delete appinst", "undoErr", undoErr)
			}
		}
		log.DebugLog(log.DebugLevelApi, "Delete app inst failed",
			"in", in, "err", err)
	}
	return err
}

func (s *AppInstApi) UpdateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_UpdateAppInstServer) error {
	_, err := s.UpdateNoWait(cb.Context(), in)
	if err != nil {
		return err
	}
	err = appInstInfoApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.AppState_AppStateReady, UpdateAppInstTransitions, UpdateAppInstTimeout, "Updated successfully", cb.Send)
	if err != nil {
		// TODO: undo and mod rev checking.
	}
	return err
}
