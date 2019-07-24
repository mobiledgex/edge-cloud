package main

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var AuditId uint64

func AuditUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	_, cmd := cloudcommon.ParseGrpcMethod(info.FullMethod)
	if strings.HasPrefix(cmd, "Show") {
		return handler(ctx, req)
	}

	id := atomic.AddUint64(&AuditId, 1)
	pr, ok := peer.FromContext(ctx)
	client := "unknown"
	if ok {
		client = pr.Addr.String()
	}

	span := log.StartSpan(log.DebugLevelApi, cmd)
	defer span.Finish()
	ctx = log.ContextWithSpan(ctx, span)
	span.SetTag("organization", edgeproto.GetOrg(req))
	span.SetTag("level", "audit")
	span.SetTag("client", client)
	span.SetTag("request", req)

	log.AuditLogStart(id, cmd, client, "notyet", "req", req)
	resp, err := handler(ctx, req)
	log.AuditLogEnd(id, err)

	log.SpanLog(ctx, log.DebugLevelApi, "finished", "err", err)

	return resp, err
}

func AuditStreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	_, cmd := cloudcommon.ParseGrpcMethod(info.FullMethod)
	if strings.HasPrefix(cmd, "Show") {
		return handler(srv, stream)
	}

	id := atomic.AddUint64(&AuditId, 1)
	pr, ok := peer.FromContext(stream.Context())
	client := "unknown"
	if ok {
		client = pr.Addr.String()
	}

	span := log.StartSpan(log.DebugLevelApi, cmd)
	defer span.Finish()
	ctx := log.ContextWithSpan(stream.Context(), span)
	span.SetTag("level", "audit")
	span.SetTag("client", client)

	if info.IsClientStream {
		log.AuditLogStart(id, cmd, client, "notyet")
	} else {
		// client API is not actually declared as streaming, but client
		// API argument is passed in via the stream.
		// Defer audit log until we can capture the client's argument.
		stream = NewAuditRecvOne(stream, ctx, id, cmd, client, "notyet")
	}
	err := handler(srv, stream)
	log.AuditLogEnd(id, err)
	log.SpanLog(ctx, log.DebugLevelApi, "finished", "err", err)

	return err
}

type AuditRecvOne struct {
	grpc.ServerStream
	id     uint64
	client string
	cmd    string
	user   string
	ctx    context.Context
}

func NewAuditRecvOne(stream grpc.ServerStream, ctx context.Context, id uint64, cmd, client, user string) *AuditRecvOne {
	s := AuditRecvOne{
		ServerStream: stream,
		id:           id,
		cmd:          cmd,
		client:       client,
		user:         user,
		ctx:          ctx,
	}
	return &s
}

func (s *AuditRecvOne) RecvMsg(m interface{}) error {
	// grpc handler will call RecvMsg once and only once,
	// to get the object to pass as the argument to our
	// implemented function. We log once the object has been
	// read from the stream. This way the audit log can capture
	// the API argument sent by the client.
	err := s.ServerStream.RecvMsg(m)
	span := opentracing.SpanFromContext(s.ctx)
	if span != nil {
		span.SetTag("organization", edgeproto.GetOrg(m))
		span.SetTag("request", m)
	}
	log.AuditLogStart(s.id, s.cmd, s.client, s.user, "req", m)
	return err
}

func (s *AuditRecvOne) Context() context.Context {
	return s.ctx
}
