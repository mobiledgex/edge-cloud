package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
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
	return s.store.Put(ctx, in, s.sync.syncWait, objstore.WithLease(controllerAliveLease))
}

func (s *CloudletInfoApi) EvictCloudletInfo(ctx context.Context, in *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	return s.store.Delete(ctx, in, s.sync.syncWait)
}

func (s *CloudletInfoApi) ShowCloudletInfo(in *edgeproto.CloudletInfo, cb edgeproto.CloudletInfoApi_ShowCloudletInfoServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletInfo) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *CloudletInfoApi) UpdateAll(ctx context.Context, in *edgeproto.CloudletInfo) {
	in.Fields = edgeproto.CloudletInfoAllFields
	s.store.Put(ctx, in, nil, objstore.WithLease(controllerAliveLease))
}

func (s *CloudletInfoApi) Update(ctx context.Context, in *edgeproto.CloudletInfo, rev int64) {
	// for now assume all fields have been specified
	in.Fields = []string{}
	for _, field := range edgeproto.CloudletInfoAllFields {
		// Ignore outdated field, as it is managed by controller
		if field == edgeproto.CloudletInfoFieldOutdated {
			continue
		}
		in.Fields = append(in.Fields, field)
	}
	in.Controller = ControllerId
	s.store.Put(ctx, in, nil, objstore.WithLease(controllerAliveLease))
}

func (s *CloudletInfoApi) Del(ctx context.Context, key *edgeproto.CloudletKey, wait func(int64)) {
	in := edgeproto.CloudletInfo{Key: *key}
	s.store.Delete(ctx, &in, wait)
}

// Delete from notify just marks the cloudlet offline
func (s *CloudletInfoApi) Delete(ctx context.Context, in *edgeproto.CloudletInfo, rev int64) {
	buf := edgeproto.CloudletInfo{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
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

func (s *CloudletInfoApi) Flush(ctx context.Context, notifyId int64) {
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
		err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
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

func (s *CloudletInfoApi) Prune(ctx context.Context, keys map[edgeproto.CloudletKey]struct{}) {}

func (s *CloudletInfoApi) getCloudletState(key *edgeproto.CloudletKey) edgeproto.CloudletState {
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
	// Get tracked state, it could be that cloudlet has initiated
	// upgrade but cloudletInfo is yet to transition to it
	cloudlet := edgeproto.Cloudlet{}
	if !cloudletApi.cache.Get(key, &cloudlet) {
		return objstore.ErrKVStoreKeyNotFound
	}
	if cloudlet.State == edgeproto.TrackedState_UPDATE_REQUESTED ||
		cloudlet.State == edgeproto.TrackedState_UPDATING {
		return fmt.Errorf("Cloudlet %v is upgrading", key)
	}

	if cloudlet.State == edgeproto.TrackedState_UPDATE_ERROR {
		return fmt.Errorf("Cloudlet %v is in failed upgrade state, please upgrade it manually", key)
	}

	upgradeReq, err := isCloudletUpgradeRequired(key)
	if err != nil {
		if !*testMode {
			return fmt.Errorf("Cloudlet %v version check failed: %v", key, err)
		}
	}
	// Ignore cloudlet version check in testMode
	if !*testMode && upgradeReq {
		return fmt.Errorf("Cloudlet %v version is outdated, please upgrade it", key)
	}

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
