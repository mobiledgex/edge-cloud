package node_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	edgetls "github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/features/proto/echo"
	"google.golang.org/grpc/grpclog"
)

// Note file package is not node, so avoids node package having
// dependencies on process package.

func TestInternalPki(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	// grcp logs not showing up in unit tests for some reason.
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, os.Stderr))
	// Set up local Vault process.
	// Note that this test depends on the approles and
	// pki configuration done by the vault setup scripts
	// that are run as part of running this Vault process.
	vp := process.Vault{
		Common: process.Common{
			Name: "vault",
		},
		Regions: "us,eu",
	}
	vroles, err := vp.StartLocalRoles()
	require.Nil(t, err, "start local vault")
	defer vp.StopLocal()

	node.BadAuthDelay = time.Millisecond
	node.VerifyDelay = time.Millisecond
	node.VerifyRetry = 3

	vaultAddr := "http://127.0.0.1:8200"
	// Set up fake Controller to serve access key API
	dcUS := &DummyController{}
	dcUS.Init(ctx, "us", vroles, vaultAddr)
	dcUS.Start(ctx)
	defer dcUS.Stop()

	// Set up fake Controller to serve access key API
	dcEU := &DummyController{}
	dcEU.Init(ctx, "eu", vroles, vaultAddr)
	dcEU.Start(ctx)
	defer dcEU.Stop()

	// create access key for US cloudlet
	edgeboxCloudlet := true
	tc1 := dcUS.CreateCloudlet(ctx, "pkitc1", !edgeboxCloudlet)
	err = dcUS.UpdateKey(ctx, tc1.Cloudlet.Key)
	require.Nil(t, err)
	// create access key for EU cloudlet
	tc2 := dcEU.CreateCloudlet(ctx, "pkitc2", !edgeboxCloudlet)
	err = dcEU.UpdateKey(ctx, tc2.Cloudlet.Key)
	require.Nil(t, err)

	// Most positive testing is done by e2e tests.

	// Negative testing for issuing certs.
	// These primarily test Vault certificate role permissions,
	// so work in conjunction with the vault setup in vault/setup-region.sh
	// Apparently CA certs are always readable from Vault approles.
	var cfgTests cfgTestList
	// regional Controller cannot issue global cert
	cfgTests.add(ConfigTest{
		NodeType:    node.NodeTypeController,
		Region:      "us",
		LocalIssuer: node.CertIssuerRegional,
		TestIssuer:  node.CertIssuerGlobal,
		ExpectErr:   "write failure pki-global/issue/us",
	})
	// regional Controller can issue RegionalCloudlet, for access-key services.
	cfgTests.add(ConfigTest{
		NodeType:    node.NodeTypeController,
		Region:      "us",
		LocalIssuer: node.CertIssuerRegional,
		TestIssuer:  node.CertIssuerRegionalCloudlet,
		ExpectErr:   "",
	})
	// global node cannot issue regional cert
	cfgTests.add(ConfigTest{
		NodeType:    node.NodeTypeNotifyRoot,
		LocalIssuer: node.CertIssuerGlobal,
		TestIssuer:  node.CertIssuerRegional,
		ExpectErr:   "write failure pki-regional/issue/default",
	})
	// global node cannot issue regional-cloudlet cert
	cfgTests.add(ConfigTest{
		NodeType:    node.NodeTypeNotifyRoot,
		LocalIssuer: node.CertIssuerGlobal,
		TestIssuer:  node.CertIssuerRegionalCloudlet,
		ExpectErr:   "write failure pki-regional-cloudlet/issue/default",
	})
	// cloudlet node cannot issue global cert
	cfgTests.add(ConfigTest{
		NodeType:      node.NodeTypeCRM,
		Region:        "us",
		LocalIssuer:   node.CertIssuerRegionalCloudlet,
		TestIssuer:    node.CertIssuerGlobal,
		AccessKeyFile: tc1.KeyClient.AccessKeyFile,
		AccessApiAddr: tc1.KeyClient.AccessApiAddr,
		CloudletKey:   &tc1.Cloudlet.Key,
		ExpectErr:     "Controller will only issue RegionalCloudlet certs",
	})
	// cloudlet node cannot issue regional cert
	cfgTests.add(ConfigTest{
		NodeType:      node.NodeTypeCRM,
		Region:        "us",
		LocalIssuer:   node.CertIssuerRegionalCloudlet,
		TestIssuer:    node.CertIssuerRegional,
		AccessKeyFile: tc1.KeyClient.AccessKeyFile,
		AccessApiAddr: tc1.KeyClient.AccessApiAddr,
		CloudletKey:   &tc1.Cloudlet.Key,
		ExpectErr:     "Controller will only issue RegionalCloudlet certs",
	})
	// cloudlet node can issue RegionalCloudlet cert
	cfgTests.add(ConfigTest{
		NodeType:      node.NodeTypeCRM,
		Region:        "us",
		LocalIssuer:   node.CertIssuerRegionalCloudlet,
		TestIssuer:    node.CertIssuerRegionalCloudlet,
		AccessKeyFile: tc1.KeyClient.AccessKeyFile,
		AccessApiAddr: tc1.KeyClient.AccessApiAddr,
		CloudletKey:   &tc1.Cloudlet.Key,
		ExpectErr:     "",
	})

	for _, cfg := range cfgTests {
		testGetTlsConfig(t, ctx, vroles, &cfg)
	}

	// define nodes for certificate exchange tests
	notifyRootServer := &PkiConfig{
		Type:        node.NodeTypeNotifyRoot,
		LocalIssuer: node.CertIssuerGlobal,
		UseVaultPki: true,
		RemoteCAs: []node.MatchCA{
			node.AnyRegionalMatchCA(),
			node.GlobalMatchCA(),
		},
	}
	controllerClientUS := &PkiConfig{
		Region:      "us",
		Type:        node.NodeTypeController,
		LocalIssuer: node.CertIssuerRegional,
		UseVaultPki: true,
		RemoteCAs: []node.MatchCA{
			node.GlobalMatchCA(),
		},
	}
	controllerServerUS := &PkiConfig{
		Region:      "us",
		Type:        node.NodeTypeController,
		LocalIssuer: node.CertIssuerRegional,
		UseVaultPki: true,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalMatchCA(),
			node.SameRegionalCloudletMatchCA(),
		},
	}
	controllerApiServerUS := &PkiConfig{
		Region:      "us",
		Type:        node.NodeTypeController,
		LocalIssuer: node.CertIssuerRegional,
		UseVaultPki: true,
		RemoteCAs: []node.MatchCA{
			node.GlobalMatchCA(),
			node.SameRegionalMatchCA(),
		},
	}
	controllerApiServerEU := &PkiConfig{
		Region:      "eu",
		Type:        node.NodeTypeController,
		LocalIssuer: node.CertIssuerRegional,
		UseVaultPki: true,
		RemoteCAs: []node.MatchCA{
			node.GlobalMatchCA(),
			node.SameRegionalMatchCA(),
		},
	}
	crmClientUS := &PkiConfig{
		Region:        "us",
		Type:          node.NodeTypeCRM,
		LocalIssuer:   node.CertIssuerRegionalCloudlet,
		UseVaultPki:   true,
		AccessKeyFile: tc1.KeyClient.AccessKeyFile,
		AccessApiAddr: tc1.KeyClient.AccessApiAddr,
		CloudletKey:   &tc1.Cloudlet.Key,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalMatchCA(),
		},
	}
	crmClientEU := &PkiConfig{
		Region:        "eu",
		Type:          node.NodeTypeCRM,
		LocalIssuer:   node.CertIssuerRegionalCloudlet,
		UseVaultPki:   true,
		AccessKeyFile: tc2.KeyClient.AccessKeyFile,
		AccessApiAddr: tc2.KeyClient.AccessApiAddr,
		CloudletKey:   &tc2.Cloudlet.Key,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalMatchCA(),
		},
	}
	dmeClientRegionalUS := &PkiConfig{
		Region:      "us",
		Type:        node.NodeTypeDME,
		LocalIssuer: node.CertIssuerRegional,
		UseVaultPki: true,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalMatchCA(),
		},
	}
	mc := &PkiConfig{
		Type:        node.NodeTypeNotifyRoot,
		LocalIssuer: node.CertIssuerGlobal,
		UseVaultPki: true,
		RemoteCAs: []node.MatchCA{
			node.AnyRegionalMatchCA(),
		},
	}
	// assume attacker stole crm EU certs, and vault login
	// so has regional-cloudlet cert and can pull all CAs.
	crmRogueEU := &PkiConfig{
		Region:        "eu",
		Type:          node.NodeTypeCRM,
		LocalIssuer:   node.CertIssuerRegionalCloudlet,
		UseVaultPki:   true,
		AccessKeyFile: tc2.KeyClient.AccessKeyFile,
		AccessApiAddr: tc2.KeyClient.AccessApiAddr,
		CloudletKey:   &tc2.Cloudlet.Key,
		RemoteCAs: []node.MatchCA{
			node.GlobalMatchCA(),
			node.AnyRegionalMatchCA(),
			node.SameRegionalMatchCA(),
			node.SameRegionalCloudletMatchCA(),
		},
	}
	edgeTurnEU := &PkiConfig{
		Region:      "eu",
		Type:        node.NodeTypeEdgeTurn,
		LocalIssuer: node.CertIssuerRegional,
		UseVaultPki: true,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalCloudletMatchCA(),
		},
	}
	edgeTurnUS := &PkiConfig{
		Region:      "us",
		Type:        node.NodeTypeEdgeTurn,
		LocalIssuer: node.CertIssuerRegional,
		UseVaultPki: true,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalCloudletMatchCA(),
		},
	}

	// Testing for certificate exchange.
	var csTests clientServerList
	// controller can connect to notifyroot
	csTests.add(ClientServer{
		Server: notifyRootServer,
		Client: controllerClientUS,
	})
	// mc can connect to controller
	csTests.add(ClientServer{
		Server: controllerApiServerUS,
		Client: mc,
	})
	csTests.add(ClientServer{
		Server: controllerApiServerEU,
		Client: mc,
	})
	// crm can connect to controller
	csTests.add(ClientServer{
		Server: controllerServerUS,
		Client: crmClientUS,
	})
	// crm from EU cannot connect to US controller
	csTests.add(ClientServer{
		Server:          controllerServerUS,
		Client:          crmClientEU,
		ExpectClientErr: "region mismatch",
		ExpectServerErr: "remote error: tls: bad certificate",
	})
	// crm cannot connect to notifyroot
	csTests.add(ClientServer{
		Server:          notifyRootServer,
		Client:          crmClientUS,
		ExpectClientErr: "certificate signed by unknown authority",
		ExpectServerErr: "remote error: tls: bad certificate",
	})
	// crm can connect to edgeturn
	csTests.add(ClientServer{
		Server: edgeTurnUS,
		Client: crmClientUS,
	})
	// crm from US cannot connect to EU edgeturn
	csTests.add(ClientServer{
		Server:          edgeTurnEU,
		Client:          crmClientUS,
		ExpectClientErr: "region mismatch",
		ExpectServerErr: "remote error: tls: bad certificate",
	})
	// crm from EU cannot connect to US edgeturn
	csTests.add(ClientServer{
		Server:          edgeTurnUS,
		Client:          crmClientEU,
		ExpectClientErr: "region mismatch",
		ExpectServerErr: "remote error: tls: bad certificate",
	})
	// rogue crm cannot connect to notify root
	csTests.add(ClientServer{
		Server:          notifyRootServer,
		Client:          crmRogueEU,
		ExpectClientErr: "remote error: tls: bad certificate",
		ExpectServerErr: "certificate signed by unknown authority",
	})
	// rogue crm cannot connect to other region controller
	csTests.add(ClientServer{
		Server:          controllerServerUS,
		Client:          crmRogueEU,
		ExpectClientErr: "region mismatch",
		ExpectServerErr: "remote error: tls: bad certificate",
	})
	// rogue crm cannot pretend to be controller
	csTests.add(ClientServer{
		Server:          crmRogueEU,
		Client:          crmClientEU,
		ExpectClientErr: "certificate signed by unknown authority",
		ExpectServerErr: "remote error: tls: bad certificate",
	})
	// rogue crm cannot pretend to be notifyroot
	csTests.add(ClientServer{
		Server:          crmRogueEU,
		Client:          controllerClientUS,
		ExpectClientErr: "certificate signed by unknown authority",
		ExpectServerErr: "remote error: tls: bad certificate",
	})

	// These test config options and rollout phases
	nodeNoTls := &PkiConfig{
		Region: "us",
		Type:   node.NodeTypeController,
	}
	nodeFileOnly := &PkiConfig{
		Region:   "us",
		Type:     node.NodeTypeController,
		CertFile: "./ctrl.crt",
		CertKey:  "./ctrl.key",
		CAFile:   "./mex-ca.crt",
	}
	nodePhase2 := &PkiConfig{
		Region:      "us",
		Type:        node.NodeTypeController,
		CertFile:    "./ctrl.crt",
		CertKey:     "./ctrl.key",
		CAFile:      "./mex-ca.crt",
		UseVaultPki: true,
		LocalIssuer: node.CertIssuerRegional,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalMatchCA(),
		},
	}
	nodePhase3 := &PkiConfig{
		Region:      "us",
		Type:        node.NodeTypeController,
		UseVaultPki: true,
		LocalIssuer: node.CertIssuerRegional,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalMatchCA(),
		},
	}
	// local testing
	csTests.add(ClientServer{
		Server: nodeNoTls,
		Client: nodeNoTls,
	})
	// existing
	csTests.add(ClientServer{
		Server: nodeFileOnly,
		Client: nodeFileOnly,
	})
	// phase3
	csTests.add(ClientServer{
		Server: nodePhase3,
		Client: nodePhase3,
	})
	csTests.add(ClientServer{
		Server: nodePhase2,
		Client: nodePhase3,
	})
	csTests.add(ClientServer{
		Server: nodePhase3,
		Client: nodePhase2,
	})
	csTests.add(ClientServer{
		Server:          nodePhase3,
		Client:          nodeFileOnly,
		ExpectClientErr: "certificate signed by unknown authority",
		ExpectServerErr: "remote error: tls: bad certificate",
	})

	for _, test := range csTests {
		testExchange(t, ctx, vroles, &test)
	}

	// Tests for Tls interceptor that allows access for Global/Regional
	// clients, but requires an additional access key for RegionalCloudlet.
	// This will be used on the notify API.
	var ccTests clientControllerList
	// crm can connect within same region
	ccTests.add(ClientController{
		Controller:                 dcUS,
		Client:                     crmClientUS,
		ControllerRequireAccessKey: true,
	})
	ccTests.add(ClientController{
		Controller:                 dcEU,
		Client:                     crmClientEU,
		ControllerRequireAccessKey: true,
	})
	// crm cannot connect to different region
	ccTests.add(ClientController{
		Controller:                 dcUS,
		Client:                     crmClientEU,
		ControllerRequireAccessKey: true,
		ExpectErr:                  "region mismatch, expected local uri sans for eu",
	})
	ccTests.add(ClientController{
		Controller:                 dcEU,
		Client:                     crmClientUS,
		ControllerRequireAccessKey: true,
		ExpectErr:                  "region mismatch, expected local uri sans for us",
	})
	// test invalid keys
	ccTests.add(ClientController{
		Controller:                 dcUS,
		Client:                     crmClientUS,
		ControllerRequireAccessKey: true,
		InvalidateClientKey:        true,
		ExpectErr:                  "Failed to verify cloudlet access key signature",
	})
	ccTests.add(ClientController{
		Controller:                 dcEU,
		Client:                     crmClientEU,
		ControllerRequireAccessKey: true,
		InvalidateClientKey:        true,
		ExpectErr:                  "Failed to verify cloudlet access key signature",
	})
	// ignore invalid keys or missing keys for backwards compatibility
	// with CRMs that were not upgraded
	ccTests.add(ClientController{
		Controller:                 dcUS,
		Client:                     crmClientUS,
		ControllerRequireAccessKey: false,
		InvalidateClientKey:        true,
	})
	ccTests.add(ClientController{
		Controller:                 dcEU,
		Client:                     crmClientEU,
		ControllerRequireAccessKey: false,
		InvalidateClientKey:        true,
	})
	// same regional service is ok (regional dme, etc)
	ccTests.add(ClientController{
		Controller: dcUS,
		Client:     dmeClientRegionalUS,
	})
	for _, test := range ccTests {
		testTlsConnect(t, ctx, &test)
	}
}

