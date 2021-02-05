package fake

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
)

type Platform struct {
	consoleServer *httptest.Server
}

var (
	ResourceExternalIps = "External IPs"

	FakeRamUsed         = uint64(0)
	FakeVcpusUsed       = uint64(0)
	FakeDiskUsed        = uint64(0)
	FakeExternalIpsUsed = uint64(0)

	FakeRamMax         = uint64(40960)
	FakeVcpusMax       = uint64(50)
	FakeDiskMax        = uint64(5000)
	FakeExternalIpsMax = uint64(10)
)

var FakeAppDNSRoot = "fake.net"

var FakeClusterVMs = map[edgeproto.ClusterInstKey][]edgeproto.VmInfo{}

var FakeFlavorList = []*edgeproto.FlavorInfo{
	&edgeproto.FlavorInfo{
		Name:  "x1.small",
		Vcpus: uint64(2),
		Ram:   uint64(4096),
		Disk:  uint64(40),
	},
}

var RootLBFlavor = edgeproto.Flavor{
	Key:   edgeproto.FlavorKey{Name: "rootlb-flavor"},
	Vcpus: uint64(2),
	Ram:   uint64(4096),
	Disk:  uint64(40),
}

func (s *Platform) GetType() string {
	return "fake"
}

var fakeProps = map[string]*edgeproto.PropertyInfo{
	// Property: Default-Value
	"PROP_1": &edgeproto.PropertyInfo{
		Name:        "Property 1",
		Description: "First Property",
		Secret:      true,
		Mandatory:   true,
	},
	"PROP_2": &edgeproto.PropertyInfo{
		Name:        "Property 2",
		Description: "Second Property",
		Mandatory:   true,
	},
}

func (s *Platform) Init(ctx context.Context, platformConfig *platform.PlatformConfig, caches *platform.Caches, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "running in fake cloudlet mode")
	platformConfig.NodeMgr.Debug.AddDebugFunc("fakecmd", s.runDebug)

	updateCallback(edgeproto.UpdateTask, "Done intializing fake platform")
	s.consoleServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Console Content")
	}))
	// Update resource info for platformVM and RootLBVM
	FakeRamUsed += 4096 + 4096
	FakeVcpusUsed += 2 + 2
	FakeDiskUsed += 40 + 40
	FakeExternalIpsUsed += 1
	return nil
}

func (s *Platform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	info.OsMaxRam = FakeRamMax
	info.OsMaxVcores = FakeVcpusMax
	info.OsMaxVolGb = FakeDiskMax
	info.Flavors = FakeFlavorList
	return nil
}

