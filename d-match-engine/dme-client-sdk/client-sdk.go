package main


import (
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

const (
	//port = "192.168.1.27:50051"
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

	client := dme.NewMatch_Engine_ApiClient(conn)


	_, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()


	find_cloudlets(client)
	TestLocations(client)
}
