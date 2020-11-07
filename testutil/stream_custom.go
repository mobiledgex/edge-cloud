package testutil

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
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
			ClusterInstKey: clusterInst.Key,
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
			ClusterInstKey: edgeproto.ClusterInstKey{
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
