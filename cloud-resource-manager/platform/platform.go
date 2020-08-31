package platform

import (
	"context"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	ssh "github.com/mobiledgex/golang-ssh"
)

type PlatformConfig struct {
	CloudletKey         *edgeproto.CloudletKey
	PhysicalName        string
	VaultAddr           string
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
}

type Caches struct {
	SettingsCache      *edgeproto.SettingsCache
	FlavorCache        *edgeproto.FlavorCache
	PrivacyPolicyCache *edgeproto.PrivacyPolicyCache
	ClusterInstCache   *edgeproto.ClusterInstCache
	AppInstCache       *edgeproto.AppInstCache
	AppCache           *edgeproto.AppCache
	ResTagTableCache   *edgeproto.ResTagTableCache
	CloudletCache      *edgeproto.CloudletCache
	VMPoolCache        *edgeproto.VMPoolCache
	VMPoolInfoCache    *edgeproto.VMPoolInfoCache

	// VMPool object managed by CRM
	VMPool    *edgeproto.VMPool
	VMPoolMux *sync.Mutex
}

// Platform abstracts the underlying cloudlet platform.
type Platform interface {
	// GetType returns the cloudlet's stack type, i.e. Openstack, Azure, etc.
	GetType() string
	// Init is called once during CRM startup.
	Init(ctx context.Context, platformConfig *PlatformConfig, caches *Caches, updateCallback edgeproto.CacheUpdateCallback) error
	// Gather information about the cloudlet platform.
	// This includes available resources, flavors, etc.
	// Returns true if sync with controller is required
	GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error
	// Create a Kubernetes Cluster on the cloudlet.
	CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error
	// Delete a Kuberentes Cluster on the cloudlet.
	DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst) error
	// Update the cluster
	UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error
	// Create an AppInst on a Cluster
	CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete an AppInst on a Cluster
	DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error
	// Update an AppInst
	UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Get AppInst runtime information
	GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error)
	// Get the client to manage the ClusterInst
	GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst, clientType string) (ssh.Client, error)
	// Get the client to manage the specified platform management node
	GetNodePlatformClient(ctx context.Context, node *edgeproto.CloudletMgmtNode) (ssh.Client, error)
	// List the cloudlet management nodes used by this platform
	ListCloudletMgmtNodes(ctx context.Context, clusterInsts []edgeproto.ClusterInst) ([]edgeproto.CloudletMgmtNode, error)
	// Get the command to pass to PlatformClient for the container command
	GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error)
	// Get the console URL of the VM app
	GetConsoleUrl(ctx context.Context, app *edgeproto.App) (string, error)
	// Set power state of the AppInst
	SetPowerState(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Create Cloudlet
	CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, caches *Caches, updateCallback edgeproto.CacheUpdateCallback) error
	UpdateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete Cloudlet
	DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, caches *Caches, updateCallback edgeproto.CacheUpdateCallback) error
	// Save Cloudlet AccessVars
	SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete Cloudlet AccessVars
	DeleteCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error
	// Sync data with controller
	SyncControllerCache(ctx context.Context, caches *Caches, cloudletState edgeproto.CloudletState) error
	// Get Cloudlet Manifest Config
	GetCloudletManifest(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, caches *Caches) (*edgeproto.CloudletManifest, error)
	// Verify VM
	VerifyVMs(ctx context.Context, vms []edgeproto.VM) error
	// Get Cloudlet Properties
	GetCloudletProps(ctx context.Context) (*edgeproto.CloudletProps, error)
}

type ClusterSvc interface {
	GetAppInstConfigs(ctx context.Context, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst, autoScalePolicy *edgeproto.AutoScalePolicy) ([]*edgeproto.ConfigFile, error)
}
