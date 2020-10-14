package main

import (
	"context"
	"encoding/json"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmeutil "github.com/mobiledgex/edge-cloud/d-match-engine/dme-util"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type AppInstLatencyApi struct {
	sync *Sync
}

var appInstLatencyApi = AppInstLatencyApi{}

func InitAppInstLatencyApi(sync *Sync) {
	appInstLatencyApi.sync = sync
}

func (s *AppInstLatencyApi) RequestAppInstLatency(ctx context.Context, in *edgeproto.AppInstLatency) (*edgeproto.Result, error) {
	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}
	// populate the clusterinst developer from the app developer if not already present
	if in.Key.ClusterInstKey.Organization == "" {
		in.Key.ClusterInstKey.Organization = in.Key.AppKey.Organization
	}
	// AppInst information to be parsed in dme-debug.go
	appKey := in.Key.AppKey
	clusterinstKey := in.Key.ClusterInstKey
	cloudletKey := clusterinstKey.CloudletKey
	clusterKey := in.Key.ClusterInstKey.ClusterKey
	args := appKey.Name + " " + appKey.Organization + " " + appKey.Version + " " + cloudletKey.Name + " " + cloudletKey.Organization + " " + clusterKey.Name + " " + clusterinstKey.Organization
	// Create Debug Request
	req := &edgeproto.DebugRequest{
		Node: edgeproto.NodeKey{
			Type: node.NodeTypeDME,
			// CloudletKey: in.Key.ClusterInstKey.CloudletKey, // When DME per cloudlet is implemented
			Region: nodeMgr.Region, // should be region of appinst
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
	newcb := &ControllerRunDebugServer{
		ReplyHandler: replyHandler,
		ctx:          ctx,
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
	// AppInst information to be parsed in dme-debug.go
	appKey := in.Key.AppKey
	clusterinstKey := in.Key.ClusterInstKey
	cloudletKey := clusterinstKey.CloudletKey
	clusterKey := in.Key.ClusterInstKey.ClusterKey
	args := appKey.Name + " " + appKey.Organization + " " + appKey.Version + " " + cloudletKey.Name + " " + cloudletKey.Organization + " " + clusterKey.Name + " " + clusterinstKey.Organization
	// Create Debug Request
	req := &edgeproto.DebugRequest{
		Node: edgeproto.NodeKey{
			Type: node.NodeTypeDME,
			// CloudletKey: in.Key.ClusterInstKey.CloudletKey, // When DME per cloudlet is implemented
			Region: nodeMgr.Region, // should be region of appinst
		},
		Cmd:  dmeutil.ShowAppInstLatency,
		Args: args,
	}

	// Create function to handle DebugReply
	replyHandler := func(m *edgeproto.DebugReply) error {
		// Unmarshal
		b := []byte(m.Output)
		var latency dme.Latency
		err := json.Unmarshal(b, &latency)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Unable to unmarshal DebugReply to latency")
			return err
		}
		appInstLatency := &edgeproto.AppInstLatency{
			Key:     in.Key,
			Latency: &latency,
		}
		return cb.Send(appInstLatency)
	}

	newcb := &ControllerRunDebugServer{
		ReplyHandler: replyHandler,
		ctx:          ctx,
	}
	err = nodeMgr.Debug.DebugRequest(req, newcb)
	return err
}
