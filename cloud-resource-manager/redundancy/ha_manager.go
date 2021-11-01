package redundancy

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"

	"github.com/mobiledgex/edge-cloud/log"
)

const RedisPingFail string = "Redis Ping Fail"
const HighAvailabilityManagerDisabled = "HighAvailabilityManagerDisabled"

const MaxRedisWait = time.Second * 30
const CheckActiveLogInterval = time.Second * 10

var PlatformInstanceActive bool

type HighAvailabilityManager struct {
	RedisAddr          string
	NodeGroupKey       string
	RedisClient        *redis.Client
	HARole             string
	Platform           pf.Platform
	activeDuration     time.Duration
	activePollInterval time.Duration
}

func (s *HighAvailabilityManager) InitFlags() {
	flag.StringVar(&s.RedisAddr, "redisAddr", "127.0.0.1:6379", "redis address")
	flag.StringVar(&s.HARole, "HARole", "", string(process.HARolePrimary+" or "+process.HARoleSecondary))
}

func (s *HighAvailabilityManager) SetPlatform(platform pf.Platform) {
	s.Platform = platform
}

func (s *HighAvailabilityManager) Init(nodeGroupKey string, activeDuration, activePollInterval edgeproto.Duration) error {
	ctx := log.ContextWithSpan(context.Background(), log.NoTracingSpan())
	log.SpanLog(ctx, log.DebugLevelInfo, "HighAvailabilityManager init", "nodeGroupKey", nodeGroupKey, "role", s.HARole, "activeDuration", activeDuration, "activePollInterval", activePollInterval)
	s.activeDuration = activeDuration.TimeDuration()
	s.activePollInterval = activePollInterval.TimeDuration()

	if s.HARole == "" {
		PlatformInstanceActive = true
		return fmt.Errorf("%s HA Role not specified", HighAvailabilityManagerDisabled)
	}
	if s.HARole != string(process.HARolePrimary) && s.HARole != string(process.HARoleSecondary) {
		return fmt.Errorf("invalid node type")
	}
	if s.RedisAddr == "" {
		return fmt.Errorf("Redis address not specified")
	}
	s.NodeGroupKey = nodeGroupKey
	if s.NodeGroupKey == "" {
		return fmt.Errorf("group key node specified")
	}
	err := s.connectRedis(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *HighAvailabilityManager) pingRedis(ctx context.Context) error {
	log.SpanLog(ctx, log.DebugLevelInfo, "pingRedis")

	pong, err := s.RedisClient.Ping().Result()
	log.SpanLog(ctx, log.DebugLevelInfo, "redis ping done", "pong", pong, "err", err)

	if err != nil {
		return fmt.Errorf("%s - %v", RedisPingFail, err)
	}
	return nil
}

func (s *HighAvailabilityManager) connectRedis(ctx context.Context) error {
	log.SpanLog(ctx, log.DebugLevelInfo, "connectRedis")
	if s.RedisAddr == "" {
		return fmt.Errorf("Redis address not specified")
	}
	if s.NodeGroupKey == "" {
		return fmt.Errorf("group key node not specified")
	}
	s.RedisClient = redis.NewClient(&redis.Options{
		Addr: s.RedisAddr,
	})
	start := time.Now()
	var err error
	for {
		err = s.pingRedis(ctx)
		if err == nil {
			return nil
		}
		elapsed := time.Since(start)
		if elapsed >= (MaxRedisWait) {
			// for now we will return no errors when we time out.  In future we will use some other state or status
			// field to reflect this and employ health checks to track these appinsts
			log.InfoLog("redis wait timed out")
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "pingRedis failed", "err", err)
	return fmt.Errorf("pingRedis failed - %v", err)
}

func (s *HighAvailabilityManager) TryActive(ctx context.Context) bool {
	cmd := s.RedisClient.SetNX(s.NodeGroupKey, s.HARole, s.activeDuration)
	v, err := cmd.Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "TryActive setNX error", "key", s.NodeGroupKey, "cmd", cmd, "v", v, "err", err)
	}
	PlatformInstanceActive = v
	return v
}

func (s *HighAvailabilityManager) BumpActiveExpire(ctx context.Context) error {
	cmd := s.RedisClient.Set(s.NodeGroupKey, s.HARole, s.activeDuration)
	v, err := cmd.Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "BumpActiveExpire error", "key", s.NodeGroupKey, "cmd", cmd, "v", v, "err", err)
		return err
	}
	if v != "OK" {
		return fmt.Errorf("BumpActiveExpire returned unexpected value - %s", v)
	}
	return nil
}

func (s *HighAvailabilityManager) CheckActiveLoop(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfo, "CheckActiveLoop")
	timeSinceLog := time.Now() // log only once every X seconds in this loop
	for {
		elaspsed := time.Since(timeSinceLog)
		if !PlatformInstanceActive {
			if elaspsed >= CheckActiveLogInterval {
				log.SpanLog(ctx, log.DebugLevelInfo, "Platform inactive, doing TryActive")
				timeSinceLog = time.Now()
			}
			newActive := s.TryActive(ctx)
			if newActive {
				log.SpanLog(ctx, log.DebugLevelInfo, "Platform became active")
				PlatformInstanceActive = true
				if s.Platform != nil {
					s.Platform.BecomeActive(ctx, s.HARole)
				}
			}
		} else {
			if elaspsed >= CheckActiveLogInterval {
				log.SpanLog(ctx, log.DebugLevelInfo, "Platform active, doing BumpActiveExpire")
				timeSinceLog = time.Now()
			}
			err := s.BumpActiveExpire(ctx)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "BumpActiveExpire failed, retry", "err", err)
				err = s.BumpActiveExpire(ctx)
				if err != nil {
					log.FatalLog("BumpActiveExpire failed!", "err", err)
				}
			}
		}
		time.Sleep(s.activePollInterval)
	}
}