type ConfigTest struct {
	NodeType      string
	Region        string
	LocalIssuer   string
	TestIssuer    string
	RemoteCAs     []node.MatchCA
	ExpectErr     string
	Line          string
	AccessKeyFile string
	AccessApiAddr string
	CloudletKey   *edgeproto.CloudletKey
}

func testGetTlsConfig(t *testing.T, ctx context.Context, vroles *process.VaultRoles, cfg *ConfigTest) {
	log.SpanLog(ctx, log.DebugLevelInfo, "run testGetTlsConfig", "cfg", cfg)
	vc := getVaultConfig(cfg.NodeType, cfg.Region, vroles)
	mgr := node.NodeMgr{}
	mgr.InternalPki.UseVaultPki = true
	mgr.InternalDomain = "mobiledgex.net"
	if cfg.AccessKeyFile != "" && cfg.AccessApiAddr != "" {
		mgr.AccessKeyClient.AccessKeyFile = cfg.AccessKeyFile
		mgr.AccessKeyClient.AccessApiAddr = cfg.AccessApiAddr
		mgr.AccessKeyClient.TestSkipTlsVerify = true
	}
	// nodeMgr init will attempt to issue a cert to be able to talk
	// to Jaeger/ElasticSearch
	opts := []node.NodeOp{
		node.WithRegion(cfg.Region),
		node.WithVaultConfig(vc),
	}
	if cfg.CloudletKey != nil {
		opts = append(opts, node.WithCloudletKey(cfg.CloudletKey))
	}
	_, _, err := mgr.Init(cfg.NodeType, cfg.LocalIssuer, opts...)
	require.Nil(t, err, "nodeMgr init %s", cfg.Line)
	_, err = mgr.InternalPki.GetServerTlsConfig(ctx,
		mgr.CommonName(),
		cfg.TestIssuer,
		cfg.RemoteCAs)
	if cfg.ExpectErr == "" {
		require.Nil(t, err, "get tls config %s", cfg.Line)
	} else {
		require.NotNil(t, err, "get tls config %s", cfg.Line)
		require.Contains(t, err.Error(), cfg.ExpectErr, "get tls config %s", cfg.Line)
	}
}

