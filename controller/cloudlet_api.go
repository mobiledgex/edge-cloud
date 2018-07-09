package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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
	cloudletApi.cache.SetNotifyCb(notify.ServerMgrOne.UpdateCloudlet)
	cloudletApi.cache.SetUpdatedCb(cloudletApi.UpdatedCb)
	sync.RegisterCache(edgeproto.CloudletKeyTypeString(), &cloudletApi.cache)
}

func (s *CloudletApi) GetAllKeys(keys map[edgeproto.CloudletKey]struct{}) {
	s.cache.GetAllKeys(keys)
}

func (s *CloudletApi) Get(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool {
	return s.cache.Get(key, buf)
}

func (s *CloudletApi) UsesOperator(in *edgeproto.OperatorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, _ := range s.cache.Objs {
		if key.OperatorKey.Matches(in) {
			return true
		}
	}
	return false
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
	dynInsts := make(map[edgeproto.AppInstKey]struct{})
	if appInstApi.UsesCloudlet(&in.Key, dynInsts) {
		// disallow delete if static instances are present
		return &edgeproto.Result{}, errors.New("Cloudlet in use by static Application Instance")
	}
	res, err := s.store.Delete(in, s.sync.syncWait)
	if len(dynInsts) > 0 {
		// delete dynamic instances
		for key, _ := range dynInsts {
			appInst := edgeproto.AppInst{Key: key}
			_, derr := appInstApi.DeleteAppInst(ctx, &appInst)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic app inst",
					"err", derr)
			}
		}
	}
	return res, err
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
