package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type AppInstApi struct {
	store       edgeproto.AppInstStore
	cache       *edgeproto.AppInstCache
	appApi      *AppApi
	cloudletApi *CloudletApi
}

func InitAppInstApi(objStore objstore.ObjStore, appApi *AppApi, cloudletApi *CloudletApi) *AppInstApi {
	api := &AppInstApi{
		store:       edgeproto.NewAppInstStore(objStore),
		appApi:      appApi,
		cloudletApi: cloudletApi,
	}
	api.cache = edgeproto.NewAppInstCache(&api.store)
	api.cache.SetNotifyCb(notify.UpdateAppInst)
	return api
}

func (s *AppInstApi) WaitInitDone() {
	s.cache.WaitInitSyncDone()
}

func (s *AppInstApi) GetAllKeys(keys map[edgeproto.AppInstKey]struct{}) {
	s.cache.GetAllKeys(keys)
}

func (s *AppInstApi) GetAppInst(key *edgeproto.AppInstKey, val *edgeproto.AppInst) bool {
	return s.cache.Get(key, val)
}

func (s *AppInstApi) CreateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	// cache location of cloudlet in app inst
	if !s.appApi.HasApp(&in.Key.AppKey) {
		return &edgeproto.Result{}, errors.New("Specified app not found")
	}
	var cloudlet edgeproto.Cloudlet
	if s.cloudletApi.GetCloudlet(&in.Key.CloudletKey, &cloudlet) {
		in.CloudletLoc = cloudlet.Location
	}
	return s.store.Create(in, s.cache.SyncWait)
}

func (s *AppInstApi) UpdateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	return s.store.Update(in, s.cache.SyncWait)
}

func (s *AppInstApi) DeleteAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	return s.store.Delete(in, s.cache.SyncWait)
}

func (s *AppInstApi) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
