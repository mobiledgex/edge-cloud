package testutil

import (
	fmt "fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

var FlavorData = []edgeproto.Flavor{
	edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.tiny",
		},
		Ram:   1024,
		Vcpus: 1,
		Disk:  1,
	},
	edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.small",
		},
		Ram:   2048,
		Vcpus: 2,
		Disk:  2,
	},
	edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.medium",
		},
		Ram:   4096,
		Vcpus: 4,
		Disk:  4,
	},
	edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.large",
		},
		Ram:   8192,
		Vcpus: 10,
		Disk:  40,
	},
	edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.tiny.gpu",
		},
		Ram:   1024,
		Vcpus: 1,
		Disk:  1,
		OptResMap: map[string]string{
			"gpu": "pci:1",
		},
	},
}

var DevData = []string{
	"NianticInc",
	"Ever.ai",
	"1000realities",
	"SierrawareLLC",
}
var ClusterKeys = []edgeproto.ClusterKey{
	edgeproto.ClusterKey{
		Name: "Pokemons",
	},
	edgeproto.ClusterKey{
		Name: "Ever.Ai",
	},
	edgeproto.ClusterKey{
		Name: "1000realities",
	},
	edgeproto.ClusterKey{
		Name: "Big-Pokemons",
	},
	edgeproto.ClusterKey{
		Name: "Reservable",
	},
	edgeproto.ClusterKey{
		Name: "defaultmtclust", // cloudcommon.DefaultMultiTenantCluster
	},
}

