// Run Etcd as a child process.
// May be useful for testing and initial development.

package main

import (
	"path/filepath"
	"runtime"

	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
)

const EtcdLocalData string = "etcdLocal_data"
const EtcdLocalLog string = "etcdLocal.log"

func StartLocalEtcdServer(opts ...process.StartOp) (*process.Etcd, error) {
	_, filename, _, _ := runtime.Caller(0)
	testdir := filepath.Dir(filename) + "/" + EtcdLocalData

	etcd := &process.Etcd{
		Common: process.Common{
			Name: "etcd-local",
		},
		DataDir:        testdir,
		PeerAddrs:      "http://127.0.0.1:52379",
		ClientAddrs:    "http://127.0.0.1:52380",
		InitialCluster: "etcd-local=http://127.0.0.1:52379",
	}
	log.InfoLog("Starting local etcd", "clientUrls", etcd.ClientAddrs)
	err := etcd.StartLocal("", opts...)
	if err != nil {
		return nil, err
	}
	return etcd, nil
}
