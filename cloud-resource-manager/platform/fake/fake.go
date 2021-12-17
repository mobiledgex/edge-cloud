package fake

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
)

type Platform struct {
	consoleServer *httptest.Server
	caches        *platform.Caches
	clusterTPEs   map[cloudcommon.TrustPolicyExceptionKeyClusterInstKey]struct{}
	mux           sync.Mutex
	cloudletKey   *edgeproto.CloudletKey
}

var (
	ResourceAdd    = true
	ResourceRemove = false

	FakeRamUsed         = uint64(0)
	FakeVcpusUsed       = uint64(0)
	FakeDiskUsed        = uint64(0)
	FakeExternalIpsUsed = uint64(0)
	FakeInstancesUsed   = uint64(0)

	FakeRamMax         = uint64(40960)
	FakeVcpusMax       = uint64(50)
	FakeDiskMax        = uint64(5000)
	FakeExternalIpsMax = uint64(30)
)

var FakeAppDNSRoot = "fake.net"

var FakeClusterVMs = map[edgeproto.ClusterInstKey][]edgeproto.VmInfo{}

var FakeFlavorList = []*edgeproto.FlavorInfo{
	&edgeproto.FlavorInfo{
		Name:  "x1.tiny",
		Vcpus: uint64(1),
		Ram:   uint64(1024),
		Disk:  uint64(20),
	},
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

var maxPrimaryCrmStartupWait = 10 * time.Second

func UpdateResourcesMax() error {
	// Make fake resource limits configurable for QA testing
	ramMax := os.Getenv("FAKE_RAM_MAX")
	if ramMax != "" {
		ram, err := strconv.Atoi(ramMax)
		if err != nil {
			return err
		}
		if ram > 0 {
			FakeRamMax = uint64(ram)
		}
	}
	vcpusMax := os.Getenv("FAKE_VCPUS_MAX")
	if vcpusMax != "" {
		vcpus, err := strconv.Atoi(vcpusMax)
		if err != nil {
			return err
		}
		if vcpus > 0 {
			FakeVcpusMax = uint64(vcpus)
		}
	}
	diskMax := os.Getenv("FAKE_DISK_MAX")
	if diskMax != "" {
		disk, err := strconv.Atoi(diskMax)
		if err != nil {
			return err
		}
		if disk > 0 {
			FakeDiskMax = uint64(disk)
		}
	}
	return nil
}

func (s *Platform) Init(ctx context.Context, platformConfig *platform.PlatformConfig, caches *platform.Caches, platformActive bool, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "running in fake cloudlet mode")
	platformConfig.NodeMgr.Debug.AddDebugFunc("fakecmd", s.runDebug)

	s.caches = caches
	s.cloudletKey = platformConfig.CloudletKey
	updateCallback(edgeproto.UpdateTask, "Done initializing fake platform")
	s.consoleServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Console Content")
	}))
	// Update resource info for platformVM and RootLBVM
	FakeRamUsed += 4096 + 4096
	FakeVcpusUsed += 2 + 2
	FakeDiskUsed += 40 + 40
	FakeExternalIpsUsed += 1
	FakeInstancesUsed += 2

	err := UpdateResourcesMax()
	if err != nil {
		return err
	}
	s.clusterTPEs = make(map[cloudcommon.TrustPolicyExceptionKeyClusterInstKey]struct{})

	return nil
}

func (s *Platform) GetFeatures() *platform.Features {
	return &platform.Features{
		SupportsMultiTenantCluster:       true,
		SupportsSharedVolume:             true,
		SupportsTrustPolicy:              true,
		CloudletServicesLocal:            true,
		IsFake:                           true,
		SupportsAdditionalNetworks:       true,
		SupportsPlatformHighAvailability: true,
	}
}

func (s *Platform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	info.OsMaxRam = FakeRamMax
	info.OsMaxVcores = FakeVcpusMax
	info.OsMaxVolGb = FakeDiskMax
	info.Flavors = FakeFlavorList
	return nil
}