func (s *Platform) UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "Updating Cluster Inst")
	vmNameSuffix := k8smgmt.GetCloudletClusterName(&clusterInst.Key)
	fakeNodes := make(map[string]struct{})
	for ii := uint32(0); ii < clusterInst.NumNodes; ii++ {
		nodeName := fmt.Sprintf("fake-node-%d-%s", ii+1, vmNameSuffix)
		fakeNodes[nodeName] = struct{}{}
	}
	fakeMasters := make(map[string]struct{})
	for ii := uint32(0); ii < clusterInst.NumMasters; ii++ {
		masterName := fmt.Sprintf("fake-master-%d-%s", ii+1, vmNameSuffix)
		fakeMasters[masterName] = struct{}{}
	}
	cVMs, ok := FakeClusterVMs[clusterInst.Key]
	if !ok {
		return fmt.Errorf("missing cluster vms for %v", clusterInst.Key)
	}
	for _, vmInfo := range cVMs {
		if vmInfo.Type == cloudcommon.VMTypeClusterNode {
			if _, ok := fakeNodes[vmInfo.Name]; ok {
				delete(fakeNodes, vmInfo.Name)
			}
		} else if vmInfo.Type == cloudcommon.VMTypeClusterMaster {
			if _, ok := fakeMasters[vmInfo.Name]; ok {
				delete(fakeMasters, vmInfo.Name)
			}
		}
	}
	for vmName, _ := range fakeNodes {
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        vmName,
			Type:        cloudcommon.VMTypeClusterNode,
			InfraFlavor: "x1.small",
			Status:      "ACTIVE",
		})
		FakeRamUsed += 4096
		FakeVcpusUsed += 2
		FakeDiskUsed += 40
	}
	for vmName, _ := range fakeMasters {
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        vmName,
			Type:        cloudcommon.VMTypeClusterMaster,
			InfraFlavor: "x1.small",
			Status:      "ACTIVE",
		})
		FakeRamUsed += 4096
		FakeVcpusUsed += 2
		FakeDiskUsed += 40
	}
	return nil
}
func (s *Platform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "fake CreateClusterInst", "clusterInst", clusterInst)
	updateCallback(edgeproto.UpdateTask, "First Create Task")
	updateCallback(edgeproto.UpdateTask, "Second Create Task")
	vmNameSuffix := k8smgmt.GetCloudletClusterName(&clusterInst.Key)
	if len(FakeClusterVMs) == 0 {
		FakeClusterVMs = make(map[edgeproto.ClusterInstKey][]edgeproto.VmInfo)
	}
	FakeClusterVMs[clusterInst.Key] = []edgeproto.VmInfo{}
	for ii := uint32(0); ii < clusterInst.NumMasters; ii++ {
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        fmt.Sprintf("fake-master-%d-%s", ii+1, vmNameSuffix),
			Type:        cloudcommon.VMTypeClusterMaster,
			InfraFlavor: "x1.small",
			Status:      "ACTIVE",
		})
		FakeRamUsed += 4096
		FakeVcpusUsed += 2
		FakeDiskUsed += 40
	}
	for ii := uint32(0); ii < clusterInst.NumNodes; ii++ {
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        fmt.Sprintf("fake-node-%d-%s", ii+1, vmNameSuffix),
			Type:        cloudcommon.VMTypeClusterNode,
			InfraFlavor: "x1.small",
			Status:      "ACTIVE",
		})
		FakeRamUsed += 4096
		FakeVcpusUsed += 2
		FakeDiskUsed += 40
	}
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		rootLBFQDN := cloudcommon.GetDedicatedLBFQDN(&clusterInst.Key.CloudletKey, &clusterInst.Key.ClusterKey, FakeAppDNSRoot)
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        rootLBFQDN,
			Type:        cloudcommon.VMTypeRootLB,
			InfraFlavor: "x1.small",
			Status:      "ACTIVE",
		})
		FakeRamUsed += 4096
		FakeVcpusUsed += 2
		FakeDiskUsed += 40
		FakeExternalIpsUsed += 1
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "fake ClusterInst ready")
	return nil
}

func (s *Platform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "First Delete Task")
	updateCallback(edgeproto.UpdateTask, "Second Delete Task")
	rootLBFQDN := cloudcommon.GetDedicatedLBFQDN(&clusterInst.Key.CloudletKey, &clusterInst.Key.ClusterKey, FakeAppDNSRoot)
	vms := make(map[string]struct{})
	vms[rootLBFQDN] = struct{}{}
	vmNameSuffix := k8smgmt.GetCloudletClusterName(&clusterInst.Key)
	for ii := uint32(0); ii < clusterInst.NumMasters; ii++ {
		vmName := fmt.Sprintf("fake-master-%d-%s", ii+1, vmNameSuffix)
		vms[vmName] = struct{}{}
	}
	for ii := uint32(0); ii < clusterInst.NumNodes; ii++ {
		vmName := fmt.Sprintf("fake-node-%d-%s", ii+1, vmNameSuffix)
		vms[vmName] = struct{}{}
	}
	newVMs := []edgeproto.VmInfo{}
	if clusterVMs, ok := FakeClusterVMs[clusterInst.Key]; ok {
		for _, vm := range clusterVMs {
			if _, ok := vms[vm.Name]; ok {
				FakeRamUsed -= 4096
				FakeVcpusUsed -= 2
				FakeDiskUsed -= 40
				if vm.Name == rootLBFQDN {
					FakeExternalIpsUsed -= 1
				}
				continue
			}
			newVMs = append(newVMs, vm)
		}
		delete(FakeClusterVMs, clusterInst.Key)
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "fake ClusterInst deleted")
	return nil
}

