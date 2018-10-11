package main

import (
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

func (s *ClusterInstInfoApi) Update(in *edgeproto.ClusterInstInfo, rev int64) {
	clusterInstApi.UpdateFromInfo(in)
}

func (s *ClusterInstInfoApi) Delete(in *edgeproto.ClusterInstInfo, rev int64) {
	clusterInstApi.DeleteFromInfo(in)
}

func (s *ClusterInstInfoApi) Flush(notifyId int64) {
	// no-op
}
