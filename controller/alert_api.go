package main

import (
	"context"
	"strconv"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type AlertApi struct {
	sync        *Sync
	store       edgeproto.AlertStore
	cache       edgeproto.AlertCache
	sourceCache edgeproto.AlertCache // source of truth from crm/etc
}

var alertApi = AlertApi{}

func InitAlertApi(sync *Sync) {
	alertApi.sync = sync
	alertApi.store = edgeproto.NewAlertStore(sync.store)
	edgeproto.InitAlertCache(&alertApi.cache)
	edgeproto.InitAlertCache(&alertApi.sourceCache)
	alertApi.sourceCache.SetUpdatedCb(alertApi.StoreUpdate)
	alertApi.sourceCache.SetDeletedCb(alertApi.StoreDelete)
	sync.RegisterCache(&alertApi.cache)
}

// AppInstDown alert needs to set the HealthCheck in AppInst
func appInstSetStateFromHealthCheckAlert(ctx context.Context, alert *edgeproto.Alert, state dme.HealthCheck) {
	appOrg, ok := alert.Labels[edgeproto.AppKeyTagOrganization]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find AppInst Org label in Alert", "alert", alert)
		return
	}
	clorg, ok := alert.Labels[edgeproto.CloudletKeyTagOrganization]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find Cloudlet Org label in Alert", "alert", alert)
		return
	}
	cloudlet, ok := alert.Labels[edgeproto.CloudletKeyTagName]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find Cloudlet label in Alert", "alert", alert)
		return
	}
	cluster, ok := alert.Labels[edgeproto.ClusterKeyTagName]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find Cluster label in Alert", "alert", alert)
		return
	}
	clusterOrg, ok := alert.Labels[edgeproto.ClusterInstKeyTagOrganization]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find Cluster Org label in Alert", "alert", alert)
		return
	}
	appName, ok := alert.Labels[edgeproto.AppKeyTagName]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find App Name label in Alert", "alert", alert)
		return
	}
	appVer, ok := alert.Labels[edgeproto.AppKeyTagVersion]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "Could not find App Version label in Alert", "alert", alert)
		return
	}
	appInst := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				Organization: appOrg,
				Name:         appName,
				Version:      appVer,
			},
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
				ClusterKey: edgeproto.ClusterKey{
					Name: cluster,
				},
				CloudletKey: edgeproto.CloudletKey{
					Organization: clorg,
					Name:         cloudlet,
				},
				Organization: clusterOrg,
			},
		},
	}
	appInstApi.HealthCheckUpdate(ctx, &appInst, state)

}

func (s *AlertApi) setAlertMetadata(in *edgeproto.Alert) {
	in.Controller = ControllerId
	// Add a region label
	in.Labels["region"] = *region
}

func (s *AlertApi) Update(ctx context.Context, in *edgeproto.Alert, rev int64) {
	// for now, only store needed alerts
	name, ok := in.Labels["alertname"]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelNotify, "alertname not found", "labels", in.Labels)
		return
	}
	if !cloudcommon.IsMonitoredAlert(in.Labels) {
		log.SpanLog(ctx, log.DebugLevelNotify, "ignoring alert", "name", name)
		return
	}
	s.setAlertMetadata(in)
	// The CRM is the source of truth for Alerts.
	// We keep a local copy (sourceCache) of all alerts sent by the CRM.
	// If we lose the keep-alive lease with etcd and it deletes all these
	// alerts, we can push them back again from the source cache once the
	// keep-alive lease is reestablished.
	// All alert changes must pass through the source cache before going
	// to etcd.
	s.sourceCache.Update(ctx, in, rev)
	// Note that any further actions should done as part of StoreUpdate.
	// This is because if the keep-alive is lost and we resync, then
	// these additional actions should be performed again as part of StoreUpdate.
}

func (s *AlertApi) StoreUpdate(ctx context.Context, old, new *edgeproto.Alert) {
	s.store.Put(ctx, new, nil, objstore.WithLease(ControllerAliveLease()))
	name, ok := new.Labels["alertname"]
	if !ok {
		return
	}
	if name == cloudcommon.AlertAppInstDown {
		state, ok := new.Labels[cloudcommon.AlertHealthCheckStatus]
		if !ok {
			log.SpanLog(ctx, log.DebugLevelNotify, "HealthCheck status not found",
				"labels", new.Labels)
			return
		}
		hcState, err := strconv.Atoi(state)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelNotify, "failed to parse Health Check state",
				"state", state, "error", err)
			return
		}
		appInstSetStateFromHealthCheckAlert(ctx, new, dme.HealthCheck(hcState))
	}
}

func (s *AlertApi) Delete(ctx context.Context, in *edgeproto.Alert, rev int64) {
	// Add a region label
	in.Labels["region"] = *region
	s.sourceCache.DeleteCondFunc(ctx, in, rev, func(old *edgeproto.Alert) bool {
		if old.NotifyId != in.NotifyId {
			// already updated by another thread, don't delete
			return false
		}
		return true
	})
	// Note that any further actions should done as part of StoreDelete.
}

func (s *AlertApi) StoreDelete(ctx context.Context, in *edgeproto.Alert) {
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
	// Reset HealthCheck state back to OK
	name, ok := in.Labels["alertname"]
	if ok && foundAlert && name == cloudcommon.AlertAppInstDown {
		appInstSetStateFromHealthCheckAlert(ctx, in, dme.HealthCheck_HEALTH_CHECK_OK)
	}
}

func (s *AlertApi) Flush(ctx context.Context, notifyId int64) {
	// Delete all data from sourceCache. This will trigger StoreDelete calls
	// for every item.
	s.sourceCache.Flush(ctx, notifyId)
}

func (s *AlertApi) Prune(ctx context.Context, keys map[edgeproto.AlertKey]struct{}) {}

func (s *AlertApi) ShowAlert(in *edgeproto.Alert, cb edgeproto.AlertApi_ShowAlertServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Alert) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AlertApi) CleanupCloudletAlerts(ctx context.Context, key *edgeproto.CloudletKey) {
	matches := []*edgeproto.Alert{}
	s.cache.Mux.Lock()
	for _, data := range s.cache.Objs {
		val := data.Obj
		if cloudletName, found := val.Labels[edgeproto.CloudletKeyTagName]; !found ||
			cloudletName != key.Name {
			continue
		}
		if cloudletOrg, found := val.Labels[edgeproto.CloudletKeyTagOrganization]; !found ||
			cloudletOrg != key.Organization {
			continue
		}
		matches = append(matches, val)
	}
	s.cache.Mux.Unlock()
	for _, val := range matches {
		s.sourceCache.Delete(ctx, val, 0)
		s.store.Delete(ctx, val, s.sync.syncWait)
	}
}

func (s *AlertApi) syncSourceData(ctx context.Context, lease int64) error {
	// Note that we don't need to delete "stale" data, because
	// if the lease expired, it will be deleted automatically.
	alerts := make([]*edgeproto.Alert, 0)
	s.sourceCache.Mux.Lock()
	for _, data := range s.sourceCache.Objs {
		alert := edgeproto.Alert{}
		alert.DeepCopyIn(data.Obj)
		alerts = append(alerts, &alert)
	}
	s.sourceCache.Mux.Unlock()

	for _, alert := range alerts {
		s.StoreUpdate(ctx, nil, alert)
	}
	return nil
}
