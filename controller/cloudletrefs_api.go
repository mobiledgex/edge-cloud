package main

import "github.com/mobiledgex/edge-cloud/edgeproto"

type CloudletRefsApi struct {
	sync  *Sync
	store edgeproto.CloudletRefsStore
	cache edgeproto.CloudletRefsCache
}

var cloudletRefsApi = CloudletRefsApi{}

func InitCloudletRefsApi(sync *Sync) {
	cloudletRefsApi.sync = sync
	cloudletRefsApi.store = edgeproto.NewCloudletRefsStore(sync.store)
	edgeproto.InitCloudletRefsCache(&cloudletRefsApi.cache)
	sync.RegisterCache(&cloudletRefsApi.cache)
}

func (s *CloudletRefsApi) Delete(key *edgeproto.CloudletKey, wait func(int64)) {
	in := edgeproto.CloudletRefs{Key: *key}
	s.store.Delete(&in, wait)
}

func (s *CloudletRefsApi) ShowCloudletRefs(in *edgeproto.CloudletRefs, cb edgeproto.CloudletRefsApi_ShowCloudletRefsServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletRefs) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func initCloudletRefs(refs *edgeproto.CloudletRefs, key *edgeproto.CloudletKey) {
	refs.Key = *key
	refs.RootLbPorts = make(map[int32]edgeproto.CloudletRefsPortProto)
}
