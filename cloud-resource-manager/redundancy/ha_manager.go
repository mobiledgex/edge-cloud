package redundancy

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"

	"github.com/mobiledgex/edge-cloud/log"
)

const RedisPingFail string = "Redis Ping Fail"
const HighAvailabilityManagerDisabled = "HighAvailabilityManagerDisabled"

const ActiveDuration = time.Second * 10
const ActivePollInterval = time.Second * 3

var PlatformInstanceActive bool

type HighAvailabilityManager struct {
	RedisAddr    string
	NodeGroupKey string
	RedisClient  *redis.Client
	HARole       string
	Platform     pf.Platform
}

func (s *HighAvailabilityManager) InitFlags() {
	flag.StringVar(&s.RedisAddr, "redisAddr", "127.0.0.1:6379", "redis address")
	flag.StringVar(&s.HARole, "HARole", cloudcommon.HARolePrimary, cloudcommon.HARolePrimary+" or "+cloudcommon.HARoleSecondary)
}

func (s *HighAvailabilityManager) SetPlatform(platform pf.Platform) {
	s.Platform = platform
}

func (s *HighAvailabilityManager) Init(nodeGroupKey string) error {
	ctx := log.ContextWithSpan(context.Background(), log.NoTracingSpan())
	log.SpanLog(ctx, log.DebugLevelInfo, "HighAvailabilityManager init")

	if s.HARole == "" {
		PlatformInstanceActive = true
		return fmt.Errorf("%s HA Role not specified", HighAvailabilityManagerDisabled)
	}
	if s.HARole != cloudcommon.HARolePrimary && s.HARole != cloudcommon.HARoleSecondary {
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
	return s.pingRedis(ctx)
}

func (s *HighAvailabilityManager) TryActive(ctx context.Context) bool {
	log.SpanLog(ctx, log.DebugLevelInfo, "TryActive")

	cmd := s.RedisClient.SetNX(s.NodeGroupKey, s.HARole, ActiveDuration)
	v, e := cmd.Result()
	log.SpanLog(ctx, log.DebugLevelInfo, "TryActive setNX result", "key", s.NodeGroupKey, "cmd", cmd, "v", v, "e", e)
	PlatformInstanceActive = v
	return v
}

func (s *HighAvailabilityManager) BumpActiveExpire(ctx context.Context) error {
	log.SpanLog(ctx, log.DebugLevelInfo, "BumpActiveExpire")

	cmd := s.RedisClient.Set(s.NodeGroupKey, s.HARole, ActiveDuration)
	v, err := cmd.Result()
	log.SpanLog(ctx, log.DebugLevelInfo, "BumpActiveExpire result", "key", s.NodeGroupKey, "cmd", cmd, "v", v, "err", err)
	if err != nil {
		return err
	}
	if v != "OK" {
		return fmt.Errorf("BumpActiveExpire returned unexpected value - %s", v)
	}
	return nil
}

func (s *HighAvailabilityManager) CheckActiveLoop(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfo, "CheckActiveLoop")
	//....if redis dies and active, remain active..
	for {
		if !PlatformInstanceActive {
			newActive := s.TryActive(ctx)
			if newActive {
				log.SpanLog(ctx, log.DebugLevelInfo, "Platform became active")
				PlatformInstanceActive = true
				if s.Platform != nil {
					s.Platform.BecomeActive(ctx, s.HARole)
				}
			}
		} else {
			err := s.BumpActiveExpire(ctx)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "BumpActiveExpire failed, retry", "err", err)
				err = s.BumpActiveExpire(ctx)
				if err != nil {
					log.FatalLog("BumpActiveExpire failed!", "err", err)
				}
			}
		}
		time.Sleep(ActivePollInterval)
	}
}
