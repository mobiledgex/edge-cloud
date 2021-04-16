package cloudcommon

import (
	"context"
	ctls "crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gogo/gateway"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mobiledgex/edge-cloud/cloudcommon/ratelimit"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
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

// TODO: Move these interceptors into ratelimit folder
// TODO: Add grpc metatdata with rate limit information (limit, reqs remaining, limit restart)

// TODO: Implement one of these in DME, Controller, and MC depending on API (or just pass in a RateLimitManager from those??)
// Grpc unary server interceptor that does rate limiting per user, org, and/or ip
/*func GetUnaryApiLimiterInterceptor(apiRateLimitMgr *ratelimit.ApiRateLimitManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		api := info.FullMethod
		usr := ""
		if apiRateLimitMgr.LimitByUser(api) {
			// get usr
			if u, ok := ctx.Value("username").(string); ok {
				usr = u
			}
		}
		org := ""
		if apiRateLimitMgr.LimitByOrg(api) {
			// get org
			org = edgeproto.GetOrg(req)
		}
		ip := ""
		if apiRateLimitMgr.LimitByIp(api) {
			// get ip
			pr, ok := peer.FromContext(ctx)
			ip = "unknown"
			if ok {
				ip = pr.Addr.String()
			}
		}
		limit, err := apiRateLimitMgr.Limit(ctx, api, usr, org, ip)
		if limit {
			errMsg := fmt.Sprintf("%s is rejected by grpc_ratelimit middleware, please retry later.", api)
			if err != nil {
				errMsg += fmt.Sprintf(" error is %s.", err.Error())
			}
			return nil, status.Errorf(codes.ResourceExhausted, errMsg)

		}
		return handler(ctx, req)
	}
}*/

// Grpc unary server interceptor that does rate limit throttling
func GetUnaryRateLimiterInterceptor(limiter ratelimit.Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		rateLimitCtx := ratelimit.Context{Context: ctx}
		_, method := ParseGrpcMethod(info.FullMethod)
		rateLimitCtx.Api = method
		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, errors.New("unable to get peer IP info")
		}
		rateLimitCtx.Ip = p.Addr.String()
		limit, err := limiter.Limit(rateLimitCtx)
		if limit {
			errMsg := fmt.Sprintf("%s is rejected by grpc ratelimit middleware, please retry later.", info.FullMethod)
			if err != nil {
				errMsg += fmt.Sprintf(" Error is: %s.", err.Error())
			}
			return nil, status.Errorf(codes.ResourceExhausted, errMsg)

		}
		return handler(ctx, req)
	}
}

type LimiterStreamWrapper struct {
	grpc.ServerStream
	limiter ratelimit.Limiter
	ctx     ratelimit.Context
}

func (s *LimiterStreamWrapper) Context() context.Context {
	return s.ctx.Context
}

func (s *LimiterStreamWrapper) Limit() (bool, error) {
	return s.limiter.Limit(s.ctx)
}

// Return a grpc stream server interceptor that does rate limit throttling
func GetStreamRateLimiterInterceptor(limiter ratelimit.Limiter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Set up a child span for the stream interceptor
		span := log.StartSpan(log.DebugLevelDmereq, "stream limiter interceptor")
		defer span.Finish()
		cctx := log.ContextWithSpan(ss.Context(), span)

		rateLimitCtx, ok := cctx.(ratelimit.Context)
		if !ok {
			// return error
		}
		wrapper := &LimiterStreamWrapper{ServerStream: ss, limiter: limiter, ctx: rateLimitCtx}
		limit, err := wrapper.Limit()
		if limit {
			errMsg := fmt.Sprintf("%s is rejected by grpc_ratelimit middleware, please retry later.", info.FullMethod)
			if err != nil {
				errMsg += fmt.Sprintf(" error is %s.", err.Error())
			}
			return status.Errorf(codes.ResourceExhausted, errMsg)

		}
		return handler(srv, wrapper)
	}
}
