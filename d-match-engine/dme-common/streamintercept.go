package dmecommon

import (
	"context"
	"reflect"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// A wrapper for the real grpc.ServerStream
type ServerStreamWrapper struct {
	inner grpc.ServerStream
	ctx   context.Context
}

func (s *ServerStreamWrapper) SetHeader(m metadata.MD) error {
	return s.inner.SetHeader(m)
}

func (s *ServerStreamWrapper) SendHeader(m metadata.MD) error {
	return s.inner.SendHeader(m)
}

func (s *ServerStreamWrapper) SetTrailer(m metadata.MD) {
	s.inner.SetTrailer(m)
}

func (s *ServerStreamWrapper) Context() context.Context {
	return s.ctx
}

func (s *ServerStreamWrapper) SendMsg(m interface{}) error {
	log.SpanLog(s.Context(), log.DebugLevelDmereq, "SendMsg auth stream message", "type", reflect.TypeOf(m).String())
	return s.inner.SendMsg(m)
}

func (s *ServerStreamWrapper) RecvMsg(m interface{}) error {
	log.SpanLog(s.Context(), log.DebugLevelDmereq, "RecvMsg auth stream message", "type", reflect.TypeOf(m).String())
	var cookie string

	err := s.inner.RecvMsg(m)
	ctx := s.Context()
	switch typ := m.(type) {
	case *dme.QosPositionRequest:
		cookie = typ.SessionCookie
		// Verify session cookie
		ckey, err := VerifyCookie(ctx, cookie)
		log.SpanLog(s.Context(), log.DebugLevelDmereq, "QosPosition VerifyCookie result", "ckey", ckey, "err", err)
		if err != nil {
			return grpc.Errorf(codes.Unauthenticated, err.Error())
		}
		s.ctx = NewCookieContext(ctx, ckey)
	case *dme.ClientEdgeEvent:
		// Verify session cookie
		cookie = typ.SessionCookie
		ckey, err := VerifyCookie(ctx, cookie)
		log.SpanLog(s.Context(), log.DebugLevelDmereq, "EdgeEvent VerifyCookie result", "ckey", ckey, "err", err)
		if err != nil {
			return grpc.Errorf(codes.Unauthenticated, err.Error())
		}
		ctx = NewCookieContext(ctx, ckey)
		// Verify EdgeEventsCookieKey
		eeCookie := typ.EdgeEventsCookie
		eekey, err := VerifyEdgeEventsCookie(ctx, eeCookie)
		log.SpanLog(ctx, log.DebugLevelDmereq, "EdgeEvent VerifyEdgeEventsCookie result", "key", eekey, "err", err)
		if err != nil {
			return grpc.Errorf(codes.Unauthenticated, err.Error())
		}
		ctx = NewEdgeEventsCookieContext(ctx, eekey)
		// Add session cookies to context
		s.ctx = ctx
		log.SpanLog(s.Context(), log.DebugLevelDmereq, "Auth intercept new context", "ctx", s.ctx)
	default:
		log.InfoLog("Unknown streaming operation, cannot verify cookie", "type", reflect.TypeOf(m).String())
		return grpc.Errorf(codes.Unauthenticated, err.Error())
	}

	return err
}

func GetStreamAuthInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Set up a child span for the stream interceptor
		span := log.StartSpan(log.DebugLevelDmereq, "stream auth interceptor")
		defer span.Finish()
		cctx := log.ContextWithSpan(ss.Context(), span)

		wrapper := &ServerStreamWrapper{inner: ss, ctx: cctx}
		return handler(srv, wrapper)
	}
}
