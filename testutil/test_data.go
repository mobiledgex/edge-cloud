package testutil

import (
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
}
var ClusterFlavorData = []edgeproto.ClusterFlavor{
	edgeproto.ClusterFlavor{
		Key: edgeproto.ClusterFlavorKey{
			Name: "c1.tiny",
		},
		NodeFlavor:   FlavorData[0].Key,
		MasterFlavor: FlavorData[0].Key,
		NumNodes:     2,
		MaxNodes:     2,
		NumMasters:   1,
	},
	edgeproto.ClusterFlavor{
		Key: edgeproto.ClusterFlavorKey{
			Name: "c1.small",
		},
		NodeFlavor:   FlavorData[1].Key,
		MasterFlavor: FlavorData[1].Key,
		NumNodes:     3,
		MaxNodes:     3,
		NumMasters:   1,
	},
	edgeproto.ClusterFlavor{
		Key: edgeproto.ClusterFlavorKey{
			Name: "c1.medium",
		},
		NodeFlavor:   FlavorData[2].Key,
		MasterFlavor: FlavorData[2].Key,
		NumNodes:     3,
		MaxNodes:     4,
		NumMasters:   1,
	},
}
var DevData = []edgeproto.Developer{
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "Niantic, Inc.",
		},
		Address: "1230 Midas Way #200, Sunnyvale, CA 94085",
		Email:   "edge.niantic.com",
	},
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "Ever.ai",
		},
		Address: "1 Letterman Drive Building C, San Francisco, CA 94129",
		Email:   "edge.everai.com",
	},
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "1000 realities",
		},
		Address: "Kamienna 43, 31-403 Krakow, Poland",
		Email:   "edge.1000realities.com",
	},
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "Sierraware LLC",
		},
		Address: "1250 Oakmead Parkway Suite 210, Sunnyvalue, CA 94085",
		Email:   "support@sierraware.com",
	},
}
var ClusterData = []edgeproto.Cluster{
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: "Pokemons",
		},
		DefaultFlavor: ClusterFlavorData[0].Key,
	},
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: "Ever.Ai",
		},
		DefaultFlavor: ClusterFlavorData[1].Key,
	},
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: "1000realities",
		},
		DefaultFlavor: ClusterFlavorData[2].Key,
	},
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: "Big-Pokemons",
		},
		DefaultFlavor: ClusterFlavorData[2].Key,
	},
}