func (s *Platform) GetCloudletInfraResources(ctx context.Context) (*edgeproto.InfraResourcesSnapshot, error) {
	var resources edgeproto.InfraResourcesSnapshot
	platvm := edgeproto.VmInfo{
		Name:        "fake-platform-vm",
		Type:        "platform",
		InfraFlavor: "x1.small",
		Status:      "ACTIVE",
		Ipaddresses: []edgeproto.IpAddr{
			{ExternalIp: "10.101.100.10"},
		},
	}
	resources.PlatformVms = append(resources.PlatformVms, platvm)
	rlbvm := edgeproto.VmInfo{
		Name:        "fake-rootlb-vm",
		Type:        "rootlb",
		InfraFlavor: "x1.small",
		Status:      "ACTIVE",
		Ipaddresses: []edgeproto.IpAddr{
			{ExternalIp: "10.101.100.11"},
		},
	}
	resources.PlatformVms = append(resources.PlatformVms, rlbvm)

	resources.Info = []edgeproto.InfraResource{
		edgeproto.InfraResource{
			Name:          cloudcommon.ResourceRamMb,
			Value:         FakeRamUsed,
			InfraMaxValue: FakeRamMax,
			Units:         cloudcommon.ResourceRamUnits,
		},
		edgeproto.InfraResource{
			Name:          cloudcommon.ResourceVcpus,
			Value:         FakeVcpusUsed,
			InfraMaxValue: FakeVcpusMax,
		},
		edgeproto.InfraResource{
			Name:          cloudcommon.ResourceDiskGb,
			Value:         FakeDiskUsed,
			InfraMaxValue: FakeDiskMax,
			Units:         cloudcommon.ResourceDiskUnits,
		},
		edgeproto.InfraResource{
			Name:          ResourceExternalIps,
			Value:         FakeExternalIpsUsed,
			InfraMaxValue: FakeExternalIpsMax,
		},
	}

	return &resources, nil
}

// called by controller, make sure it doesn't make any calls to infra API
func (s *Platform) GetClusterAdditionalResources(ctx context.Context, cloudlet *edgeproto.Cloudlet, vmResources []edgeproto.VMResource, infraResMap map[string]*edgeproto.InfraResource) map[string]*edgeproto.InfraResource {
	// resource name -> resource units
	cloudletRes := map[string]string{
		ResourceExternalIps: "",
	}
	resInfo := make(map[string]*edgeproto.InfraResource)
	for resName, resUnits := range cloudletRes {
		resMax := uint64(0)
		if infraRes, ok := infraResMap[resName]; ok {
			resMax = infraRes.InfraMaxValue
		}
		resInfo[resName] = &edgeproto.InfraResource{
			Name:          resName,
			InfraMaxValue: resMax,
			Units:         resUnits,
		}
	}

	for _, vmRes := range vmResources {
		if vmRes.Type == cloudcommon.VMTypeRootLB {
			resInfo[ResourceExternalIps].Value += 1
		}
	}
	return resInfo
}

func (s *Platform) GetCloudletResourceQuotaProps(ctx context.Context) (*edgeproto.CloudletResourceQuotaProps, error) {
	return &edgeproto.CloudletResourceQuotaProps{
		Props: []edgeproto.InfraResource{
			edgeproto.InfraResource{
				Name:        ResourceExternalIps,
				Description: "Limit on external IPs available",
			},
		},
	}, nil
}

