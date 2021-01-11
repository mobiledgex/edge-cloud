package cloudcommon

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func GetClusterInstVMRequirements(ctx context.Context, clusterInst *edgeproto.ClusterInst, pfFlavorList []*edgeproto.FlavorInfo, rootLBFlavor *edgeproto.FlavorInfo) ([]edgeproto.VMResource, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetClusterInstVMResources", "clusterinst key", clusterInst.Key, "platform flavors", pfFlavorList, "root lb flavor", rootLBFlavor)
	vmResources := []edgeproto.VMResource{}
	nodeFlavor := &edgeproto.FlavorInfo{}
	masterNodeFlavor := &edgeproto.FlavorInfo{}
	nodeFlavorFound := false
	masterNodeFlavorFound := false
	for _, flavor := range pfFlavorList {
		if flavor.Name == clusterInst.NodeFlavor {
			nodeFlavor = flavor
			nodeFlavorFound = true
		}
		if flavor.Name == clusterInst.MasterNodeFlavor {
			masterNodeFlavor = flavor
			masterNodeFlavorFound = true
		}
	}

	if !nodeFlavorFound {
		return nil, fmt.Errorf("Node flavor %s does not exist", clusterInst.NodeFlavor)
	}
	if clusterInst.MasterNodeFlavor != "" && !masterNodeFlavorFound {
		return nil, fmt.Errorf("Master node flavor %s does not exist", clusterInst.MasterNodeFlavor)
	}
	if clusterInst.Deployment == DeploymentTypeDocker {
		vmResources = append(vmResources, edgeproto.VMResource{
			VmFlavor: nodeFlavor,
		})
	} else {
		for ii := uint32(0); ii < clusterInst.NumMasters; ii++ {
			if clusterInst.MasterNodeFlavor == "" {
				vmResources = append(vmResources, edgeproto.VMResource{
					VmFlavor: nodeFlavor,
				})
			} else {
				vmResources = append(vmResources, edgeproto.VMResource{
					VmFlavor: masterNodeFlavor,
				})
			}
		}
		for ii := uint32(0); ii < clusterInst.NumNodes; ii++ {
			vmResources = append(vmResources, edgeproto.VMResource{
				VmFlavor: nodeFlavor,
			})
		}
	}

	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		if rootLBFlavor == nil {
			return nil, fmt.Errorf("missing rootlb flavor")
		}
		vmResources = append(vmResources, edgeproto.VMResource{
			VmFlavor: rootLBFlavor,
		})
	}
	return vmResources, nil
}
