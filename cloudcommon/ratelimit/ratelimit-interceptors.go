package ratelimit

import (
	"context"
	"errors"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// Grpc unary server interceptor that does rate limiting for DME
func GetDmeUnaryRateLimiterInterceptor(limiter Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// intialize span
		_, method := cloudcommon.ParseGrpcMethod(info.FullMethod)
		span := log.NewSpanFromGrpc(ctx, log.DebugLevelDmereq, method)
		defer span.Finish()
		ctx = log.ContextWithSpan(ctx, span)
		// Create ctx with rateLimitInfo
		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, errors.New("unable to get peer IP info")
		}
		rateLimitInfo := &LimiterInfo{
			Api: method,
			Ip:  p.Addr.String(),
		}
		ctx = NewLimiterInfoContext(ctx, rateLimitInfo)
		// Call dynamic value's Limit function
		limit, err := limiter.Limit(ctx)
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

// Grpc unary server interceptor that does rate limiting for Controller
func GetControllerUnaryRateLimiterInterceptor(limiter Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Create ctx with rateLimitInfo
		_, method := cloudcommon.ParseGrpcMethod(info.FullMethod)
		pr, ok := peer.FromContext(ctx)
		client := "unknown"
		if ok {
			client = pr.Addr.String()
		}
		rateLimitInfo := &LimiterInfo{
			Api: method,
			Ip:  client,
		}
		ctx = NewLimiterInfoContext(ctx, rateLimitInfo)
		// Call dynamic value's Limit function
		limit, err := limiter.Limit(ctx)
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
	ctx context.Context
}

func (s *LimiterStreamWrapper) Context() context.Context {
	return s.ctx
}

// Return a grpc stream server interceptor that does rate limiting for DME
func GetDmeStreamRateLimiterInterceptor(limiter Limiter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Set up a child span for the stream interceptor
		span := log.StartSpan(log.DebugLevelDmereq, "dme stream limiter interceptor")
		defer span.Finish()
		cctx := log.ContextWithSpan(ss.Context(), span)
		// Create ctx with rateLimitInfo
		_, method := cloudcommon.ParseGrpcMethod(info.FullMethod)
		p, ok := peer.FromContext(cctx)
		if !ok {
			return errors.New("unable to get peer IP info")
		}
		rateLimitInfo := &LimiterInfo{
			Api: method,
			Ip:  p.Addr.String(),
		}
		cctx = NewLimiterInfoContext(cctx, rateLimitInfo)
		// Call dynamic value's Limit function
		limit, err := limiter.Limit(cctx)
		if limit {
			errMsg := fmt.Sprintf("%s is rejected by grpc_ratelimit middleware, please retry later.", info.FullMethod)
			if err != nil {
				errMsg += fmt.Sprintf(" error is %s.", err.Error())
			}
			return status.Errorf(codes.ResourceExhausted, errMsg)
		}

		wrapper := &LimiterStreamWrapper{ServerStream: ss, ctx: cctx}
		return handler(srv, wrapper)
	}
}

// Return a grpc stream server interceptor that does rate limiting for Controller
func GetControllerStreamRateLimiterInterceptor(limiter Limiter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Set up a child span for the stream interceptor
		span := log.StartSpan(log.DebugLevelApi, "controller stream limiter interceptor")
		defer span.Finish()
		cctx := log.ContextWithSpan(ss.Context(), span)
		// Create ctx with rateLimitInfo
		_, method := cloudcommon.ParseGrpcMethod(info.FullMethod)
		p, ok := peer.FromContext(cctx)
		if !ok {
			return errors.New("unable to get peer IP info")
		}
		rateLimitInfo := &LimiterInfo{
			Api: method,
			Ip:  p.Addr.String(),
		}
		cctx = NewLimiterInfoContext(cctx, rateLimitInfo)
		// Call dynamic value's Limit function
		limit, err := limiter.Limit(cctx)
		if limit {
			errMsg := fmt.Sprintf("%s is rejected by grpc_ratelimit middleware, please retry later.", info.FullMethod)
			if err != nil {
				errMsg += fmt.Sprintf(" error is %s.", err.Error())
			}
			return status.Errorf(codes.ResourceExhausted, errMsg)
		}

		wrapper := &LimiterStreamWrapper{ServerStream: ss, ctx: cctx}
		return handler(srv, wrapper)
	}
}

// TODO: Add grpc metatdata with rate limit information (limit, reqs remaining, limit restart) when MaxReqs is used
