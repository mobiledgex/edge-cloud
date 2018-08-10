package edgeproto

// Common extra support code for caches

// GetAppInstsForCloudlets finds all AppInsts associated with the given cloudlets
func (s *AppInstCache) GetAppInstsForCloudlets(cloudlets map[CloudletKey]struct{}, appInsts map[AppInstKey]struct{}) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for k, v := range s.Objs {
		if _, found := cloudlets[v.Key.CloudletKey]; found {
			appInsts[k] = struct{}{}
		}
	}
}

// GetClusterInstsForCloudlets finds all ClusterInsts associated with the
// given cloudlets
func (s *ClusterInstCache) GetClusterInstsForCloudlets(cloudlets map[CloudletKey]struct{}, clusterInsts map[ClusterInstKey]struct{}) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for k, v := range s.Objs {
		if _, found := cloudlets[v.Key.CloudletKey]; found {
			clusterInsts[k] = struct{}{}
		}
	}
}

func (s *ClusterInstInfoCache) SetState(key *ClusterInstKey, state ClusterState) {
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = nil
	info.State = state
	s.Update(&info, 0)
}

func (s *ClusterInstInfoCache) SetError(key *ClusterInstKey, err string) {
	info := ClusterInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = append(info.Errors, err)
	info.State = ClusterState_ClusterStateErrors
	s.Update(&info, 0)
}

func (s *AppInstInfoCache) SetState(key *AppInstKey, state AppState) {
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = nil
	info.State = state
	s.Update(&info, 0)
}

func (s *AppInstInfoCache) SetError(key *AppInstKey, err string) {
	info := AppInstInfo{}
	if !s.Get(key, &info) {
		info.Key = *key
	}
	info.Errors = append(info.Errors, err)
	info.State = AppState_AppStateErrors
	s.Update(&info, 0)
}
