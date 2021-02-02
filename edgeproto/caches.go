package edgeproto

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/log"
	context "golang.org/x/net/context"
)

// Common extra support code for caches

type CacheUpdateType int

const (
	UpdateTask CacheUpdateType = 0
	UpdateStep CacheUpdateType = 1

	ResetStatus   bool = true
	NoResetStatus bool = false
)

type ClusterInstCacheUpdateParms struct {
	cache      *ClusterInstInfoCache
	updateType CacheUpdateType
	value      string
}

// CacheUpdateCallback updates either state or task with the given value
type CacheUpdateCallback func(updateType CacheUpdateType, value string)

// DummyUpdateCallback is used when we don't want any cache status updates
func DummyUpdateCallback(updateType CacheUpdateType, value string) {}

// GetAppInstsForCloudlets finds all AppInsts associated with the given cloudlets
func (s *AppInstCache) GetForCloudlet(key *CloudletKey, cb func(appInstKey *AppInstKey, modRev int64)) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for k, v := range s.Objs {
		if v.Obj.Key.ClusterInstKey.CloudletKey == *key {
			cb(&k, v.ModRev)
		}
	}
}

func (s *AppInstCache) GetForRealClusterInstKey(key *ClusterInstKey, cb func(appInst *AppInst)) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for _, v := range s.Objs {
		obj := v.Obj
		clusterInstKey := obj.ClusterInstKey()
		if !key.Matches(clusterInstKey) {
			continue
		}
		cb(obj)
	}
}

// GetForCloudlet finds all ClusterInsts associated with the
// given cloudlets
func (s *ClusterInstCache) GetForCloudlet(key *CloudletKey, cb func(clusterInstKey *ClusterInstKey, modRev int64)) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for k, v := range s.Objs {
		if v.Obj.Key.CloudletKey == *key {
			cb(&k, v.ModRev)
		}
	}
}

func (s *ClusterInstInfoCache) SetState(ctx context.Context, key *ClusterInstKey, state TrackedState) error {
	var err error
	s.UpdateModFunc(ctx, key, 0, func(old *ClusterInstInfo) (newObj *ClusterInstInfo, changed bool) {
		info := &ClusterInstInfo{}
		if old == nil {
			info.Key = *key
			info.Status = StatusInfo{}
		} else {
			err = StateConflict(old.State, state)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelApi, "SetState conflict", "oldState", old.State, "newState", state, "err", err)
				return old, false
			}
			*info = *old
		}
		info.Errors = nil
		info.State = state
		return info, true
	})
	return err
}

func (s *ClusterInstInfoCache) SetResources(ctx context.Context, key *ClusterInstKey, resources *InfraResources) error {
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		log.SpanLog(ctx, log.DebugLevelApi, "SetResources failed, did not find clusterInst in cache")
		return fmt.Errorf("ClusterInst not found in cache: %s", key.String())
	}
	info.Resources = *resources
	s.Update(ctx, &info, 0)
	return nil
}

func (s *ClusterInstInfoCache) SetStatusTask(ctx context.Context, key *ClusterInstKey, taskName string, resetStatus bool) {
	log.SpanLog(ctx, log.DebugLevelApi, "SetStatusTask", "key", key, "taskName", taskName)
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusTask failed, did not find clusterInst in cache")
		return
	}
	if resetStatus {
		info.Status.StatusReset()
	}
	info.Status.SetTask(taskName)
	s.Update(ctx, &info, 0)
}

func (s *ClusterInstInfoCache) SetStatusMaxTasks(ctx context.Context, key *ClusterInstKey, maxTasks uint32) {
	log.SpanLog(ctx, log.DebugLevelApi, "SetStatusMaxTasks", "key", key, "maxTasks", maxTasks)
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusMaxTasks failed, did not find clusterInst in cache")
		return
	}
	info.Status.SetMaxTasks(maxTasks)
	s.Update(ctx, &info, 0)
}

