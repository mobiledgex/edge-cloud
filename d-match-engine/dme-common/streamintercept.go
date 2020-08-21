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

// TODO: This file handles authentication only.  We need to handle stats also.

// A wrapper for the real grpc.ServerStream
type ServerStreamWrapper struct {
	inner grpc.ServerStream
	ctx   context.Context
}

func (s ServerStreamWrapper) SetHeader(m metadata.MD) error {
	return s.inner.SetHeader(m)
}

func (s ServerStreamWrapper) SendHeader(m metadata.MD) error {
	return s.inner.SendHeader(m)
}

func (s ServerStreamWrapper) SetTrailer(m metadata.MD) {
	s.inner.SetTrailer(m)
}

func (s ServerStreamWrapper) Context() context.Context {
	return s.ctx
}

func (s ServerStreamWrapper) SendMsg(m interface{}) error {
	log.SpanLog(s.Context(), log.DebugLevelDmereq, "SendMsg Streamed message", "type", reflect.TypeOf(m).String())
	return s.inner.SendMsg(m)
}

func (a ServerStreamWrapper) RecvMsg(m interface{}) error {
	log.SpanLog(a.Context(), log.DebugLevelDmereq, "RecvMsg Streamed message", "type", reflect.TypeOf(m).String())
	var cookie string

	err := a.inner.RecvMsg(m)
	switch typ := m.(type) {
	case *dme.QosPositionRequest:
		cookie = typ.SessionCookie
		// Verify session cookie
		ckey, err := VerifyCookie(a.Context(), cookie)
		log.SpanLog(a.Context(), log.DebugLevelDmereq, "QosPosition VerifyCookie result", "ckey", ckey, "err", err)
		if err != nil {
			return grpc.Errorf(codes.Unauthenticated, err.Error())
		}
	case *dme.ClientEdgeEvent:
		// CONSOLIDATE CODE
		cookie = typ.SessionCookie
		// Verify session cookie
		ckey, err := VerifyCookie(a.Context(), cookie)
		log.SpanLog(a.Context(), log.DebugLevelDmereq, "EdgeEvent VerifyCookie result", "ckey", ckey, "err", err)
		if err != nil {
			return grpc.Errorf(codes.Unauthenticated, err.Error())
		}
		break
	default:
		log.InfoLog("Unknown streaming operation, cannot verify cookie", "type", reflect.TypeOf(m).String())
		return grpc.Errorf(codes.Unauthenticated, err.Error())
	}
	return err
}

func GetStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Set up a child span for the stream interceptor
		span := log.StartSpan(log.DebugLevelDmereq, "stream interceptor")
		defer span.Finish()
		cctx := log.ContextWithSpan(ss.Context(), span)

		wrapper := ServerStreamWrapper{inner: ss, ctx: cctx}
		return handler(srv, wrapper)
	}
}
