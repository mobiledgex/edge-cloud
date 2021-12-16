package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type TrustPolicyExceptionApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.TrustPolicyExceptionStore
	cache edgeproto.TrustPolicyExceptionCache
}

func NewTrustPolicyExceptionApi(sync *Sync, all *AllApis) *TrustPolicyExceptionApi {
	trustPolicyExceptionApi := TrustPolicyExceptionApi{}
	trustPolicyExceptionApi.all = all
	trustPolicyExceptionApi.sync = sync
	trustPolicyExceptionApi.store = edgeproto.NewTrustPolicyExceptionStore(sync.store)
	edgeproto.InitTrustPolicyExceptionCache(&trustPolicyExceptionApi.cache)
	sync.RegisterCache(&trustPolicyExceptionApi.cache)
	return &trustPolicyExceptionApi
}

func (s *TrustPolicyExceptionApi) fixupPortRangeMax(ctx context.Context, in *edgeproto.TrustPolicyException) {
	// port range max is optional, set it to min if min is present but not max
	for i, o := range in.OutboundSecurityRules {
		if o.PortRangeMax == 0 {
			log.SpanLog(ctx, log.DebugLevelApi, "Setting PortRangeMax equal to min", "PortRangeMin", o.PortRangeMin)
			in.OutboundSecurityRules[i].PortRangeMax = o.PortRangeMin
		}
	}
}

