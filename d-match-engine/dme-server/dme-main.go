package main

import (
	"flag"
	"fmt"
	"net"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/tls"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Command line options
var rootDir = flag.String("r", "", "root directory for testing")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var apiAddr = flag.String("apiAddr", "localhost:50051", "API listener address")
var standalone = flag.Bool("standalone", false, "Standalone mode. AppInst data is pre-populated. Dme does not interact with controller. AppInsts can be created directly on Dme using controller AppInst API")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var locVerUrl = flag.String("locverurl", "", "location verification REST API URL to connect to")
var tokSrvUrl = flag.String("toksrvurl", "", "token service URL to provide to client on register")
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")

var carrier = flag.String("carrier", "standalone", "carrier name for API connection, or standalone for internal DME")

// server is used to implement helloworld.GreeterServer.
type server struct{}

func (s *server) FindCloudlet(ctx context.Context, req *dme.Match_Engine_Request) (*dme.Match_Engine_Reply,
	error) {

	_, err := dmecommon.VerifyCookie(req.SessionCookie)
	if err != nil {
		return nil, err
	}

	mreq := new(dme.Match_Engine_Reply)
	findCloudlet(req, mreq)
	return mreq, nil
}

func (s *server) GetCloudlets(ctx context.Context, req *dme.Match_Engine_Request) (*dme.Match_Engine_Cloudlet_List, error) {
	_, err := dmecommon.VerifyCookie(req.SessionCookie)
	if err != nil {
		return nil, err
	}
	if req.GpsLocation == nil {
		log.DebugLog(log.DebugLevelDmereq, "Invalid GetCloudlets request", "Error", "Missing GpsLocation")
		return nil, fmt.Errorf("missing GPS location")
	}
	clist := new(dme.Match_Engine_Cloudlet_List)
	getCloudlets(req, clist)
	return clist, nil
}

func (s *server) VerifyLocation(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Loc_Verify, error) {

	var mreq *dme.Match_Engine_Loc_Verify
	mreq = new(dme.Match_Engine_Loc_Verify)

	peerIp, err := dmecommon.VerifyCookie(req.SessionCookie)
	if err != nil {
		return nil, err
	}

	err = VerifyClientLoc(req, mreq, *carrier, peerIp, *locVerUrl)
	if err != nil {
		return nil, err
	}
	return mreq, nil
}

func (s *server) GetLocation(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Loc, error) {

	var mloc *dme.Match_Engine_Loc
	mloc = new(dme.Match_Engine_Loc)

	_, err := dmecommon.VerifyCookie(req.SessionCookie)
	if err != nil {
		return nil, err
	}

	GetClientLoc(req, mloc)
	if mloc.Status == dme.Match_Engine_Loc_LOC_FOUND {
		fmt.Printf("GetLocation: Found Location\n")
	} else {
		fmt.Printf("GetLocation: Location NOT Found\n")
	}

	return mloc, nil
}

func (s *server) RegisterClient(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Status, error) {

	var mstatus *dme.Match_Engine_Status
	mstatus = new(dme.Match_Engine_Status)

	cookie, err := dmecommon.GenerateCookie(req.AppName, ctx)
	if err != nil {
		mstatus.Status = dme.Match_Engine_Status_ME_FAIL
		return mstatus, err
	}
	mstatus.SessionCookie = cookie
	mstatus.TokenServerURI = *tokSrvUrl
	mstatus.Status = dme.Match_Engine_Status_ME_SUCCESS
	return mstatus, nil
}

func (s *server) AddUserToGroup(ctx context.Context,
	req *dme.DynamicLocGroupAdd) (*dme.Match_Engine_Status, error) {

	var mreq *dme.Match_Engine_Status
	mreq = new(dme.Match_Engine_Status)
	mreq.Status = dme.Match_Engine_Status_ME_SUCCESS

	return mreq, nil
}

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)

	setupMatchEngine()

	if *standalone {
		fmt.Printf("Running in Standalone Mode with test instances\n")
		appInsts := dmetest.GenerateAppInsts()
		for _, inst := range appInsts {
			addApp(inst)
		}
		listAppinstTbl()
	} else {
		notifyClient := initNotifyClient(*notifyAddrs, *tlsCertFile)
		notifyClient.Start()
		defer notifyClient.Stop()
	}

	lis, err := net.Listen("tcp", *apiAddr)
	if err != nil {
		log.FatalLog("Failed to listen", "addr", *apiAddr, "err", err)
	}

	creds, err := tls.GetTLSServerCreds(*tlsCertFile)
	if err != nil {
		log.FatalLog("get TLS Credentials", "error", err)
	}
	s := grpc.NewServer(grpc.Creds(creds))

	dme.RegisterMatch_Engine_ApiServer(s, &server{})

	if *standalone {
		saServer := standaloneServer{}
		edgeproto.RegisterAppInstApiServer(s, &saServer)
	}

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.FatalLog("Failed to server", "err", err)
	}
}
