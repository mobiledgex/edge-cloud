package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type DeveloperApi struct {
	store edgeproto.DeveloperStore
	cache *edgeproto.DeveloperCache
}

func InitDeveloperApi(objStore objstore.ObjStore) *DeveloperApi {
	api := &DeveloperApi{
		store: edgeproto.NewDeveloperStore(objStore),
	}
	api.cache = edgeproto.NewDeveloperCache(&api.store)
	return api
}

func (s *DeveloperApi) WaitInitDone() {
	s.cache.WaitInitSyncDone()
}

func (s *DeveloperApi) HasDeveloper(key *edgeproto.DeveloperKey) bool {
	return s.cache.HasKey(key)
}

func (s *DeveloperApi) CreateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	return s.store.Create(in, s.cache.SyncWait)
}

func (s *DeveloperApi) UpdateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	return s.store.Update(in, s.cache.SyncWait)
}

func (s *DeveloperApi) DeleteDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	return s.store.Delete(in, s.cache.SyncWait)
}

func (s *DeveloperApi) ShowDeveloper(in *edgeproto.Developer, cb edgeproto.DeveloperApi_ShowDeveloperServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Developer) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
