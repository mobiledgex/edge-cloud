// app config

package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
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

func (s *AppApi) UsesFlavor(key *edgeproto.FlavorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		app := data.Obj
		if app.DefaultFlavor.Matches(key) && app.DelOpt != edgeproto.DeleteType_AUTO_DELETE {
			return true
		}
	}
	return false
}

func (s *AppApi) UsesAutoProvPolicy(key *edgeproto.PolicyKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		app := data.Obj
		if app.Key.Organization == key.Organization {
			if app.AutoProvPolicy == key.Name {
				return true
			}
			for _, name := range app.AutoProvPolicies {
				if name == key.Name {
					return true
				}
			}
		}
	}
	return false
}

func (s *AppApi) UsesPrivacyPolicy(key *edgeproto.PolicyKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		app := data.Obj
		if app.Key.Organization == key.Organization && app.DefaultPrivacyPolicy == key.Name {
			return true
		}
	}
	return false
}

func (s *AppApi) AutoDeleteAppsForOrganization(ctx context.Context, org string) {
	apps := make(map[edgeproto.AppKey]*edgeproto.App)
	log.DebugLog(log.DebugLevelApi, "Auto-deleting Apps ", "org", org)
	s.cache.Mux.Lock()
	for k, data := range s.cache.Objs {
		app := data.Obj
		if app.Key.Organization == org && app.DelOpt == edgeproto.DeleteType_AUTO_DELETE {
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
	for k, data := range s.cache.Objs {
		app := data.Obj
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
	for _, data := range s.cache.Objs {
		app := data.Obj
		if app.AndroidPackageName == a.AndroidPackageName {
			if (a.Key.Organization != app.Key.Organization) || (a.Key.Name != app.Key.Name) {
				return true
			}
		}
	}
	return false
}

func validatePortRangeForAccessType(ports []dme.AppPort, accessType edgeproto.AccessType, deploymentType string) error {
	maxPorts := settingsApi.Get().LoadBalancerMaxPortRange
	for ii, _ := range ports {
		// dont allow tls on vms or docker with direct access
		if ports[ii].Tls && accessType == edgeproto.AccessType_ACCESS_TYPE_DIRECT &&
			(deploymentType == cloudcommon.DeploymentTypeVM || deploymentType == cloudcommon.DeploymentTypeDocker) {
			return fmt.Errorf("TLS unsupported on VM and docker deployments with direct access")
		}
		ports[ii].PublicPort = ports[ii].InternalPort
		if ports[ii].EndPort != 0 {
			numPortsInRange := ports[ii].EndPort - ports[ii].PublicPort + 1
			// this is checked in app_api also, but this in case there are pre-existing apps which violate this new restriction
			if accessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER && numPortsInRange > maxPorts {
				return fmt.Errorf("Port range greater than max of %d for load balanced application", maxPorts)
			}
		}
	}
	return nil
}

func validateSkipHcPorts(app *edgeproto.App) error {
	if app.SkipHcPorts == "" {
		return nil
	}
	if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_DIRECT {
		return fmt.Errorf("skipHcPorts not supported for type: %s", edgeproto.AccessType_ACCESS_TYPE_DIRECT)
	}
	if app.SkipHcPorts == "all" {
		return nil
	}
	ports, err := edgeproto.ParseAppPorts(app.AccessPorts)
	if err != nil {
		return err
	}
	skipHcPorts, err := edgeproto.ParseAppPorts(app.SkipHcPorts)
	if err != nil {
		return fmt.Errorf("Cannot parse skipHcPorts: %v", err)
	}
	for _, skipPort := range skipHcPorts {
		// for now we only support health checking for tcp ports
		if skipPort.Proto != dme.LProto_L_PROTO_TCP {
			return fmt.Errorf("Protocol %s unsupported for healthchecks", skipPort.Proto)
		}
		endSkip := skipPort.EndPort
		if endSkip == 0 {
			endSkip = skipPort.InternalPort
		}
		// break up skipHc port ranges into individual numbers
		for skipPortNum := skipPort.InternalPort; skipPortNum <= endSkip; skipPortNum++ {
			found := false
			for _, port := range ports {
				if port.Proto != skipPort.Proto {
					continue
				}
				endPort := port.EndPort
				if endPort == 0 {
					endPort = port.InternalPort
				}
				// for port ranges
				if port.InternalPort <= skipPortNum && skipPortNum <= endPort {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("skipHcPort %d not found in accessPorts", skipPortNum)
			}
		}
	}
	return nil
}

// updates fields that need manipulation on setting, or fetched remotely
func updateAppFields(ctx context.Context, in *edgeproto.App, revision string) error {
	// for update, trigger regenerating deployment manifest
	if in.DeploymentGenerator != "" {
		in.DeploymentManifest = ""
	}

	if in.ImagePath == "" {
		// ImagePath is required unless the image path is contained
		// within a DeploymentManifest specified by the user.
		// For updates where the controller generated a deployment
		// manifest, DeploymentManifest will be cleared because
		// DeploymentGenerator will have been set during create.
		if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_DOCKER && in.DeploymentManifest == "" {
			if *registryFQDN == "" {
				return fmt.Errorf("No image path specified and no default registryFQDN to fall back upon. Please specify the image path")
			}
			in.ImagePath = *registryFQDN + "/" +
				util.DockerSanitize(in.Key.Organization) + "/images/" +
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
				in.Key.Organization + "/" +
				in.Key.Name + ".qcow2#md5:" + in.Md5Sum
		} else if in.Deployment == cloudcommon.DeploymentTypeHelm {
			if *registryFQDN == "" {
				return fmt.Errorf("No image path specified and no default registryFQDN to fall back upon. Please specify the image path")
			}
			in.ImagePath = *registryFQDN + "/" +
				util.DockerSanitize(in.Key.Organization) + "/images/" +
				util.DockerSanitize(in.Key.Name)
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

	if in.ScaleWithCluster && in.Deployment != cloudcommon.DeploymentTypeKubernetes {
		return fmt.Errorf("app scaling is only supported for Kubernetes deployments")
	}

	if !cloudcommon.IsPlatformApp(in.Key.Organization, in.Key.Name) {
		if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_DOCKER && in.ImagePath != "" {
			parts := strings.Split(in.ImagePath, "/")
			if parts[0] == "localhost" {
				in.ImagePath = strings.Replace(in.ImagePath, "localhost/", "", -1)
			} else {
				// Append default registry address for internal image paths
				if len(parts) < 2 || !strings.Contains(parts[0], ".") {
					in.ImagePath = cloudcommon.DockerHub + "/" + in.ImagePath
					log.SpanLog(ctx, log.DebugLevelApi, "Using default docker registry", "ImagePath", in.ImagePath)
				}
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
		!cloudcommon.IsPlatformApp(in.Key.Organization, in.Key.Name) &&
		in.ImagePath != "" {
		err := cloudcommon.ValidateDockerRegistryPath(ctx, in.ImagePath, vaultConfig)
		if err != nil {
			if *testMode {
				log.SpanLog(ctx, log.DebugLevelApi, "Warning, could not validate docker registry image path", "path", in.ImagePath, "err", err)
			} else {
				return fmt.Errorf("failed to validate docker registry image, path %s, %v", in.ImagePath, err)
			}
		}
	}
	deploymf, err := cloudcommon.GetAppDeploymentManifest(ctx, vaultConfig, in)
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
	if !cloudcommon.IsValidDeploymentType(in.Deployment, cloudcommon.ValidAppDeployments) {
		return &edgeproto.Result{}, fmt.Errorf("Invalid deployment, must be one of %v", cloudcommon.ValidAppDeployments)
	}
	if !cloudcommon.IsValidDeploymentForImage(in.ImageType, in.Deployment) {
		return &edgeproto.Result{}, fmt.Errorf("Deployment is not valid for image type")
	}
	if err := validateAppConfigsForDeployment(in.Configs, in.Deployment); err != nil {
		return &edgeproto.Result{}, err
	}
	if in.Deployment == cloudcommon.DeploymentTypeKubernetes {
		_, err := k8smgmt.GetAppEnvVars(ctx, in, vaultConfig, &k8smgmt.TestReplacementVars)
		if err != nil {
			return &edgeproto.Result{}, err
		}
	}
	newAccessType, err := cloudcommon.GetMappedAccessType(in.AccessType, in.Deployment, in.DeploymentManifest)
	if err != nil {
		return &edgeproto.Result{}, err
	}
	if in.AccessType != newAccessType {
		log.SpanLog(ctx, log.DebugLevelApi, "updating access type", "newAccessType", newAccessType)
		in.AccessType = newAccessType
	}

	if in.Deployment == cloudcommon.DeploymentTypeVM && in.Command != "" {
		return &edgeproto.Result{}, fmt.Errorf("Invalid argument, command is not supported for VM based deployments")
	}
	err = updateAppFields(ctx, in, in.Revision)
	if err != nil {
		return &edgeproto.Result{}, err
	}
	ports, err := edgeproto.ParseAppPorts(in.AccessPorts)
	if err != nil {
		return &edgeproto.Result{}, err
	}

	err = validateSkipHcPorts(in)
	if err != nil {
		return &edgeproto.Result{}, err
	}

	if in.DeploymentManifest != "" {
		err = cloudcommon.IsValidDeploymentManifest(in.Deployment, in.Command, in.DeploymentManifest, ports)
		if err != nil {
			return &edgeproto.Result{}, fmt.Errorf("Invalid deployment manifest, %v", err)
		}
	}
	err = validatePortRangeForAccessType(ports, in.AccessType, in.Deployment)
	if err != nil {
		return nil, err
	}
	if s.AndroidPackageConflicts(in) {
		return &edgeproto.Result{}, fmt.Errorf("AndroidPackageName: %s in use by another App", in.AndroidPackageName)
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if in.DefaultFlavor.Name != "" && !flavorApi.store.STMGet(stm, &in.DefaultFlavor, nil) {
			return in.DefaultFlavor.NotFoundError()
		}
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		if err := s.validatePolicies(stm, in); err != nil {
			return err
		}
		if in.DefaultPrivacyPolicy != "" {
			apKey := edgeproto.PolicyKey{}
			apKey.Organization = in.Key.Organization
			apKey.Name = in.DefaultPrivacyPolicy
			if !privacyPolicyApi.store.STMGet(stm, &apKey, nil) {
				return apKey.NotFoundError()
			}
		}
		appInstRefsApi.createRef(stm, &in.Key)
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AppApi) UpdateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	err := in.ValidateUpdateFields()
	if err != nil {
		return &edgeproto.Result{}, err
	}
	if s.AndroidPackageConflicts(in) {
		return &edgeproto.Result{}, fmt.Errorf("AndroidPackageName: %s in use by another App", in.AndroidPackageName)
	}

	// if the app is already deployed, there are restrictions on what can be changed.
	dynInsts := make(map[edgeproto.AppInstKey]struct{})

	appInstExists := false
	for _, field := range in.Fields {
		if appInstApi.UsesApp(&in.Key, dynInsts) {
			appInstExists = true
			if field == edgeproto.AppFieldAccessPorts || field == edgeproto.AppFieldSkipHcPorts {
				return &edgeproto.Result{}, fmt.Errorf("Field cannot be modified when AppInsts exist")
			}
		}
	}
	fields := edgeproto.MakeFieldMap(in.Fields)
	if err := in.Validate(fields); err != nil {
		return &edgeproto.Result{}, err
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.App{}

		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		if appInstExists {
			if cur.Deployment != cloudcommon.DeploymentTypeKubernetes &&
				cur.Deployment != cloudcommon.DeploymentTypeDocker &&
				cur.Deployment != cloudcommon.DeploymentTypeHelm {
				return fmt.Errorf("Update App not supported for deployment: %s when AppInsts exist", cur.Deployment)
			}
			// don't allow change from regular docker to docker-compose or docker-compose zip if instances exist
			if cur.Deployment == cloudcommon.DeploymentTypeDocker {
				curType := cloudcommon.GetDockerDeployType(cur.DeploymentManifest)
				newType := cloudcommon.GetDockerDeployType(in.DeploymentManifest)
				if curType != newType {
					return fmt.Errorf("Cannot change App manifest from : %s to: %s when AppInsts exist", curType, newType)
				}
			}
		}
		// If config is being updated, make sure that it's valid for the DeploymentType
		if _, found := fields[edgeproto.AppFieldConfigs]; found {
			if err := validateAppConfigsForDeployment(in.Configs, cur.Deployment); err != nil {
				return err
			}
		}
		if in.DeploymentManifest != "" {
			// reset the deployment generator
			cur.DeploymentGenerator = ""
		} else if in.AccessPorts != "" {
			if cur.DeploymentGenerator == "" && cur.Deployment == cloudcommon.DeploymentTypeKubernetes {
				// No generator means the user previously provided a manifest.  Force them to do so again when changing ports so
				// that they do not accidentally lose their provided manifest details
				return fmt.Errorf("kubernetes manifest which was previously specified must be provided again when changing access ports")
			} else if cur.Deployment == cloudcommon.DeploymentTypeDocker {
				// there's no way to tell if we generated this manifest or it was provided, so disallow this change
				return fmt.Errorf("Changing access ports on docker apps not allowed unless manifest is specified")
			}
			// force regeneration of manifest
			cur.DeploymentManifest = ""
		}
		cur.CopyInFields(in)
		if err := s.validatePolicies(stm, &cur); err != nil {
			return err
		}
		ports, err := edgeproto.ParseAppPorts(cur.AccessPorts)
		if err != nil {
			return err
		}
		newRevision := in.Revision
		if newRevision == "" {
			newRevision = time.Now().Format("2006-01-02T150405")
			log.SpanLog(ctx, log.DebugLevelApi, "Setting new revision to current timestamp", "Revision", in.Revision)
		}
		if err := updateAppFields(ctx, &cur, newRevision); err != nil {
			return err
		}
		err = validateSkipHcPorts(&cur)
		if err != nil {
			return err
		}
		if in.DeploymentManifest != "" {
			err = cloudcommon.IsValidDeploymentManifest(cur.Deployment, cur.Command, cur.DeploymentManifest, ports)
			if err != nil {
				return fmt.Errorf("Invalid deployment manifest, %v", err)
			}
		}
		if cur.Deployment == cloudcommon.DeploymentTypeKubernetes {
			_, err := k8smgmt.GetAppEnvVars(ctx, &cur, vaultConfig, &k8smgmt.TestReplacementVars)
			if err != nil {
				return err
			}
		}
		err = validatePortRangeForAccessType(ports, cur.AccessType, cur.Deployment)
		if err != nil {
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

	// set state to prevent new AppInsts from being created from this App
	var dynInsts map[edgeproto.AppInstKey]*edgeproto.AppInst
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return in.Key.NotFoundError()
		}
		// use refs to check existing AppInsts to avoid race conditions
		dynInsts = make(map[edgeproto.AppInstKey]*edgeproto.AppInst)
		refs := edgeproto.AppInstRefs{}
		appInstRefsApi.store.STMGet(stm, &in.Key, &refs)
		for k, _ := range refs.Insts {
			// disallow delete if static instances are present
			inst := edgeproto.AppInst{}
			edgeproto.AppInstKeyStringParse(k, &inst.Key)
			if !appInstApi.store.STMGet(stm, &inst.Key, &inst) {
				// no inst?
				log.SpanLog(ctx, log.DebugLevelApi, "AppInst not found by refs, skipping for delete", "AppInst", inst.Key)
				continue
			}
			if inst.Liveness == edgeproto.Liveness_LIVENESS_STATIC {
				return errors.New("Application in use by static AppInst")
			}
			dynInsts[inst.Key] = &inst
		}

		in.DeletePrepare = true
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}

	// delete auto-appinsts
	log.SpanLog(ctx, log.DebugLevelApi, "Auto-deleting AppInsts for App", "app", in.Key)
	if err = appInstApi.AutoDelete(ctx, dynInsts); err != nil {
		// failed, so remove delete prepare and don't delete
		unseterr := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			if !s.store.STMGet(stm, &in.Key, in) {
				return in.Key.NotFoundError()
			}
			in.DeletePrepare = false
			s.store.STMPut(stm, in)
			return nil
		})
		if unseterr != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Delete App unset delete prepare", "unseterr", unseterr)
		}
		return &edgeproto.Result{}, err
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		app := edgeproto.App{}
		if !s.store.STMGet(stm, &in.Key, &app) {
			// already deleted
			return nil
		}
		// delete app
		s.store.STMDel(stm, &in.Key)
		// delete refs
		appInstRefsApi.deleteRef(stm, &in.Key)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AppApi) ShowApp(in *edgeproto.App, cb edgeproto.AppApi_ShowAppServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.App) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AppApi) AddAppAutoProvPolicy(ctx context.Context, in *edgeproto.AppAutoProvPolicy) (*edgeproto.Result, error) {
	cur := edgeproto.App{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.AppKey, &cur) {
			return in.AppKey.NotFoundError()
		}
		for _, name := range cur.AutoProvPolicies {
			if name == in.AutoProvPolicy {
				return fmt.Errorf("AutoProvPolicy %s already on App", name)
			}
		}
		cur.AutoProvPolicies = append(cur.AutoProvPolicies, in.AutoProvPolicy)
		if err := s.validatePolicies(stm, &cur); err != nil {
			return err
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AppApi) RemoveAppAutoProvPolicy(ctx context.Context, in *edgeproto.AppAutoProvPolicy) (*edgeproto.Result, error) {
	cur := edgeproto.App{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.AppKey, &cur) {
			return in.AppKey.NotFoundError()
		}
		changed := false
		for ii, name := range cur.AutoProvPolicies {
			if name != in.AutoProvPolicy {
				continue
			}
			cur.AutoProvPolicies = append(cur.AutoProvPolicies[:ii], cur.AutoProvPolicies[ii+1:]...)
			changed = true
			break
		}
		if cur.AutoProvPolicy == in.AutoProvPolicy {
			cur.AutoProvPolicy = ""
			changed = true
		}
		if !changed {
			return nil
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func validateAppConfigsForDeployment(configs []*edgeproto.ConfigFile, deployment string) error {
	for _, cfg := range configs {
		if cfg.Kind == edgeproto.AppConfigHelmYaml && deployment != cloudcommon.DeploymentTypeHelm {
			return fmt.Errorf("Invalid Config Kind(%s) for deployment type(%s)", cfg.Kind, deployment)
		}
	}
	return nil
}

func (s *AppApi) validatePolicies(stm concurrency.STM, app *edgeproto.App) error {
	// make sure policies exist
	for name, _ := range app.GetAutoProvPolicies() {
		policyKey := edgeproto.PolicyKey{}
		policyKey.Organization = app.Key.Organization
		policyKey.Name = name
		policy := edgeproto.AutoProvPolicy{}
		if !autoProvPolicyApi.store.STMGet(stm, &policyKey, &policy) {
			return policyKey.NotFoundError()
		}
	}
	return nil
}
