package main

import (
	"context"
	"net/http"

	"github.com/gogo/gateway"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mobiledgex/edge-cloud/proto"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
)

func grpcGateway(addr string) (http.Handler, error) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		util.FatalLog("Failed to start REST gateway", "error", err)
	}

	jsonpb := &gateway.JSONPb{
		EmitDefaults: true,
		Indent:       " ",
		OrigName:     true,
	}
	mux := gwruntime.NewServeMux(
		// this avoids a marshaling issue with grpc-gateway and
		// gogo protobuf non-nullable embedded structs
		gwruntime.WithMarshalerOption(gwruntime.MIMEWildcard, jsonpb),
		// this is necessary to get error details properly
		// marshalled in unary requests
		gwruntime.WithProtoErrorHandler(gwruntime.DefaultHTTPProtoErrorHandler),
	)
	for _, f := range []func(context.Context, *gwruntime.ServeMux, *grpc.ClientConn) error{
		proto.RegisterDeveloperApiHandler,
		proto.RegisterAppApiHandler,
		proto.RegisterOperatorApiHandler,
		proto.RegisterCloudletApiHandler,
	} {
		if err := f(ctx, mux, conn); err != nil {
			return nil, err
		}
	}
	return mux, nil
}
