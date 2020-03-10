// app config

package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
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

func (s *AppApi) UsesAutoProvPolicy(key *edgeproto.PolicyKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, app := range s.cache.Objs {
		if app.Key.DeveloperKey.Name == key.Developer && app.AutoProvPolicy == key.Name {
			return true
		}
	}
	return false
}

func (s *AppApi) UsesPrivacyPolicy(key *edgeproto.PolicyKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, app := range s.cache.Objs {
		if app.Key.DeveloperKey.Name == key.Developer && app.DefaultPrivacyPolicy == key.Name {
			return true
		}
	}
	return false
}

func (s *AppApi) AutoDeleteAppsForDeveloper(ctx context.Context, key *edgeproto.DeveloperKey) {
	apps := make(map[edgeproto.AppKey]*edgeproto.App)
	log.DebugLog(log.DebugLevelApi, "Auto-deleting Apps ", "developer", key)
	s.cache.Mux.Lock()
	for k, app := range s.cache.Objs {
		if app.Key.DeveloperKey.Matches(key) && app.DelOpt == edgeproto.DeleteType_AUTO_DELETE {
			apps[k] = app
		}
	}
	s.cache.Mux.Unlock()
	for _, val := range apps {
		log.DebugLog(log.DebugLevelApi, "Auto-deleting App ", "app", val.Key.Name)
		if _, err := s.DeleteApp(ctx, val); err != nil {
			log.DebugLog(log.DebugLevelApi, "unable to auto-delete App", "app", val, "err", err)
		}
	}
}

