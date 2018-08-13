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
	sync.RegisterCache(&cloudletApi.cache)
}

func (s *CloudletApi) GetAllKeys(keys map[edgeproto.CloudletKey]struct{}) {
	s.cache.GetAllKeys(keys)
}

func (s *CloudletApi) Get(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool {
	return s.cache.Get(key, buf)
}

func (s *CloudletApi) HasKey(key *edgeproto.CloudletKey) bool {
	return s.cache.HasKey(key)
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
	if in.IpSupport == edgeproto.IpSupport_IpSupportUnknown {
		in.IpSupport = edgeproto.IpSupport_IpSupportDynamic
	}
	// TODO: support static IP assignment.
	if in.IpSupport != edgeproto.IpSupport_IpSupportDynamic {
		return &edgeproto.Result{}, errors.New("Only dynamic IPs are supported currently")
	}
	if in.IpSupport == edgeproto.IpSupport_IpSupportStatic {
		// TODO: Validate static ips
	} else {
		// dynamic
		if in.NumDynamicIps < 1 {
			return &edgeproto.Result{}, errors.New("Must specify at least one dynamic public IP available")
		}
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
	clDynInsts := make(map[edgeproto.ClusterInstKey]struct{})
	if clusterInstApi.UsesCloudlet(&in.Key, clDynInsts) {
		return &edgeproto.Result{}, errors.New("Cloudlet in use by static Cluster Instance")
	}
	res, err := s.store.Delete(in, s.sync.syncWait)
	// also delete associated info
	// Note: don't delete cloudletinfo, that will get deleted once CRM
	// disconnects. Otherwise if admin deletes/recreates Cloudlet with
	// CRM connected the whole time, we will end up without cloudletInfo.
	cloudletRefsApi.Delete(&in.Key, s.sync.syncWait)
	// also delete dynamic instances
	if len(dynInsts) > 0 {
		// delete dynamic instances
		for key, _ := range dynInsts {
			appInst := edgeproto.AppInst{Key: key}
			_, derr := appInstApi.DeleteNoWait(ctx, &appInst)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic app inst",
					"key", key, "err", derr)
			}
		}
	}
	if len(clDynInsts) > 0 {
		for key, _ := range clDynInsts {
			clInst := edgeproto.ClusterInst{Key: key}
			_, derr := clusterInstApi.DeleteNoWait(ctx, &clInst)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic cluster inst",
					"key", key, "err", derr)
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
				old := inst
				inst.CloudletLoc = new.Location
				if appInstApi.cache.NotifyCb != nil {
					appInstApi.cache.NotifyCb(inst.GetKey(), old)
				}
			}
		}
		appInstApi.cache.Mux.Unlock()
	}
}
