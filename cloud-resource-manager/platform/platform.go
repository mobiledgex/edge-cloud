package platform

import (
	"context"
	"strings"
	"sync"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/chefmgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
)

type PlatformConfig struct {
	CloudletKey         *edgeproto.CloudletKey
	PhysicalName        string
	Region              string
	TestMode            bool
	CloudletVMImagePath string
	VMImageVersion      string
	PackageVersion      string
	EnvVars             map[string]string
	NodeMgr             *node.NodeMgr
	AppDNSRoot          string
	ChefServerPath      string
	DeploymentTag       string
	Upgrade             bool
	AccessApi           AccessApi
	TrustPolicy         string
	CacheDir            string
	GPUConfig           *edgeproto.GPUConfig
}

type Caches struct {
	SettingsCache             *edgeproto.SettingsCache
	FlavorCache               *edgeproto.FlavorCache
	TrustPolicyCache          *edgeproto.TrustPolicyCache
	TrustPolicyExceptionCache *edgeproto.TrustPolicyExceptionCache
	CloudletPoolCache         *edgeproto.CloudletPoolCache
	ClusterInstCache          *edgeproto.ClusterInstCache
	AppInstCache              *edgeproto.AppInstCache
	AppCache                  *edgeproto.AppCache
	ResTagTableCache          *edgeproto.ResTagTableCache
	CloudletCache             *edgeproto.CloudletCache
	CloudletInternalCache     *edgeproto.CloudletInternalCache
	VMPoolCache               *edgeproto.VMPoolCache
	VMPoolInfoCache           *edgeproto.VMPoolInfoCache
	GPUDriverCache            *edgeproto.GPUDriverCache
	NetworkCache              *edgeproto.NetworkCache
	CloudletInfoCache         *edgeproto.CloudletInfoCache
	// VMPool object managed by CRM
	VMPool    *edgeproto.VMPool
	VMPoolMux *sync.Mutex
}

// Features that the platform supports or enables
type Features struct {
	SupportsMultiTenantCluster       bool
	SupportsSharedVolume             bool
	SupportsTrustPolicy              bool
	SupportsKubernetesOnly           bool // does not support docker/VM
	KubernetesRequiresWorkerNodes    bool // k8s cluster cannot be master only
	CloudletServicesLocal            bool // cloudlet services running locally to controller
	IPAllocatedPerService            bool // Every k8s service gets a public IP (GCP/etc)
	SupportsImageTypeOVF             bool // Supports OVF images for VM deployments
	IsVMPool                         bool // cloudlet is just a pool of pre-existing VMs
	IsFake                           bool // Just for unit-testing/e2e-testing
	SupportsAdditionalNetworks       bool // Additional networks can be added
	IsSingleKubernetesCluster        bool // Entire platform is just a single K8S cluster
	SupportsAppInstDedicatedIP       bool // Supports per AppInst dedicated IPs
	SupportsPlatformHighAvailability bool // Supports High Availablity with 2 CRMs
}

