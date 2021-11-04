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
	redisAddr          string
	nodeGroupKey       string
	cloudletKey        *edgeproto.CloudletKey
	redisClient        *redis.Client
	HARole             string
	platform           pf.Platform
	activeDuration     time.Duration
	activePollInterval time.Duration
	cloudletInfoCache  *edgeproto.CloudletInfoCache
}

func (s *HighAvailabilityManager) InitFlags() {
	flag.StringVar(&s.redisAddr, "redisAddr", "127.0.0.1:6379", "redis address")
	flag.StringVar(&s.HARole, "HARole", "", string(process.HARolePrimary+" or "+process.HARoleSecondary))
}

func (s *HighAvailabilityManager) Init(nodeGroupKey string, activeDuration, activePollInterval edgeproto.Duration, platform pf.Platform, cloudletKey *edgeproto.CloudletKey, cloudletInfoCache *edgeproto.CloudletInfoCache) error {
	ctx := log.ContextWithSpan(context.Background(), log.NoTracingSpan())
	log.SpanLog(ctx, log.DebugLevelInfo, "HighAvailabilityManager init", "nodeGroupKey", nodeGroupKey, "role", s.HARole, "activeDuration", activeDuration, "activePollInterval", activePollInterval)
	s.activeDuration = activeDuration.TimeDuration()
	s.activePollInterval = activePollInterval.TimeDuration()
	s.platform = platform
	s.cloudletKey = cloudletKey
	s.cloudletInfoCache = cloudletInfoCache

	if s.HARole == "" {
		PlatformInstanceActive = true
		return fmt.Errorf("%s HA Role not specified", HighAvailabilityManagerDisabled)
	}
	if s.HARole != string(process.HARolePrimary) && s.HARole != string(process.HARoleSecondary) {
		return fmt.Errorf("invalid node type")
	}
	if s.redisAddr == "" {
		return fmt.Errorf("Redis address not specified")
	}
	s.nodeGroupKey = nodeGroupKey
	if s.nodeGroupKey == "" {
		return fmt.Errorf("group key node specified")
	}
	err := s.connectRedis(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *HighAvailabilityManager) pingRedis(ctx context.Context) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "pingRedis")

	pong, err := s.redisClient.Ping().Result()
	log.SpanLog(ctx, log.DebugLevelInfra, "redis ping done", "pong", pong, "err", err)

	if err != nil {
		return fmt.Errorf("%s - %v", RedisPingFail, err)
	}
	return nil
}

func (s *HighAvailabilityManager) connectRedis(ctx context.Context) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "connectRedis")
	if s.redisAddr == "" {
		return fmt.Errorf("Redis address not specified")
	}
	if s.nodeGroupKey == "" {
		return fmt.Errorf("group key node not specified")
	}
	s.redisClient = redis.NewClient(&redis.Options{
		Addr: s.redisAddr,
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
			log.SpanLog(ctx, log.DebugLevelInfra, "redis wait timed out", "err", err)
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "pingRedis failed", "err", err)
	return fmt.Errorf("pingRedis failed - %v", err)
}

func (s *HighAvailabilityManager) TryActive(ctx context.Context) bool {

	if PlatformInstanceActive {
		// this should not happen. Only 1 thread should be doing TryActive
		log.FatalLog("Platform already active")
	}
	cmd := s.redisClient.SetNX(s.nodeGroupKey, s.HARole, s.activeDuration)
	v, err := cmd.Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "TryActive setNX error", "key", s.nodeGroupKey, "cmd", cmd, "v", v, "err", err)
	}
	PlatformInstanceActive = v
	return v
}

func (s *HighAvailabilityManager) BumpActiveExpire(ctx context.Context) error {
	cmd := s.redisClient.Set(s.nodeGroupKey, s.HARole, s.activeDuration)
	v, err := cmd.Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "BumpActiveExpire error", "key", s.nodeGroupKey, "cmd", cmd, "v", v, "err", err)
		return err
	}
	if v != "OK" {
		return fmt.Errorf("BumpActiveExpire returned unexpected value - %s", v)
	}
	return nil
}
func (s *HighAvailabilityManager) UpdateCloudletInfoForActive(ctx context.Context) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateCloudletInfoForActive")

	var cloudletInfo edgeproto.CloudletInfo
	if !s.cloudletInfoCache.Get(s.cloudletKey, &cloudletInfo) {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to update cloudlet info, cannot find in cache", "cloudletKey", s.cloudletKey)
		return fmt.Errorf("Cannot find in cloudlet info in cache for key %s", s.cloudletKey.String())
	}
	cloudletInfo.ActiveCrmInstance = s.HARole
	cloudletInfo.StandbyCrm = false
	s.cloudletInfoCache.Update(ctx, &cloudletInfo, 0)
	return nil
}

func (s *HighAvailabilityManager) CheckActiveLoop(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CheckActiveLoop")
	timeSinceLog := time.Now() // log only once every X seconds in this loop
	for {
		elaspsed := time.Since(timeSinceLog)
		if !PlatformInstanceActive {
			if elaspsed >= CheckActiveLogInterval {
				log.SpanLog(ctx, log.DebugLevelInfra, "Platform inactive, doing TryActive")
				timeSinceLog = time.Now()
			}
			newActive := s.TryActive(ctx)
			if newActive {
				log.SpanLog(ctx, log.DebugLevelInfra, "Platform became active")
				if s.platform != nil {
					err := s.UpdateCloudletInfoForActive(ctx)
					if err != nil {
						log.FatalLog("Unable to update cloudlet info", "err", err)
					}
					s.platform.BecomeActive(ctx, s.HARole)
				}
			}
		} else {
			if elaspsed >= CheckActiveLogInterval {
				log.SpanLog(ctx, log.DebugLevelInfra, "Platform active, doing BumpActiveExpire")
				timeSinceLog = time.Now()
			}
			err := s.BumpActiveExpire(ctx)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "BumpActiveExpire failed, retry", "err", err)
				err = s.BumpActiveExpire(ctx)
				if err != nil {
					log.FatalLog("BumpActiveExpire failed!", "err", err)
				}
			}
		}
		time.Sleep(s.activePollInterval)
	}
}
