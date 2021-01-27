package cloudcommon

import (
	"context"
	"fmt"
	"strings"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var (
	// Common platform resources
	ResourceRamMb  = "RAM"
	ResourceVcpus  = "vCPUs"
	ResourceDiskGb = "disk"
	ResourceGpus   = "GPUs"

	ResourceRamUnits  = "MB"
	ResourceDiskUnits = "GB"

	CloudletResources = []edgeproto.InfraResource{
		edgeproto.InfraResource{
			Name:        ResourceRamMb,
			Description: "Limit on RAM available (MB)",
		},
		edgeproto.InfraResource{
			Name:        ResourceVcpus,
			Description: "Limit on vCPUs available",
		},
		edgeproto.InfraResource{
			Name:        ResourceDiskGb,
			Description: "Limit on disk available (GB)",
		},
		edgeproto.InfraResource{
			Name:        ResourceGpus,
			Description: "Limit on GPUs available",
		},
	}
)

func GetClusterInstVMRequirements(ctx context.Context, clusterInst *edgeproto.ClusterInst, pfFlavorList []*edgeproto.FlavorInfo, rootLBFlavor *edgeproto.FlavorInfo, provState edgeproto.VMProvState) ([]edgeproto.VMResource, error) {
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
	vmResState := edgeproto.VMProvState_PROV_STATE_ADD
	if provState == edgeproto.VMProvState_PROV_STATE_REMOVE {
		vmResState = provState
	}
	if clusterInst.Deployment == DeploymentTypeDocker {
		vmResources = append(vmResources, edgeproto.VMResource{
			Key:       clusterInst.Key,
			VmFlavor:  nodeFlavor,
			ProvState: vmResState,
		})
	} else {
		for ii := uint32(0); ii < clusterInst.NumMasters; ii++ {
			if clusterInst.MasterNodeFlavor == "" {
				vmResources = append(vmResources, edgeproto.VMResource{
					Key:       clusterInst.Key,
					VmFlavor:  nodeFlavor,
					ProvState: vmResState,
					Type:      VMTypeClusterMaster,
				})
			} else {
				vmResources = append(vmResources, edgeproto.VMResource{
					Key:       clusterInst.Key,
					VmFlavor:  masterNodeFlavor,
					ProvState: vmResState,
					Type:      VMTypeClusterMaster,
				})
			}
		}
		for ii := uint32(0); ii < clusterInst.NumNodes; ii++ {
			vmResources = append(vmResources, edgeproto.VMResource{
				Key:       clusterInst.Key,
				VmFlavor:  nodeFlavor,
				ProvState: vmResState,
				Type:      VMTypeClusterNode,
			})
		}
	}

	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		if rootLBFlavor == nil {
			return nil, fmt.Errorf("missing rootlb flavor")
		}
		vmResources = append(vmResources, edgeproto.VMResource{
			Key:       clusterInst.Key,
			VmFlavor:  rootLBFlavor,
			ProvState: vmResState,
			Type:      VMTypeRootLB,
		})
	}
	return vmResources, nil
}

func usageAlertWarningLabels(ctx context.Context, key *edgeproto.CloudletKey, alertname, warning string) map[string]string {
	labels := make(map[string]string)
	labels["alertname"] = alertname
	labels[AlertScopeTypeTag] = AlertScopeCloudlet
	labels[edgeproto.CloudletKeyTagName] = key.Name
	labels[edgeproto.CloudletKeyTagOrganization] = key.Organization
	labels["warning"] = warning
	return labels
}

// Raise the alarm when there are cloudlet resource usage warnings
func CloudletResourceUsageAlerts(ctx context.Context, key *edgeproto.CloudletKey, warnings []string) []edgeproto.Alert {
	alerts := []edgeproto.Alert{}
	for _, warning := range warnings {
		alert := edgeproto.Alert{}
		alert.State = "firing"
		alert.ActiveAt = dme.Timestamp{}
		ts := time.Now()
		alert.ActiveAt.Seconds = ts.Unix()
		alert.ActiveAt.Nanos = int32(ts.Nanosecond())
		alert.Labels = usageAlertWarningLabels(ctx, key, AlertCloudletResourceUsage, warning)
		alert.Annotations = make(map[string]string)
		alert.Annotations[AlertAnnotationTitle] = AlertCloudletResourceUsage
		alert.Annotations[AlertAnnotationDescription] = warning
		alerts = append(alerts, alert)
	}
	return alerts
}

func ValidateCloudletResourceQuotas(ctx context.Context, infraResources *edgeproto.InfraResources, resourceQuotas []edgeproto.ResourceQuota) error {
	if infraResources == nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Failed to validate cloudlet resource quotas, missing infra resources info")
		return nil
	}
	validQuotas := make(map[string]uint64)
	for _, info := range infraResources.Info {
		validQuotas[info.Name] = info.MaxValue
	}
	for _, commonRes := range CloudletResources {
		validQuotas[commonRes.Name] = 0
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
		if qMaxVal > 0 {
			if resQuota.Value > qMaxVal {
				return fmt.Errorf("Resource quota %s exceeded max supported value: %d", resQuota.Name, qMaxVal)
			}
		}
	}
	return nil
}
