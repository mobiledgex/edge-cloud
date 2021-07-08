package cloudcommon

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var (
	// Common platform resources
	ResourceRamMb       = "RAM"
	ResourceVcpus       = "vCPUs"
	ResourceGpus        = "GPUs"
	ResourceExternalIPs = "External IPs"

	// Platform specific resources
	ResourceInstances   = "Instances"
	ResourceFloatingIPs = "Floating IPs"

	// Resource units
	ResourceRamUnits = "MB"

	// Resource metrics
	ResourceMetricRamMB       = "ramUsed"
	ResourceMetricVcpus       = "vcpusUsed"
	ResourceMetricsGpus       = "gpusUsed"
	ResourceMetricInstances   = "instancesUsed"
	ResourceMetricExternalIPs = "externalIpsUsed"
	ResourceMetricFloatingIPs = "floatingIpsUsed"

	// Common cloudlet resources
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
			Name:        ResourceGpus,
			Description: "Limit on GPUs available",
		},
		edgeproto.InfraResource{
			Name:        ResourceExternalIPs,
			Description: "Limit on external IPs available",
		},
	}

	ResourceQuotaDesc = map[string]string{
		ResourceInstances:   "Limit on number of instances that can be provisioned",
		ResourceFloatingIPs: "Limit on number of floating IPs that can be created",
	}

	ResourceMetricsDesc = map[string]string{
		ResourceMetricRamMB:       "RAM Usage (MB)",
		ResourceMetricVcpus:       "vCPU Usage",
		ResourceMetricsGpus:       "GPU Usage",
		ResourceMetricInstances:   "VM Instance Usage",
		ResourceMetricExternalIPs: "External IP Usage",
		ResourceMetricFloatingIPs: "Floating IP Usage",
	}
)

// GetClusterInstVMRequirements uses the nodeFlavor and masterNodeFlavor if it cannot find a platform flavor
func GetClusterInstVMRequirements(ctx context.Context, clusterInst *edgeproto.ClusterInst, nodeFlavor, masterNodeFlavor, rootLBFlavor *edgeproto.FlavorInfo) ([]edgeproto.VMResource, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "GetClusterInstVMResources", "clusterinst key", clusterInst.Key, "nodeFlavor", nodeFlavor.Name, "masterNodeFlavor", masterNodeFlavor.Name, "root lb flavor", rootLBFlavor)
	vmResources := []edgeproto.VMResource{}

	if clusterInst.Deployment == DeploymentTypeDocker {
		vmResources = append(vmResources, edgeproto.VMResource{
			Key:      clusterInst.Key,
			VmFlavor: nodeFlavor,
			Type:     VMTypeClusterDockerNode,
		})
	} else {
		for ii := uint32(0); ii < clusterInst.NumMasters; ii++ {
			if clusterInst.MasterNodeFlavor == "" {
				vmResources = append(vmResources, edgeproto.VMResource{
					Key:      clusterInst.Key,
					VmFlavor: nodeFlavor,
					Type:     VMTypeClusterMaster,
				})
			} else {
				vmResources = append(vmResources, edgeproto.VMResource{
					Key:      clusterInst.Key,
					VmFlavor: masterNodeFlavor,
					Type:     VMTypeClusterMaster,
				})
			}
		}
		for ii := uint32(0); ii < clusterInst.NumNodes; ii++ {
			vmResources = append(vmResources, edgeproto.VMResource{
				Key:      clusterInst.Key,
				VmFlavor: nodeFlavor,
				Type:     VMTypeClusterK8sNode,
			})
		}
	}

	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		if rootLBFlavor == nil {
			return nil, fmt.Errorf("missing rootlb flavor")
		}
		vmResources = append(vmResources, edgeproto.VMResource{
			Key:      clusterInst.Key,
			VmFlavor: rootLBFlavor,
			Type:     VMTypeRootLB,
		})
	}
	return vmResources, nil
}

func GetVMAppRequirements(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, pfFlavorList []*edgeproto.FlavorInfo, rootLBFlavor *edgeproto.FlavorInfo) ([]edgeproto.VMResource, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "GetVMAppRequirements", "appinst key", appInst.Key, "platform flavors", pfFlavorList)
	vmResources := []edgeproto.VMResource{}
	vmFlavor := &edgeproto.FlavorInfo{}
	vmFlavorFound := false
	for _, flavor := range pfFlavorList {
		if flavor.Name == appInst.VmFlavor {
			vmFlavor = flavor
			vmFlavorFound = true
			break
		}
	}
	if !vmFlavorFound && len(pfFlavorList) > 0 {
		return nil, fmt.Errorf("VM flavor %s does not exist", appInst.VmFlavor)
	}
	if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
		vmResources = append(vmResources, edgeproto.VMResource{
			Key:      *appInst.ClusterInstKey(),
			VmFlavor: rootLBFlavor,
			Type:     VMTypeRootLB,
		})
	}
	vmResources = append(vmResources, edgeproto.VMResource{
		Key:           *appInst.ClusterInstKey(),
		VmFlavor:      vmFlavor,
		Type:          VMTypeAppVM,
		AppAccessType: app.AccessType,
	})
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

func ValidateCloudletResourceQuotas(ctx context.Context, curRes map[string]edgeproto.InfraResource, resourceQuotas []edgeproto.ResourceQuota) error {
	quotaNames := []string{}
	for name, _ := range curRes {
		quotaNames = append(quotaNames, name)
	}
	for _, resQuota := range resourceQuotas {
		infraRes, ok := curRes[resQuota.Name]
		if !ok {
			return fmt.Errorf("Invalid resource quota name: %s, valid names are %s", resQuota.Name, strings.Join(quotaNames, ","))
		}
		if infraRes.InfraMaxValue > 0 && resQuota.Value > infraRes.InfraMaxValue {
			return fmt.Errorf("Resource quota %s exceeded max supported value: %d", resQuota.Name, infraRes.InfraMaxValue)
		}
		if resQuota.Value > 0 && resQuota.Value < infraRes.Value {
			return fmt.Errorf("Resource quota value for %s is less than currently used value. Should be atleast %d", resQuota.Name, infraRes.Value)
		}
	}
	return nil
}

var GPUResourceLimitName = "nvidia.com/gpu"

func IsGPUFlavor(flavor *edgeproto.Flavor) (bool, int) {
	if flavor == nil {
		return false, 0
	}
	resStr, ok := flavor.OptResMap["gpu"]
	if !ok {
		return ok, 0
	}
	count, err := ParseGPUResourceCount(resStr)
	if err != nil {
		return false, 0
	}
	return true, count
}

func ParseGPUResourceCount(resStr string) (int, error) {
	count := 0
	values := strings.Split(resStr, ":")
	if len(values) == 1 {
		return count, fmt.Errorf("Missing manditory resource count, ex: optresmap=gpu=gpu:1")
	}
	var countStr string
	var err error
	if len(values) == 2 {
		countStr = values[1]
	} else if len(values) == 3 {
		countStr = values[2]
	} else {
		return count, fmt.Errorf("Invalid optresmap syntax encountered: ex: optresmap=gpu=gpu:1")
	}
	if count, err = strconv.Atoi(countStr); err != nil {
		return count, fmt.Errorf("Non-numeric resource count encountered, found %s", values[1])
	}
	return count, nil
}
