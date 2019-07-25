package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstInfoApi struct {
	sync  *Sync
	store edgeproto.AppInstInfoStore
}

var appInstInfoApi = AppInstInfoApi{}

func InitAppInstInfoApi(sync *Sync) {
	appInstInfoApi.sync = sync
	appInstInfoApi.store = edgeproto.NewAppInstInfoStore(sync.store)
}

func (s *AppInstInfoApi) Update(ctx context.Context, in *edgeproto.AppInstInfo, rev int64) {
	appInstApi.UpdateFromInfo(ctx, in)
}

func (s *AppInstInfoApi) Delete(ctx context.Context, in *edgeproto.AppInstInfo, rev int64) {
	appInstApi.DeleteFromInfo(ctx, in)
}

func (s *AppInstInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *AppInstInfoApi) Prune(ctx context.Context, keys map[edgeproto.AppInstKey]struct{}) {
	// no-op
}
