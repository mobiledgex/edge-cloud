package node

import (
	"context"
	"flag"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/version"
	"github.com/opentracing/opentracing-go"
)

var NodeTypeCRM = "crm"
var NodeTypeDME = "dme"
var NodeTypeController = "controller"
var NodeTypeClusterSvc = "cluster-svc"
var NodeTypeNotifyRoot = "notifyroot"
var NodeTypeEdgeTurn = "edgeturn"
var NodeTypeMC = "mc"

// Node tracks all the nodes connected via notify, and handles common
// requests over all nodes.
type NodeMgr struct {
	iTlsCertFile string
	iTlsKeyFile  string
	iTlsCAFile   string
	VaultAddr    string

	MyNode             edgeproto.Node
	NodeCache          RegionNodeCache
	Debug              DebugNode
	VaultConfig        *vault.Config
	Region             string
	InternalPki        internalPki
	InternalDomain     string
	ESClient           *elasticsearch.Client
	tlsClientIssuer    string
	commonName         string
	DeploymentTag      string
	AccessKeyClient    AccessKeyClient
	CloudletPoolLookup CloudletPoolLookup

	unitTestMode bool
}

// Most of the time there will only be one NodeMgr per process, and these
// settings will come from command line input.
func (s *NodeMgr) InitFlags() {
	// itls uses a set of file-based certs for internal mTLS auth
	// between services. It is not production-safe and should only be
	// used if Vault-PKI cannot be used.
	flag.StringVar(&s.iTlsCertFile, "itlsCert", "", "internal mTLS cert file for communication between services")
	flag.StringVar(&s.iTlsKeyFile, "itlsKey", "", "internal mTLS key file for communication between services")
	flag.StringVar(&s.iTlsCAFile, "itlsCA", "", "internal mTLS CA file for communication between servcies")
	flag.StringVar(&s.VaultAddr, "vaultAddr", "", "Vault address; local vault runs at http://127.0.0.1:8200")
	flag.BoolVar(&s.InternalPki.UseVaultCAs, "useVaultCAs", false, "Include use of Vault CAs for internal mTLS authentication")
	flag.BoolVar(&s.InternalPki.UseVaultCerts, "useVaultCerts", false, "Use Vault Certs for internal mTLS; implies useVaultCAs")
	flag.StringVar(&s.InternalDomain, "internalDomain", "mobiledgex.net", "domain name for internal PKI")
	flag.StringVar(&s.commonName, "commonName", "", "common name to use for vault internal pki issued certificates")
	flag.StringVar(&s.DeploymentTag, "deploymentTag", "", "Tag to indicate type of deployment setup. Ex: production, staging, etc")
}

func (s *NodeMgr) Init(nodeType, tlsClientIssuer string, ops ...NodeOp) (context.Context, opentracing.Span, error) {
	initCtx := log.ContextWithSpan(context.Background(), log.NoTracingSpan())
	log.SpanLog(initCtx, log.DebugLevelInfo, "start main nodeMgr init")

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
	s.MyNode.BuildDate = version.BuildDate
	s.MyNode.Hostname = cloudcommon.Hostname()
	s.MyNode.ContainerVersion = opts.containerVersion
	s.Region = opts.region
	s.tlsClientIssuer = tlsClientIssuer
	s.CloudletPoolLookup = opts.cloudletPoolLookup

	if err := s.AccessKeyClient.init(initCtx, nodeType, tlsClientIssuer, opts.cloudletKey, s.DeploymentTag); err != nil {
		log.SpanLog(initCtx, log.DebugLevelInfo, "access key client init failed", "err", err)
		return initCtx, nil, err
	}
	if s.AccessKeyClient.enabled {
		// no vault, Controller replaces Vault for issuing certs
	} else {
		// init vault before pki
		s.VaultConfig = opts.vaultConfig
		if s.VaultConfig == nil {
			var err error
			s.VaultConfig, err = vault.BestConfig(s.VaultAddr)
			if err != nil {
				return initCtx, nil, err
			}
			log.SpanLog(initCtx, log.DebugLevelInfo, "vault auth", "type", s.VaultConfig.Auth.Type())
		}
	}

	// init pki before logging, because access to logger needs pki certs
	log.SpanLog(initCtx, log.DebugLevelInfo, "init internal pki")
	err := s.initInternalPki(initCtx)
	if err != nil {
		return initCtx, nil, err
	}

	// init logger
	log.SpanLog(initCtx, log.DebugLevelInfo, "get logger tls")
	loggerTls, err := s.GetPublicClientTlsConfig(initCtx)
	if err != nil {
		return initCtx, nil, err
	}
	log.InitTracer(loggerTls)

	// logging is initialized so start the real span
	// nodemgr init should always be started from main.
	// Caller needs to handle span.Finish()
	var span opentracing.Span
	if opts.parentSpan != "" {
		span = log.NewSpanFromString(log.DebugLevelInfo, opts.parentSpan, "main")
	} else {
		span = log.StartSpan(log.DebugLevelInfo, "main")
	}
	ctx := log.ContextWithSpan(context.Background(), span)

	// start pki refresh after logging initialized
	s.InternalPki.start()

	if s.CloudletPoolLookup == nil {
		// single region lookup for events
		lookup := &CloudletPoolCache{}
		lookup.Init()
		s.CloudletPoolLookup = lookup
	}
	err = s.initEvents(ctx, opts)
	if err != nil {
		span.Finish()
		return initCtx, nil, err
	}

	edgeproto.InitNodeCache(&s.NodeCache.NodeCache)
	s.NodeCache.setRegion = opts.region
	s.Debug.Init(s)
	if opts.updateMyNode {
		s.UpdateMyNode(ctx)
	}
	return ctx, span, nil
}

