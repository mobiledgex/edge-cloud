package node

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"google.golang.org/grpc"
)

// Struct that allows use of Debug framework from any node
// Supply a ReplyHandler to handle the DebugReply
// See notifyroot/appinstlatency_api.go for example use
type RunDebugServer struct {
	grpc.ServerStream
	Ctx          context.Context
	ReplyHandler func(m *edgeproto.DebugReply) error
}

func (c *RunDebugServer) Send(m *edgeproto.DebugReply) error {
	return c.ReplyHandler(m)
}

func (c *RunDebugServer) Context() context.Context {
	return c.Ctx
}
