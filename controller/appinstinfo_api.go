package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstInfoApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.AppInstInfoStore
}

func NewAppInstInfoApi(sync *Sync, all *AllApis) *AppInstInfoApi {
	appInstInfoApi := AppInstInfoApi{}
	appInstInfoApi.all = all
	appInstInfoApi.sync = sync
	appInstInfoApi.store = edgeproto.NewAppInstInfoStore(sync.store)
	return &appInstInfoApi
}

func (s *AppInstInfoApi) Update(ctx context.Context, in *edgeproto.AppInstInfo, rev int64) {
	s.all.appInstApi.UpdateFromInfo(ctx, in)
}

func (s *AppInstInfoApi) Delete(ctx context.Context, in *edgeproto.AppInstInfo, rev int64) {
	// for backwards compatibility
	s.all.appInstApi.DeleteFromInfo(ctx, in)
}

func (s *AppInstInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *AppInstInfoApi) Prune(ctx context.Context, keys map[edgeproto.AppInstKey]struct{}) {
	// no-op
}
