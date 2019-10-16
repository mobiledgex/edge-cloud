package platform

import (
	"context"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type PlatformConfig struct {
	CloudletKey  *edgeproto.CloudletKey
	PhysicalName string
	VaultAddr    string
	EnvVars      map[string]string
	TestMode     bool
}

// Platform abstracts the underlying cloudlet platform.
type Platform interface {
	// GetType Returns the Cloudlet's stack type, i.e. Openstack, Azure, etc.
	GetType() string
	// Init is called once during CRM startup.
	Init(ctx context.Context, platformConfig *PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error
	// Gather information about the cloudlet platform.
	// This includes available resources, flavors, etc.
	GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error
	// Create a Kubernetes Cluster on the cloudlet.
	CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error
	// Delete a Kuberentes Cluster on the cloudlet.
	DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst) error
	// Update the cluster
	UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Create an AppInst on a Cluster
	CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete an AppInst on a Cluster
	DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error
	// Update an AppInst
	UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Get AppInst runtime information
	GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error)
	// Get the Platform Client to run commands against
	GetPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error)
	// Get the command to pass to PlatformClient for the container command
	GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error)
	// Get the console URL of the VM app
	GetConsoleUrl(ctx context.Context, app *edgeproto.App) (string, error)
	// Create Cloudlet
	CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete Cloudlet
	DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error
}
