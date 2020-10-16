package testutil

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AllDataStreamOut struct {
	StreamObjs []edgeproto.StreamObj
}

func RunAllDataStreamApis(run *Run, in *edgeproto.AllData, out *AllDataStreamOut) {
	run.Mode = "streamappinst"
	for _, appInst := range in.AppInstances {
		appInstKeys := []edgeproto.AppInstKey{appInst.Key}
		outMsgs := [][]edgeproto.StreamMsg{}
		run.StreamObjApi_AppInstKey(&appInstKeys, nil, &outMsgs)
		outObj := edgeproto.StreamObj{Key: appInst.Key}
		for _, objsMsgs := range outMsgs {
			for _, msg := range objsMsgs {
				addMsg := edgeproto.StreamMsg{}
				addMsg = msg
				outObj.Msgs = append(outObj.Msgs, &addMsg)
			}
		}
		out.StreamObjs = append(out.StreamObjs, outObj)
	}

	run.Mode = "streamclusterinst"
	for _, clusterInst := range in.ClusterInsts {
		clusterInstKeys := []edgeproto.ClusterInstKey{clusterInst.Key}
		outMsgs := [][]edgeproto.StreamMsg{}
		run.StreamObjApi_ClusterInstKey(&clusterInstKeys, nil, &outMsgs)
		streamKey := edgeproto.AppInstKey{
			ClusterInstKey: clusterInst.Key,
		}
		outObj := edgeproto.StreamObj{Key: streamKey}
		for _, objsMsgs := range outMsgs {
			for _, msg := range objsMsgs {
				addMsg := edgeproto.StreamMsg{}
				addMsg = msg
				outObj.Msgs = append(outObj.Msgs, &addMsg)
			}
		}
		out.StreamObjs = append(out.StreamObjs, outObj)
	}

	run.Mode = "streamcloudlet"
	for _, cloudlet := range in.Cloudlets {
		cloudletKeys := []edgeproto.CloudletKey{cloudlet.Key}
		outMsgs := [][]edgeproto.StreamMsg{}
		run.StreamObjApi_CloudletKey(&cloudletKeys, nil, &outMsgs)
		streamKey := edgeproto.AppInstKey{
			ClusterInstKey: edgeproto.ClusterInstKey{
				CloudletKey: cloudlet.Key,
			},
		}
		outObj := edgeproto.StreamObj{Key: streamKey}
		for _, objsMsgs := range outMsgs {
			for _, msg := range objsMsgs {
				addMsg := edgeproto.StreamMsg{}
				addMsg = msg
				outObj.Msgs = append(outObj.Msgs, &addMsg)
			}
		}
		out.StreamObjs = append(out.StreamObjs, outObj)
	}
}
