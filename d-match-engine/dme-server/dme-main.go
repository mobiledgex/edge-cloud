package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Command line options
var rootDir = flag.String("r", "", "root directory for testing")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var apiAddr = flag.String("apiAddr", "0.0.0.0:50051", "API listener address")

// server is used to implement helloworld.GreeterServer.
type server struct{}

func (s *server) FindCloudlet(ctx context.Context, req *dme.Match_Engine_Request) (*dme.Match_Engine_Reply,
	error) {

	var mreq *dme.Match_Engine_Reply
	var ipaddr net.IP

	mreq = new(dme.Match_Engine_Reply)
	find_cloudlet(req, mreq)
	ipaddr = mreq.ServiceIp
	fmt.Printf("FindCloudlet: Found Service IP %s\n", ipaddr.String())
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

func (s *server) AddUserToGroup(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Status, error) {

	var mreq *dme.Match_Engine_Status
	mreq = new(dme.Match_Engine_Status)
	mreq.Status = 0

	return mreq, nil
}

func main() {
	flag.Parse()

	setup_match_engine()

	notifyHandler := &NotifyHandler{}
	notifyClient := notify.NewDMEClient(strings.Split(*notifyAddrs, ","), notifyHandler)
	go notifyClient.Run()
	defer notifyClient.Stop()
	util.InfoLog("notify client to", "addrs", *notifyAddrs)

	lis, err := net.Listen("tcp", *apiAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	dme.RegisterMatch_Engine_ApiServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