func (s *Platform) GetClusterAdditionalResourceMetric(ctx context.Context, cloudlet *edgeproto.Cloudlet, resMetric *edgeproto.Metric, resources []edgeproto.VMResource) error {
	externalIpsUsed := uint64(0)
	for _, vmRes := range resources {
		if vmRes.Type == cloudcommon.VMTypeRootLB {
			externalIpsUsed += 1
		}
	}

	resMetric.AddIntVal("externalIpsUsed", externalIpsUsed)
	return nil
}

func (s *Platform) GetClusterInfraResources(ctx context.Context, clusterKey *edgeproto.ClusterInstKey) (*edgeproto.InfraResources, error) {
	var resources edgeproto.InfraResources
	if vms, ok := FakeClusterVMs[*clusterKey]; ok {
		resources.Vms = append(resources.Vms, vms...)
	}
	return &resources, nil
}

func (s *Platform) CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "Creating App Inst")
	log.SpanLog(ctx, log.DebugLevelInfra, "fake AppInst ready")
	if app.Deployment == cloudcommon.DeploymentTypeVM {
		appFQN := cloudcommon.GetAppFQN(&app.Key)
		clusterInst.Key.ClusterKey.Name = appFQN + "-" + appInst.Key.ClusterInstKey.ClusterKey.Name
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        appFQN,
			Type:        cloudcommon.VMTypeAppVM,
			InfraFlavor: "x1.small",
			Status:      "ACTIVE",
		})
		FakeRamUsed += 4096
		FakeVcpusUsed += 2
		FakeDiskUsed += 40
		if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_DIRECT {
			FakeExternalIpsUsed += 1
		}
	}
	return nil
}

func (s *Platform) DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "First Delete Task")
	updateCallback(edgeproto.UpdateTask, "Second Delete Task")
	log.SpanLog(ctx, log.DebugLevelInfra, "fake AppInst deleted")
	if app.Deployment == cloudcommon.DeploymentTypeVM {
		appFQN := cloudcommon.GetAppFQN(&app.Key)
		clusterInst.Key.ClusterKey.Name = appFQN + "-" + appInst.Key.ClusterInstKey.ClusterKey.Name
		FakeRamUsed -= 4096
		FakeVcpusUsed -= 2
		FakeDiskUsed -= 40
		if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_DIRECT {
			FakeExternalIpsUsed -= 1
		}
		delete(FakeClusterVMs, clusterInst.Key)
	}
	return nil
}

func (s *Platform) UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "fake appInst updated")
	return nil
}

func (s *Platform) GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	if app.Deployment == cloudcommon.DeploymentTypeKubernetes {
		rt := &edgeproto.AppInstRuntime{}
		for ii := uint32(0); ii < clusterInst.NumNodes; ii++ {
			rt.ContainerIds = append(rt.ContainerIds, fmt.Sprintf("appOnClusterNode%d", ii))
		}
		return rt, nil
	}
	return &edgeproto.AppInstRuntime{}, nil
}

func (s *Platform) GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst, clientType string) (ssh.Client, error) {
	return &pc.LocalClient{}, nil
}

func (s *Platform) GetNodePlatformClient(ctx context.Context, node *edgeproto.CloudletMgmtNode, ops ...pc.SSHClientOp) (ssh.Client, error) {
	return &pc.LocalClient{}, nil
}

func (s *Platform) ListCloudletMgmtNodes(ctx context.Context, clusterInsts []edgeproto.ClusterInst) ([]edgeproto.CloudletMgmtNode, error) {
	return []edgeproto.CloudletMgmtNode{
		edgeproto.CloudletMgmtNode{
			Type: "platformvm",
			Name: "platformvmname",
		},
	}, nil
}

func (s *Platform) GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	if req.Cmd != nil {
		return req.Cmd.Command, nil
	}
	if req.Log != nil {
		return "echo \"here's some logs\"", nil
	}
	return "", fmt.Errorf("no cmd or log specified in exec request")
}