func UpdateCommonResourcesUsed(flavor string, add bool) {
	if flavor == "x1.tiny" {
		if add {
			FakeRamUsed += FakeFlavorList[0].Ram
			FakeVcpusUsed += FakeFlavorList[0].Vcpus
			FakeDiskUsed += FakeFlavorList[0].Disk
		} else {
			FakeRamUsed -= FakeFlavorList[0].Ram
			FakeVcpusUsed -= FakeFlavorList[0].Vcpus
			FakeDiskUsed -= FakeFlavorList[0].Disk
		}
	} else {
		if add {
			FakeRamUsed += FakeFlavorList[1].Ram
			FakeVcpusUsed += FakeFlavorList[1].Vcpus
			FakeDiskUsed += FakeFlavorList[1].Disk
		} else {
			FakeRamUsed -= FakeFlavorList[1].Ram
			FakeVcpusUsed -= FakeFlavorList[1].Vcpus
			FakeDiskUsed -= FakeFlavorList[1].Disk
		}
	}
	if add {
		FakeInstancesUsed += 1
	} else {
		FakeInstancesUsed -= 1
	}
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
	newcVMs := []edgeproto.VmInfo{}
	for _, vmInfo := range cVMs {
		if vmInfo.Type == cloudcommon.VMTypeClusterK8sNode {
			if _, ok := fakeNodes[vmInfo.Name]; ok {
				delete(fakeNodes, vmInfo.Name)
				newcVMs = append(newcVMs, vmInfo)
			} else {
				UpdateCommonResourcesUsed(clusterInst.NodeFlavor, ResourceRemove)
			}
		} else if vmInfo.Type == cloudcommon.VMTypeClusterMaster {
			if _, ok := fakeMasters[vmInfo.Name]; ok {
				delete(fakeMasters, vmInfo.Name)
				newcVMs = append(newcVMs, vmInfo)
			} else {
				UpdateCommonResourcesUsed(clusterInst.MasterNodeFlavor, ResourceRemove)
			}
		} else {
			// rootlb
			newcVMs = append(newcVMs, vmInfo)
		}
	}
	FakeClusterVMs[clusterInst.Key] = newcVMs
	for vmName, _ := range fakeNodes {
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        vmName,
			Type:        cloudcommon.VMTypeClusterK8sNode,
			InfraFlavor: clusterInst.NodeFlavor,
			Status:      "ACTIVE",
		})
		UpdateCommonResourcesUsed(clusterInst.NodeFlavor, ResourceAdd)
	}
	for vmName, _ := range fakeMasters {
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        vmName,
			Type:        cloudcommon.VMTypeClusterMaster,
			InfraFlavor: clusterInst.MasterNodeFlavor,
			Status:      "ACTIVE",
		})
		UpdateCommonResourcesUsed(clusterInst.MasterNodeFlavor, ResourceAdd)
	}
	return nil
}

