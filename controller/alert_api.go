package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type AlertApi struct {
	all                     *AllApis
	sync                    *Sync
	cache                   edgeproto.AlertCache
	sourceCache             edgeproto.AlertCache // source of truth from crm/etc
	syncMux                 sync.Mutex           // for unit-testing
	doneKeepaliveRefresh    bool
	triggerKeepaliveRefresh chan bool
}

var (
	ControllerCreatedAlerts = "ControllerCreatedAlerts"

	AlertTTL          = time.Duration(1 * time.Minute)
	RedisTxMaxRetries = 10
	RedisSyncTimeout  = time.Duration(10 * time.Second)
)

func NewAlertApi(sync *Sync, all *AllApis) *AlertApi {
	alertApi := AlertApi{}
	alertApi.all = all
	alertApi.sync = sync
	alertApi.triggerKeepaliveRefresh = make(chan bool, 1)
	edgeproto.InitAlertCache(&alertApi.cache)
	edgeproto.InitAlertCache(&alertApi.sourceCache)
	alertApi.sourceCache.SetUpdatedCb(alertApi.StoreUpdate)
	alertApi.sourceCache.SetDeletedCb(alertApi.StoreDelete)
	sync.RegisterCache(&alertApi.cache)
	return &alertApi
}

func getAlertStoreKey(in *edgeproto.Alert) string {
	return objstore.DbKeyString("Alert", in.GetKey())
}

func getAllAlertsKeyPattern() string {
	return objstore.DbKeyPrefixString("Alert") + "/*"
}

// AppInstDown alert needs to set the HealthCheck in AppInst
func (s *AlertApi) appInstSetStateFromHealthCheckAlert(ctx context.Context, alert *edgeproto.Alert, state dme.HealthCheck) {
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
	s.all.appInstApi.HealthCheckUpdate(ctx, &appInst, state)

}

func setAlertMetadata(in *edgeproto.Alert) {
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
	setAlertMetadata(in)
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

	// Wait for alerts to synced with controller cache
	recvdCache := make(chan bool, 1)
	watchCancel := s.cache.WatchKey(in.GetKey(), func(ctx context.Context) {
		recvdCache <- true
	})
	defer watchCancel()
	// check if object is already synced with cache
	if s.cache.Get(in.GetKey(), &edgeproto.Alert{}) {
		return
	}
	select {
	case <-recvdCache:
	case <-time.After(RedisSyncTimeout):
		log.SpanLog(ctx, log.DebugLevelNotify, "Timed out waiting to sync alerts from redis to cache", "key", in.GetKeyVal())
	}
}

func (s *AlertApi) StoreUpdate(ctx context.Context, old, new *edgeproto.Alert) {
	// Alerts are stored in redis cache, instead of etcd for the following reasons:
	// * They are transient in nature, this causes etcd fragmentation and hinders scalability
	// * It can be re-triggered/re-computed and hence it doesn't need a persistent storage
	key := getAlertStoreKey(new)
	val, err := json.Marshal(new)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "Failed to marshal alert object", "key", new.GetKeyVal(), "err", err)
		return
	}
	_, err = redisClient.Set(key, string(val), AlertTTL).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "Failed to store alert in rediscache", "key", new.GetKeyVal(), "err", err)
		return
	}

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
		s.appInstSetStateFromHealthCheckAlert(ctx, new, dme.HealthCheck(hcState))
	}
}

func (s *AlertApi) Delete(ctx context.Context, in *edgeproto.Alert, rev int64) {
	// Add a region label
	in.Labels["region"] = *region
	// Controller created alerts, so delete directly
	_, ok := ctx.Value(ControllerCreatedAlerts).(*string)
	if ok {
		s.sourceCache.Delete(ctx, in, rev)
		// Reset HealthCheck state back to OK
		name, ok := in.Labels["alertname"]
		if ok && name == cloudcommon.AlertAppInstDown {
			s.appInstSetStateFromHealthCheckAlert(ctx, in, dme.HealthCheck_HEALTH_CHECK_OK)
		}
	} else {
		s.sourceCache.DeleteCondFunc(ctx, in, rev, func(old *edgeproto.Alert) bool {
			if old.NotifyId != in.NotifyId {
				// already updated by another thread, don't delete
				return false
			}
			return true
		})
	}

	// Wait for alerts to synced with controller cache
	syncedCache := make(chan bool, 1)
	watchCancel := s.cache.WatchKey(in.GetKey(), func(ctx context.Context) {
		if !s.cache.Get(in.GetKey(), &edgeproto.Alert{}) {
			syncedCache <- true
		}
	})
	defer watchCancel()
	// check if object is already synced with cache
	if !s.cache.Get(in.GetKey(), &edgeproto.Alert{}) {
		return
	}
	select {
	case <-syncedCache:
	case <-time.After(RedisSyncTimeout):
		log.SpanLog(ctx, log.DebugLevelNotify, "Timed out waiting to sync alerts from redis to cache", "key", in.GetKeyVal())
	}

	// Note that any further actions should done as part of StoreDelete.
}