func (s *NodeMgr) Name() string {
	return s.MyNode.Key.Name
}

func (s *NodeMgr) Finish() {
	log.FinishTracer()
}

func (s *NodeMgr) CommonName() string {
	if s.commonName != "" {
		return s.commonName
	}
	cn := s.MyNode.Key.Type
	if cn == NodeTypeController {
		cn = "ctrl"
	}
	return cn + "." + s.InternalDomain
}

type NodeOptions struct {
	name               string
	cloudletKey        edgeproto.CloudletKey
	updateMyNode       bool
	containerVersion   string
	region             string
	vaultConfig        *vault.Config
	esUrls             string
	parentSpan         string
	cloudletPoolLookup CloudletPoolLookup
}

type CloudletInPoolFunc func(region, key edgeproto.CloudletKey) bool

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

func WithESUrls(urls string) NodeOp {
	return func(opts *NodeOptions) { opts.esUrls = urls }
}

func WithParentSpan(parentSpan string) NodeOp {
	return func(opts *NodeOptions) { opts.parentSpan = parentSpan }
}

func WithCloudletPoolLookup(cloudletPoolLookup CloudletPoolLookup) NodeOp {
	return func(opts *NodeOptions) { opts.cloudletPoolLookup = cloudletPoolLookup }
}

func (s *NodeMgr) UpdateMyNode(ctx context.Context) {
	s.NodeCache.Update(ctx, &s.MyNode, 0)
}

func (s *NodeMgr) RegisterClient(client *notify.Client) {
	client.RegisterSendNodeCache(&s.NodeCache)
	s.Debug.RegisterClient(client)
	// MC notify handling of CloudletPoolCache is done outside of nodemgr.
	if s.MyNode.Key.Type != NodeTypeMC && s.MyNode.Key.Type != NodeTypeNotifyRoot && s.MyNode.Key.Type != NodeTypeController {
		cache := s.CloudletPoolLookup.GetCloudletPoolCache(s.Region)
		client.RegisterRecvCloudletPoolCache(cache)
	}
}

func (s *NodeMgr) RegisterServer(server *notify.ServerMgr) {
	server.RegisterRecvNodeCache(&s.NodeCache)
	s.Debug.RegisterServer(server)
	// MC notify handling of CloudletPoolCache is done outside of nodemgr.
	if s.MyNode.Key.Type != NodeTypeMC && s.MyNode.Key.Type != NodeTypeNotifyRoot {
		cache := s.CloudletPoolLookup.GetCloudletPoolCache(s.Region)
		server.RegisterSendCloudletPoolCache(cache)
	}
}

func (s *NodeMgr) GetInternalTlsCertFile() string {
	return s.iTlsCertFile
}

func (s *NodeMgr) GetInternalTlsKeyFile() string {
	return s.iTlsKeyFile
}

func (s *NodeMgr) GetInternalTlsCAFile() string {
	return s.iTlsCAFile
}

// setters are only used for unit testing
func (s *NodeMgr) SetInternalTlsCertFile(file string) {
	s.iTlsCertFile = file
}

func (s *NodeMgr) SetInternalTlsKeyFile(file string) {
	s.iTlsKeyFile = file
}

func (s *NodeMgr) SetInternalTlsCAFile(file string) {
	s.iTlsCAFile = file
}
