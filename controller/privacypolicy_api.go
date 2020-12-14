package main

import (
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

func (s *PrivacyPolicyApi) CreatePrivacyPolicy(in *edgeproto.PrivacyPolicy, cb edgeproto.PrivacyPolicyApi_CreatePrivacyPolicyServer) error {
	ctx := cb.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "CreatePrivacyPolicy", "policy", in)

	// port range max is optional, set it to min if min is present but not max
	for i, o := range in.OutboundSecurityRules {
		if o.PortRangeMax == 0 {
			log.SpanLog(ctx, log.DebugLevelApi, "Setting PortRangeMax equal to min", "PortRangeMin", o.PortRangeMin)
			in.OutboundSecurityRules[i].PortRangeMax = o.PortRangeMin
		}
	}
	if err := in.Validate(nil); err != nil {
		return err
	}
	_, err := s.store.Create(ctx, in, s.sync.syncWait)
	return err

}

func (s *PrivacyPolicyApi) UpdatePrivacyPolicy(in *edgeproto.PrivacyPolicy, cb edgeproto.PrivacyPolicyApi_UpdatePrivacyPolicyServer) error {
	ctx := cb.Context()
	cur := edgeproto.PrivacyPolicy{}
	changed := 0

	// if there are cloudlets in sync state forbid this operation
	if cloudletApi.UsesPrivacyPolicy(&in.Key, edgeproto.TrackedState_UPDATING) {
		return fmt.Errorf("Policy in use by Cloudlet")
	}
	// port range max is optional, set it to min if min is present but not max
	for i, o := range in.OutboundSecurityRules {
		if o.PortRangeMax == 0 {
			log.SpanLog(ctx, log.DebugLevelApi, "Setting PortRangeMax equal to min", "PortRangeMin", o.PortRangeMin)
			in.OutboundSecurityRules[i].PortRangeMax = o.PortRangeMin
		}
	}

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
	if err != nil {
		return err
	}
	return cloudletApi.UpdateCloudletsUsingPrivacyPolicy(ctx, &cur, cb)
}

func (s *PrivacyPolicyApi) DeletePrivacyPolicy(in *edgeproto.PrivacyPolicy, cb edgeproto.PrivacyPolicyApi_DeletePrivacyPolicyServer) error {
	ctx := cb.Context()
	if !s.cache.HasKey(&in.Key) {
		return in.Key.NotFoundError()
	}
	// look for cloudlets in any state
	if cloudletApi.UsesPrivacyPolicy(&in.Key, edgeproto.TrackedState_TRACKED_STATE_UNKNOWN) {
		return fmt.Errorf("Policy in use by Cloudlet")
	}
	_, err := s.store.Delete(ctx, in, s.sync.syncWait)
	return err
}

func (s *PrivacyPolicyApi) ShowPrivacyPolicy(in *edgeproto.PrivacyPolicy, cb edgeproto.PrivacyPolicyApi_ShowPrivacyPolicyServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.PrivacyPolicy) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *PrivacyPolicyApi) STMFind(stm concurrency.STM, name, org string, policy *edgeproto.PrivacyPolicy) error {
	key := edgeproto.PolicyKey{}
	key.Name = name
	key.Organization = org
	if !s.store.STMGet(stm, &key, policy) {
		return fmt.Errorf("PrivacyPolicy %s for organization %s not found", name, org)
	}
	return nil
}

func (s *PrivacyPolicyApi) GetPrivacyPolicies(policies map[edgeproto.PolicyKey]*edgeproto.PrivacyPolicy) {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		pol := data.Obj
		policies[pol.Key] = pol
	}
}
