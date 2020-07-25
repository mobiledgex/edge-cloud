package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type VMPoolInfoApi struct {
	cache edgeproto.VMPoolInfoCache
}

var vmPoolInfoApi = VMPoolInfoApi{}

func InitVMPoolInfoApi(sync *Sync) {
	edgeproto.InitVMPoolInfoCache(&vmPoolInfoApi.cache)
}

func (s *VMPoolInfoApi) Update(ctx context.Context, in *edgeproto.VMPoolInfo, rev int64) {
	vmPoolApi.UpdateFromInfo(ctx, in)
}

func (s *VMPoolInfoApi) Delete(ctx context.Context, in *edgeproto.VMPoolInfo, rev int64) {
	// no-op
}

func (s *VMPoolInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *VMPoolInfoApi) Prune(ctx context.Context, keys map[edgeproto.VMPoolKey]struct{}) {
	// no-op
}
