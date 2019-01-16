package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type ClusterFlavorApi struct {
	sync  *Sync
	store edgeproto.ClusterFlavorStore
	cache edgeproto.ClusterFlavorCache
}

var clusterFlavorApi = ClusterFlavorApi{}

func InitClusterFlavorApi(sync *Sync) {
	clusterFlavorApi.sync = sync
	clusterFlavorApi.store = edgeproto.NewClusterFlavorStore(sync.store)
	edgeproto.InitClusterFlavorCache(&clusterFlavorApi.cache)
	sync.RegisterCache(&clusterFlavorApi.cache)
}

func (s *ClusterFlavorApi) HasClusterFlavor(key *edgeproto.ClusterFlavorKey) bool {
	return s.cache.HasKey(key)
}

func (s *ClusterFlavorApi) Get(key *edgeproto.ClusterFlavorKey, buf *edgeproto.ClusterFlavor) bool {
	return s.cache.Get(key, buf)
}

func (s *ClusterFlavorApi) GetAllKeys(keys map[edgeproto.ClusterFlavorKey]struct{}) {
	s.cache.GetAllKeys(keys)
}

func (s *ClusterFlavorApi) UsesFlavor(key *edgeproto.FlavorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, flavor := range s.cache.Objs {
		if flavor.NodeFlavor.Matches(key) ||
			flavor.MasterFlavor.Matches(key) {
			return true
		}
	}
	return false
}

func (s *ClusterFlavorApi) CreateClusterFlavor(ctx context.Context, in *edgeproto.ClusterFlavor) (*edgeproto.Result, error) {
	if in.NodeFlavor.Name == "" {
		return &edgeproto.Result{}, errors.New("Please specify node flavor")
	}
	if in.MasterFlavor.Name == "" {
		return &edgeproto.Result{}, errors.New("Please specify master flavor")
	}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyExists
		}
		if !flavorApi.store.STMGet(stm, &in.NodeFlavor, nil) {
			return fmt.Errorf("Node flavor %s not found", in.NodeFlavor.Name)
		}
		if !flavorApi.store.STMGet(stm, &in.MasterFlavor, nil) {
			return fmt.Errorf("Master flavor %s not found", in.MasterFlavor.Name)
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *ClusterFlavorApi) UpdateClusterFlavor(ctx context.Context, in *edgeproto.ClusterFlavor) (*edgeproto.Result, error) {
	// Unsupported for now
	return &edgeproto.Result{}, errors.New("Update cluster flavor not supported")
	//return s.store.Update(in, s.sync.syncWait)
}

func (s *ClusterFlavorApi) DeleteClusterFlavor(ctx context.Context, in *edgeproto.ClusterFlavor) (*edgeproto.Result, error) {
	if clusterApi.UsesClusterFlavor(&in.Key) {
		return &edgeproto.Result{}, errors.New("ClusterFlavor in use by Cluster")
	}
	if clusterInstApi.UsesClusterFlavor(&in.Key) {
		return &edgeproto.Result{}, errors.New("ClusterFlavor in use by Cluster Instance")
	}
	return s.store.Delete(in, s.sync.syncWait)
}

func (s *ClusterFlavorApi) ShowClusterFlavor(in *edgeproto.ClusterFlavor, cb edgeproto.ClusterFlavorApi_ShowClusterFlavorServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.ClusterFlavor) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
