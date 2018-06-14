package setups

import "github.com/mobiledgex/edge-cloud/integration/process"

var localBasicEtcdCluster = "etcd1=http://127.0.0.1:30011,etcd2=http://127.0.0.1:30012,etcd3=http://127.0.0.1:30013"
var localBasicEtcdAddrs = "http://127.0.0.1:30001,http://127.0.0.1:30002,http://127.0.0.1:30003"
var localBasicNotifyAddrs = "127.0.0.1:37001,127.0.0.1:37002,127.0.0.1:37003"

var LocalBasic = process.ProcessSetup{
	Etcds: []process.EtcdProcess{
		&process.EtcdLocal{
			Name:           "etcd1",
			DataDir:        "/var/tmp/edge-cloud-local-etcd/etcd1",
			PeerAddrs:      "http://127.0.0.1:30011",
			ClientAddrs:    "http://127.0.0.1:30001",
			InitialCluster: localBasicEtcdCluster,
		},
		&process.EtcdLocal{
			Name:           "etcd2",
			DataDir:        "/var/tmp/edge-cloud-local-etcd/etcd2",
			PeerAddrs:      "http://127.0.0.1:30012",
			ClientAddrs:    "http://127.0.0.1:30002",
			InitialCluster: localBasicEtcdCluster,
		},
		&process.EtcdLocal{
			Name:           "etcd3",
			DataDir:        "/var/tmp/edge-cloud-local-etcd/etcd3",
			PeerAddrs:      "http://127.0.0.1:30013",
			ClientAddrs:    "http://127.0.0.1:30003",
			InitialCluster: localBasicEtcdCluster,
		},
	},
	Controllers: []process.ControllerProcess{
		&process.ControllerLocal{
			Name:       "ctrl1",
			EtcdAddrs:  localBasicEtcdAddrs,
			ApiAddr:    "127.0.0.1:35001",
			HttpAddr:   "127.0.0.1:36001",
			NotifyAddr: "127.0.0.1:37001",
		},
		&process.ControllerLocal{
			Name:       "ctrl2",
			EtcdAddrs:  localBasicEtcdAddrs,
			ApiAddr:    "127.0.0.1:35002",
			HttpAddr:   "127.0.0.1:36002",
			NotifyAddr: "127.0.0.1:37002",
		},
		&process.ControllerLocal{
			Name:       "ctrl3",
			EtcdAddrs:  localBasicEtcdAddrs,
			ApiAddr:    "127.0.0.1:35003",
			HttpAddr:   "127.0.0.1:36003",
			NotifyAddr: "127.0.0.1:37003",
		},
	},
	Dmes: []process.DmeProcess{
		&process.DmeLocal{
			Name:        "dme1",
			ApiAddr:     "127.0.0.1:50051", //TODO: DME has hardcoded listen port
			NotifyAddrs: localBasicNotifyAddrs,
		},
		&process.DmeLocal{
			Name:        "dme2",
			ApiAddr:     "127.0.0.1:50051", //TODO: DME has hardcoded listen port
			NotifyAddrs: localBasicNotifyAddrs,
		},
	},
	Crms: []process.CrmProcess{
		&process.CrmLocal{
			Name:        "crm1",
			ApiAddr:     "127.0.0.1:55099", //TODO: CRM has hardcoded listen port
			NotifyAddrs: localBasicNotifyAddrs,
		},
		&process.CrmLocal{
			Name:        "crm2",
			ApiAddr:     "127.0.0.1:55099", //TODO: CRM has hardcoded listen port
			NotifyAddrs: localBasicNotifyAddrs,
		},
	},
}
