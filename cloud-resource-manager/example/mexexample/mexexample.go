package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/example/mexexample/server"
)

var sigChan chan os.Signal

func main() {
	fmt.Println(os.Args)
	debug := flag.Bool("debug", false, "debug")
	grpcAddress := flag.String("grpc", ":27272", "GRPC address")
	restAddress := flag.String("rest", ":27273", "REST API address")
	httpAddress := flag.String("http", ":27274", "HTTP address")
	tcpAddress := flag.String("tcp", ":27275", "TCP address")
	udpAddress := flag.String("udp", ":27276", "UDP address")
	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	log.Debugln("starting HTTP Server at", *httpAddress)
	http.HandleFunc("/", frontpage)
	go func() {
		err := http.ListenAndServe(*httpAddress, nil)
		if err != nil {
			log.Fatal("cannot run HTTP Server", err)
		}
	}()
	log.Debugf("starting REST Server at %s", *restAddress) //really just POST
	go func() {
		err := server.ListenAndServeREST(*restAddress, *grpcAddress)
		if err != nil {
			log.Fatalf("cannot run REST server, %v", err)
		}
	}()
	log.Debugf("starting GRPC Server at %s", *grpcAddress)
	go func() {
		if err := server.ListenAndServeGRPC(*grpcAddress); err != nil {
			log.Fatalf("cannot run GRPC server, %v", err)
		}
	}()
	log.Debugf("starting TCP Server at %s", *tcpAddress)
	go func() {
		if err := server.ListenAndServeTCP(*tcpAddress); err != nil {
			log.Fatalf("cannot run TCP server, %v", err)
		}
	}()
	log.Debugf("starting UDP Server at %s", *udpAddress)
	go func() {
		if err := server.ListenAndServeUDP(*udpAddress); err != nil {
			log.Fatalf("cannot run UDP server, %v", err)
		}
	}()
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	os.Exit(0)
}

func frontpage(w http.ResponseWriter, r *http.Request) {
	log.Debugln("frontpage")
	hn, err := os.Hostname()
	if err == nil {
		fmt.Fprintf(w, "hostname %s", hn)
	}
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		fmt.Fprintf(w, "outbound ip %v", localAddr.IP)
	}
	conn.Close() // nolint
	fmt.Fprintf(w, "real outbound ip %s", server.GetRealOutboundIP())
	fmt.Fprintf(w, "totaltcp %v", server.TotalTCP)
	fmt.Fprintf(w, "totaludp %v", server.TotalUDP)
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, intf := range interfaces {
			addrs, err := intf.Addrs()
			if err == nil {
				fmt.Fprintf(w, "%v %v ", intf, addrs)
			}
		}
	}
}