var AppData = []edgeproto.App{
	edgeproto.App{ // 0
		Key: edgeproto.AppKey{
			Organization: DevData[0],
			Name:         "Pokemon Go!",
			Version:      "1.0.0",
		},
		ImageType:       edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:     "tcp:443,tcp:10002,udp:10002",
		AccessType:      edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor:   FlavorData[0].Key,
		AllowServerless: true,
		ServerlessConfig: &edgeproto.ServerlessConfig{
			Vcpus: 0.5,
			Ram:   20,
		},
	},
	edgeproto.App{ // 1
		Key: edgeproto.AppKey{
			Organization: DevData[0],
			Name:         "Pokemon Go!",
			Version:      "1.0.1",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:   "tcp:80,tcp:443,tcp:81:tls",
		AccessType:    edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: FlavorData[0].Key,
	},
	edgeproto.App{ // 2
		Key: edgeproto.AppKey{
			Organization: DevData[0],
			Name:         "Harry Potter Go! Go!",
			Version:      "0.0.1",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:   "tcp:443,udp:11111",
		AccessType:    edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: FlavorData[1].Key,
	},
	edgeproto.App{ // 3
		Key: edgeproto.AppKey{
			Organization: DevData[1],
			Name:         "AI",
			Version:      "1.2.0",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_QCOW,
		ImagePath:     "http://somerepo/image/path/ai/1.2.0#md5:7e9cfcb763e83573a4b9d9315f56cc5f",
		AccessPorts:   "tcp:8080",
		AccessType:    edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: FlavorData[1].Key,
	},
	edgeproto.App{ // 4
		Key: edgeproto.AppKey{
			Organization: DevData[2],
			Name:         "my reality",
			Version:      "0.0.1",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_QCOW,
		ImagePath:     "http://somerepo/image/path/myreality/0.0.1#md5:7e9cfcb763e83573a4b9d9315f56cc5f",
		AccessPorts:   "udp:1024",
		AccessType:    edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: FlavorData[2].Key,
	},
	edgeproto.App{ // 5
		Key: edgeproto.AppKey{
			Organization: DevData[3],
			Name:         "helmApp",
			Version:      "0.0.1",
		},
		Deployment:    "helm",
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_HELM,
		AccessPorts:   "udp:2024",
		AccessType:    edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: FlavorData[2].Key,
	},
	edgeproto.App{ // 6
		Key: edgeproto.AppKey{
			Organization: DevData[0],
			Name:         "Neon",
			Version:      "0.0.2",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:   "tcp:80,udp:8001,tcp:065535",
		AccessType:    edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: FlavorData[1].Key,
	},
	edgeproto.App{ // 7
		Key: edgeproto.AppKey{
			Organization: DevData[0],
			Name:         "NoPorts",
			Version:      "1.0.0",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessType:    edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: FlavorData[0].Key,
	},
	edgeproto.App{ // 8
		Key: edgeproto.AppKey{
			Organization: DevData[0],
			Name:         "PortRangeApp",
			Version:      "1.0.0",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:   "tcp:80,tcp:443,udp:10002,tcp:5000-5002", // new port range notation
		AccessType:    edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: FlavorData[0].Key,
	},
	edgeproto.App{ // 9
		Key: edgeproto.AppKey{
			Organization: "MobiledgeX", // cloudcommon.OrganizationMobiledgeX
			Name:         "AutoDeleteApp",
			Version:      "1.0.0",
		},
		ImageType:       edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessType:      edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor:   FlavorData[0].Key,
		DelOpt:          edgeproto.DeleteType_AUTO_DELETE,
		AllowServerless: true,
		ServerlessConfig: &edgeproto.ServerlessConfig{
			Vcpus: 0.2,
			Ram:   10,
		},
	},
	edgeproto.App{ // 10
		Key: edgeproto.AppKey{
			Organization: DevData[1],
			Name:         "Dev1App",
			Version:      "0.0.1",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:   "tcp:443,udp:11111",
		AccessType:    edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: FlavorData[1].Key,
	},
	edgeproto.App{ // 11
		Key: edgeproto.AppKey{
			Organization: DevData[0],
			Name:         "Pokemon Go!",
			Version:      "1.0.2",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:   "tcp:10003",
		AccessType:    edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: FlavorData[0].Key,
		AutoProvPolicies: []string{
			AutoProvPolicyData[0].Key.Name,
			AutoProvPolicyData[3].Key.Name,
		},
	},
	edgeproto.App{ // 12
		Key: edgeproto.AppKey{
			Organization: DevData[0],
			Name:         "vm lb",
			Version:      "1.0.2",
		},
		Deployment:    "vm",
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_QCOW,
		ImagePath:     "http://somerepo/image/path/myreality/0.0.1#md5:7e9cfcb763e83573a4b9d9315f56cc5f",
		AccessPorts:   "tcp:10003",
		AccessType:    edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor: FlavorData[0].Key,
	},
	edgeproto.App{ // 13 - MobiledgeX app
		Key: edgeproto.AppKey{
			Organization: "MobiledgeX", // cloudcommon.OrganizationMobiledgeX
			Name:         "SampleApp",
			Version:      "1.0.0",
		},
		ImageType:       edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessType:      edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		AccessPorts:     "tcp:889",
		DefaultFlavor:   FlavorData[0].Key,
		AllowServerless: true,
		ServerlessConfig: &edgeproto.ServerlessConfig{
			Vcpus: 0.5,
			Ram:   20,
		},
	},
	edgeproto.App{ // 14
		Key: edgeproto.AppKey{
			Organization: DevData[0],
			Name:         "Pokemon MT",
			Version:      "1.0.0",
		},
		ImageType:       edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:     "tcp:444",
		AccessType:      edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER,
		DefaultFlavor:   FlavorData[0].Key,
		AllowServerless: true,
		ServerlessConfig: &edgeproto.ServerlessConfig{
			Vcpus: 0.5,
			Ram:   20,
		},
	},
}
var OperatorData = []string{
	"AT&T Inc.",
	"T-Mobile",
	"Verizon",
	"Deutsche Telekom",
}

var OperatorCodeData = []edgeproto.OperatorCode{
	edgeproto.OperatorCode{
		Code:         "31170",
		Organization: "AT&T Inc.",
	},
	edgeproto.OperatorCode{
		Code:         "31026",
		Organization: "T-Mobile",
	},
	edgeproto.OperatorCode{
		Code:         "310110",
		Organization: "Verizon",
	},
	edgeproto.OperatorCode{
		Code:         "2621",
		Organization: "Deutsche Telekom",
	},
}

var CloudletData = []edgeproto.Cloudlet{
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Organization: OperatorData[0],
			Name:         "San Jose Site",
		},
		IpSupport:     edgeproto.IpSupport_IP_SUPPORT_DYNAMIC,
		NumDynamicIps: 100,
		Location: dme.Loc{
			Latitude:  37.338207,
			Longitude: -121.886330,
		},
		PlatformType:                  edgeproto.PlatformType_PLATFORM_TYPE_FAKE,
		Flavor:                        FlavorData[0].Key,
		NotifySrvAddr:                 "127.0.0.1:51001",
		CrmOverride:                   edgeproto.CRMOverride_IGNORE_CRM,
		PhysicalName:                  "SanJoseSite",
		Deployment:                    "docker",
		DefaultResourceAlertThreshold: 80,
		ResourceQuotas: []edgeproto.ResourceQuota{
			edgeproto.ResourceQuota{
				Name:           "GPUs",
				Value:          10,
				AlertThreshold: 10,
			},
			edgeproto.ResourceQuota{
				Name:           "RAM",
				AlertThreshold: 30,
			},
			edgeproto.ResourceQuota{
				Name:           "vCPUs",
				Value:          99,
				AlertThreshold: 20,
			},
		},
		ResTagMap: map[string]*edgeproto.ResTagTableKey{
			"gpu": &Restblkeys[3],
		},
		GpuConfig: edgeproto.GPUConfig{
			Driver: GPUDriverData[0].Key,
		},
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Organization: OperatorData[0],
			Name:         "New York Site",
		},
		IpSupport:     edgeproto.IpSupport_IP_SUPPORT_DYNAMIC,
		NumDynamicIps: 100,
		Location: dme.Loc{
			Latitude:  40.712776,
			Longitude: -74.005974,
		},
		PlatformType:                  edgeproto.PlatformType_PLATFORM_TYPE_FAKE,
		Flavor:                        FlavorData[0].Key,
		NotifySrvAddr:                 "127.0.0.1:51002",
		CrmOverride:                   edgeproto.CRMOverride_IGNORE_CRM,
		PhysicalName:                  "NewYorkSite",
		Deployment:                    "docker",
		DefaultResourceAlertThreshold: 80,
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Organization: OperatorData[1],
			Name:         "San Francisco Site",
		},
		IpSupport:     edgeproto.IpSupport_IP_SUPPORT_DYNAMIC,
		NumDynamicIps: 100,
		Location: dme.Loc{
			Latitude:  37.774929,
			Longitude: -122.419418,
		},
		Flavor:         FlavorData[0].Key,
		PlatformType:   edgeproto.PlatformType_PLATFORM_TYPE_FAKE,
		NotifySrvAddr:  "127.0.0.1:51003",
		InfraApiAccess: edgeproto.InfraApiAccess_RESTRICTED_ACCESS,
		InfraConfig: edgeproto.InfraConfig{
			FlavorName:          FlavorData[0].Key.Name,
			ExternalNetworkName: "testnet",
		},
		PhysicalName:                  "SanFranciscoSite",
		Deployment:                    "docker",
		DefaultResourceAlertThreshold: 80,
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Organization: OperatorData[2],
			Name:         "Hawaii Site",
		},
		IpSupport:     edgeproto.IpSupport_IP_SUPPORT_DYNAMIC,
		NumDynamicIps: 10,
		Location: dme.Loc{
			Latitude:  21.306944,
			Longitude: -157.858337,
		},
		Flavor:                        FlavorData[0].Key,
		PlatformType:                  edgeproto.PlatformType_PLATFORM_TYPE_FAKE,
		NotifySrvAddr:                 "127.0.0.1:51004",
		CrmOverride:                   edgeproto.CRMOverride_IGNORE_CRM,
		PhysicalName:                  "HawaiiSite",
		Deployment:                    "docker",
		DefaultResourceAlertThreshold: 80,
	},
}
var ClusterInstData = []edgeproto.ClusterInst{
	edgeproto.ClusterInst{ // 0
		Key: edgeproto.ClusterInstKey{
			ClusterKey:   ClusterKeys[0],
			CloudletKey:  CloudletData[0].Key,
			Organization: DevData[0],
		},
		Flavor:     FlavorData[0].Key,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_DEDICATED,
		NumMasters: 1,
		NumNodes:   2,
	},
	edgeproto.ClusterInst{ // 1
		Key: edgeproto.ClusterInstKey{
			ClusterKey:   ClusterKeys[0],
			CloudletKey:  CloudletData[1].Key,
			Organization: DevData[0],
		},
		Flavor:     FlavorData[0].Key,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_SHARED,
		NumMasters: 1,
		NumNodes:   2,
	},
	edgeproto.ClusterInst{ // 2
		Key: edgeproto.ClusterInstKey{
			ClusterKey:   ClusterKeys[0],
			CloudletKey:  CloudletData[2].Key,
			Organization: DevData[3],
		},
		Flavor:          FlavorData[0].Key,
		NumMasters:      1,
		NumNodes:        2,
		AutoScalePolicy: AutoScalePolicyData[2].Key.Name,
	},
	edgeproto.ClusterInst{ // 3
		Key: edgeproto.ClusterInstKey{
			ClusterKey:   ClusterKeys[1],
			CloudletKey:  CloudletData[0].Key,
			Organization: DevData[0],
		},
		Flavor:          FlavorData[1].Key,
		IpAccess:        edgeproto.IpAccess_IP_ACCESS_DEDICATED,
		NumMasters:      1,
		NumNodes:        3,
		AutoScalePolicy: AutoScalePolicyData[0].Key.Name,
	},
	edgeproto.ClusterInst{ // 4
		Key: edgeproto.ClusterInstKey{
			ClusterKey:   ClusterKeys[1],
			CloudletKey:  CloudletData[1].Key,
			Organization: DevData[3],
		},
		Flavor:     FlavorData[1].Key,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_SHARED,
		NumMasters: 1,
		NumNodes:   3,
	},
	edgeproto.ClusterInst{ // 5
		Key: edgeproto.ClusterInstKey{
			ClusterKey:   ClusterKeys[2],
			CloudletKey:  CloudletData[2].Key,
			Organization: DevData[3],
		},
		Flavor:     FlavorData[2].Key,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_DEDICATED,
		NumMasters: 1,
		NumNodes:   4,
	},
	edgeproto.ClusterInst{ // 6
		Key: edgeproto.ClusterInstKey{
			ClusterKey:   ClusterKeys[3],
			CloudletKey:  CloudletData[3].Key,
			Organization: DevData[3],
		},
		Flavor:     FlavorData[2].Key,
		NumMasters: 1,
		NumNodes:   3,
	},
	edgeproto.ClusterInst{ // 7
		Key: edgeproto.ClusterInstKey{
			ClusterKey:   ClusterKeys[4],
			CloudletKey:  CloudletData[0].Key,
			Organization: "MobiledgeX", // cloudcommon.OrganizationMobiledgeX
		},
		Flavor:     FlavorData[0].Key,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_SHARED,
		NumMasters: 1,
		NumNodes:   2,
		Reservable: true,
	},
	edgeproto.ClusterInst{ // 8
		Key: edgeproto.ClusterInstKey{
			ClusterKey:   ClusterKeys[5], // multi-tenant cluster
			CloudletKey:  CloudletData[0].Key,
			Organization: "MobiledgeX", // cloudcommon.OrganizationMobiledgeX
		},
		Flavor:           FlavorData[0].Key,
		IpAccess:         edgeproto.IpAccess_IP_ACCESS_SHARED,
		NumMasters:       1,
		NumNodes:         5,
		MasterNodeFlavor: FlavorData[2].Key.Name, // medium
		MultiTenant:      true,
	},
}

// These are the cluster insts that will be created automatically
// from appinsts that have not specified a cluster.
var ClusterInstAutoData = []edgeproto.ClusterInst{
	// from AppInstData[3] -> AppData[1]
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey: edgeproto.ClusterKey{
				Name: "reservable0",
			},
			CloudletKey:  CloudletData[1].Key,
			Organization: "MobiledgeX", // cloudcommon.OrganizationMobiledgeX
		},
		Flavor:     FlavorData[0].Key,
		NumMasters: 1,
		NumNodes:   1,
		State:      edgeproto.TrackedState_READY,
		Auto:       true,
		Reservable: true,
		ReservedBy: DevData[0],
	},
	// from AppInstData[4] -> AppData[2]
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey: edgeproto.ClusterKey{
				Name: "reservable0",
			},
			CloudletKey:  CloudletData[2].Key,
			Organization: "MobiledgeX", // cloudcommon.OrganizationMobiledgeX
		},
		Flavor:     FlavorData[1].Key,
		NumMasters: 1,
		NumNodes:   1,
		State:      edgeproto.TrackedState_READY,
		Auto:       true,
		Reservable: true,
		ReservedBy: DevData[0],
	},
	// from AppInstData[6] -> AppData[6]
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey: edgeproto.ClusterKey{
				Name: "reservable1",
			},
			CloudletKey:  CloudletData[2].Key,
			Organization: "MobiledgeX", // cloudcommon.OrganizationMobiledgeX
		},
		Flavor:     FlavorData[1].Key,
		NumMasters: 1,
		NumNodes:   1,
		State:      edgeproto.TrackedState_READY,
		Auto:       true,
		Reservable: true,
		ReservedBy: DevData[0],
	},
	// from AppInstData[12] -> AppData[13]
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey: edgeproto.ClusterKey{
				Name: "reservable0",
			},
			CloudletKey:  CloudletData[3].Key,
			Organization: "MobiledgeX", // cloudcommon.OrganizationMobiledgeX
		},
		Flavor:     FlavorData[0].Key,
		NumMasters: 1,
		NumNodes:   1,
		State:      edgeproto.TrackedState_READY,
		Auto:       true,
		Reservable: true,
		ReservedBy: "MobiledgeX",
	},
}
var AppInstData = []edgeproto.AppInst{
	edgeproto.AppInst{ // 0
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[0].Key,
			ClusterInstKey: *ClusterInstData[0].Key.Virtual(""),
		},
		CloudletLoc: CloudletData[0].Location,
	},
	edgeproto.AppInst{ // 1
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[0].Key,
			ClusterInstKey: *ClusterInstData[3].Key.Virtual(""),
		},
		CloudletLoc: CloudletData[0].Location,
	},
	edgeproto.AppInst{ // 2
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[0].Key,
			ClusterInstKey: *ClusterInstData[1].Key.Virtual(""),
		},
		CloudletLoc: CloudletData[1].Location,
	},
	edgeproto.AppInst{ // 3
		Key: edgeproto.AppInstKey{
			AppKey: AppData[1].Key,
			// ClusterInst is ClusterInstAutoData[0]
			ClusterInstKey: *ClusterInstAutoData[0].Key.Virtual(util.K8SSanitize("autocluster" + AppData[1].Key.Name)),
		},
		CloudletLoc: CloudletData[1].Location,
	},
	edgeproto.AppInst{ // 4
		Key: edgeproto.AppInstKey{
			AppKey: AppData[2].Key,
			// ClusterInst is ClusterInstAutoData[1]
			ClusterInstKey: *ClusterInstAutoData[1].Key.Virtual(util.K8SSanitize("autocluster" + AppData[2].Key.Name)),
		},
		CloudletLoc: CloudletData[2].Location,
	},
	edgeproto.AppInst{ // 5
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[5].Key,
			ClusterInstKey: *ClusterInstData[2].Key.Virtual(""),
		},
		CloudletLoc: CloudletData[2].Location,
	},
	edgeproto.AppInst{ // 6
		Key: edgeproto.AppInstKey{
			AppKey: AppData[6].Key,
			// ClusterInst is ClusterInstAutoData[2]
			ClusterInstKey: *ClusterInstAutoData[2].Key.Virtual(util.K8SSanitize("autocluster" + AppData[6].Key.Name)),
		},
		CloudletLoc: CloudletData[2].Location,
	},
	edgeproto.AppInst{ // 7
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[6].Key,
			ClusterInstKey: *ClusterInstData[0].Key.Virtual(""),
		},
		CloudletLoc: CloudletData[0].Location,
	},
	edgeproto.AppInst{ // 8
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[7].Key,
			ClusterInstKey: *ClusterInstData[0].Key.Virtual(""),
		},
		CloudletLoc: CloudletData[0].Location,
	},
	edgeproto.AppInst{ // 9
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[9].Key, //auto-delete app
			ClusterInstKey: *ClusterInstData[0].Key.Virtual(""),
		},
		CloudletLoc: CloudletData[0].Location,
	},
	edgeproto.AppInst{ // 10
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[9].Key, //auto-delete app
			ClusterInstKey: *ClusterInstAutoData[0].Key.Virtual(""),
		},
		CloudletLoc:     CloudletData[1].Location,
		RealClusterName: ClusterInstAutoData[0].Key.ClusterKey.Name,
	},
	edgeproto.AppInst{ // 11
		Key: edgeproto.AppInstKey{
			AppKey: AppData[12].Key, //vm app with lb
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
				CloudletKey: CloudletData[0].Key,
			},
		},
		CloudletLoc: CloudletData[1].Location,
	},
	edgeproto.AppInst{ // 12 - deploy MobiledgeX app to reservable autocluster
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[13].Key, // mobiledgex sample app
			ClusterInstKey: *ClusterInstAutoData[3].Key.Virtual(util.K8SSanitize("autocluster" + AppData[13].Key.Name)),
		},
		CloudletLoc: CloudletData[3].Location,
	},
	edgeproto.AppInst{ // 13
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[0].Key,
			ClusterInstKey: *ClusterInstData[8].Key.Virtual("autocluster-mt1"),
		},
		CloudletLoc: CloudletData[0].Location,
	},
	edgeproto.AppInst{ // 14
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[9].Key, // sidecar app
			ClusterInstKey: *ClusterInstData[8].Key.Virtual(""),
		},
		CloudletLoc: CloudletData[0].Location,
	},
	edgeproto.AppInst{ // 15
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[13].Key,
			ClusterInstKey: *ClusterInstData[8].Key.Virtual("autocluster-mt3"),
		},
		CloudletLoc: CloudletData[0].Location,
	},
	edgeproto.AppInst{ // 16
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[14].Key,
			ClusterInstKey: *ClusterInstData[8].Key.Virtual("autocluster-mt2"),
		},
		CloudletLoc: CloudletData[0].Location,
	},
}

