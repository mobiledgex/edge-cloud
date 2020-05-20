package node

import (
	"context"
	"flag"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/version"
)

var NodeTypeCRM = "crm"
var NodeTypeDME = "dme"
var NodeTypeController = "controller"
var NodeTypeClusterSvc = "cluster-svc"
var NodeTypeNotifyRoot = "notifyroot"
var NodeTypeEdgeTurn = "edgeturn"

// Node tracks all the nodes connected via notify, and handles common
// requests over all nodes.
type NodeMgr struct {
	TlsCertFile string
	VaultAddr   string

	MyNode         edgeproto.Node
	NodeCache      RegionNodeCache
	Debug          DebugNode
	VaultConfig    *vault.Config
	Region         string
	InternalPki    internalPki
	InternalDomain string
}

// Most of the time there will only be one NodeMgr per process, and these
// settings will come from command line input.
func (s *NodeMgr) InitFlags() {
	// TlsCertFile remains for backwards compatibility. It will eventually be
	// removed once all CRMs transition over to Vault internal PKI.
	flag.StringVar(&s.TlsCertFile, "tls", "", "server tls cert file. Keyfile and CA file must be in same directory, CA file should be \"mex-ca.crt\", and key file should be same name as cert file but extension \".key\"")
	flag.StringVar(&s.VaultAddr, "vaultAddr", "", "Vault address; local vault runs at http://127.0.0.1:8200")
	flag.BoolVar(&s.InternalPki.UseVaultCAs, "useVaultCAs", false, "Include use of Vault CAs for internal TLS authentication")
	flag.BoolVar(&s.InternalPki.UseVaultCerts, "useVaultCerts", false, "Use Vault Certs for internal TLS; implies useVaultCAs")
	flag.StringVar(&s.InternalDomain, "internalDomain", "mobiledgex.net", "domain name for internal PKI")
}

func (s *NodeMgr) Init(ctx context.Context, nodeType string, ops ...NodeOp) error {
	opts := &NodeOptions{}
	opts.updateMyNode = true
	for _, op := range ops {
		op(opts)
	}
	s.MyNode.Key.Type = nodeType
	if opts.name != "" {
		s.MyNode.Key.Name = opts.name
	} else {
		s.MyNode.Key.Name = cloudcommon.Hostname()
	}
	s.MyNode.Key.Region = opts.region
	s.MyNode.Key.CloudletKey = opts.cloudletKey
	s.MyNode.BuildMaster = version.BuildMaster
	s.MyNode.BuildHead = version.BuildHead
	s.MyNode.BuildAuthor = version.BuildAuthor
	s.MyNode.Hostname = cloudcommon.Hostname()
	s.MyNode.ContainerVersion = opts.containerVersion
	s.Region = opts.region

	// init vault before pki
	s.VaultConfig = opts.vaultConfig
	if s.VaultConfig == nil {
		var err error
		s.VaultConfig, err = vault.BestConfig(s.VaultAddr)
		if err != nil {
			return err
		}
		log.SpanLog(ctx, log.DebugLevelInfo, "vault auth", "type", s.VaultConfig.Auth.Type())
	}

	err := s.initInternalPki(ctx)
	if err != nil {
		return err
	}

	edgeproto.InitNodeCache(&s.NodeCache.NodeCache)
	s.NodeCache.setRegion = opts.region
	s.Debug.Init(s)

	if opts.updateMyNode {
		s.UpdateMyNode(ctx)
	}
	return nil
}

type NodeOptions struct {
	name             string
	cloudletKey      edgeproto.CloudletKey
	updateMyNode     bool
	containerVersion string
	region           string
	vaultConfig      *vault.Config
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

func WithRegion(region string) NodeOp {
	return func(opts *NodeOptions) { opts.region = region }
}

func WithVaultConfig(vaultConfig *vault.Config) NodeOp {
	return func(opts *NodeOptions) { opts.vaultConfig = vaultConfig }
}

func (s *NodeMgr) UpdateMyNode(ctx context.Context) {
	s.NodeCache.Update(ctx, &s.MyNode, 0)
}

func (s *NodeMgr) RegisterClient(client *notify.Client) {
	client.RegisterSendNodeCache(&s.NodeCache)
	s.Debug.RegisterClient(client)
}

func (s *NodeMgr) RegisterServer(server *notify.ServerMgr) {
	server.RegisterRecvNodeCache(&s.NodeCache)
	s.Debug.RegisterServer(server)
}
