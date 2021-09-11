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

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		if cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE || cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED {
			return fmt.Errorf("Already in state:%s", cur.State.String())
		}
		cur.State = edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED
		log.SpanLog(ctx, log.DebugLevelApi, "Setting TrustPolicyExceptionState", "state:", cur.State)
		if err := cur.Validate(nil); err != nil {
			return err
		}

		s.store.STMPut(stm, &cur)

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

func (s *TrustPolicyExceptionApi) GetTrustPolicyExceptionRules(ckey *edgeproto.CloudletPoolKey, appKey *edgeproto.AppKey) []*edgeproto.SecurityRule {
	var rules []*edgeproto.SecurityRule
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		pol := data.Obj
		if ckey.Organization != pol.Key.CloudletPoolKey.Organization || ckey.Name != pol.Key.CloudletPoolKey.Name {
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
}

var trustPolicyExceptionResponseApi = TrustPolicyExceptionResponseApi{}

func InitTrustPolicyExceptionResponseApi(sync *Sync) {
}

func (s *TrustPolicyExceptionResponseApi) CreateTrustPolicyExceptionResponse(in *edgeproto.TrustPolicyExceptionResponse, cb edgeproto.TrustPolicyExceptionResponseApi_CreateTrustPolicyExceptionResponseServer) error {
	ctx := cb.Context()
	cur := edgeproto.TrustPolicyException{}

	activated := 0

	err := trustPolicyExceptionApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !trustPolicyExceptionApi.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		if in.State != edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE &&
			in.State != edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_REJECTED {
			return fmt.Errorf("Not allowed to change to new state:%s", in.State.String())
		}
		if cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED {
			cur.State = in.State
			log.SpanLog(ctx, log.DebugLevelApi, "Setting TrustPolicyExceptionResponseState", "state:", cur.State)
			trustPolicyExceptionApi.store.STMPut(stm, &cur)
			if in.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE {
				activated = 1
			}
		} else {
			return fmt.Errorf("Not allowed to change TrustPolicyExceptionResponse in state:%s", cur.State.String())
		}
		return nil
	})

	if err != nil {
		return err
	}

	if activated == 1 {
		// If App is already deployed and TrustPolicyException is created later and approved now, we should automatically program the TrustPolicyException rules
		// FIXME TO DO
	}
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
