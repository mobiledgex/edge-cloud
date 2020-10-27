package node

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/features/proto/echo"
)

// Test access key verification via grpc interceptors.
// This only tests the "required" interceptors, because we need Vault certs
// to test the "tls" interceptors. Those are tested in the pki tests.
func TestAccessClientServer(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()

	// reduce timeout
	BadAuthDelay = time.Millisecond
	vaultRole = ""
	vaultSecret = ""

	ctx := log.StartTestSpan(context.Background())
	// use initCtx to test that init call to controller works without span
	initCtx := log.ContextWithSpan(context.Background(), log.NoTracingSpan())

	// dummy controller
	es := &EchoServer{}
	dc := DummyController{
		ApiRegisterCb: func(serv *grpc.Server) {
			echo.RegisterEchoServer(serv, es)
		},
	}
	dc.Init()
	dc.KeyServer.SetCrmVaultAuth("", "")
	dc.Start()
	defer dc.Stop()

	// ----------------------------------------------------------------
	log.SpanLog(ctx, log.DebugLevelInfo,
		"---- client with no auth credentials, expect failure  ----")
	clientConn, err := grpc.Dial(dc.ApiAddr(), grpc.WithInsecure())
	require.Nil(t, err)
	// API calls should fail
	EchoApisTest(t, ctx, clientConn, "access-key-data not found in metadata")
	clientConn.Close()

	// ----------------------------------------------------------------
	log.SpanLog(ctx, log.DebugLevelInfo,
		"---- client with valid access key ----")
	tc1 := dc.CreateCloudlet(ctx, "tc1")
	// set up access key
	err = dc.UpdateKey(ctx, tc1.Cloudlet.Key)
	require.Nil(t, err)
	// init client
	err = tc1.KeyClient.init(initCtx, NodeTypeCRM, CertIssuerRegionalCloudlet, tc1.Cloudlet.Key)
	require.Nil(t, err)
	// API calls should succeed
	clientConn = startClient(t, ctx, tc1.KeyClient)
	EchoApisTest(t, ctx, clientConn, "")
	clientConn.Close()

	// ----------------------------------------------------------------
	log.SpanLog(ctx, log.DebugLevelInfo,
		"---- manually rotate access key on server to invalidate it ----")
	err = dc.UpdateKey(ctx, tc1.Cloudlet.Key)
	require.Nil(t, err)
	clientConn = startClient(t, ctx, tc1.KeyClient)
	EchoApisTest(t, ctx, clientConn, "Failed to verify cloudlet access key signature")
	clientConn.Close()
	tc1.Cleanup()

	// ----------------------------------------------------------------
	log.SpanLog(ctx, log.DebugLevelInfo, "---- new crm upgrade ---")
	tc2 := dc.CreateCloudlet(ctx, "tc2")
	// set up access key
	err = dc.UpdateKey(ctx, tc2.Cloudlet.Key)
	require.Nil(t, err)
	// save current private key
	privKey := tc2.privateKeyPEM
	// mark key for upgrade
	tc2.Cloudlet.CrmAccessKeyUpgradeRequired = true
	dc.Cache.Update(ctx, &tc2.Cloudlet, 0)
	// run init (upgrade)
	err = tc2.KeyClient.init(initCtx, NodeTypeCRM, CertIssuerRegionalCloudlet, tc2.Cloudlet.Key)
	require.Nil(t, err)
	// check that backup file exists and contains old key
	dat, err := ioutil.ReadFile(tc2.KeyClient.backupKeyFile())
	require.Nil(t, err)
	require.Equal(t, privKey, string(dat))
	// check the primary file exists and is different from old key
	dat, err = ioutil.ReadFile(tc2.KeyClient.AccessKeyFile)
	require.Nil(t, err)
	require.NotEqual(t, privKey, string(dat))
	// check that access works
	clientConn = startClient(t, ctx, tc2.KeyClient)
	EchoApisTest(t, ctx, clientConn, "")
	clientConn.Close()
	tc2.Cleanup()

	// ----------------------------------------------------------------
	log.SpanLog(ctx, log.DebugLevelInfo, "---- old crm upgrade ----")
	tc3 := dc.CreateCloudlet(ctx, "tc3")
	// access key should not exist
	_, err = os.Stat(tc3.KeyClient.AccessKeyFile)
	require.NotNil(t, err)
	// do not generate access key, instead set client vault role and secret
	vaultRole = "role"
	vaultSecret = "secret"
	// init client will fail because no access key, and controller does not
	// have matching vault creds
	err = tc3.KeyClient.init(initCtx, NodeTypeCRM, CertIssuerRegionalCloudlet, tc3.Cloudlet.Key)
	require.NotNil(t, err)
	// Set server vault creds to match, init should now succeed
	dc.KeyServer.SetCrmVaultAuth(vaultRole, vaultSecret)
	err = tc3.KeyClient.init(initCtx, NodeTypeCRM, CertIssuerRegionalCloudlet, tc3.Cloudlet.Key)
	require.Nil(t, err)
	// access key should now exist
	_, err = os.Stat(tc3.KeyClient.AccessKeyFile)
	require.Nil(t, err)
	// check that access works
	clientConn = startClient(t, ctx, tc3.KeyClient)
	EchoApisTest(t, ctx, clientConn, "")
	clientConn.Close()
	tc3.Cleanup()

	// ----------------------------------------------------------------
	log.SpanLog(ctx, log.DebugLevelInfo, "---- non-crm verify only ----")
	tc4 := dc.CreateCloudlet(ctx, "tc4")
	// set up access key
	err = dc.UpdateKey(ctx, tc4.Cloudlet.Key)
	require.Nil(t, err)
	// save current private key
	privKey = tc4.privateKeyPEM
	// init client, should succeed to verify access key
	err = tc4.KeyClient.init(initCtx, NodeTypeDME, CertIssuerRegionalCloudlet, tc4.Cloudlet.Key)
	require.Nil(t, err)
	// mark key for upgrade
	tc4.Cloudlet.CrmAccessKeyUpgradeRequired = true
	dc.Cache.Update(ctx, &tc4.Cloudlet, 0)
	// init client, should fail because upgrade required
	err = tc4.KeyClient.init(initCtx, NodeTypeDME, CertIssuerRegionalCloudlet, tc4.Cloudlet.Key)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "upgrade required")
	// non crm should not touch private key, make sure it's still the same
	dat, err = ioutil.ReadFile(tc4.KeyClient.AccessKeyFile)
	require.Nil(t, err)
	require.Equal(t, privKey, string(dat))
	tc4.Cleanup()
}

