package main

import (
	"context"
	"encoding/json"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstLatencyApi struct{}

var appInstLatencyApi = AppInstLatencyApi{}

func (s *AppInstLatencyApi) RequestAppInstLatency(ctx context.Context, in *edgeproto.AppInstLatency) (*edgeproto.Result, error) {
	// Create args for DebugRequest (AppInstKey)
	b, err := json.Marshal(in.Key)
	if err != nil {
		return nil, err
	}
	args := string(b)
	// Create Debug Request
	req := &edgeproto.DebugRequest{
		Node: edgeproto.NodeKey{
			Type: node.NodeTypeDME,
		},
		Cmd:  dmecommon.RequestAppInstLatency,
		Args: args,
	}

	msg := ""
	// Create function to handle DebugReply
	replyHandler := func(m *edgeproto.DebugReply) error {
		msg = m.Output
		return nil
	}
	// Initialize ControllerRunDebugServer will DebugReply handler to be called in Send
	newcb := &node.RunDebugServer{
		ReplyHandler: replyHandler,
		Ctx:          ctx,
	}
	err = nodeMgr.Debug.DebugRequest(req, newcb)
	if err != nil {
		return nil, err
	}
	return &edgeproto.Result{Message: msg}, err
}
