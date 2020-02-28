package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Nodes are not stored in the etcd database, because they
// are dynamic. DMEs will be instantiated as load requires on Cloudlets.
// Instead, connected DMEs are tracked independently by each controller.
// To get a list of all DMEs, we query each controller and get each
// one's list of connected DMEs/CRMs.

type NodeApi struct{}

var nodeApi = NodeApi{}

func (s *NodeApi) ShowNode(in *edgeproto.Node, cb edgeproto.NodeApi_ShowNodeServer) error {
	if *notifyParentAddrs == "" {
		// assume this is the root
		return nodeMgr.NodeCache.Show(in, func(obj *edgeproto.Node) error {
			err := cb.Send(obj)
			return err
		})
	}

	conn, err := notifyRootConnect()
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
