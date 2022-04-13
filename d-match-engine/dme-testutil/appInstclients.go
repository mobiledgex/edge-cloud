// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
				ClusterInstKey: edgeproto.VirtualClusterInstKey{
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
				ClusterInstKey: edgeproto.VirtualClusterInstKey{
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
				ClusterInstKey: edgeproto.VirtualClusterInstKey{
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
				ClusterInstKey: edgeproto.VirtualClusterInstKey{
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
				ClusterInstKey: edgeproto.VirtualClusterInstKey{
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
