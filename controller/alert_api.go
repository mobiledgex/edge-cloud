package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/rediscache"
)

type AlertApi struct {
	all         *AllApis
	sync        *Sync
	cache       edgeproto.AlertCache
	sourceCache edgeproto.AlertCache // source of truth from crm/etc
}

var (
	ControllerCreatedAlerts = "ControllerCreatedAlerts"

	AlertTTL = time.Duration(1 * time.Minute)
)

func NewAlertApi(sync *Sync, all *AllApis) *AlertApi {
	alertApi := AlertApi{}
	alertApi.all = all
	alertApi.sync = sync
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
	// Alerts are stored in redis cache, instead of etcd for the following reasons:
	// * They are transient in nature, this causes etcd fragmentation and hinders scalability
	// * It can be re-triggered/re-computed and hence it doesn't need a persistent storage
	key := getAlertStoreKey(new)
	val, err := json.Marshal(new)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "Failed to marshal alert object", "key", new.GetKeyVal(), "err", err)
		return
	}
	_, err = redisClient.Set(key, val, AlertTTL).Result()
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
	// Note that any further actions should done as part of StoreDelete.
}

func (s *AlertApi) StoreDelete(ctx context.Context, in *edgeproto.Alert) {
	buf := edgeproto.Alert{}
	var foundAlert bool
	key := getAlertStoreKey(in)
	_, err := redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		alertVal, err := pipe.Get(key).Result()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelNotify, "Failed to get alert from redis", "key", in.GetKeyVal(), "err", err)
			return nil
		}
		err = json.Unmarshal([]byte(alertVal), &buf)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to unmarshal alert from redis", "alert", alertVal, "err", err)
			return nil
		}
		if buf.NotifyId != in.NotifyId || buf.Controller != ControllerId {
			// updated by another thread or controller
			return nil
		}
		foundAlert = true
		_, err = pipe.Del(key).Result()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelNotify, "Failed to delete alert from redis", "key", in.GetKeyVal(), "err", err)
		}
		return nil
	})
	if err != nil {
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
		s.sourceCache.Delete(ctx, val, 0)
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
		s.sourceCache.Delete(ctx, val, 0)
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
		s.sourceCache.Delete(ctx, val, 0)
	}
}

func (s *AlertApi) syncSourceData(ctx context.Context) error {
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

func (s *AlertApi) refreshAlertKeepAliveThread() {
	for {
		refreshInterval := s.all.settingsApi.Get().AlertKeepaliveRefreshInterval.TimeDuration()
		time.Sleep(refreshInterval)
		span := log.StartSpan(log.DebugLevelApi, "Alert keepalive refresh thread")
		ctx := log.ContextWithSpan(context.Background(), span)
		s.refreshAlertKeepAlive(ctx)
		span.Finish()
	}
}

func (s *AlertApi) refreshAlertKeepAlive(ctx context.Context) {
	alerts := make([]*edgeproto.Alert, 0)
	s.sourceCache.Mux.Lock()
	for _, data := range s.sourceCache.Objs {
		alert := edgeproto.Alert{}
		alert.DeepCopyIn(data.Obj)
		alerts = append(alerts, &alert)
	}
	s.sourceCache.Mux.Unlock()
	_, err := redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for _, alert := range alerts {
			if alert.Controller != ControllerId {
				// not owned by this controller
				continue
			}
			alertKey := getAlertStoreKey(alert)
			_, err := pipe.Expire(alertKey, AlertTTL).Result()
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Refresh alert TTL failed", "alertKey", alertKey, "err", err)
				continue
			}
		}
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to refresh alert keepalive", "err", err)
	}
	return
}

func (s *AlertApi) syncRedisWithControllerCache(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfo, "Sync redis data with controller cache")
	// this is telling redis to publish events since it's off by default.
	_, err := redisClient.ConfigSet("notify-keyspace-events", "Kg$").Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "unable to set keyspace events", "err", err.Error())
		return
	}

	pubsub := redisClient.PSubscribe(fmt.Sprintf("__keyspace@*__:%s", getAllAlertsKeyPattern()))
	// Go channel to receives messages.
	ch := pubsub.Channel()
	for {
		select {
		case chObj := <-ch:
			parts := strings.Split(chObj.Channel, ":")
			alertKey := parts[1]
			alertVal, err := redisClient.Get(alertKey).Result()
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Failed to get alert from redis", "alertKey", alertKey, "err", err)
				continue
			}
			var obj edgeproto.Alert
			err = json.Unmarshal([]byte(alertVal), &obj)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Failed to unmarshal alert from redis", "alert", alertVal, "err", err)
				continue
			}
			event := chObj.Payload
			switch event {
			case rediscache.RedisEventSet:
				s.cache.Update(ctx, &obj, 0)
			case rediscache.RedisEventDel:
				s.cache.Delete(ctx, &obj, 0)
			}

		}
	}

}
