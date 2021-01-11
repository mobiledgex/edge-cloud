package main

import (
	"context"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type ClusterInstRefsApi struct {
	sync  *Sync
	store edgeproto.ClusterInstRefsStore
	cache edgeproto.ClusterInstRefsCache
}

var clusterInstRefsApi = ClusterInstRefsApi{}

func InitClusterInstRefsApi(sync *Sync) {
	clusterInstRefsApi.sync = sync
	clusterInstRefsApi.store = edgeproto.NewClusterInstRefsStore(sync.store)
	edgeproto.InitClusterInstRefsCache(&clusterInstRefsApi.cache)
	sync.RegisterCache(&clusterInstRefsApi.cache)
}

func (s *ClusterInstRefsApi) ShowClusterInstRefs(in *edgeproto.ClusterInstRefs, cb edgeproto.ClusterInstRefsApi_ShowClusterInstRefsServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.ClusterInstRefs) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *ClusterInstRefsApi) CreateResourcesRef(stm concurrency.STM, key *edgeproto.ClusterInstKey, res []edgeproto.VMResource) {
	refs := edgeproto.ClusterInstRefs{}
	refs.Key = *key
	refs.ReservedResources = res
	s.store.STMPut(stm, &refs)
}

func (s *ClusterInstRefsApi) DeleteResourcesRef(ctx context.Context, key *edgeproto.ClusterInstKey) {
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		refs := edgeproto.ClusterInstRefs{}
		if !s.store.STMGet(stm, key, &refs) {
			// got deleted in the meantime
			return nil
		}
		s.store.STMDel(stm, key)
		return nil
	})
}
