package main

import (
	"context"
	"encoding/json"
	"fmt"

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

	apiStatCall := &dmecommon.ApiStatCall{
		Key: dmecommon.StatKey{
			AppKey:         appInstKey.AppKey,
			CloudletFound:  appInstKey.ClusterInstKey.CloudletKey,
			Method:         dmecommon.EdgeEventLatencyMethod,
			ClusterKey:     appInstKey.ClusterInstKey.ClusterKey,
			ClusterInstOrg: appInstKey.ClusterInstKey.Organization,
		},
	}

	apiStat, found := dmecommon.Stats.LookupApiStatCall(apiStatCall)
	if !found {
		return "unable to find apiStat"
	}
	apiStat.Mux.Lock()
	defer apiStat.Mux.Unlock()

	latency := apiStat.RollingLatencyTemp.Latency
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

	b := []byte(req.Args)
	var appInstKey edgeproto.AppInstKey
	err := json.Unmarshal(b, &appInstKey)
	if err != nil {
		return nil, err
	}

	return &appInstKey, nil
}
