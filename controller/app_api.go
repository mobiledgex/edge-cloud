// app config

package main

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
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
	appApi.cache.SetNotifyCb(notify.ServerMgrOne.UpdateApp)
	sync.RegisterCache(&appApi.cache)
}

func (s *AppApi) HasApp(key *edgeproto.AppKey) bool {
	return s.cache.HasKey(key)
}

func (s *AppApi) GetAllKeys(keys map[edgeproto.AppKey]struct{}) {
	s.cache.GetAllKeys(keys)
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
		if app.DefaultFlavor.Matches(key) {
			return true
		}
	}
	return false
}

func (s *AppApi) UsesCluster(key *edgeproto.ClusterKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, app := range s.cache.Objs {
		if app.Cluster.Matches(key) && app.DelFlags == edgeproto.DeleteType_NoAutoDelete {
			return true
		}
	}
	return false
}

func (s *AppApi) AutoDeleteApps(ctx context.Context, key *edgeproto.ClusterKey) error {
	apps := make(map[edgeproto.AppKey]*edgeproto.App)
	log.DebugLog(log.DebugLevelApi, "Auto-deleting appinsts ", "cluster", key.Name)
	s.cache.Mux.Lock()
	for k, app := range s.cache.Objs {
		if app.Cluster.Matches(key) && app.DelFlags == edgeproto.DeleteType_AutoDelete {
			apps[k] = app
		}
	}
	s.cache.Mux.Unlock()
	for _, val := range apps {
		log.DebugLog(log.DebugLevelApi, "Auto-deleting app ", "app", val.Key.Name)
		if _, err := s.DeleteApp(ctx, val); err != nil {
			return err
		}
	}
	return nil
}

// AndroidPackageConflicts returns true if an app with a different developer+name
// has the same package.  It is expect that different versions of the same app with
// the same package however so we do not do a full key comparison
func (s *AppApi) AndroidPackageConflicts(a *edgeproto.App) bool {
	if a.AndroidPackageName == "" {
		return false
	}
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, app := range s.cache.Objs {
		if app.AndroidPackageName == a.AndroidPackageName {
			if (a.Key.DeveloperKey.Name != app.Key.DeveloperKey.Name) || (a.Key.Name != app.Key.Name) {
				return true
			}
		}
	}
	return false
}

func (s *AppApi) CreateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	var err error

	if in.IpAccess == edgeproto.IpAccess_IpAccessUnknown {
		// default to shared
		in.IpAccess = edgeproto.IpAccess_IpAccessShared
	}

	if err = in.Validate(edgeproto.AppAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}
	if in.Deployment == "" {
		in.Deployment, err = cloudcommon.GetDefaultDeploymentType(in.ImageType)
		if err != nil {
			return &edgeproto.Result{}, err
		}
	}
	if !cloudcommon.IsValidDeploymentType(in.Deployment) {
		return &edgeproto.Result{}, fmt.Errorf("invalid deployment, must be one of %v", cloudcommon.ValidDeployments)
	}
	if !cloudcommon.IsValidDeploymentForImage(in.ImageType, in.Deployment) {
		return &edgeproto.Result{}, fmt.Errorf("deployment is not valid for image type")
	}
	if in.ImagePath == "" {
		if in.ImageType == edgeproto.ImageType_ImageTypeDocker {
			in.ImagePath = "mobiledgex_" +
				util.DockerSanitize(in.Key.DeveloperKey.Name) + "/" +
				util.DockerSanitize(in.Key.Name) + ":" +
				util.DockerSanitize(in.Key.Version)
		} else if in.Deployment == cloudcommon.AppDeploymentTypeHelm {
			in.ImagePath = "mobiledgex/" +
				util.DockerSanitize(in.Key.DeveloperKey.Name) + "/" +
				util.DockerSanitize(in.Key.Name)
		} else {
			in.ImagePath = "qcow path not determined yet"
		}
	}

	if in.Config != "" {
		configStr, err := cloudcommon.GetAppConfig(in)
		if err != nil {
			return &edgeproto.Result{}, err
		}
		in.Config = configStr
		// do a quick parse just to make sure it's valid
		_, err = cloudcommon.ParseAppConfig(in.Config)
		if err != nil {
			return &edgeproto.Result{}, err
		}
	}

	deploymf, err := cloudcommon.GetAppDeploymentManifest(in)
	if err != nil {
		return &edgeproto.Result{}, err
	}
	// Save manifest to app in case it was a remote target.
	// Manifest is required on app delete and we'll be in trouble
	// if remote target is unreachable or changed at that time.
	in.DeploymentManifest = deploymf

	if s.AndroidPackageConflicts(in) {
		return &edgeproto.Result{}, fmt.Errorf("AndroidPackageName: %s in use by another app", in.AndroidPackageName)
	}

	// make sure cluster exists
	// This is a separate STM to avoid ordering issues between
	// auto-cluster create and app create in watch cb.
	if in.Cluster.Name == "" {
		if in.DefaultFlavor.Name == "" {
			return &edgeproto.Result{}, errors.New("DefaultFlavor is required if Cluster is not specified")
		}
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
			clusterFlavorKey, lookup := GetClusterFlavorForFlavor(&in.DefaultFlavor)
			if lookup != nil {
				return lookup
			}
			cluster.DefaultFlavor = *clusterFlavorKey
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
				s.deleteClusterAuto(ctx, &in.Cluster)
			}
		}()
	}

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !developerApi.store.STMGet(stm, &in.Key.DeveloperKey, nil) {
			return errors.New("Specified developer not found")
		}
		if !flavorApi.store.STMGet(stm, &in.DefaultFlavor, nil) {
			return errors.New("Specified default flavor not found")
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
	if s.AndroidPackageConflicts(in) {
		return &edgeproto.Result{}, fmt.Errorf("AndroidPackageName: %s in use by another app", in.AndroidPackageName)
	}
	return s.store.Update(in, s.sync.syncWait)
}

