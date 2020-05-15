package main

import (
	"context"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type CloudletPoolMemberApi struct {
	sync  *Sync
	store edgeproto.CloudletPoolMemberStore
	cache edgeproto.CloudletPoolMemberCache
}

var cloudletPoolMemberApi = CloudletPoolMemberApi{}

func InitCloudletPoolMemberApi(sync *Sync) {
	cloudletPoolMemberApi.sync = sync
	cloudletPoolMemberApi.store = edgeproto.NewCloudletPoolMemberStore(sync.store)
	edgeproto.InitCloudletPoolMemberCache(&cloudletPoolMemberApi.cache)
	sync.RegisterCache(&cloudletPoolMemberApi.cache)
}

func (s *CloudletPoolMemberApi) CreateCloudletPoolMember(ctx context.Context, in *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	if err := in.ValidateKey(); err != nil {
		return &edgeproto.Result{}, err
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, in, nil) {
			return in.ExistsError()
		}
		if !cloudletPoolApi.store.STMGet(stm, &in.PoolKey, nil) {
			return in.PoolKey.NotFoundError()
		}
		if !cloudletApi.store.STMGet(stm, &in.CloudletKey, nil) {
			return in.CloudletKey.NotFoundError()
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolMemberApi) DeleteCloudletPoolMember(ctx context.Context, in *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, in, nil) {
			return in.NotFoundError()
		}
		s.store.STMDel(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolMemberApi) ShowCloudletPoolMember(in *edgeproto.CloudletPoolMember, cb edgeproto.CloudletPoolMemberApi_ShowCloudletPoolMemberServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletPoolMember) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *CloudletPoolMemberApi) ShowPoolsForCloudlet(in *edgeproto.CloudletKey, cb edgeproto.CloudletPoolShowApi_ShowPoolsForCloudletServer) error {
	poolKeys := make(map[edgeproto.CloudletPoolKey]struct{})
	filter := edgeproto.CloudletPoolMember{
		CloudletKey: *in,
	}
	err := s.cache.Show(&filter, func(obj *edgeproto.CloudletPoolMember) error {
		poolKeys[obj.PoolKey] = struct{}{}
		return nil
	})
	if err != nil {
		return err
	}
	return cloudletPoolApi.showPoolsByKeys(poolKeys, cb.Send)
}

func (s *CloudletPoolMemberApi) ShowCloudletsForPool(in *edgeproto.CloudletPoolKey, cb edgeproto.CloudletPoolShowApi_ShowCloudletsForPoolServer) error {
	keys := s.getCloudletKeysForPools(in.Name)
	return cloudletApi.showCloudletsByKeys(keys, cb.Send)
}

func (s *CloudletPoolMemberApi) getCloudletKeysForPools(names ...string) map[edgeproto.CloudletKey]struct{} {
	keys := make(map[edgeproto.CloudletKey]struct{})

	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()

	for _, data := range s.cache.Objs {
		obj := data.Obj
		for _, name := range names {
			if name == obj.PoolKey.Name {
				keys[obj.CloudletKey] = struct{}{}
				break
			}
		}
	}
	return keys
}

// clean up CloudletPoolMembers after Cloudlet delete
func (s *CloudletPoolMemberApi) cloudletDeleted(ctx context.Context, in *edgeproto.CloudletKey) {
	members := make(map[edgeproto.CloudletPoolMember]struct{})
	s.cache.Mux.Lock()
	for member, _ := range s.cache.Objs {
		if member.CloudletKey.Matches(in) {
			members[member] = struct{}{}
		}
	}
	s.cache.Mux.Unlock()
	s.cleanup(ctx, members)
}

func (s *CloudletPoolMemberApi) poolDeleted(ctx context.Context, in *edgeproto.CloudletPoolKey) {
	members := make(map[edgeproto.CloudletPoolMember]struct{})
	s.cache.Mux.Lock()
	for member, _ := range s.cache.Objs {
		if member.PoolKey.Matches(in) {
			members[member] = struct{}{}
		}
	}
	s.cache.Mux.Unlock()
	s.cleanup(ctx, members)
}

func (s *CloudletPoolMemberApi) cleanup(ctx context.Context, members map[edgeproto.CloudletPoolMember]struct{}) {
	for obj, _ := range members {
		_, err := s.store.Delete(ctx, &obj, nil)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "cloudletDeleted failed to clean up", "member", obj, "err", err)
		}
	}
}
