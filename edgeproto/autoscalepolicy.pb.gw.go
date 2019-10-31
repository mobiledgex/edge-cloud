// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: autoscalepolicy.proto

/*
Package edgeproto is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package edgeproto

import (
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray

func request_AutoScalePolicyApi_CreateAutoScalePolicy_0(ctx context.Context, marshaler runtime.Marshaler, client AutoScalePolicyApiClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq AutoScalePolicy
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.CreateAutoScalePolicy(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func request_AutoScalePolicyApi_DeleteAutoScalePolicy_0(ctx context.Context, marshaler runtime.Marshaler, client AutoScalePolicyApiClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq AutoScalePolicy
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.DeleteAutoScalePolicy(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func request_AutoScalePolicyApi_UpdateAutoScalePolicy_0(ctx context.Context, marshaler runtime.Marshaler, client AutoScalePolicyApiClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq AutoScalePolicy
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.UpdateAutoScalePolicy(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func request_AutoScalePolicyApi_ShowAutoScalePolicy_0(ctx context.Context, marshaler runtime.Marshaler, client AutoScalePolicyApiClient, req *http.Request, pathParams map[string]string) (AutoScalePolicyApi_ShowAutoScalePolicyClient, runtime.ServerMetadata, error) {
	var protoReq AutoScalePolicy
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	stream, err := client.ShowAutoScalePolicy(ctx, &protoReq)
	if err != nil {
		return nil, metadata, err
	}
	header, err := stream.Header()
	if err != nil {
		return nil, metadata, err
	}
	metadata.HeaderMD = header
	return stream, metadata, nil

}

// RegisterAutoScalePolicyApiHandlerFromEndpoint is same as RegisterAutoScalePolicyApiHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterAutoScalePolicyApiHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterAutoScalePolicyApiHandler(ctx, mux, conn)
}

// RegisterAutoScalePolicyApiHandler registers the http handlers for service AutoScalePolicyApi to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterAutoScalePolicyApiHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterAutoScalePolicyApiHandlerClient(ctx, mux, NewAutoScalePolicyApiClient(conn))
}

// RegisterAutoScalePolicyApiHandlerClient registers the http handlers for service AutoScalePolicyApi
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "AutoScalePolicyApiClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "AutoScalePolicyApiClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "AutoScalePolicyApiClient" to call the correct interceptors.
func RegisterAutoScalePolicyApiHandlerClient(ctx context.Context, mux *runtime.ServeMux, client AutoScalePolicyApiClient) error {

	mux.Handle("POST", pattern_AutoScalePolicyApi_CreateAutoScalePolicy_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_AutoScalePolicyApi_CreateAutoScalePolicy_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_AutoScalePolicyApi_CreateAutoScalePolicy_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_AutoScalePolicyApi_DeleteAutoScalePolicy_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_AutoScalePolicyApi_DeleteAutoScalePolicy_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_AutoScalePolicyApi_DeleteAutoScalePolicy_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_AutoScalePolicyApi_UpdateAutoScalePolicy_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_AutoScalePolicyApi_UpdateAutoScalePolicy_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_AutoScalePolicyApi_UpdateAutoScalePolicy_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_AutoScalePolicyApi_ShowAutoScalePolicy_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_AutoScalePolicyApi_ShowAutoScalePolicy_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_AutoScalePolicyApi_ShowAutoScalePolicy_0(ctx, mux, outboundMarshaler, w, req, func() (proto.Message, error) { return resp.Recv() }, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_AutoScalePolicyApi_CreateAutoScalePolicy_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"create", "autoscalepolicy"}, ""))

	pattern_AutoScalePolicyApi_DeleteAutoScalePolicy_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"delete", "autoscalepolicy"}, ""))

	pattern_AutoScalePolicyApi_UpdateAutoScalePolicy_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"update", "autoscalepolicy"}, ""))

	pattern_AutoScalePolicyApi_ShowAutoScalePolicy_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1}, []string{"show", "autoscalepolicy"}, ""))
)

var (
	forward_AutoScalePolicyApi_CreateAutoScalePolicy_0 = runtime.ForwardResponseMessage

	forward_AutoScalePolicyApi_DeleteAutoScalePolicy_0 = runtime.ForwardResponseMessage

	forward_AutoScalePolicyApi_UpdateAutoScalePolicy_0 = runtime.ForwardResponseMessage

	forward_AutoScalePolicyApi_ShowAutoScalePolicy_0 = runtime.ForwardResponseStream
)
