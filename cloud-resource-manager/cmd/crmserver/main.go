package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/bobbae/q"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var bindAddress = flag.String("bind", "0.0.0.0:55099", "Address to bind")
var controllerAddress = flag.String("controller", "127.0.0.1:55001", "Address of controller API")
var notifyAddr = flag.String("notifyAddr", "127.0.0.1:51001", "Notify listener address")

var sigChan chan os.Signal

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", *bindAddress)
	if err != nil {
		log.Fatalf("Failed to bind to %v, %v", *bindAddress, err)
	}

	controllerData := crmutil.NewControllerData()
	srv, err := crmutil.NewCloudResourceManagerServer(controllerData)

	grpcServer := grpc.NewServer()
	edgeproto.RegisterCloudResourceManagerServer(grpcServer, srv)

	recvHandler := NewNotifyHandler(controllerData)
	recv := notify.NewNotifyReceiver("tcp", *notifyAddr, recvHandler)
	go recv.Run()
	defer recv.Stop()
	q.Q("running notify handler at %v", *notifyAddr)

	q.Q("registered CRM API server")
	reflection.Register(grpcServer)

	go func() {
		q.Q("running grpc server")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve grpc, %v", err)
		}
	}()
	defer grpcServer.Stop()

	log.Printf("Server started at %v", *bindAddress)

	conn, err := grpc.Dial(*controllerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to controller %v, %v", *controllerAddress, err)
	}
	defer conn.Close()

	// Cloudlet, Operator, AppInst data should come from controller
	// via notify API, but may be pre-populated here programmatically
	// for testing.
	cloudletAPI := edgeproto.NewCloudletApiClient(conn)
	operatorAPI := edgeproto.NewOperatorApiClient(conn)

	log.Printf("Registering Operators...")

	err = crmutil.RegisterOperators(operatorAPI)
	if err != nil {
		log.Printf("Can't register operators, %v", err)
	}

	log.Printf("Cloudlet API client at %v", *controllerAddress)

	go func() {
		if err := crmutil.CloudletClient(cloudletAPI, srv); err != nil {
			log.Printf("client API error, %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c

	os.Exit(0)
}
