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
	ActiveChangedPreSwitch(ctx context.Context, platformActive bool) error  // actions before setting PlatformInstanceActive
	ActiveChangedPostSwitch(ctx context.Context, platformActive bool) error // actions after setting PlatformInstanceActive
}

type HighAvailabilityManager struct {
	redisAddr                  string
	nodeGroupKey               string
	redisClient                *redis.Client
	HARole                     string
	HAEnabled                  bool
	activeDuration             time.Duration
	activePollInterval         time.Duration
	PlatformInstanceActive     bool
	ActiveTransitionInProgress bool // active from a redis standpoint but still doing switchover actions
	RedisConnectionFailed      bool // used when primary becomes active due to redis unavailability
	haWatcher                  HAWatcher
	nodeMgr                    *node.NodeMgr
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
	s.HAEnabled = true
	err := s.connectRedis(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "connectRedis failed", "err", err)
		if s.HARole == string(process.HARolePrimary) {
			log.SpanLog(ctx, log.DebugLevelInfo, "Assuming active status for primary node due to redis failure")
			s.PlatformInstanceActive = true
			s.RedisConnectionFailed = true
			return nil
		}
		return err
	}
	// prior to the first active check, give the other unit an opportunity to gain activity in case this is a restart case
	time.Sleep(s.activePollInterval)
	s.PlatformInstanceActive, err = s.tryActive(ctx)
	return err
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
func (s *HighAvailabilityManager) tryActive(ctx context.Context) (bool, error) {
	if !s.RedisConnectionFailed {
		if s.PlatformInstanceActive || s.ActiveTransitionInProgress {
			// this should not happen. Only 1 thread should be doing tryActive
			log.FatalLog("Platform already active", "PlatformInstanceActive", s.PlatformInstanceActive, "ActiveTransitionInProgress", s.ActiveTransitionInProgress)
		}
		// see if we are already active, which can happen if the process just died and was restarted quickly
		// before the SetNX expired
		alreadyActive, err := s.CheckActive(ctx)
		if err != nil {
			return false, err
		}
		if alreadyActive {
			log.SpanLog(ctx, log.DebugLevelInfra, "Platform already active", "key", s.nodeGroupKey)
			// bump the active timer in case it is close to expiry
			s.BumpActiveExpire(ctx)
			return true, nil
		}
	}
	cmd := s.redisClient.SetNX(s.nodeGroupKey, s.HARole, s.activeDuration)
	v, err := cmd.Result()
	if err != nil {
		// don't log this if the redis is already down because it result in excessive logs
		if !s.RedisConnectionFailed {
			log.SpanLog(ctx, log.DebugLevelInfra, "tryActive setNX error", "key", s.nodeGroupKey, "cmd", cmd, "v", v, "err", err)
		}
		return false, err
	}
	return v, nil
}

