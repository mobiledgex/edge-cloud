package rediscache

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/mobiledgex/edge-cloud/log"
)

const (
	MaxRedisWait = time.Second * 30

	// Special IDs in the streams API
	RedisSmallestId = "-"
	RedisGreatestId = "+"
	RedisLastId     = "$"

	DefaultCfgRedisOptional int = iota
	DefaultCfgRedisStandalone
	DefaultCfgRedisHA

	RedisStandalonePort         = "6379"
	DefaultRedisStandaloneAddr  = "127.0.0.1:" + RedisStandalonePort
	RedisHeadlessService        = "redis-headless"
	RedisCloudletStandaloneAddr = RedisHeadlessService + ":" + RedisStandalonePort // for redis running in a cloudlet
	DefaultRedisMasterName      = "redismaster"
	DefaultRedisSentinelAddrs   = "127.0.0.1:26379,127.0.0.1:26380,127.0.0.1:26381"
)

type RedisConfig struct {
	MasterName     string
	SentinelAddrs  string
	StandaloneAddr string
}

func (r *RedisConfig) InitFlags(defaultCfgType int) {
	defaultMasterName := ""
	defaultSentinelAddrs := ""
	defaultStandaloneAddr := ""
	switch defaultCfgType {
	case DefaultCfgRedisOptional:
		// defaults set to empty
	case DefaultCfgRedisStandalone:
		defaultStandaloneAddr = DefaultRedisStandaloneAddr
	case DefaultCfgRedisHA:
		defaultMasterName = DefaultRedisMasterName
		defaultSentinelAddrs = DefaultRedisSentinelAddrs
	}

	flag.StringVar(&r.MasterName, "redisMasterName", defaultMasterName, "Name of the redis master node as specified in sentinel config")
	flag.StringVar(&r.SentinelAddrs, "redisSentinelAddrs", defaultSentinelAddrs, "comma separated list of redis sentinel addresses")
	flag.StringVar(&r.StandaloneAddr, "redisStandaloneAddr", defaultStandaloneAddr, "Redis standalone server address")
}

func (r *RedisConfig) AddrSpecified() bool {
	if r.SentinelAddrs != "" {
		return true
	}
	if r.StandaloneAddr != "" {
		return true
	}
	return false
}

// Supports both modes of redis server deployment:
// 1. Standalone server
// 2. Redis Sentinels (for HA)
func NewClient(ctx context.Context, cfg *RedisConfig) (*redis.Client, error) {
	if cfg.StandaloneAddr != "" {
		log.SpanLog(ctx, log.DebugLevelInfo, "init redis client for standalone server",
			"addr", cfg.StandaloneAddr)
		client := redis.NewClient(&redis.Options{
			Addr: cfg.StandaloneAddr,
		})
		return client, nil
	}

	sentinelAddrs := strings.Split(cfg.SentinelAddrs, ",")
	if len(sentinelAddrs) == 0 {
		return nil, fmt.Errorf("At least one redis sentinel address is required")
	}
	masterName := cfg.MasterName
	if masterName == "" {
		masterName = DefaultRedisMasterName
	}
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
	})
	log.SpanLog(ctx, log.DebugLevelInfo, "init redis client for HA server",
		"addr", cfg.SentinelAddrs, "mastername", cfg.MasterName)
	return client, nil
}

func IsServerReady(client *redis.Client, timeout time.Duration) error {
	start := time.Now()
	var err error
	for {
		_, err = client.Ping().Result()
		if err == nil {
			return nil
		}
		elapsed := time.Since(start)
		if elapsed >= (timeout) {
			err = fmt.Errorf("timed out")
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("Failed to ping redis - %v", err)
}