var AppInstInfoData = []edgeproto.AppInstInfo{
	edgeproto.AppInstInfo{
		Key: AppInstData[0].Key,
	},
	edgeproto.AppInstInfo{
		Key: AppInstData[1].Key,
	},
	edgeproto.AppInstInfo{
		Key: AppInstData[2].Key,
	},
	edgeproto.AppInstInfo{
		Key: AppInstData[3].Key,
	},
	edgeproto.AppInstInfo{
		Key: AppInstData[4].Key,
	},
	edgeproto.AppInstInfo{
		Key: AppInstData[5].Key,
	},
	edgeproto.AppInstInfo{
		Key: AppInstData[6].Key,
	},
	edgeproto.AppInstInfo{
		Key: AppInstData[7].Key,
	},
}

func GetAppInstRefsData() []edgeproto.AppInstRefs {
	createdAppInsts := CreatedAppInstData()
	return []edgeproto.AppInstRefs{
		edgeproto.AppInstRefs{
			Key: AppData[0].Key,
			Insts: map[string]uint32{
				AppInstData[0].Key.GetKeyString():  1,
				AppInstData[1].Key.GetKeyString():  1,
				AppInstData[2].Key.GetKeyString():  1,
				AppInstData[13].Key.GetKeyString(): 1,
			},
		},
		edgeproto.AppInstRefs{
			Key: AppData[1].Key,
			Insts: map[string]uint32{
				AppInstData[3].Key.GetKeyString(): 1,
			},
		},
		edgeproto.AppInstRefs{
			Key: AppData[2].Key,
			Insts: map[string]uint32{
				AppInstData[4].Key.GetKeyString(): 1,
			},
		},
		edgeproto.AppInstRefs{
			Key:   AppData[3].Key,
			Insts: map[string]uint32{},
		},
		edgeproto.AppInstRefs{
			Key:   AppData[4].Key,
			Insts: map[string]uint32{},
		},
		edgeproto.AppInstRefs{
			Key: AppData[5].Key,
			Insts: map[string]uint32{
				AppInstData[5].Key.GetKeyString(): 1,
			},
		},
		edgeproto.AppInstRefs{
			Key: AppData[6].Key,
			Insts: map[string]uint32{
				AppInstData[6].Key.GetKeyString(): 1,
				AppInstData[7].Key.GetKeyString(): 1,
			},
		},
		edgeproto.AppInstRefs{
			Key: AppData[7].Key,
			Insts: map[string]uint32{
				AppInstData[8].Key.GetKeyString(): 1,
			},
		},
		edgeproto.AppInstRefs{
			Key:   AppData[8].Key,
			Insts: map[string]uint32{},
		},
		edgeproto.AppInstRefs{
			Key: AppData[9].Key,
			Insts: map[string]uint32{
				AppInstData[9].Key.GetKeyString():  1,
				AppInstData[10].Key.GetKeyString(): 1,
				AppInstData[14].Key.GetKeyString(): 1,
			},
		},
		edgeproto.AppInstRefs{
			Key:   AppData[10].Key,
			Insts: map[string]uint32{},
		},
		edgeproto.AppInstRefs{
			Key:   AppData[11].Key,
			Insts: map[string]uint32{},
		},
		edgeproto.AppInstRefs{
			Key: AppData[12].Key,
			Insts: map[string]uint32{
				createdAppInsts[11].Key.GetKeyString(): 1,
			},
		},
		edgeproto.AppInstRefs{
			Key: AppData[13].Key,
			Insts: map[string]uint32{
				AppInstData[12].Key.GetKeyString(): 1,
				AppInstData[15].Key.GetKeyString(): 1,
			},
		},
		edgeproto.AppInstRefs{
			Key: AppData[14].Key,
			Insts: map[string]uint32{
				AppInstData[16].Key.GetKeyString(): 1,
			},
		},
	}
}