func (s *AlertApi) StoreDelete(ctx context.Context, in *edgeproto.Alert) {
	buf := edgeproto.Alert{}
	var foundAlert bool
	key := getAlertStoreKey(in)

	// Transactional function.
	txf := func(tx *redis.Tx) error {
		// Get the current value or zero.
		alertVal, err := tx.Get(key).Result()
		if err != nil && err != redis.Nil {
			return err
		}
		err = json.Unmarshal([]byte(alertVal), &buf)
		if err != nil {
			return fmt.Errorf("Failed to unmarshal alert from redis: %s, %v", alertVal, err)
		}
		if buf.NotifyId != in.NotifyId || buf.Controller != ControllerId {
			// updated by another thread or controller
			return nil
		}
		foundAlert = true

		// Operation is commited only if the watched keys remain unchanged.
		_, err = tx.TxPipelined(func(pipe redis.Pipeliner) error {
			pipe.Del(key)
			return nil
		})
		return err
	}

	// Retry if the key has been changed.
	for i := 0; i < RedisTxMaxRetries; i++ {
		err := redisClient.Watch(txf, key)
		if err == nil {
			// Success.
			break
		}
		if err == redis.TxFailedErr {
			// Optimistic lock lost. Retry.
			continue
		}
		// Return any other error.
		log.SpanLog(ctx, log.DebugLevelNotify, "Failed to delete alert from redis", "key", in.GetKeyVal(), "err", err)
	}

	// Reset HealthCheck state back to OK
	name, ok := in.Labels["alertname"]
	if ok && foundAlert && name == cloudcommon.AlertAppInstDown {
		s.appInstSetStateFromHealthCheckAlert(ctx, in, dme.HealthCheck_HEALTH_CHECK_OK)
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
		// s.sourceCache.Delete(ctx, val, 0)
		s.Delete(ctx, val, 0)
	}
}

func (s *AlertApi) CleanupAppInstAlerts(ctx context.Context, key *edgeproto.AppInstKey) {
	log.SpanLog(ctx, log.DebugLevelApi, "CleanupAppInstAlerts", "key", key)

	matches := []*edgeproto.Alert{}
	s.cache.Mux.Lock()
	labels := key.GetTags()
	for _, data := range s.cache.Objs {
		val := data.Obj
		matched := true
		for appLabelName, appLabelVal := range labels {
			if val, found := val.Labels[appLabelName]; !found || val != appLabelVal {
				matched = false
				break
			}
		}
		if matched {
			matches = append(matches, val)
		}
	}
	s.cache.Mux.Unlock()
	for _, val := range matches {
		s.Delete(ctx, val, 0)
	}
}

func (s *AlertApi) CleanupClusterInstAlerts(ctx context.Context, key *edgeproto.ClusterInstKey) {
	matches := []*edgeproto.Alert{}
	s.cache.Mux.Lock()
	for _, data := range s.cache.Objs {
		val := data.Obj
		if cloudletName, found := val.Labels[edgeproto.CloudletKeyTagName]; !found ||
			cloudletName != key.CloudletKey.Name {
			continue
		}
		if cloudletOrg, found := val.Labels[edgeproto.CloudletKeyTagOrganization]; !found ||
			cloudletOrg != key.CloudletKey.Organization {
			continue
		}
		if clusterName, found := val.Labels[edgeproto.ClusterKeyTagName]; !found ||
			clusterName != key.ClusterKey.Name {
			continue
		}
		if clusterOrg, found := val.Labels[edgeproto.ClusterInstKeyTagOrganization]; !found ||
			clusterOrg != key.Organization {
			continue
		}
		matches = append(matches, val)
	}
	s.cache.Mux.Unlock()
	for _, val := range matches {
		s.Delete(ctx, val, 0)
	}
}

