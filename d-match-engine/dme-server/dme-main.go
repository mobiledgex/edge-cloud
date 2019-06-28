package main

import (
	ctls "crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	baselog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	log "github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
)

// Command line options
var rootDir = flag.String("r", "", "root directory for testing")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var apiAddr = flag.String("apiAddr", "localhost:50051", "API listener address")
var httpAddr = flag.String("httpAddr", "127.0.0.1:38001", "HTTP listener address")
var standalone = flag.Bool("standalone", false, "Standalone mode. AppInst data is pre-populated. Dme does not interact with controller. AppInsts can be created directly on Dme using controller AppInst API")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var locVerUrl = flag.String("locverurl", "", "location verification REST API URL to connect to")
var tokSrvUrl = flag.String("toksrvurl", "", "token service URL to provide to client on register")
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var tlsApiCertFile = flag.String("tlsApiCertFile", "", "Public-CA signed TLS cert file for serving DME APIs")
var tlsApiKeyFile = flag.String("tlsApiKeyFile", "", "Public-CA signed TLS key file for serving DME APIs")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")
var scaleID = flag.String("scaleID", "", "ID to distinguish multiple DMEs in the same cloudlet. Defaults to hostname if unspecified.")
var vaultAddr = flag.String("vaultAddr", "http://127.0.0.1:8200", "Vault address")
var statsInterval = flag.Int("statsInterval", 1, "interval in seconds between sending stats")
var statsShards = flag.Uint("statsShards", 10, "number of shards (locks) in memory for parallel stat collection")
var cookieExpiration = flag.Duration("cookieExpiration", time.Hour*24, "Cookie expiration time")
var region = flag.String("region", "local", "region name")

// TODO: carrier arg is redundant with OperatorKey.Name in myCloudletKey, and
// should be replaced by it, but requires dealing with carrier-specific
// verify location API behavior and e2e test setups.
var carrier = flag.String("carrier", "standalone", "carrier name for API connection, or standalone for internal DME")

// server is used to implement helloworld.GreeterServer.
type server struct{}

// myCloudlet is the information for the cloudlet in which the DME is instantiated.
// The key for myCloudlet is provided as a configuration - either command line or
// from a file.
var myCloudletKey edgeproto.CloudletKey
var myNode edgeproto.Node

var sigChan chan os.Signal

func (s *server) FindCloudlet(ctx context.Context, req *dme.FindCloudletRequest) (*dme.FindCloudletReply, error) {
	reply := new(dme.FindCloudletReply)
	ckey, ok := dmecommon.CookieFromContext(ctx)
	if !ok {
		return reply, grpc.Errorf(codes.InvalidArgument, "No valid session cookie")
	}
	if req.CarrierName == "" {
		log.DebugLog(log.DebugLevelDmereq, "Invalid FindCloudlet request", "Error", "Missing CarrierName")

		return reply, grpc.Errorf(codes.InvalidArgument, "Missing carrierName")
	}
	if req.GpsLocation == nil || (req.GpsLocation.Latitude == 0 && req.GpsLocation.Longitude == 0) {
		log.DebugLog(log.DebugLevelDmereq, "Invalid FindCloudlet request", "Error", "Missing GpsLocation")
		return reply, grpc.Errorf(codes.InvalidArgument, "Missing GpsLocation")
	}

	if !util.IsLatitudeValid(req.GpsLocation.Latitude) || !util.IsLongitudeValid(req.GpsLocation.Longitude) {
		log.DebugLog(log.DebugLevelDmereq, "Invalid FindCloudlet GpsLocation", "lat", req.GpsLocation.Latitude, "long", req.GpsLocation.Longitude)
		return reply, grpc.Errorf(codes.InvalidArgument, "Invalid GpsLocation")
	}

	err := findCloudlet(ckey, req, reply)
	log.DebugLog(log.DebugLevelDmereq, "FindCloudlet returns", "reply", reply, "error", err)
	return reply, err
}

func (s *server) GetFqdnList(ctx context.Context, req *dme.FqdnListRequest) (*dme.FqdnListReply, error) {
	log.DebugLog(log.DebugLevelDmereq, "GetFqdnList", "req", req)
	flist := new(dme.FqdnListReply)

	ckey, ok := dmecommon.CookieFromContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.InvalidArgument, "No valid session cookie")
	}
	// normal applications are not allowed to access this, only special platform developer/app combos
	if !cloudcommon.IsPlatformApp(ckey.DevName, ckey.AppName) {
		return nil, grpc.Errorf(codes.PermissionDenied, "API Not allowed for developer: %s app: %s", ckey.DevName, ckey.AppName)
	}

	getFqdnList(req, flist)
	log.DebugLog(log.DebugLevelDmereq, "GetFqdnList returns", "status", flist.Status)
	return flist, nil
}