func (s *AppApi) AutoDeleteApps(ctx context.Context, key *edgeproto.FlavorKey) {
	apps := make(map[edgeproto.AppKey]*edgeproto.App)
	log.DebugLog(log.DebugLevelApi, "Auto-deleting Apps ", "flavor", key)
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
			log.DebugLog(log.DebugLevelApi, "Unable to auto-delete App", "app", val, "err", err)
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

func validatePortRangeForAccessType(ports []dme.AppPort, accessType edgeproto.AccessType) error {
	maxPorts := settingsApi.Get().LoadBalancerMaxPortRange
	for ii, _ := range ports {
		ports[ii].PublicPort = ports[ii].InternalPort
		if ports[ii].EndPort != 0 {
			numPortsInRange := ports[ii].EndPort - ports[ii].PublicPort
			// this is checked in app_api also, but this in case there are pre-existing apps which violate this new restriction
			if accessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER && numPortsInRange > maxPorts {
				return fmt.Errorf("Port range greater than max of %d for load balanced application", maxPorts)
			}
		}
	}
	return nil
}

// updates fields that need manipulation on setting, or fetched remotely
func updateAppFields(ctx context.Context, in *edgeproto.App, revision int32) error {

	if in.ImagePath == "" {
		if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_DOCKER {
			if *registryFQDN == "" {
				return fmt.Errorf("No image path specified and no default registryFQDN to fall back upon. Please specify the image path")
			}
			in.ImagePath = *registryFQDN + "/" +
				util.DockerSanitize(in.Key.DeveloperKey.Name) + "/images/" +
				util.DockerSanitize(in.Key.Name) + ":" +
				util.DockerSanitize(in.Key.Version)
		} else if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW {
			if in.Md5Sum == "" {
				return fmt.Errorf("md5sum should be provided if imagepath is not specified")
			}
			if *artifactoryFQDN == "" {
				return fmt.Errorf("No image path specified and no default artifactoryFQDN to fall back upon. Please specify the image path")
			}
			in.ImagePath = *artifactoryFQDN + "repo-" +
				in.Key.DeveloperKey.Name + "/" +
				in.Key.Name + ".qcow2#md5:" + in.Md5Sum
		} else if in.Deployment == cloudcommon.AppDeploymentTypeHelm {
			if *registryFQDN == "" {
				return fmt.Errorf("No image path specified and no default registryFQDN to fall back upon. Please specify the image path")
			}
			in.ImagePath = *registryFQDN + "/" +
				util.DockerSanitize(in.Key.DeveloperKey.Name) + "/images/" +
				util.DockerSanitize(in.Key.Name)
		} else {
			in.ImagePath = "image path required"
		}
		log.DebugLog(log.DebugLevelApi, "derived imagepath", "imagepath", in.ImagePath)
	}
	if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_DOCKER {
		if strings.HasPrefix(in.ImagePath, "http://") {
			in.ImagePath = in.ImagePath[len("http://"):]
		}
		if strings.HasPrefix(in.ImagePath, "https://") {
			in.ImagePath = in.ImagePath[len("https://"):]
		}
	}

	if in.ScaleWithCluster && in.Deployment != cloudcommon.AppDeploymentTypeKubernetes {
		return fmt.Errorf("app scaling is only supported for Kubernetes deployments")
	}

	if !cloudcommon.IsPlatformApp(in.Key.DeveloperKey.Name, in.Key.Name) {
		if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_DOCKER {
			parts := strings.Split(in.ImagePath, "/")
			// Append default registry address for internal image paths
			if len(parts) < 2 || !strings.Contains(parts[0], ".") {
				in.ImagePath = cloudcommon.DockerHub + "/" + in.ImagePath
				log.SpanLog(ctx, log.DebugLevelApi, "Using default docker registry", "ImagePath", in.ImagePath)
			}
		}
	}

	if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW {
		err := util.ValidateImagePath(in.ImagePath)
		if err != nil {
			return err
		}
		err = cloudcommon.ValidateVMRegistryPath(ctx, in.ImagePath, vaultConfig)
		if err != nil {
			if *testMode {
				log.SpanLog(ctx, log.DebugLevelApi, "Warning, could not validate VM registry path.", "path", in.ImagePath, "err", err)
			} else {
				return fmt.Errorf("failed to validate VM registry image, path %s, %v", in.ImagePath, err)
			}
		}
	}

	if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_DOCKER &&
		!cloudcommon.IsPlatformApp(in.Key.DeveloperKey.Name, in.Key.Name) {
		err := cloudcommon.ValidateDockerRegistryPath(ctx, in.ImagePath, vaultConfig)
		if err != nil {
			if *testMode {
				log.SpanLog(ctx, log.DebugLevelApi, "Warning, could not validate docker registry image path", "path", in.ImagePath, "err", err)
			} else {
				return fmt.Errorf("failed to validate docker registry image, path %s, %v", in.ImagePath, err)
			}
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

	log.DebugLog(log.DebugLevelApi, "setting App revision", "revision", revision)
	in.Revision = revision

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
	} else if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_UNKNOWN {
		in.ImageType, err = cloudcommon.GetImageTypeForDeployment(in.Deployment)
		if err != nil {
			return &edgeproto.Result{}, err
		}
	}
	if !cloudcommon.IsValidDeploymentType(in.Deployment) {
		return &edgeproto.Result{}, fmt.Errorf("Invalid deployment, must be one of %v", cloudcommon.ValidDeployments)
	}
	if !cloudcommon.IsValidDeploymentForImage(in.ImageType, in.Deployment) {
		return &edgeproto.Result{}, fmt.Errorf("Deployment is not valid for image type")
	}
	newAccessType, err := cloudcommon.GetMappedAccessType(in.AccessType, in.Deployment, in.DeploymentManifest)
	if err != nil {
		return &edgeproto.Result{}, err
	}
	if in.AccessType != newAccessType {
		log.SpanLog(ctx, log.DebugLevelApi, "updating access type", "newAccessType", newAccessType)
		in.AccessType = newAccessType
	}

	if in.Deployment == cloudcommon.AppDeploymentTypeDocker && in.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
		dtype := cloudcommon.GetDockerDeployType(in.DeploymentManifest)
		if dtype != "docker" {
			// docker-compose manifests introduce a lot of complexity for LB solution because the port mappings will have
			// to change, and there may be multiple containers which communicate together over the host network.
			return &edgeproto.Result{}, fmt.Errorf("ACCESS_TYPE_LOAD_BALANCER not supported for docker deployment type: %s", dtype)
		}
	}

	if in.Deployment == cloudcommon.AppDeploymentTypeDocker || in.Deployment == cloudcommon.AppDeploymentTypeVM {
		if strings.Contains(strings.ToLower(in.AccessPorts), "http") {
			return &edgeproto.Result{}, fmt.Errorf("Deployment Type and HTTP access ports are incompatible")
		}
	}
	if in.Deployment == cloudcommon.AppDeploymentTypeVM && in.Command != "" {
		return &edgeproto.Result{}, fmt.Errorf("Invalid argument, command is not supported for VM based deployments")
	}
	err = updateAppFields(ctx, in, 0)
	if err != nil {
		return &edgeproto.Result{}, err
	}
	ports, err := edgeproto.ParseAppPorts(in.AccessPorts)
	if err != nil {
		return &edgeproto.Result{}, err
	}
	if in.DeploymentManifest != "" {
		err = cloudcommon.IsValidDeploymentManifest(in.Deployment, in.Command, in.DeploymentManifest, ports)
		if err != nil {
			return &edgeproto.Result{}, fmt.Errorf("Invalid deployment manifest, %v", err)
		}
	}
	err = validatePortRangeForAccessType(ports, in.AccessType)
	if err != nil {
		return nil, err
	}
	if s.AndroidPackageConflicts(in) {
		return &edgeproto.Result{}, fmt.Errorf("AndroidPackageName: %s in use by another App", in.AndroidPackageName)
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !flavorApi.store.STMGet(stm, &in.DefaultFlavor, nil) {
			return in.DefaultFlavor.NotFoundError()
		}
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		if in.AutoProvPolicy != "" {
			apKey := edgeproto.PolicyKey{}
			apKey.Developer = in.Key.DeveloperKey.Name
			apKey.Name = in.AutoProvPolicy
			if !autoProvPolicyApi.store.STMGet(stm, &apKey, nil) {
				return apKey.NotFoundError()
			}
		}
		if in.DefaultPrivacyPolicy != "" {
			apKey := edgeproto.PolicyKey{}
			apKey.Developer = in.Key.DeveloperKey.Name
			apKey.Name = in.DefaultPrivacyPolicy
			if !privacyPolicyApi.store.STMGet(stm, &apKey, nil) {
				return apKey.NotFoundError()
			}
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AppApi) UpdateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	if s.AndroidPackageConflicts(in) {
		return &edgeproto.Result{}, fmt.Errorf("AndroidPackageName: %s in use by another App", in.AndroidPackageName)
	}

	// if the app is already deployed, there are restrictions on what can be changed.
	dynInsts := make(map[edgeproto.AppInstKey]struct{})

	appInstExists := false
	for _, field := range in.Fields {
		if field == edgeproto.AppFieldDeployment {
			return &edgeproto.Result{}, fmt.Errorf("Field cannot be modified")
		}
		if appInstApi.UsesApp(&in.Key, dynInsts) {
			appInstExists = true
			if field == edgeproto.AppFieldAccessPorts {
				return &edgeproto.Result{}, fmt.Errorf("Field cannot be modified when AppInsts exist")
			}
		}
	}
	if err := in.Validate(edgeproto.MakeFieldMap(in.Fields)); err != nil {
		return &edgeproto.Result{}, err
	}

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.App{}

		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		if appInstExists {
			if cur.Deployment != cloudcommon.AppDeploymentTypeKubernetes &&
				cur.Deployment != cloudcommon.AppDeploymentTypeDocker &&
				cur.Deployment != cloudcommon.AppDeploymentTypeHelm {
				return fmt.Errorf("Update App not supported for deployment: %s when AppInsts exist", cur.Deployment)
			}
			// don't allow change from regular docker to docker-compose or docker-compose zip if instances exist
			if cur.Deployment == cloudcommon.AppDeploymentTypeDocker {
				curType := cloudcommon.GetDockerDeployType(cur.DeploymentManifest)
				newType := cloudcommon.GetDockerDeployType(in.DeploymentManifest)
				if curType != newType {
					return fmt.Errorf("Cannot change App manifest from : %s to: %s when AppInsts exist", curType, newType)
				}
			}
		}

		cur.CopyInFields(in)
		newRevision := cur.Revision + 1
		if err := updateAppFields(ctx, &cur, newRevision); err != nil {
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
		return &edgeproto.Result{}, in.Key.NotFoundError()
	}
	dynInsts := make(map[edgeproto.AppInstKey]struct{})
	if appInstApi.UsesApp(&in.Key, dynInsts) {
		// disallow delete if static instances are present
		return &edgeproto.Result{}, errors.New("Application in use by static AppInst")
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
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
			stream := streamoutAppInst{}
			stream.ctx = ctx
			stream.debugLvl = log.DebugLevelApi
			derr := appInstApi.DeleteAppInst(&appInst, &stream)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic AppInst",
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
