package testutil

import (
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
}

var DevData = []edgeproto.Developer{
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "NianticInc",
		},
	},
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "Ever.ai",
		},
	},
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "1000realities",
		},
	},
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "SierrawareLLC",
		},
	},
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
}

var AppData = []edgeproto.App{
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pokemon Go!",
			Version:      "1.0.0",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:   "http:443,tcp:10002,udp:10002",
		DefaultFlavor: FlavorData[0].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pokemon Go!",
			Version:      "1.0.1",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:   "tcp:80,http:443",
		DefaultFlavor: FlavorData[0].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Harry Potter Go! Go!",
			Version:      "0.0.1",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:   "tcp:443,udp:11111",
		DefaultFlavor: FlavorData[1].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[1].Key,
			Name:         "AI",
			Version:      "1.2.0",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_QCOW,
		ImagePath:     "http://somerepo/image/path/ai/1.2.0#md5:7e9cfcb763e83573a4b9d9315f56cc5f",
		AccessPorts:   "tcp:8080",
		DefaultFlavor: FlavorData[1].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[2].Key,
			Name:         "my reality",
			Version:      "0.0.1",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_QCOW,
		ImagePath:     "http://somerepo/image/path/myreality/0.0.1#md5:7e9cfcb763e83573a4b9d9315f56cc5f",
		AccessPorts:   "udp:1024",
		DefaultFlavor: FlavorData[2].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[3].Key,
			Name:         "helmApp",
			Version:      "0.0.1",
		},
		Deployment:    "helm",
		AccessPorts:   "udp:2024",
		DefaultFlavor: FlavorData[2].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Neon",
			Version:      "0.0.2",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:   "tcp:80,udp:8001,tcp:065535",
		DefaultFlavor: FlavorData[1].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "NoPorts",
			Version:      "1.0.0",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		DefaultFlavor: FlavorData[0].Key,
	},

	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "PortRangeApp",
			Version:      "1.0.0",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:   "tcp:80,http:443,udp:10002,tcp:5000-5002", // new port range notation
		DefaultFlavor: FlavorData[0].Key,
	},

	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "AutoDeleteApp",
			Version:      "1.0.0",
		},
		ImageType:     edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		DefaultFlavor: FlavorData[0].Key,
		DelOpt:        edgeproto.DeleteType_AUTO_DELETE,
	},
}
var OperatorData = []edgeproto.Operator{
	edgeproto.Operator{
		Key: edgeproto.OperatorKey{
			Name: "AT&T Inc.",
		},
	},
	edgeproto.Operator{
		Key: edgeproto.OperatorKey{
			Name: "T-Mobile",
		},
	},
	edgeproto.Operator{
		Key: edgeproto.OperatorKey{
			Name: "Verizon",
		},
	},
	edgeproto.Operator{
		Key: edgeproto.OperatorKey{
			Name: "Deutsche Telekom",
		},
	},
}

