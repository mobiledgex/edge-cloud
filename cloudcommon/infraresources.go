package cloudcommon

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	v1 "k8s.io/api/core/v1"
)

var (
	// Common platform resources
	ResourceRamMb       = "RAM"
	ResourceVcpus       = "vCPUs"
	ResourceDiskGb      = "Disk"
	ResourceGpus        = "GPUs"
	ResourceExternalIPs = "External IPs"

	// Platform specific resources
	ResourceInstances             = "Instances"
	ResourceFloatingIPs           = "Floating IPs"
	ResourceK8sClusters           = "K8s Clusters"
	ResourceMaxK8sNodesPerCluster = "Maximum K8s Nodes Per Cluster"
	ResourceTotalK8sNodes         = "Total Number Of K8s Nodes"
	ResourceNetworkLBs            = "Network Load Balancers"

	// Resource units
	ResourceRamUnits  = "MB"
	ResourceDiskUnits = "GB"

	// Resource metrics
	ResourceMetricRamMB                 = "ramUsed"
	ResourceMetricVcpus                 = "vcpusUsed"
	ResourceMetricDisk                  = "diskUsed"
	ResourceMetricGpus                  = "gpusUsed"
	ResourceMetricInstances             = "instancesUsed"
	ResourceMetricExternalIPs           = "externalIpsUsed"
	ResourceMetricFloatingIPs           = "floatingIpsUsed"
	ResourceMetricK8sClusters           = "k8sClustersUsed"
	ResourceMetricMaxK8sNodesPerCluster = "maxK8sNodesPerClusterUsed"
	ResourceMetricTotalK8sNodes         = "totalK8sNodesUsed"
	ResourceMetricNetworkLBs            = "networkLBsUsed"

	// Common cloudlet resources
	CommonCloudletResources = map[string]string{
		ResourceRamMb:       ResourceRamUnits,
		ResourceVcpus:       "",
		ResourceDiskGb:      "",
		ResourceGpus:        "",
		ResourceExternalIPs: "",
	}

	ResourceQuotaDesc = map[string]string{
		ResourceRamMb:                 "Limit on RAM available (MB)",
		ResourceVcpus:                 "Limit on vCPUs available",
		ResourceDiskGb:                "Limit on disk available (GB)",
		ResourceGpus:                  "Limit on GPUs available",
		ResourceExternalIPs:           "Limit on external IPs available",
		ResourceInstances:             "Limit on number of instances that can be provisioned",
		ResourceFloatingIPs:           "Limit on number of floating IPs that can be created",
		ResourceK8sClusters:           "Limit on number of k8s clusters than can be created",
		ResourceMaxK8sNodesPerCluster: "Limit on maximum number of k8s nodes that can be created as part of k8s cluster",
		ResourceTotalK8sNodes:         "Limit on total number of k8s nodes that can be created altogether",
		ResourceNetworkLBs:            "Limit on maximum number of network load balancers that can be created in a region",
	}

	ResourceMetricsDesc = map[string]string{
		ResourceMetricRamMB:                 "RAM Usage (MB)",
		ResourceMetricVcpus:                 "vCPU Usage",
		ResourceMetricDisk:                  "Disk Usage (GB)",
		ResourceMetricGpus:                  "GPU Usage",
		ResourceMetricExternalIPs:           "External IP Usage",
		ResourceMetricInstances:             "VM Instance Usage",
		ResourceMetricFloatingIPs:           "Floating IP Usage",
		ResourceMetricK8sClusters:           "K8s Cluster Usage",
		ResourceMetricMaxK8sNodesPerCluster: "Maximum K8s Nodes Per Cluster Usage",
		ResourceMetricTotalK8sNodes:         "Total K8s Nodes Usage",
		ResourceMetricNetworkLBs:            "Network Load Balancer Usage",
	}
)

