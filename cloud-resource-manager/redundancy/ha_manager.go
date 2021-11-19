package redundancy

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"

	"github.com/mobiledgex/edge-cloud/log"
)

const RedisPingFail string = "Redis Ping Fail"
const HighAvailabilityManagerDisabled = "HighAvailabilityManagerDisabled"

const MaxRedisWait = time.Second * 30
const CheckActiveLogInterval = time.Second * 10

type HAWatcher interface {
	ActiveChanged(ctx context.Context, platformActive bool) error
}

type HighAvailabilityManager struct {
	redisAddr              string
	nodeGroupKey           string
	redisClient            *redis.Client
	HARole                 string
	HAEnabled              bool
	activeDuration         time.Duration
	activePollInterval     time.Duration
	PlatformInstanceActive bool
	haWatcher              HAWatcher
	nodeMgr                *node.NodeMgr
}

func (s *HighAvailabilityManager) InitFlags() {
	flag.StringVar(&s.redisAddr, "redisAddr", "", "redis address")
	flag.StringVar(&s.HARole, "HARole", "", string(process.HARolePrimary+" or "+process.HARoleSecondary))
}

func (s *HighAvailabilityManager) Init(nodeGroupKey string, nodeMgr *node.NodeMgr, activeDuration, activePollInterval edgeproto.Duration, haWatcher HAWatcher) error {
	ctx := log.ContextWithSpan(context.Background(), log.NoTracingSpan())
	log.SpanLog(ctx, log.DebugLevelInfo, "HighAvailabilityManager init", "nodeGroupKey", nodeGroupKey, "role", s.HARole, "activeDuration", activeDuration, "activePollInterval", activePollInterval)
	s.activeDuration = activeDuration.TimeDuration()
	s.activePollInterval = activePollInterval.TimeDuration()
	s.nodeMgr = nodeMgr
	s.haWatcher = haWatcher
	if s.HARole != string(process.HARolePrimary) && s.HARole != string(process.HARoleSecondary) {
		return fmt.Errorf("invalid HA Role type")
	}
	if s.redisAddr == "" {
		s.PlatformInstanceActive = true
		return fmt.Errorf("%s Redis Addr for HA not specified", HighAvailabilityManagerDisabled)
	}
	s.nodeGroupKey = nodeGroupKey
	s.HAEnabled = true
	err := s.connectRedis(ctx)
	if err != nil {
		return err
	}
	s.PlatformInstanceActive = s.tryActive(ctx)
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

// TryActive is called on startup
func (s *HighAvailabilityManager) tryActive(ctx context.Context) bool {

	if s.PlatformInstanceActive {
		// this should not happen. Only 1 thread should be doing tryActive
		log.FatalLog("Platform already active")
	}
	// see if we are already active, which can happen if the process just died and was restarted quickly
	// before the SetNX expired
	alreadyActive := s.CheckActive(ctx)
	if alreadyActive {
		log.SpanLog(ctx, log.DebugLevelInfra, "Platform already active", "key", s.nodeGroupKey)
		s.PlatformInstanceActive = true
		// bump the active timer in case it is close to expiry
		s.BumpActiveExpire(ctx)
		return true
	}
	cmd := s.redisClient.SetNX(s.nodeGroupKey, s.HARole, s.activeDuration)
	v, err := cmd.Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "tryActive setNX error", "key", s.nodeGroupKey, "cmd", cmd, "v", v, "err", err)
	}
	return v
}

func (s *HighAvailabilityManager) CheckActive(ctx context.Context) bool {

	cmd := s.redisClient.Get(s.nodeGroupKey)
	v, err := cmd.Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "CheckActive error", "key", s.nodeGroupKey, "cmd", cmd, "v", v, "err", err)
		return false
	}
	isActive := v == s.HARole
	return isActive
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

func (s *HighAvailabilityManager) CheckActiveLoop(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CheckActiveLoop", "activePollInterval", s.activePollInterval, "activeDuration", s.activeDuration)
	timeLastLog := time.Now() // log only once every X seconds in this loop
	timeLastBumpActive := time.Now()
	timeLastCheckActive := time.Now()
	for {
		elapsedSinceLog := time.Since(timeLastLog)
		elapsedSinceBumpActive := time.Since(timeLastBumpActive)
		elapsedSinceCheckActive := time.Since(timeLastCheckActive)

		if !s.PlatformInstanceActive {
			if elapsedSinceLog >= CheckActiveLogInterval {
				log.SpanLog(ctx, log.DebugLevelInfra, "Platform inactive, doing TryActive")
				timeLastLog = time.Now()
			}
			newActive := s.tryActive(ctx)
			if newActive {
				log.SpanLog(ctx, log.DebugLevelInfra, "Platform became active")
				s.PlatformInstanceActive = true
				elapsedSinceBumpActive = time.Since(time.Now())
				s.haWatcher.ActiveChanged(ctx, true)
				s.nodeMgr.Event(ctx, "High Availability Node Active", s.nodeMgr.MyNode.Key.CloudletKey.Organization, s.nodeMgr.MyNode.Key.CloudletKey.GetTags(), nil, "Node Type", s.nodeMgr.MyNode.Key.Type, "Newly Active Instance", s.HARole)
			}
		} else {
			if elapsedSinceLog >= CheckActiveLogInterval {
				log.SpanLog(ctx, log.DebugLevelInfra, "Platform active, doing BumpActiveExpire")
				timeLastLog = time.Now()
			}
			checkStillActive := false
			if elapsedSinceBumpActive >= s.activePollInterval*2 {
				// Somehow we missed at least one poll. this can cause another process to seize activity
				// We need to check we are still active to avoid a split brain situation
				log.SpanLog(ctx, log.DebugLevelInfra, "Missed active poll", "elapsedSinceBumpActive", elapsedSinceBumpActive)
				checkStillActive = true
			}
			// periodically we check that we are still active to detect an unexpected split brain scenario
			if elapsedSinceCheckActive > s.activeDuration {
				checkStillActive = true
			}
			if checkStillActive {
				stillActive := s.CheckActive(ctx)
				if !stillActive {
					// somehow we lost activity, this is a big problem
					log.SpanLog(ctx, log.DebugLevelInfo, "ERROR!: Lost activity")
					s.PlatformInstanceActive = false
					s.haWatcher.ActiveChanged(ctx, false)
				}
				timeLastCheckActive = time.Now()
			}
			err := s.BumpActiveExpire(ctx)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "BumpActiveExpire failed, retry", "err", err)
				err = s.BumpActiveExpire(ctx)
				if err != nil {
					log.FatalLog("BumpActiveExpire failed!", "err", err)
				}
			}
			timeLastBumpActive = time.Now()

		}
		time.Sleep(s.activePollInterval)
	}
}
