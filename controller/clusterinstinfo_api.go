package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type ClusterInstInfoApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.ClusterInstInfoStore
}

func NewClusterInstInfoApi(sync *Sync, all *AllApis) *ClusterInstInfoApi {
	clusterInstInfoApi := ClusterInstInfoApi{}
	clusterInstInfoApi.all = all
	clusterInstInfoApi.sync = sync
	clusterInstInfoApi.store = edgeproto.NewClusterInstInfoStore(sync.store)
	return &clusterInstInfoApi
}

func (s *ClusterInstInfoApi) Update(ctx context.Context, in *edgeproto.ClusterInstInfo, rev int64) {
	s.all.clusterInstApi.UpdateFromInfo(ctx, in)
}

func (s *ClusterInstInfoApi) Delete(ctx context.Context, in *edgeproto.ClusterInstInfo, rev int64) {
	// for backwards compatibility
	s.all.clusterInstApi.DeleteFromInfo(ctx, in)
}

func (s *ClusterInstInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *ClusterInstInfoApi) Prune(ctx context.Context, keys map[edgeproto.ClusterInstKey]struct{}) {
	// no-op
}
