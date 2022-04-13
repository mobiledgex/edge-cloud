// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func (s *FreeReservableClusterInstCache) GetForCloudlet(key *CloudletKey, deployment, flavor string, deploymentTransformFunc func(string) string) *ClusterInstKey {
	// need a transform func to avoid import cycle
	deployment = deploymentTransformFunc(deployment)
	s.Mux.Lock()
	defer s.Mux.Unlock()
	cinsts, found := s.InstsByCloudlet[*key]
	log.DebugLog(log.DebugLevelDmereq, "GetForCloudlet", "key", *key, "found", found, "num-insts", len(cinsts))
	if found && len(cinsts) > 0 {
		for key, clust := range cinsts {
			if deployment == clust.Deployment && flavor == clust.Flavor.Name {
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