var CloudletInfoData = []edgeproto.CloudletInfo{
	edgeproto.CloudletInfo{
		Key:         CloudletData[0].Key,
		State:       dme.CloudletState_CLOUDLET_STATE_READY,
		OsMaxRam:    65536,
		OsMaxVcores: 16,
		OsMaxVolGb:  500,
		Flavors: []*edgeproto.FlavorInfo{
			&edgeproto.FlavorInfo{
				Name:  "flavor.tiny1",
				Vcpus: uint64(1),
				Ram:   uint64(512),
				Disk:  uint64(10),
			},
			&edgeproto.FlavorInfo{
				Name:  "flavor.tiny2",
				Vcpus: uint64(1),
				Ram:   uint64(1024),
				Disk:  uint64(10),
			},
			&edgeproto.FlavorInfo{
				Name:  "flavor.small",
				Vcpus: uint64(2),
				Ram:   uint64(1024),
				Disk:  uint64(20),
			},
			&edgeproto.FlavorInfo{
				Name:  "flavor.medium",
				Vcpus: uint64(4),
				Ram:   uint64(4096),
				Disk:  uint64(40),
			},
			&edgeproto.FlavorInfo{
				Name:  "flavor.lg-master",
				Vcpus: uint64(4),
				Ram:   uint64(8192),
				Disk:  uint64(60),
			},
			// restagtbl/clouldlet resource map tests
			&edgeproto.FlavorInfo{
				Name:    "flavor.large",
				Vcpus:   uint64(10),
				Ram:     uint64(8192),
				Disk:    uint64(40),
				PropMap: map[string]string{"pci_passthrough": "alias=t4:1"},
			},
			&edgeproto.FlavorInfo{
				Name:    "flavor.large2",
				Vcpus:   uint64(10),
				Ram:     uint64(8192),
				Disk:    uint64(40),
				PropMap: map[string]string{"pci_passthrough": "alias=t4:1", "nas": "ceph-20:1"},
			},
			&edgeproto.FlavorInfo{
				Name:    "flavor.large-pci",
				Vcpus:   uint64(10),
				Ram:     uint64(8192),
				Disk:    uint64(40),
				PropMap: map[string]string{"pci": "NP4:1"},
			},
			&edgeproto.FlavorInfo{
				Name:    "flavor.large-nvidia",
				Vcpus:   uint64(10),
				Ram:     uint64(8192),
				Disk:    uint64(40),
				PropMap: map[string]string{"vgpu": "nvidia-63:1"},
			},
			&edgeproto.FlavorInfo{
				Name:    "flavor.large-generic-gpu",
				Vcpus:   uint64(10),
				Ram:     uint64(8192),
				Disk:    uint64(80),
				PropMap: map[string]string{"vmware": "vgpu=1"},
			},
			// A typical case where two flavors are identical in their
			// nominal resources, and differ only by gpu vs vgpu
			// These cases require a fully qualifed request in the mex flavors optresmap
			&edgeproto.FlavorInfo{
				Name:    "flavor.m4.large-gpu",
				Vcpus:   uint64(12),
				Ram:     uint64(4096),
				Disk:    uint64(20),
				PropMap: map[string]string{"pci_passthrough": "alias=t4gpu:1"},
			},
			&edgeproto.FlavorInfo{
				Name:    "flavor.m4.large-vgpu",
				Vcpus:   uint64(12),
				Ram:     uint64(4096),
				Disk:    uint64(20),
				PropMap: map[string]string{"resources": "VGPU=1"},
			},
		},
		ResourcesSnapshot: edgeproto.InfraResourcesSnapshot{
			Info: []edgeproto.InfraResource{
				edgeproto.InfraResource{
					Name:          "RAM",
					Value:         uint64(1024),
					InfraMaxValue: uint64(102400),
				},
				edgeproto.InfraResource{
					Name:          "vCPUs",
					Value:         uint64(10),
					InfraMaxValue: uint64(109),
				},
				edgeproto.InfraResource{
					Name:          "Disk",
					Value:         uint64(20),
					InfraMaxValue: uint64(5000),
				},
				edgeproto.InfraResource{
					Name:          "GPUs",
					Value:         uint64(6),
					InfraMaxValue: uint64(20),
				},
				edgeproto.InfraResource{
					Name:          "External IPs",
					Value:         uint64(2),
					InfraMaxValue: uint64(10),
				},
			},
		},
		CompatibilityVersion: 1, // cloudcommon.GetCRMCompatibilityVersion()
		Properties: map[string]string{
			"supports-mt": "true", // cloudcommon.CloudletSupportsMT
		},
	},
	edgeproto.CloudletInfo{
		Key:         CloudletData[1].Key,
		State:       dme.CloudletState_CLOUDLET_STATE_READY,
		OsMaxRam:    65536,
		OsMaxVcores: 16,
		OsMaxVolGb:  500,
		Flavors: []*edgeproto.FlavorInfo{
			&edgeproto.FlavorInfo{
				Name:  "flavor.small1",
				Vcpus: uint64(2),
				Ram:   uint64(2048),
				Disk:  uint64(10),
			},
			&edgeproto.FlavorInfo{
				Name:  "flavor.small2",
				Vcpus: uint64(2),
				Ram:   uint64(1024),
				Disk:  uint64(20),
			},
			&edgeproto.FlavorInfo{
				Name:  "flavor.medium1",
				Vcpus: uint64(2),
				Ram:   uint64(4096),
				Disk:  uint64(40),
			},
		},
		ResourcesSnapshot: edgeproto.InfraResourcesSnapshot{
			Info: []edgeproto.InfraResource{
				edgeproto.InfraResource{
					Name:          "RAM",
					Value:         uint64(1024),
					InfraMaxValue: uint64(61440),
				},
				edgeproto.InfraResource{
					Name:          "vCPUs",
					Value:         uint64(10),
					InfraMaxValue: uint64(100),
				},
				edgeproto.InfraResource{
					Name:          "Disk",
					Value:         uint64(20),
					InfraMaxValue: uint64(5000),
				},
				edgeproto.InfraResource{
					Name:          "External IPs",
					Value:         uint64(2),
					InfraMaxValue: uint64(10),
				},
			},
		},
		CompatibilityVersion: 1, // cloudcommon.GetCRMCompatibilityVersion()
	},
	edgeproto.CloudletInfo{
		Key:         CloudletData[2].Key,
		State:       dme.CloudletState_CLOUDLET_STATE_READY,
		OsMaxRam:    65536,
		OsMaxVcores: 16,
		OsMaxVolGb:  500,
		Flavors: []*edgeproto.FlavorInfo{
			&edgeproto.FlavorInfo{
				Name:  "flavor.medium1",
				Vcpus: uint64(4),
				Ram:   uint64(8192),
				Disk:  uint64(80),
			},
			&edgeproto.FlavorInfo{
				Name:  "flavor.medium2",
				Vcpus: uint64(4),
				Ram:   uint64(4096),
				Disk:  uint64(40),
			},
			&edgeproto.FlavorInfo{
				Name:  "flavor.medium3",
				Vcpus: uint64(4),
				Ram:   uint64(2048),
				Disk:  uint64(20),
			},
		},
		ResourcesSnapshot: edgeproto.InfraResourcesSnapshot{
			Info: []edgeproto.InfraResource{
				edgeproto.InfraResource{
					Name:          "RAM",
					Value:         uint64(1024),
					InfraMaxValue: uint64(61440),
				},
				edgeproto.InfraResource{
					Name:          "vCPUs",
					Value:         uint64(10),
					InfraMaxValue: uint64(100),
				},
				edgeproto.InfraResource{
					Name:          "Disk",
					Value:         uint64(20),
					InfraMaxValue: uint64(5000),
				},
				edgeproto.InfraResource{
					Name:          "External IPs",
					Value:         uint64(2),
					InfraMaxValue: uint64(10),
				},
			},
		},
		CompatibilityVersion: 1, // cloudcommon.GetCRMCompatibilityVersion()
	},
	edgeproto.CloudletInfo{
		Key:         CloudletData[3].Key,
		State:       dme.CloudletState_CLOUDLET_STATE_READY,
		OsMaxRam:    65536,
		OsMaxVcores: 16,
		OsMaxVolGb:  500,
		Flavors: []*edgeproto.FlavorInfo{
			&edgeproto.FlavorInfo{
				Name:  "flavor.large",
				Vcpus: uint64(8),
				Ram:   uint64(101024),
				Disk:  uint64(100),
			},
			&edgeproto.FlavorInfo{
				Name:  "flavor.medium",
				Vcpus: uint64(4),
				Ram:   uint64(1),
				Disk:  uint64(1),
			},
		},
		ResourcesSnapshot: edgeproto.InfraResourcesSnapshot{
			Info: []edgeproto.InfraResource{
				edgeproto.InfraResource{
					Name:          "RAM",
					Value:         uint64(1024),
					InfraMaxValue: uint64(1024000),
				},
				edgeproto.InfraResource{
					Name:          "vCPUs",
					Value:         uint64(10),
					InfraMaxValue: uint64(100),
				},
				edgeproto.InfraResource{
					Name:          "Disk",
					Value:         uint64(20),
					InfraMaxValue: uint64(5000),
				},
				edgeproto.InfraResource{
					Name:          "External IPs",
					Value:         uint64(2),
					InfraMaxValue: uint64(10),
				},
			},
		},
		CompatibilityVersion: 1, // cloudcommon.GetCRMCompatibilityVersion()
	},
}

