package testutil

import (
	"context"
	fmt "fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc"
)

func (s *DummyServer) AddDummyObjs(ctx context.Context, num int) {
	for ii := 0; ii < num; ii++ {
		name := fmt.Sprintf("%d", ii)

		flavor := edgeproto.Flavor{}
		flavor.Key.Name = name
		s.FlavorCache.Update(ctx, &flavor, int64(ii))
	}
}

func (s *DummyServer) AddDummyOrgObjs(ctx context.Context, org string, num int) {
	for ii := 0; ii < num; ii++ {
		name := fmt.Sprintf("%d", ii)

		app := edgeproto.App{}
		app.Key.Organization = org
		app.Key.Name = name
		s.AppCache.Update(ctx, &app, int64(ii))

		appinst := edgeproto.AppInst{}
		appinst.Key.AppKey.Organization = org
		appinst.Key.AppKey.Name = name
		s.AppInstCache.Update(ctx, &appinst, int64(ii))

		cinst := edgeproto.ClusterInst{}
		cinst.Key.Organization = org
		cinst.Key.ClusterKey.Name = name
		s.ClusterInstCache.Update(ctx, &cinst, int64(ii))

		cloudlet := edgeproto.Cloudlet{}
		cloudlet.Key.Organization = org
		cloudlet.Key.Name = name
		s.CloudletCache.Update(ctx, &cloudlet, int64(ii))

		cloudletInfo := edgeproto.CloudletInfo{}
		cloudletInfo.Key.Organization = org
		cloudletInfo.Key.Name = name
		s.CloudletInfoCache.Update(ctx, &cloudletInfo, int64(ii))

		pool := edgeproto.CloudletPool{}
		pool.Key.Name = name
		pool.Key.Organization = org
		pool.Cloudlets = []string{"cloudlet1", "cloudlet2", "cloudlet3"}
		s.CloudletPoolCache.Update(ctx, &pool, int64(ii))
	}
}

func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	span := log.StartSpan(log.DebugLevelApi, info.FullMethod)
	defer span.Finish()
	ctx = log.ContextWithSpan(ctx, span)
	return handler(ctx, req)
}

func StreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	span := log.StartSpan(log.DebugLevelApi, info.FullMethod)
	defer span.Finish()
	ctx := log.ContextWithSpan(stream.Context(), span)
	ss := ServerStream{
		ServerStream: stream,
		ctx:          ctx,
	}
	return handler(srv, &ss)
}

type ServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *ServerStream) Context() context.Context {
	return s.ctx
}
