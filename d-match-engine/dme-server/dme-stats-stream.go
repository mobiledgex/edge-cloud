package main

import (
	"reflect"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	log "github.com/mobiledgex/edge-cloud/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
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
	start := time.Now()

	call := ApiStatCall{}

	err := w.inner.RecvMsg(m)
	w.ctx = context.WithValue(w.inner.Context(), dmecommon.StatKeyContextKey, &call.key)

	_, call.key.Method = cloudcommon.ParseGrpcMethod(w.info.FullMethod)

	switch typ := m.(type) {
	case *dme.ClientEdgeEvent:
		ckey, ok := dmecommon.CookieFromContext(w.ctx)
		if !ok {
			log.SpanLog(w.Context(), log.DebugLevelDmereq, "ClientEdgeEvent unable to get session cookie from context")
			return err
		}
		call.key.AppKey.Organization = ckey.OrgName
		call.key.AppKey.Name = ckey.AppName
		call.key.AppKey.Version = ckey.AppVers

		// Separate metric for Edge Events latency -> appinst-latency
		if typ.Event == dme.EventType_EVENT_LATENCY {
			eekey, ok := dmecommon.EdgeEventsCookieFromContext(w.ctx)
			if !ok {
				log.SpanLog(w.Context(), log.DebugLevelDmereq, "ClientEdgeEvent latency unable to get edge events cookie from context")
				return err
			}
			call.key.CloudletFound.Organization = eekey.CloudletOrg
			call.key.CloudletFound.Name = eekey.CloudletName
			call.key.ClusterKey.Name = eekey.ClusterName
			call.key.ClusterInstOrg = eekey.ClusterOrg

			// Create AppInstKey from SessionCookie and EdgeEventsCookie
			appInstKey := &edgeproto.AppInstKey{
				AppKey: edgeproto.AppKey{
					Organization: ckey.OrgName,
					Name:         ckey.AppName,
					Version:      ckey.AppVers,
				},
				ClusterInstKey: edgeproto.ClusterInstKey{
					ClusterKey: edgeproto.ClusterKey{
						Name: eekey.ClusterName,
					},
					CloudletKey: edgeproto.CloudletKey{
						Organization: eekey.CloudletOrg,
						Name:         eekey.CloudletName,
					},
					Organization: eekey.ClusterOrg,
				},
			}
			// Grab Peer info
			p, ok := peer.FromContext(w.Context())
			if !ok {
				log.SpanLog(w.Context(), log.DebugLevelDmereq, "ClientEdgeEvent latency unable to get peer info from context")
				return err
			}
			// Handle latency samples from client
			latency, ok := dmecommon.ProcessLatencySamples(w.Context(), appInstKey, p, typ.Samples)
			if !ok {
				log.SpanLog(w.Context(), log.DebugLevelDmereq, "ClientEdgeEvent latency unable to process latency samples")
				return err
			}
			call.latency = time.Duration(latency.Avg * float64(time.Millisecond))

			call.key.Method = EdgeEventLatencyMethod // override method name
			log.SpanLog(w.Context(), log.DebugLevelDmereq, "ClientEdgeEvent latency processing results", "latency", call.latency)
			w.stats.RecordApiStatCall(&call)
			return nil
		}
		// TODO: GPS UPDATE
	// Other stream apis provide same metric as Unary
	default:
		ckey, ok := dmecommon.CookieFromContext(w.ctx)
		if !ok {
			log.SpanLog(w.Context(), log.DebugLevelDmereq, "Default unable to get session cookie from context")
			return err
		}
		call.key.AppKey.Organization = ckey.OrgName
		call.key.AppKey.Name = ckey.AppName
		call.key.AppKey.Version = ckey.AppVers
	}
	call.latency = time.Since(start)

	w.stats.RecordApiStatCall(&call)
	return err
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
