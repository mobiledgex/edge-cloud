package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type ClusterApi struct {
	sync  *Sync
	store edgeproto.ClusterStore
	cache edgeproto.ClusterCache
}

var clusterApi = ClusterApi{}

const ClusterAutoPrefix = "autocluster"

var ClusterAutoPrefixErr = fmt.Sprintf("Cluster name prefix \"%s\" is reserved",
	ClusterAutoPrefix)

func InitClusterApi(sync *Sync) {
	clusterApi.sync = sync
	clusterApi.store = edgeproto.NewClusterStore(sync.store)
	edgeproto.InitClusterCache(&clusterApi.cache)
	sync.RegisterCache(&clusterApi.cache)
}

func (s *ClusterApi) UsesClusterFlavor(key *edgeproto.ClusterFlavorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, cluster := range s.cache.Objs {
		if cluster.DefaultFlavor.Matches(key) {
			return true
		}
	}
	return false
}

func (s *ClusterApi) HasKey(key *edgeproto.ClusterKey) bool {
	return s.cache.HasKey(key)
}

func (s *ClusterApi) Get(key *edgeproto.ClusterKey, buf *edgeproto.Cluster) bool {
	return s.cache.Get(key, buf)
}

func (s *ClusterApi) CreateCluster(ctx context.Context, in *edgeproto.Cluster) (*edgeproto.Result, error) {
	if strings.HasPrefix(in.Key.Name, ClusterAutoPrefix) {
		return &edgeproto.Result{}, errors.New(ClusterAutoPrefixErr)
	}
	if in.DefaultFlavor.Name != "" && !clusterFlavorApi.HasClusterFlavor(&in.DefaultFlavor) {
		return &edgeproto.Result{}, fmt.Errorf("default flavor %s not found", in.DefaultFlavor.Name)
	}
	in.Auto = false
	return s.store.Create(in, s.sync.syncWait)
}

func (s *ClusterApi) UpdateCluster(ctx context.Context, in *edgeproto.Cluster) (*edgeproto.Result, error) {
	// Unsupported for now
	return &edgeproto.Result{}, errors.New("Update cluster not supported")
	//return s.store.Update(in, s.sync.syncWait)
}

func (s *ClusterApi) DeleteCluster(ctx context.Context, in *edgeproto.Cluster) (*edgeproto.Result, error) {
	if clusterInstApi.UsesCluster(&in.Key) {
		return &edgeproto.Result{}, errors.New("Cluster in use by ClusterInst")
	}
	return s.store.Delete(in, s.sync.syncWait)
}

func (s *ClusterApi) ShowCluster(in *edgeproto.Cluster, cb edgeproto.ClusterApi_ShowClusterServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Cluster) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
