package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
)

type ClusterInstApi struct {
	sync  *Sync
	store edgeproto.ClusterInstStore
	cache edgeproto.ClusterInstCache
}

var clusterInstApi = ClusterInstApi{}

func InitClusterInstApi(sync *Sync) {
	clusterInstApi.sync = sync
	clusterInstApi.store = edgeproto.NewClusterInstStore(sync.store)
	edgeproto.InitClusterInstCache(&clusterInstApi.cache)
	clusterInstApi.cache.SetNotifyCb(notify.ServerMgrOne.UpdateClusterInst)
	sync.RegisterCache(&clusterInstApi.cache)
}

func (s *ClusterInstApi) HasKey(key *edgeproto.ClusterInstKey) bool {
	return s.cache.HasKey(key)
}

func (s *ClusterInstApi) Get(key *edgeproto.ClusterInstKey, buf *edgeproto.ClusterInst) bool {
	return s.cache.Get(key, buf)
}

func (s *ClusterInstApi) GetClusterInstsForCloudlets(cloudlets map[edgeproto.CloudletKey]struct{}, clusterInsts map[edgeproto.ClusterInstKey]struct{}) {
	s.cache.GetClusterInstsForCloudlets(cloudlets, clusterInsts)
}

func (s *ClusterInstApi) CreateClusterInst(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	in.Liveness = edgeproto.Liveness_LivenessStatic
	return s.createClusterInstInternal(in)
}

func (s *ClusterInstApi) createClusterInstInternal(in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	if in.Liveness == edgeproto.Liveness_LivenessUnknown {
		in.Liveness = edgeproto.Liveness_LivenessDynamic
	}
	if !cloudletApi.HasKey(&in.Key.CloudletKey) {
		return &edgeproto.Result{}, errors.New("Specified Cloudlet not found")
	}
	// cache data
	var cluster edgeproto.Cluster
	if !clusterApi.Get(&in.Key.ClusterKey, &cluster) {
		return &edgeproto.Result{}, errors.New("Specified Cluster not found")
	}
	in.Flavor = cluster.Flavor
	in.Nodes = cluster.Nodes

	return s.store.Create(in, s.sync.syncWait)
}

func (s *ClusterInstApi) UpdateClusterInst(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	// Unsupported for now
	return &edgeproto.Result{}, errors.New("Update cluster not supported")
	//return s.store.Update(in, s.sync.syncWait)
}

func (s *ClusterInstApi) DeleteClusterInst(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	return s.deleteClusterInstInternal(in)
}

func (s *ClusterInstApi) deleteClusterInstInternal(in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	if appInstApi.UsesClusterInst(&in.Key) {
		return &edgeproto.Result{}, errors.New("ClusterInst in use by Application Instance")
	}
	return s.store.Delete(in, s.sync.syncWait)
}

func (s *ClusterInstApi) ShowClusterInst(in *edgeproto.ClusterInst, cb edgeproto.ClusterInstApi_ShowClusterInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.ClusterInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
