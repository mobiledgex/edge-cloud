package cloudcommon

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func AuditUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	_, cmd := ParseGrpcMethod(info.FullMethod)
	pr, ok := peer.FromContext(ctx)
	client := "unknown"
	if ok {
		client = pr.Addr.String()
	}

	span := log.NewSpanFromGrpc(ctx, log.DebugLevelApi, cmd)
	defer span.Finish()
	ctx = log.ContextWithSpan(ctx, span)
	span.SetTag("organization", edgeproto.GetOrg(req))
	span.SetTag("client", client)
	span.SetTag("request", req)

	resp, err := handler(ctx, req)
	// Make sure first letter is capitalized in error message
	if err != nil {
		modmsg := util.CapitalizeMessage(err.Error())
		err = errors.New(modmsg)
	}
	log.SpanLog(ctx, log.DebugLevelApi, "finished", "err", err)

	return resp, err
}

func AuditStreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	_, cmd := ParseGrpcMethod(info.FullMethod)
	pr, ok := peer.FromContext(stream.Context())
	client := "unknown"
	if ok {
		client = pr.Addr.String()
	}

	ctx := stream.Context()
	span := log.NewSpanFromGrpc(ctx, log.DebugLevelApi, cmd)
	defer span.Finish()
	ctx = log.ContextWithSpan(ctx, span)
	span.SetTag("client", client)

	if !info.IsClientStream {
		// client API is not actually declared as streaming, but client
		// API argument is passed in via the stream.
		// Defer audit log until we can capture the client's argument.
		stream = NewAuditRecvOne(stream, ctx)
	}
	err := handler(srv, stream)
	// Make sure first letter is capitalized in error message
	if err != nil {
		modmsg := util.CapitalizeMessage(err.Error())
		err = errors.New(modmsg)
	}
	log.SpanLog(ctx, log.DebugLevelApi, "finished", "err", err)

	return err
}

type AuditRecvOne struct {
	grpc.ServerStream
	ctx context.Context
}

func NewAuditRecvOne(stream grpc.ServerStream, ctx context.Context) *AuditRecvOne {
	s := AuditRecvOne{
		ServerStream: stream,
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
	return err
}

func (s *AuditRecvOne) Context() context.Context {
	return s.ctx
}
