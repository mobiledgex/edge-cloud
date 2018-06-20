// app config

package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

// Should only be one of these instantiated in main
type AppApi struct {
	sync  *Sync
	store edgeproto.AppStore
	cache edgeproto.AppCache
}

var appApi = AppApi{}

func InitAppApi(sync *Sync) {
	appApi.sync = sync
	appApi.store = edgeproto.NewAppStore(sync.store)
	edgeproto.InitAppCache(&appApi.cache)
	sync.RegisterCache(edgeproto.AppKeyTypeString(), &appApi.cache)
	appApi.cache.SetUpdatedCb(appApi.UpdatedCb)
}

func (s *AppApi) HasApp(key *edgeproto.AppKey) bool {
	return s.cache.HasKey(key)
}

func (s *AppApi) GetApp(key *edgeproto.AppKey, buf *edgeproto.App) bool {
	return s.cache.Get(key, buf)
}

func (s *AppApi) CreateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	if !developerApi.HasDeveloper(&in.Key.DeveloperKey) {
		return &edgeproto.Result{}, errors.New("Specified developer not found")
	}
	return s.store.Create(in, s.sync.syncWait)
}

func (s *AppApi) UpdateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	return s.store.Update(in, s.sync.syncWait)
}

func (s *AppApi) DeleteApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	return s.store.Delete(in, s.sync.syncWait)
}

func (s *AppApi) ShowApp(in *edgeproto.App, cb edgeproto.AppApi_ShowAppServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.App) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AppApi) UpdatedCb(old *edgeproto.App, new *edgeproto.App) {
	if old == nil {
		return
	}
	if old.AppPath != new.AppPath {
		util.DebugLog(util.DebugLevelApi, "updating app path")
		appInstApi.cache.Mux.Lock()
		for _, inst := range appInstApi.cache.Objs {
			if inst.Key.AppKey.Matches(&new.Key) {
				inst.AppPath = new.AppPath
				if appInstApi.cache.NotifyCb != nil {
					appInstApi.cache.NotifyCb(inst.GetKey())
				}
			}
		}
		appInstApi.cache.Mux.Unlock()
	}
}
