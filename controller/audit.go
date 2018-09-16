package main

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
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

	log.AuditLogStart(id, cmd, client, "notyet", "req", req)
	resp, err := handler(ctx, req)
	log.AuditLogEnd(id, err)

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

	if info.IsClientStream {
		log.AuditLogStart(id, cmd, client, "notyet")
	} else {
		// client API is not actually declared as streaming, but client
		// API argument is passed in via the stream.
		// Defer audit log until we can capture the client's argument.
		stream = NewAuditRecvOne(stream, id, cmd, client, "notyet")
	}
	err := handler(srv, stream)
	log.AuditLogEnd(id, err)

	return err
}

type AuditRecvOne struct {
	grpc.ServerStream
	id     uint64
	client string
	cmd    string
	user   string
}

func NewAuditRecvOne(stream grpc.ServerStream, id uint64, cmd, client, user string) *AuditRecvOne {
	s := AuditRecvOne{
		ServerStream: stream,
		id:           id,
		cmd:          cmd,
		client:       client,
		user:         user,
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
	log.AuditLogStart(s.id, s.cmd, s.client, s.user, "req", m)
	return err
}