// GetClusterInstVMRequirements uses the nodeFlavor and masterNodeFlavor if it cannot find a platform flavor
func GetClusterInstVMRequirements(ctx context.Context, clusterInst *edgeproto.ClusterInst, nodeFlavor, masterNodeFlavor, rootLBFlavor *edgeproto.FlavorInfo, isManagedK8s bool) ([]edgeproto.VMResource, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "GetClusterInstVMResources", "clusterinst key", clusterInst.Key, "nodeFlavor", nodeFlavor.Name, "masterNodeFlavor", masterNodeFlavor.Name, "root lb flavor", rootLBFlavor, "managed k8s", isManagedK8s)
	vmResources := []edgeproto.VMResource{}

	if clusterInst.Deployment == DeploymentTypeDocker {
		vmResources = append(vmResources, edgeproto.VMResource{
			Key:      clusterInst.Key,
			VmFlavor: nodeFlavor,
			Type:     NodeTypeDockerClusterNode.String(),
		})
	} else {
		// For managed-k8s platforms, ignore master node for resource calculation
		if !isManagedK8s {
			for ii := uint32(0); ii < clusterInst.NumMasters; ii++ {
				if clusterInst.MasterNodeFlavor == "" {
					vmResources = append(vmResources, edgeproto.VMResource{
						Key:      clusterInst.Key,
						VmFlavor: nodeFlavor,
						Type:     NodeTypeK8sClusterMaster.String(),
					})
				} else {
					vmResources = append(vmResources, edgeproto.VMResource{
						Key:      clusterInst.Key,
						VmFlavor: masterNodeFlavor,
						Type:     NodeTypeK8sClusterMaster.String(),
					})
				}
			}
		}
		for ii := uint32(0); ii < clusterInst.NumNodes; ii++ {
			vmResources = append(vmResources, edgeproto.VMResource{
				Key:      clusterInst.Key,
				VmFlavor: nodeFlavor,
				Type:     NodeTypeK8sClusterNode.String(),
			})
		}
	}

	// For managed-k8s platforms, ignore rootLB for resource calculation
	if !isManagedK8s {
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			if rootLBFlavor == nil {
				return nil, fmt.Errorf("missing rootlb flavor")
			}
			vmResources = append(vmResources, edgeproto.VMResource{
				Key:      clusterInst.Key,
				VmFlavor: rootLBFlavor,
				Type:     NodeTypeDedicatedRootLB.String(),
			})
		}
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
			Type:     NodeTypeDedicatedRootLB.String(),
		})
	}
	vmResources = append(vmResources, edgeproto.VMResource{
		Key:           *appInst.ClusterInstKey(),
		VmFlavor:      vmFlavor,
		Type:          NodeTypeAppVM.String(),
		AppAccessType: app.AccessType,
	})
	return vmResources, nil
}

func GetK8sAppRequirements(ctx context.Context, app *edgeproto.App) ([]edgeproto.VMResource, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "GetK8sAppRequirements", "app", app)
	resources := []edgeproto.VMResource{}
	if app.Deployment != DeploymentTypeKubernetes && app.Deployment != DeploymentTypeHelm {
		return resources, nil
	}
	objs, _, err := DecodeK8SYaml(app.DeploymentManifest)
	if err != nil {
		return nil, fmt.Errorf("parse kubernetes deployment yaml failed for app %v, %v", app.Key, err)
	}
	for _, obj := range objs {
		ksvc, ok := obj.(*v1.Service)
		if !ok {
			continue
		}
		if ksvc.Spec.Type != v1.ServiceTypeLoadBalancer {
			continue
		}
		// add LB svc as a resource for public-cloud based platform validations
		resources = append(resources, edgeproto.VMResource{
			Type: ResourceTypeK8sLBSvc,
		})
	}
	return resources, nil
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

func ValidateCloudletResourceQuotas(ctx context.Context, clSpecificProps *edgeproto.CloudletResourceQuotaProps, curRes map[string]edgeproto.InfraResource, resourceQuotas []edgeproto.ResourceQuota) error {
	resPropsMap := make(map[string]struct{})
	resPropsNames := []string{}
	for _, prop := range clSpecificProps.Properties {
		resPropsMap[prop.Name] = struct{}{}
		resPropsNames = append(resPropsNames, prop.Name)
	}
	for resName, _ := range CommonCloudletResources {
		resPropsMap[resName] = struct{}{}
		resPropsNames = append(resPropsNames, resName)
	}
	sort.Strings(resPropsNames)
	for _, resQuota := range resourceQuotas {
		if _, ok := resPropsMap[resQuota.Name]; !ok {
			return fmt.Errorf("Invalid quota name: %s, valid names are %s", resQuota.Name, strings.Join(resPropsNames, ", "))
		}
		if curRes == nil {
			continue
		}
		infraRes, ok := curRes[resQuota.Name]
		if !ok {
			continue
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
