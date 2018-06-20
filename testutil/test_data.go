package testutil

import "github.com/mobiledgex/edge-cloud/edgeproto"

var DevData = []edgeproto.Developer{
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "Atlantic, Inc.",
		},
		Address: "1230 Midas Way #200, Sunnyvale, CA 94085",
		Email:   "edge.atlantic.com",
	},
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "Eaiever",
		},
		Address: "1 Letterman Drive Building C, San Francisco, CA 94129",
		Email:   "edge.everai.com",
	},
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "1000 realities",
		},
		Address: "Kamienna 43, 31-403 Kraken, Poland",
		Email:   "edge.Untomt.com",
	},
	edgeproto.Developer{
		Key: edgeproto.DeveloperKey{
			Name: "Sierraware LLC",
		},
		Address: "1250 Oakmead Parkway Suite 210, Sunnyvalue, CA 94085",
		Email:   "support@sierraware.com",
	},
}
var AppData = []edgeproto.App{
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pillimo Go!",
			Version:      "1.0.0",
		},
		AppPath: "/foo/bar/bin",
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pillimo Go!",
			Version:      "1.0.1",
		},
		AppPath: "foo/bar/bin/1.0.1",
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Hunna Stoll Go! Go!",
			Version:      "0.0.1",
		},
		AppPath: "/some/path",
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[1].Key,
			Name:         "AI",
			Version:      "1.2.0",
		},
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[2].Key,
			Name:         "my reality",
			Version:      "0.0.1",
		},
	},
}
var OperatorData = []edgeproto.Operator{
	edgeproto.Operator{
		Key: edgeproto.OperatorKey{
			Name: "UFGT Inc.",
		},
	},
	edgeproto.Operator{
		Key: edgeproto.OperatorKey{
			Name: "xmobx",
		},
	},
	edgeproto.Operator{
		Key: edgeproto.OperatorKey{
			Name: "Zerilu",
		},
	},
	edgeproto.Operator{
		Key: edgeproto.OperatorKey{
			Name: "Denton telecom",
		},
	},
}
var CloudletData = []edgeproto.Cloudlet{
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[0].Key,
			Name:        "San Jose Site",
		},
		AccessIp: []byte{10, 100, 0, 1},
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[0].Key,
			Name:        "New York Site",
		},
		AccessIp: []byte{254, 8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[1].Key,
			Name:        "San Francisco Site",
		},
		AccessIp: []byte{172, 24, 0, 1},
	},
	edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			OperatorKey: OperatorData[2].Key,
			Name:        "Hawaii Site",
		},
		AccessIp: []byte{172, 30, 0, 1},
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
		Liveness:    edgeproto.AppInst_STATIC,
		Ip:          []byte{10, 100, 10, 1},
		Port:        8089,
		AppPath:     AppData[0].AppPath,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[0].Key,
			CloudletKey: CloudletData[0].Key,
			Id:          2,
		},
		CloudletLoc: CloudletData[0].Location,
		Liveness:    edgeproto.AppInst_DYNAMIC,
		Ip:          []byte{10, 100, 10, 2},
		Port:        8089,
		AppPath:     AppData[0].AppPath,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[0].Key,
			CloudletKey: CloudletData[1].Key,
			Id:          1,
		},
		CloudletLoc: CloudletData[1].Location,
		Liveness:    edgeproto.AppInst_STATIC,
		Ip:          []byte{172, 24, 1, 1},
		Port:        1443,
		AppPath:     AppData[0].AppPath,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[1].Key,
			CloudletKey: CloudletData[1].Key,
			Id:          1,
		},
		CloudletLoc: CloudletData[1].Location,
		Liveness:    edgeproto.AppInst_STATIC,
		Ip:          []byte{172, 24, 1, 1},
		Port:        2443,
		AppPath:     AppData[1].AppPath,
	},
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[2].Key,
			CloudletKey: CloudletData[2].Key,
			Id:          1,
		},
		CloudletLoc: CloudletData[2].Location,
		Liveness:    edgeproto.AppInst_STATIC,
		Ip:          []byte{192, 168, 1, 1},
		Port:        54321,
		AppPath:     AppData[2].AppPath,
	},
}
