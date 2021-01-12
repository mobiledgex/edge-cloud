package dmecommon

import (
	"context"
	"reflect"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
)

// A wrapper for the real grpc.ServerStream
type ServerStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *ServerStreamWrapper) Context() context.Context {
	return s.ctx
}

func (s *ServerStreamWrapper) SendMsg(m interface{}) error {
	log.SpanLog(s.Context(), log.DebugLevelDmereq, "SendMsg auth stream message", "type", reflect.TypeOf(m).String())
	return s.ServerStream.SendMsg(m)
}

func (s *ServerStreamWrapper) RecvMsg(m interface{}) error {
	log.SpanLog(s.Context(), log.DebugLevelDmereq, "RecvMsg auth stream message", "type", reflect.TypeOf(m).String())
	var cookie string

	err := s.ServerStream.RecvMsg(m)
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
		// Authentication of Session Cookie and EdgeEvents Session Cookie occurs in match-engine.go:StreamEdgeEvent
		break
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

		wrapper := &ServerStreamWrapper{ServerStream: ss, ctx: cctx}
		return handler(srv, wrapper)
	}
}
