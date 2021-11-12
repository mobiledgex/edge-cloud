package k8smgmt

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestGetNodeInfos(t *testing.T) {
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	client := pc.DummyClient{}
	client.Out = getNodesSampleOutput

	nodeInfos, err := GetNodeInfos(ctx, &client, "")
	require.Nil(t, err)

	expNodes := []*edgeproto.NodeInfo{{
		Name: "aks-agentpool-30520393-vmss000000",
		Allocatable: map[string]*edgeproto.Udec64{
			cloudcommon.ResourceVcpus: edgeproto.NewUdec64(1, 900*edgeproto.DecMillis),
			cloudcommon.ResourceRamMb: edgeproto.NewUdec64(5368, 0),
			cloudcommon.ResourceDisk:  edgeproto.NewUdec64(111, 0),
		},
		Capacity: map[string]*edgeproto.Udec64{
			cloudcommon.ResourceVcpus: edgeproto.NewUdec64(2, 0),
			cloudcommon.ResourceRamMb: edgeproto.NewUdec64(7961, 0),
			cloudcommon.ResourceDisk:  edgeproto.NewUdec64(123, 0),
		},
	}, {
		Name: "aks-agentpool-30520393-vmss000001",
		Allocatable: map[string]*edgeproto.Udec64{
			cloudcommon.ResourceVcpus: edgeproto.NewUdec64(1, 900*edgeproto.DecMillis),
			cloudcommon.ResourceRamMb: edgeproto.NewUdec64(5368, 0),
			cloudcommon.ResourceDisk:  edgeproto.NewUdec64(111, 0),
		},
		Capacity: map[string]*edgeproto.Udec64{
			cloudcommon.ResourceVcpus: edgeproto.NewUdec64(2, 0),
			cloudcommon.ResourceRamMb: edgeproto.NewUdec64(7961, 0),
			cloudcommon.ResourceDisk:  edgeproto.NewUdec64(123, 0),
		},
	}, {
		Name: "aks-agentpool-30520393-vmss000002",
		Allocatable: map[string]*edgeproto.Udec64{
			cloudcommon.ResourceVcpus: edgeproto.NewUdec64(1, 900*edgeproto.DecMillis),
			cloudcommon.ResourceRamMb: edgeproto.NewUdec64(5368, 0),
			cloudcommon.ResourceDisk:  edgeproto.NewUdec64(111, 0),
		},
		Capacity: map[string]*edgeproto.Udec64{
			cloudcommon.ResourceVcpus: edgeproto.NewUdec64(2, 0),
			cloudcommon.ResourceRamMb: edgeproto.NewUdec64(7961, 0),
			cloudcommon.ResourceDisk:  edgeproto.NewUdec64(123, 0),
		},
	}}
	require.Equal(t, expNodes, nodeInfos)
}

