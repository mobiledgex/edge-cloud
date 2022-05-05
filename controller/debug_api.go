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
	"fmt"
	"io"

	"github.com/edgexr/edge-cloud/cloudcommon/node"
	"github.com/edgexr/edge-cloud/edgeproto"
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
