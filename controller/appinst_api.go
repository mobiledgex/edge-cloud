package main

import (
	"context"
	"errors"
	"strings"

	"github.com/coreos/etcd/clientv3/concurrency"
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

var appInstApi = AppInstApi{}

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

func (s *AppInstApi) CreateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
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
		if s.store.STMGet(stm, in.GetKey(), nil) {
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

		// CloudletKey is duplicated in both the ClusterInstKey and the
		// AppInstKey. It's inconsistent to create an AppInst on a ClusterInst
		// when the cloudlet keys don't match, so don't require the user to
		// specify the CloudletKey in the cluster during AppInst create, just
		// take it from the AppInst Key.
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
		in.ConfigMap = app.ConfigMap
		in.AccessLayer = app.AccessLayer
		in.Flavor = app.Flavor

		if !clusterInstApi.store.STMGet(stm, &in.ClusterInstKey, nil) {
			return errors.New("Cluster instance does not exist for app")
		}

		// set up URI for this instance
		in.Uri = util.DNSSanitize(in.Key.AppKey.Name) + "." +
			util.DNSSanitize(in.Key.CloudletKey.Name) + "." +
			util.AppDNSRoot
		// TODO:
		// Allocate mapped ports and mapped path(s)
		// Reserve resources
		in.MappedPorts = app.AccessPorts
		in.MappedPath = util.DNSSanitize(app.Key.Name)

		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AppInstApi) UpdateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
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

func (s *AppInstApi) DeleteAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	clusterInstKey := edgeproto.ClusterInstKey{}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		appinst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, in.GetKey(), &appinst) {
			// already deleted
			return nil
		}
		clusterInstKey = appinst.ClusterInstKey
		// delete associated info
		appInstInfoApi.internalDelete(stm, in.GetKey())
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
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		clusterInst := edgeproto.ClusterInst{}
		if clusterInstApi.store.STMGet(stm, key, &clusterInst) && clusterInst.Auto {
			clusterInstApi.store.STMDel(stm, key)
		}
		return nil
	})
	if err != nil {
		log.InfoLog("Failed to delete auto cluster inst",
			"clusterInst", key, "err", err)
	}
}

func (s *AppInstApi) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
