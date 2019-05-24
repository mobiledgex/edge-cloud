package main

import (
	"log"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
	version = 1
)

func main() {
	// Set up a connection to the server
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := dme.NewMatchEngineApiClient(conn)

	_, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	FindCloudlets(client)
	TestLocations(client)
}