type PkiConfig struct {
	Region        string
	Type          string
	LocalIssuer   string
	CertFile      string
	CertKey       string
	CAFile        string
	UseVaultPki   bool
	RemoteCAs     []node.MatchCA
	AccessKeyFile string
	AccessApiAddr string
	CloudletKey   *edgeproto.CloudletKey
}

type ClientServer struct {
	Server          *PkiConfig
	Client          *PkiConfig
	ExpectServerErr string
	ExpectClientErr string
	Line            string
}

func (s *PkiConfig) setupNodeMgr(vroles *process.VaultRoles) (*node.NodeMgr, error) {
	vaultCfg := getVaultConfig(s.Type, s.Region, vroles)
	nodeMgr := node.NodeMgr{}
	nodeMgr.SetInternalTlsCertFile(s.CertFile)
	nodeMgr.SetInternalTlsKeyFile(s.CertKey)
	nodeMgr.SetInternalTlsCAFile(s.CAFile)
	nodeMgr.InternalPki.UseVaultPki = s.UseVaultPki
	nodeMgr.InternalDomain = "mobiledgex.net"
	if s.AccessKeyFile != "" && s.AccessApiAddr != "" {
		nodeMgr.AccessKeyClient.AccessKeyFile = s.AccessKeyFile
		nodeMgr.AccessKeyClient.AccessApiAddr = s.AccessApiAddr
		nodeMgr.AccessKeyClient.TestSkipTlsVerify = true
	}
	opts := []node.NodeOp{
		node.WithRegion(s.Region),
		node.WithVaultConfig(vaultCfg),
	}
	if s.CloudletKey != nil {
		opts = append(opts, node.WithCloudletKey(s.CloudletKey))
	}
	_, _, err := nodeMgr.Init(s.Type, s.LocalIssuer, opts...)
	return &nodeMgr, err
}

