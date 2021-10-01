package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func InitDebug(nodeMgr *node.NodeMgr) {
	nodeMgr.Debug.AddDebugFunc(dmecommon.RequestAppInstLatency, requestAppInstLatency)
	nodeMgr.Debug.AddDebugFunc("spew-rate-limit-mgr", spewRateLimitMgr)
}

func requestAppInstLatency(ctx context.Context, req *edgeproto.DebugRequest) string {
	appInstKey, err := createAppInstKeyFromRequest(req)
	if err != nil {
		return err.Error()
	}

	dmecommon.EEHandler.SendLatencyRequestEdgeEvent(ctx, *appInstKey)
	return "successfully sent latency request"
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

func spewRateLimitMgr(ctx context.Context, req *edgeproto.DebugRequest) string {
	if dmecommon.RateLimitMgr == nil {
		return "nil"
	}
	dmecommon.RateLimitMgr.Lock()
	defer dmecommon.RateLimitMgr.Unlock()
	return spew.Sdump(dmecommon.RateLimitMgr)
}