func (s *server) GetAppInstList(ctx context.Context, req *dme.AppInstListRequest) (*dme.AppInstListReply, error) {
	ckey, ok := dmecommon.CookieFromContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.InvalidArgument, "No valid session cookie")
	}

	log.DebugLog(log.DebugLevelDmereq, "GetAppInstList", "carrier", req.CarrierName, "ckey", ckey)

	if req.GpsLocation == nil {
		log.DebugLog(log.DebugLevelDmereq, "Invalid GetAppInstList request", "Error", "Missing GpsLocation")
		return nil, grpc.Errorf(codes.InvalidArgument, "Missing GPS location")
	}
	alist := new(dme.AppInstListReply)
	getAppInstList(ckey, req, alist)
	log.DebugLog(log.DebugLevelDmereq, "GetAppInstList returns", "status", alist.Status)
	return alist, nil
}

func (s *server) VerifyLocation(ctx context.Context,
	req *dme.VerifyLocationRequest) (*dme.VerifyLocationReply, error) {

	reply := new(dme.VerifyLocationReply)

	ckey, ok := dmecommon.CookieFromContext(ctx)
	if !ok {
		return reply, grpc.Errorf(codes.InvalidArgument, "No valid session cookie")
	}
	err := VerifyClientLoc(req, reply, *carrier, ckey, *locVerUrl, *tokSrvUrl)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *server) GetLocation(ctx context.Context,
	req *dme.GetLocationRequest) (*dme.GetLocationReply, error) {

	reply := new(dme.GetLocationReply)

	GetClientLoc(req, reply)
	if reply.Status == dme.GetLocationReply_LOC_FOUND {
		fmt.Printf("GetLocation: Found Location\n")
	} else {
		fmt.Printf("GetLocation: Location NOT Found\n")
	}

	return reply, nil
}

func getAuthPublicKey(devname string, appname string, appvers string) (string, error) {
	var key edgeproto.AppKey
	var tbl *dmeApps
	tbl = dmeAppTbl

	key.DeveloperKey.Name = devname
	key.Name = appname
	key.Version = appvers
	tbl.Lock()
	defer tbl.Unlock()

	app, ok := tbl.apps[key]
	if ok {
		return app.authPublicKey, nil
	}
	return "", grpc.Errorf(codes.NotFound, "app not found")
}

func (s *server) RegisterClient(ctx context.Context,
	req *dme.RegisterClientRequest) (*dme.RegisterClientReply, error) {

	mstatus := new(dme.RegisterClientReply)

	log.DebugLog(log.DebugLevelDmereq, "RegisterClient received", "request", req)

	if req.DevName == "" {
		log.DebugLog(log.DebugLevelDmereq, "DevName cannot be empty")
		mstatus.Status = dme.ReplyStatus_RS_FAIL
		return mstatus, grpc.Errorf(codes.InvalidArgument, "DevName cannot be empty")
	}
	if req.AppName == "" {
		log.DebugLog(log.DebugLevelDmereq, "AppName cannot be empty")
		mstatus.Status = dme.ReplyStatus_RS_FAIL
		return mstatus, grpc.Errorf(codes.InvalidArgument, "AppName cannot be empty")
	}
	if req.AppVers == "" {
		log.DebugLog(log.DebugLevelDmereq, "AppVers cannot be empty")
		mstatus.Status = dme.ReplyStatus_RS_FAIL
		return mstatus, grpc.Errorf(codes.InvalidArgument, "AppVers cannot be empty")
	}
	authkey, err := getAuthPublicKey(req.DevName, req.AppName, req.AppVers)
	if err != nil {
		log.DebugLog(log.DebugLevelDmereq, "fail to get public key", "err", err)
		mstatus.Status = dme.ReplyStatus_RS_FAIL
		return mstatus, err
	}

	//the token is currently optional, but once the SDK is enhanced to send one, it should
	// be a mandatory parameter.  For now, only validate the token if we receive one
	if req.AuthToken == "" {
		if authkey != "" {
			// we provisioned a key, and one was not provided.
			log.DebugLog(log.DebugLevelDmereq, "App has key, no token received")
			mstatus.Status = dme.ReplyStatus_RS_FAIL
			return mstatus, grpc.Errorf(codes.InvalidArgument, "No authtoken received")
		}
		// for now we will allow a tokenless register to pass if the app does not have one
		log.DebugLog(log.DebugLevelDmereq, "Allowing register without token")

	} else {
		if authkey == "" {
			log.DebugLog(log.DebugLevelDmereq, "No authkey provisioned to validate token")
			mstatus.Status = dme.ReplyStatus_RS_FAIL
			return mstatus, grpc.Errorf(codes.Unauthenticated, "No authkey found to validate token")
		}
		err := dmecommon.VerifyAuthToken(req.AuthToken, authkey, req.DevName, req.AppName, req.AppVers)
		if err != nil {
			log.DebugLog(log.DebugLevelDmereq, "Failed to verify token", "err", err)
			mstatus.Status = dme.ReplyStatus_RS_FAIL
			return mstatus, grpc.Errorf(codes.Unauthenticated, "failed to verify token - %s", err.Error())
		}
	}
	key := dmecommon.CookieKey{
		DevName: req.DevName,
		AppName: req.AppName,
		AppVers: req.AppVers,
	}
	cookie, err := dmecommon.GenerateCookie(&key, ctx, cookieExpiration)
	if err != nil {
		return mstatus, grpc.Errorf(codes.Internal, err.Error())
	}
	mstatus.SessionCookie = cookie
	mstatus.TokenServerUri = *tokSrvUrl
	mstatus.Status = dme.ReplyStatus_RS_SUCCESS
	return mstatus, nil
}

