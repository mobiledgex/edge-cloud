// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package node

import (
	"context"

	"github.com/edgexr/edge-cloud/edgeproto"
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