func (s *ClusterInstInfoCache) SetStatusStep(ctx context.Context, key *ClusterInstKey, stepName string, resetStatus bool) {
	log.SpanLog(ctx, log.DebugLevelApi, "SetStatusStep", "key", key, "stepName", stepName)
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusStep failed, did not find clusterInst in cache")
		return
	}
	if resetStatus {
		info.Status.StatusReset()
	}
	info.Status.SetStep(stepName)
	s.Update(ctx, &info, 0)
}

func (s *ClusterInstInfoCache) SetError(ctx context.Context, key *ClusterInstKey, errState TrackedState, err string) {
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = append(info.Errors, err)
	info.State = errState
	s.Update(ctx, &info, 0)
}

// If CRM crashes or reconnects to controller, controller will resend
// current state. This is needed to:
// -restart actions that were lost due to a crash
// -update cache for dependent objects (AppInst looks up ClusterInst from
// cache).
// If it was a disconnect and not a restart, there may already be a
// thread in progress. To prevent multiple conflicting threads, check
// the state which can tell us if a thread is in progress.
func StateConflict(oldState, newState TrackedState) error {
	busyStates := []TrackedState{
		TrackedState_CREATING,
		TrackedState_UPDATING,
		TrackedState_DELETING,
	}

	oldBusy := false
	newBusy := false
	for _, state := range busyStates {
		if oldState == state {
			oldBusy = true
		}
		if newState == state {
			newBusy = true
		}
	}
	if oldBusy && newBusy {
		return fmt.Errorf("conflicting state: %s", oldState)
	}
	return nil
}

func PowerStateConflict(oldState, newState PowerState) error {
	busyStates := []PowerState{
		PowerState_POWERING_ON,
		PowerState_POWERING_OFF,
		PowerState_REBOOTING,
	}

	oldBusy := false
	newBusy := false
	for _, state := range busyStates {
		if oldState == state {
			oldBusy = true
		}
		if newState == state {
			newBusy = true
		}
	}
	if oldBusy && newBusy {
		return fmt.Errorf("conflicting state: %s", oldState)
	}
	return nil
}

func IsTransientState(state TrackedState) bool {
	if state == TrackedState_CREATING ||
		state == TrackedState_CREATING_DEPENDENCIES ||
		state == TrackedState_CREATE_REQUESTED ||
		state == TrackedState_UPDATE_REQUESTED ||
		state == TrackedState_DELETE_REQUESTED ||
		state == TrackedState_UPDATING ||
		state == TrackedState_DELETING ||
		state == TrackedState_DELETE_PREPARE {
		return true
	}
	return false
}

type PowerStateType int

const (
	RequestState   PowerStateType = 0
	TransientState PowerStateType = 1
	FinalState     PowerStateType = 2
)

func GetNextPowerState(state PowerState, stateType PowerStateType) PowerState {
	switch stateType {
	case RequestState:
		if state == PowerState_POWER_ON {
			return PowerState_POWER_ON_REQUESTED
		} else if state == PowerState_POWER_OFF {
			return PowerState_POWER_OFF_REQUESTED
		} else if state == PowerState_REBOOT {
			return PowerState_REBOOT_REQUESTED
		}
	case TransientState:
		if state == PowerState_POWER_ON_REQUESTED {
			return PowerState_POWERING_ON
		} else if state == PowerState_POWER_OFF_REQUESTED {
			return PowerState_POWERING_OFF
		} else if state == PowerState_REBOOT_REQUESTED {
			return PowerState_REBOOTING
		}
	case FinalState:
		if state == PowerState_POWERING_ON {
			return PowerState_POWER_ON
		} else if state == PowerState_POWERING_OFF {
			return PowerState_POWER_OFF
		} else if state == PowerState_REBOOTING {
			return PowerState_POWER_ON
		}
	}
	return PowerState_POWER_STATE_UNKNOWN
}

