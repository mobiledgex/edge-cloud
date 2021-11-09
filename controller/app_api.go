//app config

package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/deploygen"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	appsv1 "k8s.io/api/apps/v1"
)

// Should only be one of these instantiated in main
type AppApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.AppStore
	cache edgeproto.AppCache
}

func NewAppApi(sync *Sync, all *AllApis) *AppApi {
	appApi := AppApi{}
	appApi.all = all
	appApi.sync = sync
	appApi.store = edgeproto.NewAppStore(sync.store)
	edgeproto.InitAppCache(&appApi.cache)
	sync.RegisterCache(&appApi.cache)
	return &appApi
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

func (s *AppApi) GetAllApps(apps map[edgeproto.AppKey]*edgeproto.App) {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		app := data.Obj
		apps[app.Key] = app
	}
}

func (s *AppApi) validateAppExists(key *edgeproto.AppKey) bool {
	return s.HasApp(key)
}

func (s *AppApi) CheckAppCompatibleWithTrustPolicy(ctx context.Context, ckey *edgeproto.CloudletKey, app *edgeproto.App, trustPolicy *edgeproto.TrustPolicy) error {
	if !app.Trusted {
		return fmt.Errorf("Non trusted app: %s not compatible with trust policy: %s", strings.TrimSpace(app.Key.String()), trustPolicy.Key.String())
	}

	allowedRules := []*edgeproto.SecurityRule{}
	list, err := s.all.cloudletPoolApi.GetCloudletPoolKeysForCloudletKey(ckey)
	if err == nil {
		for _, cloudletPoolKey := range list {
			rules := s.all.trustPolicyExceptionApi.GetTrustPolicyExceptionRules(&cloudletPoolKey, &app.Key)
			log.SpanLog(ctx, log.DebugLevelApi, "CheckAppCompatibleWithTrustPolicy() GetTrustPolicyExceptionRules returned", "rules", rules)

			allowedRules = append(allowedRules, rules...)
		}
	} else {
		log.SpanLog(ctx, log.DebugLevelApi, "CheckAppCompatibleWithTrustPolicy() returned", "err", err)
	}
	// Combine the trustPolicy rules with the trustPolicyException rules.
	for i, r := range trustPolicy.OutboundSecurityRules {
		log.SpanLog(ctx, log.DebugLevelApi, "CheckAppCompatibleWithTrustPolicy()  trustPolicy:", "rule", r)
		allowedRules = append(allowedRules, &trustPolicy.OutboundSecurityRules[i])
	}
	for _, r := range app.RequiredOutboundConnections {
		log.SpanLog(ctx, log.DebugLevelApi, "CheckAppCompatibleWithTrustPolicy()  Checking for app:", "rule", r)
		policyMatchFound := false
		_, appNet, err := net.ParseCIDR(r.RemoteCidr)
		if err != nil {
			return fmt.Errorf("Invalid remote CIDR in RequiredOutboundConnections: %s - %v", r.RemoteCidr, err)
		}
		for _, outboundRule := range allowedRules {
			if strings.ToLower(r.Protocol) != strings.ToLower(outboundRule.Protocol) {
				continue
			}
			_, trustPolNet, err := net.ParseCIDR(outboundRule.RemoteCidr)
			if err != nil {
				return fmt.Errorf("Invalid remote CIDR in policy: %s - %v", outboundRule.RemoteCidr, err)
			}
			if !cloudcommon.CidrContainsCidr(trustPolNet, appNet) {
				continue
			}
			if strings.ToLower(r.Protocol) != "icmp" {
				if r.PortRangeMin < outboundRule.PortRangeMin || r.PortRangeMax > outboundRule.PortRangeMax {
					continue
				}
			}
			policyMatchFound = true
			break
		}
		if !policyMatchFound {
			return fmt.Errorf("No outbound rule in policy or exception to match required connection %s:%s:%d-%d for App %s", r.Protocol, r.RemoteCidr, r.PortRangeMin, r.PortRangeMax, app.Key.GetKeyString())
		}
	}
	return nil
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

func validateSkipHcPorts(app *edgeproto.App) error {
	if app.SkipHcPorts == "" {
		return nil
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

// Configure and validate App. Common code for Create and Update.
func (s *AppApi) configureApp(ctx context.Context, stm concurrency.STM, in *edgeproto.App, revision string) error {
	log.SpanLog(ctx, log.DebugLevelApi, "configureApp", "app", in.Key.String())
	var err error
	if s.AndroidPackageConflicts(in) {
		return fmt.Errorf("AndroidPackageName: %s in use by another App", in.AndroidPackageName)
	}
	if in.Deployment == "" {
		in.Deployment, err = cloudcommon.GetDefaultDeploymentType(in.ImageType)
		if err != nil {
			return err
		}
	} else if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_UNKNOWN {
		in.ImageType, err = cloudcommon.GetImageTypeForDeployment(in.Deployment)
		if err != nil {
			return err
		}
	}
	if in.Deployment == cloudcommon.DeploymentTypeHelm && in.DeploymentManifest != "" {
		return fmt.Errorf("Manifest is not used for Helm deployments. Use config files for customizations")
	}
	if !cloudcommon.IsValidDeploymentType(in.Deployment, cloudcommon.ValidAppDeployments) {
		return fmt.Errorf("Invalid deployment, must be one of %v", cloudcommon.ValidAppDeployments)
	}
	if !cloudcommon.IsValidDeploymentForImage(in.ImageType, in.Deployment) {
		return fmt.Errorf("Deployment is not valid for image type")
	}
	if err := validateAppConfigsForDeployment(ctx, in.Configs, in.Deployment); err != nil {
		return err
	}
	for i, r := range in.RequiredOutboundConnections {
		if r.PortRangeMax == 0 {
			in.RequiredOutboundConnections[i].PortRangeMax = r.PortRangeMin
		}
	}
	if err := validateRequiredOutboundConnections(in.RequiredOutboundConnections); err != nil {
		return err
	}
	newAccessType, err := cloudcommon.GetMappedAccessType(in.AccessType, in.Deployment, in.DeploymentManifest)
	if err != nil {
		return err
	}
	if in.AccessType != newAccessType {
		log.SpanLog(ctx, log.DebugLevelApi, "updating access type", "newAccessType", newAccessType)
		in.AccessType = newAccessType
	}

	if in.Deployment == cloudcommon.DeploymentTypeVM && in.Command != "" {
		return fmt.Errorf("Invalid argument, command is not supported for VM based deployments")
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
	if in.VmAppOsType != edgeproto.VmAppOsType_VM_APP_OS_UNKNOWN && in.Deployment != cloudcommon.DeploymentTypeVM {
		return fmt.Errorf("VM App OS Type is only supported for VM deployments")
	}

	if !cloudcommon.IsPlatformApp(in.Key.Organization, in.Key.Name) {
		if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_DOCKER && in.ImagePath != "" {
			in.ImagePath = k8smgmt.FixImagePath(in.ImagePath)
		}
	}

	authApi := &cloudcommon.VaultRegistryAuthApi{
		VaultConfig: vaultConfig,
	}
	if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_QCOW {
		if !strings.Contains(in.ImagePath, "://") {
			in.ImagePath = "https://" + in.ImagePath
		}
		err := util.ValidateImagePath(in.ImagePath)
		if err != nil {
			return err
		}
		err = cloudcommon.ValidateVMRegistryPath(ctx, in.ImagePath, authApi)
		if err != nil {
			if *testMode {
				log.SpanLog(ctx, log.DebugLevelApi, "Warning, could not validate VM registry path.", "path", in.ImagePath, "err", err)
			} else {
				return fmt.Errorf("failed to validate VM registry image, path %s, %v", in.ImagePath, err)
			}
		}
	}
	if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_OVF {
		if !strings.Contains(in.ImagePath, "://") {
			in.ImagePath = "https://" + in.ImagePath
		}
		// need to check for both an OVF and the corresponding VMDK
		if !strings.Contains(in.ImagePath, ".ovf") {
			return fmt.Errorf("image path does not specify an OVF file %s, %v", in.ImagePath, err)
		}
		err = cloudcommon.ValidateOvfRegistryPath(ctx, in.ImagePath, authApi)
		if err != nil {
			if *testMode {
				log.SpanLog(ctx, log.DebugLevelApi, "Warning, could not validate ovf file path.", "path", in.ImagePath, "err", err)
			} else {
				return fmt.Errorf("failed to validate ovf file path, path %s, %v", in.ImagePath, err)
			}
		}
	}

	if in.ImageType == edgeproto.ImageType_IMAGE_TYPE_DOCKER &&
		!cloudcommon.IsPlatformApp(in.Key.Organization, in.Key.Name) &&
		in.ImagePath != "" {
		err := cloudcommon.ValidateDockerRegistryPath(ctx, in.ImagePath, authApi)
		if err != nil {
			if *testMode {
				log.SpanLog(ctx, log.DebugLevelApi, "Warning, could not validate docker registry image path", "path", in.ImagePath, "err", err)
			} else {
				return fmt.Errorf("failed to validate docker registry image, path %s, %v", in.ImagePath, err)
			}
		}
	}
	deploymf, err := cloudcommon.GetAppDeploymentManifest(ctx, authApi, in)
	if err != nil {
		return err
	}
	if in.ScaleWithCluster && !manifestContainsDaemonSet(deploymf) {
		return fmt.Errorf("DaemonSet required in manifest when ScaleWithCluster set")
	}

	flavor := &edgeproto.Flavor{}
	if in.DefaultFlavor.Name == "" {
		return fmt.Errorf("Default flavor must be specified")
	}
	if !s.all.flavorApi.store.STMGet(stm, &in.DefaultFlavor, flavor) {
		return in.DefaultFlavor.NotFoundError()
	}

	if in.AllowServerless {
		if in.Deployment != cloudcommon.DeploymentTypeKubernetes {
			return fmt.Errorf("Allow serverless only supported for deployment type Kubernetes")
		}
		if in.ScaleWithCluster {
			return fmt.Errorf("Allow serverless does not support scale with cluster")
		}
		if in.DeploymentGenerator != deploygen.KubernetesBasic {
			// For now, only allow system generated manifests.
			// In order to support user-supplied manifests, we will
			// need to parse the manifest and look for any objects
			// that cannot be segregated by namespace.
			return fmt.Errorf("Allow serverless only allowed for system generated manifests")
		}
		if in.ServerlessConfig == nil {
			in.ServerlessConfig = &edgeproto.ServerlessConfig{}
		}
		if in.ServerlessConfig.Vcpus.IsZero() {
			in.ServerlessConfig.Vcpus.Set(flavor.Vcpus, 0)
		}
		if in.ServerlessConfig.Ram == 0 {
			in.ServerlessConfig.Ram = flavor.Ram
		}
		if in.ServerlessConfig.Vcpus.Nanos%(edgeproto.DecMillis) != 0 {
			return fmt.Errorf("Serverless config vcpus cannot have precision less than 0.001")
		}
		if in.ServerlessConfig.MinReplicas == 0 {
			in.ServerlessConfig.MinReplicas = 1
		}
	} else {
		if in.ServerlessConfig != nil {
			return fmt.Errorf("Serverless config cannot be specified without allow serverless true")
		}
	}

	// Save manifest to app in case it was a remote target.
	// Manifest is required on app delete and we'll be in trouble
	// if remote target is unreachable or changed at that time.
	in.DeploymentManifest = deploymf

	log.DebugLog(log.DebugLevelApi, "setting App revision", "revision", revision)
	in.Revision = revision

	err = validateSkipHcPorts(in)
	if err != nil {
		return err
	}
	ports, err := edgeproto.ParseAppPorts(in.AccessPorts)
	if err != nil {
		return err
	}

	if in.DeploymentManifest != "" {
		err = cloudcommon.IsValidDeploymentManifest(in.Deployment, in.Command, in.DeploymentManifest, ports, flavor)
		if err != nil {
			return fmt.Errorf("Invalid deployment manifest, %v", err)
		}
	}

	if in.Deployment == cloudcommon.DeploymentTypeKubernetes {
		authApi := &cloudcommon.VaultRegistryAuthApi{
			VaultConfig: vaultConfig,
		}
		_, err = k8smgmt.GetAppEnvVars(ctx, in, authApi, &k8smgmt.TestReplacementVars)
		if err != nil {
			return err
		}
	}

	if err := s.validatePolicies(stm, in); err != nil {
		return err
	}
	if err := s.validateAlertPolicies(stm, in); err != nil {
		return err
	}
	return nil
}

func (s *AppApi) CreateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "CreateApp", "app", in.Key.String())
	var err error

	if err = in.Validate(edgeproto.AppAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}

	start := time.Now()
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		log.SpanLog(ctx, log.DebugLevelApi, "CreateApp begin ApplySTMWait", "app", in.Key.String())
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}

		err = s.configureApp(ctx, stm, in, in.Revision)
		if err != nil {
			return err
		}
		s.all.appInstRefsApi.createRef(stm, &in.Key)

		in.CreatedAt = cloudcommon.TimeToTimestamp(time.Now())
		s.store.STMPut(stm, in)
		elapsed := time.Since(start)
		log.SpanLog(ctx, log.DebugLevelApi, "CreateApp finish ApplySTMWait", "app", in.Key.String(), "elapsed", elapsed, "err", err)
		return nil
	})
	log.SpanLog(ctx, log.DebugLevelApi, "CreateApp done", "app", in.Key.String())
	return &edgeproto.Result{}, err
}

