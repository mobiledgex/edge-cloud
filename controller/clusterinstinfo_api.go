package main

import "github.com/mobiledgex/edge-cloud/edgeproto"

type ClusterInstInfoApi struct {
	sync  *Sync
	store edgeproto.ClusterInstInfoStore
	cache edgeproto.ClusterInstInfoCache
}

var clusterInstInfoApi = ClusterInstInfoApi{}

func InitClusterInstInfoApi(sync *Sync) {
	clusterInstInfoApi.sync = sync
	clusterInstInfoApi.store = edgeproto.NewClusterInstInfoStore(sync.store)
	edgeproto.InitClusterInstInfoCache(&clusterInstInfoApi.cache)
	sync.RegisterCache(&clusterInstInfoApi.cache)
}

func (s *ClusterInstInfoApi) ShowClusterInstInfo(in *edgeproto.ClusterInstInfo, cb edgeproto.ClusterInstInfoApi_ShowClusterInstInfoServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.ClusterInstInfo) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *ClusterInstInfoApi) Update(in *edgeproto.ClusterInstInfo, notifyId int64) {
	if !clusterInstApi.HasKey(in.GetKey()) {
		return
	}
	// for now assume all fields have been specified
	in.Fields = edgeproto.ClusterInstInfoAllFields
	s.store.Put(in, nil)
}

func (s *ClusterInstInfoApi) internalDelete(key *edgeproto.ClusterInstKey, wait func(int64)) {
	in := edgeproto.ClusterInstInfo{Key: *key}
	s.store.Delete(&in, wait)
}

// Delete for notify never actually deletes the data
func (s *ClusterInstInfoApi) Delete(in *edgeproto.ClusterInstInfo, notifyId int64) {
	// no-op
}

func (s *ClusterInstInfoApi) Flush(notifyId int64) {
	// no-op
}
