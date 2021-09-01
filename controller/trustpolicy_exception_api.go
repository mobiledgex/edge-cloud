package main

import (
	"fmt"

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

func InitTrustPolicyExceptionApi(sync *Sync) {
	trustPolicyExceptionApi.sync = sync
	trustPolicyExceptionApi.store = edgeproto.NewTrustPolicyExceptionStore(sync.store)
	edgeproto.InitTrustPolicyExceptionCache(&trustPolicyExceptionApi.cache)
	sync.RegisterCache(&trustPolicyExceptionApi.cache)
}

func (s *TrustPolicyExceptionApi) CreateTrustPolicyException(in *edgeproto.TrustPolicyException, cb edgeproto.TrustPolicyExceptionApi_CreateTrustPolicyExceptionServer) error {
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

	cur := edgeproto.TrustPolicyException{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.ExistsError()
		}
		return nil
	})

	_, err = s.store.Create(ctx, in, s.sync.syncWait)
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
		if cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED || cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE {
			return fmt.Errorf("Not allowed to modify TrustPolicyException in state:%s", cur.State.String())
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
	return nil
}

func (s *TrustPolicyExceptionApi) RequestTrustPolicyException(in *edgeproto.TrustPolicyException, cb edgeproto.TrustPolicyExceptionApi_RequestTrustPolicyExceptionServer) error {
	ctx := cb.Context()
	cur := edgeproto.TrustPolicyException{}
	changed := 0

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		if cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE {
			return nil
		}
		if cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED {
			// Just a hack for now. FIXME FIXME
			cur.State = edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE
		} else {
			cur.State = edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED
		}

		log.SpanLog(ctx, log.DebugLevelApi, "Setting TrustPolicyExceptionState", "state:", cur.State)
		changed = 1
		//changed = cur.CopyInFields(in) // FIXME FIXME
		if err := cur.Validate(nil); err != nil {
			return err
		}

		if changed == 0 {
			return nil
		}
		s.store.STMPut(stm, &cur)

		if cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE {
			// If App is already deployed and TrustPolicyException is created later, we should automatically program the TrustPolicyException rules

		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *TrustPolicyExceptionApi) DeleteTrustPolicyException(in *edgeproto.TrustPolicyException, cb edgeproto.TrustPolicyExceptionApi_DeleteTrustPolicyExceptionServer) error {
	ctx := cb.Context()
	if !s.cache.HasKey(&in.Key) {
		return in.Key.NotFoundError()
	}

	cur := edgeproto.TrustPolicyException{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		if cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED || cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE {
			return fmt.Errorf("Not allowed to delete TrustPolicyException in state:%s", cur.State.String())
		}
		return nil
	})
	if err != nil {
		return err
	}

	_, err = s.store.Delete(ctx, in, s.sync.syncWait)
	return err
}
func (s *TrustPolicyExceptionApi) ShowTrustPolicyException(in *edgeproto.TrustPolicyException, cb edgeproto.TrustPolicyExceptionApi_ShowTrustPolicyExceptionServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.TrustPolicyException) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *TrustPolicyExceptionApi) STMFind(stm concurrency.STM, appName, appOrg, appVer, cloudletName, cloudletOrg string, policy *edgeproto.TrustPolicyException) error {
	key := edgeproto.TrustPolicyExceptionKey{}
	key.AppKey.Organization = appOrg
	key.AppKey.Name = appName
	key.AppKey.Version = appVer
	key.CloudletKey.Organization = cloudletOrg
	key.CloudletKey.Name = cloudletName

	if !s.store.STMGet(stm, &key, policy) {
		return fmt.Errorf("TrustPolicyException for app %s version %s organization %s not found", appName, appVer, appOrg)
	}
	return nil
}

// Pass cloudletKey
func (s *TrustPolicyExceptionApi) GetTrustPolicyExceptionRules(ckey *edgeproto.CloudletKey, appKey *edgeproto.AppKey) []*edgeproto.SecurityRule {
	var rules []*edgeproto.SecurityRule
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		pol := data.Obj
		if ckey.Organization != pol.Key.CloudletKey.Organization || ckey.Name != pol.Key.CloudletKey.Name {
			continue
		}

		if pol.Key.AppKey.Organization == appKey.Organization && pol.Key.AppKey.Name == appKey.Name && pol.Key.AppKey.Version == appKey.Version {
			for _, r := range pol.OutboundSecurityRules {
				rules = append(rules, &r)
			}
		}
	}
	return rules
}

type TrustPolicyExceptionResponseApi struct {
	sync  *Sync
	store edgeproto.TrustPolicyExceptionResponseStore
	cache edgeproto.TrustPolicyExceptionResponseCache
}

var trustPolicyExceptionResponseApi = TrustPolicyExceptionResponseApi{}

func InitTrustPolicyExceptionResponseApi(sync *Sync) {
	trustPolicyExceptionResponseApi.sync = sync
	trustPolicyExceptionResponseApi.store = edgeproto.NewTrustPolicyExceptionResponseStore(sync.store)
	edgeproto.InitTrustPolicyExceptionResponseCache(&trustPolicyExceptionResponseApi.cache)
	sync.RegisterCache(&trustPolicyExceptionResponseApi.cache)
}

func (s *TrustPolicyExceptionResponseApi) CreateTrustPolicyExceptionResponse(in *edgeproto.TrustPolicyExceptionResponse, cb edgeproto.TrustPolicyExceptionResponseApi_CreateTrustPolicyExceptionResponseServer) error {
	return nil
}

func (s *TrustPolicyExceptionResponseApi) UpdateTrustPolicyExceptionResponse(in *edgeproto.TrustPolicyExceptionResponse, cb edgeproto.TrustPolicyExceptionResponseApi_UpdateTrustPolicyExceptionResponseServer) error {
	return nil
}

func (s *TrustPolicyExceptionResponseApi) DeleteTrustPolicyExceptionResponse(in *edgeproto.TrustPolicyExceptionResponse, cb edgeproto.TrustPolicyExceptionResponseApi_DeleteTrustPolicyExceptionResponseServer) error {
	return nil
}

func (s *TrustPolicyExceptionResponseApi) ShowTrustPolicyExceptionResponse(in *edgeproto.TrustPolicyExceptionResponse, cb edgeproto.TrustPolicyExceptionResponseApi_ShowTrustPolicyExceptionResponseServer) error {
	return nil
}
