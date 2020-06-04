package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

type AutoProvPolicyApi struct {
	sync    *Sync
	store   edgeproto.AutoProvPolicyStore
	cache   edgeproto.AutoProvPolicyCache
	influxQ *influxq.InfluxQ
}

var autoProvPolicyApi = AutoProvPolicyApi{}

func InitAutoProvPolicyApi(sync *Sync) {
	autoProvPolicyApi.sync = sync
	autoProvPolicyApi.store = edgeproto.NewAutoProvPolicyStore(sync.store)
	edgeproto.InitAutoProvPolicyCache(&autoProvPolicyApi.cache)
	sync.RegisterCache(&autoProvPolicyApi.cache)
}

func (s *AutoProvPolicyApi) SetInfluxQ(influxQ *influxq.InfluxQ) {
	s.influxQ = influxQ
}

func (s *AutoProvPolicyApi) CreateAutoProvPolicy(ctx context.Context, in *edgeproto.AutoProvPolicy) (*edgeproto.Result, error) {
	if err := in.Validate(nil); err != nil {
		return &edgeproto.Result{}, err
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
	cur := edgeproto.AutoProvPolicy{}
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
		if err := s.configureCloudlets(stm, &cur); err != nil {
			return err
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AutoProvPolicyApi) DeleteAutoProvPolicy(ctx context.Context, in *edgeproto.AutoProvPolicy) (*edgeproto.Result, error) {
	if appApi.UsesAutoProvPolicy(&in.Key) {
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
	cur := edgeproto.AutoProvPolicy{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		provCloudlet := edgeproto.AutoProvCloudlet{}
		provCloudlet.Key = in.CloudletKey
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
	cur := edgeproto.AutoProvPolicy{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
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
		go func() {
			span := log.StartSpan(log.DebugLevelMetrics, "AutoProvCreateAppInst", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
			ctx := log.ContextWithSpan(context.Background(), span)
			stream := streamoutAppInst{
				ctx:      ctx,
				debugLvl: log.DebugLevelMetrics,
			}
			appInst := edgeproto.AppInst{}
			appInst.Key.AppKey = target.AppKey
			appInst.Key.ClusterInstKey = target.DeployNowKey
			appInst.Liveness = edgeproto.Liveness_LIVENESS_DYNAMIC
			err := appInstApi.createAppInstInternal(DefCallContext(), &appInst, &stream)
			log.SpanLog(ctx, log.DebugLevelMetrics, "auto prov now", "appInst", appInst.Key, "err", err)
			span.Finish()
		}()
		return
	}
	// push stats to influxdb
	log.SpanLog(ctx, log.DebugLevelMetrics, "push auto-prov counts to influxdb", "msg", msg)
	err := s.influxQ.PushAutoProvCounts(ctx, msg)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMetrics, "failed to push auto-prov counts to influxdb", "err", err)
	}
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
		if !cloudletApi.store.STMGet(stm, &policy.Cloudlets[ii].Key, &cloudlet) {
			return policy.Cloudlets[ii].Key.NotFoundError()
		}
		policy.Cloudlets[ii].Loc = cloudlet.Location
	}
	return nil
}
