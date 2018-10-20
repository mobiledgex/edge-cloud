package cloudcommon

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogo/gateway"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
)

func ParseGrpcMethod(method string) (path string, cmd string) {
	if i := strings.LastIndexByte(method, '/'); i > 0 {
		path = method[:i]
		cmd = method[i+1:]
	} else {
		path = ""
		cmd = method
	}
	return path, cmd
}

type GrpcGWConfig struct {
	ApiAddr     string
	TlsCertFile string
	ApiHandles  []func(context.Context, *gwruntime.ServeMux, *grpc.ClientConn) error
}

func GrpcGateway(cfg *GrpcGWConfig) (http.Handler, error) {
	ctx := context.Background()
	dialOption, err := tls.GetTLSClientDialOption(cfg.ApiAddr, cfg.TlsCertFile)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.DialContext(ctx, cfg.ApiAddr, dialOption)
	if err != nil {
		log.FatalLog("Failed to start REST gateway", "error", err)
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
	for _, f := range cfg.ApiHandles {
		if err := f(ctx, mux, conn); err != nil {
			return nil, err
		}
	}
	return mux, nil
}

func GrpcGatewayServe(cfg *GrpcGWConfig, server *http.Server) {
	// Serve REST gateway
	if cfg.TlsCertFile != "" {
		tlsKeyFile := strings.Replace(cfg.TlsCertFile, ".crt", ".key", -1)
		if err := server.ListenAndServeTLS(cfg.TlsCertFile, tlsKeyFile); err != http.ErrServerClosed {
			log.FatalLog("Failed to serve HTTP TLS", "error", err)
		}
	} else {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.FatalLog("Failed to serve HTTP", "error", err)
		}
	}
}
