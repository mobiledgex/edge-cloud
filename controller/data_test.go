package main

import "github.com/mobiledgex/edge-cloud/proto"

var DevData = []proto.Developer{
	proto.Developer{
		Key: proto.DeveloperKey{
			Name: "Niantic, Inc.",
		},
		Address: "1230 Midas Way #200, Sunnyvale, CA 94085",
		Email:   "edge.niantic.com",
	},
	proto.Developer{
		Key: proto.DeveloperKey{
			Name: "Ever.ai",
		},
		Address: "1 Letterman Drive Building C, San Francisco, CA 94129",
		Email:   "edge.everai.com",
	},
	proto.Developer{
		Key: proto.DeveloperKey{
			Name: "1000 realities",
		},
		Address: "Kamienna 43, 31-403 Krakow, Poland",
		Email:   "edge.1000realities.com",
	},
	proto.Developer{
		Key: proto.DeveloperKey{
			Name: "Sierraware LLC",
		},
		Address: "1250 Oakmead Parkway Suite 210, Sunnyvalue, CA 94085",
		Email:   "support@sierraware.com",
	},
}
var AppData = []proto.App{
	proto.App{
		Key: proto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pokemon Go!",
			Version:      "1.0.0",
		},
		AppPath: "/foo/bar/bin",
	},
	proto.App{
		Key: proto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Pokemon Go!",
			Version:      "1.0.1",
		},
		AppPath: "foo/bar/bin/1.0.1",
	},
	proto.App{
		Key: proto.AppKey{
			DeveloperKey: DevData[0].Key,
			Name:         "Harry Potter Go! Go!",
			Version:      "0.0.1",
		},
		AppPath: "/some/path",
	},
	proto.App{
		Key: proto.AppKey{
			DeveloperKey: DevData[1].Key,
			Name:         "AI",
			Version:      "1.2.0",
		},
	},
	proto.App{
		Key: proto.AppKey{
			DeveloperKey: DevData[2].Key,
			Name:         "my reality",
			Version:      "0.0.1",
		},
	},
}
var OperatorData = []proto.Operator{
	proto.Operator{
		Key: proto.OperatorKey{
			Name: "AT&T Inc.",
		},
	},
	proto.Operator{
		Key: proto.OperatorKey{
			Name: "T-Mobile",
		},
	},
	proto.Operator{
		Key: proto.OperatorKey{
			Name: "Verizon",
		},
	},
	proto.Operator{
		Key: proto.OperatorKey{
			Name: "Deutsche Telekom",
		},
	},
}
var CloudletData = []proto.Cloudlet{
	proto.Cloudlet{
		Key: proto.CloudletKey{
			OperatorKey: OperatorData[0].Key,
			Name:        "San Jose Site",
		},
		AccessIp: []byte{10, 100, 0, 1},
	},
	proto.Cloudlet{
		Key: proto.CloudletKey{
			OperatorKey: OperatorData[0].Key,
			Name:        "New York Site",
		},
		AccessIp: []byte{254, 8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	},
	proto.Cloudlet{
		Key: proto.CloudletKey{
			OperatorKey: OperatorData[1].Key,
			Name:        "San Francisco Site",
		},
		AccessIp: []byte{172, 24, 0, 1},
	},
	proto.Cloudlet{
		Key: proto.CloudletKey{
			OperatorKey: OperatorData[2].Key,
			Name:        "Hawaii Site",
		},
		AccessIp: []byte{172, 30, 0, 1},
	},
}
