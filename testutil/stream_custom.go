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

package testutil

import (
	"github.com/edgexr/edge-cloud/edgeproto"
)

type AllDataStreamOut struct {
	StreamObjs []StreamObj
}

type StreamObj struct {
	Key  edgeproto.AppInstKey
	Msgs []edgeproto.Result
}

func RunAllDataStreamApis(run *Run, in *edgeproto.AllData, out *AllDataStreamOut) {
	run.Mode = "streamappinst"
	for _, appInst := range in.AppInstances {
		appInstKeys := []edgeproto.AppInstKey{appInst.Key}
		outMsgs := [][]edgeproto.Result{}
		run.StreamObjApi_AppInstKey(&appInstKeys, nil, &outMsgs)
		outObj := StreamObj{Key: appInst.Key}
		for _, objsMsgs := range outMsgs {
			outObj.Msgs = objsMsgs
		}
		out.StreamObjs = append(out.StreamObjs, outObj)
	}

	run.Mode = "streamclusterinst"
	for _, clusterInst := range in.ClusterInsts {
		clusterInstKeys := []edgeproto.ClusterInstKey{clusterInst.Key}
		outMsgs := [][]edgeproto.Result{}
		run.StreamObjApi_ClusterInstKey(&clusterInstKeys, nil, &outMsgs)
		streamKey := edgeproto.AppInstKey{
			ClusterInstKey: *clusterInst.Key.Virtual(""),
		}
		outObj := StreamObj{Key: streamKey}
		for _, objsMsgs := range outMsgs {
			outObj.Msgs = objsMsgs
		}
		out.StreamObjs = append(out.StreamObjs, outObj)
	}

	run.Mode = "streamcloudlet"
	for _, cloudlet := range in.Cloudlets {
		cloudletKeys := []edgeproto.CloudletKey{cloudlet.Key}
		outMsgs := [][]edgeproto.Result{}
		run.StreamObjApi_CloudletKey(&cloudletKeys, nil, &outMsgs)
		streamKey := edgeproto.AppInstKey{
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
				CloudletKey: cloudlet.Key,
			},
		}
		outObj := StreamObj{Key: streamKey}
		for _, objsMsgs := range outMsgs {
			outObj.Msgs = objsMsgs
		}
		out.StreamObjs = append(out.StreamObjs, outObj)
	}
}
