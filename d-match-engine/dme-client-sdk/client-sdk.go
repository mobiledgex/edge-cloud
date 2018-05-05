package main


import (
	"log"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	dme "../dme-proto"
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

	defaultversion := version
	if len(os.Args) > 1 {
		defaultversion = int(os.Args[1][0])
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var req = &dme.Match_Engine_Request{};
	req.Ver = uint32(defaultversion);
	resp, err := client.FindCloudlet(ctx, req);
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Got Version: %d", resp.Ver)
}
