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
	fmt.Printf("FindCloudlet: Found Sewrvice IP %s\n", ipaddr.String())
	return mreq, nil
}

func (s *server) VerifyLocation(ctx context.Context,
	req *dme.Match_Engine_Request) (*dme.Match_Engine_Loc_Verify, error) {
	
	log.Printf("VerifyLocation - Got Version: %d", req.Ver)
	var loc = &dme.Match_Engine_Loc_Verify{}
	loc.Ver = 6
	return loc, nil
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
