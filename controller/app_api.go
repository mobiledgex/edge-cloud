// app config

package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/util"
)

// Should only be one of these instantiated in main
type AppApi struct {
	sync  *Sync
	store edgeproto.AppStore
	cache edgeproto.AppCache
}

var appApi = AppApi{}

func InitAppApi(sync *Sync) {
	appApi.sync = sync
	appApi.store = edgeproto.NewAppStore(sync.store)
	edgeproto.InitAppCache(&appApi.cache)
	sync.RegisterCache(&appApi.cache)
	appApi.cache.SetUpdatedCb(appApi.UpdatedCb)
}

func (s *AppApi) HasApp(key *edgeproto.AppKey) bool {
	return s.cache.HasKey(key)
}

func (s *AppApi) Get(key *edgeproto.AppKey, buf *edgeproto.App) bool {
	return s.cache.Get(key, buf)
}

func (s *AppApi) UsesDeveloper(in *edgeproto.DeveloperKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, _ := range s.cache.Objs {
		if key.DeveloperKey.Matches(in) {
			return true
		}
	}
	return false
}

func (s *AppApi) UsesFlavor(key *edgeproto.FlavorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, app := range s.cache.Objs {
		if app.Flavor.Matches(key) {
			return true
		}
	}
	return false
}

func (s *AppApi) UsesCluster(key *edgeproto.ClusterKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, app := range s.cache.Objs {
		if app.Cluster.Matches(key) {
			return true
		}
	}
	return false
}

func (s *AppApi) CreateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	var err error
	if in.ImageType == edgeproto.ImageType_ImageTypeUnknown {
		return &edgeproto.Result{}, errors.New("Please specify Image Type")
	}
	if in.AccessLayer == edgeproto.AccessLayer_AccessLayerUnknown {
		// default to L7
		in.AccessLayer = edgeproto.AccessLayer_AccessLayerL7
	}
	if in.AccessPorts == "" && (in.AccessLayer == edgeproto.AccessLayer_AccessLayerL4 || in.AccessLayer == edgeproto.AccessLayer_AccessLayerL4L7) {
		return &edgeproto.Result{}, errors.New("Please specify ports for L4 access types")
	}
	if in.ImageType == edgeproto.ImageType_ImageTypeDocker {
		in.ImagePath = "mobiledgex_" +
			util.DockerSanitize(in.Key.DeveloperKey.Name) + "/" +
			util.DockerSanitize(in.Key.Name) + ":" +
			util.DockerSanitize(in.Key.Version)
	}

	// make sure cluster exists
	// This is a separate STM to avoid ordering issues between
	// auto-cluster create and app create in watch cb.
	if in.Cluster.Name == "" {
		err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
			if s.store.STMGet(stm, &in.Key, nil) {
				return objstore.ErrKVStoreKeyExists
			}
			// cluster not specified, create one automatically
			in.Cluster.Name = fmt.Sprintf("%s%s", ClusterAutoPrefix,
				in.Key.Name)
			in.Cluster.Name = util.K8SSanitize(in.Cluster.Name)
			cluster := edgeproto.Cluster{}
			cluster.Key = in.Cluster
			cluster.Flavor = in.Flavor
			cluster.Nodes = ClusterAutoNodes
			cluster.Auto = true
			// it may be possible that cluster already exists
			if !clusterApi.store.STMGet(stm, &cluster.Key, nil) {
				log.DebugLog(log.DebugLevelApi,
					"Create auto-cluster",
					"key", cluster.Key,
					"app", in)
				clusterApi.store.STMPut(stm, &cluster)
			}
			return nil
		})
		if err != nil {
			return &edgeproto.Result{}, err
		}
		defer func() {
			if err != nil {
				s.deleteClusterAuto(&in.Cluster)
			}
		}()
	}

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !developerApi.store.STMGet(stm, &in.Key.DeveloperKey, nil) {
			return errors.New("Specified developer not found")
		}
		if !flavorApi.store.STMGet(stm, &in.Flavor, nil) {
			return errors.New("Specified flavor not found")
		}
		if s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyExists
		}
		if !clusterApi.store.STMGet(stm, &in.Cluster, nil) {
			return errors.New("Specified Cluster not found")
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AppApi) UpdateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	return s.store.Update(in, s.sync.syncWait)
}

func (s *AppApi) DeleteApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	dynInsts := make(map[edgeproto.AppInstKey]struct{})
	if appInstApi.UsesApp(&in.Key, dynInsts) {
		// disallow delete if static instances are present
		return &edgeproto.Result{}, errors.New("Application in use by static Application Instance")
	}
	clusterKey := edgeproto.ClusterKey{}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		app := edgeproto.App{}
		if !s.store.STMGet(stm, in.GetKey(), &app) {
			// already deleted
			return nil
		}
		clusterKey = app.Cluster
		// delete app
		s.store.STMDel(stm, in.GetKey())
		return nil
	})
	if err == nil {
		// delete cluster afterwards if it was auto-created
		s.deleteClusterAuto(&clusterKey)
	}
	if err == nil && len(dynInsts) > 0 {
		// delete dynamic instances
		for key, _ := range dynInsts {
			appInst := edgeproto.AppInst{Key: key}
			_, derr := appInstApi.DeleteAppInst(ctx, &appInst)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic app inst",
					"err", derr)
			}
		}
	}
	return &edgeproto.Result{}, err
}

func (s *AppApi) deleteClusterAuto(key *edgeproto.ClusterKey) {
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		cluster := edgeproto.Cluster{}
		if clusterApi.store.STMGet(stm, key, &cluster) && cluster.Auto {
			clusterApi.store.STMDel(stm, key)
		}
		return nil
	})
	if err != nil {
		log.InfoLog("Failed to delete auto cluster",
			"clusterInst", key, "err", err)
	}
}

func (s *AppApi) ShowApp(in *edgeproto.App, cb edgeproto.AppApi_ShowAppServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.App) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AppApi) UpdatedCb(old *edgeproto.App, new *edgeproto.App) {
	if old == nil {
		return
	}
	if old.ImagePath != new.ImagePath || old.ImageType != new.ImageType ||
		old.ConfigMap != new.ConfigMap || !old.Flavor.Matches(&new.Flavor) {
		log.DebugLog(log.DebugLevelApi, "updating image path")
		appInstApi.cache.Mux.Lock()
		for _, inst := range appInstApi.cache.Objs {
			if inst.Key.AppKey.Matches(&new.Key) {
				inst.ImagePath = new.ImagePath
				inst.ImageType = new.ImageType
				inst.ConfigMap = new.ConfigMap
				inst.Flavor = new.Flavor
				// TODO: update mapped ports if needed
				if appInstApi.cache.NotifyCb != nil {
					appInstApi.cache.NotifyCb(inst.GetKey())
				}
			}
		}
		appInstApi.cache.Mux.Unlock()
	}
}
