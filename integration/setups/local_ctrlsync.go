package setups

import "github.com/mobiledgex/edge-cloud/integration/process"

// This setup directs each controller to a single separate etcd instance so
// that any update to one controller must flow to etcd, to another etcd,
// and then back to another controller. This allows us to test the
// synchronization between controllers via etcd watch.

var localCtrlSyncEtcdCluster = "etcd1=http://127.0.0.1:30011,etcd2=http://127.0.0.1:30012,etcd3=http://127.0.0.1:30013"

var LocalCtrlSync = process.ProcessSetup{
	Etcds: []process.EtcdProcess{
		&process.EtcdLocal{
			Name:           "etcd1",
			DataDir:        "/var/tmp/edge-cloud-local-etcd/etcd1",
			PeerAddrs:      "http://127.0.0.1:30011",
			ClientAddrs:    "http://127.0.0.1:30001",
			InitialCluster: localCtrlSyncEtcdCluster,
		},
		&process.EtcdLocal{
			Name:           "etcd2",
			DataDir:        "/var/tmp/edge-cloud-local-etcd/etcd2",
			PeerAddrs:      "http://127.0.0.1:30012",
			ClientAddrs:    "http://127.0.0.1:30002",
			InitialCluster: localCtrlSyncEtcdCluster,
		},
		&process.EtcdLocal{
			Name:           "etcd3",
			DataDir:        "/var/tmp/edge-cloud-local-etcd/etcd3",
			PeerAddrs:      "http://127.0.0.1:30013",
			ClientAddrs:    "http://127.0.0.1:30003",
			InitialCluster: localCtrlSyncEtcdCluster,
		},
	},
	Controllers: []process.ControllerProcess{
		&process.ControllerLocal{
			Name:      "ctrl1",
			EtcdAddrs: "http://127.0.0.1:30001",
			ApiAddr:   "127.0.0.1:35001",
			HttpAddr:  "127.0.0.1:36001",
		},
		&process.ControllerLocal{
			Name:      "ctrl2",
			EtcdAddrs: "http://127.0.0.1:30002",
			ApiAddr:   "127.0.0.1:35002",
			HttpAddr:  "127.0.0.1:36002",
		},
		&process.ControllerLocal{
			Name:      "ctrl3",
			EtcdAddrs: "http://127.0.0.1:30003",
			ApiAddr:   "127.0.0.1:35003",
			HttpAddr:  "127.0.0.1:36003",
		},
	},
}
