package main

import (
	"context"
	"errors"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
)

type AppInstApi struct {
	sync        *Sync
	store       edgeproto.AppInstStore
	cache       edgeproto.AppInstCache
	appApi      *AppApi
	cloudletApi *CloudletApi
}

var appInstApi = AppInstApi{}

func InitAppInstApi(sync *Sync) {
	appInstApi.sync = sync
	appInstApi.store = edgeproto.NewAppInstStore(sync.store)
	edgeproto.InitAppInstCache(&appInstApi.cache)
	appInstApi.cache.SetNotifyCb(notify.ServerMgrOne.UpdateAppInst)
	sync.RegisterCache(&appInstApi.cache)
}

func (s *AppInstApi) GetAllKeys(keys map[edgeproto.AppInstKey]struct{}) {
	s.cache.GetAllKeys(keys)
}

func (s *AppInstApi) Get(key *edgeproto.AppInstKey, val *edgeproto.AppInst) bool {
	return s.cache.Get(key, val)
}

func (s *AppInstApi) HasKey(key *edgeproto.AppInstKey) bool {
	return s.cache.HasKey(key)
}

func (s *AppInstApi) UsesCloudlet(in *edgeproto.CloudletKey, dynInsts map[edgeproto.AppInstKey]struct{}) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	static := false
	for key, val := range s.cache.Objs {
		if key.CloudletKey.Matches(in) {
			if val.Liveness == edgeproto.AppInst_STATIC {
				static = true
			} else if val.Liveness == edgeproto.AppInst_DYNAMIC {
				dynInsts[key] = struct{}{}
			}
		}
	}
	return static
}

func (s *AppInstApi) UsesApp(in *edgeproto.AppKey, dynInsts map[edgeproto.AppInstKey]struct{}) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	static := false
	for key, val := range s.cache.Objs {
		if key.AppKey.Matches(in) {
			if val.Liveness == edgeproto.AppInst_STATIC {
				static = true
			} else if val.Liveness == edgeproto.AppInst_DYNAMIC {
				dynInsts[key] = struct{}{}
			}
		}
	}
	return static
}

func (s *AppInstApi) CreateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	// cache app path in app inst
	var app edgeproto.App
	if !appApi.Get(&in.Key.AppKey, &app) {
		return &edgeproto.Result{}, errors.New("Specified app not found")
	} else {
		in.AppPath = app.AppPath
	}
	// cache location of cloudlet in app inst
	var cloudlet edgeproto.Cloudlet
	if !cloudletApi.Get(&in.Key.CloudletKey, &cloudlet) {
		return &edgeproto.Result{}, errors.New("Specified cloudlet not found")
	} else {
		in.CloudletLoc = cloudlet.Location
	}
	return s.store.Create(in, s.sync.syncWait)
}

func (s *AppInstApi) UpdateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	// don't allow updates to cached fields
	if in.Fields != nil {
		for _, field := range in.Fields {
			if field == edgeproto.AppInstFieldAppPath {
				return &edgeproto.Result{}, errors.New("Cannot specify app path as it is inherited from specified app")
			} else if strings.HasPrefix(field, edgeproto.AppInstFieldCloudletLoc) {
				return &edgeproto.Result{}, errors.New("Cannot specify cloudlet location fields as they are inherited from specified cloudlet")
			}
		}
	}
	return s.store.Update(in, s.sync.syncWait)
}

func (s *AppInstApi) DeleteAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	resp, err := s.store.Delete(in, s.sync.syncWait)
	// also delete associated info
	appInstInfoApi.Del(&in.Key, s.sync.syncWait)
	return resp, err
}

func (s *AppInstApi) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