func (s *AlertApi) refreshAlertKeepAliveThread() {
	s.doneKeepaliveRefresh = false
	for {
		select {
		case <-time.After(s.all.settingsApi.Get().AlertKeepaliveRefreshInterval.TimeDuration()):
		case <-s.triggerKeepaliveRefresh:
		}
		span := log.StartSpan(log.DebugLevelApi, "Alert keepalive refresh thread")
		ctx := log.ContextWithSpan(context.Background(), span)
		if s.doneKeepaliveRefresh {
			log.SpanLog(ctx, log.DebugLevelInfo, "Alert keepalive refresh done")
			span.Finish()
			break
		}
		err := s.refreshAlertKeepAlive(ctx)
		if err != nil {
			s.syncAllAlerts(ctx)
		}
		span.Finish()
	}
}

func (s *AlertApi) StopAlertKeepAliveRefresh() {
	s.doneKeepaliveRefresh = true
	select {
	case s.triggerKeepaliveRefresh <- true:
	default:
	}
}

func (s *AlertApi) refreshAlertKeepAlive(ctx context.Context) error {
	s.syncMux.Lock()
	defer s.syncMux.Unlock()
	alerts := make([]*edgeproto.Alert, 0)
	s.sourceCache.Mux.Lock()
	for _, data := range s.sourceCache.Objs {
		alert := edgeproto.Alert{}
		alert.DeepCopyIn(data.Obj)
		alerts = append(alerts, &alert)
	}
	s.sourceCache.Mux.Unlock()
	cmdOuts, err := redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for _, alert := range alerts {
			if alert.Controller != ControllerId {
				// not owned by this controller
				continue
			}
			alertKey := getAlertStoreKey(alert)
			pipe.Expire(alertKey, AlertTTL)
		}
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to refresh alert keepalive", "err", err)
		return err
	}

	// return error on failure so that we can perform a complete refresh of all the alerts
	for _, cmdOut := range cmdOuts {
		out, ok := cmdOut.(*redis.BoolCmd)
		if !ok {
			// not possible, as `Expire()` will always return BoolCmd
			return fmt.Errorf("Invalid command output type: %v", cmdOut)
		}
		expiryRefreshed, err := out.Result()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Refresh alert TTL failed", "err", err)
			return fmt.Errorf("Refresh alert TTL failed: %v", err)
		}
		if !expiryRefreshed {
			log.SpanLog(ctx, log.DebugLevelInfo, "Refresh alert TTL failed")
			return fmt.Errorf("Refresh alert TTL failed")
		}
	}
	return nil
}

func (s *AlertApi) syncAllAlerts(ctx context.Context) error {
	log.SpanLog(ctx, log.DebugLevelNotify, "Sync all alerts")
	alerts := make(map[string]string)
	s.sourceCache.Mux.Lock()
	for _, data := range s.sourceCache.Objs {
		alertObj := data.Obj
		key := getAlertStoreKey(alertObj)
		val, err := json.Marshal(alertObj)
		if err != nil {
			s.sourceCache.Mux.Unlock()
			log.SpanLog(ctx, log.DebugLevelNotify, "Failed to marshal alert object", "key", alertObj.GetKeyVal(), "err", err)
			return fmt.Errorf("Failed to marshal alert object %v, %v", alertObj, err)
		}
		alerts[key] = string(val)
	}
	s.sourceCache.Mux.Unlock()
	cmdOuts, err := redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for key, val := range alerts {
			pipe.Set(key, val, AlertTTL)
		}
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to refresh alert keepalive", "err", err)
		return err
	}

	syncCount := 0
	errors := []error{}
	for _, cmdOut := range cmdOuts {
		out, ok := cmdOut.(*redis.StatusCmd)
		if !ok {
			// not possible, as `Set()` will always return StatusCmd
			errors = append(errors, fmt.Errorf("Invalid command output type: %v", cmdOut))
			continue
		}
		_, err := out.Result()
		if err != nil {
			errors = append(errors, fmt.Errorf("Failed to set alert: %v", err))
			continue
		}
		syncCount++
	}
	log.SpanLog(ctx, log.DebugLevelNotify, "Synced all alerts", "sync count", syncCount, "errors", errors)
	return nil
}
