package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
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

var carrier = flag.String("carrier", "standalone", "carrier name for API connection, or standalone for internal DME")

// server is used to implement helloworld.GreeterServer.
type server struct{}

func verifyCookie(ctx context.Context, sessionCookie string) (int, error, string) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return -1, errors.New("unable to get peer IP info"), ""
	}
	//peer address is ip:port
	ss := strings.Split(p.Addr.String(), ":")
	if len(ss) != 2 {
		return -1, errors.New("unable to parse peer address " + p.Addr.String()), ""
	}
	peerIp := ss[0]

	// This will be encrypted on our public key and will need to be decrypted
	// For now just verify the clear txt IP
	fmt.Printf("SessionCookie is %s\n", sessionCookie)
	if sessionCookie != peerIp {
		return -1, errors.New("unable to verify SessionCookie"), ""
	}
	return 0, nil, peerIp
}

func (s *server) FindCloudlet(ctx context.Context, req *dme.Match_Engine_Request) (*dme.Match_Engine_Reply,
	error) {

	ok, err, _ := verifyCookie(ctx, req.SessionCookie)
	if ok != 0 {
		return nil, err
	}

	mreq := new(dme.Match_Engine_Reply)
	findCloudlet(req, mreq)
	return mreq, nil
}

func (s *server) VerifyLocation(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Loc_Verify, error) {

	var mreq *dme.Match_Engine_Loc_Verify
	mreq = new(dme.Match_Engine_Loc_Verify)

	ok, err, peerIp := verifyCookie(ctx, req.SessionCookie)
	if ok != 0 {
		return nil, err
	}

	VerifyClientLoc(req, mreq, *carrier, peerIp, *locVerUrl)
	return mreq, nil
}

func (s *server) GetLocation(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Loc, error) {

	var mloc *dme.Match_Engine_Loc
	mloc = new(dme.Match_Engine_Loc)

	ok, err, _ := verifyCookie(ctx, req.SessionCookie)
	if ok != 0 {
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

	mstatus.TokenServerURI = *tokSrvUrl

	// Set the src IP as the session cookie for now so we can verify the client later
	// without needing to call into the operator backend and also have some context
	// when subsequent call comes. Todo: this part will get enhanced as we go along. For
	// now teach the sdk to store the cookie and send it in each subsequent calls
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("unable to get peer IP info")
	}
	//peer address is ip:port
	ss := strings.Split(p.Addr.String(), ":")
	if len(ss) != 2 {
		return nil, errors.New("unable to parse peer address " + p.Addr.String())
	}
	peerIp := ss[0]

	// For now, just send the unencrypoted cookie back
	// Fix me to return a cookie encryoted on DME public key
	mstatus.SessionCookie = peerIp
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
		notifyClient := initNotifyClient(*notifyAddrs)
		notifyClient.Start()
		defer notifyClient.Stop()
	}

	lis, err := net.Listen("tcp", *apiAddr)
	if err != nil {
		log.FatalLog("Failed to listen", "addr", *apiAddr, "err", err)
	}
	s := grpc.NewServer()
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
