// app config

package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

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
	for key, app := range s.cache.Objs {
		if key.DeveloperKey.Matches(in) && app.DelOpt != edgeproto.DeleteType_AUTO_DELETE {
			return true
		}
	}
	return false
}

func (s *AppApi) UsesFlavor(key *edgeproto.FlavorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, app := range s.cache.Objs {
		if app.DefaultFlavor.Matches(key) && app.DelOpt != edgeproto.DeleteType_AUTO_DELETE {
			return true
		}
	}
	return false
}

func (s *AppApi) AutoDeleteAppsForDeveloper(ctx context.Context, key *edgeproto.DeveloperKey) {
	apps := make(map[edgeproto.AppKey]*edgeproto.App)
	log.DebugLog(log.DebugLevelApi, "Auto-deleting apps ", "developer", key)
	s.cache.Mux.Lock()
	for k, app := range s.cache.Objs {
		if app.Key.DeveloperKey.Matches(key) && app.DelOpt == edgeproto.DeleteType_AUTO_DELETE {
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

func (s *AppApi) AutoDeleteApps(ctx context.Context, key *edgeproto.FlavorKey) {
	apps := make(map[edgeproto.AppKey]*edgeproto.App)
	log.DebugLog(log.DebugLevelApi, "Auto-deleting apps ", "flavor", key)
	s.cache.Mux.Lock()
	for k, app := range s.cache.Objs {
		if app.DefaultFlavor.Matches(key) && app.DelOpt == edgeproto.DeleteType_AUTO_DELETE {
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
		if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_DOCKER {
			in.ImagePath = *registryFQDN + "/" +
				util.DockerSanitize(in.Key.DeveloperKey.Name) + "/images/" +
				util.DockerSanitize(in.Key.Name) + ":" +
				util.DockerSanitize(in.Key.Version)
		} else if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW {
			return fmt.Errorf("imagepath is required for imagetype %s", in.ImageType)
		} else if in.Deployment == cloudcommon.AppDeploymentTypeHelm {
			in.ImagePath = *registryFQDN + "/" +
				util.DockerSanitize(in.Key.DeveloperKey.Name) + "/images/" +
				util.DockerSanitize(in.Key.Name)
		} else {
			in.ImagePath = "qcow path not determined yet"
		}
	}

	if !cloudcommon.IsPlatformApp(in.Key.DeveloperKey.Name, in.Key.Name) {
		if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_DOCKER {
			parts := strings.Split(in.ImagePath, "/")
			// Append default registry address for internal image paths
			if len(parts) < 2 || !strings.Contains(parts[0], ".") {
				return fmt.Errorf("imagepath should be full registry URL: <domain-name>/<registry-path>")
			}
			if !*testMode {
				err := cloudcommon.ValidateDockerRegistryPath(in.ImagePath, *vaultAddr)
				if err != nil {
					return err
				}
			}
		}
	}

	if in.ScaleWithCluster && in.Deployment != cloudcommon.AppDeploymentTypeKubernetes {
		return fmt.Errorf("app scaling is only supported for Kubernetes deployments")
	}

	if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW {
		if !*testMode {
			err := cloudcommon.ValidateVMRegistryPath(in.ImagePath, *vaultAddr)
			if err != nil {
				return err
			}
		}
		urlInfo := strings.Split(in.ImagePath, "#")
		if len(urlInfo) != 2 {
			return fmt.Errorf("md5 checksum of image is required. Please append checksum to imagepath: \"<url>#md5:checksum\"")
		}
		cSum := strings.Split(urlInfo[1], ":")
		if len(cSum) != 2 {
			return fmt.Errorf("incorrect checksum format, valid format: \"<url>#md5:checksum\"")
		}
		if cSum[0] != "md5" {
			return fmt.Errorf("only md5 checksum is supported")
		}
		if len(cSum[1]) < 32 {
			return fmt.Errorf("md5 checksum must be at least 32 characters")
		}
		_, err := hex.DecodeString(cSum[1])
		if err != nil {
			return fmt.Errorf("invalid md5 checksum")
		}
	}

	// for update, trigger regenerating deployment manifest
	if in.DeploymentGenerator != "" {
		in.DeploymentManifest = ""
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
	if in.Deployment == cloudcommon.AppDeploymentTypeDocker {
		if in.AccessPorts != "" {
			if strings.Contains(in.AccessPorts, "http") {
				return &edgeproto.Result{}, fmt.Errorf("Deployment Type Docker and HTTP access ports are incompatable")
			}
		}
	}
	err = updateAppFields(in)
	if err != nil {
		return &edgeproto.Result{}, err
	}
	if in.DeploymentManifest != "" {
		err = cloudcommon.IsValidDeploymentManifest(in.Deployment, in.Command, in.DeploymentManifest)
		if err != nil {
			return &edgeproto.Result{}, fmt.Errorf("invalid deploymentment manifest %v", err)
		}
	}

	if s.AndroidPackageConflicts(in) {
		return &edgeproto.Result{}, fmt.Errorf("AndroidPackageName: %s in use by another app", in.AndroidPackageName)
	}

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !flavorApi.store.STMGet(stm, &in.DefaultFlavor, nil) {
			return edgeproto.ErrEdgeApiFlavorNotFound
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
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		cur := edgeproto.App{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return objstore.ErrKVStoreKeyNotFound
		}
		cur.CopyInFields(in)
		if err := updateAppFields(&cur); err != nil {
			return err
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
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
