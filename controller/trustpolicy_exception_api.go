package main

import (
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type TrustPolicyExceptionApi struct {
	sync  *Sync
	store edgeproto.TrustPolicyExceptionStore
	cache edgeproto.TrustPolicyExceptionCache
}

var trustPolicyExceptionApi = TrustPolicyExceptionApi{}

func InitTrustPolicyExcptionApi(sync *Sync) {
	trustPolicyApi.sync = sync
	trustPolicyApi.store = edgeproto.NewTrustPolicyStore(sync.store)
	edgeproto.InitTrustPolicyCache(&trustPolicyApi.cache)
	sync.RegisterCache(&trustPolicyApi.cache)
}

func (s *TrustPolicyExceptionApi) CreateTrustPolicyException(in *edgeproto.TrustPolicyException, cb edgeproto.TrustPolicyApi_CreateTrustPolicyServer) error {
	ctx := cb.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "CreateTrustPolicyException", "policy", in)

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

func (s *TrustPolicyExceptionApi) UpdateTrustPolicyException(in *edgeproto.TrustPolicyException, cb edgeproto.TrustPolicyExceptionApi_UpdateTrustPolicyExceptionServer) error {
	ctx := cb.Context()
	cur := edgeproto.TrustPolicyException{}
	changed := 0

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
		//if err := cloudletApi.ValidateCloudletsUsingTrustPolicy(ctx, &cur); err != nil {
		//	return err
		//}
		if changed == 0 {
			return nil
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
	//return cloudletApi.UpdateCloudletsUsingTrustPolicy(ctx, &cur, cb)
}

func (s *TrustPolicyExceptionApi) DeleteTrustExceptionPolicy(in *edgeproto.TrustPolicyException, cb edgeproto.TrustPolicyApi_DeleteTrustPolicyServer) error {
	ctx := cb.Context()
	if !s.cache.HasKey(&in.Key) {
		return in.Key.NotFoundError()
	}
	// look for cloudlets in any state
	//if cloudletApi.UsesTrustPolicy(&in.Key, edgeproto.TrackedState_TRACKED_STATE_UNKNOWN) {
	//	return fmt.Errorf("Policy in use by Cloudlet")
	//}
	_, err := s.store.Delete(ctx, in, s.sync.syncWait)
	return err
}

func (s *TrustPolicyExceptionApi) ShowTrustPolicyException(in *edgeproto.TrustPolicyException, cb edgeproto.TrustPolicyExceptionApi_ShowTrustPolicyExceptionServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.TrustPolicyException) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

/*func (s *TrustPolicyExceptionApi) STMFind(stm concurrency.STM, name, appOrg, cloudletOrg string, polic *edgeproto.TrustPolicyException) error {
	key := edgeproto.TrustPolicyExceptionKey{}
	key.Name = name
	key.Organization = org
	if !s.store.STMGet(stm, &key, policy) {
		return fmt.Errorf("TrustPolicy %s for organization %s not found", name, org)
	}
	return nil
}*/

func (s *TrustPolicyExceptionApi) GetTrustPolicyExceptionRules(appKey *edgeproto.AppKey) []*edgeproto.SecurityRule {
	var rules []*edgeproto.SecurityRule
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		pol := data.Obj
		if pol.Key.AppKey.Organization == appKey.Organization && pol.Key.AppKey.Name == appKey.Name && pol.Key.AppKey.Version == appKey.Version {
			for _, r := range pol.OutboundSecurityRules {
				rules = append(rules, &r)
			}
		}
	}
	return rules
}
