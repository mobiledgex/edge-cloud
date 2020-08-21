package dmecommon

import (
	"reflect"

	log "github.com/mobiledgex/edge-cloud/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const EdgeEventLatencyMethod = "appinst-latency"

// A wrapper for grpc.ServerStream to handle stats
type StatsStreamWrapper struct {
	stats *DmeStats
	inner grpc.ServerStream
	info  *grpc.StreamServerInfo
	ctx   context.Context
}

func (w *StatsStreamWrapper) SetHeader(m metadata.MD) error {
	return w.inner.SetHeader(m)
}

func (w *StatsStreamWrapper) SendHeader(m metadata.MD) error {
	return w.inner.SendHeader(m)
}

func (w *StatsStreamWrapper) SetTrailer(m metadata.MD) {
	w.inner.SetTrailer(m)
}

func (w *StatsStreamWrapper) Context() context.Context {
	return w.ctx
}

func (w *StatsStreamWrapper) SendMsg(m interface{}) error {
	log.SpanLog(w.Context(), log.DebugLevelDmereq, "SendMsg stats stream interceptor", "type", reflect.TypeOf(m).String())
	return w.inner.SendMsg(m)
}

func (w *StatsStreamWrapper) RecvMsg(m interface{}) error {
	log.SpanLog(w.Context(), log.DebugLevelDmereq, "RecvMsg stats stream interceptor", "type", reflect.TypeOf(m).String())
	return w.inner.RecvMsg(m)
}

func (s *DmeStats) GetStreamStatsInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Set up a child span for the stream interceptor
		span := log.StartSpan(log.DebugLevelDmereq, "stream stats interceptor")
		defer span.Finish()
		cctx := log.ContextWithSpan(ss.Context(), span)

		wrapper := &StatsStreamWrapper{stats: s, inner: ss, info: info, ctx: cctx}
		return handler(srv, wrapper)
	}
}