func startClient(t *testing.T, ctx context.Context, keyClient *AccessKeyClient) *grpc.ClientConn {
	// start connection
	clientConn, err := keyClient.ConnectController(ctx)
	require.Nil(t, err)
	return clientConn
}

func EchoApisTest(t *testing.T, ctx context.Context, conn *grpc.ClientConn, errMsg string) {
	// All apis start new context to avoid baggage (vars, metadata)
	// on existing context.
	client := echo.NewEchoClient(conn)
	// Unary
	span, cctx := log.ChildSpan(ctx, log.DebugLevelApi, "unary-request")
	defer span.Finish()
	_, err := client.UnaryEcho(cctx, &echo.EchoRequest{})
	EchoApisTestCheckErr(t, err, errMsg)
	// Server streaming
	span, cctx = log.ChildSpan(ctx, log.DebugLevelApi, "stream-request")
	defer span.Finish()
	sstream, err := client.ServerStreamingEcho(cctx, &echo.EchoRequest{})
	if err == nil {
		_, err = sstream.Recv()
	}
	EchoApisTestCheckErr(t, err, errMsg)
	// Client streaming
	span, cctx = log.ChildSpan(ctx, log.DebugLevelApi, "stream-request")
	defer span.Finish()
	cstream, err := client.ClientStreamingEcho(cctx)
	if err == nil {
		_, err = cstream.CloseAndRecv()
	}
	EchoApisTestCheckErr(t, err, errMsg)
	// Bidir streaming
	span, cctx = log.ChildSpan(ctx, log.DebugLevelApi, "stream-request")
	defer span.Finish()
	bstream, err := client.BidirectionalStreamingEcho(cctx)
	if err == nil {
		_, err = bstream.Recv()
	}
	EchoApisTestCheckErr(t, err, errMsg)
}

func EchoApisTestCheckErr(t *testing.T, err error, errMsg string) {
	if errMsg == "" {
		require.Nil(t, err)
	} else {
		require.NotNil(t, err)
		require.Contains(t, err.Error(), errMsg)
	}
}

type EchoServer struct{}

func (s *EchoServer) UnaryEcho(ctx context.Context, req *echo.EchoRequest) (*echo.EchoResponse, error) {
	return &echo.EchoResponse{}, nil
}

func (s *EchoServer) ServerStreamingEcho(req *echo.EchoRequest, stream echo.Echo_ServerStreamingEchoServer) error {
	return stream.Send(&echo.EchoResponse{})
}

func (s *EchoServer) ClientStreamingEcho(stream echo.Echo_ClientStreamingEchoServer) error {
	return stream.SendAndClose(&echo.EchoResponse{})
}

