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
