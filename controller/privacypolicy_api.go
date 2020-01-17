package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type PrivacyPolicyApi struct {
	sync  *Sync
	store edgeproto.PrivacyPolicyStore
	cache edgeproto.PrivacyPolicyCache
}

var privacyPolicyApi = PrivacyPolicyApi{}

func InitPrivacyPolicyApi(sync *Sync) {
	privacyPolicyApi.sync = sync
	privacyPolicyApi.store = edgeproto.NewPrivacyPolicyStore(sync.store)
	edgeproto.InitPrivacyPolicyCache(&privacyPolicyApi.cache)
	sync.RegisterCache(&privacyPolicyApi.cache)
}

func (s *PrivacyPolicyApi) CreatePrivacyPolicy(ctx context.Context, in *edgeproto.PrivacyPolicy) (*edgeproto.Result, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "CreatePrivacyPolicy", "policy", in)

	// port range max is optional, set it to min if min is present but not max
	for i, o := range in.OutboundSecurityRules {
		if o.PortRangeMax == 0 {
			log.SpanLog(ctx, log.DebugLevelApi, "Setting PortRangeMax equal to min", "PortRangeMin", o.PortRangeMin)
			in.OutboundSecurityRules[i].PortRangeMax = o.PortRangeMin
		}
	}
	if err := in.Validate(nil); err != nil {
		return &edgeproto.Result{}, err
	}
	return s.store.Create(ctx, in, s.sync.syncWait)
}

func (s *PrivacyPolicyApi) UpdatePrivacyPolicy(ctx context.Context, in *edgeproto.PrivacyPolicy) (*edgeproto.Result, error) {
	cur := edgeproto.PrivacyPolicy{}
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

func (s *PrivacyPolicyApi) DeletePrivacyPolicy(ctx context.Context, in *edgeproto.PrivacyPolicy) (*edgeproto.Result, error) {
	if !s.cache.HasKey(&in.Key) {
		return &edgeproto.Result{}, in.Key.NotFoundError()
	}
	if clusterInstApi.UsesPrivacyPolicy(&in.Key) {
		return &edgeproto.Result{}, fmt.Errorf("Policy in use by ClusterInst")
	}
	return s.store.Delete(ctx, in, s.sync.syncWait)
}

func (s *PrivacyPolicyApi) ShowPrivacyPolicy(in *edgeproto.PrivacyPolicy, cb edgeproto.PrivacyPolicyApi_ShowPrivacyPolicyServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.PrivacyPolicy) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *PrivacyPolicyApi) STMFind(stm concurrency.STM, name, dev string, policy *edgeproto.PrivacyPolicy) error {
	key := edgeproto.PolicyKey{}
	key.Name = name
	key.Developer = dev
	if !s.store.STMGet(stm, &key, policy) {
		return fmt.Errorf("PrivacyPolicy %s for developer %s not found", name, dev)
	}
	return nil
}
