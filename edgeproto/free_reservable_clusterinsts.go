package edgeproto

import (
	"sync"

	context "golang.org/x/net/context"
)

type AppCloudletKey struct {
	AppKey      AppKey
	CloudletKey CloudletKey
}

type FreeReservableClusterInstCache struct {
	InstsByCloudlet map[CloudletKey]map[ClusterInstKey]*ClusterInst
	Mux             sync.Mutex
}

func (s *FreeReservableClusterInstCache) Init() {
	s.InstsByCloudlet = make(map[CloudletKey]map[ClusterInstKey]*ClusterInst)
}

func (s *FreeReservableClusterInstCache) Update(ctx context.Context, in *ClusterInst, rev int64) {
	if !in.Reservable {
		return
	}
	s.Mux.Lock()
	defer s.Mux.Unlock()
	cinsts, found := s.InstsByCloudlet[in.Key.CloudletKey]
	if !found {
		cinsts = make(map[ClusterInstKey]*ClusterInst)
		s.InstsByCloudlet[in.Key.CloudletKey] = cinsts
	}
	if in.ReservedBy != "" {
		delete(cinsts, in.Key)
	} else {
		cinsts[in.Key] = in
	}
}

func (s *FreeReservableClusterInstCache) Delete(ctx context.Context, in *ClusterInst, rev int64) {
	if !in.Reservable {
		return
	}
	s.Mux.Lock()
	defer s.Mux.Unlock()
	cinsts, found := s.InstsByCloudlet[in.Key.CloudletKey]
	if !found {
		return
	}
	delete(cinsts, in.Key)
	if len(cinsts) == 0 {
		delete(s.InstsByCloudlet, in.Key.CloudletKey)
	}
}

func (s *FreeReservableClusterInstCache) Prune(ctx context.Context, keys map[ClusterInstKey]struct{}) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for key, _ := range keys {
		cinsts, found := s.InstsByCloudlet[key.CloudletKey]
		if !found {
			continue
		}
		delete(cinsts, key)
		if len(cinsts) == 0 {
			delete(s.InstsByCloudlet, key.CloudletKey)
		}
	}
}

func (s *FreeReservableClusterInstCache) Flush(ctx context.Context, notifyId int64) {}

func (s *FreeReservableClusterInstCache) GetForCloudlet(key *CloudletKey) *ClusterInstKey {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	cinsts, found := s.InstsByCloudlet[*key]
	if found && len(cinsts) > 0 {
		for key, _ := range cinsts {
			return &key
		}
	}
	return nil
}
