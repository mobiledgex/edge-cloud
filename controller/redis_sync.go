package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-redis/redis"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/rediscache"
)

type RedisSync struct {
	syncCancel context.CancelFunc
	syncDone   chan bool
	allApis    *AllApis
}

func InitRedisSync(allApis *AllApis) *RedisSync {
	sync := RedisSync{}
	sync.allApis = allApis
	sync.syncDone = make(chan bool)
	return &sync
}

// Sync data from redis with controller cache
func (s *RedisSync) Start(ctx context.Context) {

	// Perform initial sync of redis data with controller cache
	err := s.syncAll(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "redis sync all failed", "err", err)
	}

	// this is telling redis to publish events since it's off by default.
	// set appropriate config options to listen for alert events:
	// - K: Keyspace events, published with __keyspace@<db>__ prefix.
	// - g: Generic commands (non-type specific) like DEL, EXPIRE, RENAME, ...
	// - x: Expired events (events generated every time a key expires)
	// - $: String commands
	_, err = redisClient.ConfigSet("notify-keyspace-events", "Kgx$").Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "unable to set keyspace events", "err", err.Error())
		return
	}

	// sync any updates to redis with controller cache
	s.syncWithNotifyCache(ctx)
}

func (s *RedisSync) syncAll(ctx context.Context) error {

	log.SpanLog(ctx, log.DebugLevelInfo, "Start sync of redis data")

	// fetch alert data from redis and sync it with controller cache
	s.syncAlertCache(ctx)

	log.SpanLog(ctx, log.DebugLevelInfo, "Done sync of redis data")
	return nil
}

func (s *RedisSync) syncAlertCache(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfo, "Start sync of redis alert data")
	pattern := getAllAlertsKeyPattern()
	syncCount := 0
	alertKeys, err := redisClient.Keys(pattern).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to find alerts from redis", "pattern", pattern, "err", err)
		return
	}

	cmdOuts, err := redisClient.Pipelined(func(pipe redis.Pipeliner) error {
		for _, alertKey := range alertKeys {
			pipe.Get(alertKey)
		}
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to sync alerts from redis", "err", err)
		return
	}

	for _, cmdOut := range cmdOuts {
		alertVal, err := cmdOut.(*redis.StringCmd).Result()
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
		s.allApis.alertApi.cache.Update(ctx, &obj, 0)
		syncCount++
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "Done sync of redis alert data", "sync-count", syncCount)
}

// Watch on all alert keys changes in a single thread.
// Set/Del event from redis will update controller cache respectively
func (s *RedisSync) syncWithNotifyCache(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfo, "Sync redis data with controller cache")

	// PSubscribe subscribes the client to the specified channels based on pattern specified.
	// Note that this method does not wait on a response from Redis, so the
	// subscription may not be active immediately. To force the connection to wait,
	// we call the Receive() method on the returned *PubSub
	pubsub := redisClient.PSubscribe(fmt.Sprintf("__keyspace@*__:%s", getAllAlertsKeyPattern()))
	_, err := pubsub.Receive()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to subscribe to keyspace notification stream", "err", err)
		return
	}

	// Go channel to receives messages.
	ch := pubsub.Channel()

	ctx, cancel := context.WithCancel(context.Background())
	s.syncCancel = cancel
	go func() {
		for {
			select {
			case chObj := <-ch:
				if chObj == nil {
					continue
				}
				span := log.StartSpan(log.DebugLevelInfo, "redis-sync-start")
				ctx := log.ContextWithSpan(ctx, span)
				parts := strings.SplitN(chObj.Channel, ":", 2)
				alertKey := parts[1]
				event := chObj.Payload
				switch event {
				case rediscache.RedisEventSet:
					alertVal, err := redisClient.Get(alertKey).Result()
					if err != nil {
						log.SpanLog(ctx, log.DebugLevelInfo, "Failed to get alert from redis", "alertKey", alertKey, "err", err)
						span.Finish()
						continue
					}
					var obj edgeproto.Alert
					err = json.Unmarshal([]byte(alertVal), &obj)
					if err != nil {
						log.SpanLog(ctx, log.DebugLevelInfo, "Failed to unmarshal alert from redis", "alert", alertVal, "err", err)
						span.Finish()
						continue
					}
					s.allApis.alertApi.cache.Update(ctx, &obj, 0)
				case rediscache.RedisEventDel:
					fallthrough
				case rediscache.RedisEventExpired:
					var obj edgeproto.Alert
					aKey := objstore.DbKeyPrefixRemove(alertKey)
					edgeproto.AlertKeyStringParse(aKey, &obj)
					s.allApis.alertApi.cache.Delete(ctx, &obj, 0)
				}
				span.Finish()
			case <-s.syncDone:
				s.syncCancel()
				pubsub.Close()
			}
		}
	}()
}

func (s *RedisSync) Done() {
	close(s.syncDone)
}