// Output of "kubectl get nodes --output=json"
var getNodesSampleOutput = `
{
    "apiVersion": "v1",
    "items": [
        {
            "apiVersion": "v1",
            "kind": "Node",
            "metadata": {
                "annotations": {
                    "node.alpha.kubernetes.io/ttl": "0",
                    "volumes.kubernetes.io/controller-managed-attach-detach": "true"
                },
                "creationTimestamp": "2021-07-18T09:42:46Z",
                "labels": {
                    "agentpool": "agentpool",
                    "beta.kubernetes.io/arch": "amd64",
                    "beta.kubernetes.io/instance-type": "Standard_D2s_v3",
                    "beta.kubernetes.io/os": "linux",
                    "failure-domain.beta.kubernetes.io/region": "southcentralus",
                    "failure-domain.beta.kubernetes.io/zone": "0",
                    "kubernetes.azure.com/node-image-version": "AKSUbuntu-1804gen2-2021.06.12",
                    "kubernetes.azure.com/os-sku": "Ubuntu",
                    "kubernetes.azure.com/role": "agent",
                    "kubernetes.io/arch": "amd64",
                    "kubernetes.io/hostname": "aks-agentpool-30520393-vmss000000",
                    "kubernetes.io/os": "linux",
                    "kubernetes.io/role": "agent",
                    "node-role.kubernetes.io/agent": "",
                    "node.kubernetes.io/instance-type": "Standard_D2s_v3",
                    "storageprofile": "managed",
                    "storagetier": "Premium_LRS",
                    "topology.kubernetes.io/region": "southcentralus",
                    "topology.kubernetes.io/zone": "0"
                },
                "managedFields": [
                    {
                        "apiVersion": "v1",
                        "fieldsType": "FieldsV1",
                        "fieldsV1": {
                            "f:metadata": {
                                "f:labels": {
                                    "f:kubernetes.io/role": {},
                                    "f:node-role.kubernetes.io/agent": {}
                                }
                            }
                        },
                        "manager": "kubectl",
                        "operation": "Update",
                        "time": "2021-07-18T09:43:02Z"
                    },
                    {
                        "apiVersion": "v1",
                        "fieldsType": "FieldsV1",
                        "fieldsV1": {
                            "f:metadata": {
                                "f:annotations": {
                                    ".": {},
                                    "f:volumes.kubernetes.io/controller-managed-attach-detach": {}
                                },
                                "f:labels": {
                                    ".": {},
                                    "f:agentpool": {},
                                    "f:beta.kubernetes.io/arch": {},
                                    "f:beta.kubernetes.io/instance-type": {},
                                    "f:beta.kubernetes.io/os": {},
                                    "f:failure-domain.beta.kubernetes.io/region": {},
                                    "f:failure-domain.beta.kubernetes.io/zone": {},
                                    "f:kubernetes.azure.com/cluster": {},
                                    "f:kubernetes.azure.com/node-image-version": {},
                                    "f:kubernetes.azure.com/os-sku": {},
                                    "f:kubernetes.azure.com/role": {},
                                    "f:kubernetes.io/arch": {},
                                    "f:kubernetes.io/hostname": {},
                                    "f:kubernetes.io/os": {},
                                    "f:node.kubernetes.io/instance-type": {},
                                    "f:storageprofile": {},
                                    "f:storagetier": {},
                                    "f:topology.kubernetes.io/region": {},
                                    "f:topology.kubernetes.io/zone": {}
                                }
                            },
                            "f:spec": {
                                "f:providerID": {}
                            },
                            "f:status": {
                                "f:addresses": {
                                    ".": {},
                                    "k:{\"type\":\"Hostname\"}": {
                                        ".": {},
                                        "f:address": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"InternalIP\"}": {
                                        ".": {},
                                        "f:address": {},
                                        "f:type": {}
                                    }
                                },
                                "f:allocatable": {
                                    ".": {},
                                    "f:attachable-volumes-azure-disk": {},
                                    "f:cpu": {},
                                    "f:ephemeral-storage": {},
                                    "f:hugepages-1Gi": {},
                                    "f:hugepages-2Mi": {},
                                    "f:memory": {},
                                    "f:pods": {}
                                },
                                "f:capacity": {
                                    ".": {},
                                    "f:attachable-volumes-azure-disk": {},
                                    "f:cpu": {},
                                    "f:ephemeral-storage": {},
                                    "f:hugepages-1Gi": {},
                                    "f:hugepages-2Mi": {},
                                    "f:memory": {},
                                    "f:pods": {}
                                },
                                "f:conditions": {
                                    ".": {},
                                    "k:{\"type\":\"DiskPressure\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"MemoryPressure\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"PIDPressure\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"Ready\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    }
                                },
                                "f:config": {},
                                "f:daemonEndpoints": {
                                    "f:kubeletEndpoint": {
                                        "f:Port": {}
                                    }
                                },
                                "f:images": {},
                                "f:nodeInfo": {
                                    "f:architecture": {},
                                    "f:bootID": {},
                                    "f:containerRuntimeVersion": {},
                                    "f:kernelVersion": {},
                                    "f:kubeProxyVersion": {},
                                    "f:kubeletVersion": {},
                                    "f:machineID": {},
                                    "f:operatingSystem": {},
                                    "f:osImage": {},
                                    "f:systemUUID": {}
                                },
                                "f:volumesInUse": {}
                            }
                        },
                        "manager": "kubelet",
                        "operation": "Update",
                        "time": "2021-07-18T12:35:04Z"
                    },
                    {
                        "apiVersion": "v1",
                        "fieldsType": "FieldsV1",
                        "fieldsV1": {
                            "f:metadata": {
                                "f:annotations": {
                                    "f:node.alpha.kubernetes.io/ttl": {}
                                }
                            },
                            "f:spec": {
                                "f:podCIDR": {},
                                "f:podCIDRs": {
                                    ".": {},
                                    "v:\"10.244.1.0/24\"": {}
                                }
                            },
                            "f:status": {
                                "f:conditions": {
                                    "k:{\"type\":\"NetworkUnavailable\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    }
                                },
                                "f:volumesAttached": {}
                            }
                        },
                        "manager": "kube-controller-manager",
                        "operation": "Update",
                        "time": "2021-07-18T12:35:36Z"
                    }
                ],
                "name": "aks-agentpool-30520393-vmss000000",
                "resourceVersion": "25296131",
                "selfLink": "/api/v1/nodes/aks-agentpool-30520393-vmss000000",
                "uid": "31af118e-5959-440a-8417-f3abbf6b9ed9"
            },
            "spec": {
                "podCIDR": "10.244.1.0/24",
                "podCIDRs": [
                    "10.244.1.0/24"
                ]
            },
            "status": {
                "addresses": [
                    {
                        "address": "aks-agentpool-30520393-vmss000000",
                        "type": "Hostname"
                    },
                    {
                        "address": "10.240.0.4",
                        "type": "InternalIP"
                    }
                ],
                "allocatable": {
                    "attachable-volumes-azure-disk": "4",
                    "cpu": "1900m",
                    "ephemeral-storage": "119716326407",
                    "hugepages-1Gi": "0",
                    "hugepages-2Mi": "0",
                    "memory": "5497568Ki",
                    "pods": "110"
                },
                "capacity": {
                    "attachable-volumes-azure-disk": "4",
                    "cpu": "2",
                    "ephemeral-storage": "129900528Ki",
                    "hugepages-1Gi": "0",
                    "hugepages-2Mi": "0",
                    "memory": "8152800Ki",
                    "pods": "110"
                },
                "conditions": [
                    {
                        "lastHeartbeatTime": "2021-07-18T09:43:09Z",
                        "lastTransitionTime": "2021-07-18T09:43:09Z",
                        "message": "RouteController created a route",
                        "reason": "RouteCreated",
                        "status": "False",
                        "type": "NetworkUnavailable"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:53:39Z",
                        "lastTransitionTime": "2021-07-18T09:42:46Z",
                        "message": "kubelet has sufficient memory available",
                        "reason": "KubeletHasSufficientMemory",
                        "status": "False",
                        "type": "MemoryPressure"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:53:39Z",
                        "lastTransitionTime": "2021-07-18T09:42:46Z",
                        "message": "kubelet has no disk pressure",
                        "reason": "KubeletHasNoDiskPressure",
                        "status": "False",
                        "type": "DiskPressure"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:53:39Z",
                        "lastTransitionTime": "2021-07-18T09:42:46Z",
                        "message": "kubelet has sufficient PID available",
                        "reason": "KubeletHasSufficientPID",
                        "status": "False",
                        "type": "PIDPressure"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:53:39Z",
                        "lastTransitionTime": "2021-07-18T09:42:47Z",
                        "message": "kubelet is posting ready status. AppArmor enabled",
                        "reason": "KubeletReady",
                        "status": "True",
                        "type": "Ready"
                    }
                ],
                "config": {},
                "daemonEndpoints": {
                    "kubeletEndpoint": {
                        "Port": 10211
                    }
                },
                "nodeInfo": {
                    "architecture": "amd64",
                    "bootID": "0615c9d0-7dc4-4000-b898-8d434ea2bb14",
                    "containerRuntimeVersion": "docker://19.3.14",
                    "kernelVersion": "5.4.0-1049-azure",
                    "kubeProxyVersion": "v1.18.19",
                    "kubeletVersion": "v1.18.19",
                    "machineID": "aae645a2ac5644fa8edfd13dcc420af4",
                    "operatingSystem": "linux",
                    "osImage": "Ubuntu 18.04.5 LTS",
                    "systemUUID": "5baa476c-6d93-4888-8afc-b5970d23e731"
                }
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Node",
            "metadata": {
                "annotations": {
                    "node.alpha.kubernetes.io/ttl": "0",
                    "volumes.kubernetes.io/controller-managed-attach-detach": "true"
                },
                "creationTimestamp": "2021-07-18T09:43:29Z",
                "labels": {
                    "agentpool": "agentpool",
                    "beta.kubernetes.io/arch": "amd64",
                    "beta.kubernetes.io/instance-type": "Standard_D2s_v3",
                    "beta.kubernetes.io/os": "linux",
                    "failure-domain.beta.kubernetes.io/region": "southcentralus",
                    "failure-domain.beta.kubernetes.io/zone": "1",
                    "kubernetes.azure.com/node-image-version": "AKSUbuntu-1804gen2-2021.06.12",
                    "kubernetes.azure.com/os-sku": "Ubuntu",
                    "kubernetes.azure.com/role": "agent",
                    "kubernetes.io/arch": "amd64",
                    "kubernetes.io/hostname": "aks-agentpool-30520393-vmss000001",
                    "kubernetes.io/os": "linux",
                    "kubernetes.io/role": "agent",
                    "node-role.kubernetes.io/agent": "",
                    "node.kubernetes.io/instance-type": "Standard_D2s_v3",
                    "storageprofile": "managed",
                    "storagetier": "Premium_LRS",
                    "topology.kubernetes.io/region": "southcentralus",
                    "topology.kubernetes.io/zone": "1"
                },
                "managedFields": [
                    {
                        "apiVersion": "v1",
                        "fieldsType": "FieldsV1",
                        "fieldsV1": {
                            "f:metadata": {
                                "f:labels": {
                                    "f:kubernetes.io/role": {},
                                    "f:node-role.kubernetes.io/agent": {}
                                }
                            }
                        },
                        "manager": "kubectl",
                        "operation": "Update",
                        "time": "2021-07-18T09:44:02Z"
                    },
                    {
                        "apiVersion": "v1",
                        "fieldsType": "FieldsV1",
                        "fieldsV1": {
                            "f:metadata": {
                                "f:annotations": {
                                    ".": {},
                                    "f:volumes.kubernetes.io/controller-managed-attach-detach": {}
                                },
                                "f:labels": {
                                    ".": {},
                                    "f:agentpool": {},
                                    "f:beta.kubernetes.io/arch": {},
                                    "f:beta.kubernetes.io/instance-type": {},
                                    "f:beta.kubernetes.io/os": {},
                                    "f:failure-domain.beta.kubernetes.io/region": {},
                                    "f:failure-domain.beta.kubernetes.io/zone": {},
                                    "f:kubernetes.azure.com/cluster": {},
                                    "f:kubernetes.azure.com/node-image-version": {},
                                    "f:kubernetes.azure.com/os-sku": {},
                                    "f:kubernetes.azure.com/role": {},
                                    "f:kubernetes.io/arch": {},
                                    "f:kubernetes.io/hostname": {},
                                    "f:kubernetes.io/os": {},
                                    "f:node.kubernetes.io/instance-type": {},
                                    "f:storageprofile": {},
                                    "f:storagetier": {},
                                    "f:topology.kubernetes.io/region": {},
                                    "f:topology.kubernetes.io/zone": {}
                                }
                            },
                            "f:spec": {
                                "f:providerID": {}
                            },
                            "f:status": {
                                "f:addresses": {
                                    ".": {},
                                    "k:{\"type\":\"Hostname\"}": {
                                        ".": {},
                                        "f:address": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"InternalIP\"}": {
                                        ".": {},
                                        "f:address": {},
                                        "f:type": {}
                                    }
                                },
                                "f:allocatable": {
                                    ".": {},
                                    "f:attachable-volumes-azure-disk": {},
                                    "f:cpu": {},
                                    "f:ephemeral-storage": {},
                                    "f:hugepages-1Gi": {},
                                    "f:hugepages-2Mi": {},
                                    "f:memory": {},
                                    "f:pods": {}
                                },
                                "f:capacity": {
                                    ".": {},
                                    "f:attachable-volumes-azure-disk": {},
                                    "f:cpu": {},
                                    "f:ephemeral-storage": {},
                                    "f:hugepages-1Gi": {},
                                    "f:hugepages-2Mi": {},
                                    "f:memory": {},
                                    "f:pods": {}
                                },
                                "f:conditions": {
                                    ".": {},
                                    "k:{\"type\":\"DiskPressure\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"MemoryPressure\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"PIDPressure\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"Ready\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    }
                                },
                                "f:config": {},
                                "f:daemonEndpoints": {
                                    "f:kubeletEndpoint": {
                                        "f:Port": {}
                                    }
                                },
                                "f:images": {},
                                "f:nodeInfo": {
                                    "f:architecture": {},
                                    "f:bootID": {},
                                    "f:containerRuntimeVersion": {},
                                    "f:kernelVersion": {},
                                    "f:kubeProxyVersion": {},
                                    "f:kubeletVersion": {},
                                    "f:machineID": {},
                                    "f:operatingSystem": {},
                                    "f:osImage": {},
                                    "f:systemUUID": {}
                                },
                                "f:volumesInUse": {}
                            }
                        },
                        "manager": "kubelet",
                        "operation": "Update",
                        "time": "2021-07-18T12:36:05Z"
                    },
                    {
                        "apiVersion": "v1",
                        "fieldsType": "FieldsV1",
                        "fieldsV1": {
                            "f:metadata": {
                                "f:annotations": {
                                    "f:node.alpha.kubernetes.io/ttl": {}
                                }
                            },
                            "f:spec": {
                                "f:podCIDR": {},
                                "f:podCIDRs": {
                                    ".": {},
                                    "v:\"10.244.2.0/24\"": {}
                                }
                            },
                            "f:status": {
                                "f:conditions": {
                                    "k:{\"type\":\"NetworkUnavailable\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    }
                                },
                                "f:volumesAttached": {}
                            }
                        },
                        "manager": "kube-controller-manager",
                        "operation": "Update",
                        "time": "2021-07-18T12:36:15Z"
                    }
                ],
                "name": "aks-agentpool-30520393-vmss000001",
                "resourceVersion": "25295719",
                "selfLink": "/api/v1/nodes/aks-agentpool-30520393-vmss000001",
                "uid": "3335beb8-ed1f-41cd-af49-9cbe3f063514"
            },
            "spec": {
                "podCIDR": "10.244.2.0/24",
                "podCIDRs": [
                    "10.244.2.0/24"
                ]
            },
            "status": {
                "addresses": [
                    {
                        "address": "aks-agentpool-30520393-vmss000001",
                        "type": "Hostname"
                    },
                    {
                        "address": "10.240.0.5",
                        "type": "InternalIP"
                    }
                ],
                "allocatable": {
                    "attachable-volumes-azure-disk": "4",
                    "cpu": "1900m",
                    "ephemeral-storage": "119716326407",
                    "hugepages-1Gi": "0",
                    "hugepages-2Mi": "0",
                    "memory": "5497568Ki",
                    "pods": "110"
                },
                "capacity": {
                    "attachable-volumes-azure-disk": "4",
                    "cpu": "2",
                    "ephemeral-storage": "129900528Ki",
                    "hugepages-1Gi": "0",
                    "hugepages-2Mi": "0",
                    "memory": "8152800Ki",
                    "pods": "110"
                },
                "conditions": [
                    {
                        "lastHeartbeatTime": "2021-07-18T09:43:50Z",
                        "lastTransitionTime": "2021-07-18T09:43:50Z",
                        "message": "RouteController created a route",
                        "reason": "RouteCreated",
                        "status": "False",
                        "type": "NetworkUnavailable"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:51:47Z",
                        "lastTransitionTime": "2021-07-18T09:43:29Z",
                        "message": "kubelet has sufficient memory available",
                        "reason": "KubeletHasSufficientMemory",
                        "status": "False",
                        "type": "MemoryPressure"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:51:47Z",
                        "lastTransitionTime": "2021-07-18T09:43:29Z",
                        "message": "kubelet has no disk pressure",
                        "reason": "KubeletHasNoDiskPressure",
                        "status": "False",
                        "type": "DiskPressure"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:51:47Z",
                        "lastTransitionTime": "2021-07-18T09:43:29Z",
                        "message": "kubelet has sufficient PID available",
                        "reason": "KubeletHasSufficientPID",
                        "status": "False",
                        "type": "PIDPressure"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:51:47Z",
                        "lastTransitionTime": "2021-07-18T09:43:39Z",
                        "message": "kubelet is posting ready status. AppArmor enabled",
                        "reason": "KubeletReady",
                        "status": "True",
                        "type": "Ready"
                    }
                ],
                "config": {},
                "daemonEndpoints": {
                    "kubeletEndpoint": {
                        "Port": 10211
                    }
                },
                "nodeInfo": {
                    "architecture": "amd64",
                    "bootID": "417a9d39-f61b-4cc7-b43b-1c58d2669f6f",
                    "containerRuntimeVersion": "docker://19.3.14",
                    "kernelVersion": "5.4.0-1049-azure",
                    "kubeProxyVersion": "v1.18.19",
                    "kubeletVersion": "v1.18.19",
                    "machineID": "643edc584d36482aa3134748a2d288b4",
                    "operatingSystem": "linux",
                    "osImage": "Ubuntu 18.04.5 LTS",
                    "systemUUID": "128c4972-c980-4043-8f62-ed3dc9afac31"
                }
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Node",
            "metadata": {
                "annotations": {
                    "node.alpha.kubernetes.io/ttl": "0",
                    "volumes.kubernetes.io/controller-managed-attach-detach": "true"
                },
                "creationTimestamp": "2021-07-18T09:42:45Z",
                "labels": {
                    "agentpool": "agentpool",
                    "beta.kubernetes.io/arch": "amd64",
                    "beta.kubernetes.io/instance-type": "Standard_D2s_v3",
                    "beta.kubernetes.io/os": "linux",
                    "failure-domain.beta.kubernetes.io/region": "southcentralus",
                    "failure-domain.beta.kubernetes.io/zone": "2",
                    "kubernetes.azure.com/node-image-version": "AKSUbuntu-1804gen2-2021.06.12",
                    "kubernetes.azure.com/os-sku": "Ubuntu",
                    "kubernetes.azure.com/role": "agent",
                    "kubernetes.io/arch": "amd64",
                    "kubernetes.io/hostname": "aks-agentpool-30520393-vmss000002",
                    "kubernetes.io/os": "linux",
                    "kubernetes.io/role": "agent",
                    "node-role.kubernetes.io/agent": "",
                    "node.kubernetes.io/instance-type": "Standard_D2s_v3",
                    "storageprofile": "managed",
                    "storagetier": "Premium_LRS",
                    "topology.kubernetes.io/region": "southcentralus",
                    "topology.kubernetes.io/zone": "2"
                },
                "managedFields": [
                    {
                        "apiVersion": "v1",
                        "fieldsType": "FieldsV1",
                        "fieldsV1": {
                            "f:metadata": {
                                "f:labels": {
                                    "f:kubernetes.io/role": {},
                                    "f:node-role.kubernetes.io/agent": {}
                                }
                            }
                        },
                        "manager": "kubectl",
                        "operation": "Update",
                        "time": "2021-07-18T09:43:02Z"
                    },
                    {
                        "apiVersion": "v1",
                        "fieldsType": "FieldsV1",
                        "fieldsV1": {
                            "f:metadata": {
                                "f:annotations": {
                                    ".": {},
                                    "f:volumes.kubernetes.io/controller-managed-attach-detach": {}
                                },
                                "f:labels": {
                                    ".": {},
                                    "f:agentpool": {},
                                    "f:beta.kubernetes.io/arch": {},
                                    "f:beta.kubernetes.io/instance-type": {},
                                    "f:beta.kubernetes.io/os": {},
                                    "f:failure-domain.beta.kubernetes.io/region": {},
                                    "f:failure-domain.beta.kubernetes.io/zone": {},
                                    "f:kubernetes.azure.com/cluster": {},
                                    "f:kubernetes.azure.com/node-image-version": {},
                                    "f:kubernetes.azure.com/os-sku": {},
                                    "f:kubernetes.azure.com/role": {},
                                    "f:kubernetes.io/arch": {},
                                    "f:kubernetes.io/hostname": {},
                                    "f:kubernetes.io/os": {},
                                    "f:node.kubernetes.io/instance-type": {},
                                    "f:storageprofile": {},
                                    "f:storagetier": {},
                                    "f:topology.kubernetes.io/region": {},
                                    "f:topology.kubernetes.io/zone": {}
                                }
                            },
                            "f:spec": {
                                "f:providerID": {}
                            },
                            "f:status": {
                                "f:addresses": {
                                    ".": {},
                                    "k:{\"type\":\"Hostname\"}": {
                                        ".": {},
                                        "f:address": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"InternalIP\"}": {
                                        ".": {},
                                        "f:address": {},
                                        "f:type": {}
                                    }
                                },
                                "f:allocatable": {
                                    ".": {},
                                    "f:attachable-volumes-azure-disk": {},
                                    "f:cpu": {},
                                    "f:ephemeral-storage": {},
                                    "f:hugepages-1Gi": {},
                                    "f:hugepages-2Mi": {},
                                    "f:memory": {},
                                    "f:pods": {}
                                },
                                "f:capacity": {
                                    ".": {},
                                    "f:attachable-volumes-azure-disk": {},
                                    "f:cpu": {},
                                    "f:ephemeral-storage": {},
                                    "f:hugepages-1Gi": {},
                                    "f:hugepages-2Mi": {},
                                    "f:memory": {},
                                    "f:pods": {}
                                },
                                "f:conditions": {
                                    ".": {},
                                    "k:{\"type\":\"DiskPressure\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"MemoryPressure\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"PIDPressure\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    },
                                    "k:{\"type\":\"Ready\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    }
                                },
                                "f:config": {},
                                "f:daemonEndpoints": {
                                    "f:kubeletEndpoint": {
                                        "f:Port": {}
                                    }
                                },
                                "f:images": {},
                                "f:nodeInfo": {
                                    "f:architecture": {},
                                    "f:bootID": {},
                                    "f:containerRuntimeVersion": {},
                                    "f:kernelVersion": {},
                                    "f:kubeProxyVersion": {},
                                    "f:kubeletVersion": {},
                                    "f:machineID": {},
                                    "f:operatingSystem": {},
                                    "f:osImage": {},
                                    "f:systemUUID": {}
                                },
                                "f:volumesInUse": {}
                            }
                        },
                        "manager": "kubelet",
                        "operation": "Update",
                        "time": "2021-07-18T12:34:38Z"
                    },
                    {
                        "apiVersion": "v1",
                        "fieldsType": "FieldsV1",
                        "fieldsV1": {
                            "f:metadata": {
                                "f:annotations": {
                                    "f:node.alpha.kubernetes.io/ttl": {}
                                }
                            },
                            "f:spec": {
                                "f:podCIDR": {},
                                "f:podCIDRs": {
                                    ".": {},
                                    "v:\"10.244.0.0/24\"": {}
                                }
                            },
                            "f:status": {
                                "f:conditions": {
                                    "k:{\"type\":\"NetworkUnavailable\"}": {
                                        ".": {},
                                        "f:lastHeartbeatTime": {},
                                        "f:lastTransitionTime": {},
                                        "f:message": {},
                                        "f:reason": {},
                                        "f:status": {},
                                        "f:type": {}
                                    }
                                },
                                "f:volumesAttached": {}
                            }
                        },
                        "manager": "kube-controller-manager",
                        "operation": "Update",
                        "time": "2021-07-18T12:34:45Z"
                    }
                ],
                "name": "aks-agentpool-30520393-vmss000002",
                "resourceVersion": "25295824",
                "selfLink": "/api/v1/nodes/aks-agentpool-30520393-vmss000002",
                "uid": "0dc68414-ce91-4115-8fc7-364f56449572"
            },
            "spec": {
                "podCIDR": "10.244.0.0/24",
                "podCIDRs": [
                    "10.244.0.0/24"
                ]
            },
            "status": {
                "addresses": [
                    {
                        "address": "aks-agentpool-30520393-vmss000002",
                        "type": "Hostname"
                    },
                    {
                        "address": "10.240.0.6",
                        "type": "InternalIP"
                    }
                ],
                "allocatable": {
                    "attachable-volumes-azure-disk": "4",
                    "cpu": "1900m",
                    "ephemeral-storage": "119716326407",
                    "hugepages-1Gi": "0",
                    "hugepages-2Mi": "0",
                    "memory": "5497564Ki",
                    "pods": "110"
                },
                "capacity": {
                    "attachable-volumes-azure-disk": "4",
                    "cpu": "2",
                    "ephemeral-storage": "129900528Ki",
                    "hugepages-1Gi": "0",
                    "hugepages-2Mi": "0",
                    "memory": "8152796Ki",
                    "pods": "110"
                },
                "conditions": [
                    {
                        "lastHeartbeatTime": "2021-07-18T09:43:09Z",
                        "lastTransitionTime": "2021-07-18T09:43:09Z",
                        "message": "RouteController created a route",
                        "reason": "RouteCreated",
                        "status": "False",
                        "type": "NetworkUnavailable"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:52:16Z",
                        "lastTransitionTime": "2021-07-18T09:42:45Z",
                        "message": "kubelet has sufficient memory available",
                        "reason": "KubeletHasSufficientMemory",
                        "status": "False",
                        "type": "MemoryPressure"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:52:16Z",
                        "lastTransitionTime": "2021-07-18T09:42:45Z",
                        "message": "kubelet has no disk pressure",
                        "reason": "KubeletHasNoDiskPressure",
                        "status": "False",
                        "type": "DiskPressure"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:52:16Z",
                        "lastTransitionTime": "2021-07-18T09:42:45Z",
                        "message": "kubelet has sufficient PID available",
                        "reason": "KubeletHasSufficientPID",
                        "status": "False",
                        "type": "PIDPressure"
                    },
                    {
                        "lastHeartbeatTime": "2021-10-05T21:52:16Z",
                        "lastTransitionTime": "2021-07-18T09:42:49Z",
                        "message": "kubelet is posting ready status. AppArmor enabled",
                        "reason": "KubeletReady",
                        "status": "True",
                        "type": "Ready"
                    }
                ],
                "config": {},
                "daemonEndpoints": {
                    "kubeletEndpoint": {
                        "Port": 10211
                    }
                },
                "nodeInfo": {
                    "architecture": "amd64",
                    "bootID": "dc5a2572-50c4-4eca-abef-e351a27b5ea2",
                    "containerRuntimeVersion": "docker://19.3.14",
                    "kernelVersion": "5.4.0-1049-azure",
                    "kubeProxyVersion": "v1.18.19",
                    "kubeletVersion": "v1.18.19",
                    "machineID": "a0f3ff3098744fe795119b2019c06253",
                    "operatingSystem": "linux",
                    "osImage": "Ubuntu 18.04.5 LTS",
                    "systemUUID": "4c191a84-2005-45b5-8f39-fc647c71a75f"
                }
            }
        }
    ],
    "kind": "List",
    "metadata": {
        "resourceVersion": "",
        "selfLink": ""
    }
}
`
