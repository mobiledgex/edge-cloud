package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstClientApi struct {
	sync  *Sync
	store edgeproto.AppInstClientStore
	cache edgeproto.AppInstClientCache
}

var appInstClientApi = AppInstClientApi{}

func InitAppInstClientApi(sync *Sync) {
	appInstClientApi.sync = sync
	appInstClientApi.store = edgeproto.NewAppInstClientStore(sync.store)
	edgeproto.InitAppInstClientCache(&appInstClientApi.cache)
	sync.RegisterCache(&appInstClientApi.cache)
}

func (s *AppInstClientApi) Prune(ctx context.Context, keys map[edgeproto.AppInstClientKey]struct{}) {}

// TODO: stream forever
func (s *AppInstClientApi) ShowAppInstClient(in *edgeproto.AppInstClient, cb edgeproto.AppInstClientApi_ShowAppInstClientServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInstClient) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AppInstClientApi) Update(ctx context.Context, in *edgeproto.AppInstClient, rev int64) {
	s.cache.Update(ctx, in, rev)
}

func (s *AppInstClientApi) Delete(ctx context.Context, in *edgeproto.AppInstClient, rev int64) {
	s.cache.Delete(ctx, in, rev)
}

func (s *AppInstClientApi) Flush(ctx context.Context, notifyId int64) {
	s.cache.Flush(ctx, notifyId)
}
