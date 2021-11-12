package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AutoScalePolicyApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.AutoScalePolicyStore
	cache edgeproto.AutoScalePolicyCache
}

func NewAutoScalePolicyApi(sync *Sync, all *AllApis) *AutoScalePolicyApi {
	autoScalePolicyApi := AutoScalePolicyApi{}
	autoScalePolicyApi.all = all
	autoScalePolicyApi.sync = sync
	autoScalePolicyApi.store = edgeproto.NewAutoScalePolicyStore(sync.store)
	edgeproto.InitAutoScalePolicyCache(&autoScalePolicyApi.cache)
	sync.RegisterCache(&autoScalePolicyApi.cache)
	return &autoScalePolicyApi
}

func (s *AutoScalePolicyApi) CreateAutoScalePolicy(ctx context.Context, in *edgeproto.AutoScalePolicy) (*edgeproto.Result, error) {
	if err := in.Validate(nil); err != nil {
		return &edgeproto.Result{}, err
	}
	return s.store.Create(ctx, in, s.sync.syncWait)
}

func (s *AutoScalePolicyApi) UpdateAutoScalePolicy(ctx context.Context, in *edgeproto.AutoScalePolicy) (*edgeproto.Result, error) {
	cur := edgeproto.AutoScalePolicy{}
	changed := 0
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed = cur.CopyInFields(in)
		if err := cur.Validate(nil); err != nil {
			return err
		}
		if changed == 0 {
			return nil
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AutoScalePolicyApi) DeleteAutoScalePolicy(ctx context.Context, in *edgeproto.AutoScalePolicy) (*edgeproto.Result, error) {
	if !s.cache.HasKey(&in.Key) {
		return &edgeproto.Result{}, in.Key.NotFoundError()
	}
	if s.all.clusterInstApi.UsesAutoScalePolicy(&in.Key) {
		return &edgeproto.Result{}, fmt.Errorf("Policy in use by ClusterInst")
	}
	return s.store.Delete(ctx, in, s.sync.syncWait)
}

func (s *AutoScalePolicyApi) ShowAutoScalePolicy(in *edgeproto.AutoScalePolicy, cb edgeproto.AutoScalePolicyApi_ShowAutoScalePolicyServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AutoScalePolicy) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AutoScalePolicyApi) STMFind(stm concurrency.STM, name, dev string, policy *edgeproto.AutoScalePolicy) error {
	key := edgeproto.PolicyKey{}
	key.Name = name
	key.Organization = dev
	if !s.store.STMGet(stm, &key, policy) {
		return fmt.Errorf("AutoScalePolicy %s for developer %s not found", name, dev)
	}
	return nil
}
