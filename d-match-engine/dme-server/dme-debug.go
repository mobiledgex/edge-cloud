package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	log "github.com/mobiledgex/edge-cloud/log"
)

func InitDebug(nodeMgr *node.NodeMgr) {
	nodeMgr.Debug.AddDebugFunc("measure-appinst-latency", runDmeCmd)
}

func runDmeCmd(ctx context.Context, req *edgeproto.DebugRequest) string {
	log.SpanLog(ctx, log.DebugLevelDmereq, "Received rundebug in dme", "request", req)
	if req.Args == "" {
		return "appinst info in args required"
	}
	rd := csv.NewReader(strings.NewReader(req.Args))
	rd.Comma = ' '
	args, err := rd.Read()
	if err != nil {
		return fmt.Sprintf("failed to split args string, %v", err)
	}
	appname := args[0]
	apporg := args[1]
	appvers := args[2]
	cloudlet := args[3]
	cloudletorg := args[4]
	cluster := args[5]
	clusterorg := args[6]

	appInstKey := &edgeproto.AppInstKey{
		AppKey: edgeproto.AppKey{
			Organization: apporg,
			Name:         appname,
			Version:      appvers,
		},
		ClusterInstKey: edgeproto.ClusterInstKey{
			ClusterKey: edgeproto.ClusterKey{
				Name: cluster,
			},
			CloudletKey: edgeproto.CloudletKey{
				Organization: cloudletorg,
				Name:         cloudlet,
			},
			Organization: clusterorg,
		},
	}

	/*if req.Node.CloudletKey == nil {
		return "CloudletKey required"
	}*/

	dmecommon.EEHandler.SendLatencyRequestEdgeEvent(ctx, *appInstKey)
	return "successfully sent latency request"
}
