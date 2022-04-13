// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// The Group is created as an AppInst from the Controller
// The AddUser call comes from Client to DME and somehow needs to be communicated
// to this guy

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
