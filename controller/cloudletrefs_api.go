package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type CloudletRefsApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.CloudletRefsStore
	cache edgeproto.CloudletRefsCache
}

func NewCloudletRefsApi(sync *Sync, all *AllApis) *CloudletRefsApi {
	cloudletRefsApi := CloudletRefsApi{}
	cloudletRefsApi.all = all
	cloudletRefsApi.sync = sync
	cloudletRefsApi.store = edgeproto.NewCloudletRefsStore(sync.store)
	edgeproto.InitCloudletRefsCache(&cloudletRefsApi.cache)
	sync.RegisterCache(&cloudletRefsApi.cache)
	return &cloudletRefsApi
}

func (s *CloudletRefsApi) Delete(ctx context.Context, key *edgeproto.CloudletKey, wait func(int64)) {
	in := edgeproto.CloudletRefs{Key: *key}
	s.store.Delete(ctx, &in, wait)
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
	refs.RootLbPorts = make(map[int32]int32)
}
