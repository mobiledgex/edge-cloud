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

func (s *AlertApi) Update(ctx context.Context, in *edgeproto.Alert, rev int64) {
	// for now, only store needed alerts
	name, ok := in.Labels["alertname"]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "alertname not found", "labels", in.Labels)
		return
	}
	if name != cloudcommon.AlertAutoScaleUp && name != cloudcommon.AlertAutoScaleDown {
		log.SpanLog(ctx, log.DebugLevelNotify, "ignoring alert", "name", name)
		return
	}
	in.Controller = ControllerId
	s.store.Put(ctx, in, nil, objstore.WithLease(controllerAliveLease))
}

func (s *AlertApi) Delete(ctx context.Context, in *edgeproto.Alert, rev int64) {
	buf := edgeproto.Alert{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, in.GetKey(), &buf) {
			return nil
		}
		if buf.NotifyId != in.NotifyId || buf.Controller != ControllerId {
			// updated by another thread or controller
			return nil
		}
		s.store.STMDel(stm, in.GetKey())
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "notify delete Alert", "key", in.GetKeyVal(), "err", err)
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
