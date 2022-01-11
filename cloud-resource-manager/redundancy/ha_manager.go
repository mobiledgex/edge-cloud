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
	"github.com/mobiledgex/edge-cloud/rediscache"

	"github.com/mobiledgex/edge-cloud/log"
)

const RedisPingFail string = "Redis Ping Fail"
const HighAvailabilityManagerDisabled = "HighAvailabilityManagerDisabled"

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
	flag.StringVar(&s.HARole, "HARole", string(process.HARolePrimary), string(process.HARolePrimary+" or "+process.HARoleSecondary))
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
	if s.nodeGroupKey == "" {
		return fmt.Errorf("group key node not specified")
	}
	s.HAEnabled = true

	var err error
	s.redisClient, err = rediscache.NewClient(s.redisAddr)
	if err != nil {
		return err
	}
	if err := rediscache.IsServerReady(s.redisClient, rediscache.MaxRedisWait); err != nil {
		return err
	}

	s.PlatformInstanceActive = s.tryActive(ctx)
	return nil
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
	v, err := s.redisClient.SetNX(s.nodeGroupKey, s.HARole, s.activeDuration).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "tryActive setNX error", "key", s.nodeGroupKey, "v", v, "err", err)
	}
	return v
}

func (s *HighAvailabilityManager) CheckActive(ctx context.Context) bool {

	v, err := s.redisClient.Get(s.nodeGroupKey).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "CheckActive error", "key", s.nodeGroupKey, "v", v, "err", err)
		return false
	}
	isActive := v == s.HARole
	return isActive
}

func (s *HighAvailabilityManager) BumpActiveExpire(ctx context.Context) error {
	v, err := s.redisClient.Set(s.nodeGroupKey, s.HARole, s.activeDuration).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "BumpActiveExpire error", "key", s.nodeGroupKey, "v", v, "err", err)
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