// To figure out what resources are used on each Cloudlet,
// see ClusterInstData to see what clusters are instantiated on what Cloudlet.
var CloudletRefsData = []edgeproto.CloudletRefs{
	// ClusterInstData[0,3,7,8]:
	edgeproto.CloudletRefs{
		Key: CloudletData[0].Key,
		ClusterInsts: []edgeproto.ClusterInstRefKey{
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[0].Key.ClusterKey,
				Organization: ClusterInstData[0].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[3].Key.ClusterKey,
				Organization: ClusterInstData[3].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[7].Key.ClusterKey,
				Organization: ClusterInstData[7].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[8].Key.ClusterKey,
				Organization: ClusterInstData[8].Key.Organization,
			},
		},
		UsedDynamicIps: 2,
	},
	// ClusterInstData[1,4]:
	edgeproto.CloudletRefs{
		Key: CloudletData[1].Key,
		ClusterInsts: []edgeproto.ClusterInstRefKey{
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[1].Key.ClusterKey,
				Organization: ClusterInstData[1].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[4].Key.ClusterKey,
				Organization: ClusterInstData[4].Key.Organization,
			},
		},
	},
	// ClusterInstData[2,5]:
	edgeproto.CloudletRefs{
		Key: CloudletData[2].Key,
		ClusterInsts: []edgeproto.ClusterInstRefKey{
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[2].Key.ClusterKey,
				Organization: ClusterInstData[2].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[5].Key.ClusterKey,
				Organization: ClusterInstData[5].Key.Organization,
			},
		},
		UsedDynamicIps: 1,
	},
	// ClusterInstData[6]:
	edgeproto.CloudletRefs{
		Key: CloudletData[3].Key,
		ClusterInsts: []edgeproto.ClusterInstRefKey{
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[6].Key.ClusterKey,
				Organization: ClusterInstData[6].Key.Organization,
			},
		},
	},
}

// These Refs are after creating both cluster insts and app insts.
// Some of the app insts trigger creating auto-clusterinsts,
// and ports are reserved with the creation of app insts.
var CloudletRefsWithAppInstsData = []edgeproto.CloudletRefs{
	// ClusterInstData[0,3,7,8]: (dedicated,dedicated,shared,shared)
	// AppInstData[0,1] -> ports[tcp:443;tcp:443]:
	// AppInstData[13,14,15,16] -> App[0,9,13,14] -> ports[tcp:443,tcp:10002,udp:10002;;tcp:889;tcp:444]
	edgeproto.CloudletRefs{
		Key: CloudletData[0].Key,
		ClusterInsts: []edgeproto.ClusterInstRefKey{
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[0].Key.ClusterKey,
				Organization: ClusterInstData[0].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[3].Key.ClusterKey,
				Organization: ClusterInstData[3].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[7].Key.ClusterKey,
				Organization: ClusterInstData[7].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[8].Key.ClusterKey,
				Organization: ClusterInstData[8].Key.Organization,
			},
		},
		RootLbPorts: map[int32]int32{443: 1, 10002: 3, 889: 1, 444: 1},
		VmAppInsts: []edgeproto.AppInstRefKey{
			edgeproto.AppInstRefKey{
				AppKey: AppInstData[11].Key.AppKey,
				ClusterInstKey: edgeproto.ClusterInstRefKey{
					ClusterKey:   AppInstData[11].Key.ClusterInstKey.ClusterKey,
					Organization: AppInstData[11].Key.ClusterInstKey.Organization,
				},
			},
		},
		UsedDynamicIps: 2,
	},
	// ClusterInstData[1,4], ClusterInstAutoData[0]: (shared,shared,shared)
	// AppInstData[2,3] -> ports[tcp:443;tcp:80,tcp:443,tcp:81,udp:10002]
	edgeproto.CloudletRefs{
		Key: CloudletData[1].Key,
		ClusterInsts: []edgeproto.ClusterInstRefKey{
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[1].Key.ClusterKey,
				Organization: ClusterInstData[1].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[4].Key.ClusterKey,
				Organization: ClusterInstData[4].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstAutoData[0].Key.ClusterKey,
				Organization: ClusterInstAutoData[0].Key.Organization,
			},
		},
		RootLbPorts:            map[int32]int32{80: 1, 81: 1, 443: 1, 10000: 1, 10002: 3},
		ReservedAutoClusterIds: 1,
	},
	// ClusterInstData[2,5], ClusterInstAutoData[1,2]: (shared,dedicated,shared,shared)
	// AppInstData[4,5,6] -> ports[tcp:443,udp:11111;udp:2024;tcp:80,udp:8001,tcp:65535]
	edgeproto.CloudletRefs{
		Key: CloudletData[2].Key,
		ClusterInsts: []edgeproto.ClusterInstRefKey{
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[2].Key.ClusterKey,
				Organization: ClusterInstData[2].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[5].Key.ClusterKey,
				Organization: ClusterInstData[5].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstAutoData[1].Key.ClusterKey,
				Organization: ClusterInstAutoData[1].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstAutoData[2].Key.ClusterKey,
				Organization: ClusterInstAutoData[2].Key.Organization,
			},
		},
		UsedDynamicIps:         1,
		RootLbPorts:            map[int32]int32{443: 1, 11111: 2, 2024: 2, 80: 1, 8001: 2, 65535: 1},
		ReservedAutoClusterIds: 3,
	},
	// ClusterInstData[6]: (no app insts on this clusterinst) (shared),
	// ClusterInstAutoData[3]: (shared)
	// AppInstData[12] -> ports[tcp:889]
	edgeproto.CloudletRefs{
		Key: CloudletData[3].Key,
		ClusterInsts: []edgeproto.ClusterInstRefKey{
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstData[6].Key.ClusterKey,
				Organization: ClusterInstData[6].Key.Organization,
			},
			edgeproto.ClusterInstRefKey{
				ClusterKey:   ClusterInstAutoData[3].Key.ClusterKey,
				Organization: ClusterInstAutoData[3].Key.Organization,
			},
		},
		RootLbPorts:            map[int32]int32{889: 1},
		ReservedAutoClusterIds: 1,
	},
}