func testExchange(t *testing.T, ctx context.Context, vroles *process.VaultRoles, cs *ClientServer) {
	if !strings.HasPrefix(runtime.Version(), "go1.12") {
		// After go1.12, the client side Dial/Handshake does not return
		// error when the server decides to abort the connection.
		// Only the server side returns an error.
		if cs.ExpectClientErr == "remote error: tls: bad certificate" {
			cs.ExpectClientErr = ""
		}
	}
	fmt.Printf("******************* testExchange %s *********************\n", cs.Line)
	serverNode, err := cs.Server.setupNodeMgr(vroles)
	require.Nil(t, err, "serverNode init %s", cs.Line)
	serverTls, err := serverNode.InternalPki.GetServerTlsConfig(ctx,
		serverNode.CommonName(),
		cs.Server.LocalIssuer,
		cs.Server.RemoteCAs)
	require.Nil(t, err, "get server tls config %s", cs.Line)
	if cs.Server.CertFile != "" || cs.Server.UseVaultPki {
		require.NotNil(t, serverTls)
	}

	clientNode, err := cs.Client.setupNodeMgr(vroles)
	require.Nil(t, err, "clientNode init %s", cs.Line)
	clientTls, err := clientNode.InternalPki.GetClientTlsConfig(ctx,
		clientNode.CommonName(),
		cs.Client.LocalIssuer,
		cs.Client.RemoteCAs)
	require.Nil(t, err, "get client tls config %s", cs.Line)
	if cs.Client.CertFile != "" || cs.Client.UseVaultPki {
		require.NotNil(t, clientTls)
		// must set ServerName due to the way this test is set up
		clientTls.ServerName = serverNode.CommonName()
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.Nil(t, err)
	defer lis.Close()

	connDone := make(chan bool, 2)
	// loop twice so we can test cert refresh
	for i := 0; i < 2; i++ {
		var serr error
		go func() {
			var sconn net.Conn
			sconn, serr = lis.Accept()
			defer sconn.Close()
			if serverTls != nil {
				srv := tls.Server(sconn, serverTls)
				serr = srv.Handshake()
			}
			connDone <- true
		}()
		log.SpanLog(ctx, log.DebugLevelInfo, "client dial", "addr", lis.Addr().String())
		var err error
		if clientTls == nil {
			var conn net.Conn
			conn, err = net.Dial("tcp", lis.Addr().String())
			if err == nil {
				defer conn.Close()
			}
		} else {
			var conn *tls.Conn
			conn, err = tls.Dial("tcp", lis.Addr().String(), clientTls)
			if err == nil {
				defer conn.Close()
				err = conn.Handshake()
			}
		}
		<-connDone
		if cs.ExpectClientErr == "" {
			require.Nil(t, err, "client dial/handshake [%d] %s", i, cs.Line)
		} else {
			require.NotNil(t, err, "client dial [%d] %s", i, cs.Line)
			require.Contains(t, err.Error(), cs.ExpectClientErr, "client error check for [%d] %s", i, cs.Line)
		}
		if cs.ExpectServerErr == "" {
			require.Nil(t, serr, "server dial/handshake [%d] %s", i, cs.Line)
		} else {
			require.NotNil(t, serr, "server accept [%d] %s", i, cs.Line)
			require.Contains(t, serr.Error(), cs.ExpectServerErr, "server error check for [%d] %s", i, cs.Line)
		}
		if i == 1 {
			// no need to refresh on last iteration
			break
		}
		// refresh certs. same tls config should pick up new certs.
		err = serverNode.InternalPki.RefreshNow(ctx)
		require.Nil(t, err, "refresh server certs [%d] %s", i, cs.Line)
		err = clientNode.InternalPki.RefreshNow(ctx)
		require.Nil(t, err, "refresh client certs [%d] %s", i, cs.Line)
	}
}

type ClientController struct {
	Controller                 *DummyController
	Client                     *PkiConfig
	ControllerRequireAccessKey bool
	InvalidateClientKey        bool
	ExpectErr                  string
	Line                       string
}

func testTlsConnect(t *testing.T, ctx context.Context, cc *ClientController) {
	// This tests the TLS interceptors that will require
	// access keys if client uses the RegionalCloudlet cert.
	fmt.Printf("******************* testTlsConnect %s *********************\n", cc.Line)
	cc.Controller.DummyController.KeyServer.SetRequireTlsAccessKey(cc.ControllerRequireAccessKey)

	clientNode, err := cc.Client.setupNodeMgr(cc.Controller.vroles)
	require.Nil(t, err, "clientNode init %s", cc.Line)
	clientTls, err := clientNode.InternalPki.GetClientTlsConfig(ctx,
		clientNode.CommonName(),
		cc.Client.LocalIssuer,
		cc.Client.RemoteCAs)
	require.Nil(t, err, "get client tls config %s", cc.Line)
	if cc.Client.CertFile != "" || cc.Client.UseVaultPki {
		require.NotNil(t, clientTls)
		// must set ServerName due to the way this test is set up
		clientTls.ServerName = cc.Controller.nodeMgr.CommonName()
	}
	// for negative testing with invalid key
	if cc.InvalidateClientKey && cc.Client.CloudletKey != nil {
		cc.Controller.UpdateKey(ctx, *cc.Client.CloudletKey)
	}
	// interceptors will add access key to grpc metadata if access key present
	unaryInterceptor := log.UnaryClientTraceGrpc
	if cc.Client.AccessKeyFile != "" {
		unaryInterceptor = grpc_middleware.ChainUnaryClient(
			log.UnaryClientTraceGrpc,
			clientNode.AccessKeyClient.UnaryAddAccessKey)
	}
	streamInterceptor := log.StreamClientTraceGrpc
	if cc.Client.AccessKeyFile != "" {
		streamInterceptor = grpc_middleware.ChainStreamClient(
			log.StreamClientTraceGrpc,
			clientNode.AccessKeyClient.StreamAddAccessKey)
	}

	clientConn, err := grpc.Dial(cc.Controller.TlsAddr(),
		edgetls.GetGrpcDialOption(clientTls),
		grpc.WithUnaryInterceptor(unaryInterceptor),
		grpc.WithStreamInterceptor(streamInterceptor),
	)
	require.Nil(t, err, "create client conn %s", cc.Line)
	node.EchoApisTest(t, ctx, clientConn, cc.ExpectErr)
}

func getVaultConfig(nodetype, region string, vroles *process.VaultRoles) *vault.Config {
	var roleid string
	var secretid string

	if nodetype == node.NodeTypeNotifyRoot {
		roleid = vroles.NotifyRootRoleID
		secretid = vroles.NotifyRootSecretID
	} else {
		if region == "" {
			// for testing, map to us region"
			region = "us"
		}
		rr := vroles.GetRegionRoles(region)
		if rr == nil {
			panic("no roles for region")
		}
		switch nodetype {
		case node.NodeTypeDME:
			roleid = rr.DmeRoleID
			secretid = rr.DmeSecretID
		case node.NodeTypeCRM:
			// no vault access for crm
			return nil
		case node.NodeTypeController:
			roleid = rr.CtrlRoleID
			secretid = rr.CtrlSecretID
		case node.NodeTypeClusterSvc:
			roleid = rr.ClusterSvcRoleID
			secretid = rr.ClusterSvcSecretID
		case node.NodeTypeEdgeTurn:
			roleid = rr.EdgeTurnRoleID
			secretid = rr.EdgeTurnSecretID
		default:
			panic("invalid node type")
		}
	}
	auth := vault.NewAppRoleAuth(roleid, secretid)
	return vault.NewConfig(process.VaultAddress, auth)
}

// Track line number of objects added to list to make it easier
// to debug if one of them fails.

type cfgTestList []ConfigTest

func (list *cfgTestList) add(cfg ConfigTest) {
	_, file, line, _ := runtime.Caller(1)
	cfg.Line = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	*list = append(*list, cfg)
}

type clientServerList []ClientServer

func (list *clientServerList) add(cs ClientServer) {
	_, file, line, _ := runtime.Caller(1)
	cs.Line = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	*list = append(*list, cs)
}

type clientControllerList []ClientController

func (list *clientControllerList) add(cc ClientController) {
	_, file, line, _ := runtime.Caller(1)
	cc.Line = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	*list = append(*list, cc)
}

// Dummy controller serves Vault certs to access key clients
type DummyController struct {
	node.DummyController
	nodeMgr       node.NodeMgr
	vroles        *process.VaultRoles
	TlsLis        net.Listener
	TlsServ       *grpc.Server
	TlsRegisterCb func(server *grpc.Server)
}

func (s *DummyController) Init(ctx context.Context, region string, vroles *process.VaultRoles, vaultAddr string) error {
	s.DummyController.Init(vaultAddr)
	s.DummyController.RegisterCloudletAccess = false // register it here
	s.DummyController.ApiRegisterCb = func(serv *grpc.Server) {
		// add APIs to issue certs to CRM/etc
		edgeproto.RegisterCloudletAccessApiServer(serv, s)
	}
	es := &node.EchoServer{}
	s.TlsRegisterCb = func(serv *grpc.Server) {
		// echo server for testing
		echo.RegisterEchoServer(serv, es)
	}
	// no crm vault role/secret env vars for controller (no backwards compatability)
	s.vroles = vroles

	vc := getVaultConfig(node.NodeTypeController, region, vroles)
	s.nodeMgr.InternalPki.UseVaultPki = true
	s.nodeMgr.InternalDomain = "mobiledgex.net"
	_, _, err := s.nodeMgr.Init(node.NodeTypeController, node.NoTlsClientIssuer, node.WithRegion(region), node.WithVaultConfig(vc))
	return err
}

func (s *DummyController) Start(ctx context.Context) {
	s.DummyController.Start(ctx, "127.0.0.1:0")
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err.Error())
	}
	s.TlsLis = lis
	// same config as Controller's notify server
	tlsConfig, err := s.nodeMgr.InternalPki.GetServerTlsConfig(ctx,
		s.nodeMgr.CommonName(),
		node.CertIssuerRegional,
		[]node.MatchCA{
			node.SameRegionalMatchCA(),
			node.SameRegionalCloudletMatchCA(),
		})
	if err != nil {
		panic(err.Error())
	}
	// The "tls" interceptors, which only require an access-key
	// if the client is using a RegionalCloudlet cert.
	s.TlsServ = grpc.NewServer(
		cloudcommon.GrpcCreds(tlsConfig),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			cloudcommon.AuditUnaryInterceptor,
			s.KeyServer.UnaryTlsAccessKey,
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			cloudcommon.AuditStreamInterceptor,
			s.KeyServer.StreamTlsAccessKey,
		)),
	)
	if s.TlsRegisterCb != nil {
		s.TlsRegisterCb(s.TlsServ)
	}
	go func() {
		err := s.TlsServ.Serve(s.TlsLis)
		if err != nil {
			panic(err.Error())
		}
	}()
}

