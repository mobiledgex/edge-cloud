package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
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
			return fmt.Errorf("Member already exists")
		}
		if !cloudletPoolApi.store.STMGet(stm, &in.PoolKey, nil) {
			return fmt.Errorf("Specified cloudlet pool not found")
		}
		if !cloudletApi.store.STMGet(stm, &in.CloudletKey, nil) {
			return fmt.Errorf("Specified cloudlet not found")
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolMemberApi) DeleteCloudletPoolMember(ctx context.Context, in *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, in, nil) {
			return objstore.ErrKVStoreKeyNotFound
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

func (s *CloudletPoolMemberApi) ShowPoolsForCloudlet(in *edgeproto.CloudletKey, cb edgeproto.CloudletPoolMemberApi_ShowPoolsForCloudletServer) error {
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

func (s *CloudletPoolMemberApi) ShowCloudletsForPool(in *edgeproto.CloudletPoolKey, cb edgeproto.CloudletPoolMemberApi_ShowCloudletsForPoolServer) error {
	keys := s.getCloudletKeysForPools(in.Name)
	return cloudletApi.showCloudletsByKeys(keys, cb.Send)
}

func (s *CloudletPoolMemberApi) ShowCloudletsForPoolList(in *edgeproto.CloudletPoolList, cb edgeproto.CloudletPoolMemberApi_ShowCloudletsForPoolListServer) error {
	if len(in.PoolName) == 0 {
		return fmt.Errorf("No pool names specified")
	}

	keys := s.getCloudletKeysForPools(in.PoolName...)
	return cloudletApi.showCloudletsByKeys(keys, cb.Send)
}

func (s *CloudletPoolMemberApi) getCloudletKeysForPools(names ...string) map[edgeproto.CloudletKey]struct{} {
	keys := make(map[edgeproto.CloudletKey]struct{})

	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()

	for _, obj := range s.cache.Objs {
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
		log.DebugLog(log.DebugLevelApi, "check obj", "key", member.CloudletKey)
		if member.CloudletKey.Matches(in) {
			log.DebugLog(log.DebugLevelApi, "obj matches")
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