var CloudletData = []edgeproto.Cloudlet{
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[0].Key,
			Name:        "San Jose Site",
		},
		IpSupport:     edgeproto.IpSupport_IP_SUPPORT_DYNAMIC,
		NumDynamicIps: 100,
		Location: dme.Loc{
			Latitude:  37.338207,
			Longitude: -121.886330,
		},
		PlatformType:  edgeproto.PlatformType_PLATFORM_TYPE_FAKE,
		Flavor:        FlavorData[0].Key,
		NotifySrvAddr: "127.0.0.1:51001",
		CrmOverride:   edgeproto.CRMOverride_IGNORE_CRM,
		PhysicalName:  "SanJoseSite",
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[0].Key,
			Name:        "New York Site",
		},
		IpSupport:     edgeproto.IpSupport_IP_SUPPORT_DYNAMIC,
		NumDynamicIps: 100,
		Location: dme.Loc{
			Latitude:  40.712776,
			Longitude: -74.005974,
		},
		PlatformType:  edgeproto.PlatformType_PLATFORM_TYPE_FAKE,
		Flavor:        FlavorData[0].Key,
		NotifySrvAddr: "127.0.0.1:51002",
		CrmOverride:   edgeproto.CRMOverride_IGNORE_CRM,
		PhysicalName:  "NewYorkSite",
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[1].Key,
			Name:        "San Francisco Site",
		},
		IpSupport:     edgeproto.IpSupport_IP_SUPPORT_DYNAMIC,
		NumDynamicIps: 100,
		Location: dme.Loc{
			Latitude:  37.774929,
			Longitude: -122.419418,
		},
		Flavor:        FlavorData[0].Key,
		PlatformType:  edgeproto.PlatformType_PLATFORM_TYPE_FAKE,
		NotifySrvAddr: "127.0.0.1:51003",
		CrmOverride:   edgeproto.CRMOverride_IGNORE_CRM,
		PhysicalName:  "SanFranciscoSite",
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[2].Key,
			Name:        "Hawaii Site",
		},
		IpSupport:     edgeproto.IpSupport_IP_SUPPORT_DYNAMIC,
		NumDynamicIps: 10,
		Location: dme.Loc{
			Latitude:  21.306944,
			Longitude: -157.858337,
		},
		Flavor:        FlavorData[0].Key,
		PlatformType:  edgeproto.PlatformType_PLATFORM_TYPE_FAKE,
		NotifySrvAddr: "127.0.0.1:51004",
		CrmOverride:   edgeproto.CRMOverride_IGNORE_CRM,
		PhysicalName:  "HawaiiSite",
	},
}
var ClusterInstData = []edgeproto.ClusterInst{
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterKeys[0],
			CloudletKey: CloudletData[0].Key,
			Developer:   DevData[0].Key.Name,
		},
		Flavor:     FlavorData[0].Key,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_DEDICATED,
		NumMasters: 1,
		NumNodes:   2,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterKeys[0],
			CloudletKey: CloudletData[1].Key,
			Developer:   DevData[0].Key.Name,
		},
		Flavor:     FlavorData[0].Key,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_SHARED,
		NumMasters: 1,
		NumNodes:   2,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterKeys[0],
			CloudletKey: CloudletData[2].Key,
			Developer:   DevData[3].Key.Name,
		},
		Flavor:     FlavorData[0].Key,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_DEDICATED_OR_SHARED,
		NumMasters: 1,
		NumNodes:   2,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterKeys[1],
			CloudletKey: CloudletData[0].Key,
			Developer:   DevData[0].Key.Name,
		},
		Flavor:     FlavorData[1].Key,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_DEDICATED,
		NumMasters: 1,
		NumNodes:   3,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterKeys[1],
			CloudletKey: CloudletData[1].Key,
			Developer:   DevData[3].Key.Name,
		},
		Flavor:     FlavorData[1].Key,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_SHARED,
		NumMasters: 1,
		NumNodes:   3,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterKeys[2],
			CloudletKey: CloudletData[2].Key,
			Developer:   DevData[3].Key.Name,
		},
		Flavor:     FlavorData[2].Key,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_DEDICATED,
		NumMasters: 1,
		NumNodes:   4,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterKeys[3],
			CloudletKey: CloudletData[3].Key,
			Developer:   DevData[3].Key.Name,
		},
		Flavor:     FlavorData[2].Key,
		NumMasters: 1,
		NumNodes:   3,
	},
}

