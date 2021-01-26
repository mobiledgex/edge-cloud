package fake

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
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
	ResourceRam   = "RAM"
	ResourceVcpus = "vCPUs"
	ResourceDisk  = "disk"

	ResourceRamUnits  = "MB"
	ResourceDiskUnits = "GB"

	FakeRamUsed   = uint64(0)
	FakeVcpusUsed = uint64(0)
	FakeDiskUsed  = uint64(0)

	FakeRamMax   = uint64(40960)
	FakeVcpusMax = uint64(50)
	FakeDiskMax  = uint64(5000)
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
	FakeRamUsed += uint64(4096) + uint64(4096)
	FakeVcpusUsed += uint64(2) + uint64(2)
	FakeDiskUsed += uint64(40) + uint64(40)
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
			Type:        "cluster-master",
			InfraFlavor: "m4.small",
			Status:      "ACTIVE",
		})
		FakeRamUsed += uint64(4096)
		FakeVcpusUsed += uint64(2)
		FakeDiskUsed += uint64(40)
	}
	for ii := uint32(0); ii < clusterInst.NumNodes; ii++ {
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        fmt.Sprintf("fake-node-%d-%s", ii+1, vmNameSuffix),
			Type:        "cluster-node",
			InfraFlavor: "m4.small",
			Status:      "ACTIVE",
		})
		FakeRamUsed += uint64(4096)
		FakeVcpusUsed += uint64(2)
		FakeDiskUsed += uint64(40)
	}
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		rootLBFQDN := cloudcommon.GetDedicatedLBFQDN(&clusterInst.Key.CloudletKey, &clusterInst.Key.ClusterKey, FakeAppDNSRoot)
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        rootLBFQDN,
			Type:        "rootlb",
			InfraFlavor: "m4.small",
			Status:      "ACTIVE",
		})
		FakeRamUsed += uint64(4096)
		FakeVcpusUsed += uint64(2)
		FakeDiskUsed += uint64(40)
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
				FakeRamUsed -= uint64(4096)
				FakeVcpusUsed -= uint64(2)
				FakeDiskUsed -= uint64(40)
				continue
			}
			newVMs = append(newVMs, vm)
		}
		delete(FakeClusterVMs, clusterInst.Key)
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "fake ClusterInst deleted")
	return nil
}

func (s *Platform) GetCloudletInfraResources(ctx context.Context) (*edgeproto.InfraResources, error) {
	var resources edgeproto.InfraResources
	platvm := edgeproto.VmInfo{
		Name:        "fake-platform-vm",
		Type:        "platform",
		InfraFlavor: "x1.small",
		Status:      "ACTIVE",
		Ipaddresses: []edgeproto.IpAddr{
			{ExternalIp: "10.101.100.10"},
		},
	}
	resources.Vms = append(resources.Vms, platvm)
	rlbvm := edgeproto.VmInfo{
		Name:        "fake-rootlb-vm",
		Type:        "rootlb",
		InfraFlavor: "x1.small",
		Status:      "ACTIVE",
		Ipaddresses: []edgeproto.IpAddr{
			{ExternalIp: "10.101.100.11"},
		},
	}
	resources.Vms = append(resources.Vms, rlbvm)

	resources.Info = []edgeproto.ResourceInfo{
		edgeproto.ResourceInfo{
			Name:     ResourceRam,
			Value:    FakeRamUsed,
			MaxValue: FakeRamMax,
			Units:    ResourceRamUnits,
		},
		edgeproto.ResourceInfo{
			Name:     ResourceVcpus,
			Value:    FakeVcpusUsed,
			MaxValue: FakeVcpusMax,
		},
		edgeproto.ResourceInfo{
			Name:     ResourceDisk,
			Value:    FakeDiskUsed,
			MaxValue: FakeDiskMax,
			Units:    ResourceDiskUnits,
		},
	}

	return &resources, nil
}