func (s *server) AddUserToGroup(ctx context.Context,
	req *dme.DynamicLocGroupRequest) (*dme.DynamicLocGroupReply, error) {

	mreq := new(dme.DynamicLocGroupReply)
	mreq.Status = dme.ReplyStatus_RS_SUCCESS

	return mreq, nil
}

func (s *server) GetQosPositionKpi(ctx context.Context,
	req *dme.QosPositionKpiRequest) (*dme.QosPositionKpiReply, error) {
	log.DebugLog(log.DebugLevelDmereq, "GetQosPositionKpi", "request", req)

	mrep := new(dme.QosPositionKpiReply)
	getQosPositionKpi(req, mrep)
	return mrep, nil
}

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	cloudcommon.ParseMyCloudletKey(*standalone, cloudletKeyStr, &myCloudletKey)
	cloudcommon.SetNodeKey(scaleID, edgeproto.NodeType_NODE_DME, &myCloudletKey, &myNode.Key)

	dmecommon.InitVault(*vaultAddr, *region)

	setupMatchEngine()
	grpcOpts := make([]grpc.ServerOption, 0)

	if *standalone {
		fmt.Printf("Running in Standalone Mode with test instances\n")
		for _, app := range dmetest.GenerateApps() {
			addApp(app)
		}
		for _, inst := range dmetest.GenerateAppInsts() {
			addAppInst(inst)
		}
		listAppinstTbl()
	} else {
		notifyClient := initNotifyClient(*notifyAddrs, *tlsCertFile)
		sendMetric := notify.NewMetricSend()
		notifyClient.RegisterSend(sendMetric)

		notifyClient.Start()
		defer notifyClient.Stop()

		interval := time.Duration(*statsInterval) * time.Second
		stats := NewDmeStats(interval, *statsShards, sendMetric.Update)
		stats.Start()
		defer stats.Stop()
		grpcOpts = append(grpcOpts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(dmecommon.UnaryAuthInterceptor, stats.UnaryStatsInterceptor)))
	}
	nodeCache.Update(&myNode, 0)

	lis, err := net.Listen("tcp", *apiAddr)
	if err != nil {
		log.FatalLog("Failed to listen", "addr", *apiAddr, "err", err)
	}

	creds, err := tls.ServerAuthServerCreds(*tlsApiCertFile, *tlsApiKeyFile)
	if err != nil {
		log.FatalLog("get TLS Credentials", "error", err)
	}
	grpcOpts = append(grpcOpts, grpc.Creds(creds))
	s := grpc.NewServer(grpcOpts...)

	dme.RegisterMatchEngineApiServer(s, &server{})

	if *standalone {
		saServer := standaloneServer{}
		edgeproto.RegisterAppApiServer(s, &saServer)
		edgeproto.RegisterAppInstApiServer(s, &saServer)
	}

	// Register reflection service on gRPC server.
	reflection.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.FatalLog("Failed to server", "err", err)
		}
	}()
	defer s.Stop()

	// REST service
	mux := http.NewServeMux()
	gwcfg := &cloudcommon.GrpcGWConfig{
		ApiAddr:     *apiAddr,
		TlsCertFile: *tlsCertFile,
		ApiHandles: []func(context.Context, *gwruntime.ServeMux, *grpc.ClientConn) error{
			dme.RegisterMatchEngineApiHandler,
		},
	}
	gw, err := cloudcommon.GrpcGateway(gwcfg)
	if err != nil {
		log.FatalLog("Failed to start grpc Gateway", "err", err)
	}
	mux.Handle("/", gw)
	tlscfg := &ctls.Config{
		MinVersion:               ctls.VersionTLS12,
		CurvePreferences:         []ctls.CurveID{ctls.CurveP521, ctls.CurveP384, ctls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			ctls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			ctls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			ctls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			ctls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			ctls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	// Suppress contant stream of TLS error logs due to LB health check. There is discussion in the community
	//to get rid of some of these logs, but as of now this a the way around it.   We could miss other logs here but
	// the excessive error logs are drowning out everthing else.
	var nullLogger baselog.Logger
	nullLogger.SetOutput(ioutil.Discard)

	httpServer := &http.Server{
		Addr:      *httpAddr,
		Handler:   mux,
		TLSConfig: tlscfg,
		ErrorLog:  &nullLogger,
	}

	go cloudcommon.GrpcGatewayServe(gwcfg, httpServer)
	defer httpServer.Shutdown(context.Background())

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	log.InfoLog("Ready")

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}
