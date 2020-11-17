package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dmeutil "github.com/mobiledgex/edge-cloud/d-match-engine/dme-util"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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

func (s *AppInstLatencyApi) ShowAppInstLatency(in *edgeproto.AppInstLatency, cb edgeproto.AppInstLatencyApi_ShowAppInstLatencyServer) error {
	ctx := cb.Context()
	err := in.Key.ValidateKey()
	if err != nil {
		return err
	}

	// populate the clusterinst developer from the app developer if not already present
	if in.Key.ClusterInstKey.Organization == "" {
		in.Key.ClusterInstKey.Organization = in.Key.AppKey.Organization
		cb.Send(&edgeproto.AppInstLatency{Message: "Setting ClusterInst developer to match App developer"})
	}
	// Create args for DebugRequest (AppInstKey)
	b, err := json.Marshal(in.Key)
	if err != nil {
		return err
	}
	args := string(b)
	// Create Debug Request
	req := &edgeproto.DebugRequest{
		Node: edgeproto.NodeKey{
			Type: node.NodeTypeDME,
			// CloudletKey: in.Key.ClusterInstKey.CloudletKey, // When DME per cloudlet is implemented
			// Region: nodeMgr.Region, // should be region of appinst
		},
		Cmd:  dmeutil.ShowAppInstLatency,
		Args: args,
	}

	// Create function to handle DebugReply
	replyHandler := func(m *edgeproto.DebugReply) error {
		// Unmarshal
		b := []byte(m.Output)
		var latencyStats dmeutil.AppInstLatencyStats
		err := json.Unmarshal(b, &latencyStats)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Unable to unmarshal DebugReply to RollingLatency")
			return err
		}
		appInstLatency := &edgeproto.AppInstLatency{
			Key: in.Key,
		}
		return cb.Send(appInstLatency)
	}

	newcb := &node.RunDebugServer{
		ReplyHandler: replyHandler,
		Ctx:          ctx,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err = nodeMgr.Debug.DebugRequest(req, newcb); err != nil {
			return err
		}
		time.Sleep(10 * time.Second)
	}
}
