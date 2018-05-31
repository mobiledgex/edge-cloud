package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type CloudletApi struct {
	store       edgeproto.CloudletStore
	cache       *edgeproto.CloudletCache
	operatorApi *OperatorApi
}

func InitCloudletApi(objStore objstore.ObjStore, opApi *OperatorApi) *CloudletApi {
	api := &CloudletApi{
		store:       edgeproto.NewCloudletStore(objStore),
		operatorApi: opApi,
	}
	api.cache = edgeproto.NewCloudletCache(&api.store)
	api.cache.SetNotifyCb(notify.UpdateCloudlet)
	return api
}

func (s *CloudletApi) WaitInitDone() {
	s.cache.WaitInitSyncDone()
}

func (s *CloudletApi) GetAllKeys(keys map[edgeproto.CloudletKey]struct{}) {
	s.cache.GetAllKeys(keys)
}

func (s *CloudletApi) GetCloudlet(key *edgeproto.CloudletKey, val *edgeproto.Cloudlet) bool {
	return s.cache.Get(key, val)
}

func (s *CloudletApi) CreateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	if !s.operatorApi.HasOperator(&in.Key.OperatorKey) {
		return &edgeproto.Result{}, errors.New("Specified cloudlet operator not found")
	}
	return s.store.Create(in, s.cache.SyncWait)
}

func (s *CloudletApi) UpdateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	return s.store.Update(in, s.cache.SyncWait)
}

func (s *CloudletApi) DeleteCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	return s.store.Delete(in, s.cache.SyncWait)
}

func (s *CloudletApi) ShowCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_ShowCloudletServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Cloudlet) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
