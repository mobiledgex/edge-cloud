package main

import (
	"context"
	"encoding/json"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// Sync data from redis with controller cache
func syncRedisData(ctx context.Context, allApis *AllApis) {
	log.SpanLog(ctx, log.DebugLevelInfo, "Start sync of redis data")

	// fetch alert data from redis and sync it with controller cache
	syncAlertCache(ctx, &allApis.alertApi.cache)

	log.SpanLog(ctx, log.DebugLevelInfo, "Done sync of redis data")
}

func syncAlertCache(ctx context.Context, cache *edgeproto.AlertCache) {
	log.SpanLog(ctx, log.DebugLevelInfo, "Start sync of redis alert data")
	pattern := getAllAlertsKeyPattern()
	alertKeys, err := redisClient.Keys(pattern).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to sync alerts from redis", "err", err)
		return
	}
	syncCount := 0
	for _, alertKey := range alertKeys {
		alertVal, err := redisClient.Get(alertKey).Result()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to sync alerts from redis", "alert", alertVal, "err", err)
			continue
		}
		var obj edgeproto.Alert
		err = json.Unmarshal([]byte(alertVal), &obj)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to sync alerts from redis", "alert", alertVal, "err", err)
			continue
		}
		cache.Update(ctx, &obj, 0)
		syncCount++
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "Done sync of redis alert data", "sync-count", syncCount)
}
