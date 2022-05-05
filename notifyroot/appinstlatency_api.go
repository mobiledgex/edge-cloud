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

package main

import (
	"context"
	"encoding/json"

	"github.com/edgexr/edge-cloud/cloudcommon/node"
	dmecommon "github.com/edgexr/edge-cloud/d-match-engine/dme-common"
	"github.com/edgexr/edge-cloud/edgeproto"
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
