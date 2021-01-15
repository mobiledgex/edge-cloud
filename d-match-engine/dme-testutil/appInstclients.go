package dmetest

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// AppInstClients - test the Clients
var AppInstClientData = []edgeproto.AppInstClient{
	edgeproto.AppInstClient{
		ClientKey: edgeproto.AppInstClientKey{
			AppInstKey: edgeproto.AppInstKey{
				AppKey: edgeproto.AppKey{
					Name:         "app1",
					Organization: "devorg1",
					Version:      "1.0",
				},
				ClusterInstKey: edgeproto.ClusterInstKey{
					CloudletKey: edgeproto.CloudletKey{
						Name:         "cloudlet1",
						Organization: "operator1",
					},
				},
			},
			UniqueId:     "1",
			UniqueIdType: "testuuid",
		},
	},
	edgeproto.AppInstClient{
		ClientKey: edgeproto.AppInstClientKey{
			AppInstKey: edgeproto.AppInstKey{
				AppKey: edgeproto.AppKey{
					Name:         "app2",
					Organization: "devorg2",
					Version:      "1.0",
				},
				ClusterInstKey: edgeproto.ClusterInstKey{
					CloudletKey: edgeproto.CloudletKey{
						Name:         "cloudlet1",
						Organization: "operator1",
					},
				},
			},
			UniqueId:     "2",
			UniqueIdType: "testuuid",
		},
	},
	// Same as AppInstClientData[0], but on a different cloudlet
	edgeproto.AppInstClient{
		ClientKey: edgeproto.AppInstClientKey{
			AppInstKey: edgeproto.AppInstKey{
				AppKey: edgeproto.AppKey{
					Name:         "app1",
					Organization: "devorg1",
					Version:      "1.0",
				},
				ClusterInstKey: edgeproto.ClusterInstKey{
					CloudletKey: edgeproto.CloudletKey{
						Name:         "cloudlet2",
						Organization: "operator2",
					},
				},
			},
			UniqueId:     "1",
			UniqueIdType: "testuuid",
		},
	},
	edgeproto.AppInstClient{
		ClientKey: edgeproto.AppInstClientKey{
			AppInstKey: edgeproto.AppInstKey{
				AppKey: edgeproto.AppKey{
					Name:         "app1",
					Organization: "devorg1",
					Version:      "1.0",
				},
				ClusterInstKey: edgeproto.ClusterInstKey{
					CloudletKey: edgeproto.CloudletKey{
						Name:         "cloudlet1",
						Organization: "operator1",
					},
				},
			},
			UniqueId:     "3",
			UniqueIdType: "testuuid",
		},
	},
	edgeproto.AppInstClient{
		ClientKey: edgeproto.AppInstClientKey{
			AppInstKey: edgeproto.AppInstKey{
				AppKey: edgeproto.AppKey{
					Name:         "app1",
					Organization: "devorg1",
					Version:      "1.0",
				},
				ClusterInstKey: edgeproto.ClusterInstKey{
					CloudletKey: edgeproto.CloudletKey{
						Name:         "cloudlet1",
						Organization: "operator1",
					},
				},
			},
			UniqueId:     "4",
			UniqueIdType: "testuuid",
		},
	},
}