func (s *Platform) GetConsoleUrl(ctx context.Context, app *edgeproto.App) (string, error) {
	if s.consoleServer != nil {
		return s.consoleServer.URL + "?token=xyz", nil
	}
	return "", fmt.Errorf("no console server to fetch URL from")
}

func (s *Platform) CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, caches *platform.Caches, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "create fake cloudlet", "key", cloudlet.Key)
	if cloudlet.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS {
		return nil
	}
	updateCallback(edgeproto.UpdateTask, "Creating Cloudlet")
	updateCallback(edgeproto.UpdateTask, "Starting CRMServer")
	err := cloudcommon.StartCRMService(ctx, cloudlet, pfConfig)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "fake cloudlet create failed", "err", err)
		return err
	}
	return nil
}

func (s *Platform) UpdateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelInfra, "update fake Cloudlet", "cloudlet", cloudlet)
	for key, val := range cloudlet.EnvVar {
		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Updating envvar, %s=%s", key, val))
	}
	return nil
}

func (s *Platform) UpdateTrustPolicy(ctx context.Context, TrustPolicy *edgeproto.TrustPolicy) error {
	log.DebugLog(log.DebugLevelInfra, "fake UpdateTrustPolicy begin", "policy", TrustPolicy)
	return nil
}

func (s *Platform) DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, caches *platform.Caches, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelInfra, "delete fake Cloudlet", "key", cloudlet.Key)
	updateCallback(edgeproto.UpdateTask, "Deleting Cloudlet")
	updateCallback(edgeproto.UpdateTask, "Stopping CRMServer")
	err := cloudcommon.StopCRMService(ctx, cloudlet)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "fake cloudlet delete failed", "err", err)
		return err
	}

	return nil
}

func (s *Platform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Saving cloudlet access vars", "cloudletName", cloudlet.Key.Name)
	return nil
}

func (s *Platform) GetCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config) (map[string]string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "Get cloudlet access vars", "cloudletName", cloudlet.Key.Name)
	return map[string]string{}, nil
}

func (s *Platform) DeleteCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Deleting cloudlet access vars", "cloudletName", cloudlet.Key.Name)
	return nil
}

func (s *Platform) SetPowerState(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Setting power state", "state", appInst.PowerState)
	return nil
}

func (s *Platform) runDebug(ctx context.Context, req *edgeproto.DebugRequest) string {
	return "ran some debug"
}

func (s *Platform) SyncControllerCache(ctx context.Context, caches *platform.Caches, cloudletState dme.CloudletState) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SyncControllerCache", "state", cloudletState)
	return nil
}

func (s *Platform) GetCloudletManifest(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, accessApi platform.AccessApi, flavor *edgeproto.Flavor, caches *platform.Caches) (*edgeproto.CloudletManifest, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "Get cloudlet manifest", "cloudletName", cloudlet.Key.Name)
	return &edgeproto.CloudletManifest{Manifest: "fake manifest\n" + pfConfig.CrmAccessPrivateKey}, nil
}

func (s *Platform) VerifyVMs(ctx context.Context, vms []edgeproto.VM) error {
	for _, vm := range vms {
		// For unit testing
		if vm.Name == "vmFailVerification" {
			return fmt.Errorf("failed to verify VM")
		}
	}
	return nil
}

func (s *Platform) GetCloudletProps(ctx context.Context) (*edgeproto.CloudletProps, error) {
	return &edgeproto.CloudletProps{Properties: fakeProps}, nil
}

func (s *Platform) GetAccessData(ctx context.Context, cloudlet *edgeproto.Cloudlet, region string, vaultConfig *vault.Config, dataType string, arg []byte) (map[string]string, error) {
	return nil, nil
}

func (s *Platform) GetRestrictedCloudletStatus(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "Setting up cloudlet")
	return nil
}

func (s *Platform) GetRootLBClients(ctx context.Context) (map[string]ssh.Client, error) {
	return nil, nil
}

func (s *Platform) GetRootLBFlavor(ctx context.Context) (*edgeproto.Flavor, error) {
	return &RootLBFlavor, nil
}
