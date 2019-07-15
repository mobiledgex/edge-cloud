package dmecommon

import (
	"context"
	"reflect"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc"

	"google.golang.org/grpc/metadata"
)

// A wrapper for the real grpc.ServerStream
type ServerStreamWrapper struct {
	inner grpc.ServerStream
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
	return s.inner.Context()
}

func (s ServerStreamWrapper) SendMsg(m interface{}) error {
	log.DebugLog(log.DebugLevelDmereq, "SendMsg Streamed message", "type", reflect.TypeOf(m).String())
	return s.inner.SendMsg(m)
}

func (a ServerStreamWrapper) RecvMsg(m interface{}) error {
	log.DebugLog(log.DebugLevelDmereq, "RecvMsg Streamed message", "type", reflect.TypeOf(m).String())
	var cookie string

	err := a.inner.RecvMsg(m)
	switch typ := m.(type) {
	case *dme.QosPositionKpiRequest:
		cookie = typ.SessionCookie
		// Verify session cookie
		ckey, err := VerifyCookie(cookie)
		log.DebugLog(log.DebugLevelDmereq, "VerifyCookie result", "ckey", ckey, "err", err)
		//	if err != nil {
		//		return grpc.Errorf(codes.Unauthenticated, err.Error())
		//	}
	}
	return err
}

func GetStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := ServerStreamWrapper{inner: ss}
		return handler(srv, wrapper)
	}
}
