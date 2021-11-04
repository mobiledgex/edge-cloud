package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util/tasks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AutoProvPolicyApi struct {
	all              *AllApis
	sync             *Sync
	store            edgeproto.AutoProvPolicyStore
	cache            edgeproto.AutoProvPolicyCache
	influxQ          *influxq.InfluxQ
	deployImmWorkers tasks.KeyWorkers
}

func NewAutoProvPolicyApi(sync *Sync, all *AllApis) *AutoProvPolicyApi {
	autoProvPolicyApi := AutoProvPolicyApi{}
	autoProvPolicyApi.all = all
	autoProvPolicyApi.sync = sync
	autoProvPolicyApi.store = edgeproto.NewAutoProvPolicyStore(sync.store)
	edgeproto.InitAutoProvPolicyCache(&autoProvPolicyApi.cache)
	sync.RegisterCache(&autoProvPolicyApi.cache)
	autoProvPolicyApi.deployImmWorkers.Init("AutoProvDeployImmediate", autoProvPolicyApi.deployImmediate)
	return &autoProvPolicyApi
}

func (s *AutoProvPolicyApi) SetInfluxQ(influxQ *influxq.InfluxQ) {
	s.influxQ = influxQ
}

func (s *AutoProvPolicyApi) CreateAutoProvPolicy(ctx context.Context, in *edgeproto.AutoProvPolicy) (*edgeproto.Result, error) {
	if err := in.Validate(nil); err != nil {
		return &edgeproto.Result{}, err
	}
	if int(in.MinActiveInstances) > len(in.Cloudlets) {
		return &edgeproto.Result{}, fmt.Errorf("Minimum Active Instances cannot be larger than the number of Cloudlets")
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		if err := s.configureCloudlets(stm, in); err != nil {
			return err
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AutoProvPolicyApi) UpdateAutoProvPolicy(ctx context.Context, in *edgeproto.AutoProvPolicy) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.AutoProvPolicy{}
		changed := 0
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
		if err := s.configureCloudlets(stm, &cur); err != nil {
			return err
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AutoProvPolicyApi) DeleteAutoProvPolicy(ctx context.Context, in *edgeproto.AutoProvPolicy) (*edgeproto.Result, error) {
	if s.all.appApi.UsesAutoProvPolicy(&in.Key) {
		return &edgeproto.Result{}, fmt.Errorf("Policy in use by App")
	}
	return s.store.Delete(ctx, in, s.sync.syncWait)
}

func (s *AutoProvPolicyApi) ShowAutoProvPolicy(in *edgeproto.AutoProvPolicy, cb edgeproto.AutoProvPolicyApi_ShowAutoProvPolicyServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AutoProvPolicy) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AutoProvPolicyApi) AddAutoProvPolicyCloudlet(ctx context.Context, in *edgeproto.AutoProvPolicyCloudlet) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.AutoProvPolicy{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		provCloudlet := edgeproto.AutoProvCloudlet{}
		provCloudlet.Key = in.CloudletKey
		for ii, _ := range cur.Cloudlets {
			if cur.Cloudlets[ii].Key.Matches(&in.CloudletKey) {
				return fmt.Errorf("Cloudlet already part of policy")
			}
		}
		cur.Cloudlets = append(cur.Cloudlets, &provCloudlet)
		if err := s.configureCloudlets(stm, &cur); err != nil {
			return err
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AutoProvPolicyApi) RemoveAutoProvPolicyCloudlet(ctx context.Context, in *edgeproto.AutoProvPolicyCloudlet) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.AutoProvPolicy{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed := false
		for ii, cloudlet := range cur.Cloudlets {
			if cloudlet.Key.Matches(&in.CloudletKey) {
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

func (s *AutoProvPolicyApi) STMFind(stm concurrency.STM, name, dev string, policy *edgeproto.AutoProvPolicy) error {
	key := edgeproto.PolicyKey{}
	key.Name = name
	key.Organization = dev
	if !s.store.STMGet(stm, &key, policy) {
		return fmt.Errorf("AutoProvPolicy %s for developer %s not found", name, dev)
	}
	return nil
}

func (s *AutoProvPolicyApi) RecvAutoProvCounts(ctx context.Context, msg *edgeproto.AutoProvCounts) {
	if len(msg.Counts) == 1 && msg.Counts[0].ProcessNow {
		target := msg.Counts[0]
		log.SpanLog(ctx, log.DebugLevelMetrics, "auto-prov count recv immedate", "target", target)
		appInstKey := edgeproto.AppInstKey{
			AppKey:         target.AppKey,
			ClusterInstKey: *target.DeployNowKey.Virtual(""),
		}
		s.deployImmWorkers.NeedsWork(ctx, appInstKey)
		return
	}
	// push stats to influxdb
	log.SpanLog(ctx, log.DebugLevelMetrics, "push auto-prov counts to influxdb", "msg", msg)
	err := s.influxQ.PushAutoProvCounts(ctx, msg)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMetrics, "failed to push auto-prov counts to influxdb", "err", err)
	}
}

func (s *AutoProvPolicyApi) deployImmediate(ctx context.Context, k interface{}) {
	key, ok := k.(edgeproto.AppInstKey)
	if !ok {
		log.SpanLog(ctx, log.DebugLevelApi, "Unexpected failure, key not AppInstKey", "key", k)
		return
	}
	log.SetContextTags(ctx, key.GetTags())
	log.SpanLog(ctx, log.DebugLevelApi, "deploy immediate", "key", key)
	md := metadata.Pairs(
		cloudcommon.CallerAutoProv, "",
		cloudcommon.AutoProvReason,
		cloudcommon.AutoProvReasonDemand,
		cloudcommon.AutoProvPolicyName, "dme-process-now")
	ctx = metadata.NewIncomingContext(ctx, md)
	appInst := edgeproto.AppInst{}
	appInst.Key = key
	stream := streamoutAppInst{
		ctx:      ctx,
		debugLvl: log.DebugLevelApi,
	}
	err := s.all.appInstApi.createAppInstInternal(DefCallContext(), &appInst, &stream)
	log.SpanLog(ctx, log.DebugLevelApi, "auto prov now", "appInst", appInst.Key, "err", err)
}

type streamoutAppInst struct {
	grpc.ServerStream
	ctx      context.Context
	debugLvl uint64
}

func (s *streamoutAppInst) Send(res *edgeproto.Result) error {
	log.SpanLog(s.Context(), s.debugLvl, res.Message)
	return nil
}

func (s *streamoutAppInst) Context() context.Context {
	return s.ctx
}

func (s *AutoProvPolicyApi) configureCloudlets(stm concurrency.STM, policy *edgeproto.AutoProvPolicy) error {
	// make sure cloudlets exist and location is copied
	for ii, _ := range policy.Cloudlets {
		cloudlet := edgeproto.Cloudlet{}
		if !s.all.cloudletApi.store.STMGet(stm, &policy.Cloudlets[ii].Key, &cloudlet) {
			return policy.Cloudlets[ii].Key.NotFoundError()
		}
		policy.Cloudlets[ii].Loc = cloudlet.Location
	}
	return nil
}

// Checks any AppInst create/delete API calls originating from the
// AutoProv service. Because of the distributed nature of the data,
// the AutoProv service may act on stale data, and cause the
// number of AppInsts to be invalid or sub-optimal.
// One way to deal with this is to reject any API calls that are
// done on stale data by tracking the database revisions of all data
// used by the AutoProv service. However because we must track the
// count of AppInsts via the AppInstRefs object, this effectively
// limits the AutoProv service to only operating serially for each
// AppInst associated with an App (because making a change to an AppInst
// will update the refs, which will necessarily invalidate any further
// API calls for that App). This is highly sub-optimal when we
// should be able to create AppInsts on multiple Cloudlets (for the same
// App) in parallel.
// The second way we are using here is to validate the intent of the
// API call. Based on the grpc metadata passed along with the API call,
// we know what the intent of the AutoProv service was when making the
// call. If the current state does not match that intent, we reject the
// API call.

// create/delete AppInsts:
// The first is to meet auto-deploy and auto-undeploy criteria. In this
// case, we need to ensure that calls do not violate the min/max limits.
// The second is to meet the min/max requirements. In this case we need
// to ensure that we do not create more than the min, or delete below
// the max.
func (s *AutoProvPolicyApi) appInstCheck(ctx context.Context, stm concurrency.STM, action cloudcommon.Action, app *edgeproto.App, inst *edgeproto.AppInst) error {
	log.SpanLog(ctx, log.DebugLevelApi, "autoprov check", "inst", inst.Key)
	// Check caller from GRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.SpanLog(ctx, log.DebugLevelApi, "autoprov check no metadata")
		// not AutoProv service
		return nil
	}
	if _, found := md[cloudcommon.CallerAutoProv]; !found {
		log.SpanLog(ctx, log.DebugLevelApi, "autoprov check no caller")
		// not AutoProv service
		return nil
	}
	reasons, found := md[cloudcommon.AutoProvReason]
	if !found || len(reasons) == 0 {
		log.SpanLog(ctx, log.DebugLevelApi, "autoprov check but no reason", "metadata", md)
		return nil
	}
	reason := reasons[0]

	policyNames, found := md[cloudcommon.AutoProvPolicyName]
	if !found || len(policyNames) == 0 {
		log.SpanLog(ctx, log.DebugLevelApi, "autoprov check but no policy name", "metadata", md)
		return nil
	}
	policyName := policyNames[0]

	if action == cloudcommon.Create {
		// mark auto-provisioned AppInsts so we can know which ones they are.
		inst.Liveness = edgeproto.Liveness_LIVENESS_AUTOPROV
	}

	refs := edgeproto.AppInstRefs{}
	s.all.appInstRefsApi.store.STMGet(stm, &app.Key, &refs)

	if action == cloudcommon.Create {
		// Make sure that no AppInst already exists on the Cloudlet.
		// The goal of AutoProv is to provide Cloudlet redundancy and
		// best service by distributing instances across Cloudlets, so we
		// avoid auto-provisioning AppInsts when one already exists on the
		// Cloudlet. Additionally this avoids a race condition for the
		// immediate deployment from DME where multiple clients may trigger
		// multiple calls from the DME to deploy to the same cloudlet.
		// It does not matter if the existing AppInst was manually or
		// automatically created.
		for k, _ := range refs.Insts {
			instKey := edgeproto.AppInstKey{}
			edgeproto.AppInstKeyStringParse(k, &instKey)
			if inst.Key.ClusterInstKey.CloudletKey.Matches(&instKey.ClusterInstKey.CloudletKey) {
				return fmt.Errorf("already an AppInst on ClusterInst %v on the Cloudlet", inst.Key.ClusterInstKey.GetKeyString())
			}
		}
	}

	var err error
	switch reason {
	case cloudcommon.AutoProvReasonDemand:
		err = s.checkDemand(ctx, stm, action, app, inst, &refs)
	case cloudcommon.AutoProvReasonMinMax:
		err = s.checkMinMax(ctx, stm, action, app, inst, &refs, policyName)
	case cloudcommon.AutoProvReasonOrphaned:
		err = s.checkOrphaned(ctx, stm, action, app, inst)
	default:
		log.SpanLog(ctx, log.DebugLevelApi, "unsupported reason", "reason", reason)
		return nil
	}
	log.SpanLog(ctx, log.DebugLevelApi, "AppInst autoprov check result", "action", action, "AppInst", inst.Key, "reason", reason, "err", err)
	return err
}

// The intent is to satisfy client demand, so the only limitations
// are to not exceed the min or max constraints on any of the
// policies governing the target AppInst's Cloudlet.
func (s *AutoProvPolicyApi) checkDemand(ctx context.Context, stm concurrency.STM, action cloudcommon.Action, app *edgeproto.App, inst *edgeproto.AppInst, refs *edgeproto.AppInstRefs) error {
	onlineOnly := false
	if action == cloudcommon.Delete {
		// If the AppInst we are deleting is not online, then it will
		// not affect the min active count so we can safely delete it.
		online, err := s.autoProvAppInstOnline(ctx, stm, &inst.Key)
		if err != nil {
			return err
		}
		if !online {
			return nil
		}
		onlineOnly = true
	}

	// For create, we are bounded by the MaxInstances value,
	// which is for any AppInst (regardless of online state).
	// For delete, we are bounded by the MinActiveInstances value,
	// which is for online AppInsts only.
	countsByCloudlet, err := s.getAppInstCountsForAutoProv(ctx, stm, refs, onlineOnly)
	if err != nil {
		return err
	}

	for pname, _ := range app.GetAutoProvPolicies() {
		policy := edgeproto.AutoProvPolicy{}
		policyKey := edgeproto.PolicyKey{
			Name:         pname,
			Organization: app.Key.Organization,
		}
		if !s.store.STMGet(stm, &policyKey, &policy) {
			return policyKey.NotFoundError()
		}
		withinPolicy := false
		count := 0
		for _, apCloudlet := range policy.Cloudlets {
			if apCloudlet.Key.Matches(&inst.Key.ClusterInstKey.CloudletKey) {
				withinPolicy = true
			}
			count += countsByCloudlet[apCloudlet.Key]
		}
		log.SpanLog(ctx, log.DebugLevelApi, "autoprov check demand stats", "action", action.String(), "inst", inst.Key, "policy", pname, "count", count, "withinPolicy", withinPolicy, "min", policy.MinActiveInstances, "max", policy.MaxInstances)
		if !withinPolicy {
			continue
		}
		if action == cloudcommon.Create && policy.MaxInstances > 0 {
			// Do not exceed max instances (0 is no limit)
			if count >= int(policy.MaxInstances) {
				return fmt.Errorf("Create would exceed max instances %d of policy %s", policy.MaxInstances, policy.Key.Name)
			}
		}
		if action == cloudcommon.Delete && policy.MinActiveInstances > 0 {
			if count <= int(policy.MinActiveInstances) {
				return fmt.Errorf("Delete would violate min active instances %d of policy %s", policy.MinActiveInstances, policy.Key.Name)
			}
		}
	}
	return nil
}

// Intent is to satisfy min/max constraints. Reject any calls that would
// overshoot the min/max constraints.
func (s *AutoProvPolicyApi) checkMinMax(ctx context.Context, stm concurrency.STM, action cloudcommon.Action, app *edgeproto.App, inst *edgeproto.AppInst, refs *edgeproto.AppInstRefs, policyName string) error {
	// For create, we are trying to satisfy the min constraint,
	// which counts online AppInsts only.
	// For delete, we are trying to satisfy the max constraint,
	// which counts all AppInsts.
	onlineOnly := false
	if action == cloudcommon.Create {
		onlineOnly = true
	}
	countsByCloudlet, err := s.getAppInstCountsForAutoProv(ctx, stm, refs, onlineOnly)
	if err != nil {
		return err
	}
	policy := edgeproto.AutoProvPolicy{}
	policyKey := edgeproto.PolicyKey{
		Name:         policyName,
		Organization: app.Key.Organization,
	}
	if !s.store.STMGet(stm, &policyKey, &policy) {
		return policyKey.NotFoundError()
	}
	count := 0
	for _, apCloudlet := range policy.Cloudlets {
		count += countsByCloudlet[apCloudlet.Key]
	}
	log.SpanLog(ctx, log.DebugLevelApi, "autoprov check minmax stats", "action", action.String(), "inst", inst.Key, "policy", policyName, "count", count, "min", policy.MinActiveInstances, "max", policy.MaxInstances)
	if action == cloudcommon.Create && policy.MinActiveInstances > 0 {
		if count >= int(policy.MinActiveInstances) {
			return cloudcommon.AutoProvMinAlreadyMetError
		}
	}
	if action == cloudcommon.Delete && policy.MaxInstances > 0 {
		if count <= int(policy.MaxInstances) {
			return fmt.Errorf("Delete to satisfy max already met, ignoring")
		}
	}
	return nil
}

// Intent is to remove any auto-provisioned AppInsts that no longer
// belong to any policies on the App. This can happen if the user removes
// a policy from the App, for example.
func (s *AutoProvPolicyApi) checkOrphaned(ctx context.Context, stm concurrency.STM, action cloudcommon.Action, app *edgeproto.App, inst *edgeproto.AppInst) error {
	if action != cloudcommon.Delete {
		return nil
	}
	// Orphaned instances should not be on any policies on the app
	for pname, _ := range app.GetAutoProvPolicies() {
		policy := edgeproto.AutoProvPolicy{}
		policyKey := edgeproto.PolicyKey{
			Name:         pname,
			Organization: app.Key.Organization,
		}
		if !s.store.STMGet(stm, &policyKey, &policy) {
			continue
		}
		for _, apCloudlet := range policy.Cloudlets {
			if apCloudlet.Key.Matches(&inst.Key.ClusterInstKey.CloudletKey) {
				return fmt.Errorf("AppInst %s on Cloudlet in policy %s on App, not orphaned", inst.Key.GetKeyString(), pname)
			}
		}
	}
	return nil
}

func (s *AutoProvPolicyApi) getAppInstCountsForAutoProv(ctx context.Context, stm concurrency.STM, refs *edgeproto.AppInstRefs, onlineOnly bool) (map[edgeproto.CloudletKey]int, error) {
	countsByCloudlet := make(map[edgeproto.CloudletKey]int)

	for k, _ := range refs.Insts {
		instKey := edgeproto.AppInstKey{}
		edgeproto.AppInstKeyStringParse(k, &instKey)
		inst := edgeproto.AppInst{}
		if !s.all.appInstApi.store.STMGet(stm, &instKey, &inst) {
			// no inst?
			continue
		}
		if cloudcommon.AppInstBeingDeleted(&inst) {
			// don't count it
			continue
		}
		if onlineOnly {
			online, err := s.autoProvAppInstOnline(ctx, stm, &instKey)
			if err != nil {
				return countsByCloudlet, err
			}
			if !online {
				continue
			}
		}
		countsByCloudlet[instKey.ClusterInstKey.CloudletKey]++
	}
	return countsByCloudlet, nil
}

func (s *AutoProvPolicyApi) autoProvAppInstOnline(ctx context.Context, stm concurrency.STM, key *edgeproto.AppInstKey) (bool, error) {
	appInst := edgeproto.AppInst{}
	if !s.all.appInstApi.store.STMGet(stm, key, &appInst) {
		return false, key.NotFoundError()
	}
	cloudletInfo := edgeproto.CloudletInfo{}
	if !s.all.cloudletInfoApi.store.STMGet(stm, &key.ClusterInstKey.CloudletKey, &cloudletInfo) {
		return false, key.ClusterInstKey.CloudletKey.NotFoundError()
	}
	cloudlet := edgeproto.Cloudlet{}
	if !s.all.cloudletApi.store.STMGet(stm, &key.ClusterInstKey.CloudletKey, &cloudlet) {
		return false, key.ClusterInstKey.CloudletKey.NotFoundError()
	}
	return cloudcommon.AutoProvAppInstOnline(&appInst, &cloudletInfo, &cloudlet), nil
}

func (s *AutoProvPolicyApi) UsesCloudlet(key *edgeproto.CloudletKey) []edgeproto.PolicyKey {
	policyKeys := []edgeproto.PolicyKey{}
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for policyKey, data := range s.cache.Objs {
		policy := data.Obj
		for _, autoProvCloudlet := range policy.Cloudlets {
			if autoProvCloudlet.Key.Matches(key) {
				policyKeys = append(policyKeys, policyKey)
				break
			}
		}
	}
	return policyKeys
}
