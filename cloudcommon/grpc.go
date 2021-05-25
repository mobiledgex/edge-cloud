package cloudcommon

import (
	"context"
	ctls "crypto/tls"
	"net/http"
	"strings"
	"time"

	"github.com/gogo/gateway"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	ApiAddr        string
	GetCertificate func(*ctls.CertificateRequestInfo) (*ctls.Certificate, error)
	TlsCertFile    string
	ApiHandles     []func(context.Context, *gwruntime.ServeMux, *grpc.ClientConn) error
}

func GrpcGateway(cfg *GrpcGWConfig) (http.Handler, error) {
	ctx := context.Background()
	// GRPC GW does not validate the GRPC server cert because it may be public signed and therefore
	// may not work with internal addressing
	skipVerify := true
	tlsMode := tls.MutualAuthTLS
	if cfg.GetCertificate == nil && cfg.TlsCertFile == "" {
		tlsMode = tls.NoTLS
	}
	dialOption, err := tls.GetTLSClientDialOption(tlsMode, cfg.ApiAddr, cfg.GetCertificate, cfg.TlsCertFile, skipVerify)
	if err != nil {
		log.FatalLog("Unable to get TLSClient Dial Option", "error", err)
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

func GrpcGatewayServe(server *http.Server, tlsCertFile string) {
	// Serve REST gateway
	cfg := server.TLSConfig
	if cfg == nil || (cfg.GetCertificate == nil && tlsCertFile == "") {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.FatalLog("Failed to serve HTTP", "error", err)
		}
	} else if cfg.GetCertificate != nil {
		if err := server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			log.FatalLog("Failed to serve HTTP TLS", "error", err)
		}
	} else {
		tlsKeyFile := strings.Replace(tlsCertFile, ".crt", ".key", -1)
		if err := server.ListenAndServeTLS(tlsCertFile, tlsKeyFile); err != http.ErrServerClosed {
			log.FatalLog("Failed to serve HTTP TLS", "error", err)
		}
	}
}

func TimeToTimestamp(t time.Time) dme.Timestamp {
	ts := dme.Timestamp{}
	ts.Seconds = t.Unix()
	ts.Nanos = int32(t.Nanosecond())
	return ts
}

func TimestampToTime(ts dme.Timestamp) time.Time {
	return time.Unix(ts.Seconds, int64(ts.Nanos))
}

func GrpcCreds(cfg *ctls.Config) grpc.ServerOption {
	if cfg == nil {
		return grpc.Creds(nil)
	} else {
		return grpc.Creds(credentials.NewTLS(cfg))
	}
}
