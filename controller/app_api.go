// app config

package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
)

// Should only be one of these instantiated in main
type AppApi struct {
	store  edgeproto.AppStore
	cache  *edgeproto.AppCache
	devApi *DeveloperApi
}

func InitAppApi(objStore objstore.ObjStore, devApi *DeveloperApi) *AppApi {
	api := &AppApi{
		store:  edgeproto.NewAppStore(objStore),
		devApi: devApi,
	}
	api.cache = edgeproto.NewAppCache(&api.store)
	return api
}

func (s *AppApi) WaitInitDone() {
	s.cache.WaitInitSyncDone()
}

func (s *AppApi) HasApp(key *edgeproto.AppKey) bool {
	return s.cache.HasKey(key)
}

func (s *AppApi) CreateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	if !s.devApi.HasDeveloper(&in.Key.DeveloperKey) {
		return &edgeproto.Result{}, errors.New("Specified developer not found")
	}
	return s.store.Create(in, s.cache.SyncWait)
}

func (s *AppApi) UpdateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	return s.store.Update(in, s.cache.SyncWait)
}

func (s *AppApi) DeleteApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	return s.store.Delete(in, s.cache.SyncWait)
}

func (s *AppApi) ShowApp(in *edgeproto.App, cb edgeproto.AppApi_ShowAppServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.App) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
