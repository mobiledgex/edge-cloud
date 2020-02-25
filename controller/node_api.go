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

func (s *NodeApi) ShowNodeLocal(in *edgeproto.Node, cb edgeproto.NodeApi_ShowNodeLocalServer) error {
	err := nodeMgr.NodeCache.Show(in, func(obj *edgeproto.Node) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *NodeApi) ShowNode(in *edgeproto.Node, cb edgeproto.NodeApi_ShowNodeServer) error {
	err := controllerApi.RunJobs(func(arg interface{}, addr string) error {
		if addr == *externalApiAddr {
			// local node
			return s.ShowNodeLocal(in, cb)
		}
		// connect to remote node
		conn, err := ControllerConnect(addr)
		if err != nil {
			return nil
		}
		defer conn.Close()

		cmd := edgeproto.NewNodeApiClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		stream, err := cmd.ShowNodeLocal(ctx, in)
		if err != nil {
			return err
		}
		for {
			obj, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("ShowNodeLocal %s: failed: %s",
					addr, err.Error())
			}
			err = cb.Send(obj)
			if err != nil {
				return err
			}
		}
		return nil
	}, nil)
	return err
}
