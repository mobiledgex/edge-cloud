package main

import (
	"fmt"
	"flag"
	"log"
	"net"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	//port = "192.168.1.27:50051"
	port = ":50051"
)

// Command line options
var rootDir = flag.String("r", "", "root directory for testing")
var notifyAddr = flag.String("notifyAddr", "127.0.0.1:50001", "Notify listener address")

// server is used to implement helloworld.GreeterServer.
type server struct{}

func (s *server) FindCloudlet(ctx context.Context, req *dme.Match_Engine_Request) (*dme.Match_Engine_Reply,
	error) {
	
	var mreq *dme.Match_Engine_Reply;
	var ipaddr net.IP

	mreq = new (dme.Match_Engine_Reply)
	find_cloudlet(req, mreq)
	ipaddr = mreq.ServiceIp
	fmt.Printf("FindCloudlet: Found Service IP %s\n", ipaddr.String())
	return mreq, nil
}

func (s *server) VerifyLocation(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Loc_Verify, error) {

	var mreq *dme.Match_Engine_Loc_Verify;
	mreq = new (dme.Match_Engine_Loc_Verify)
	VerifyClientLoc(req, mreq)
	return mreq, nil
}

func (s *server) GetLocation(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Loc, error) {

	var mloc *dme.Match_Engine_Loc;
	mloc = new (dme.Match_Engine_Loc)
	//Todo: Implement the function to actually get the location
	GetClientLoc(req, mloc)
	if (mloc.Status == 1) {
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
	mstatus = new (dme.Match_Engine_Status)
	mstatus.Status = 0

	return mstatus, nil
}
	

func main() {
	flag.Parse()

	setup_match_engine()
	
	recvHandler := &NotifyHandler{}
	recv := notify.NewNotifyReceiver("tcp", *notifyAddr, recvHandler)
	go recv.Run()
	defer recv.Stop()
	util.InfoLog("notify listener", "addr", *notifyAddr)

	lis, err := net.Listen("tcp", port)
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