func (s *AppInstInfoCache) SetPowerState(ctx context.Context, key *AppInstKey, state PowerState) error {
	var err error
	s.UpdateModFunc(ctx, key, 0, func(old *AppInstInfo) (newObj *AppInstInfo, changed bool) {
		info := &AppInstInfo{}
		if old == nil {
			info.Key = *key
			info.Status = StatusInfo{}
		} else {
			err = PowerStateConflict(old.PowerState, state)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelApi, "SetPowerState conflict", "oldState", old.PowerState, "newState", state, "err", err)
				return old, false
			}
			*info = *old
		}
		info.Errors = nil
		info.PowerState = state
		return info, true
	})
	return err
}

func (s *AppInstInfoCache) SetState(ctx context.Context, key *AppInstKey, state TrackedState) error {
	var err error
	s.UpdateModFunc(ctx, key, 0, func(old *AppInstInfo) (newObj *AppInstInfo, changed bool) {
		info := &AppInstInfo{}
		if old == nil {
			info.Key = *key
			info.Status = StatusInfo{}
		} else {
			err = StateConflict(old.State, state)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelApi, "SetState conflict", "oldState", old.State, "newState", state, "err", err)
				return old, false
			}
			*info = *old
		}
		info.Errors = nil
		info.State = state
		return info, true
	})
	return err
}

func (s *AppInstInfoCache) SetStateRuntime(ctx context.Context, key *AppInstKey, state TrackedState, rt *AppInstRuntime) {
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
		info.Status = StatusInfo{}
	}
	info.Errors = nil
	info.State = state
	info.RuntimeInfo = *rt
	s.Update(ctx, &info, 0)
}

func (s *AppInstInfoCache) SetStatusMaxTasks(ctx context.Context, key *AppInstKey, maxTasks uint32) {
	log.SpanLog(ctx, log.DebugLevelApi, "SetStatusMaxTasks", "key", key, "maxTasks", maxTasks)
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusTaskMax failed, did not find appInstInfo in cache")
		return
	}
	info.Status.SetMaxTasks(maxTasks)
	s.Update(ctx, &info, 0)
}

func (s *AppInstInfoCache) SetStatusTask(ctx context.Context, key *AppInstKey, taskName string, resetStatus bool) {
	log.SpanLog(ctx, log.DebugLevelApi, "SetStatusTask", "key", key, "taskName", taskName)
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusTask failed, did not find appInstInfo in cache")
		return
	}
	if resetStatus {
		info.Status.StatusReset()
	}
	info.Status.SetTask(taskName)
	s.Update(ctx, &info, 0)
}

func (s *AppInstInfoCache) SetStatusStep(ctx context.Context, key *AppInstKey, stepName string, resetStatus bool) {
	log.SpanLog(ctx, log.DebugLevelApi, "SetStatusStep", "key", key, "stepName", stepName)
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusStep failed, did not find appInstInfo in cache")
		return
	}
	if resetStatus {
		info.Status.StatusReset()
	}
	info.Status.SetStep(stepName)
	s.Update(ctx, &info, 0)
}

func (s *AppInstInfoCache) SetError(ctx context.Context, key *AppInstKey, errState TrackedState, err string) {
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = append(info.Errors, err)
	info.State = errState
	s.Update(ctx, &info, 0)
}

func (s *CloudletInfoCache) SetStatusTask(ctx context.Context, key *CloudletKey, taskName string) {
	log.SpanLog(ctx, log.DebugLevelApi, "SetStatusTask", "key", key, "taskName", taskName)
	info := CloudletInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusTask failed, did not find CloudletInfo in cache")
		return
	}
	info.Status.SetTask(taskName)
	s.Update(ctx, &info, 0)
}

func (s *CloudletInfoCache) SetStatusStep(ctx context.Context, key *CloudletKey, stepName string) {
	log.SpanLog(ctx, log.DebugLevelApi, "SetStatusStep", "key", key, "stepName", stepName)
	info := CloudletInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusStep failed, did not find CloudletInfo in cache")
		return
	}
	info.Status.SetStep(stepName)
	s.Update(ctx, &info, 0)
}
