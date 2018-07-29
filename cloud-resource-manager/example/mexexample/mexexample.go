package main

import (
	"flag"
	"fmt"
	log "gitlab.com/bobbae/logrus"
	"os"
	"os/signal"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/example/mexexample/server"
)

var sigChan chan os.Signal

func main() {
	fmt.Println(os.Args)
	debug := flag.Bool("debug", false, "debug")
	grpcAddress := flag.String("grpc", ":27272", "GRPC address")
	restAddress := flag.String("rest", ":27273", "REST API address")
	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	log.Debugf("starting HTTP Server at %s", *restAddress)
	go func() {
		err := server.ListenAndServeREST(*restAddress, *grpcAddress)
		if err != nil {
			log.Fatalf("cannot run HTTP server, %v", err)
		}
	}()
	log.Debugf("starting GRPC Server at %s", *grpcAddress)
	go func() {
		if err := server.ListenAndServeGRPC(*grpcAddress); err != nil {
			log.Fatalf("cannot run GRPC server, %v", err)
		}
	}()
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	<-sigChan
	os.Exit(0)
}
