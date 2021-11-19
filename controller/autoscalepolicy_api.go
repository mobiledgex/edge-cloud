package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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

func (s *AutoScalePolicyApi) DeleteAutoScalePolicy(ctx context.Context, in *edgeproto.AutoScalePolicy) (res *edgeproto.Result, reterr error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.AutoScalePolicy{}
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
			cur := edgeproto.AutoScalePolicy{}
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

	if ciKey := s.all.clusterInstApi.UsesAutoScalePolicy(&in.Key); ciKey != nil {
		return &edgeproto.Result{}, fmt.Errorf("Policy in use by ClusterInst %s", ciKey.GetKeyString())
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