func updateClusterResCount(clusterInst *edgeproto.ClusterInst) {
	vmNameSuffix := k8smgmt.GetCloudletClusterName(&clusterInst.Key)
	if len(FakeClusterVMs) == 0 {
		FakeClusterVMs = make(map[edgeproto.ClusterInstKey][]edgeproto.VmInfo)
	}
	if _, ok := FakeClusterVMs[clusterInst.Key]; !ok {
		FakeClusterVMs[clusterInst.Key] = []edgeproto.VmInfo{}
	}
	for ii := uint32(0); ii < clusterInst.NumMasters; ii++ {
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        fmt.Sprintf("fake-master-%d-%s", ii+1, vmNameSuffix),
			Type:        cloudcommon.VMTypeClusterMaster,
			InfraFlavor: clusterInst.MasterNodeFlavor,
			Status:      "ACTIVE",
		})
		UpdateCommonResourcesUsed(clusterInst.MasterNodeFlavor, ResourceAdd)
	}
	for ii := uint32(0); ii < clusterInst.NumNodes; ii++ {
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        fmt.Sprintf("fake-node-%d-%s", ii+1, vmNameSuffix),
			Type:        cloudcommon.VMTypeClusterK8sNode,
			InfraFlavor: clusterInst.NodeFlavor,
			Status:      "ACTIVE",
		})
		UpdateCommonResourcesUsed(clusterInst.NodeFlavor, ResourceAdd)
	}
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		rootLBFQDN := cloudcommon.GetDedicatedLBFQDN(&clusterInst.Key.CloudletKey, &clusterInst.Key.ClusterKey, FakeAppDNSRoot)
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        rootLBFQDN,
			Type:        cloudcommon.VMTypeRootLB,
			InfraFlavor: "x1.small",
			Status:      "ACTIVE",
		})
		UpdateCommonResourcesUsed("x1.small", ResourceAdd)
		FakeExternalIpsUsed += 1
	}
}

func updateVmAppResCount(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) {
	if app.Deployment == cloudcommon.DeploymentTypeVM {
		appFQN := cloudcommon.GetAppFQN(&app.Key)
		clusterInst.Key.ClusterKey.Name = appFQN + "-" + appInst.Key.ClusterInstKey.ClusterKey.Name
		if len(FakeClusterVMs) == 0 {
			FakeClusterVMs = make(map[edgeproto.ClusterInstKey][]edgeproto.VmInfo)
		}
		if _, ok := FakeClusterVMs[clusterInst.Key]; !ok {
			FakeClusterVMs[clusterInst.Key] = []edgeproto.VmInfo{}
		}
		FakeClusterVMs[clusterInst.Key] = append(FakeClusterVMs[clusterInst.Key], edgeproto.VmInfo{
			Name:        appFQN,
			Type:        cloudcommon.VMTypeAppVM,
			InfraFlavor: appInst.VmFlavor,
			Status:      "ACTIVE",
		})
		UpdateCommonResourcesUsed(appInst.VmFlavor, ResourceAdd)
		FakeExternalIpsUsed += 1 // VMApp create a dedicated LB that consumes one IP
	}
}

func (s *Platform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "fake CreateClusterInst", "clusterInst", clusterInst)
	updateCallback(edgeproto.UpdateTask, "First Create Task")
	updateCallback(edgeproto.UpdateTask, "Second Create Task")
	updateClusterResCount(clusterInst)

	// verify we can find any provisioned networks
	if len(clusterInst.Networks) > 0 {
		networks, err := crmutil.GetNetworksForClusterInst(ctx, clusterInst, s.caches.NetworkCache)
		if err != nil {
			return fmt.Errorf("Error getting cluster networks - %v", err)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "found networks from cache", "networks", networks)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "fake ClusterInst ready")
	return nil
}

func (s *Platform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "First Delete Task")
	updateCallback(edgeproto.UpdateTask, "Second Delete Task")
	rootLBFQDN := cloudcommon.GetDedicatedLBFQDN(&clusterInst.Key.CloudletKey, &clusterInst.Key.ClusterKey, FakeAppDNSRoot)
	vms := make(map[string]string)
	vms[rootLBFQDN] = "x1.small"
	vmNameSuffix := k8smgmt.GetCloudletClusterName(&clusterInst.Key)
	for ii := uint32(0); ii < clusterInst.NumMasters; ii++ {
		vmName := fmt.Sprintf("fake-master-%d-%s", ii+1, vmNameSuffix)
		vms[vmName] = clusterInst.MasterNodeFlavor
	}
	for ii := uint32(0); ii < clusterInst.NumNodes; ii++ {
		vmName := fmt.Sprintf("fake-node-%d-%s", ii+1, vmNameSuffix)
		vms[vmName] = clusterInst.NodeFlavor
	}
	if clusterVMs, ok := FakeClusterVMs[clusterInst.Key]; ok {
		for _, vm := range clusterVMs {
			if vmFlavor, ok := vms[vm.Name]; ok {
				UpdateCommonResourcesUsed(vmFlavor, ResourceRemove)
				if vm.Name == rootLBFQDN {
					FakeExternalIpsUsed -= 1
				}
				continue
			}
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
			Name:          cloudcommon.ResourceExternalIPs,
			Value:         FakeExternalIpsUsed,
			InfraMaxValue: FakeExternalIpsMax,
		},
		edgeproto.InfraResource{
			Name:  cloudcommon.ResourceInstances,
			Value: FakeInstancesUsed,
		},
	}

	return &resources, nil
}

