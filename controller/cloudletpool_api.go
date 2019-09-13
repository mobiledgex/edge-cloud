package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type CloudletPoolApi struct {
	sync  *Sync
	store edgeproto.CloudletPoolStore
	cache edgeproto.CloudletPoolCache
}

var cloudletPoolApi = CloudletPoolApi{}

func InitCloudletPoolApi(sync *Sync) {
	cloudletPoolApi.sync = sync
	cloudletPoolApi.store = edgeproto.NewCloudletPoolStore(sync.store)
	edgeproto.InitCloudletPoolCache(&cloudletPoolApi.cache)
	sync.RegisterCache(&cloudletPoolApi.cache)
}

func (s *CloudletPoolApi) registerPublicPool(ctx context.Context) error {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		pool := edgeproto.CloudletPool{}
		pool.Key.Name = cloudcommon.PublicCloudletPool
		if s.store.STMGet(stm, &pool.Key, &pool) {
			// already present
			return nil
		}
		s.store.STMPut(stm, &pool)
		return nil
	})
	return err
}

func (s *CloudletPoolApi) CreateCloudletPool(ctx context.Context, in *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	if err := in.Validate(edgeproto.CloudletPoolAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyExists
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) DeleteCloudletPool(ctx context.Context, in *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	if in.Key.Name == cloudcommon.PublicCloudletPool {
		return &edgeproto.Result{}, fmt.Errorf("cannot delete Public pool")
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyNotFound
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) ShowCloudletPool(in *edgeproto.CloudletPool, cb edgeproto.CloudletPoolApi_ShowCloudletPoolServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletPool) error {
		// Don't show members because they're stored in json format.
		// This is a restriction of using proto buffers where the
		// map key cannot be a struct.
		showObj := *obj
		showObj.Members = nil
		err := cb.Send(&showObj)
		return err
	})
	return err
}

func (s *CloudletPoolApi) AddCloudletPoolMember(ctx context.Context, member *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		pool := edgeproto.CloudletPool{}
		if !s.store.STMGet(stm, &member.PoolKey, &pool) {
			return fmt.Errorf("Specified cloudlet pool not found")
		}
		if !cloudletApi.store.STMGet(stm, &member.CloudletKey, nil) {
			return fmt.Errorf("Specified cloudlet not found")
		}
		if pool.Members == nil {
			pool.Members = make(map[string]string)
		}
		memberStr := member.CloudletKey.GetKeyString()
		if _, found := pool.Members[memberStr]; found {
			return fmt.Errorf("Already a member of the pool")
		}
		pool.Members[memberStr] = ""
		s.store.STMPut(stm, &pool)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) RemoveCloudletPoolMember(ctx context.Context, member *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		pool := edgeproto.CloudletPool{}
		if !s.store.STMGet(stm, &member.PoolKey, &pool) {
			return objstore.ErrKVStoreKeyNotFound
		}
		memberStr := member.CloudletKey.GetKeyString()
		if _, found := pool.Members[memberStr]; !found {
			return fmt.Errorf("Specified cloudlet not a member")
		}
		delete(pool.Members, memberStr)
		s.store.STMPut(stm, &pool)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) ShowCloudletPoolMember(in *edgeproto.CloudletPoolMember, cb edgeproto.CloudletPoolMemberApi_ShowCloudletPoolMemberServer) error {
	poolFilter := edgeproto.CloudletPool{}
	poolFilter.Key = in.PoolKey
	ctx := cb.Context()

	err := s.cache.Show(&poolFilter, func(obj *edgeproto.CloudletPool) error {
		members := make([]edgeproto.CloudletPoolMember, 0)
		for memberStr, _ := range obj.Members {
			member := edgeproto.CloudletPoolMember{}
			err := json.Unmarshal([]byte(memberStr), &member.CloudletKey)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelApi, "unable to unmarshal CloudletPoolMember", "str", memberStr, "err", err)
				continue
			}
			member.PoolKey = obj.Key

			if !member.Matches(in, edgeproto.MatchFilter()) {
				continue
			}
			members = append(members, member)
		}
		sort.Slice(members, func(i, j int) bool {
			return members[i].CloudletKey.GetKeyString() < members[j].CloudletKey.GetKeyString()
		})

		for _, member := range members {
			err := cb.Send(&member)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (s *CloudletPoolApi) ShowPoolsForCloudlet(in *edgeproto.CloudletKey, cb edgeproto.CloudletPoolMemberApi_ShowPoolsForCloudletServer) error {
	str := in.GetKeyString()

	filter := edgeproto.CloudletPool{}
	err := s.cache.Show(&filter, func(obj *edgeproto.CloudletPool) error {
		if _, found := obj.Members[str]; !found {
			return nil
		}
		showObj := *obj
		showObj.Members = nil
		err := cb.Send(&showObj)
		return err
	})
	return err
}

func (s *CloudletPoolApi) ShowCloudletsForPool(in *edgeproto.CloudletPoolKey, cb edgeproto.CloudletPoolMemberApi_ShowCloudletsForPoolServer) error {
	members, err := s.getCloudletKeysForPools(in.Name)
	if err != nil {
		return err
	}
	err = cloudletApi.showCloudletsByKeys(members, func(obj *edgeproto.Cloudlet) error {
		return cb.Send(obj)
	})
	return err
}

func (s *CloudletPoolApi) ShowCloudletsForPoolList(in *edgeproto.CloudletPoolList, cb edgeproto.CloudletPoolMemberApi_ShowCloudletsForPoolListServer) error {
	if len(in.PoolName) == 0 {
		return fmt.Errorf("No pool names specified")
	}

	members, err := s.getCloudletKeysForPools(in.PoolName...)
	if err != nil {
		return err
	}
	err = cloudletApi.showCloudletsByKeys(members, func(obj *edgeproto.Cloudlet) error {
		return cb.Send(obj)
	})
	return err
}

func (s *CloudletPoolApi) getCloudletKeysForPools(names ...string) (map[string]struct{}, error) {
	members := make(map[string]struct{})

	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()

	for _, name := range names {
		poolKey := edgeproto.CloudletPoolKey{}
		poolKey.Name = name
		pool, found := s.cache.Objs[poolKey]
		if !found {
			return nil, fmt.Errorf("Pool %s not found", name)
		}
		for key, _ := range pool.Members {
			members[key] = struct{}{}
		}
	}
	return members, nil
}
