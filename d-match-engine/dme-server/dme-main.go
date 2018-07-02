package main

import (
	"flag"
	"fmt"
	"net"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Command line options
var rootDir = flag.String("r", "", "root directory for testing")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var apiAddr = flag.String("apiAddr", "0.0.0.0:50051", "API listener address")
var standalone = flag.Bool("standalone", false, "Standalone mode. AppInst data is pre-populated. Dme does not interact with controller. AppInsts can be created directly on Dme using controller AppInst API")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))

// server is used to implement helloworld.GreeterServer.
type server struct{}

func (s *server) FindCloudlet(ctx context.Context, req *dme.Match_Engine_Request) (*dme.Match_Engine_Reply,
	error) {

	mreq := new(dme.Match_Engine_Reply)
	findCloudlet(req, mreq)
	return mreq, nil
}

func (s *server) VerifyLocation(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Loc_Verify, error) {

	var mreq *dme.Match_Engine_Loc_Verify
	mreq = new(dme.Match_Engine_Loc_Verify)
	VerifyClientLoc(req, mreq)
	return mreq, nil
}

func (s *server) GetLocation(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Loc, error) {

	var mloc *dme.Match_Engine_Loc
	mloc = new(dme.Match_Engine_Loc)
	//Todo: Implement the function to actually get the location
	GetClientLoc(req, mloc)
	if mloc.Status == 1 {
		fmt.Printf("GetLocation: Found Location\n")
	} else {
		fmt.Printf("GetLocation: Location NOT Found\n")
	}

	return mloc, nil
}

func (s *server) RegisterClient(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Status, error) {

	//Todo: Implement the Reqister client/token Function
	var mstatus *dme.Match_Engine_Status
	mstatus = new(dme.Match_Engine_Status)
	mstatus.Status = 0

	return mstatus, nil
}

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)

	setupMatchEngine()

	if *standalone {
		appInsts := dmetest.GenerateAppInsts()
		for _, inst := range appInsts {
			addApp(inst)
		}
	} else {
		notifyHandler := &NotifyHandler{}
		notifyClient := initNotifyClient(*notifyAddrs, notifyHandler)
		go notifyClient.Run()
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