func (s *HighAvailabilityManager) CheckActive(ctx context.Context) (bool, error) {

	cmd := s.redisClient.Get(s.nodeGroupKey)
	v, err := cmd.Result()
	if err != nil {
		if err == redis.Nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "CheckActive returns nil -- neither unit is active")
			return false, nil
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "CheckActive error", "key", s.nodeGroupKey, "cmd", cmd, "v", v, "err", err)
			return false, err
		}
	}
	isActive := v == s.HARole
	return isActive, nil
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

		if !s.PlatformInstanceActive && !s.ActiveTransitionInProgress {
			if elapsedSinceLog >= CheckActiveLogInterval {
				log.SpanLog(ctx, log.DebugLevelInfra, "Platform inactive, doing TryActive")
				timeLastLog = time.Now()
			}
			newActive, err := s.tryActive(ctx)
			if err != nil {
				log.FatalLog("Error in tryActive - %v", err)
			}
			if newActive {
				log.SpanLog(ctx, log.DebugLevelInfra, "Platform became active", "role", s.HARole)
				timeLastBumpActive = time.Now()
				// switchover is handled in a separate thread so we do not miss polling redis
				s.ActiveTransitionInProgress = true
				go func() {
					err := s.haWatcher.ActiveChangedPreSwitch(ctx, true)
					if err != nil {
						log.FatalLog("ActiveChangedPreSwitch failed - %v", err)
					}
					s.PlatformInstanceActive = true
					s.ActiveTransitionInProgress = false
					s.haWatcher.ActiveChangedPostSwitch(ctx, true)
					if err != nil {
						log.FatalLog("ActiveChangedPostSwitch failed - %v", err)
					}
					s.nodeMgr.Event(ctx, "High Availability Node Active", s.nodeMgr.MyNode.Key.CloudletKey.Organization, s.nodeMgr.MyNode.Key.CloudletKey.GetTags(), nil, "Node Type", s.nodeMgr.MyNode.Key.Type, "Newly Active Instance", s.HARole)
				}()
			}
		} else {
			if elapsedSinceLog >= CheckActiveLogInterval {
				log.SpanLog(ctx, log.DebugLevelInfra, "Unit is active", "ActiveTransitionInProgress", s.ActiveTransitionInProgress, "PlatformInstanceActive", s.PlatformInstanceActive, "RedisConnectionFailed", s.RedisConnectionFailed, "role", s.HARole)
				timeLastLog = time.Now()
			}
			if s.RedisConnectionFailed {
				// redis went down at some point, we cannot bump the active state, try to regain it
				active, err := s.tryActive(ctx)
				if err == nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "redis connection re-established", "active", active)
					if !active {
						// activity was stolen when redis came back. This is possible as the secondary will restart
						log.FatalLog("activity lost when redis connection re-established")
					}
					elapsedSinceBumpActive = time.Since(time.Now())
					s.RedisConnectionFailed = false
				}
				if err != nil && elapsedSinceLog >= CheckActiveLogInterval {
					log.SpanLog(ctx, log.DebugLevelInfra, "error in tryActive", "err", err)
				}
			}
			if !s.RedisConnectionFailed {
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
					stillActive, err := s.CheckActive(ctx)
					if !stillActive {
						// somehow we lost activity
						if err != nil {
							if s.HARole == string(process.HARolePrimary) {
								log.SpanLog(ctx, log.DebugLevelInfra, "Maintaining active status for primary due to redis error", "err", err)
								s.RedisConnectionFailed = true
								s.nodeMgr.Event(ctx, "Redis error", s.nodeMgr.MyNode.Key.CloudletKey.Organization, s.nodeMgr.MyNode.Key.CloudletKey.GetTags(), err, "Node Type", s.nodeMgr.MyNode.Key.Type, "HA Role", s.HARole, "Redis Addr", s.redisAddr)
								continue
							} else {
								log.FatalLog("secondary unit lost activity due to redis error")
							}
						} else {
							// this is unexpected
							log.SpanLog(ctx, log.DebugLevelInfra, "Activity Lost Unexpectedly")
						}
					}
					timeLastCheckActive = time.Now()
				}
				err := s.BumpActiveExpire(ctx)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "BumpActiveExpire failed, retry", "err", err)
					err = s.BumpActiveExpire(ctx)
					if err != nil {
						if s.HARole == string(process.HARolePrimary) {
							log.SpanLog(ctx, log.DebugLevelInfra, "Maintaining active status for primary due to redis error", "err", err)
							s.RedisConnectionFailed = true
							s.nodeMgr.Event(ctx, "Redis error", s.nodeMgr.MyNode.Key.CloudletKey.Organization, s.nodeMgr.MyNode.Key.CloudletKey.GetTags(), err, "Node Type", s.nodeMgr.MyNode.Key.Type, "HA Role", s.HARole, "Redis Addr", s.redisAddr)
						} else {
							log.FatalLog("BumpActiveExpire failed!", "err", err)
						}
					}
				}
				timeLastBumpActive = time.Now()
			}
		}
		time.Sleep(s.activePollInterval)
	}
}

func (s *HighAvailabilityManager) DumpActive(ctx context.Context, req *edgeproto.DebugRequest) string {
	return fmt.Sprintf("PlatformActive: %t", s.PlatformInstanceActive)
}
