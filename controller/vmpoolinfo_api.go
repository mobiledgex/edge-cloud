package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type VMPoolInfoApi struct {
	all   *AllApis
	cache edgeproto.VMPoolInfoCache
}

func NewVMPoolInfoApi(sync *Sync, all *AllApis) *VMPoolInfoApi {
	vmPoolInfoApi := VMPoolInfoApi{}
	vmPoolInfoApi.all = all
	edgeproto.InitVMPoolInfoCache(&vmPoolInfoApi.cache)
	return &vmPoolInfoApi
}

func (s *VMPoolInfoApi) Update(ctx context.Context, in *edgeproto.VMPoolInfo, rev int64) {
	s.all.vmPoolApi.UpdateFromInfo(ctx, in)
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
