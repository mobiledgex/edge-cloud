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
	"github.com/mobiledgex/edge-cloud/rediscache"
)

const RedisPingFail string = "Redis Ping Fail"
const HighAvailabilityManagerDisabled = "HighAvailabilityManagerDisabled"

const CheckActiveLogInterval = time.Second * 10
const MaxRedisUnreachableRetries = 3

type HAWatcher interface {
	ActiveChangedPreSwitch(ctx context.Context, platformActive bool) error  // actions before setting PlatformInstanceActive
	ActiveChangedPostSwitch(ctx context.Context, platformActive bool) error // actions after setting PlatformInstanceActive
	PlatformActiveOnStartup(ctx context.Context)                            // actions if the platform is active on first
}

type HighAvailabilityManager struct {
	redisCfg                   rediscache.RedisConfig
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
	s.redisCfg.InitFlags(rediscache.DefaultCfgRedisOptional)
	flag.StringVar(&s.HARole, "HARole", string(process.HARolePrimary), string(process.HARolePrimary+" or "+process.HARoleSecondary))
}

func (s *HighAvailabilityManager) Init(ctx context.Context, nodeGroupKey string, nodeMgr *node.NodeMgr, activeDuration, activePollInterval edgeproto.Duration, haWatcher HAWatcher) error {
	log.SpanLog(ctx, log.DebugLevelInfo, "HighAvailabilityManager init", "nodeGroupKey", nodeGroupKey, "redisCfg", s.redisCfg, "role", s.HARole, "activeDuration", activeDuration, "activePollInterval", activePollInterval)
	s.activeDuration = activeDuration.TimeDuration()
	s.activePollInterval = activePollInterval.TimeDuration()
	s.nodeMgr = nodeMgr
	s.haWatcher = haWatcher
	if s.HARole != string(process.HARolePrimary) && s.HARole != string(process.HARoleSecondary) {
		return fmt.Errorf("invalid HA Role type")
	}
	defer func() {
		if s.PlatformInstanceActive {
			// perform any actions needed when the platform is active on start
			s.haWatcher.PlatformActiveOnStartup(ctx)
		}
	}()
	if !s.redisCfg.AddrSpecified() {
		s.PlatformInstanceActive = true
		return fmt.Errorf("%s Redis Addr for HA not specified", HighAvailabilityManagerDisabled)
	}
	s.nodeGroupKey = nodeGroupKey
	if s.nodeGroupKey == "" {
		return fmt.Errorf("group key node not specified")
	}
	s.HAEnabled = true

	err := s.connectRedis(ctx)
	if err != nil {
		s.updateRedisFailed(ctx, true, err)
		log.SpanLog(ctx, log.DebugLevelInfo, "connectRedis failed", "err", err)
		if s.HARole == string(process.HARolePrimary) {
			log.SpanLog(ctx, log.DebugLevelInfo, "Assuming active status for primary node due to redis failure")
			s.PlatformInstanceActive = true
			return nil
		} else {
			log.SpanLog(ctx, log.DebugLevelInfo, "Assuming inactive status for secondary node due to redis failure")
		}
	} else {
		log.SpanLog(ctx, log.DebugLevelInfo, "redis is online, do tryActive")
		s.PlatformInstanceActive, err = s.tryActive(ctx)
		return err
	}
	return nil
}

func (s *HighAvailabilityManager) pingRedis(ctx context.Context, genLog bool) error {
	pong, err := s.redisClient.Ping().Result()
	if genLog {
		log.SpanLog(ctx, log.DebugLevelInfra, "redis ping done", "pong", pong, "err", err)
	}
	if err != nil {
		return fmt.Errorf("%s - %v", RedisPingFail, err)
	}
	return nil
}

func (s *HighAvailabilityManager) connectRedis(ctx context.Context) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "connectRedis")
	var err error
	s.redisClient, err = rediscache.NewClient(ctx, &s.redisCfg)
	if err != nil {
		return err
	}

	err = rediscache.IsServerReady(s.redisClient, rediscache.MaxRedisWait)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "pingRedis failed", "err", err)
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "redis ping done successfully")
	return nil
}