var CloudletPoolData = []edgeproto.CloudletPool{
	edgeproto.CloudletPool{
		Key: edgeproto.CloudletPoolKey{
			Organization: OperatorData[1],
			Name:         "private",
		},
		Cloudlets: []string{
			CloudletData[2].Key.Name,
		},
	},
	edgeproto.CloudletPool{
		Key: edgeproto.CloudletPoolKey{
			Organization: OperatorData[2],
			Name:         "test-and-dev",
		},
		Cloudlets: []string{
			CloudletData[3].Key.Name,
		},
	},
	edgeproto.CloudletPool{
		Key: edgeproto.CloudletPoolKey{
			Organization: OperatorData[2],
			Name:         "enterprise",
		},
		Cloudlets: []string{
			CloudletData[3].Key.Name,
		},
	},
}

var Restblkeys = []edgeproto.ResTagTableKey{
	edgeproto.ResTagTableKey{
		Name:         "gpu",
		Organization: "AT&T Inc.",
	},
	edgeproto.ResTagTableKey{
		Name:         "nas",
		Organization: "AT&T Inc.",
	},
	edgeproto.ResTagTableKey{
		Name:         "nic",
		Organization: "AT&T Inc.",
	},
	edgeproto.ResTagTableKey{
		Name:         "gput4",
		Organization: "AT&T Inc.",
	},
}

var ResTagTableData = []edgeproto.ResTagTable{

	edgeproto.ResTagTable{
		Key:  Restblkeys[0],
		Tags: map[string]string{"vmware": "vgpu=1"},
	},

	edgeproto.ResTagTable{
		Key:  Restblkeys[1],
		Tags: map[string]string{"vcpu": "nvidia-72", "pci-passthru": "NP4:2"},
	},

	edgeproto.ResTagTable{
		Key:  Restblkeys[2],
		Tags: map[string]string{"vcpu": "nvidia-63", "pci-passthru": "T4:1"},
	},
	edgeproto.ResTagTable{
		Key:  Restblkeys[3],
		Tags: map[string]string{"pci": "t4:1"},
	},
}

var AlertData = []edgeproto.Alert{
	edgeproto.Alert{
		Labels: map[string]string{
			"alertname":   "AutoScaleUp",
			"cloudletorg": ClusterInstData[0].Key.CloudletKey.Organization,
			"cloudlet":    ClusterInstData[0].Key.CloudletKey.Name,
			"cluster":     ClusterInstData[0].Key.ClusterKey.Name,
			"clusterorg":  ClusterInstData[0].Key.Organization,
			"severity":    "none",
		},
		Annotations: map[string]string{
			"message": "Policy threshold to scale up cluster reached",
		},
		State: "firing",
		ActiveAt: dme.Timestamp{
			Seconds: 1257894000,
			Nanos:   2343569,
		},
		Value: 1,
	},
	edgeproto.Alert{
		Labels: map[string]string{
			"alertname":   "AutoScaleDown",
			"cloudletorg": ClusterInstData[0].Key.CloudletKey.Organization,
			"cloudlet":    ClusterInstData[0].Key.CloudletKey.Name,
			"cluster":     ClusterInstData[0].Key.ClusterKey.Name,
			"clusterorg":  ClusterInstData[0].Key.Organization,
			"severity":    "none",
		},
		Annotations: map[string]string{
			"message": "Policy threshold to scale down cluster reached",
		},
		State: "pending",
		ActiveAt: dme.Timestamp{
			Seconds: 1257894001,
			Nanos:   642398,
		},
		Value: 1,
	},
	edgeproto.Alert{
		Labels: map[string]string{
			"alertname":   "AutoScaleUp",
			"cloudletorg": ClusterInstData[1].Key.CloudletKey.Organization,
			"cloudlet":    ClusterInstData[1].Key.CloudletKey.Name,
			"cluster":     ClusterInstData[1].Key.ClusterKey.Name,
			"clusterorg":  ClusterInstData[1].Key.Organization,
			"severity":    "critical",
		},
		Annotations: map[string]string{
			"message": "Cluster offline",
		},
		State: "firing",
		ActiveAt: dme.Timestamp{
			Seconds: 1257894002,
			Nanos:   42398457,
		},
		Value: 1,
	},
	edgeproto.Alert{
		Labels: map[string]string{
			"alertname":   "AppInstDown",
			"app":         AppInstData[0].Key.AppKey.Name,
			"appver":      AppInstData[0].Key.AppKey.Version,
			"apporg":      AppInstData[0].Key.AppKey.Organization,
			"cloudletorg": ClusterInstData[7].Key.CloudletKey.Organization,
			"cloudlet":    ClusterInstData[7].Key.CloudletKey.Name,
			"cluster":     ClusterInstData[7].Key.ClusterKey.Name,
			"clusterorg":  ClusterInstData[7].Key.Organization,
			"status":      "1",
		},
		State: "firing",
		ActiveAt: dme.Timestamp{
			Seconds: 1257894002,
			Nanos:   42398457,
		},
	},
	edgeproto.Alert{
		Labels: map[string]string{
			"alertname":   "AppInstDown",
			"app":         AppInstData[0].Key.AppKey.Name,
			"appver":      AppInstData[0].Key.AppKey.Version,
			"apporg":      AppInstData[0].Key.AppKey.Organization,
			"cloudletorg": AppInstData[0].Key.ClusterInstKey.CloudletKey.Organization,
			"cloudlet":    AppInstData[0].Key.ClusterInstKey.CloudletKey.Name,
			"cluster":     AppInstData[0].Key.ClusterInstKey.ClusterKey.Name,
			"clusterorg":  AppInstData[0].Key.ClusterInstKey.Organization,
			"status":      "2",
		},
		State: "firing",
		ActiveAt: dme.Timestamp{
			Seconds: 1257894002,
			Nanos:   42398457,
		},
	},
}

var AutoScalePolicyData = []edgeproto.AutoScalePolicy{
	edgeproto.AutoScalePolicy{
		Key: edgeproto.PolicyKey{
			Name:         "auto-scale-policy",
			Organization: DevData[0],
		},
		MinNodes:           1,
		MaxNodes:           3,
		ScaleUpCpuThresh:   80,
		ScaleDownCpuThresh: 20,
		TriggerTimeSec:     60,
	},
	edgeproto.AutoScalePolicy{
		Key: edgeproto.PolicyKey{
			Name:         "auto-scale-policy",
			Organization: DevData[1],
		},
		MinNodes:           4,
		MaxNodes:           8,
		ScaleUpCpuThresh:   60,
		ScaleDownCpuThresh: 40,
		TriggerTimeSec:     30,
	},
	edgeproto.AutoScalePolicy{
		Key: edgeproto.PolicyKey{
			Name:         "auto-scale-policy",
			Organization: DevData[3],
		},
		MinNodes:           1,
		MaxNodes:           3,
		ScaleUpCpuThresh:   90,
		ScaleDownCpuThresh: 10,
		TriggerTimeSec:     60,
	},
}