func (s *DummyController) Stop() {
	s.DummyController.Stop()
	s.TlsServ.Stop()
	s.TlsLis.Close()
}

func (s *DummyController) TlsAddr() string {
	return s.TlsLis.Addr().String()
}

func (s *DummyController) IssueCert(ctx context.Context, req *edgeproto.IssueCertRequest) (*edgeproto.IssueCertReply, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "dummy controller issue cert", "req", req)
	reply := &edgeproto.IssueCertReply{}
	certId := node.CertId{
		CommonName: req.CommonName,
		Issuer:     node.CertIssuerRegionalCloudlet,
	}
	vc, err := s.nodeMgr.InternalPki.IssueVaultCertDirect(ctx, certId)
	if err != nil {
		return reply, err
	}
	reply.PublicCertPem = string(vc.PublicCertPEM)
	reply.PrivateKeyPem = string(vc.PrivateKeyPEM)
	return reply, nil
}

func (s *DummyController) GetCas(ctx context.Context, req *edgeproto.GetCasRequest) (*edgeproto.GetCasReply, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "dummy controller get cas", "req", req)
	reply := &edgeproto.GetCasReply{}
	cab, err := s.nodeMgr.InternalPki.GetVaultCAsDirect(ctx, req.Issuer)
	if err != nil {
		return reply, err
	}
	reply.CaChainPem = string(cab)
	return reply, err
}

func (s *DummyController) GetAccessData(ctx context.Context, in *edgeproto.AccessDataRequest) (*edgeproto.AccessDataReply, error) {
	return &edgeproto.AccessDataReply{}, nil
}
