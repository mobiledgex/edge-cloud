package main

import "github.com/mobiledgex/edge-cloud/edgeproto"

type NodeApi struct{}

var nodeApi = NodeApi{}

func (s *NodeApi) ShowNode(in *edgeproto.Node, cb edgeproto.NodeApi_ShowNodeServer) error {
	err := nodeMgr.NodeCache.Show(in, func(obj *edgeproto.Node) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
