package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	log "github.com/mobiledgex/edge-cloud/log"
)

const (
	RequestAppInstLatency = "request-appinst-latency"
	DisplayAppInstLatency = "display-appinst-latency"
)

func InitDebug(nodeMgr *node.NodeMgr) {
	nodeMgr.Debug.AddDebugFunc(RequestAppInstLatency, requestAppInstLatency)
	nodeMgr.Debug.AddDebugFunc(DisplayAppInstLatency, displayAppInstLatency)
}

func requestAppInstLatency(ctx context.Context, req *edgeproto.DebugRequest) string {
	log.SpanLog(ctx, log.DebugLevelDmereq, "Received request-appinst-latency in dme", "request", req)

	appInstKey, err := createAppInstKeyFromRequest(req)
	if err != nil {
		return err.Error()
	}

	dmecommon.EEHandler.SendLatencyRequestEdgeEvent(ctx, *appInstKey)
	return "successfully sent latency request"
}

func displayAppInstLatency(ctx context.Context, req *edgeproto.DebugRequest) string {
	log.SpanLog(ctx, log.DebugLevelDmereq, "Received display-appinst-latency in dme", "request", req)

	appInstKey, err := createAppInstKeyFromRequest(req)
	if err != nil {
		return err.Error()
	}

	apiStatCall := &ApiStatCall{
		key: dmecommon.StatKey{
			AppKey:         appInstKey.AppKey,
			CloudletFound:  appInstKey.ClusterInstKey.CloudletKey,
			Method:         EdgeEventLatencyMethod,
			ClusterKey:     appInstKey.ClusterInstKey.ClusterKey,
			ClusterInstOrg: appInstKey.ClusterInstKey.Organization,
		},
	}

	apiStat, found := stats.LookupApiStatCall(apiStatCall)
	if !found {
		return "unable to find apiStat"
	}
	apiStat.mux.Lock()
	defer apiStat.mux.Unlock()

	latency := apiStat.rollinglatency
	b, err := json.Marshal(latency)
	if err != nil {
		return "unable to marshall latency"
	}
	return string(b)
}

func createAppInstKeyFromRequest(req *edgeproto.DebugRequest) (*edgeproto.AppInstKey, error) {
	if req.Args == "" {
		return nil, fmt.Errorf("appinst info in args required")
	}
	rd := csv.NewReader(strings.NewReader(req.Args))
	rd.Comma = ' '
	args, err := rd.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to split args string, %v", err)
	}
	appname := args[0]
	apporg := args[1]
	appvers := args[2]
	cloudlet := args[3]
	cloudletorg := args[4]
	cluster := args[5]
	clusterorg := args[6]

	// TODO: Check if args exist

	/*if req.Node.CloudletKey == nil {
		return "CloudletKey required"
	}*/

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

	return appInstKey, nil
}
