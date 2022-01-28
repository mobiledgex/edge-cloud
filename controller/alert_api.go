package main

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type AlertApi struct {
	all   *AllApis
	sync  *Sync
	cache edgeproto.AlertCache
}

func NewAlertApi(sync *Sync, all *AllApis) *AlertApi {
	alertApi := AlertApi{}
	alertApi.all = all
	alertApi.sync = sync
	edgeproto.InitAlertCache(&alertApi.cache)
	sync.RegisterCache(&alertApi.cache)
	return &alertApi
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

func getAlertStoreKey(in *edgeproto.Alert) string {
	return objstore.DbKeyString("Alert", in.GetKey())
}

func getAllAlertsKeyPattern() string {
	return objstore.DbKeyPrefixString("Alert") + "/*"
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

	// Store alerts in controller managed cache and also in redis cache.
	s.cache.Update(ctx, in, rev)
	// Alerts are stored in redis cache, instead of etcd for the following reasons:
	// * They are transient in nature, this causes etcd fragmentation and hinders scalability
	// * It can be re-triggered/re-computed and hence it doesn't need a persistent storage
	key := getAlertStoreKey(in)
	val, err := json.Marshal(in)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "Failed to marshal alert object", "key", in.GetKeyVal(), "err", err)
		return
	}
	_, err = redisClient.Set(key, val, 0).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "Failed to store alert in rediscache", "key", in.GetKeyVal(), "err", err)
		return
	}

	if name == cloudcommon.AlertAppInstDown {
		state, ok := in.Labels[cloudcommon.AlertHealthCheckStatus]
		if !ok {
			log.SpanLog(ctx, log.DebugLevelNotify, "HealthCheck status not found",
				"labels", in.Labels)
			return
		}
		hcState, err := strconv.Atoi(state)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelNotify, "failed to parse Health Check state",
				"state", state, "error", err)
			return
		}
		s.appInstSetStateFromHealthCheckAlert(ctx, in, dme.HealthCheck(hcState))
	}
}

func (s *AlertApi) Delete(ctx context.Context, in *edgeproto.Alert, rev int64) {
	// Add a region label
	in.Labels["region"] = *region
	s.cache.Delete(ctx, in, rev)

	key := getAlertStoreKey(in)
	_, err := redisClient.Del(key).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "notify delete alert redis failure", "key", in.GetKeyVal(), "err", err)
	}

	// Reset HealthCheck state back to OK
	name, ok := in.Labels["alertname"]
	if ok && name == cloudcommon.AlertAppInstDown {
		s.appInstSetStateFromHealthCheckAlert(ctx, in, dme.HealthCheck_HEALTH_CHECK_OK)
	}
}

func (s *AlertApi) Flush(ctx context.Context, notifyId int64) {
	// Delete all data from controller cache and redis cache
	s.cache.Flush(ctx, notifyId)
	pattern := getAllAlertsKeyPattern()
	alertKeys, err := redisClient.Keys(pattern).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelNotify, "notify flush alert redis failure", "err", err)
		return
	}
	if len(alertKeys) > 0 {
		_, err = redisClient.Del(alertKeys...).Result()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelNotify, "notify flush alert redis failure", "err", err)
			return
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