func (s *AppApi) UpdateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	err := in.ValidateUpdateFields()
	if err != nil {
		return &edgeproto.Result{}, err
	}
	inUseCannotUpdate := []string{
		edgeproto.AppFieldAccessPorts,
		edgeproto.AppFieldSkipHcPorts,
		edgeproto.AppFieldDeployment,
		edgeproto.AppFieldDeploymentGenerator,
	}
	canAlwaysUpdate := map[string]bool{
		edgeproto.AppFieldTrusted:     true,
		edgeproto.AppFieldVmAppOsType: true, // will not affect current AppInsts, but needed to launch existing apps on VCD
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
		refs := edgeproto.AppInstRefs{}
		s.all.appInstRefsApi.store.STMGet(stm, &in.Key, &refs)
		appInstExists := len(refs.Insts) > 0
		if appInstExists {
			if cur.Deployment != cloudcommon.DeploymentTypeKubernetes &&
				cur.Deployment != cloudcommon.DeploymentTypeDocker &&
				cur.Deployment != cloudcommon.DeploymentTypeHelm {
				for f := range fields {
					if in.IsKeyField(f) {
						continue
					}
					if !canAlwaysUpdate[f] {
						return fmt.Errorf("Update App field %s not supported for deployment: %s when AppInsts exist", edgeproto.AppAllFieldsStringMap[f], cur.Deployment)
					}
				}
			}
			for _, field := range inUseCannotUpdate {
				if _, found := fields[field]; found {
					return fmt.Errorf("Cannot update %s when AppInst exists", edgeproto.AppAllFieldsStringMap[field])
				}
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
		_, deploymentSpecified := fields[edgeproto.AppFieldDeployment]
		if deploymentSpecified {
			// likely any existing manifest is no longer valid,
			// reset it in case a new manifest was not specified
			// and it needs to be auto-generated.
			// If a new manifest is specified during update, it
			// will overwrite this blank setting.
			cur.DeploymentManifest = ""
		}
		_, deploymentManifestSpecified := fields[edgeproto.AppFieldDeploymentManifest]
		_, accessPortSpecified := fields[edgeproto.AppFieldAccessPorts]
		_, TrustedSpecified := fields[edgeproto.AppFieldTrusted]
		_, requiredOutboundSpecified := fields[edgeproto.AppFieldRequiredOutboundConnections]

		if deploymentManifestSpecified {
			// reset the deployment generator
			cur.DeploymentGenerator = ""
		} else if accessPortSpecified {
			if cur.DeploymentGenerator == "" && cur.Deployment == cloudcommon.DeploymentTypeKubernetes {
				// No generator means the user previously provided a manifest.  Force them to do so again when changing ports so
				// that they do not accidentally lose their provided manifest details
				return fmt.Errorf("kubernetes manifest which was previously specified must be provided again when changing access ports")
			} else if cur.Deployment == cloudcommon.DeploymentTypeKubernetes {
				// force regeneration of manifest for k8s
				cur.DeploymentManifest = ""
			}
		}
		cur.CopyInFields(in)
		// for any changes that can affect trust policy, verify the app is still valid for all
		// cloudlets onto which it is deployed.
		if requiredOutboundSpecified ||
			(TrustedSpecified && !in.Trusted) {
			appInstKeys := make(map[edgeproto.AppInstKey]struct{})

			for k, _ := range refs.Insts {
				// disallow delete if static instances are present
				inst := edgeproto.AppInst{}
				edgeproto.AppInstKeyStringParse(k, &inst.Key)
				appInstKeys[inst.Key] = struct{}{}
			}
			err = s.all.cloudletApi.VerifyTrustPoliciesForAppInsts(ctx, &cur, appInstKeys)
			if err != nil {
				if TrustedSpecified && !in.Trusted {
					// override the usual errmsg to be clear for this scenario
					return fmt.Errorf("Cannot set app to untrusted which has an instance on a trusted cloudlet")
				}
				return err
			}
		}
		// for update, trigger regenerating deployment manifest
		if cur.DeploymentGenerator != "" {
			cur.DeploymentManifest = ""
		}
		newRevision := in.Revision
		if newRevision == "" {
			newRevision = time.Now().Format("2006-01-02T150405")
			log.SpanLog(ctx, log.DebugLevelApi, "Setting new revision to current timestamp", "Revision", in.Revision)
		}
		if err := s.configureApp(ctx, stm, &cur, newRevision); err != nil {
			return err
		}
		cur.UpdatedAt = cloudcommon.TimeToTimestamp(time.Now())
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AppApi) DeleteApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	if !s.HasApp(&in.Key) {
		// key doesn't exist
		return &edgeproto.Result{}, in.Key.NotFoundError()
	}
	if s.all.trustPolicyExceptionApi.TrustPolicyExceptionForAppKeyExists(&in.Key) {
		return &edgeproto.Result{}, errors.New("Application in use by Trust Policy Exception")
	}

	// set state to prevent new AppInsts from being created from this App
	dynInsts := []*edgeproto.AppInst{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return in.Key.NotFoundError()
		}
		// use refs to check existing AppInsts to avoid race conditions
		refs := edgeproto.AppInstRefs{}
		s.all.appInstRefsApi.store.STMGet(stm, &in.Key, &refs)
		for k, _ := range refs.Insts {
			// disallow delete if static instances are present
			inst := edgeproto.AppInst{}
			edgeproto.AppInstKeyStringParse(k, &inst.Key)
			if !s.all.appInstApi.store.STMGet(stm, &inst.Key, &inst) {
				// no inst?
				log.SpanLog(ctx, log.DebugLevelApi, "AppInst not found by refs, skipping for delete", "AppInst", inst.Key)
				continue
			}
			if inst.Liveness == edgeproto.Liveness_LIVENESS_STATIC {
				return errors.New("Application in use by static AppInst")
			}
			dynInsts = append(dynInsts, &inst)
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
	if err = s.all.appInstApi.AutoDelete(ctx, dynInsts); err != nil {
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
		s.all.appInstRefsApi.deleteRef(stm, &in.Key)
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

func validateAppConfigsForDeployment(ctx context.Context, configs []*edgeproto.ConfigFile, deployment string) error {
	log.SpanLog(ctx, log.DebugLevelApi, "validateAppConfigsForDeployment")
	for _, cfg := range configs {
		invalid := false
		switch cfg.Kind {
		case edgeproto.AppConfigHelmYaml:
			if deployment != cloudcommon.DeploymentTypeHelm {
				invalid = true
			}
			// Validate that this is a valid url
			_, err := cloudcommon.GetDeploymentManifest(ctx, nil, cfg.Config)
			log.SpanLog(ctx, log.DebugLevelApi, "error getting deployment manifest for app", "err", err)
			if err != nil {
				return err
			}
		case edgeproto.AppConfigEnvYaml:
			if deployment != cloudcommon.DeploymentTypeKubernetes {
				invalid = true
			}
		}
		if invalid {
			return fmt.Errorf("Invalid Config Kind(%s) for deployment type(%s)", cfg.Kind, deployment)
		}
		if cfg.Config == "" {
			return fmt.Errorf("Empty config for config kind %s", cfg.Kind)
		}
	}
	log.SpanLog(ctx, log.DebugLevelApi, "validateAppConfigsForDeployment success")
	return nil
}

func validateRequiredOutboundConnections(rules []edgeproto.SecurityRule) error {
	return edgeproto.ValidateSecurityRules(rules)
}

func (s *AppApi) validatePolicies(stm concurrency.STM, app *edgeproto.App) error {
	// make sure policies exist
	numPolicies := 0
	for name, _ := range app.GetAutoProvPolicies() {
		policyKey := edgeproto.PolicyKey{}
		policyKey.Organization = app.Key.Organization
		policyKey.Name = name
		policy := edgeproto.AutoProvPolicy{}
		if !s.all.autoProvPolicyApi.store.STMGet(stm, &policyKey, &policy) {
			return policyKey.NotFoundError()
		}
		numPolicies++
	}
	if numPolicies > 0 {
		if err := validateAutoDeployApp(stm, app); err != nil {
			return err
		}
	}
	return nil
}

func validateAutoDeployApp(stm concurrency.STM, app *edgeproto.App) error {
	// to reduce the number of permutations of reservable autocluster
	// configurations, we only support a subset of all features
	// for autoclusters and auto-provisioning.
	if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_DIRECT {
		return fmt.Errorf("For auto-provisioning or auto-clusters, App access type direct is not supported")
	}
	if app.DefaultFlavor.Name == "" {
		return fmt.Errorf("For auto-provisioning or auto-clusters, App must have default flavor defined")
	}
	validDeployments := []string{
		cloudcommon.DeploymentTypeKubernetes,
		cloudcommon.DeploymentTypeHelm,
		cloudcommon.DeploymentTypeDocker,
	}
	validDep := false
	for _, dep := range validDeployments {
		if app.Deployment == dep {
			validDep = true
			break
		}
	}
	if !validDep {
		return fmt.Errorf("For auto-provisioning or auto-clusters, App deployment types are limited to %s", strings.Join(validDeployments, ", "))
	}
	return nil
}

func manifestContainsDaemonSet(manifest string) bool {
	objs, _, err := cloudcommon.DecodeK8SYaml(manifest)
	if err != nil {
		return false
	}
	for ii, _ := range objs {
		switch objs[ii].(type) {
		case *appsv1.DaemonSet:
			return true
		}
	}
	return false
}

func (s *AppApi) AddAppAlertPolicy(ctx context.Context, in *edgeproto.AppAlertPolicy) (*edgeproto.Result, error) {
	cur := edgeproto.App{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.AppKey, &cur) {
			return in.AppKey.NotFoundError()
		}
		for _, name := range cur.AlertPolicies {
			if name == in.AlertPolicy {
				return fmt.Errorf("Alert %s already monitored on App", name)
			}
		}
		cur.AlertPolicies = append(cur.AlertPolicies, in.AlertPolicy)
		if err := s.validateAlertPolicies(stm, &cur); err != nil {
			return err
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AppApi) RemoveAppAlertPolicy(ctx context.Context, in *edgeproto.AppAlertPolicy) (*edgeproto.Result, error) {
	cur := edgeproto.App{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.AppKey, &cur) {
			return in.AppKey.NotFoundError()
		}
		changed := false
		for ii, name := range cur.AlertPolicies {
			if name != in.AlertPolicy {
				continue
			}
			cur.AlertPolicies = append(cur.AlertPolicies[:ii], cur.AlertPolicies[ii+1:]...)
			changed = true
			break
		}
		if changed {
			s.store.STMPut(stm, &cur)
			return nil
		}
		return (&edgeproto.AlertPolicyKey{}).NotFoundError()
	})
	return &edgeproto.Result{}, err
}

func (s *AppApi) validateAlertPolicies(stm concurrency.STM, app *edgeproto.App) error {
	// make sure alerts exist
	for ii := range app.AlertPolicies {
		alertKey := edgeproto.AlertPolicyKey{
			Name:         app.AlertPolicies[ii],
			Organization: app.Key.Organization,
		}
		alert := edgeproto.AlertPolicy{}
		if !s.all.alertPolicyApi.store.STMGet(stm, &alertKey, &alert) {
			return alertKey.NotFoundError()
		}
	}
	return nil
}

func (s *AppApi) UsesAlertPolicy(key *edgeproto.AlertPolicyKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		app := data.Obj
		if app.Key.Organization == key.Organization {
			for _, name := range app.AlertPolicies {
				if name == key.Name {
					return true
				}
			}
		}
	}
	return false
}

func (s *AppApi) tryDeployApp(ctx context.Context, stm concurrency.STM, app *edgeproto.App, appInst *edgeproto.AppInst, cloudlet *edgeproto.Cloudlet, cloudletInfo *edgeproto.CloudletInfo,
	cloudletRefs *edgeproto.CloudletRefs, numNodes uint32) error {

	deployment := app.Deployment
	if deployment == cloudcommon.DeploymentTypeHelm {
		log.SpanLog(ctx, log.DebugLevelApi, "DryRunDeploy deployment Helm replaced with Kube")
		deployment = cloudcommon.DeploymentTypeKubernetes
	}
	if app.Deployment == cloudcommon.DeploymentTypeKubernetes && app.AllowServerless {
		log.SpanLog(ctx, log.DebugLevelApi, "DryRunDeploy check multi-tenant TBI for kube+serverless")
		// note: helm not allowed
		// TODO: check if multi-tenant cluster exists on cloudlet, and check if resources are available
	}
	if deployment == cloudcommon.DeploymentTypeKubernetes || deployment == cloudcommon.DeploymentTypeDocker {
		// look for reservable ClusterInst. This emulates behavior in createAppInstInternal().
		s.all.clusterInstApi.cache.Mux.Lock()
		canDeploy := false
		for _, data := range s.all.clusterInstApi.cache.Objs {
			if !cloudlet.Key.Matches(&data.Obj.Key.CloudletKey) {
				continue
			}
			if data.Obj.Reservable && data.Obj.ReservedBy == "" && data.Obj.Deployment == deployment && data.Obj.Flavor.Name == app.DefaultFlavor.Name {
				// can deploy to existing reservable ClusterInst
				canDeploy = true
				break
			}
		}
		s.all.clusterInstApi.cache.Mux.Unlock()
		if canDeploy {
			log.SpanLog(ctx, log.DebugLevelApi, "DryRunDeploy Ok for reservable cluster", "cloudlet", cloudlet.Key.Name)
			return nil
		}
		// see if we can create a new ClusterInst
		targetCluster := edgeproto.ClusterInst{}
		if deployment == cloudcommon.DeploymentTypeKubernetes {
			targetCluster.MasterNodeFlavor = s.all.settingsApi.Get().MasterNodeFlavor
		}
		targetCluster.NodeFlavor = app.DefaultFlavor.Name
		targetCluster.Deployment = deployment
		if deployment == cloudcommon.DeploymentTypeKubernetes {
			targetCluster.NumMasters = 1
			targetCluster.NumNodes = numNodes
		}
		return s.all.clusterInstApi.validateResources(ctx, stm, &targetCluster, nil, nil, cloudlet, cloudletInfo, cloudletRefs, NoGenResourceAlerts)
	}
	if deployment == cloudcommon.DeploymentTypeVM {
		return s.all.clusterInstApi.validateResources(ctx, stm, nil, app, appInst, cloudlet, cloudletInfo, cloudletRefs, NoGenResourceAlerts)
	}
	return fmt.Errorf("Unsupported deployment type %s\n", app.Deployment)
}

func (s *AppApi) ShowCloudletsForAppDeployment(in *edgeproto.DeploymentCloudletRequest, cb edgeproto.AppApi_ShowCloudletsForAppDeploymentServer) error {
	ctx := cb.Context()
	var allclds = make(map[edgeproto.CloudletKey]string)
	app := in.App
	flavor := in.App.DefaultFlavor
	var numNodes uint32 = 2

	if in.NumNodes != 0 {
		numNodes = in.NumNodes
	}
	if flavor.Name == "" {
		return fmt.Errorf("No flavor specified for App")
	}
	s.all.cloudletApi.cache.GetAllKeys(ctx, func(k *edgeproto.CloudletKey, modRev int64) {
		allclds[*k] = ""
	})

	// Generate a list of all cloudlets that find a match for app.DefaultFlavor
	for cldkey, _ := range allclds {
		fm := edgeproto.FlavorMatch{
			Key:        cldkey,
			FlavorName: flavor.Name,
		}
		// since the validate is going to consider the appInst.VmFlavor, we'll need spec.FlavorName set in the test
		// AppInst object, not the meta flavor.
		spec, err := s.all.cloudletApi.FindFlavorMatch(ctx, &fm)
		if err != nil {
			delete(allclds, cldkey)
			continue
		} else {
			// need the mapped FlavorName for appInst.VmFlavor
			allclds[cldkey] = spec.FlavorName
		}
	}
	// If an instance of this App were to be deployed now, check if our resource mgr thinks
	// there are sufficient resources to support the creation. Assumes the appInst.Flavor and appInst.VmFlavor
	// would use the App templates default flavor.
	if in.DryRunDeploy {
		var err error
		appInst := edgeproto.AppInst{}
		if app.Deployment == "" {
			log.SpanLog(ctx, log.DebugLevelApi, "DryRunDeploy mandatory app Deployment not found for App")
			return fmt.Errorf("No deployment found on candidate App")
		}
		appInst.Flavor = app.DefaultFlavor

		log.SpanLog(ctx, log.DebugLevelApi, "ShowCloudletsForAppDeployment dry run deployment")
		// For all remaining cloudlets, check available resources
		err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			for key, _ := range allclds {
				appInst.VmFlavor = allclds[key]
				cloudlet := edgeproto.Cloudlet{}
				if !s.all.cloudletApi.store.STMGet(stm, &key, &cloudlet) {
					log.SpanLog(ctx, log.DebugLevelApi, "ShowCloudletsForAppDeployment cloudlet not found", "cloudlet", key)
					continue
				}
				cloudletRefs := edgeproto.CloudletRefs{}
				if !s.all.cloudletRefsApi.store.STMGet(stm, &key, &cloudletRefs) {
					initCloudletRefs(&cloudletRefs, &key)
				}
				cloudletInfo := edgeproto.CloudletInfo{}
				if !s.all.cloudletInfoApi.store.STMGet(stm, &key, &cloudletInfo) {
					log.SpanLog(ctx, log.DebugLevelApi, "ShowCloudletsForAppDeployment cloudletinfo not found, skipping", "cloudlet", key)
					delete(allclds, key)
					continue
				}
				err = s.tryDeployApp(ctx, stm, app, &appInst, &cloudlet, &cloudletInfo, &cloudletRefs, numNodes)
				if err != nil {
					delete(allclds, key)
					log.SpanLog(ctx, log.DebugLevelApi, "DryRunDeploy failed for", "cloudlet", cloudlet.Key, "error", err)
					continue
				}
				cb.Send(&key)
				log.SpanLog(ctx, log.DebugLevelApi, "ShowCloudletsForAppDeployment dry run deployment succeeded for", "cloudlet", cloudlet.Key.Name)
			}
			return nil
		})
	}
	for key, _ := range allclds {
		cb.Send(&key)
	}
	return nil
}
