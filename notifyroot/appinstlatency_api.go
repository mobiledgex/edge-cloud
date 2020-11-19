package main

import (
	"context"
	"encoding/json"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dmeutil "github.com/mobiledgex/edge-cloud/d-match-engine/dme-util"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstLatencyApi struct{}

var appInstLatencyApi = AppInstLatencyApi{}

func (s *AppInstLatencyApi) RequestAppInstLatency(ctx context.Context, in *edgeproto.AppInstLatency) (*edgeproto.Result, error) {
	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}
	// populate the clusterinst developer from the app developer if not already present
	if in.Key.ClusterInstKey.Organization == "" {
		in.Key.ClusterInstKey.Organization = in.Key.AppKey.Organization
	}
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
			// CloudletKey: in.Key.ClusterInstKey.CloudletKey, // When DME per cloudlet is implemented
			// Region: nodeMgr.Region, // should be region of appinst
		},
		Cmd:  dmeutil.RequestAppInstLatency,
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
