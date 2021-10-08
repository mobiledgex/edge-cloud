package main

import (
	"context"
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

func (s *TrustPolicyExceptionApi) CreateTrustPolicyException(ctx context.Context, in *edgeproto.TrustPolicyException) (*edgeproto.Result, error) {

	log.SpanLog(ctx, log.DebugLevelApi, "CreateTrustPolicyException", "policy", in)

	if s.cache.HasKey(&in.Key) {
		return nil, in.Key.ExistsError()
	}

	// port range max is optional, set it to min if min is present but not max
	for i, o := range in.OutboundSecurityRules {
		if o.PortRangeMax == 0 {
			log.SpanLog(ctx, log.DebugLevelApi, "Setting PortRangeMax equal to min", "PortRangeMin", o.PortRangeMin)
			in.OutboundSecurityRules[i].PortRangeMax = o.PortRangeMin
		}
	}
	if err := in.Validate(nil); err != nil {
		return nil, err
	}

	if validateAppExists(&in.Key.AppKey) == false {
		return nil, fmt.Errorf("TrustPolicyExceptionKey: App does not exist")
	}

	if validateCloudletPoolExists(&in.Key.CloudletPoolKey) == false {
		return nil, fmt.Errorf("TrustPolicyExceptionKey: CloudletPoolKey does not exist")
	}

	in.State = edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED
	log.SpanLog(ctx, log.DebugLevelApi, "Setting TrustPolicyExceptionState", "state:", in.State)

	_, err := s.store.Create(ctx, in, s.sync.syncWait)
	return nil, err

}

func (s *TrustPolicyExceptionApi) UpdateTrustPolicyException(ctx context.Context, in *edgeproto.TrustPolicyException) (*edgeproto.Result, error) {
	return nil, fmt.Errorf("Not allowed to update TrustPolicyException")
}

func (s *TrustPolicyExceptionApi) DeleteTrustPolicyException(ctx context.Context, in *edgeproto.TrustPolicyException) (*edgeproto.Result, error) {
	if !s.cache.HasKey(&in.Key) {
		return nil, in.Key.NotFoundError()
	}
	_, err := s.store.Delete(ctx, in, s.sync.syncWait)
	return nil, err
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

	filterKey := edgeproto.TrustPolicyExceptionKey{
		CloudletPoolKey: *ckey,
		AppKey:          *appKey,
	}

	filter := edgeproto.TrustPolicyException{
		Key: filterKey,
	}

	s.cache.Show(&filter, func(pol *edgeproto.TrustPolicyException) error {
		for _, r := range pol.OutboundSecurityRules {
			rule := edgeproto.SecurityRule{}
			rule.DeepCopyIn(&r)
			rules = append(rules, &rule)
		}
		return nil
	})

	return rules
}

func (s *TrustPolicyExceptionApi) GetTrustPolicyExceptionForCloudletPoolKey(cKey *edgeproto.CloudletPoolKey) *edgeproto.TrustPolicyException {

	var TrustPolicyException *edgeproto.TrustPolicyException

	filterKey := edgeproto.TrustPolicyExceptionKey{
		CloudletPoolKey: *cKey,
	}

	filter := edgeproto.TrustPolicyException{
		Key: filterKey,
	}

	s.cache.Show(&filter, func(tpe *edgeproto.TrustPolicyException) error {
		TrustPolicyException = tpe
		return nil
	})

	return TrustPolicyException
}

func TrustPolicyExceptionForCloudletPoolKeyExists(cKey *edgeproto.CloudletPoolKey) bool {
	return trustPolicyExceptionApi.GetTrustPolicyExceptionForCloudletPoolKey(cKey) != nil
}

func (s *TrustPolicyExceptionApi) GetTrustPolicyExceptionForAppKey(appKey *edgeproto.AppKey) *edgeproto.TrustPolicyException {

	var TrustPolicyException *edgeproto.TrustPolicyException

	filterKey := edgeproto.TrustPolicyExceptionKey{
		AppKey: *appKey,
	}

	filter := edgeproto.TrustPolicyException{
		Key: filterKey,
	}

	s.cache.Show(&filter, func(tpe *edgeproto.TrustPolicyException) error {
		TrustPolicyException = tpe
		return nil
	})

	return TrustPolicyException
}

func TrustPolicyExceptionForAppKeyExists(appKey *edgeproto.AppKey) bool {
	return trustPolicyExceptionApi.GetTrustPolicyExceptionForAppKey(appKey) != nil
}

type TrustPolicyExceptionResponseApi struct {
}

var trustPolicyExceptionResponseApi = TrustPolicyExceptionResponseApi{}

func (s *TrustPolicyExceptionResponseApi) CreateTrustPolicyExceptionResponse(ctx context.Context, in *edgeproto.TrustPolicyExceptionResponse) (*edgeproto.Result, error) {
	cur := edgeproto.TrustPolicyException{}

	err := trustPolicyExceptionApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !trustPolicyExceptionApi.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		if in.State == cur.State {
			return fmt.Errorf("Current state is already %s", in.State.String())
		}
		if in.State != edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE &&
			in.State != edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_REJECTED {
			return fmt.Errorf("New state must be either Active or Rejected")
		}
		if cur.State != edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED {
			return fmt.Errorf("Not allowed to change TrustPolicyExceptionResponse in state:%s", cur.State.String())
		}
		cur.State = in.State
		log.SpanLog(ctx, log.DebugLevelApi, "Setting TrustPolicyExceptionResponseState", "state:", cur.State)
		trustPolicyExceptionApi.store.STMPut(stm, &cur)
		return nil
	})

	if err != nil {
		return nil, err
	}

	if in.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE {
		// If App is already deployed and TrustPolicyException is created later and approved now, we should automatically program the TrustPolicyException rules
		// FIXME TO DO
	}
	return nil, nil
}

func (s *TrustPolicyExceptionResponseApi) UpdateTrustPolicyExceptionResponse(ctx context.Context, in *edgeproto.TrustPolicyExceptionResponse) (*edgeproto.Result, error) {
	return nil, nil
}

func (s *TrustPolicyExceptionResponseApi) DeleteTrustPolicyExceptionResponse(ctx context.Context, in *edgeproto.TrustPolicyExceptionResponse) (*edgeproto.Result, error) {
	return nil, nil
}

func (s *TrustPolicyExceptionResponseApi) ShowTrustPolicyExceptionResponse(in *edgeproto.TrustPolicyExceptionResponse, cb edgeproto.TrustPolicyExceptionResponseApi_ShowTrustPolicyExceptionResponseServer) error {
	return nil
}
