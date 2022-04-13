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
