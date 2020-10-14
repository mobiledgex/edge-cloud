package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dmeutil "github.com/mobiledgex/edge-cloud/d-match-engine/dme-util"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func InitDebug(nodeMgr *node.NodeMgr) {
	nodeMgr.Debug.AddDebugFunc(dmeutil.RequestAppInstLatency, requestAppInstLatency)
	nodeMgr.Debug.AddDebugFunc(dmeutil.ShowAppInstLatency, showAppInstLatency)
}

func requestAppInstLatency(ctx context.Context, req *edgeproto.DebugRequest) string {
	appInstKey, err := createAppInstKeyFromRequest(req)
	if err != nil {
		return err.Error()
	}

	dmecommon.EEHandler.SendLatencyRequestEdgeEvent(ctx, *appInstKey)
	return "successfully sent latency request"
}

func showAppInstLatency(ctx context.Context, req *edgeproto.DebugRequest) string {
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
	args := strings.Split(req.Args, " ")
	if len(args) != 7 {
		return nil, fmt.Errorf("7 arguments required: appname, apporg, appvers, cloudlet, cloudletorg, cluster, clusterorg")
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

	return appInstKey, nil
}
