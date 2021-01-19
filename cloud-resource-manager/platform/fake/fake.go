package fake

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
)

type Platform struct {
	consoleServer *httptest.Server
}

const (
	RamUsed   = "totalRAMUsed"
	RamMax    = "maxTotalRAMSize"
	VcpusUsed = "totalCoresUsed"
	VcpusMax  = "maxTotalCores"
	DiskUsed  = "totalGigabytesUsed"
	DiskMax   = "maxTotalVolumeGigabytes"
)

type FakeResource struct {
	RamUsed   uint64
	RamMax    uint64
	VcpusUsed uint64
	VcpusMax  uint64
	DiskUsed  uint64
	DiskMax   uint64
}

var FakeRamUsed = uint64(0)
var FakeVcpusUsed = uint64(0)
var FakeDiskUsed = uint64(0)

var FakeAppDNSRoot = "fake.net"

var FakeClusterVMs = []edgeproto.VmInfo{}

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
	info.OsMaxRam = 40960
	info.OsMaxVcores = 50
	info.OsMaxVolGb = 5000
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
	for ii := uint32(0); ii < clusterInst.NumMasters; ii++ {
		FakeClusterVMs = append(FakeClusterVMs, edgeproto.VmInfo{
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
		FakeClusterVMs = append(FakeClusterVMs, edgeproto.VmInfo{
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
		FakeClusterVMs = append(FakeClusterVMs, edgeproto.VmInfo{
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
	for _, vm := range FakeClusterVMs {
		if _, ok := vms[vm.Name]; ok {
			FakeRamUsed -= uint64(4096)
			FakeVcpusUsed -= uint64(2)
			FakeDiskUsed -= uint64(40)
			continue
		}
		newVMs = append(newVMs, vm)
	}
	FakeClusterVMs = newVMs

	log.SpanLog(ctx, log.DebugLevelInfra, "fake ClusterInst deleted")
	return nil
}

func MarshalInfraResourceInfo(fakeRes *FakeResource) []edgeproto.ResourceInfo {
	return []edgeproto.ResourceInfo{
		edgeproto.ResourceInfo{
			Name:  RamMax,
			Value: fmt.Sprintf("%d", fakeRes.RamMax),
		},
		edgeproto.ResourceInfo{
			Name:  RamUsed,
			Value: fmt.Sprintf("%d", fakeRes.RamUsed),
		},
		edgeproto.ResourceInfo{
			Name:  VcpusMax,
			Value: fmt.Sprintf("%d", fakeRes.VcpusMax),
		},
		edgeproto.ResourceInfo{
			Name:  VcpusUsed,
			Value: fmt.Sprintf("%d", fakeRes.VcpusUsed),
		},
		edgeproto.ResourceInfo{
			Name:  DiskMax,
			Value: fmt.Sprintf("%d", fakeRes.DiskMax),
		},
		edgeproto.ResourceInfo{
			Name:  DiskUsed,
			Value: fmt.Sprintf("%d", fakeRes.DiskUsed),
		},
	}
}

func UnmarshalInfraResourceInfo(cloudletResInfo []edgeproto.ResourceInfo) (*FakeResource, error) {
	fakeRes := FakeResource{}
	for _, resInfo := range cloudletResInfo {
		switch resInfo.Name {
		case RamMax:
			maxRam, err := strconv.ParseUint(resInfo.Value, 0, 64)
			if err != nil {
				return nil, err
			}
			fakeRes.RamMax = maxRam
		case RamUsed:
			ramUsed, err := strconv.ParseUint(resInfo.Value, 0, 64)
			if err != nil {
				return nil, err
			}
			fakeRes.RamUsed = ramUsed
		case VcpusMax:
			maxVcpus, err := strconv.ParseUint(resInfo.Value, 0, 64)
			if err != nil {
				return nil, err
			}
			fakeRes.VcpusMax = maxVcpus
		case VcpusUsed:
			vcpusUsed, err := strconv.ParseUint(resInfo.Value, 0, 64)
			if err != nil {
				return nil, err
			}
			fakeRes.VcpusUsed = vcpusUsed
		case DiskMax:
			maxDisk, err := strconv.ParseUint(resInfo.Value, 0, 64)
			if err != nil {
				return nil, err
			}
			fakeRes.DiskMax = maxDisk
		case DiskUsed:
			diskUsed, err := strconv.ParseUint(resInfo.Value, 0, 64)
			if err != nil {
				return nil, err
			}
			fakeRes.DiskUsed = diskUsed
		}
	}
	return &fakeRes, nil
}

func (s *Platform) GetCloudletInfraResources(ctx context.Context) (*edgeproto.InfraResources, []string, error) {
	var resources edgeproto.InfraResources
	platvm := edgeproto.VmInfo{
		Name:        "fake-platform-vm",
		Type:        "platform",
		InfraFlavor: "m4.small",
		Status:      "ACTIVE",
		Ipaddresses: []edgeproto.IpAddr{
			{ExternalIp: "10.101.100.10"},
		},
	}
	resources.Vms = append(resources.Vms, platvm)
	rlbvm := edgeproto.VmInfo{
		Name:        "fake-rootlb-vm",
		Type:        "rootlb",
		InfraFlavor: "m4.small",
		Status:      "ACTIVE",
		Ipaddresses: []edgeproto.IpAddr{
			{ExternalIp: "10.101.100.11"},
		},
	}
	resources.Vms = append(resources.Vms, rlbvm)
	resources.Vms = append(resources.Vms, FakeClusterVMs...)

	resources.Info = MarshalInfraResourceInfo(&FakeResource{
		RamUsed:   FakeRamUsed,
		RamMax:    uint64(40960),
		VcpusUsed: FakeVcpusUsed,
		VcpusMax:  uint64(50),
		DiskUsed:  FakeDiskUsed,
		DiskMax:   uint64(5000),
	})

	warnings := []string{}

	return &resources, warnings, nil
}

func (s *Platform) GetCloudletResourceUsage(ctx context.Context, resInfo []edgeproto.ResourceInfo, existingVmResources []edgeproto.VMResource) ([]edgeproto.ResourceInfo, error) {
	fakeRes, err := UnmarshalInfraResourceInfo(resInfo)
	if err != nil {
		return nil, err
	}
	for _, vmRes := range existingVmResources {
		if vmRes.VmFlavor != nil {
			if vmRes.ProvState == edgeproto.VMProvState_PROV_STATE_REMOVE {
				fakeRes.RamUsed -= vmRes.VmFlavor.Ram
				fakeRes.VcpusUsed -= vmRes.VmFlavor.Vcpus
				fakeRes.DiskUsed -= vmRes.VmFlavor.Disk
			} else {
				fakeRes.RamUsed += vmRes.VmFlavor.Ram
				fakeRes.VcpusUsed += vmRes.VmFlavor.Vcpus
				fakeRes.DiskUsed += vmRes.VmFlavor.Disk
			}
		}
	}
	resInfo = MarshalInfraResourceInfo(fakeRes)
	return resInfo, nil
}

func (s *Platform) ValidateCloudletResources(ctx context.Context, infraResources *edgeproto.InfraResources, vmResources []edgeproto.VMResource, existingVmResources []edgeproto.VMResource) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Validate cloudlet resources", "vm resources", vmResources, "cloudlet resouurces", infraResources)
	ramReqd := uint64(0)
	vcpusReqd := uint64(0)
	diskReqd := uint64(0)
	for _, vmRes := range vmResources {
		if vmRes.VmFlavor != nil {
			ramReqd += vmRes.VmFlavor.Ram
			vcpusReqd += vmRes.VmFlavor.Vcpus
			diskReqd += vmRes.VmFlavor.Disk
		}
	}

	fakeRes, err := UnmarshalInfraResourceInfo(infraResources.Info)
	if err != nil {
		return err
	}

	for _, vmRes := range existingVmResources {
		if vmRes.VmFlavor != nil {
			if vmRes.ProvState == edgeproto.VMProvState_PROV_STATE_REMOVE {
				fakeRes.RamUsed -= vmRes.VmFlavor.Ram
				fakeRes.VcpusUsed -= vmRes.VmFlavor.Vcpus
				fakeRes.DiskUsed -= vmRes.VmFlavor.Disk
			} else {
				fakeRes.RamUsed += vmRes.VmFlavor.Ram
				fakeRes.VcpusUsed += vmRes.VmFlavor.Vcpus
				fakeRes.DiskUsed += vmRes.VmFlavor.Disk
			}
		}
	}

	availableRam := fakeRes.RamMax - fakeRes.RamUsed
	availableVcpus := fakeRes.VcpusMax - fakeRes.VcpusUsed
	availableDisk := fakeRes.DiskMax - fakeRes.DiskUsed

	if ramReqd > availableRam {
		return fmt.Errorf("Not enough RAM available, required %dMB but only %dMB is available", ramReqd, availableRam)
	}
	if vcpusReqd > availableVcpus {
		return fmt.Errorf("Not enough Vcpus available, required %d but only %d is available", vcpusReqd, availableVcpus)
	}
	if diskReqd > availableDisk {
		return fmt.Errorf("Not enough Disk available, required %dGB but only %dGB is available", diskReqd, availableDisk)
	}
	return nil
}

func (s *Platform) GetClusterInfraResources(ctx context.Context, clusterKey *edgeproto.ClusterInstKey) (*edgeproto.InfraResources, error) {
	var resources edgeproto.InfraResources
	vmtype := "cluster-master"
	for i := 0; i < 3; i++ {
		if i > 1 {
			vmtype = "cluster-node"
		}
		ipstr := fmt.Sprintf("10.100.100.1%d", i)
		vm := edgeproto.VmInfo{
			Name:        fmt.Sprintf("fake-cluster-vm-%d", i),
			Type:        vmtype,
			InfraFlavor: "m4.small",
			Status:      "ACTIVE",
			Ipaddresses: []edgeproto.IpAddr{
				{ExternalIp: ipstr},
			},
		}
		resources.Vms = append(resources.Vms, vm)
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

func (s *Platform) SyncControllerCache(ctx context.Context, caches *platform.Caches, cloudletState edgeproto.CloudletState) error {
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
