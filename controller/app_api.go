// app config

package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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
	if !developerApi.HasDeveloper(&in.Key.DeveloperKey) {
		return &edgeproto.Result{}, errors.New("Specified developer not found")
	}
	if !flavorApi.HasFlavor(&in.Flavor) {
		return &edgeproto.Result{}, errors.New("Specified flavor not found")
	}
	if in.ImageType == edgeproto.ImageType_ImageTypeUnknown {
		return &edgeproto.Result{}, errors.New("Please specify Image Type")
	}
	if in.AccessLayer == edgeproto.AccessLayer_AccessLayerUnknown {
		// default to L7
		in.AccessLayer = edgeproto.AccessLayer_AccessLayerL7
	}
	if in.AccessPorts == "" && (in.AccessLayer == edgeproto.AccessLayer_AccessLayerL4 ||
		in.AccessLayer == edgeproto.AccessLayer_AccessLayerL4L7) {
		return &edgeproto.Result{}, errors.New("Please specify ports for L4 access types")
	}
	if in.ImageType == edgeproto.ImageType_ImageTypeDocker {
		in.ImagePath = "mobiledgex_" +
			util.DockerSanitize(in.Key.DeveloperKey.Name) + "/" +
			util.DockerSanitize(in.Key.Name) + ":" +
			util.DockerSanitize(in.Key.Version)
	}
	if in.Cluster.Name == "" {
		// cluster not specified, create one automatically
		in.Cluster.Name = fmt.Sprintf("%s%s", ClusterAutoPrefix,
			in.Key.Name)
		in.Cluster.Name = util.K8SSanitize(in.Cluster.Name)
		cluster := edgeproto.Cluster{}
		cluster.Key = in.Cluster
		cluster.Flavor = in.Flavor
		cluster.Nodes = ClusterAutoNodes
		// it may be possible that cluster already exists
		if !clusterApi.HasKey(&cluster.Key) {
			resp, err := clusterApi.createClusterInternal(&cluster)
			if err != nil {
				return resp, err
			}
			defer func() {
				if err != nil {
					clusterApi.deleteClusterInternal(&cluster)
				}
			}()
		}
	} else if !clusterApi.HasKey(&in.Cluster) {
		return &edgeproto.Result{}, errors.New("Specified Cluster not found")
	}
	resp, err := s.store.Create(in, s.sync.syncWait)
	return resp, err
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
	res, err := s.store.Delete(in, s.sync.syncWait)
	if len(dynInsts) > 0 {
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
	return res, err
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