func (s *EchoServer) BidirectionalStreamingEcho(stream echo.Echo_BidirectionalStreamingEchoServer) error {
	return stream.Send(&echo.EchoResponse{})
}

type DummyController struct {
	Cache         edgeproto.CloudletCache
	Cloudlets     map[edgeproto.CloudletKey]*TestCloudlet
	KeyServer     *AccessKeyServer
	ApiLis        net.Listener
	ApiServ       *grpc.Server
	ApiRegisterCb func(server *grpc.Server)
}

func (s *DummyController) Init() {
	edgeproto.InitCloudletCache(&s.Cache)
	s.Cloudlets = make(map[edgeproto.CloudletKey]*TestCloudlet)
	s.KeyServer = NewAccessKeyServer(&s.Cache)
}

// DummyController. The optional registerCb func allows the caller to register
// more grpc handlers.
func (s *DummyController) Start() {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err.Error())
	}
	s.ApiLis = lis
	// The "required" stream interceptors, which always requires a
	// valid access key.
	s.ApiServ = grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			cloudcommon.AuditUnaryInterceptor,
			s.KeyServer.UnaryRequireAccessKey,
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			cloudcommon.AuditStreamInterceptor,
			s.KeyServer.StreamRequireAccessKey,
		)),
	)
	edgeproto.RegisterCloudletAccessKeyApiServer(s.ApiServ, s)
	if s.ApiRegisterCb != nil {
		s.ApiRegisterCb(s.ApiServ)
	}
	go func() {
		err := s.ApiServ.Serve(s.ApiLis)
		if err != nil {
			panic(err.Error())
		}
	}()
}

func (s *DummyController) Stop() {
	s.ApiServ.Stop()
	s.ApiLis.Close()
}

func (s *DummyController) ApiAddr() string {
	return s.ApiLis.Addr().String()
}

func (s *DummyController) UpgradeAccessKey(stream edgeproto.CloudletAccessKeyApi_UpgradeAccessKeyServer) error {
	return s.KeyServer.UpgradeAccessKey(stream, s.commitKey)
}

func (s *DummyController) commitKey(ctx context.Context, key *edgeproto.CloudletKey, pubPEM string) error {
	tc, ok := s.Cloudlets[*key]
	if !ok {
		return fmt.Errorf("test cloudlet key %v not found", key)
	}
	tc.Cloudlet.CrmAccessPublicKey = pubPEM
	tc.Cloudlet.CrmAccessKeyUpgradeRequired = false
	s.Cache.Update(ctx, &tc.Cloudlet, 0)
	return nil
}

// TestCloudlet is data for the client
type TestCloudlet struct {
	Cloudlet       edgeproto.Cloudlet
	privateKeyPEM  string
	privateKeyFile string
	KeyClient      *AccessKeyClient
}

// CreateCloudlet creates test client data.
func (s *DummyController) CreateCloudlet(ctx context.Context, name string) *TestCloudlet {
	tc := &TestCloudlet{}
	tc.Cloudlet.Key.Name = name
	tc.Cloudlet.Key.Organization = "testorg"
	tc.privateKeyFile = "/tmp/accesskey_unittest_" + name
	s.Cloudlets[tc.Cloudlet.Key] = tc
	s.Cache.Update(ctx, &tc.Cloudlet, 0)

	keyClient := &AccessKeyClient{}
	keyClient.AccessKeyFile = tc.privateKeyFile
	keyClient.AccessApiAddr = s.ApiLis.Addr().String()
	keyClient.TestNoTls = true
	tc.KeyClient = keyClient
	// clear out any existing key files left over by previous (failed) tests
	tc.Cleanup()
	return tc
}

func (s *TestCloudlet) Cleanup() {
	os.Remove(s.privateKeyFile)
	os.Remove(s.KeyClient.backupKeyFile())
}

func (s *DummyController) UpdateKey(ctx context.Context, key edgeproto.CloudletKey) error {
	tc, found := s.Cloudlets[key]
	if !found {
		return fmt.Errorf("cloudlet %v not found", key)
	}

	// set up access key pair
	keyPair, err := GenerateAccessKey()
	if err != nil {
		return err
	}

	tc.privateKeyPEM = keyPair.PrivatePEM
	tc.Cloudlet.CrmAccessPublicKey = keyPair.PublicPEM
	// put cloudlet in cache
	s.Cache.Update(ctx, &tc.Cloudlet, 0)
	// write private key to disk
	return ioutil.WriteFile(tc.privateKeyFile, []byte(tc.privateKeyPEM), 0600)
}