func (s *AppApi) DeleteApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	if !appApi.HasApp(&in.Key) {
		// key doesn't exist
		return &edgeproto.Result{}, objstore.ErrKVStoreKeyNotFound
	}
	dynInsts := make(map[edgeproto.AppInstKey]struct{})
	if appInstApi.UsesApp(&in.Key, dynInsts) {
		// disallow delete if static instances are present
		return &edgeproto.Result{}, errors.New("Application in use by static Application Instance")
	}
	clusterKey := edgeproto.ClusterKey{}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		app := edgeproto.App{}
		if !s.store.STMGet(stm, &in.Key, &app) {
			// already deleted
			return nil
		}
		clusterKey = app.Cluster
		// delete app
		s.store.STMDel(stm, &in.Key)
		return nil
	})
	if err == nil {
		// delete cluster afterwards if it was auto-created
		s.deleteClusterAuto(ctx, &clusterKey)
	}
	if err == nil && len(dynInsts) > 0 {
		// delete dynamic instances
		for key, _ := range dynInsts {
			appInst := edgeproto.AppInst{Key: key}
			derr := appInstApi.DeleteAppInst(&appInst, nil)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic app inst",
					"err", derr)
			}
		}
	}
	return &edgeproto.Result{}, err
}

func (s *AppApi) deleteClusterAuto(ctx context.Context, key *edgeproto.ClusterKey) {
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		cluster := edgeproto.Cluster{}
		if clusterApi.store.STMGet(stm, key, &cluster) && cluster.Auto {
			if err := s.AutoDeleteApps(ctx, key); err != nil {
				return err
			}
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

// GetClusterFlavorForFlavor finds the smallest cluster flavor whose
// node flavor matches the passed in flavor.
func GetClusterFlavorForFlavor(flavorKey *edgeproto.FlavorKey) (*edgeproto.ClusterFlavorKey, error) {
	matching := make([]*edgeproto.ClusterFlavor, 0)

	clusterFlavorApi.cache.Mux.Lock()
	defer clusterFlavorApi.cache.Mux.Unlock()
	for _, val := range clusterFlavorApi.cache.Objs {
		if val.NodeFlavor.Matches(flavorKey) {
			matching = append(matching, val)
		}
	}
	if len(matching) == 0 {
		return nil, fmt.Errorf("No cluster flavors with node flavor %s found", flavorKey.Name)
	}
	sort.Slice(matching, func(i, j int) bool {
		if matching[i].MaxNodes != matching[j].MaxNodes {
			return matching[i].MaxNodes < matching[j].MaxNodes
		}
		if matching[i].NumNodes != matching[j].NumNodes {
			return matching[i].NumNodes < matching[j].NumNodes
		}
		if matching[i].NumMasters != matching[j].NumMasters {
			return matching[i].NumMasters < matching[j].NumMasters
		}
		return matching[i].Key.Name < matching[j].Key.Name
	})
	return &matching[0].Key, nil
}
