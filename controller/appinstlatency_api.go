package main

import (
	"context"
	"encoding/json"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc"
)

type AppInstLatencyApi struct {
	sync *Sync
}

var appInstLatencyApi = AppInstLatencyApi{}

func InitAppInstLatencyApi(sync *Sync) {
	appInstLatencyApi.sync = sync
}

type ControllerRunDebugServer struct {
	grpc.ServerStream
	ctx          context.Context
	ReplyHandler func(m *edgeproto.DebugReply) error // TODO: Rename SendHandler
}

func (c *ControllerRunDebugServer) Send(m *edgeproto.DebugReply) error {
	return c.ReplyHandler(m)
	// return c.ServerStream.SendMsg(m)
}

func (c *ControllerRunDebugServer) Context() context.Context {
	return c.ctx
}

func (s *AppInstLatencyApi) RequestAppInstLatency(ctx context.Context, in *edgeproto.AppInstLatency) (*edgeproto.Result, error) {
	// ctx := cb.Context()

	log.SpanLog(ctx, log.DebugLevelApi, "request app inst latency", "ctx", ctx)

	var err error
	/*err := in.Validate(edgeproto.AppInstAllFieldsMap)
	if err != nil {
		return nil, err
	}*/

	// populate the clusterinst developer from the app developer if not already present
	if in.Key.ClusterInstKey.Organization == "" {
		in.Key.ClusterInstKey.Organization = in.Key.AppKey.Organization
		//cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
	}
	// AppInst information to be parsed in dme-debug.go
	appKey := in.Key.AppKey
	clusterinstKey := in.Key.ClusterInstKey
	cloudletKey := clusterinstKey.CloudletKey
	clusterKey := in.Key.ClusterInstKey.ClusterKey
	args := appKey.Name + " " + appKey.Organization + " " + appKey.Version + " " + cloudletKey.Name + " " + cloudletKey.Organization + " " + clusterKey.Name + " " + clusterinstKey.Organization
	// Debug Request
	req := &edgeproto.DebugRequest{
		Node: edgeproto.NodeKey{
			// Name: "FHUANG-MAC.local",
			Type:        node.NodeTypeDME,
			CloudletKey: edgeproto.CloudletKey{
				// Name:         "mexdev-cloud-1",
				// Organization: "mexdev",
			},      //in.Key.ClusterInstKey.CloudletKey,
			Region: "local", //nodeMgr.Region, should be region of appinst
		},
		Cmd:  "request-appinst-latency", //node.RequestAppInstLatency,
		Args: args,
	}
	log.SpanLog(ctx, log.DebugLevelApi, "before run debug", "req", req)

	msg := ""
	replyHandler := func(m *edgeproto.DebugReply) error {
		log.SpanLog(ctx, log.DebugLevelApi, "Received run debug reply", "reply", m)
		msg = m.Output
		return nil
	}

	newcb := &ControllerRunDebugServer{}
	newcb.ReplyHandler = replyHandler
	newcb.ctx = ctx

	// err = debugApi.RunDebug(req, newcb)
	err = nodeMgr.Debug.DebugRequest(req, newcb)
	if err != nil {
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "received debug output", "msg", msg)
	return &edgeproto.Result{Message: msg}, err
}

func (s *AppInstLatencyApi) ShowAppInstLatency(in *edgeproto.AppInstLatency, cb edgeproto.AppInstLatencyApi_ShowAppInstLatencyServer) error {
	ctx := cb.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "request app inst latency", "ctx", ctx)

	var err error
	/*err := in.Validate(edgeproto.AppInstAllFieldsMap)
	if err != nil {
		return err
	}*/

	// populate the clusterinst developer from the app developer if not already present
	if in.Key.ClusterInstKey.Organization == "" {
		in.Key.ClusterInstKey.Organization = in.Key.AppKey.Organization
		//cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
	}
	// AppInst information to be parsed in dme-debug.go
	appKey := in.Key.AppKey
	clusterinstKey := in.Key.ClusterInstKey
	cloudletKey := clusterinstKey.CloudletKey
	clusterKey := in.Key.ClusterInstKey.ClusterKey
	args := appKey.Name + " " + appKey.Organization + " " + appKey.Version + " " + cloudletKey.Name + " " + cloudletKey.Organization + " " + clusterKey.Name + " " + clusterinstKey.Organization
	// Debug Request
	req := &edgeproto.DebugRequest{
		Node: edgeproto.NodeKey{
			Type:        node.NodeTypeDME,
			CloudletKey: edgeproto.CloudletKey{
				// Name:         "mexdev-cloud-1",
				// Organization: "mexdev",
			},      //in.Key.ClusterInstKey.CloudletKey,
			Region: "local", //nodeMgr.Region, should be region of appinst
		},
		Cmd:  "show-appinst-latency", //node.ShowAppInstLatency,
		Args: args,
	}
	log.SpanLog(ctx, log.DebugLevelApi, "before run debug", "req", req)

	msg := ""
	// TODO: BETTER STREAM HANDLING (ERRORS)
	replyHandler := func(m *edgeproto.DebugReply) error {
		log.SpanLog(ctx, log.DebugLevelApi, "Received run debug reply", "reply", m)
		msg = m.Output

		// Unmarshal
		b := []byte(msg)
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
		cb.Send(appInstLatency)
		return nil
	}

	newcb := &ControllerRunDebugServer{}
	newcb.ReplyHandler = replyHandler
	newcb.ctx = ctx

	// err = debugApi.RunDebug(req, newcb)
	err = nodeMgr.Debug.DebugRequest(req, newcb)
	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelApi, "received debug output", "msg", msg)
	return err
}
