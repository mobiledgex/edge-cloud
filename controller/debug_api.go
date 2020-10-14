package main

import (
	"context"
	"fmt"
	"io"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"google.golang.org/grpc"
)

type DebugApi struct{}

var debugApi = DebugApi{}

func (s *DebugApi) EnableDebugLevels(req *edgeproto.DebugRequest, cb edgeproto.DebugApi_EnableDebugLevelsServer) error {
	req.Cmd = node.EnableDebugLevels
	return s.RunDebug(req, cb)
}

func (s *DebugApi) DisableDebugLevels(req *edgeproto.DebugRequest, cb edgeproto.DebugApi_DisableDebugLevelsServer) error {
	req.Cmd = node.DisableDebugLevels
	return s.RunDebug(req, cb)
}

func (s *DebugApi) ShowDebugLevels(req *edgeproto.DebugRequest, cb edgeproto.DebugApi_ShowDebugLevelsServer) error {
	req.Cmd = node.ShowDebugLevels
	return s.RunDebug(req, cb)
}

func (s *DebugApi) RunDebug(req *edgeproto.DebugRequest, cb edgeproto.DebugApi_RunDebugServer) error {
	if *notifyParentAddrs == "" {
		// assume this is the root
		return nodeMgr.Debug.DebugRequest(req, cb)
	}

	conn, err := notifyRootConnect(cb.Context(), *notifyParentAddrs)
	if err != nil {
		return err
	}
	client := edgeproto.NewDebugApiClient(conn)
	if req.Timeout == 0 {
		req.Timeout = edgeproto.Duration(node.DefaultDebugTimeout)
	}
	ctx, cancel := context.WithTimeout(cb.Context(), req.Timeout.TimeDuration())
	defer cancel()

	stream, err := client.RunDebug(ctx, req)
	if err != nil {
		return err
	}
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("RunDebug failed, %v", err)
		}
		err = cb.Send(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

// Struct that allows use of Debug framework from controller
// Supply a ReplyHandler to handle the DebugReply
// See appinstlatency_api.go for example use
type ControllerRunDebugServer struct {
	grpc.ServerStream
	ctx          context.Context
	ReplyHandler func(m *edgeproto.DebugReply) error
}

func (c *ControllerRunDebugServer) Send(m *edgeproto.DebugReply) error {
	return c.ReplyHandler(m)
}

func (c *ControllerRunDebugServer) Context() context.Context {
	return c.ctx
}
