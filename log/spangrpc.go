package log

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const spanKey = "mobiledgex-spankey"

func UnaryClientTraceGrpc(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	val := SpanToString(ctx)
	if val != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, spanKey, val)
	}
	return invoker(ctx, method, req, resp, cc, opts...)
}

func StreamClientTraceGrpc(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	val := SpanToString(ctx)
	if val != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, spanKey, val)
	}
	return streamer(ctx, desc, cc, method, opts...)
}

// NewSpanFromGrpc is used on server-side in controller/audit.go to extract span
func NewSpanFromGrpc(ctx context.Context, lvl uint64, spanName string) opentracing.Span {
	val := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals, ok := md[spanKey]; ok {
			val = vals[0]
		}
	}
	return NewSpanFromString(lvl, val, spanName)
}
