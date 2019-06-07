package edgeproto

import "github.com/mobiledgex/edge-cloud/log"

// Common extra support code for caches

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

// GetClusterInstsForCloudlets finds all ClusterInsts associated with the
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
	info.Status = &StatusInfo{}
	log.InfoLog("ClusterInstInfoCache SetState", "state", state)
	s.Update(&info, 0)
}

func (s *ClusterInstInfoCache) SetStatusTask(key *ClusterInstKey, taskName string) {
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusTask failed, did not find clusterInst in cache")
		return
	}
	info.Status.setTask(taskName)
	s.Update(&info, 0)
}

func (s *ClusterInstInfoCache) SetStatusStep(key *ClusterInstKey, stepName string) {
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		// we don't want to override the state in the cache if it is not present
		log.InfoLog("SetStatusStep failed, did not find clusterInst in cache")
		return
	}

	info.Status.setStep(stepName)
	s.Update(&info, 0)
}

func (s *ClusterInstInfoCache) ResetStatus(key *ClusterInstKey) {
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		// nothing to do
		return
	}
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
	info.Status = &StatusInfo{}
	s.Update(&info, 0)
}

func (s *AppInstInfoCache) SetStateRuntime(key *AppInstKey, state TrackedState, rt *AppInstRuntime) {
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = nil
	info.State = state
	info.RuntimeInfo = *rt
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

func (s *AppInstInfoCache) ResetStatus(key *AppInstKey) {
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		// nothing to do
		return
	}
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