// called by controller, make sure it doesn't make any calls to infra API
func (s *Platform) GetClusterAdditionalResources(ctx context.Context, cloudlet *edgeproto.Cloudlet, vmResources []edgeproto.VMResource, infraResMap map[string]edgeproto.InfraResource) map[string]edgeproto.InfraResource {
	// resource name -> resource units
	cloudletRes := map[string]string{
		cloudcommon.ResourceInstances: "",
	}
	resInfo := make(map[string]edgeproto.InfraResource)
	for resName, resUnits := range cloudletRes {
		resMax := uint64(0)
		if infraRes, ok := infraResMap[resName]; ok {
			resMax = infraRes.InfraMaxValue
		}
		resInfo[resName] = edgeproto.InfraResource{
			Name:          resName,
			InfraMaxValue: resMax,
			Units:         resUnits,
		}
	}

	out, ok := resInfo[cloudcommon.ResourceInstances]
	if ok {
		out.Value += uint64(len(vmResources))
		resInfo[cloudcommon.ResourceInstances] = out
	}
	return resInfo
}

func (s *Platform) GetCloudletResourceQuotaProps(ctx context.Context) (*edgeproto.CloudletResourceQuotaProps, error) {
	return &edgeproto.CloudletResourceQuotaProps{
		Properties: []edgeproto.InfraResource{
			edgeproto.InfraResource{
				Name:        cloudcommon.ResourceInstances,
				Description: cloudcommon.ResourceQuotaDesc[cloudcommon.ResourceInstances],
			},
		},
	}, nil
}

func (s *Platform) GetClusterAdditionalResourceMetric(ctx context.Context, cloudlet *edgeproto.Cloudlet, resMetric *edgeproto.Metric, resources []edgeproto.VMResource) error {
	instancesUsed := uint64(len(resources))
	resMetric.AddIntVal(cloudcommon.ResourceMetricInstances, instancesUsed)
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
	clusterSvcAppInstFail := os.Getenv("FAKE_PLATFORM_APPINST_CREATE_FAIL")
	if clusterSvcAppInstFail != "" {
		return errors.New("FAKE_PLATFORM_APPINST_CREATE_FAIL")
	}
	updateCallback(edgeproto.UpdateTask, "Creating App Inst")
	log.SpanLog(ctx, log.DebugLevelInfra, "fake AppInst ready")
	updateVmAppResCount(ctx, clusterInst, app, appInst)
	return nil
}

func (s *Platform) DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "First Delete Task")
	updateCallback(edgeproto.UpdateTask, "Second Delete Task")
	log.SpanLog(ctx, log.DebugLevelInfra, "fake AppInst deleted")
	if app.Deployment == cloudcommon.DeploymentTypeVM {
		appFQN := cloudcommon.GetAppFQN(&app.Key)
		clusterInst.Key.ClusterKey.Name = appFQN + "-" + appInst.Key.ClusterInstKey.ClusterKey.Name
		UpdateCommonResourcesUsed(appInst.VmFlavor, ResourceRemove)
		if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_DIRECT {
			FakeExternalIpsUsed -= 1
		}
		delete(FakeClusterVMs, clusterInst.Key)
	}
	return nil
}

