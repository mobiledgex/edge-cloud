package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
)

type CloudletApi struct {
	sync  *Sync
	store edgeproto.CloudletStore
	cache edgeproto.CloudletCache
}

var cloudletApi = CloudletApi{}

func InitCloudletApi(sync *Sync) {
	cloudletApi.sync = sync
	cloudletApi.store = edgeproto.NewCloudletStore(sync.store)
	edgeproto.InitCloudletCache(&cloudletApi.cache)
	cloudletApi.cache.SetNotifyCb(notify.UpdateCloudlet)
	cloudletApi.cache.SetUpdatedCb(cloudletApi.UpdatedCb)
	sync.RegisterCache(edgeproto.CloudletKeyTypeString(), &cloudletApi.cache)
}

func (s *CloudletApi) GetAllKeys(keys map[edgeproto.CloudletKey]struct{}) {
	s.cache.GetAllKeys(keys)
}

func (s *CloudletApi) GetCloudlet(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool {
	return s.cache.Get(key, buf)
}

func (s *CloudletApi) CreateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	if !operatorApi.HasOperator(&in.Key.OperatorKey) {
		return &edgeproto.Result{}, errors.New("Specified cloudlet operator not found")
	}
	return s.store.Create(in, s.sync.syncWait)
}

func (s *CloudletApi) UpdateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	return s.store.Update(in, s.sync.syncWait)
}

func (s *CloudletApi) DeleteCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	return s.store.Delete(in, s.sync.syncWait)
}

func (s *CloudletApi) ShowCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_ShowCloudletServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Cloudlet) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *CloudletApi) UpdatedCb(old *edgeproto.Cloudlet, new *edgeproto.Cloudlet) {
	if old == nil {
		return
	}
	if old.Location.Lat != new.Location.Lat ||
		old.Location.Long != new.Location.Long ||
		old.Location.Altitude != new.Location.Altitude {

		appInstApi.cache.Mux.Lock()
		for _, inst := range appInstApi.cache.Objs {
			if inst.Key.CloudletKey.Matches(&new.Key) {
				inst.CloudletLoc = new.Location
				if appInstApi.cache.NotifyCb != nil {
					appInstApi.cache.NotifyCb(inst.GetKey())
				}
			}
		}
		appInstApi.cache.Mux.Unlock()
	}
}
