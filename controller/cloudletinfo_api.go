package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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

func (s *CloudletInfoApi) InjectCloudletInfo(ctx context.Context, in *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	return s.store.Put(in, s.sync.syncWait)
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
	s.store.Put(in, nil)
}

func (s *CloudletInfoApi) Del(key *edgeproto.CloudletKey, wait func(int64)) {
	in := edgeproto.CloudletInfo{Key: *key}
	s.store.Delete(&in, wait)
}

// Delete from notify just marks the cloudlet offline
func (s *CloudletInfoApi) Delete(in *edgeproto.CloudletInfo, rev int64) {
	in.State = edgeproto.CloudletState_CloudletStateOffline
	in.Fields = []string{edgeproto.CloudletInfoFieldState}
	s.store.Put(in, nil)
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
			info.State = edgeproto.CloudletState_CloudletStateOffline
			log.DebugLog(log.DebugLevelNotify, "mark cloudlet offline", "key", matches[ii], "notifyid", notifyId)
			s.store.STMPut(stm, &info)
			return nil
		})
		if err != nil {
			log.DebugLog(log.DebugLevelNotify, "mark cloudlet offline", "key", matches[ii], "err", err)
		}
	}
}

func (s *CloudletInfoApi) getCloudletState(key *edgeproto.CloudletKey) edgeproto.CloudletState {
	if *key == cloudcommon.PublicCloudletKey {
		return edgeproto.CloudletState_CloudletStateReady
	}
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, obj := range s.cache.Objs {
		if key.Matches(&obj.Key) {
			return obj.State
		}
	}
	return edgeproto.CloudletState_CloudletStateNotPresent
}

func (s *CloudletInfoApi) checkCloudletReady(key *edgeproto.CloudletKey) error {
	// For testing, state is Errors due to openstack limits not found.
	// Errors state does indicate it is online, so consider it ok.
	state := s.getCloudletState(key)
	if state == edgeproto.CloudletState_CloudletStateReady ||
		state == edgeproto.CloudletState_CloudletStateErrors {
		return nil
	}
	return fmt.Errorf("Cloudlet %v not ready, state is %s", key,
		edgeproto.CloudletState_name[int32(state)])
}