func (s *Platform) UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
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

func (s *Platform) ListCloudletMgmtNodes(ctx context.Context, clusterInsts []edgeproto.ClusterInst, vmAppInsts []edgeproto.AppInst) ([]edgeproto.CloudletMgmtNode, error) {
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

func (s *Platform) GetConsoleUrl(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst) (string, error) {
	if s.consoleServer != nil {
		return s.consoleServer.URL + "?token=xyz", nil
	}
	return "", fmt.Errorf("no console server to fetch URL from")
}

func (s *Platform) CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, caches *platform.Caches, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) (bool, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "create fake cloudlet", "key", cloudlet.Key)
	if cloudlet.InfraApiAccess == edgeproto.InfraApiAccess_RESTRICTED_ACCESS {
		return true, nil
	}
	updateCallback(edgeproto.UpdateTask, "Creating Cloudlet")
	updateCallback(edgeproto.UpdateTask, "Starting CRMServer")
	redisAddr := ""
	if cloudlet.PlatformHighAvailability {
		redisAddr = process.LocalRedisAddr
	}
	err := cloudcommon.StartCRMService(ctx, cloudlet, pfConfig, process.HARolePrimary, redisAddr)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "fake cloudlet create failed to start CRM", "err", err)
		return true, err
	}
	if cloudlet.PlatformHighAvailability {
		log.SpanLog(ctx, log.DebugLevelInfra, "creating 2 instances for H/A", "key", cloudlet.Key)
		// Pause before starting the secondary to let the primary become active first for the sake of
		// e2e tests that need consistent ordering. Secondary will be started up in a separate thread after
		// cloudletInfo shows up from the primary
		go func() {
			start := time.Now()
			var err error
			var cloudletInfo edgeproto.CloudletInfo
			for {
				time.Sleep(time.Millisecond * 200)
				elapsed := time.Since(start)
				if elapsed >= (maxPrimaryCrmStartupWait) {
					log.SpanLog(ctx, log.DebugLevelInfra, "timed out waiting for primary CRM to report cloudlet info")
					err = fmt.Errorf("timed out waiting for primary CRM to report cloudlet info")
					break
				}
				if !caches.CloudletInfoCache.Get(&cloudlet.Key, &cloudletInfo) {
					log.SpanLog(ctx, log.DebugLevelInfra, "failed to get cloudlet info after starting primary CRM, will retry", "cloudletKey", s.cloudletKey)
				} else {
					log.SpanLog(ctx, log.DebugLevelInfra, "got cloudlet info from primary CRM, will start secondary", "cloudletKey", cloudlet.Key, "active", cloudletInfo.ActiveCrmInstance, "ci", cloudletInfo)
					err = cloudcommon.StartCRMService(ctx, cloudlet, pfConfig, process.HARoleSecondary, redisAddr)
					if err != nil {
						log.SpanLog(ctx, log.DebugLevelInfra, "fake cloudlet create failed to start secondary CRM", "err", err)
					}
					break
				}
			}
			if err != nil {
				cloudletInfo.Key = cloudlet.Key
				cloudletInfo.Errors = append(cloudletInfo.Errors, "fake cloudlet create failed to start secondary CRM: "+err.Error())
				cloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_ERRORS
				caches.CloudletInfoCache.Update(ctx, &cloudletInfo, 0)
			}
		}()
	}
	return true, nil
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

func (s *Platform) UpdateTrustPolicyException(ctx context.Context, tpe *edgeproto.TrustPolicyException, clusterInstKey *edgeproto.ClusterInstKey) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	key := cloudcommon.TrustPolicyExceptionKeyClusterInstKey{
		TpeKey:         tpe.Key,
		ClusterInstKey: *clusterInstKey,
	}
	s.clusterTPEs[key] = struct{}{}
	log.SpanLog(ctx, log.DebugLevelInfra, "fake UpdateTrustPolicyException", "ADD_TPE policyKey", key)
	return nil
}

