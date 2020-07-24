package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type VMPoolInfoApi struct {
	sync  *Sync
	store edgeproto.VMPoolInfoStore
	cache edgeproto.VMPoolInfoCache
}

var vmPoolInfoApi = VMPoolInfoApi{}

func InitVMPoolInfoApi(sync *Sync) {
	vmPoolInfoApi.sync = sync
	vmPoolInfoApi.store = edgeproto.NewVMPoolInfoStore(sync.store)
	edgeproto.InitVMPoolInfoCache(&vmPoolInfoApi.cache)
	sync.RegisterCache(&vmPoolInfoApi.cache)
}

func (s *VMPoolInfoApi) Update(ctx context.Context, in *edgeproto.VMPoolInfo, rev int64) {
	vmPoolApi.UpdateFromInfo(ctx, in)
}

func (s *VMPoolInfoApi) Delete(ctx context.Context, in *edgeproto.VMPoolInfo, rev int64) {
	s.store.Delete(ctx, in, nil)
}

func (s *VMPoolInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *VMPoolInfoApi) Prune(ctx context.Context, keys map[edgeproto.VMPoolKey]struct{}) {
	// no-op
}
