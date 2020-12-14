package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type ClusterInstInfoApi struct {
	sync  *Sync
	store edgeproto.ClusterInstInfoStore
}

var clusterInstInfoApi = ClusterInstInfoApi{}

func InitClusterInstInfoApi(sync *Sync) {
	clusterInstInfoApi.sync = sync
	clusterInstInfoApi.store = edgeproto.NewClusterInstInfoStore(sync.store)
}

func (s *ClusterInstInfoApi) Update(ctx context.Context, in *edgeproto.ClusterInstInfo, rev int64) {
	clusterInstApi.UpdateFromInfo(ctx, in)
}

func (s *ClusterInstInfoApi) Delete(ctx context.Context, in *edgeproto.ClusterInstInfo, rev int64) {
	// for backwards compatibility
	clusterInstApi.DeleteFromInfo(ctx, in)
}

func (s *ClusterInstInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *ClusterInstInfoApi) Prune(ctx context.Context, keys map[edgeproto.ClusterInstKey]struct{}) {
	// no-op
}
