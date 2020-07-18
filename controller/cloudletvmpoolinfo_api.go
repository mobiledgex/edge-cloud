package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type CloudletVMPoolInfoApi struct {
	sync  *Sync
	store edgeproto.CloudletVMPoolInfoStore
	cache edgeproto.CloudletVMPoolInfoCache
}

var cloudletVMPoolInfoApi = CloudletVMPoolInfoApi{}

func InitCloudletVMPoolInfoApi(sync *Sync) {
	cloudletVMPoolInfoApi.sync = sync
	cloudletVMPoolInfoApi.store = edgeproto.NewCloudletVMPoolInfoStore(sync.store)
	edgeproto.InitCloudletVMPoolInfoCache(&cloudletVMPoolInfoApi.cache)
	sync.RegisterCache(&cloudletVMPoolInfoApi.cache)
}

func (s *CloudletVMPoolInfoApi) Update(ctx context.Context, in *edgeproto.CloudletVMPoolInfo, rev int64) {
	s.store.Put(ctx, in, nil)
}

func (s *CloudletVMPoolInfoApi) Delete(ctx context.Context, in *edgeproto.CloudletVMPoolInfo, rev int64) {
	s.store.Delete(ctx, in, nil)
}

func (s *CloudletVMPoolInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *CloudletVMPoolInfoApi) Prune(ctx context.Context, keys map[edgeproto.CloudletKey]struct{}) {
	// no-op
}

func (s *CloudletVMPoolInfoApi) ShowCloudletVMPoolInfo(in *edgeproto.CloudletVMPoolInfo, cb edgeproto.CloudletVMPoolInfoApi_ShowCloudletVMPoolInfoServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletVMPoolInfo) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
