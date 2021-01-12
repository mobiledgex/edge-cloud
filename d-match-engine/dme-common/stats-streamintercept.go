package dmecommon

import (
	"reflect"

	log "github.com/mobiledgex/edge-cloud/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// A wrapper for grpc.ServerStream to handle stats
type StatsStreamWrapper struct {
	grpc.ServerStream
	stats *DmeStats
	info  *grpc.StreamServerInfo
	ctx   context.Context
}

func (w *StatsStreamWrapper) Context() context.Context {
	return w.ctx
}

func (w *StatsStreamWrapper) SendMsg(m interface{}) error {
	log.SpanLog(w.Context(), log.DebugLevelDmereq, "SendMsg stats stream interceptor", "type", reflect.TypeOf(m).String())
	return w.ServerStream.SendMsg(m)
}

func (w *StatsStreamWrapper) RecvMsg(m interface{}) error {
	log.SpanLog(w.Context(), log.DebugLevelDmereq, "RecvMsg stats stream interceptor", "type", reflect.TypeOf(m).String())
	return w.ServerStream.RecvMsg(m)
}

func (s *DmeStats) GetStreamStatsInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Set up a child span for the stream interceptor
		span := log.StartSpan(log.DebugLevelDmereq, "stream stats interceptor")
		defer span.Finish()
		cctx := log.ContextWithSpan(ss.Context(), span)

		wrapper := &StatsStreamWrapper{ServerStream: ss, stats: s, info: info, ctx: cctx}
		return handler(srv, wrapper)
	}
}