var AutoProvPolicyData = []edgeproto.AutoProvPolicy{
	edgeproto.AutoProvPolicy{
		Key: edgeproto.PolicyKey{
			Name:         "auto-prov-policy",
			Organization: DevData[0],
		},
		DeployClientCount:   2,
		DeployIntervalCount: 2,
	},
	edgeproto.AutoProvPolicy{
		Key: edgeproto.PolicyKey{
			Name:         "auto-prov-policy",
			Organization: DevData[1],
		},
		DeployClientCount:   1,
		DeployIntervalCount: 1,
	},
	edgeproto.AutoProvPolicy{
		Key: edgeproto.PolicyKey{
			Name:         "auto-prov-policy",
			Organization: DevData[3],
		},
		DeployClientCount:   20,
		DeployIntervalCount: 4,
	},
	edgeproto.AutoProvPolicy{
		Key: edgeproto.PolicyKey{
			Name:         "auto-prov-policy2",
			Organization: DevData[0],
		},
		DeployClientCount:   10,
		DeployIntervalCount: 10,
	},
}

var TrustPolicyData = []edgeproto.TrustPolicy{
	edgeproto.TrustPolicy{
		Key: edgeproto.PolicyKey{
			Name:         "trust-policy0",
			Organization: CloudletData[2].Key.Organization,
		},
		OutboundSecurityRules: []edgeproto.SecurityRule{
			edgeproto.SecurityRule{
				Protocol:     "tcp",
				RemoteCidr:   "8.100.0.0/16",
				PortRangeMin: 443,
				PortRangeMax: 443,
			},
			edgeproto.SecurityRule{
				Protocol:     "udp",
				RemoteCidr:   "0.0.0.0/0",
				PortRangeMin: 53,
				PortRangeMax: 53,
			},
		},
	},
	edgeproto.TrustPolicy{
		Key: edgeproto.PolicyKey{
			Name:         "trust-policy1",
			Organization: CloudletData[2].Key.Organization,
		},
		OutboundSecurityRules: []edgeproto.SecurityRule{
			edgeproto.SecurityRule{
				Protocol:     "tcp",
				RemoteCidr:   "8.100.0.0/16",
				PortRangeMin: 443,
				PortRangeMax: 443,
			},
			edgeproto.SecurityRule{
				Protocol:     "udp",
				RemoteCidr:   "0.0.0.0/0",
				PortRangeMin: 53,
				PortRangeMax: 53,
			},
		},
	},
	edgeproto.TrustPolicy{
		Key: edgeproto.PolicyKey{
			Name:         "trust-policy2",
			Organization: CloudletData[2].Key.Organization,
		},
		OutboundSecurityRules: []edgeproto.SecurityRule{
			edgeproto.SecurityRule{
				Protocol:   "icmp",
				RemoteCidr: "0.0.0.0/0",
			},
			edgeproto.SecurityRule{
				Protocol:     "tcp",
				RemoteCidr:   "10.0.0.0/8",
				PortRangeMin: 1,
				PortRangeMax: 65535,
			},
		},
	},
}

var TrustPolicyErrorData = []edgeproto.TrustPolicy{
	// Failure case, max port > min port
	edgeproto.TrustPolicy{
		Key: edgeproto.PolicyKey{
			Name:         "trust-policy3",
			Organization: CloudletData[2].Key.Organization,
		},
		OutboundSecurityRules: []edgeproto.SecurityRule{
			edgeproto.SecurityRule{
				Protocol:     "tcp",
				RemoteCidr:   "10.1.0.0/16",
				PortRangeMin: 201,
				PortRangeMax: 110,
			},
		},
	},
	// Failure case, bad CIDR
	edgeproto.TrustPolicy{
		Key: edgeproto.PolicyKey{
			Name:         "trust-policy4",
			Organization: CloudletData[2].Key.Organization,
		},
		OutboundSecurityRules: []edgeproto.SecurityRule{
			edgeproto.SecurityRule{
				Protocol:     "tcp",
				RemoteCidr:   "10.0.0.0/50",
				PortRangeMin: 22,
				PortRangeMax: 22,
			},
		},
	},
	// Failure case, missing min port but max port present
	edgeproto.TrustPolicy{
		Key: edgeproto.PolicyKey{
			Name:         "trust-policy5",
			Organization: CloudletData[2].Key.Organization,
		},
		OutboundSecurityRules: []edgeproto.SecurityRule{
			edgeproto.SecurityRule{
				Protocol:     "tcp",
				RemoteCidr:   "47.186.0.0/16",
				PortRangeMax: 22,
			},
		},
	},
}

var AppInstClientKeyData = []edgeproto.AppInstClientKey{
	edgeproto.AppInstClientKey{
		AppInstKey: AppInstData[0].Key,
	},
	edgeproto.AppInstClientKey{
		AppInstKey: AppInstData[3].Key,
	},
}

var AppInstClientData = []edgeproto.AppInstClient{
	edgeproto.AppInstClient{
		ClientKey: AppInstClientKeyData[0],
		Location: dme.Loc{
			Latitude:  1.0,
			Longitude: 1.0,
		},
	},
	edgeproto.AppInstClient{
		ClientKey: AppInstClientKeyData[0],
		Location: dme.Loc{
			Latitude:  1.0,
			Longitude: 2.0,
		},
	},
	edgeproto.AppInstClient{
		ClientKey: AppInstClientKeyData[0],
		Location: dme.Loc{
			Latitude:  1.0,
			Longitude: 3.0,
		},
	},
	edgeproto.AppInstClient{
		ClientKey: AppInstClientKeyData[1],
		Location: dme.Loc{
			Latitude:  1.0,
			Longitude: 2.0,
		},
	},
}
var PlarformDeviceClientDataKeys = []edgeproto.DeviceKey{
	edgeproto.DeviceKey{
		UniqueIdType: "Samsung",
		UniqueId:     "1",
	},
	edgeproto.DeviceKey{
		UniqueIdType: "Samsung",
		UniqueId:     "2",
	},
	edgeproto.DeviceKey{
		UniqueIdType: "Mex",
		UniqueId:     "1",
	},
	edgeproto.DeviceKey{
		UniqueIdType: "GSAFKDF:Samsung:SamsungEnablementLayer",
		UniqueId:     "1",
	},
	edgeproto.DeviceKey{
		UniqueIdType: "SAMSUNG:CaseDeviceTest",
		UniqueId:     "1",
	},
}