// called by controller, make sure it doesn't make any CRM specific calls
func (s *Platform) GetCloudletResourceInfo(ctx context.Context, vmResources []edgeproto.VMResource, infraResMap map[string]*edgeproto.ResourceInfo) map[string]*edgeproto.ResourceInfo {
	cloudletRes := map[string]string{
		ResourceRam:   ResourceRamUnits,
		ResourceVcpus: "",
		ResourceDisk:  ResourceDiskUnits,
	}
	resInfo := make(map[string]*edgeproto.ResourceInfo)
	for resName, resUnits := range cloudletRes {
		resMax := uint64(0)
		if infraRes, ok := infraResMap[resName]; ok {
			resMax = infraRes.MaxValue
		}
		resInfo[resName] = &edgeproto.ResourceInfo{
			Name:     resName,
			MaxValue: resMax,
			Units:    resUnits,
		}
	}

	for _, vmRes := range vmResources {
		if vmRes.VmFlavor != nil {
			if vmRes.ProvState == edgeproto.VMProvState_PROV_STATE_REMOVE {
				resInfo[ResourceRam].Value -= vmRes.VmFlavor.Ram
				resInfo[ResourceVcpus].Value -= vmRes.VmFlavor.Vcpus
				resInfo[ResourceDisk].Value -= vmRes.VmFlavor.Disk
			} else {
				resInfo[ResourceRam].Value += vmRes.VmFlavor.Ram
				resInfo[ResourceVcpus].Value += vmRes.VmFlavor.Vcpus
				resInfo[ResourceDisk].Value += vmRes.VmFlavor.Disk
			}
		}
	}
	return resInfo
}

func (s *Platform) ValidateCloudletResourceQuotas(ctx context.Context, resourceQuotas []edgeproto.ResourceQuota) error {
	validQuotas := map[string]uint64{
		ResourceRam:   FakeRamMax,
		ResourceVcpus: FakeVcpusMax,
		ResourceDisk:  FakeDiskMax,
	}
	quotaNames := []string{}
	for name, _ := range validQuotas {
		quotaNames = append(quotaNames, name)
	}
	for _, resQuota := range resourceQuotas {
		qMaxVal, ok := validQuotas[resQuota.Name]
		if !ok {
			return fmt.Errorf("Invalid resource quota name: %s, valid names are %s", resQuota.Name, strings.Join(quotaNames, ","))
		}
		if resQuota.Value > qMaxVal {
			return fmt.Errorf("Resource quota %s exceeded max supported value: %d", resQuota.Name, qMaxVal)
		}
	}
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
	return nil
}

func (s *Platform) DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "First Delete Task")
	updateCallback(edgeproto.UpdateTask, "Second Delete Task")
	log.SpanLog(ctx, log.DebugLevelInfra, "fake AppInst deleted")
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
	err := s.ValidateCloudletResourceQuotas(ctx, cloudlet.ResourceQuotas)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Creating Cloudlet")
	updateCallback(edgeproto.UpdateTask, "Starting CRMServer")
	err = cloudcommon.StartCRMService(ctx, cloudlet, pfConfig)
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

func (s *Platform) GetCloudletResourceQuotaProps(ctx context.Context) (*edgeproto.CloudletResourceQuotaProps, error) {
	return &edgeproto.CloudletResourceQuotaProps{
		Props: []edgeproto.ResourceInfo{
			edgeproto.ResourceInfo{
				Name:        ResourceRam,
				Description: "Limit on RAM available (MB)",
			},
			edgeproto.ResourceInfo{
				Name:        ResourceVcpus,
				Description: "Limit on vCPUs available",
			},
			edgeproto.ResourceInfo{
				Name:        ResourceDisk,
				Description: "Limit on disk available (GB)",
			},
		},
	}, nil
}

func (s *Platform) GetCloudletResourceMetric(ctx context.Context, key *edgeproto.CloudletKey, resources []edgeproto.VMResource) (*edgeproto.Metric, error) {
	resMetric := edgeproto.Metric{}
	if len(resources) == 0 {
		return nil, fmt.Errorf("missing resources")
	}

	ramUsed := uint64(0)
	vcpusUsed := uint64(0)
	diskUsed := uint64(0)
	for _, vmRes := range resources {
		if vmRes.VmFlavor != nil {
			ramUsed += vmRes.VmFlavor.Ram
			vcpusUsed += vmRes.VmFlavor.Vcpus
			diskUsed += vmRes.VmFlavor.Disk
		}
	}
	resMetric.Name = "fake-resource-usage"
	ts, _ := types.TimestampProto(time.Now())
	resMetric.Timestamp = *ts
	resMetric.AddTag("cloudletorg", key.Organization)
	resMetric.AddTag("cloudlet", key.Name)
	resMetric.AddIntVal("ramUsed", ramUsed)
	resMetric.AddIntVal("vcpusUsed", vcpusUsed)
	resMetric.AddIntVal("diskUsed", diskUsed)
	return &resMetric, nil
}
