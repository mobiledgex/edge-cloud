package platform

import (
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
	CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor) error
	// Delete an AppInst on a Cluster
	DeleteAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error
}
