package node_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/require"
)

// Note file package is not node, so avoids node package having
// dependencies on process package.

func TestInternalPki(t *testing.T) {
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
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
		LocalIssuer: node.CertIssuerGlobal,
		ExpectErr:   "write failure pki-global/issue/us",
	})
	// regional Controller cannot issue cloudlet cert
	cfgTests.add(ConfigTest{
		NodeType:    node.NodeTypeController,
		Region:      "us",
		LocalIssuer: node.CertIssuerRegionalCloudlet,
		ExpectErr:   "write failure pki-regional-cloudlet/issue/us",
	})
	// global node cannot issue regional cert
	cfgTests.add(ConfigTest{
		NodeType:    node.NodeTypeNotifyRoot,
		LocalIssuer: node.CertIssuerRegional,
		ExpectErr:   "write failure pki-regional/issue/default",
	})
	// global node cannot issue regional-cloudlet cert
	cfgTests.add(ConfigTest{
		NodeType:    node.NodeTypeNotifyRoot,
		LocalIssuer: node.CertIssuerRegionalCloudlet,
		ExpectErr:   "write failure pki-regional-cloudlet/issue/default",
	})
	// cloudlet node cannot issue global cert
	cfgTests.add(ConfigTest{
		NodeType:    node.NodeTypeCRM,
		Region:      "us",
		LocalIssuer: node.CertIssuerGlobal,
		ExpectErr:   "write failure pki-global/issue/us",
	})
	// cloudlet node cannot issue regional cert
	cfgTests.add(ConfigTest{
		NodeType:    node.NodeTypeCRM,
		Region:      "us",
		LocalIssuer: node.CertIssuerRegional,
		ExpectErr:   "write failure pki-regional/issue/us",
	})

	for _, cfg := range cfgTests {
		testGetTlsConfig(t, ctx, vroles, &cfg)
	}

	// define nodes for certificate exchange tests
	notifyRootServer := &PkiConfig{
		Type:          node.NodeTypeNotifyRoot,
		LocalIssuer:   node.CertIssuerGlobal,
		UseVaultCerts: true,
		RemoteCAs: []node.MatchCA{
			node.AnyRegionalMatchCA(),
			node.GlobalMatchCA(),
		},
	}
	controllerClientUS := &PkiConfig{
		Region:        "us",
		Type:          node.NodeTypeController,
		LocalIssuer:   node.CertIssuerRegional,
		UseVaultCerts: true,
		RemoteCAs: []node.MatchCA{
			node.GlobalMatchCA(),
		},
	}
	controllerServerUS := &PkiConfig{
		Region:        "us",
		Type:          node.NodeTypeController,
		LocalIssuer:   node.CertIssuerRegional,
		UseVaultCerts: true,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalMatchCA(),
			node.SameRegionalCloudletMatchCA(),
		},
	}
	crmClientUS := &PkiConfig{
		Region:        "us",
		Type:          node.NodeTypeCRM,
		LocalIssuer:   node.CertIssuerRegionalCloudlet,
		UseVaultCerts: true,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalMatchCA(),
		},
	}
	crmClientEU := &PkiConfig{
		Region:        "eu",
		Type:          node.NodeTypeCRM,
		LocalIssuer:   node.CertIssuerRegionalCloudlet,
		UseVaultCerts: true,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalMatchCA(),
		},
	}
	// assume attacker stole crm EU certs, and vault login
	// so has regional-cloudlet cert and can pull all CAs.
	crmRogueEU := &PkiConfig{
		Region:        "eu",
		Type:          node.NodeTypeCRM,
		LocalIssuer:   node.CertIssuerRegionalCloudlet,
		UseVaultCerts: true,
		RemoteCAs: []node.MatchCA{
			node.GlobalMatchCA(),
			node.AnyRegionalMatchCA(),
			node.SameRegionalMatchCA(),
			node.SameRegionalCloudletMatchCA(),
		},
	}
	edgeTurnEU := &PkiConfig{
		Region:        "eu",
		Type:          node.NodeTypeEdgeTurn,
		LocalIssuer:   node.CertIssuerRegional,
		UseVaultCerts: true,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalCloudletMatchCA(),
		},
	}
	edgeTurnUS := &PkiConfig{
		Region:        "us",
		Type:          node.NodeTypeEdgeTurn,
		LocalIssuer:   node.CertIssuerRegional,
		UseVaultCerts: true,
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
	// crm can connect to controller
	csTests.add(ClientServer{
		Server: controllerServerUS,
		Client: crmClientUS,
	})
	// crm from EU cannot connect to US controller
	csTests.add(ClientServer{
		Server:    controllerServerUS,
		Client:    crmClientEU,
		ExpectErr: "region mismatch",
	})
	// crm cannot connect to notifyroot
	csTests.add(ClientServer{
		Server:    notifyRootServer,
		Client:    crmClientUS,
		ExpectErr: "certificate signed by unknown authority",
	})
	// crm can connect to edgeturn
	csTests.add(ClientServer{
		Server: edgeTurnUS,
		Client: crmClientUS,
	})
	// crm from US cannot connect to EU edgeturn
	csTests.add(ClientServer{
		Server:    edgeTurnEU,
		Client:    crmClientUS,
		ExpectErr: "region mismatch",
	})
	// crm from EU cannot connect to US edgeturn
	csTests.add(ClientServer{
		Server:    edgeTurnUS,
		Client:    crmClientEU,
		ExpectErr: "region mismatch",
	})
	// rogue crm cannot connect to notify root
	csTests.add(ClientServer{
		Server:    notifyRootServer,
		Client:    crmRogueEU,
		ExpectErr: "remote error: tls: bad certificate",
	})
	// rogue crm cannot connect to other region controller
	csTests.add(ClientServer{
		Server:    controllerServerUS,
		Client:    crmRogueEU,
		ExpectErr: "region mismatch",
	})
	// rogue crm cannot pretend to be controller
	csTests.add(ClientServer{
		Server:    crmRogueEU,
		Client:    crmClientEU,
		ExpectErr: "certificate signed by unknown authority",
	})
	// rogue crm cannot pretend to be notifyroot
	csTests.add(ClientServer{
		Server:    crmRogueEU,
		Client:    controllerClientUS,
		ExpectErr: "certificate signed by unknown authority",
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
	}
	nodePhase1 := &PkiConfig{
		Region:      "us",
		Type:        node.NodeTypeController,
		CertFile:    "./ctrl.crt",
		UseVaultCAs: true,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalMatchCA(),
		},
	}
	nodePhase2 := &PkiConfig{
		Region:        "us",
		Type:          node.NodeTypeController,
		CertFile:      "./ctrl.crt",
		UseVaultCerts: true,
		LocalIssuer:   node.CertIssuerRegional,
		RemoteCAs: []node.MatchCA{
			node.SameRegionalMatchCA(),
		},
	}
	nodePhase3 := &PkiConfig{
		Region:        "us",
		Type:          node.NodeTypeController,
		UseVaultCerts: true,
		LocalIssuer:   node.CertIssuerRegional,
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
	// phase1
	csTests.add(ClientServer{
		Server: nodePhase1,
		Client: nodePhase1,
	})
	csTests.add(ClientServer{
		Server: nodeFileOnly,
		Client: nodePhase1,
	})
	csTests.add(ClientServer{
		Server: nodePhase1,
		Client: nodeFileOnly,
	})
	// phase2
	csTests.add(ClientServer{
		Server: nodePhase2,
		Client: nodePhase2,
	})
	csTests.add(ClientServer{
		Server: nodePhase1,
		Client: nodePhase2,
	})
	csTests.add(ClientServer{
		Server: nodePhase2,
		Client: nodePhase1,
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
		Server:    nodePhase3,
		Client:    nodeFileOnly,
		ExpectErr: "certificate signed by unknown authority",
	})

	for _, test := range csTests {
		testExchange(t, ctx, vroles, &test)
	}
}

type ConfigTest struct {
	NodeType      string
	Region        string
	VaultNodeType string
	VaultRegion   string
	LocalIssuer   string
	RemoteCAs     []node.MatchCA
	ExpectErr     string
	Line          string
}

func testGetTlsConfig(t *testing.T, ctx context.Context, vroles *process.VaultRoles, cfg *ConfigTest) {
	if cfg.VaultNodeType == "" {
		cfg.VaultNodeType = cfg.NodeType
	}
	if cfg.VaultRegion == "" {
		cfg.VaultRegion = cfg.Region
	}
	vc := getVaultConfig(cfg.VaultNodeType, cfg.VaultRegion, vroles)
	mgr := node.NodeMgr{}
	mgr.InternalPki.UseVaultCerts = true
	_, _, err := mgr.Init(cfg.NodeType, node.NoTlsClientIssuer, node.WithRegion(cfg.Region), node.WithVaultConfig(vc))
	require.Nil(t, err)
	_, err = mgr.InternalPki.GetServerTlsConfig(ctx,
		mgr.CommonName(),
		cfg.LocalIssuer,
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
	UseVaultCAs   bool
	UseVaultCerts bool
	RemoteCAs     []node.MatchCA
}

type ClientServer struct {
	Server    *PkiConfig
	Client    *PkiConfig
	ExpectErr string
	Line      string
}

func testExchange(t *testing.T, ctx context.Context, vroles *process.VaultRoles, cs *ClientServer) {
	fmt.Printf("******************* testExchange %s *********************\n", cs.Line)
	serverVault := getVaultConfig(cs.Server.Type, cs.Server.Region, vroles)
	serverNode := node.NodeMgr{}
	serverNode.TlsCertFile = cs.Server.CertFile
	serverNode.InternalPki.UseVaultCAs = cs.Server.UseVaultCAs
	serverNode.InternalPki.UseVaultCerts = cs.Server.UseVaultCerts
	serverNode.InternalDomain = "mobiledgex.net"
	_, _, err := serverNode.Init(cs.Server.Type, node.NoTlsClientIssuer,
		node.WithRegion(cs.Server.Region),
		node.WithVaultConfig(serverVault))
	require.Nil(t, err)
	serverTls, err := serverNode.InternalPki.GetServerTlsConfig(ctx,
		serverNode.CommonName(),
		cs.Server.LocalIssuer,
		cs.Server.RemoteCAs)
	require.Nil(t, err, "get server tls config %s", cs.Line)
	if cs.Server.CertFile != "" || cs.Server.UseVaultCerts {
		require.NotNil(t, serverTls)
	}

	clientVault := getVaultConfig(cs.Client.Type, cs.Client.Region, vroles)
	clientNode := node.NodeMgr{}
	clientNode.TlsCertFile = cs.Client.CertFile
	clientNode.InternalPki.UseVaultCAs = cs.Client.UseVaultCAs
	clientNode.InternalPki.UseVaultCerts = cs.Client.UseVaultCerts
	clientNode.InternalDomain = "mobiledgex.net"
	_, _, err = clientNode.Init(cs.Client.Type, node.NoTlsClientIssuer,
		node.WithRegion(cs.Client.Region),
		node.WithVaultConfig(clientVault))
	require.Nil(t, err)
	clientTls, err := clientNode.InternalPki.GetClientTlsConfig(ctx,
		clientNode.CommonName(),
		cs.Client.LocalIssuer,
		cs.Client.RemoteCAs)
	require.Nil(t, err, "get client tls config %s", cs.Line)
	if cs.Client.CertFile != "" || cs.Client.UseVaultCerts {
		require.NotNil(t, clientTls)
		// must set ServerName due to the way this test is set up
		clientTls.ServerName = serverNode.CommonName()
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.Nil(t, err)
	defer lis.Close()

	go func() {
		for i := 0; i < 2; i++ {
			sconn, err := lis.Accept()
			require.Nil(t, err, "accept")
			defer sconn.Close()

			if serverTls != nil {
				srv := tls.Server(sconn, serverTls)
				srv.Handshake()
			}
		}
	}()

	// loop twice so we can test cert refresh
	for i := 0; i < 2; i++ {
		log.SpanLog(ctx, log.DebugLevelInfo, "client dial")
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
		if cs.ExpectErr == "" {
			require.Nil(t, err, "client dial/handshake [%d] %s", i, cs.Line)
		} else {
			require.NotNil(t, err, "client dial [%d] %s", i, cs.Line)
			require.Contains(t, err.Error(), cs.ExpectErr, "error check for [%d] %s", i, cs.Line)
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
			roleid = rr.CRMRoleID
			secretid = rr.CRMSecretID
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
