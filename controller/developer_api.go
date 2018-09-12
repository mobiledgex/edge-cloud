package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type DeveloperApi struct {
	sync  *Sync
	store edgeproto.DeveloperStore
	cache edgeproto.DeveloperCache
}

var developerApi = DeveloperApi{}

func InitDeveloperApi(sync *Sync) {
	developerApi.sync = sync
	developerApi.store = edgeproto.NewDeveloperStore(sync.store)
	edgeproto.InitDeveloperCache(&developerApi.cache)
	sync.RegisterCache(&developerApi.cache)
}

func (s *DeveloperApi) HasDeveloper(key *edgeproto.DeveloperKey) bool {
	return s.cache.HasKey(key)
}

func (s *DeveloperApi) CreateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	return s.store.Create(in, s.sync.syncWait)
}

func (s *DeveloperApi) UpdateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	return s.store.Update(in, s.sync.syncWait)
}

func (s *DeveloperApi) DeleteDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	if appApi.UsesDeveloper(&in.Key) {
		return &edgeproto.Result{}, errors.New("Developer in use by Application")
	}
	return s.store.Delete(in, s.sync.syncWait)
}

func (s *DeveloperApi) ShowDeveloper(in *edgeproto.Developer, cb edgeproto.DeveloperApi_ShowDeveloperServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Developer) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
