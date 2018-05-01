package main

import "github.com/mobiledgex/edge-cloud/proto"

var (
	Dev1 = proto.Developer{
		Key: &proto.DeveloperKey{
			Name: "Niantic, Inc.",
		},
		Address: "1230 Midas Way #200, Sunnyvale, CA 94085",
		Email:   "edge.niantic.com",
	}
	Dev2 = proto.Developer{
		Key: &proto.DeveloperKey{
			Name: "Ever.ai",
		},
		Address: "1 Letterman Drive Building C, San Francisco, CA 94129",
		Email:   "edge.everai.com",
	}
	Dev3 = proto.Developer{
		Key: &proto.DeveloperKey{
			Name: "1000 realities",
		},
		Address: "Kamienna 43, 31-403 Krakow, Poland",
		Email:   "edge.1000realities.com",
	}
	Dev4 = proto.Developer{
		Key: &proto.DeveloperKey{
			Name: "Sierraware LLC",
		},
		Address: "1250 Oakmead Parkway Suite 210, Sunnyvalue, CA 94085",
		Email:   "support@sierraware.com",
	}
)
