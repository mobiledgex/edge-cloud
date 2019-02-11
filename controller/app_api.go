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
		if app.DefaultFlavor.Matches(key) && app.DelOpt != edgeproto.DeleteType_AutoDelete {
			return true
		}
	}
	return false
}

func (s *AppApi) AutoDeleteApps(ctx context.Context, key *edgeproto.FlavorKey) {
	apps := make(map[edgeproto.AppKey]*edgeproto.App)
	log.DebugLog(log.DebugLevelApi, "Auto-deleting apps ", "flavor", key)
	s.cache.Mux.Lock()
	for k, app := range s.cache.Objs {
		if app.DefaultFlavor.Matches(key) && app.DelOpt == edgeproto.DeleteType_AutoDelete {
			apps[k] = app
		}
	}
	s.cache.Mux.Unlock()
	for _, val := range apps {
		log.DebugLog(log.DebugLevelApi, "Auto-deleting app ", "app", val.Key.Name)
		if _, err := s.DeleteApp(ctx, val); err != nil {
			log.DebugLog(log.DebugLevelApi, "unable to auto-delete app", "app", val, "err", err)
		}
	}
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

// updates fields that need manipulation on setting, or fetched remotely
func updateAppFields(in *edgeproto.App) error {

	if in.ImagePath == "" {
		if in.ImageType == edgeproto.ImageType_ImageTypeDocker {
			in.ImagePath = cloudcommon.Registry + ":5000/" +
				util.DockerSanitize(in.Key.DeveloperKey.Name) + "/" +
				util.DockerSanitize(in.Key.Name) + ":" +
				util.DockerSanitize(in.Key.Version)
		} else if in.Deployment == cloudcommon.AppDeploymentTypeHelm {
			in.ImagePath = cloudcommon.Registry + ":5000/" +
				util.DockerSanitize(in.Key.DeveloperKey.Name) + "/" +
				util.DockerSanitize(in.Key.Name)
		} else {
			in.ImagePath = "qcow path not determined yet"
		}
	}

	if in.Config != "" {
		configStr, err := cloudcommon.GetAppConfig(in)
		if err != nil {
			return err
		}
		in.Config = configStr
		// do a quick parse just to make sure it's valid
		_, err = cloudcommon.ParseAppConfig(in.Config)
		if err != nil {
			return err
		}
	}

	deploymf, err := cloudcommon.GetAppDeploymentManifest(in)
	if err != nil {
		return err
	}
	// Save manifest to app in case it was a remote target.
	// Manifest is required on app delete and we'll be in trouble
	// if remote target is unreachable or changed at that time.
	in.DeploymentManifest = deploymf
	return nil
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

	err = updateAppFields(in)
	if err != nil {
		return &edgeproto.Result{}, err
	}

	if s.AndroidPackageConflicts(in) {
		return &edgeproto.Result{}, fmt.Errorf("AndroidPackageName: %s in use by another app", in.AndroidPackageName)
	}

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !flavorApi.store.STMGet(stm, &in.DefaultFlavor, nil) {
			return errors.New("Specified default flavor not found")
		}
		if s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyExists
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
	err := updateAppFields(in)
	if err != nil {
		return &edgeproto.Result{}, err
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
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		app := edgeproto.App{}
		if !s.store.STMGet(stm, &in.Key, &app) {
			// already deleted
			return nil
		}
		// delete app
		s.store.STMDel(stm, &in.Key)
		return nil
	})
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