// TryActive is called on startup
func (s *HighAvailabilityManager) tryActive(ctx context.Context) (bool, error) {
	if !s.RedisConnectionFailed {
		if s.PlatformInstanceActive || s.ActiveTransitionInProgress {
			// this should not happen. Only 1 thread should be doing tryActive
			log.SpanFromContext(ctx).Finish()
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
	v, err := s.redisClient.SetNX(s.nodeGroupKey, s.HARole, s.activeDuration).Result()
	if err != nil {
		// don't log this if the redis is already down because it result in excessive logs
		if !s.RedisConnectionFailed {
			log.SpanLog(ctx, log.DebugLevelInfra, "tryActive setNX error", "key", s.nodeGroupKey, "v", v, "err", err)
		}
		return false, err
	}
	return v, nil
}

func (s *HighAvailabilityManager) SetValue(ctx context.Context, key string, value string, expiration time.Duration) error {
	result, err := s.redisClient.Set(key, value, expiration).Result()
	log.SpanLog(ctx, log.DebugLevelInfra, "SetValue Done", "expiration", expiration, "result", result, "err", err)
	return err
}

func (s *HighAvailabilityManager) GetValue(ctx context.Context, key string) (string, error) {
	val, err := s.redisClient.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("Error getting value from redis: %v", err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetValue Done", "val", val)
	return val, nil
}

func (s *HighAvailabilityManager) CheckActive(ctx context.Context) (bool, error) {

	v, err := s.redisClient.Get(s.nodeGroupKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "CheckActive - key does not exist -- neither unit is active")
			return false, nil
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "CheckActive error", "key", s.nodeGroupKey, "v", v, "err", err)
			return false, err
		}
	}
	isActive := v == s.HARole
	return isActive, nil
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

func (s *HighAvailabilityManager) updateRedisFailed(ctx context.Context, newRedisFailed bool, err error) {
	if newRedisFailed == s.RedisConnectionFailed {
		// no change
		return
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "updateRedisFailed state changed", "current RedisConnectionFailed", s.RedisConnectionFailed, "newRedisFailed", newRedisFailed, "err", err)
	// generate an event if the state changed
	s.RedisConnectionFailed = newRedisFailed
	if newRedisFailed {
		s.nodeMgr.Event(ctx, "Redis offline", s.nodeMgr.MyNode.Key.CloudletKey.Organization, s.nodeMgr.MyNode.Key.CloudletKey.GetTags(), err, "Node Type", s.nodeMgr.MyNode.Key.Type, "HARole", s.HARole)
	} else {
		s.nodeMgr.Event(ctx, "Redis online", s.nodeMgr.MyNode.Key.CloudletKey.Organization, s.nodeMgr.MyNode.Key.CloudletKey.GetTags(), err, "Node Type", s.nodeMgr.MyNode.Key.Type, "HARole", s.HARole)
	}
}

func (s *HighAvailabilityManager) CheckActiveLoop(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfra, "Enter CheckActiveLoop", "activePollInterval", s.activePollInterval, "activeDuration", s.activeDuration)
	timeLastLog := time.Now() // log only once every X seconds in this loop
	timeLastBumpActive := time.Now()
	timeLastCheckActive := time.Now()
	for {
		elapsedSinceLog := time.Since(timeLastLog)
		elapsedSinceBumpActive := time.Since(timeLastBumpActive)
		elapsedSinceCheckActive := time.Since(timeLastCheckActive)
		periodicLog := elapsedSinceLog >= CheckActiveLogInterval
		isActive := s.PlatformInstanceActive || s.ActiveTransitionInProgress // either of these states indicate active from a redis standpoint
		if periodicLog {
			timeLastLog = time.Now()
			log.SpanLog(ctx, log.DebugLevelInfra, "CheckActiveLoop", "role", s.HARole, "isActive", isActive, "RedisConnectionFailed", s.RedisConnectionFailed, "ActiveTransitionInProgress", s.ActiveTransitionInProgress, "PlatformInstanceActive", s.PlatformInstanceActive)
		}
		if !isActive {
			var err error
			transitionToActive := false
			if s.RedisConnectionFailed && s.HARole == string(process.HARoleSecondary) {
				// redis is already known to be down on this pass
				if periodicLog {
					log.SpanLog(ctx, log.DebugLevelInfra, "redis unreachable, ping to see if it came back")
				}
				// rather than tryActive which will seize activity, just ping redis to see if it came back. We don't want
				// to take activity from the primary when redis comes back online
				err = s.pingRedis(ctx, periodicLog)
				if err == nil {
					s.updateRedisFailed(ctx, false, err)
					// this secondary is currently standby and redis just came back online, so the primary should be active.
					// Sleep two intervals and wait until the next pass before doing tryActive again to give
					// the primary the chance to retain activity
					log.SpanLog(ctx, log.DebugLevelInfra, "redis reachable again, wait and then tryActive")
				}
				time.Sleep(s.activePollInterval * 2)
				continue
			}
			for retry := 0; retry < MaxRedisUnreachableRetries; retry++ {
				transitionToActive, err = s.tryActive(ctx)
				if err == nil {
					break
				}
				if periodicLog {
					log.SpanLog(ctx, log.DebugLevelInfra, "redis unreachable, sleep and retry tryActive", "retry", retry)
				}
				time.Sleep(s.activePollInterval)
			}
			if err == nil {
				s.updateRedisFailed(ctx, false, nil)
			} else {
				s.updateRedisFailed(ctx, true, err)
				// Redis unreachable. Assume if it is a network or redis outage, it affects both primary and secondary.
				// If Redis itself is down, both primary and secondary will fail to reach it.
				// In this case, primary should go active, and secondary should go standby.
				// We are already not active here.
				if s.HARole == string(process.HARolePrimary) {
					// go active
					log.SpanLog(ctx, log.DebugLevelInfra, "Primary platform became active due to redis failure", "role", s.HARole)
					transitionToActive = true
				} else {
					if periodicLog {
						log.SpanLog(ctx, log.DebugLevelInfra, "Secondary remains inactive", "role", s.HARole)
					}
				}
			}
			if transitionToActive {
				log.SpanLog(ctx, log.DebugLevelInfra, "Platform becoming active", "role", s.HARole, "RedisConnectionFailed", s.RedisConnectionFailed)
				timeLastBumpActive = time.Now()
				// switchover is handled in a separate thread so we do not miss polling redis
				s.ActiveTransitionInProgress = true
				switchoverStartTime := time.Now()
				err := s.haWatcher.ActiveChangedPreSwitch(ctx, true)
				if err != nil {
					log.SpanFromContext(ctx).Finish()
					log.FatalLog("ActiveChangedPreSwitch failed", "err", err)
				}
				s.PlatformInstanceActive = true
				s.ActiveTransitionInProgress = false
				s.haWatcher.ActiveChangedPostSwitch(ctx, true)
				if err != nil {
					log.SpanFromContext(ctx).Finish()
					log.FatalLog("ActiveChangedPostSwitch failed", "err", err)
				}
				s.nodeMgr.Event(ctx, "High Availability Node Active", s.nodeMgr.MyNode.Key.CloudletKey.Organization, s.nodeMgr.MyNode.Key.CloudletKey.GetTags(), nil, "Node Type", s.nodeMgr.MyNode.Key.Type, "HARole", s.HARole)
				switchoverDuration := time.Since(switchoverStartTime)
				log.SpanLog(ctx, log.DebugLevelInfra, "switchover done", "switchoverDuration", switchoverDuration)
				if switchoverDuration > s.activePollInterval {
					// indicates some long running task was done by the watcher in ActiveChangedPreSwitch or ActiveChangedPostSwitch
					log.SpanLog(ctx, log.DebugLevelInfra, "Warning: switchover took excessive time", "switchoverDuration", switchoverDuration)
				}
			} else {
				time.Sleep(s.activePollInterval)
				continue
			}
		} else { // platform active case
			if s.RedisConnectionFailed {
				// redis went down at some point, we cannot bump the active state, try to regain it
				active, err := s.tryActive(ctx)
				if err == nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "redis connection re-established", "active", active)
					s.updateRedisFailed(ctx, false, nil)
					if !active {
						// activity was stolen when redis came back. This is possible but unlikely as the secondary should give the primary time to remain active
						log.SpanLog(ctx, log.DebugLevelInfra, "activity lost when redis connection re-established", "role", s.HARole)
						s.PlatformInstanceActive = false
						continue
					}
					elapsedSinceBumpActive = time.Since(time.Now())
				} else {
					if periodicLog {
						log.SpanLog(ctx, log.DebugLevelInfra, "error in tryActive", "err", err)
					}
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
							s.updateRedisFailed(ctx, true, err)
							if s.HARole == string(process.HARolePrimary) {
								log.SpanLog(ctx, log.DebugLevelInfra, "Maintaining active status for primary due to redis error", "err", err)
							} else {
								s.PlatformInstanceActive = false
								log.SpanLog(ctx, log.DebugLevelInfra, "secondary unit lost activity due to redis error")
							}
							time.Sleep(s.activePollInterval)
							continue // bypass bumpActive this pass since redis is down
						} else {
							// this is unexpected
							log.SpanLog(ctx, log.DebugLevelInfra, "Activity Lost Unexpectedly")
							s.PlatformInstanceActive = false
						}
					}
					timeLastCheckActive = time.Now()
				}
				var err error
				for retry := 0; retry < MaxRedisUnreachableRetries; retry++ {
					err = s.BumpActiveExpire(ctx)
					if err == nil {
						timeLastBumpActive = time.Now()
						break
					}
					log.SpanLog(ctx, log.DebugLevelInfra, "BumpActiveExpire failed", "retry", retry, "err", err)
				}
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "BumpActiveExpire failed, retries exhausted", "err", err)
					s.updateRedisFailed(ctx, true, err)
					if s.HARole == string(process.HARolePrimary) {
						log.SpanLog(ctx, log.DebugLevelInfra, "Maintaining active status for primary due to redis error", "err", err)
					} else {
						s.PlatformInstanceActive = false
						log.SpanLog(ctx, log.DebugLevelInfra, "Standby going inactive due to redis error", "err", err)
					}
				}
			}
		}
		time.Sleep(s.activePollInterval)
	}
}

func (s *HighAvailabilityManager) DumpActive(ctx context.Context, req *edgeproto.DebugRequest) string {
	return fmt.Sprintf("PlatformActive: %t", s.PlatformInstanceActive)
}
