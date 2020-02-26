package node

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/version"
)

var NodeTypeCRM = "crm"
var NodeTypeDME = "dme"
var NodeTypeController = "controller"
var NodeTypeClusterSvc = "cluster-svc"

// Node tracks all the nodes connected via notify, and handles common
// requests over all nodes.
type NodeMgr struct {
	MyNode    edgeproto.Node
	NodeCache edgeproto.NodeCache
}

func Init(ctx context.Context, nodeType string, ops ...NodeOp) *NodeMgr {
	opts := &NodeOptions{}
	opts.updateMyNode = true
	for _, op := range ops {
		op(opts)
	}
	s := NodeMgr{}
	s.MyNode.Key.Type = nodeType
	if opts.name != "" {
		s.MyNode.Key.Name = opts.name
	} else {
		s.MyNode.Key.Name = cloudcommon.Hostname()
	}
	s.MyNode.Key.CloudletKey = opts.cloudletKey
	s.MyNode.BuildMaster = version.BuildMaster
	s.MyNode.BuildHead = version.BuildHead
	s.MyNode.BuildAuthor = version.BuildAuthor
	s.MyNode.Hostname = cloudcommon.Hostname()
	s.MyNode.ContainerVersion = opts.containerVersion

	edgeproto.InitNodeCache(&s.NodeCache)
	if opts.updateMyNode {
		s.UpdateMyNode(ctx)
	}
	return &s
}

type NodeOptions struct {
	name             string
	cloudletKey      edgeproto.CloudletKey
	updateMyNode     bool
	containerVersion string
}

type NodeOp func(s *NodeOptions)

func WithName(name string) NodeOp {
	return func(opts *NodeOptions) { opts.name = name }
}

func WithCloudletKey(key *edgeproto.CloudletKey) NodeOp {
	return func(opts *NodeOptions) { opts.cloudletKey = *key }
}

func WithNoUpdateMyNode() NodeOp {
	return func(opts *NodeOptions) { opts.updateMyNode = false }
}

func WithContainerVersion(ver string) NodeOp {
	return func(opts *NodeOptions) { opts.containerVersion = ver }
}

func (s *NodeMgr) UpdateMyNode(ctx context.Context) {
	s.NodeCache.Update(ctx, &s.MyNode, 0)
}

func (s *NodeMgr) RegisterClient(client *notify.Client) {
	client.RegisterSendNodeCache(&s.NodeCache)
}

func (s *NodeMgr) RegisterServer(server *notify.ServerMgr) {
	server.RegisterRecvNodeCache(&s.NodeCache)
}