func (s *Platform) DeleteTrustPolicyException(ctx context.Context, tpeKey *edgeproto.TrustPolicyExceptionKey, clusterInstKey *edgeproto.ClusterInstKey) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	key := cloudcommon.TrustPolicyExceptionKeyClusterInstKey{
		TpeKey:         *tpeKey,
		ClusterInstKey: *clusterInstKey,
	}
	delete(s.clusterTPEs, key)
	log.SpanLog(ctx, log.DebugLevelInfra, "fake DeleteTrustPolicyException", "DELETE_TPE policyKey", key)
	return nil
}

func (s *Platform) HasTrustPolicyException(ctx context.Context, tpeKey *edgeproto.TrustPolicyExceptionKey, clusterInst *edgeproto.ClusterInst) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	key := cloudcommon.TrustPolicyExceptionKeyClusterInstKey{
		TpeKey:         *tpeKey,
		ClusterInstKey: clusterInst.Key,
	}
	_, found := s.clusterTPEs[key]
	log.SpanLog(ctx, log.DebugLevelInfra, "fake HasTrustPolicyException", "policyKey", tpeKey, "found", found)

	return found
}

func (s *Platform) TrustPolicyExceptionCount(ctx context.Context) int {
	s.mux.Lock()
	defer s.mux.Unlock()
	count := len(s.clusterTPEs)
	log.SpanLog(ctx, log.DebugLevelInfra, "fake TrustPolicyExceptionCount", "count", count)
	return count
}

func (s *Platform) DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, caches *platform.Caches, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelInfra, "delete fake Cloudlet", "key", cloudlet.Key)
	updateCallback(edgeproto.UpdateTask, "Deleting Cloudlet")
	updateCallback(edgeproto.UpdateTask, "Stopping CRMServer")
	err := cloudcommon.StopCRMService(ctx, cloudlet, process.HARoleAll)
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
	if caches == nil {
		return fmt.Errorf("caches is nil")
	}
	// Because the fake cloudlet doesn't have it's own internal database of
	// allocated objects like Openstack/VMWare, we just fake it by copying
	// what the Controller says is supposed to be here. This handles the CRM
	// restart case.
	clusterInstKeys := []edgeproto.ClusterInstKey{}
	caches.ClusterInstCache.GetAllKeys(ctx, func(k *edgeproto.ClusterInstKey, modRev int64) {
		clusterInstKeys = append(clusterInstKeys, *k)
	})
	for _, k := range clusterInstKeys {
		var clusterInst edgeproto.ClusterInst
		if caches.ClusterInstCache.Get(&k, &clusterInst) {
			updateClusterResCount(&clusterInst)
		}
	}

	appInstKeys := []edgeproto.AppInstKey{}
	caches.AppInstCache.GetAllKeys(ctx, func(k *edgeproto.AppInstKey, modRev int64) {
		appInstKeys = append(appInstKeys, *k)
	})
	for _, k := range appInstKeys {
		var appInst edgeproto.AppInst
		if caches.AppInstCache.Get(&k, &appInst) {
			var app edgeproto.App
			if caches.AppCache.Get(&k.AppKey, &app) {
				if app.Deployment == cloudcommon.DeploymentTypeVM {
					var clusterInst = edgeproto.ClusterInst{
						Key: *appInst.ClusterInstKey(),
					}
					updateVmAppResCount(ctx, &clusterInst, &app, &appInst)
				}
			}
		}
	}
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

func (s *Platform) GetVersionProperties() map[string]string {
	return map[string]string{}
}

func (s *Platform) GetRootLBFlavor(ctx context.Context) (*edgeproto.Flavor, error) {
	return &RootLBFlavor, nil
}

func (s *Platform) ActiveChanged(ctx context.Context, platformActive bool) {
	log.SpanLog(ctx, log.DebugLevelInfra, "ActiveChanged", "platformActive", platformActive)
}
