package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/example/mexexample/api"
)

func ListenAndServeGRPC(address string) error {
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	s := Server{}
	api.RegisterMexExampleServer(grpcServer, &s)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s : %v", address, err)
	}

	return grpcServer.Serve(lis)
}

func headerMatcher(headerName string) (string, bool) {
	return strings.ToLower(headerName), true
}

func ListenAndServeREST(restAddress, grpcAddress string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(headerMatcher))

	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := api.RegisterMexExampleHandlerFromEndpoint(
		ctx,
		mux,
		grpcAddress,
		opts,
	)
	if err != nil {
		return fmt.Errorf("could not register REST service: %s", err)
	}

	return http.ListenAndServe(restAddress, mux)
}

func ListenAndServeTCP(tcpAddress string) error {
	la, err := net.Listen("tcp", tcpAddress)
	if err != nil {
		return err
	}
	defer la.Close()
	for {
		cl, err := la.Accept()
		if err != nil {
			return err
		}
		go handleTCP(cl)
	}
}

var TotalTCP = uint64(0)

func handleTCP(cl net.Conn) {
	defer cl.Close()
	myip := GetRealOutboundIP()
	buffer := make([]byte, 1024)
	for {
		nr, err := cl.Read(buffer)
		if err != nil {
			fmt.Println("error reading from tcp socket", err)
			return
		}
		TotalTCP += uint64(nr)
		dat := fmt.Sprintf("%s:%v:", myip, TotalTCP)
		_, err = cl.Write([]byte(dat))
		if err != nil {
			fmt.Println("error writing to tcp socket", err)
			return
		}
		_, err = cl.Write(buffer)
		if err != nil {
			fmt.Println("error writing to tcp socket", err)
			return
		}
	}
}

var TotalUDP = uint64(0)

func ListenAndServeUDP(udpAddress string) error {
	uc, err := net.ListenPacket("udp", udpAddress)
	if err != nil {
		return err
	}
	defer uc.Close()
	fmt.Println("reading udp", udpAddress)
	myip := GetRealOutboundIP()
	buffer := make([]byte, 1024)
	for {
		nr, addr, err := uc.ReadFrom(buffer)
		if err != nil {
			fmt.Println("udp read error", err)
			return err
		}
		TotalUDP += uint64(nr)
		dat := fmt.Sprintf("%s:%v:", myip, TotalUDP)
		_, err = uc.WriteTo([]byte(dat), addr)
		if err != nil {
			fmt.Println("udp write error", err)
			return err
		}
		_, err = uc.WriteTo(buffer, addr)
		if err != nil {
			fmt.Println("udp write error", err)
			return err
		}
	}
}
