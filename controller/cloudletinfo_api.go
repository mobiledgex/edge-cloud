package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type CloudletInfoApi struct {
	sync  *Sync
	store edgeproto.CloudletInfoStore
	cache edgeproto.CloudletInfoCache
}

var cloudletInfoApi = CloudletInfoApi{}

func InitCloudletInfoApi(sync *Sync) {
	cloudletInfoApi.sync = sync
	cloudletInfoApi.store = edgeproto.NewCloudletInfoStore(sync.store)
	edgeproto.InitCloudletInfoCache(&cloudletInfoApi.cache)
	sync.RegisterCache(&cloudletInfoApi.cache)
}

// We put CloudletInfo in etcd with a lease, so in case both controller
// and CRM suddenly go away, etcd will remove the stale CloudletInfo data.

func (s *CloudletInfoApi) InjectCloudletInfo(ctx context.Context, in *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	return s.store.Put(in, s.sync.syncWait, objstore.WithLease(controllerAliveLease))
}

func (s *CloudletInfoApi) EvictCloudletInfo(ctx context.Context, in *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	return s.store.Delete(in, s.sync.syncWait)
}

func (s *CloudletInfoApi) ShowCloudletInfo(in *edgeproto.CloudletInfo, cb edgeproto.CloudletInfoApi_ShowCloudletInfoServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletInfo) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *CloudletInfoApi) Update(in *edgeproto.CloudletInfo, rev int64) {
	// for now assume all fields have been specified
	in.Fields = edgeproto.CloudletInfoAllFields
	in.Controller = ControllerId
	s.store.Put(in, nil, objstore.WithLease(controllerAliveLease))
}

func (s *CloudletInfoApi) Del(key *edgeproto.CloudletKey, wait func(int64)) {
	in := edgeproto.CloudletInfo{Key: *key}
	s.store.Delete(&in, wait)
}

// Delete from notify just marks the cloudlet offline
func (s *CloudletInfoApi) Delete(in *edgeproto.CloudletInfo, rev int64) {
	buf := edgeproto.CloudletInfo{}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &buf) {
			return nil
		}
		if buf.NotifyId != in.NotifyId || buf.Controller != ControllerId {
			// updated by another thread or controller
			return nil
		}
		buf.State = edgeproto.CloudletState_CLOUDLET_STATE_OFFLINE
		buf.Fields = []string{edgeproto.CloudletInfoFieldState}
		s.store.STMPut(stm, &buf, objstore.WithLease(controllerAliveLease))
		return nil
	})
	if err != nil {
		log.DebugLog(log.DebugLevelNotify, "notify delete CloudletInfo",
			"key", in.Key, "err", err)
	}
}

func (s *CloudletInfoApi) Flush(notifyId int64) {
	// mark all cloudlets from the client as offline
	matches := make([]edgeproto.CloudletKey, 0)
	s.cache.Mux.Lock()
	for _, val := range s.cache.Objs {
		if val.NotifyId != notifyId || val.Controller != ControllerId {
			continue
		}
		matches = append(matches, val.Key)
	}
	s.cache.Mux.Unlock()

	info := edgeproto.CloudletInfo{}
	for ii, _ := range matches {
		info.Key = matches[ii]
		err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
			if s.store.STMGet(stm, &info.Key, &info) {
				if info.NotifyId != notifyId || info.Controller != ControllerId {
					// Updated by another thread or controller
					return nil
				}
			}
			info.State = edgeproto.CloudletState_CLOUDLET_STATE_OFFLINE
			log.DebugLog(log.DebugLevelNotify, "mark cloudlet offline", "key", matches[ii], "notifyid", notifyId)
			s.store.STMPut(stm, &info, objstore.WithLease(controllerAliveLease))
			return nil
		})
		if err != nil {
			log.DebugLog(log.DebugLevelNotify, "mark cloudlet offline", "key", matches[ii], "err", err)
		}
	}
}

func (s *CloudletInfoApi) Prune(keys map[edgeproto.CloudletKey]struct{}) {}

func (s *CloudletInfoApi) getCloudletState(key *edgeproto.CloudletKey) edgeproto.CloudletState {
	if *key == cloudcommon.DefaultCloudletKey {
		return edgeproto.CloudletState_CLOUDLET_STATE_READY
	}
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, obj := range s.cache.Objs {
		if key.Matches(&obj.Key) {
			return obj.State
		}
	}
	return edgeproto.CloudletState_CLOUDLET_STATE_NOT_PRESENT
}

func (s *CloudletInfoApi) checkCloudletReady(key *edgeproto.CloudletKey) error {
	// For testing, state is Errors due to openstack limits not found.
	// Errors state does indicate it is online, so consider it ok.
	state := s.getCloudletState(key)
	if state == edgeproto.CloudletState_CLOUDLET_STATE_READY ||
		state == edgeproto.CloudletState_CLOUDLET_STATE_ERRORS {
		return nil
	}
	return fmt.Errorf("Cloudlet %v not ready, state is %s", key,
		edgeproto.CloudletState_name[int32(state)])
}