// These are the cluster insts that will be created automatically
// from appinsts that have not specified a cluster.
var ClusterInstAutoData = []edgeproto.ClusterInst{
	// from AppInstData[3] -> AppData[1]
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey: edgeproto.ClusterKey{
				Name: util.K8SSanitize("AutoCluster" + AppData[1].Key.Name),
			},
			CloudletKey: CloudletData[1].Key,
			Developer:   AppData[1].Key.DeveloperKey.Name,
		},
		Flavor:     FlavorData[0].Key,
		NumMasters: 1,
		NumNodes:   1,
		State:      edgeproto.TrackedState_READY,
		Auto:       true,
	},
	// from AppInstData[4] -> AppData[2]
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey: edgeproto.ClusterKey{
				Name: util.K8SSanitize("AutoCluster" + AppData[2].Key.Name),
			},
			CloudletKey: CloudletData[2].Key,
			Developer:   AppData[2].Key.DeveloperKey.Name,
		},
		Flavor:     FlavorData[1].Key,
		NumMasters: 1,
		NumNodes:   1,
		State:      edgeproto.TrackedState_READY,
		Auto:       true,
	},
	// from AppInstData[6] -> AppData[6]
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey: edgeproto.ClusterKey{
				Name: util.K8SSanitize("AutoCluster" + AppData[6].Key.Name),
			},
			CloudletKey: CloudletData[2].Key,
			Developer:   AppData[6].Key.DeveloperKey.Name,
		},
		Flavor:     FlavorData[1].Key,
		NumMasters: 1,
		NumNodes:   1,
		State:      edgeproto.TrackedState_READY,
		Auto:       true,
		IpAccess:   edgeproto.IpAccess_IP_ACCESS_DEDICATED,
	},
}
var AppInstData = []edgeproto.AppInst{
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[0].Key,
			ClusterInstKey: ClusterInstData[0].Key,
		},
		CloudletLoc: CloudletData[0].Location,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[0].Key,
			ClusterInstKey: ClusterInstData[3].Key,
		},
		CloudletLoc: CloudletData[0].Location,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[0].Key,
			ClusterInstKey: ClusterInstData[1].Key,
		},
		CloudletLoc: CloudletData[1].Location,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey: AppData[1].Key,
			// ClusterInst is ClusterInstAutoData[0]
			ClusterInstKey: ClusterInstAutoData[0].Key,
		},
		CloudletLoc: CloudletData[1].Location,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey: AppData[2].Key,
			// ClusterInst is ClusterInstAutoData[1]
			ClusterInstKey: ClusterInstAutoData[1].Key,
		},
		CloudletLoc: CloudletData[2].Location,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[5].Key,
			ClusterInstKey: ClusterInstData[2].Key,
		},
		CloudletLoc: CloudletData[2].Location,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey: AppData[6].Key,
			// ClusterInst is ClusterInstAutoData[2]
			ClusterInstKey: ClusterInstAutoData[2].Key,
		},
		CloudletLoc:         CloudletData[2].Location,
		AutoClusterIpAccess: edgeproto.IpAccess_IP_ACCESS_DEDICATED,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[6].Key,
			ClusterInstKey: ClusterInstData[0].Key,
		},
		CloudletLoc: CloudletData[0].Location,
	},

	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         AppData[7].Key,
			ClusterInstKey: ClusterInstData[0].Key,
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
var CloudletInfoData = []edgeproto.CloudletInfo{
	edgeproto.CloudletInfo{
		Key:         CloudletData[0].Key,
		State:       edgeproto.CloudletState_CLOUDLET_STATE_READY,
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
		},
	},
	edgeproto.CloudletInfo{
		Key:         CloudletData[1].Key,
		State:       edgeproto.CloudletState_CLOUDLET_STATE_READY,
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
		},
	},
	edgeproto.CloudletInfo{
		Key:         CloudletData[2].Key,
		State:       edgeproto.CloudletState_CLOUDLET_STATE_READY,
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
	},
	edgeproto.CloudletInfo{
		Key:         CloudletData[3].Key,
		State:       edgeproto.CloudletState_CLOUDLET_STATE_READY,
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
	},
}

// To figure out what resources are used on each Cloudlet,
// see ClusterInstData to see what clusters are instantiated on what Cloudlet.
var CloudletRefsData = []edgeproto.CloudletRefs{
	// ClusterInstData[0,3]:
	edgeproto.CloudletRefs{
		Key: CloudletData[0].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterKeys[0],
			ClusterKeys[1],
		},
		UsedRam:        GetCloudletUsedRam(0, 3),
		UsedVcores:     GetCloudletUsedVcores(0, 3),
		UsedDisk:       GetCloudletUsedDisk(0, 3),
		UsedDynamicIps: 2,
	},
	// ClusterInstData[1,4]:
	edgeproto.CloudletRefs{
		Key: CloudletData[1].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterKeys[0],
			ClusterKeys[1],
		},
		UsedRam:    GetCloudletUsedRam(1, 4),
		UsedVcores: GetCloudletUsedVcores(1, 4),
		UsedDisk:   GetCloudletUsedDisk(1, 4),
	},
	// ClusterInstData[2,5]:
	edgeproto.CloudletRefs{
		Key: CloudletData[2].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterKeys[0],
			ClusterKeys[2],
		},
		UsedRam:        GetCloudletUsedRam(2, 5),
		UsedVcores:     GetCloudletUsedVcores(2, 5),
		UsedDisk:       GetCloudletUsedDisk(2, 5),
		UsedDynamicIps: 1,
	},
	// ClusterInstData[6]:
	edgeproto.CloudletRefs{
		Key: CloudletData[3].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterKeys[3],
		},
		UsedRam:    GetCloudletUsedRam(6),
		UsedVcores: GetCloudletUsedVcores(6),
		UsedDisk:   GetCloudletUsedDisk(6),
	},
}

