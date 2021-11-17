package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type CloudletPoolApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.CloudletPoolStore
	cache *edgeproto.CloudletPoolCache
}

func NewCloudletPoolApi(sync *Sync, all *AllApis) *CloudletPoolApi {
	cloudletPoolApi := CloudletPoolApi{}
	cloudletPoolApi.all = all
	cloudletPoolApi.sync = sync
	cloudletPoolApi.store = edgeproto.NewCloudletPoolStore(sync.store)
	cloudletPoolApi.cache = nodeMgr.CloudletPoolLookup.GetCloudletPoolCache(node.NoRegion)
	sync.RegisterCache(cloudletPoolApi.cache)
	return &cloudletPoolApi
}

func (s *CloudletPoolApi) CreateCloudletPool(ctx context.Context, in *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	if err := in.Validate(edgeproto.CloudletPoolAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		if err := s.checkCloudletsExist(stm, in); err != nil {
			return err
		}
		in.CreatedAt = cloudcommon.TimeToTimestamp(time.Now())
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) DeleteCloudletPool(ctx context.Context, in *edgeproto.CloudletPool) (res *edgeproto.Result, reterr error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.CloudletPool{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		if cur.DeletePrepare {
			return in.Key.BeingDeletedError()
		}
		cur.DeletePrepare = true
		s.store.STMPut(stm, &cur)
		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}
	defer func() {
		if reterr == nil {
			return
		}
		undoErr := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			cur := edgeproto.CloudletPool{}
			if !s.store.STMGet(stm, &in.Key, &cur) {
				return nil
			}
			if cur.DeletePrepare {
				cur.DeletePrepare = false
				s.store.STMPut(stm, &cur)
			}
			return nil
		})
		if undoErr != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to undo delete prepare", "key", in.Key, "err", undoErr)
		}
	}()

	if tpeKey := s.all.trustPolicyExceptionApi.TrustPolicyExceptionForCloudletPoolKeyExists(&in.Key); tpeKey != nil {
		return &edgeproto.Result{}, fmt.Errorf("CloudletPool in use by Trust Policy Exception %s", tpeKey.GetKeyString())
	}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.NotFoundError()
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) UpdateCloudletPool(ctx context.Context, in *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.CloudletPool{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed := cur.CopyInFields(in)
		if err := cur.Validate(nil); err != nil {
			return err
		}
		if changed == 0 {
			return nil
		}
		if err := s.checkCloudletsExist(stm, &cur); err != nil {
			return err
		}
		if k := s.all.trustPolicyExceptionApi.TrustPolicyExceptionForCloudletPoolKeyExists(&in.Key); k != nil {
			return fmt.Errorf("Not allowed to update CloudletPool when TrustPolicyException %s is applied", k.GetKeyString())
		}
		cur.UpdatedAt = cloudcommon.TimeToTimestamp(time.Now())
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) checkCloudletsExist(stm concurrency.STM, in *edgeproto.CloudletPool) error {
	notFound := []string{}
	for _, name := range in.Cloudlets {
		key := edgeproto.CloudletKey{
			Name:         name,
			Organization: in.Key.Organization,
		}
		cloudlet := edgeproto.Cloudlet{}
		if !s.all.cloudletApi.store.STMGet(stm, &key, &cloudlet) {
			notFound = append(notFound, name)
		}
		if cloudlet.DeletePrepare {
			return key.BeingDeletedError()
		}
	}
	if len(notFound) > 0 {
		return fmt.Errorf("Cloudlets %s not found", strings.Join(notFound, ", "))
	}
	return nil
}

func (s *CloudletPoolApi) ShowCloudletPool(in *edgeproto.CloudletPool, cb edgeproto.CloudletPoolApi_ShowCloudletPoolServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletPool) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *CloudletPoolApi) AddCloudletPoolMember(ctx context.Context, in *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.CloudletPool{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		for _, name := range cur.Cloudlets {
			if name == in.CloudletName {
				return fmt.Errorf("Cloudlet already part of pool")
			}
		}
		cur.Cloudlets = append(cur.Cloudlets, in.CloudletName)
		if err := s.checkCloudletsExist(stm, &cur); err != nil {
			return err
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) RemoveCloudletPoolMember(ctx context.Context, in *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.CloudletPool{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed := false
		for ii, _ := range cur.Cloudlets {
			if cur.Cloudlets[ii] == in.CloudletName {
				cur.Cloudlets = append(cur.Cloudlets[:ii], cur.Cloudlets[ii+1:]...)
				changed = true
				break
			}
		}
		if !changed {
			return nil
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) GetCloudletPoolKeysForCloudletKey(in *edgeproto.CloudletKey) ([]edgeproto.CloudletPoolKey, error) {
	return s.cache.GetPoolsForCloudletKey(in)
}

func (s *CloudletPoolApi) HasCloudletPool(key *edgeproto.CloudletPoolKey) bool {
	return s.cache.HasKey(key)
}

func (s *CloudletPoolApi) validateCloudletPoolExists(key *edgeproto.CloudletPoolKey) bool {
	return s.HasCloudletPool(key)
}

func (s *CloudletPoolApi) UsesCloudlet(key *edgeproto.CloudletKey) []edgeproto.CloudletPoolKey {
	cpKeys := []edgeproto.CloudletPoolKey{}
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for cpKey, data := range s.cache.Objs {
		if cpKey.Organization != key.Organization {
			continue
		}
		pool := data.Obj
		for _, name := range pool.Cloudlets {
			if name == key.Name {
				cpKeys = append(cpKeys, cpKey)
				break
			}
		}
	}
	return cpKeys
}