var PlarformDeviceClientData = []edgeproto.Device{
	edgeproto.Device{
		Key: PlarformDeviceClientDataKeys[0],
		// 2009-11-10 23:00:00 +0000 UTC
		FirstSeen: GetTimestamp(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
	},
	edgeproto.Device{
		Key: PlarformDeviceClientDataKeys[1],
		// 2009-11-10 23:00:00 +0000 UTC
		FirstSeen: GetTimestamp(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
	},
	edgeproto.Device{
		Key: PlarformDeviceClientDataKeys[2],
		// 2009-12-10 23:00:00 +0000 UTC
		FirstSeen: GetTimestamp(time.Date(2009, time.December, 10, 23, 0, 0, 0, time.UTC)),
	},
	edgeproto.Device{
		Key: PlarformDeviceClientDataKeys[3],
		// 2009-10-10 23:30:00 +0000 UTC
		FirstSeen: GetTimestamp(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
	},
	edgeproto.Device{
		Key: PlarformDeviceClientDataKeys[4],
		// 2009-12-10 23:30:00 +0000 UTC
		FirstSeen: GetTimestamp(time.Date(2009, time.December, 10, 23, 0, 0, 0, time.UTC)),
	},
}

var VMPoolData = []edgeproto.VMPool{
	edgeproto.VMPool{
		Key: edgeproto.VMPoolKey{
			Organization: OperatorData[0],
			Name:         "San Jose Site",
		},
		Vms: []edgeproto.VM{
			edgeproto.VM{
				Name: "vm1",
				NetInfo: edgeproto.VMNetInfo{
					ExternalIp: "192.168.1.101",
					InternalIp: "192.168.100.101",
				},
				Flavor: &edgeproto.FlavorInfo{
					Name:  "vm1-flavor",
					Vcpus: uint64(2),
					Ram:   uint64(2048),
					Disk:  uint64(20),
				},
			},
			edgeproto.VM{
				Name: "vm2",
				NetInfo: edgeproto.VMNetInfo{
					ExternalIp: "192.168.1.102",
					InternalIp: "192.168.100.102",
				},
				Flavor: &edgeproto.FlavorInfo{
					Name:  "vm2-flavor",
					Vcpus: uint64(2),
					Ram:   uint64(2048),
					Disk:  uint64(20),
				},
			},
			edgeproto.VM{
				Name: "vm3",
				NetInfo: edgeproto.VMNetInfo{
					InternalIp: "192.168.100.103",
				},
				Flavor: &edgeproto.FlavorInfo{
					Name:  "vm3-flavor",
					Vcpus: uint64(2),
					Ram:   uint64(2048),
					Disk:  uint64(20),
				},
			},
			edgeproto.VM{
				Name: "vm4",
				NetInfo: edgeproto.VMNetInfo{
					InternalIp: "192.168.100.104",
				},
				Flavor: &edgeproto.FlavorInfo{
					Name:  "vm4-flavor",
					Vcpus: uint64(2),
					Ram:   uint64(2048),
					Disk:  uint64(20),
				},
			},
			edgeproto.VM{
				Name: "vm5",
				NetInfo: edgeproto.VMNetInfo{
					InternalIp: "192.168.100.105",
				},
				Flavor: &edgeproto.FlavorInfo{
					Name:  "vm5-flavor",
					Vcpus: uint64(3),
					Ram:   uint64(4096),
					Disk:  uint64(50),
				},
			},
		},
	},
	edgeproto.VMPool{
		Key: edgeproto.VMPoolKey{
			Organization: OperatorData[0],
			Name:         "New York Site",
		},
		Vms: []edgeproto.VM{
			edgeproto.VM{
				Name: "vm1",
				NetInfo: edgeproto.VMNetInfo{
					ExternalIp: "192.168.1.101",
					InternalIp: "192.168.100.101",
				},
				Flavor: &edgeproto.FlavorInfo{
					Name:  "vm1-flavor",
					Vcpus: uint64(2),
					Ram:   uint64(2048),
					Disk:  uint64(20),
				},
			},
			edgeproto.VM{
				Name: "vm2",
				NetInfo: edgeproto.VMNetInfo{
					ExternalIp: "192.168.1.102",
					InternalIp: "192.168.100.102",
				},
				Flavor: &edgeproto.FlavorInfo{
					Name:  "vm2-flavor",
					Vcpus: uint64(2),
					Ram:   uint64(2048),
					Disk:  uint64(20),
				},
			},
			edgeproto.VM{
				Name: "vm3",
				NetInfo: edgeproto.VMNetInfo{
					InternalIp: "192.168.100.103",
				},
				Flavor: &edgeproto.FlavorInfo{
					Name:  "vm3-flavor",
					Vcpus: uint64(2),
					Ram:   uint64(2048),
					Disk:  uint64(20),
				},
			},
		},
	},
	edgeproto.VMPool{
		Key: edgeproto.VMPoolKey{
			Organization: OperatorData[1],
			Name:         "San Francisco Site",
		},
		Vms: []edgeproto.VM{
			edgeproto.VM{
				Name: "vm1",
				NetInfo: edgeproto.VMNetInfo{
					InternalIp: "192.168.100.101",
				},
				Flavor: &edgeproto.FlavorInfo{
					Name:  "vm1-flavor",
					Vcpus: uint64(2),
					Ram:   uint64(2048),
					Disk:  uint64(20),
				},
			},
			edgeproto.VM{
				Name: "vm2",
				NetInfo: edgeproto.VMNetInfo{
					InternalIp: "192.168.100.102",
				},
				Flavor: &edgeproto.FlavorInfo{
					Name:  "vm2-flavor",
					Vcpus: uint64(2),
					Ram:   uint64(2048),
					Disk:  uint64(20),
				},
			},
		},
	},
}

var GPUDriverData = []edgeproto.GPUDriver{
	edgeproto.GPUDriver{
		Key: edgeproto.GPUDriverKey{
			Organization: OperatorData[0],
			Name:         "nvidia-450",
		},
	},
	edgeproto.GPUDriver{
		Key: edgeproto.GPUDriverKey{
			Name: "nvidia-490",
		},
	},
	edgeproto.GPUDriver{
		Key: edgeproto.GPUDriverKey{
			Organization: OperatorData[1],
			Name:         "nvidia-999",
		},
	},
	edgeproto.GPUDriver{
		Key: edgeproto.GPUDriverKey{
			Organization: OperatorData[0],
			Name:         "nvidia-vgpu",
		},
	},
}

var UserAlertData = []edgeproto.UserAlert{
	edgeproto.UserAlert{ // Warning alert with no labels/annotations
		Key: edgeproto.UserAlertKey{
			Name:         "testAlert1",
			Organization: DevData[0],
		},
		CpuLimit:    80,
		MemLimit:    123456,
		DiskLimit:   123456,
		Severity:    "warning",
		TriggerTime: edgeproto.Duration(30 * time.Second),
	},
	edgeproto.UserAlert{ // Warning alert with Active Connections
		Key: edgeproto.UserAlertKey{
			Name:         "testAlert2",
			Organization: DevData[0],
		},
		ActiveConnLimit: 10,
		Severity:        "info",
		TriggerTime:     edgeproto.Duration(5 * time.Minute),
	},
	edgeproto.UserAlert{ // Error alert with extra labels
		Key: edgeproto.UserAlertKey{
			Name:         "testAlert3",
			Organization: DevData[1],
		},
		CpuLimit:    100,
		Severity:    "error",
		TriggerTime: edgeproto.Duration(30 * time.Second),
		Labels: map[string]string{
			"testLabel1": "testValue1",
			"testLabel2": "testValue2",
		},
		Annotations: map[string]string{
			"testAnnotation1": "description1",
			"testAnnotation2": "description2",
		},
	},
}

func GetTimestamp(t time.Time) *types.Timestamp {
	ts, _ := types.TimestampProto(t)
	return ts
}

func IsAutoClusterAutoDeleteApp(key *edgeproto.AppInstKey) bool {
	if !strings.HasPrefix(key.ClusterInstKey.ClusterKey.Name, "autocluster") && !strings.HasPrefix(key.ClusterInstKey.ClusterKey.Name, "reservable") {
		return false
	}
	for _, app := range AppData {
		if app.Key.Matches(&key.AppKey) {
			return app.DelOpt == edgeproto.DeleteType_AUTO_DELETE
		}
	}
	panic(fmt.Sprintf("App definition not found for %v", key))
}

// Get the AppInst data after it has been created by the Controller.
// This is for tests that are using data as if it has already been
// created and processed by the Controller, given that the controller
// may modify certain fields during create.
func CreatedAppInstData() []edgeproto.AppInst {
	insts := []edgeproto.AppInst{}
	for ii, appInst := range AppInstData {
		switch ii {
		case 3:
			// grab expected autocluster real name
			appInst.RealClusterName = ClusterInstAutoData[0].Key.ClusterKey.Name
		case 4:
			appInst.RealClusterName = ClusterInstAutoData[1].Key.ClusterKey.Name
		case 6:
			appInst.RealClusterName = ClusterInstAutoData[2].Key.ClusterKey.Name
		case 11:
			appInst.Key.ClusterInstKey.Organization = appInst.Key.AppKey.Organization
			appInst.Key.ClusterInstKey.ClusterKey.Name = "DefaultCluster"
		}
		insts = append(insts, appInst)
	}
	return insts
}
