package main

import (
	"context"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type AlertApi struct {
	sync  *Sync
	store edgeproto.AlertStore
	cache edgeproto.AlertCache
}

var alertApi = AlertApi{}

func InitAlertApi(sync *Sync) {
	alertApi.sync = sync
	alertApi.store = edgeproto.NewAlertStore(sync.store)
	edgeproto.InitAlertCache(&alertApi.cache)
	sync.RegisterCache(&alertApi.cache)
}

// AppInstDown alert needs to set the TrackedState of the AppInst
func appInstSetStateFromAlert(ctx context.Context, alert *edgeproto.Alert, state edgeproto.TrackedState) {
	dev, ok := alert.Labels[cloudcommon.AlertLabelDev]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find Dev label in Alert", "alert", alert)
		return
	}
	oper, ok := alert.Labels[cloudcommon.AlertLabelOperator]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find Operator label in Alert", "alert", alert)
		return
	}
	cloudlet, ok := alert.Labels[cloudcommon.AlertLabelCloudlet]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find Cloudlet label in Alert", "alert", alert)
		return
	}
	cluster, ok := alert.Labels[cloudcommon.AlertLabelCluster]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find Cluster label in Alert", "alert", alert)
		return
	}
	appName, ok := alert.Labels[cloudcommon.AlertLabelApp]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find App Name label in Alert", "alert", alert)
		return
	}
	appVer, ok := alert.Labels[cloudcommon.AlertLabelAppVer]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find App Version label in Alert", "alert", alert)
		return
	}
	// Build up info since we use UpdateFromInfo API
	appInst := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				DeveloperKey: edgeproto.DeveloperKey{
					Name: dev,
				},
				Name:    appName,
				Version: appVer,
			},
			ClusterInstKey: edgeproto.ClusterInstKey{
				ClusterKey: edgeproto.ClusterKey{
					Name: cluster,
				},
				CloudletKey: edgeproto.CloudletKey{
					OperatorKey: edgeproto.OperatorKey{
						Name: oper,
					},
					Name: cloudlet,
				},
				Developer: dev,
			},
		},
		State: state,
	}
	appInstApi.UpdateState(ctx, &appInst)

}

func (s *AlertApi) Update(ctx context.Context, in *edgeproto.Alert, rev int64) {
	// for now, only store needed alerts
	name, ok := in.Labels["alertname"]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "alertname not found", "labels", in.Labels)
		return
	}
	if name != cloudcommon.AlertAutoScaleUp && name != cloudcommon.AlertAutoScaleDown &&
		name != cloudcommon.AlertAppInstDown {
		log.SpanLog(ctx, log.DebugLevelNotify, "ignoring alert", "name", name)
		return
	}
	if name == cloudcommon.AlertAppInstDown {
		appInstSetStateFromAlert(ctx, in, edgeproto.TrackedState_HEALTHCHECK_FAILED)
	}

	in.Controller = ControllerId
	s.store.Put(ctx, in, nil, objstore.WithLease(controllerAliveLease))
}

func (s *AlertApi) Delete(ctx context.Context, in *edgeproto.Alert, rev int64) {
	buf := edgeproto.Alert{}
	var foundAlert bool
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, in.GetKey(), &buf) {
			return nil
		}
		if buf.NotifyId != in.NotifyId || buf.Controller != ControllerId {
			// updated by another thread or controller
			return nil
		}
		s.store.STMDel(stm, in.GetKey())
		foundAlert = true
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "notify delete Alert", "key", in.GetKeyVal(), "err", err)
	}
	// Need to set the state of AppInst to disabled
	name, ok := in.Labels["alertname"]
	if ok && foundAlert && name == cloudcommon.AlertAppInstDown {
		appInstSetStateFromAlert(ctx, in, edgeproto.TrackedState_READY)
	}
}

func (s *AlertApi) Flush(ctx context.Context, notifyId int64) {
	matches := make([]edgeproto.AlertKey, 0)
	s.cache.Mux.Lock()
	for _, val := range s.cache.Objs {
		if val.NotifyId != notifyId || val.Controller != ControllerId {
			continue
		}
		matches = append(matches, val.GetKeyVal())
	}
	s.cache.Mux.Unlock()

	info := edgeproto.Alert{}
	for _, key := range matches {
		err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			if s.store.STMGet(stm, &key, &info) {
				if info.NotifyId != notifyId || info.Controller != ControllerId {
					// updated by another thread or controller
					return nil
				}
			}
			s.store.STMDel(stm, &key)
			return nil
		})
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelNotify, "flush alert", "key", key, "err", err)
		}
	}
}

func (s *AlertApi) Prune(ctx context.Context, keys map[edgeproto.AlertKey]struct{}) {}

func (s *AlertApi) ShowAlert(in *edgeproto.Alert, cb edgeproto.AlertApi_ShowAlertServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Alert) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
