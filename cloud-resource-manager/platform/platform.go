package platform

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Platform abstracts the underlying cloudlet platform.
type Platform interface {
	// GetType Returns the Cloudlet's stack type, i.e. Openstack, Azure, etc.
	GetType() string
	// Init is called once during CRM startup.
	Init(key *edgeproto.CloudletKey) error
	// Gather information about the cloudlet platform.
	// This includes available resources, flavors, etc.
	GatherCloudletInfo(info *edgeproto.CloudletInfo) error
	// Create a Kubernetes Cluster on the cloudlet.
	CreateCluster(clusterInst *edgeproto.ClusterInst, flavor *edgeproto.ClusterFlavor) error
	// Delete a Kuberentes Cluster on the cloudlet.
	DeleteCluster(clusterInst *edgeproto.ClusterInst) error
	// Create an AppInst on a Cluster
	CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error
	// Delete an AppInst on a Cluster
	DeleteAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error
	// Get AppInst runtime information
	GetAppInstRuntime(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error)
	// Get the Platform Client to run commands against
	GetPlatformClient() (pc.PlatformClient, error)
	// Get the command to pass to PlatformClient for the container command
	GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error)
}