// these are the auto clusters created by apps that don't specify a cluster
var ClusterAutoData = []edgeproto.Cluster{
	// AppData[1]:
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: util.K8SSanitize("AutoCluster" + AppData[1].Key.Name),
		},
		DefaultFlavor: ClusterFlavorData[0].Key,
		Auto:          true,
	},
	// AppData[2]:
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: util.K8SSanitize("AutoCluster" + AppData[2].Key.Name),
		},
		DefaultFlavor: ClusterFlavorData[1].Key,
		Auto:          true,
	},
	// AppData[3]:
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: util.K8SSanitize("AutoCluster" + AppData[3].Key.Name),
		},
		DefaultFlavor: ClusterFlavorData[1].Key,
		Auto:          true,
	},
}
var AppData = []edgeproto.App{
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pokemon Go!",
			Version:      "1.0.0",
		},
		ImageType:     edgeproto.ImageType_ImageTypeDocker,
		IpAccess:      edgeproto.IpAccess_IpAccessDedicated,
		AccessPorts:   "http:443",
		DefaultFlavor: FlavorData[0].Key,
		Cluster:       ClusterData[0].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pokemon Go!",
			Version:      "1.0.1",
		},
		ImageType:     edgeproto.ImageType_ImageTypeDocker,
		IpAccess:      edgeproto.IpAccess_IpAccessShared,
		AccessPorts:   "tcp:80,http:443",
		DefaultFlavor: FlavorData[0].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Harry Potter Go! Go!",
			Version:      "0.0.1",
		},
		ImageType:     edgeproto.ImageType_ImageTypeDocker,
		IpAccess:      edgeproto.IpAccess_IpAccessDedicatedOrShared,
		AccessPorts:   "tcp:443,udp:11111",
		DefaultFlavor: FlavorData[1].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[1].Key,
			Name:         "AI",
			Version:      "1.2.0",
		},
		ImageType:     edgeproto.ImageType_ImageTypeQCOW,
		IpAccess:      edgeproto.IpAccess_IpAccessDedicated,
		AccessPorts:   "http:8080",
		DefaultFlavor: FlavorData[1].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[2].Key,
			Name:         "my reality",
			Version:      "0.0.1",
		},
		ImageType:     edgeproto.ImageType_ImageTypeQCOW,
		IpAccess:      edgeproto.IpAccess_IpAccessShared,
		AccessPorts:   "udp:1024",
		DefaultFlavor: FlavorData[2].Key,
		Cluster:       ClusterData[2].Key,
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
		AccessUri:     "10.100.0.1",
		IpSupport:     edgeproto.IpSupport_IpSupportDynamic,
		NumDynamicIps: 100,
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[0].Key,
			Name:        "New York Site",
		},
		AccessUri:     "ff.f8::1",
		IpSupport:     edgeproto.IpSupport_IpSupportDynamic,
		NumDynamicIps: 100,
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[1].Key,
			Name:        "San Francisco Site",
		},
		AccessUri:     "172.24.0.1",
		IpSupport:     edgeproto.IpSupport_IpSupportDynamic,
		NumDynamicIps: 100,
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[2].Key,
			Name:        "Hawaii Site",
		},
		AccessUri:     "hi76.verizon.com",
		IpSupport:     edgeproto.IpSupport_IpSupportDynamic,
		NumDynamicIps: 10,
	},
}
var ClusterInstData = []edgeproto.ClusterInst{
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[0].Key,
			CloudletKey: CloudletData[0].Key,
		},
		Flavor: ClusterData[0].DefaultFlavor,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[0].Key,
			CloudletKey: CloudletData[1].Key,
		},
		Flavor: ClusterData[0].DefaultFlavor,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[0].Key,
			CloudletKey: CloudletData[2].Key,
		},
		Flavor: ClusterData[0].DefaultFlavor,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[1].Key,
			CloudletKey: CloudletData[0].Key,
		},
		Flavor: ClusterData[1].DefaultFlavor,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[1].Key,
			CloudletKey: CloudletData[1].Key,
		},
		Flavor: ClusterData[1].DefaultFlavor,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[2].Key,
			CloudletKey: CloudletData[2].Key,
		},
		Flavor: ClusterData[2].DefaultFlavor,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[3].Key,
			CloudletKey: CloudletData[3].Key,
		},
		Flavor: ClusterData[3].DefaultFlavor,
	},
}

