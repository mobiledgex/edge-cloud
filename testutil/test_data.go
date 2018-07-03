package testutil

import "github.com/mobiledgex/edge-cloud/edgeproto"

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
var AppData = []edgeproto.App{
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pokemon Go!",
			Version:      "1.0.0",
		},
		AppPath: "/foo/bar/bin",
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pokemon Go!",
			Version:      "1.0.1",
		},
		AppPath: "foo/bar/bin/1.0.1",
	},
	edgeproto.App{
		Key: edgeproto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Harry Potter Go! Go!",
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
var AppInstData = []edgeproto.AppInst{
	edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      AppData[0].Key,
			CloudletKey: CloudletData[0].Key,
			Id:          1,
		},
		CloudletLoc: CloudletData[0].Location,
		Liveness:    edgeproto.AppInst_STATIC,
		Uri:         "10.100.10.1",
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
		Uri:         "10.100.10.2",
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
		Uri:         "pokemon.ny.mex.io",
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
		Uri:         "pokemon.ny.mex.io",
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
		Uri:         "harrypotter.sf.mex.io",
		AppPath:     AppData[2].AppPath,
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
