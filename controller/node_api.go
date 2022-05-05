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
	"time"

	"github.com/edgexr/edge-cloud/edgeproto"
)

// Nodes are not stored in the etcd database, because they
// are dynamic. DMEs will be instantiated as load requires on Cloudlets.
// Instead, connected DMEs are tracked independently by each controller.
// To get a list of all DMEs, we query each controller and get each
// one's list of connected DMEs/CRMs.

type NodeApi struct{}

var nodeApi = NodeApi{}

func (s *NodeApi) ShowNode(in *edgeproto.Node, cb edgeproto.NodeApi_ShowNodeServer) error {
	if *notifyRootAddrs == "" && *notifyParentAddrs == "" {
		// assume this is the root
		return nodeMgr.NodeCache.Show(in, func(obj *edgeproto.Node) error {
			err := cb.Send(obj)
			return err
		})
	}

	// ShowNode should directly communicate with NotifyRoot and not via MC
	notifyAddrs := *notifyRootAddrs
	if notifyAddrs == "" {
		// In case notifyrootaddrs is not specified,
		// fallback to notifyparentaddrs
		notifyAddrs = *notifyParentAddrs
	}

	conn, err := notifyRootConnect(cb.Context(), notifyAddrs)
	if err != nil {
		return err
	}
	client := edgeproto.NewNodeApiClient(conn)
	ctx, cancel := context.WithTimeout(cb.Context(), 3*time.Second)
	defer cancel()

	stream, err := client.ShowNode(ctx, in)
	if err != nil {
		return err
	}
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ShowNode failed, %v", err)
		}
		err = cb.Send(obj)
		if err != nil {
			return err
		}
	}
	return nil
}