// These are the cluster insts that will be created automatically
// from appinsts that have not specified a cluster.
var ClusterInstAutoData = []edgeproto.ClusterInst{
	// from AppInstData[1]:
	// There is no auto-cluster created for AppInstData[1],
	// because the AppData[0] is corresponds to does have a \
	// cluster specified, and an instance for that cluster is
	// created by ClusterData[0] -> ClusterInstData[0].
	// XXX I'm not sure if this makes sense, as both AppInsts
	// now share the same clusterInst. But we haven't really
	// decided what happens if a user creates two of the same
	// AppInsts for the same App on the same Cloudlet (which is
	// what AppInstData[0] and AppInstData[1] do).
	// AppInstData[2] does not specify a cluster, but it
	// ends up computing key below, which corresponds to an
	// already created ClusterInst.
	// AppInstData[2]: (ClusterInstData[1])
	// Key: edgeproto.ClusterInstKey{
	//    ClusterKey:  ClusterData[0].Key,
	//    CloudletKey: CloudletData[1].Key,
	// },
	// from AppInstData[3] -> AppData[1] -> ClusterAutoData[0]:
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterAutoData[0].Key,
			CloudletKey: CloudletData[1].Key,
		},
		Flavor: ClusterData[0].DefaultFlavor,
		Auto:   true,
	},
	// from AppInstData[4] -> AppData[2] -> ClusterAutoData[1]:
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterAutoData[1].Key,
			CloudletKey: CloudletData[2].Key,
		},
		Flavor: ClusterData[1].DefaultFlavor,
		Auto:   true,
	},
}
var AppInstData = []edgeproto.AppInst{
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[0].Key,
			CloudletKey: CloudletData[0].Key,
			Id:          1,
		},
		CloudletLoc: CloudletData[0].Location,
		Uri:         "10.100.10.1",
		// Cluster is ClusterData[0]
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[0].Key,
			CloudletKey: CloudletData[0].Key,
			Id:          2,
		},
		CloudletLoc: CloudletData[0].Location,
		Uri:         "10.100.10.2",
		// Cluster is ClusterData[0]
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[0].Key,
			CloudletKey: CloudletData[1].Key,
			Id:          1,
		},
		CloudletLoc: CloudletData[1].Location,
		Uri:         "pokemon.ny.mex.io",
		// Cluster is ClusterData[0]
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[1].Key,
			CloudletKey: CloudletData[1].Key,
			Id:          1,
		},
		CloudletLoc: CloudletData[1].Location,
		Uri:         "pokemon.ny.mex.io",
		// Cluster is ClusterAutoData[0]
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[2].Key,
			CloudletKey: CloudletData[2].Key,
			Id:          1,
		},
		CloudletLoc: CloudletData[2].Location,
		Uri:         "harrypotter.sf.mex.io",
		// Cluster is ClusterAutoData[1]
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
}
var CloudletInfoData = []edgeproto.CloudletInfo{
	edgeproto.CloudletInfo{
		Key:         CloudletData[0].Key,
		State:       edgeproto.CloudletState_CloudletStateReady,
		OsMaxRam:    65536,
		OsMaxVcores: 16,
		OsMaxVolGb:  500,
	},
	edgeproto.CloudletInfo{
		Key:         CloudletData[1].Key,
		OsMaxRam:    65536,
		OsMaxVcores: 16,
		OsMaxVolGb:  500,
	},
	edgeproto.CloudletInfo{
		Key:         CloudletData[2].Key,
		OsMaxRam:    65536,
		OsMaxVcores: 16,
		OsMaxVolGb:  500,
	},
	edgeproto.CloudletInfo{
		Key:         CloudletData[3].Key,
		OsMaxRam:    65536,
		OsMaxVcores: 16,
		OsMaxVolGb:  500,
	},
}

// To figure out what resources are used on each Cloudlet,
// see ClusterInstData to see what clusters are instantiated on what Cloudlet.
var CloudletRefsData = []edgeproto.CloudletRefs{
	// ClusterInstData[0,3]:
	edgeproto.CloudletRefs{
		Key: CloudletData[0].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterData[0].Key,
			ClusterData[1].Key,
		},
		UsedRam:    GetCloudletUsedRam(0, 1),
		UsedVcores: GetCloudletUsedVcores(0, 1),
		UsedDisk:   GetCloudletUsedDisk(0, 1),
	},
	// ClusterInstData[1,4]:
	edgeproto.CloudletRefs{
		Key: CloudletData[1].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterData[0].Key,
			ClusterData[1].Key,
		},
		UsedRam:    GetCloudletUsedRam(0, 1),
		UsedVcores: GetCloudletUsedVcores(0, 1),
		UsedDisk:   GetCloudletUsedDisk(0, 1),
	},
	// ClusterInstData[2,5]:
	edgeproto.CloudletRefs{
		Key: CloudletData[2].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterData[0].Key,
			ClusterData[2].Key,
		},
		UsedRam:    GetCloudletUsedRam(0, 2),
		UsedVcores: GetCloudletUsedVcores(0, 2),
		UsedDisk:   GetCloudletUsedDisk(0, 2),
	},
	// ClusterInstData[2,6]:
	edgeproto.CloudletRefs{
		Key: CloudletData[3].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterData[3].Key,
		},
		UsedRam:    GetCloudletUsedRam(2),
		UsedVcores: GetCloudletUsedVcores(2),
		UsedDisk:   GetCloudletUsedDisk(2),
	},
}

