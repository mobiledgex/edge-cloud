package main

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

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
	// note: must be applied even if clusterinst doesn't exist, since this
	// will get called on error on delete.
	in.Fields = edgeproto.ClusterInstInfoAllFields
	s.store.Put(in, s.sync.syncWait)
}

func (s *ClusterInstInfoApi) Delete(in *edgeproto.ClusterInstInfo, notifyId int64) {
	s.store.Delete(in, s.sync.syncWait)
}

func (s *ClusterInstInfoApi) Flush(notifyId int64) {
	// XXX Set all states to NotConnected? Need to store notifyId in cache.
	// no-op
}
