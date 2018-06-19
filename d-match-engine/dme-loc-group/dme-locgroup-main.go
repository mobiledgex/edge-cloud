package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Command line options
var rootDir = flag.String("r", "", "root directory for testing")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var apiAddr = flag.String("apiAddr", "0.0.0.0:50058", "API listener address")

// server is used to implement helloworld.GreeterServer.
type server struct{}

func (s *server) SendToGroup(ctx context.Context, req *dme.DlgMessage) (*dme.DlgReply,
	error) {

	var mreq *dme.DlgReply

	fmt.Printf("SendToGroup: To Group %d\n", req.LgId);
	mreq = new(dme.DlgReply)
	mreq.AckId = req.MessageId
	return mreq, nil
}


func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *apiAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	dme.RegisterDynamicLocGroupApiServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