func (s *TrustPolicyExceptionApi) CreateTrustPolicyException(ctx context.Context, in *edgeproto.TrustPolicyException) (*edgeproto.Result, error) {

	log.SpanLog(ctx, log.DebugLevelApi, "CreateTrustPolicyException", "policy", in)
	s.fixupPortRangeMax(ctx, in)
	if err := in.Validate(nil); err != nil {
		return nil, err
	}
	in.State = edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED
	log.SpanLog(ctx, log.DebugLevelApi, "Setting TrustPolicyExceptionState", "state:", in.State)

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		app := edgeproto.App{}
		if !s.all.appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return in.Key.AppKey.NotFoundError()
		}
		if app.DeletePrepare {
			return in.Key.AppKey.BeingDeletedError()
		}
		cloudletPool := edgeproto.CloudletPool{}
		if !s.all.cloudletPoolApi.store.STMGet(stm, &in.Key.CloudletPoolKey, &cloudletPool) {
			return in.Key.CloudletPoolKey.NotFoundError()
		}
		if cloudletPool.DeletePrepare {
			return in.Key.CloudletPoolKey.BeingDeletedError()
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

// A developer can call this api to update OutboundSecurityRules (protocol, port range min/max, cidr).
// Such an update is allowed only when it's in approval requested state.
// An operator can call this api to update State to Active or Rejected
// Authz takes care of such 'incoming' State checks as per role, but it does not have current state.
// This api does not have information on whether this is called by operator or developer.
func (s *TrustPolicyExceptionApi) UpdateTrustPolicyException(ctx context.Context, in *edgeproto.TrustPolicyException) (*edgeproto.Result, error) {
	cur := edgeproto.TrustPolicyException{}

	fields := edgeproto.MakeFieldMap(in.Fields)

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		// Safeguard from incoming state as UNKNOWN
		if in.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_UNKNOWN {
			return fmt.Errorf("User not allowed to update TrustPolicyException state to %s", in.State.String())
		}

		log.SpanLog(ctx, log.DebugLevelApi, "UpdateTrustPolicyException", "in state:", in.State.String(), "cur state:", cur.State.String())

		// Disallow going to APPROVAL_REQUESTED from Active or Rejected state
		if (in.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED) &&
			(cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE ||
				cur.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_REJECTED) {
			log.SpanLog(ctx, log.DebugLevelApi, "UpdateTrustPolicyException not allowed", "state:", cur.State.String())
			return fmt.Errorf("New state must be either Active or Rejected, current state: %s", cur.State.String())
		}
		// Disallow OutboundSecurityRules changes if new state is going to be Active or Rejected
		if in.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE ||
			in.State == edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_REJECTED {
			if _, found := fields[edgeproto.TrustPolicyExceptionFieldOutboundSecurityRulesProtocol]; found {
				return fmt.Errorf("field %s not allowed in new state: %s",
					edgeproto.TrustPolicyExceptionAllFieldsStringMap[edgeproto.TrustPolicyExceptionFieldOutboundSecurityRulesProtocol],
					in.State.String())
			}
			if _, found := fields[edgeproto.TrustPolicyExceptionFieldOutboundSecurityRulesPortRangeMin]; found {
				return fmt.Errorf("field %s not allowed in new state: %s",
					edgeproto.TrustPolicyExceptionAllFieldsStringMap[edgeproto.TrustPolicyExceptionFieldOutboundSecurityRulesPortRangeMin],
					in.State.String())
			}
			if _, found := fields[edgeproto.TrustPolicyExceptionFieldOutboundSecurityRulesPortRangeMax]; found {
				return fmt.Errorf("field %s not allowed in new state: %s",
					edgeproto.TrustPolicyExceptionAllFieldsStringMap[edgeproto.TrustPolicyExceptionFieldOutboundSecurityRulesPortRangeMax],
					in.State.String())
			}
			if _, found := fields[edgeproto.TrustPolicyExceptionFieldOutboundSecurityRulesRemoteCidr]; found {
				return fmt.Errorf("field %s not allowed in new state: %s",
					edgeproto.TrustPolicyExceptionAllFieldsStringMap[edgeproto.TrustPolicyExceptionFieldOutboundSecurityRulesRemoteCidr],
					in.State.String())
			}
		}
		// Copy in user specified fields only
		changed := cur.CopyInFields(in)
		if changed == 0 {
			return nil // no changes
		}
		s.fixupPortRangeMax(ctx, &cur)
		if err := cur.Validate(nil); err != nil {
			return err
		}
		log.SpanLog(ctx, log.DebugLevelApi, "UpdateTrustPolicyException", "state:", cur.State.String())
		s.store.STMPut(stm, &cur)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &edgeproto.Result{}, nil
}

func (s *TrustPolicyExceptionApi) DeleteTrustPolicyException(ctx context.Context, in *edgeproto.TrustPolicyException) (*edgeproto.Result, error) {
	if !s.cache.HasKey(&in.Key) {
		return nil, in.Key.NotFoundError()
	}
	_, err := s.store.Delete(ctx, in, s.sync.syncWait)
	return &edgeproto.Result{}, err
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

	filter := edgeproto.TrustPolicyException{
		Key: edgeproto.TrustPolicyExceptionKey{
			CloudletPoolKey: *ckey,
			AppKey:          *appKey,
		},
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

	filter := edgeproto.TrustPolicyException{
		Key: edgeproto.TrustPolicyExceptionKey{
			CloudletPoolKey: *cKey,
		},
	}

	s.cache.Show(&filter, func(tpe *edgeproto.TrustPolicyException) error {
		TrustPolicyException = tpe
		return nil
	})

	return TrustPolicyException
}

func (s *TrustPolicyExceptionApi) TrustPolicyExceptionForCloudletPoolKeyExists(cKey *edgeproto.CloudletPoolKey) *edgeproto.TrustPolicyExceptionKey {
	tpe := s.GetTrustPolicyExceptionForCloudletPoolKey(cKey)
	if tpe != nil {
		return &tpe.Key
	}
	return nil
}

func (s *TrustPolicyExceptionApi) GetTrustPolicyExceptionForAppKey(appKey *edgeproto.AppKey) *edgeproto.TrustPolicyException {

	var TrustPolicyException *edgeproto.TrustPolicyException

	filter := edgeproto.TrustPolicyException{
		Key: edgeproto.TrustPolicyExceptionKey{
			AppKey: *appKey,
		},
	}

	s.cache.Show(&filter, func(tpe *edgeproto.TrustPolicyException) error {
		TrustPolicyException = tpe
		return nil
	})

	return TrustPolicyException
}

func (s *TrustPolicyExceptionApi) TrustPolicyExceptionForAppKeyExists(appKey *edgeproto.AppKey) *edgeproto.TrustPolicyExceptionKey {
	tp := s.GetTrustPolicyExceptionForAppKey(appKey)
	if tp != nil {
		return &tp.Key
	}
	return nil
}
