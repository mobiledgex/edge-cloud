package testutil

import (
	"context"
	fmt "fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc"
)

type Action int

const (
	Create Action = iota
	Delete
)

type CustomData struct {
	OrgsOnCloudlet map[edgeproto.CloudletKey][]string
}

func (s *CustomData) Init() {
	s.OrgsOnCloudlet = make(map[edgeproto.CloudletKey][]string)
}

func (s *DummyServer) SetDummyObjs(ctx context.Context, a Action, tag string, num int) {
	for ii := 0; ii < num; ii++ {
		name := fmt.Sprintf("%s%d", tag, ii)

		flavor := edgeproto.Flavor{}
		flavor.Key.Name = name
		if a == Create {
			s.FlavorCache.Update(ctx, &flavor, int64(ii))
		} else if a == Delete {
			s.FlavorCache.Delete(ctx, &flavor, int64(ii))
		}
	}
}

func (s *DummyServer) SetDummyOrgObjs(ctx context.Context, a Action, org string, num int) {
	for ii := 0; ii < num; ii++ {
		name := fmt.Sprintf("%d", ii)

		app := edgeproto.App{}
		app.Key.Organization = org
		app.Key.Name = name
		if a == Create {
			s.AppCache.Update(ctx, &app, int64(ii))
		} else if a == Delete {
			s.AppCache.Delete(ctx, &app, int64(ii))
		}

		appinst := edgeproto.AppInst{}
		appinst.Key.AppKey.Organization = org
		appinst.Key.AppKey.Name = name
		if a == Create {
			s.AppInstCache.Update(ctx, &appinst, int64(ii))
		} else if a == Delete {
			s.AppInstCache.Delete(ctx, &appinst, int64(ii))
		}

		cinst := edgeproto.ClusterInst{}
		cinst.Key.Organization = org
		cinst.Key.ClusterKey.Name = name
		if a == Create {
			s.ClusterInstCache.Update(ctx, &cinst, int64(ii))
		} else if a == Delete {
			s.ClusterInstCache.Delete(ctx, &cinst, int64(ii))
		}

		resTagTbl := edgeproto.ResTagTable{}
		resTagTbl.Key.Name = name + "resTagTbl"
		resTagTbl.Key.Organization = org
		resTagTbl.Tags = map[string]string{
			"pci": "t4gpu:1",
		}
		if a == Create {
			s.ResTagTableCache.Update(ctx, &resTagTbl, int64(ii))
		} else if a == Delete {
			s.ResTagTableCache.Delete(ctx, &resTagTbl, int64(ii))
		}

		cloudlet := edgeproto.Cloudlet{}
		cloudlet.Key.Organization = org
		cloudlet.Key.Name = name
		cloudlet.EnvVar = map[string]string{"key1": "val1"}
		if a == Create {
			s.CloudletCache.Update(ctx, &cloudlet, int64(ii))
		} else if a == Delete {
			s.CloudletCache.Delete(ctx, &cloudlet, int64(ii))
		}

		cloudletInfo := edgeproto.CloudletInfo{}
		cloudletInfo.Key.Organization = org
		cloudletInfo.Key.Name = name
		cloudletInfo.ContainerVersion = "xyz"
		if a == Create {
			s.CloudletInfoCache.Update(ctx, &cloudletInfo, int64(ii))
		} else if a == Delete {
			s.CloudletInfoCache.Delete(ctx, &cloudletInfo, int64(ii))
		}

		pool := edgeproto.CloudletPool{}
		pool.Key.Name = name
		pool.Key.Organization = org
		pool.Cloudlets = []string{"cloudlet1", "cloudlet2", "cloudlet3"}
		if a == Create {
			s.CloudletPoolCache.Update(ctx, &pool, int64(ii))
		} else if a == Delete {
			s.CloudletPoolCache.Delete(ctx, &pool, int64(ii))
		}

		vmpool := edgeproto.VMPool{}
		vmpool.Key.Name = name
		vmpool.Key.Organization = org
		if a == Create {
			s.VMPoolCache.Update(ctx, &vmpool, int64(ii))
		} else if a == Delete {
			s.VMPoolCache.Delete(ctx, &vmpool, int64(ii))
		}

		gpuDriver := edgeproto.GPUDriver{}
		gpuDriver.Key.Name = name + "gpudriver"
		gpuDriver.Key.Organization = org
		if a == Create {
			s.GPUDriverCache.Update(ctx, &gpuDriver, int64(ii))
		} else if a == Delete {
			s.GPUDriverCache.Delete(ctx, &gpuDriver, int64(ii))
		}

		autoprov := edgeproto.AutoProvPolicy{}
		autoprov.Key.Name = name + "autoprov"
		autoprov.Key.Organization = org
		if a == Create {
			s.AutoProvPolicyCache.Update(ctx, &autoprov, int64(ii))
		} else if a == Delete {
			s.AutoProvPolicyCache.Delete(ctx, &autoprov, int64(ii))
		}

		autoscale := edgeproto.AutoScalePolicy{}
		autoscale.Key.Name = name + "autoscale"
		autoscale.Key.Organization = org
		if a == Create {
			s.AutoScalePolicyCache.Update(ctx, &autoscale, int64(ii))
		} else if a == Delete {
			s.AutoScalePolicyCache.Delete(ctx, &autoscale, int64(ii))
		}

		priv := edgeproto.TrustPolicy{}
		priv.Key.Name = name + "trust"
		priv.Key.Organization = org
		if a == Create {
			s.TrustPolicyCache.Update(ctx, &priv, int64(ii))
		} else if a == Delete {
			s.TrustPolicyCache.Delete(ctx, &priv, int64(ii))
		}
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