// These Refs are after creating both cluster insts and app insts.
// Some of the app insts trigger creating auto-clusterinsts,
// and ports are reserved with the creation of app insts.
var CloudletRefsWithAppInstsData = []edgeproto.CloudletRefs{
	// ClusterInstData[0,3]: (dedicated,dedicated)
	// AppInstData[0,1] -> ports[http:443;http:443]:
	edgeproto.CloudletRefs{
		Key: CloudletData[0].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterKeys[0],
			ClusterKeys[1],
		},
		UsedRam:        GetCloudletUsedRam(0, 3),
		UsedVcores:     GetCloudletUsedVcores(0, 3),
		UsedDisk:       GetCloudletUsedDisk(0, 3),
		UsedDynamicIps: 2,
	},
	// ClusterInstData[1,4], ClusterInstAutoData[0]: (shared,shared,shared)
	// AppInstData[2,3] -> ports[http:443;tcp:80,http:443]
	edgeproto.CloudletRefs{
		Key: CloudletData[1].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterKeys[0],
			ClusterKeys[1],
			ClusterInstAutoData[0].Key.ClusterKey,
		},
		UsedRam:     GetCloudletUsedRam(1, 4, -1, 0),
		UsedVcores:  GetCloudletUsedVcores(1, 4, -1, 0),
		UsedDisk:    GetCloudletUsedDisk(1, 4, -1, 0),
		RootLbPorts: map[int32]int32{80: 1, 10002: 3},
	},
	// ClusterInstData[2,5], ClusterInstAutoData[1,2]: (shared,dedicated,shared,dedicated)
	// AppInstData[4,5] -> ports[tcp:443,udp:11111;udp:2024,tcp:80,udp:8001]
	edgeproto.CloudletRefs{
		Key: CloudletData[2].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterKeys[0],
			ClusterKeys[2],
			ClusterInstAutoData[1].Key.ClusterKey,
			ClusterInstAutoData[2].Key.ClusterKey,
		},
		UsedRam:        GetCloudletUsedRam(2, 5, -1, 1, 2),
		UsedVcores:     GetCloudletUsedVcores(2, 5, -1, 1, 2),
		UsedDisk:       GetCloudletUsedDisk(2, 5, -1, 1, 2),
		UsedDynamicIps: 2,
		RootLbPorts:    map[int32]int32{10000: 1, 11111: 2, 2024: 2},
	},
	// ClusterInstData[6]: (no app insts on this clusterinst) (shared)
	edgeproto.CloudletRefs{
		Key: CloudletData[3].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterKeys[3],
		},
		UsedRam:    GetCloudletUsedRam(6),
		UsedVcores: GetCloudletUsedVcores(6),
		UsedDisk:   GetCloudletUsedDisk(6),
	},
}

var CloudletPoolData = []edgeproto.CloudletPool{
	edgeproto.CloudletPool{
		Key: edgeproto.CloudletPoolKey{
			Name: "private",
		},
	},
	edgeproto.CloudletPool{
		Key: edgeproto.CloudletPoolKey{
			Name: "test-and-dev",
		},
	},
	edgeproto.CloudletPool{
		Key: edgeproto.CloudletPoolKey{
			Name: "enterprise",
		},
	},
}

var CloudletPoolMemberData = []edgeproto.CloudletPoolMember{
	edgeproto.CloudletPoolMember{
		PoolKey:     CloudletPoolData[0].Key, // private
		CloudletKey: CloudletData[3].Key,
	},
	edgeproto.CloudletPoolMember{
		PoolKey:     CloudletPoolData[1].Key, // test-and-dev
		CloudletKey: CloudletData[2].Key,
	},
	edgeproto.CloudletPoolMember{
		PoolKey:     CloudletPoolData[2].Key, // enterprise
		CloudletKey: CloudletData[3].Key,
	},
}

func GetCloudletUsedRam(indices ...int) uint64 {
	var ram uint64
	data := ClusterInstData
	for _, idx := range indices {
		if idx == -1 {
			data = ClusterInstAutoData
			continue
		}
		clinst := data[idx]
		clflavor := data[idx].Flavor
		flavor, _ := FindFlavorData(&clflavor, FlavorData)
		ram += flavor.Ram * uint64(clinst.NumNodes+clinst.NumMasters)
	}
	return ram
}

func GetCloudletUsedVcores(indices ...int) uint64 {
	var vcores uint64
	data := ClusterInstData
	for _, idx := range indices {
		if idx == -1 {
			data = ClusterInstAutoData
			continue
		}
		clinst := data[idx]
		clflavor := data[idx].Flavor
		flavor, _ := FindFlavorData(&clflavor, FlavorData)
		vcores += flavor.Vcpus * uint64(clinst.NumNodes+clinst.NumMasters)
	}
	return vcores
}

func GetCloudletUsedDisk(indices ...int) uint64 {
	var disk uint64
	data := ClusterInstData
	for _, idx := range indices {
		if idx == -1 {
			data = ClusterInstAutoData
			continue
		}
		clinst := data[idx]
		clflavor := data[idx].Flavor
		flavor, _ := FindFlavorData(&clflavor, FlavorData)
		disk += flavor.Disk * uint64(clinst.NumNodes+clinst.NumMasters)
	}
	return disk
}
