package edgeproto

import (
	"sync"

	"github.com/mobiledgex/edge-cloud/log"
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

func (s *FreeReservableClusterInstCache) Prune(ctx context.Context, validKeys map[ClusterInstKey]struct{}) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for cloudletKey, cmap := range s.InstsByCloudlet {
		for clusterInstKey, _ := range cmap {
			if _, ok := validKeys[clusterInstKey]; !ok {
				delete(cmap, clusterInstKey)
			}
		}
		if len(cmap) == 0 {
			delete(s.InstsByCloudlet, cloudletKey)
		}
	}
}

func (s *FreeReservableClusterInstCache) Flush(ctx context.Context, notifyId int64) {}

func (s *FreeReservableClusterInstCache) GetForCloudlet(key *CloudletKey, deployment string) *ClusterInstKey {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	cinsts, found := s.InstsByCloudlet[*key]
	log.DebugLog(log.DebugLevelDmereq, "GetForCloudlet", "key", *key, "found", found, "num-insts", len(cinsts))
	if found && len(cinsts) > 0 {
		for key, clust := range cinsts {
			if deployment == clust.Deployment {
				return &key
			}
		}
	}
	return nil
}

func (s *FreeReservableClusterInstCache) GetCount() int {
	count := 0
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for _, m := range s.InstsByCloudlet {
		count += len(m)
	}
	return count
}

func (s *FreeReservableClusterInstCache) GetTypeString() string {
	return "FreeReservableClusterInstCache"
}