// Platform abstracts the underlying cloudlet platform.
type Platform interface {
	// GetVersionProperties returns properties related to the platform version
	GetVersionProperties() map[string]string
	// Get platform features
	GetFeatures() *Features
	// Init is called once during CRM startup.
	Init(ctx context.Context, platformConfig *PlatformConfig, caches *Caches, platformActive bool, updateCallback edgeproto.CacheUpdateCallback) error
	// Gather information about the cloudlet platform.
	// This includes available resources, flavors, etc.
	// Returns true if sync with controller is required
	GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error
	// Create a Kubernetes Cluster on the cloudlet.
	CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error
	// Delete a Kuberentes Cluster on the cloudlet.
	DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Update the cluster
	UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Get resources used by the cloudlet
	GetCloudletInfraResources(ctx context.Context) (*edgeproto.InfraResourcesSnapshot, error)
	// Get cluster additional resources used by the vms specific to the platform
	GetClusterAdditionalResources(ctx context.Context, cloudlet *edgeproto.Cloudlet, vmResources []edgeproto.VMResource, infraResMap map[string]edgeproto.InfraResource) map[string]edgeproto.InfraResource
	// Get Cloudlet Resource Properties
	GetCloudletResourceQuotaProps(ctx context.Context) (*edgeproto.CloudletResourceQuotaProps, error)
	// Get cluster additional resource metric
	GetClusterAdditionalResourceMetric(ctx context.Context, cloudlet *edgeproto.Cloudlet, resMetric *edgeproto.Metric, resources []edgeproto.VMResource) error
	// Get resources used by the cluster
	GetClusterInfraResources(ctx context.Context, clusterKey *edgeproto.ClusterInstKey) (*edgeproto.InfraResources, error)
	// Create an AppInst on a Cluster
	CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete an AppInst on a Cluster
	DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Update an AppInst
	UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error
	// Get AppInst runtime information
	GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error)
	// Get the client to manage the ClusterInst
	GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst, clientType string) (ssh.Client, error)
	// Get the client to manage the specified platform management node
	GetNodePlatformClient(ctx context.Context, node *edgeproto.CloudletMgmtNode, ops ...pc.SSHClientOp) (ssh.Client, error)
	// List the cloudlet management nodes used by this platform
	ListCloudletMgmtNodes(ctx context.Context, clusterInsts []edgeproto.ClusterInst, vmAppInsts []edgeproto.AppInst) ([]edgeproto.CloudletMgmtNode, error)
	// Get the command to pass to PlatformClient for the container command
	GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error)
	// Get the console URL of the VM appInst
	GetConsoleUrl(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst) (string, error)
	// Set power state of the AppInst
	SetPowerState(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Create Cloudlet returns cloudletResourcesCreated, error
	CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, caches *Caches, accessApi AccessApi, updateCallback edgeproto.CacheUpdateCallback) (bool, error)
	UpdateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete Cloudlet
	DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, caches *Caches, accessApi AccessApi, updateCallback edgeproto.CacheUpdateCallback) error
	// Save Cloudlet AccessVars
	SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete Cloudlet AccessVars
	DeleteCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error
	// Sync data with controller
	SyncControllerCache(ctx context.Context, caches *Caches, cloudletState dme.CloudletState) error
	// Get Cloudlet Manifest Config
	GetCloudletManifest(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, accessApi AccessApi, flavor *edgeproto.Flavor, caches *Caches) (*edgeproto.CloudletManifest, error)
	// Verify VM
	VerifyVMs(ctx context.Context, vms []edgeproto.VM) error
	// Get Cloudlet Properties
	GetCloudletProps(ctx context.Context) (*edgeproto.CloudletProps, error)
	// Platform-sepcific access data lookup (only called from Controller context)
	GetAccessData(ctx context.Context, cloudlet *edgeproto.Cloudlet, region string, vaultConfig *vault.Config, dataType string, arg []byte) (map[string]string, error)
	// Update the cloudlet's Trust Policy
	UpdateTrustPolicy(ctx context.Context, TrustPolicy *edgeproto.TrustPolicy) error
	//  Create and Update TrustPolicyException
	UpdateTrustPolicyException(ctx context.Context, TrustPolicyException *edgeproto.TrustPolicyException) error
	// Delete TrustPolicyException
	DeleteTrustPolicyException(ctx context.Context, TrustPolicyExceptionKey *edgeproto.TrustPolicyExceptionKey) error
	// Get restricted cloudlet create status
	GetRestrictedCloudletStatus(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, accessApi AccessApi, updateCallback edgeproto.CacheUpdateCallback) error
	// Get ssh clients of all root LBs
	GetRootLBClients(ctx context.Context) (map[string]ssh.Client, error)
	// Get RootLB Flavor
	GetRootLBFlavor(ctx context.Context) (*edgeproto.Flavor, error)
	// Called when the platform instance switches activity
	ActiveChanged(ctx context.Context, platformActive bool)
}

type ClusterSvc interface {
	// GetVersionProperties returns properties related to the platform version
	GetVersionProperties() map[string]string
	// Get AppInst Configs
	GetAppInstConfigs(ctx context.Context, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst,
		autoScalePolicy *edgeproto.AutoScalePolicy, settings *edgeproto.Settings,
		userAlerts []edgeproto.AlertPolicy) ([]*edgeproto.ConfigFile, error)
}

// AccessApi handles functions that require secrets access, but
// may be run from either the Controller or CRM context, so may either
// use Vault directly (Controller) or may go indirectly via Controller (CRM).
type AccessApi interface {
	cloudcommon.RegistryAuthApi
	cloudcommon.GetPublicCertApi
	GetCloudletAccessVars(ctx context.Context) (map[string]string, error)
	SignSSHKey(ctx context.Context, publicKey string) (string, error)
	GetSSHPublicKey(ctx context.Context) (string, error)
	GetOldSSHKey(ctx context.Context) (*vault.MEXKey, error)
	GetChefAuthKey(ctx context.Context) (*chefmgmt.ChefAuthKey, error)
	CreateOrUpdateDNSRecord(ctx context.Context, zone, name, rtype, content string, ttl int, proxy bool) error
	GetDNSRecords(ctx context.Context, zone, fqdn string) ([]cloudflare.DNSRecord, error)
	DeleteDNSRecord(ctx context.Context, zone, recordID string) error
	GetSessionTokens(ctx context.Context, arg []byte) (map[string]string, error)
	GetKafkaCreds(ctx context.Context) (*node.KafkaCreds, error)
	GetGCSCreds(ctx context.Context) ([]byte, error)
}

var pfMaps = map[string]string{
	"fakeinfra": "fake",
	"edgebox":   "dind",
	"kindinfra": "kind",
}

func GetType(pfType string) string {
	out := strings.TrimPrefix(pfType, "PLATFORM_TYPE_")
	out = strings.ToLower(out)
	out = strings.Replace(out, "_", "", -1)
	if mapOut, found := pfMaps[out]; found {
		return mapOut
	}
	return out
}
