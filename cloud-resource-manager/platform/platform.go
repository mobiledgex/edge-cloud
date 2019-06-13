package platform

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type PlatformConfig struct {
	CloudletKey  *edgeproto.CloudletKey
	PhysicalName string
	VaultAddr    string
}

// Platform abstracts the underlying cloudlet platform.
type Platform interface {
	// GetType Returns the Cloudlet's stack type, i.e. Openstack, Azure, etc.
	GetType() string
	// Init is called once during CRM startup.
	Init(platformConfig *PlatformConfig) error
	// Gather information about the cloudlet platform.
	// This includes available resources, flavors, etc.
	GatherCloudletInfo(info *edgeproto.CloudletInfo) error
	// Create a Kubernetes Cluster on the cloudlet.
	CreateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete a Kuberentes Cluster on the cloudlet.
	DeleteClusterInst(clusterInst *edgeproto.ClusterInst) error
	// Update the cluster
	UpdateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error
	// Create an AppInst on a Cluster
	CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete an AppInst on a Cluster
	DeleteAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error
	// Get AppInst runtime information
	GetAppInstRuntime(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error)
	// Get the Platform Client to run commands against
	GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error)
	// Get the command to pass to PlatformClient for the container command
	GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error)
	// Create Cloudlet
	CreateCloudlet(cloudlet *edgeproto.Cloudlet, updateCallback edgeproto.CacheUpdateCallback) error
	// Delete Cloudlet
	DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error
}
