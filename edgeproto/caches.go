package edgeproto

import "github.com/mobiledgex/edge-cloud/log"

// Common extra support code for caches

type CacheUpdateType int

const (
	UpdateTask CacheUpdateType = 0
	UpdateStep CacheUpdateType = 1
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
func (s *AppInstCache) GetForCloudlet(key *CloudletKey, appInsts map[AppInstKey]struct{}) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for k, v := range s.Objs {
		if v.Key.ClusterInstKey.CloudletKey == *key {
			appInsts[k] = struct{}{}
		}
	}
}

// GetForCloudlet finds all ClusterInsts associated with the
// given cloudlets
func (s *ClusterInstCache) GetForCloudlet(key *CloudletKey, clusterInsts map[ClusterInstKey]struct{}) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for k, v := range s.Objs {
		if v.Key.CloudletKey == *key {
			clusterInsts[k] = struct{}{}
		}
	}
}

func (s *ClusterInstInfoCache) SetState(key *ClusterInstKey, state TrackedState) {
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = nil
	info.State = state
	info.Status = StatusInfo{}
	s.Update(&info, 0)
}

func (s *ClusterInstInfoCache) SetStatusTask(key *ClusterInstKey, taskName string) {
	log.DebugLog(log.DebugLevelApi, "SetStatusTask", "key", key, "taskName", taskName)
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusTask failed, did not find clusterInst in cache")
		return
	}
	info.Status.setTask(taskName)
	s.Update(&info, 0)
}

func (s *ClusterInstInfoCache) SetStatusMaxTasks(key *ClusterInstKey, maxTasks uint32) {
	log.DebugLog(log.DebugLevelApi, "SetStatusMaxTasks", "key", key, "maxTasks", maxTasks)
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusMaxTasks failed, did not find clusterInst in cache")
		return
	}
	info.Status.setMaxTasks(maxTasks)
	s.Update(&info, 0)
}

func (s *ClusterInstInfoCache) SetStatusStep(key *ClusterInstKey, stepName string) {
	log.DebugLog(log.DebugLevelApi, "SetStatusStep", "key", key, "stepName", stepName)
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusStep failed, did not find clusterInst in cache")
		return
	}
	info.Status.setStep(stepName)
	s.Update(&info, 0)
}

func (s *ClusterInstInfoCache) SetError(key *ClusterInstKey, errState TrackedState, err string) {
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = append(info.Errors, err)
	info.State = errState
	s.Update(&info, 0)
}

func (s *AppInstInfoCache) SetState(key *AppInstKey, state TrackedState) {
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = nil
	info.State = state
	info.Status = StatusInfo{}
	s.Update(&info, 0)
}

func (s *AppInstInfoCache) SetStateRuntime(key *AppInstKey, state TrackedState, rt *AppInstRuntime) {
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = nil
	info.State = state
	info.Status = StatusInfo{}
	info.RuntimeInfo = *rt
	s.Update(&info, 0)
}

func (s *AppInstInfoCache) SetStatusMaxTasks(key *AppInstKey, maxTasks uint32) {
	log.DebugLog(log.DebugLevelApi, "SetStatusMaxTasks", "key", key, "maxTasks", maxTasks)
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusTaskMax failed, did not find clusterInst in cache")
		return
	}
	info.Status.setMaxTasks(maxTasks)
	s.Update(&info, 0)
}

func (s *AppInstInfoCache) SetStatusTask(key *AppInstKey, taskName string) {
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusTask failed, did not find clusterInst in cache")
		return
	}
	info.Status.setTask(taskName)
	s.Update(&info, 0)
}

func (s *AppInstInfoCache) SetStatusStep(key *AppInstKey, stepName string) {
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusStep failed, did not find clusterInst in cache")
		return
	}
	info.Status.setStep(stepName)
	s.Update(&info, 0)
}

func (s *AppInstInfoCache) SetError(key *AppInstKey, errState TrackedState, err string) {
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = append(info.Errors, err)
	info.State = errState
	s.Update(&info, 0)
}