// These Refs are after creating both cluster insts and app insts.
// Some of the app insts trigger creating auto-clusterinsts,
// and ports are reserved with the creation of app insts.
var CloudletRefsWithAppInstsData = []edgeproto.CloudletRefs{
	// ClusterInstData[0,3]:
	edgeproto.CloudletRefs{
		Key: CloudletData[0].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterData[0].Key,
			ClusterData[1].Key,
		},
		UsedRam:        GetCloudletUsedRam(0, 1),
		UsedVcores:     GetCloudletUsedVcores(0, 1),
		UsedDisk:       GetCloudletUsedDisk(0, 1),
		UsedDynamicIps: 2,
	},
	// ClusterInstData[1,4], ClusterInstAutoData[0]:
	edgeproto.CloudletRefs{
		Key: CloudletData[1].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterData[0].Key,
			ClusterData[1].Key,
			ClusterAutoData[0].Key,
		},
		UsedRam:        GetCloudletUsedRam(0, 1, 0),
		UsedVcores:     GetCloudletUsedVcores(0, 1, 0),
		UsedDisk:       GetCloudletUsedDisk(0, 1, 0),
		RootLbPorts:    map[int32]int32{10000: 1, 10001: 1},
		UsedDynamicIps: 1,
	},
	// ClusterInstData[2,5], ClusterInstAutoData[1]:
	edgeproto.CloudletRefs{
		Key: CloudletData[2].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterData[0].Key,
			ClusterData[2].Key,
			ClusterAutoData[1].Key,
		},
		UsedRam:        GetCloudletUsedRam(0, 2, 1),
		UsedVcores:     GetCloudletUsedVcores(0, 2, 1),
		UsedDisk:       GetCloudletUsedDisk(0, 2, 1),
		UsedDynamicIps: 1,
	},
	// ClusterInstData[2,6]:
	edgeproto.CloudletRefs{
		Key: CloudletData[3].Key,
		Clusters: []edgeproto.ClusterKey{
			ClusterData[3].Key,
		},
		UsedRam:    GetCloudletUsedRam(2),
		UsedVcores: GetCloudletUsedVcores(2),
		UsedDisk:   GetCloudletUsedDisk(2),
	},
}

func FindFlavorData(key *edgeproto.FlavorKey) *edgeproto.Flavor {
	for ii, _ := range FlavorData {
		if FlavorData[ii].Key.Matches(key) {
			return &FlavorData[ii]
		}
	}
	return nil
}

func GetCloudletUsedRam(indices ...int) uint64 {
	var ram uint64
	for _, idx := range indices {
		clflavor := ClusterFlavorData[idx]
		flavor := FindFlavorData(&clflavor.NodeFlavor)
		ram += flavor.Ram * uint64(clflavor.MaxNodes)
	}
	return ram
}

func GetCloudletUsedVcores(indices ...int) uint64 {
	var vcores uint64
	for _, idx := range indices {
		clflavor := ClusterFlavorData[idx]
		flavor := FindFlavorData(&clflavor.NodeFlavor)
		vcores += flavor.Vcpus * uint64(clflavor.MaxNodes)
	}
	return vcores
}

func GetCloudletUsedDisk(indices ...int) uint64 {
	var disk uint64
	for _, idx := range indices {
		clflavor := ClusterFlavorData[idx]
		flavor := FindFlavorData(&clflavor.NodeFlavor)
		disk += flavor.Disk * uint64(clflavor.MaxNodes)
	}
	return disk
}
