package testutil

import "github.com/mobiledgex/edge-cloud/edgeproto"

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
		Flavor: FlavorData[0].Key,
		Nodes:  3,
	},
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: "Ever.Ai",
		},
		Flavor: FlavorData[1].Key,
		Nodes:  3,
	},
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: "1000realities",
		},
		Flavor: FlavorData[2].Key,
		Nodes:  3,
	},
	edgeproto.Cluster{
		Key: edgeproto.ClusterKey{
			Name: "Big-Pokemons",
		},
		Flavor: FlavorData[2].Key,
		Nodes:  5,
	},
}
var AppData = []edgeproto.App{
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pokemon Go!",
			Version:      "1.0.0",
		},
		ImageType:   edgeproto.ImageType_ImageTypeDocker,
		AccessLayer: edgeproto.AccessLayer_AccessLayerL7,
		Flavor:      FlavorData[0].Key,
		Cluster:     ClusterData[0].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pokemon Go!",
			Version:      "1.0.1",
		},
		ImageType:   edgeproto.ImageType_ImageTypeDocker,
		AccessLayer: edgeproto.AccessLayer_AccessLayerL7,
		Flavor:      FlavorData[0].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Harry Potter Go! Go!",
			Version:      "0.0.1",
		},
		ImageType:   edgeproto.ImageType_ImageTypeDocker,
		AccessLayer: edgeproto.AccessLayer_AccessLayerL7,
		Flavor:      FlavorData[1].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[1].Key,
			Name:         "AI",
			Version:      "1.2.0",
		},
		ImageType:   edgeproto.ImageType_ImageTypeQCOW,
		AccessLayer: edgeproto.AccessLayer_AccessLayerL4,
		AccessPorts: "12345-12349",
		Flavor:      FlavorData[1].Key,
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[2].Key,
			Name:         "my reality",
			Version:      "0.0.1",
		},
		ImageType:   edgeproto.ImageType_ImageTypeQCOW,
		AccessLayer: edgeproto.AccessLayer_AccessLayerL4L7,
		AccessPorts: "80,10080",
		Flavor:      FlavorData[2].Key,
		Cluster:     ClusterData[2].Key,
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
		AccessUri: "10.100.0.1",
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[0].Key,
			Name:        "New York Site",
		},
		AccessUri: "ff.f8::1",
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[1].Key,
			Name:        "San Francisco Site",
		},
		AccessUri: "172.24.0.1",
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[2].Key,
			Name:        "Hawaii Site",
		},
		AccessUri: "hi76.verizon.com",
	},
}
var ClusterInstData = []edgeproto.ClusterInst{
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[0].Key,
			CloudletKey: CloudletData[0].Key,
		},
		Flavor: ClusterData[0].Flavor,
		Nodes:  ClusterData[0].Nodes,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[0].Key,
			CloudletKey: CloudletData[1].Key,
		},
		Flavor: ClusterData[0].Flavor,
		Nodes:  ClusterData[0].Nodes,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[0].Key,
			CloudletKey: CloudletData[2].Key,
		},
		Flavor: ClusterData[0].Flavor,
		Nodes:  ClusterData[0].Nodes,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[1].Key,
			CloudletKey: CloudletData[0].Key,
		},
		Flavor: ClusterData[1].Flavor,
		Nodes:  ClusterData[1].Nodes,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[1].Key,
			CloudletKey: CloudletData[1].Key,
		},
		Flavor: ClusterData[1].Flavor,
		Nodes:  ClusterData[1].Nodes,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[2].Key,
			CloudletKey: CloudletData[2].Key,
		},
		Flavor: ClusterData[2].Flavor,
		Nodes:  ClusterData[2].Nodes,
	},
	edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey:  ClusterData[3].Key,
			CloudletKey: CloudletData[3].Key,
		},
		Flavor: ClusterData[3].Flavor,
		Nodes:  ClusterData[3].Nodes,
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
		ImagePath:   AppData[0].ImagePath,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[0].Key,
			CloudletKey: CloudletData[0].Key,
			Id:          2,
		},
		CloudletLoc: CloudletData[0].Location,
		Uri:         "10.100.10.2",
		ImagePath:   AppData[0].ImagePath,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[0].Key,
			CloudletKey: CloudletData[1].Key,
			Id:          1,
		},
		CloudletLoc: CloudletData[1].Location,
		Uri:         "pokemon.ny.mex.io",
		ImagePath:   AppData[0].ImagePath,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[1].Key,
			CloudletKey: CloudletData[1].Key,
			Id:          1,
		},
		CloudletLoc: CloudletData[1].Location,
		Uri:         "pokemon.ny.mex.io",
		ImagePath:   AppData[1].ImagePath,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[2].Key,
			CloudletKey: CloudletData[2].Key,
			Id:          1,
		},
		CloudletLoc: CloudletData[2].Location,
		Uri:         "harrypotter.sf.mex.io",
		ImagePath:   AppData[2].ImagePath,
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
		Key: CloudletData[0].Key,
	},
	edgeproto.CloudletInfo{
		Key: CloudletData[1].Key,
	},
	edgeproto.CloudletInfo{
		Key: CloudletData[2].Key,
	},
	edgeproto.CloudletInfo{
		Key: CloudletData[3].Key,
	},
}
