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

package main

import (
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/edgexr/edge-cloud/edgeproto"
)

type ClusterRefsApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.ClusterRefsStore
	cache edgeproto.ClusterRefsCache
}

func NewClusterRefsApi(sync *Sync, all *AllApis) *ClusterRefsApi {
	clusterRefsApi := ClusterRefsApi{}
	clusterRefsApi.all = all
	clusterRefsApi.sync = sync
	clusterRefsApi.store = edgeproto.NewClusterRefsStore(sync.store)
	edgeproto.InitClusterRefsCache(&clusterRefsApi.cache)
	sync.RegisterCache(&clusterRefsApi.cache)
	return &clusterRefsApi
}

func (s *ClusterRefsApi) ShowClusterRefs(in *edgeproto.ClusterRefs, cb edgeproto.ClusterRefsApi_ShowClusterRefsServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.ClusterRefs) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *ClusterRefsApi) deleteRef(stm concurrency.STM, key *edgeproto.ClusterInstKey) {
	s.store.STMDel(stm, key)
}

func (s *ClusterRefsApi) addRef(stm concurrency.STM, appInst *edgeproto.AppInst) {
	// note the ClusterInst key is the real cluster key, not the virtual
	key := appInst.ClusterInstKey()
	refs := edgeproto.ClusterRefs{}
	if !s.store.STMGet(stm, key, &refs) {
		refs.Key = *key
	}
	refs.Apps = append(refs.Apps, *appInst.Key.ClusterRefsAppInstKey())
	s.store.STMPut(stm, &refs)
}

func (s *ClusterRefsApi) removeRef(stm concurrency.STM, appInst *edgeproto.AppInst) {
	key := appInst.ClusterInstKey()
	refs := edgeproto.ClusterRefs{}
	if !s.store.STMGet(stm, key, &refs) {
		return
	}
	changed := false
	refKey := appInst.Key.ClusterRefsAppInstKey()
	for ii := range refs.Apps {
		if refKey.Matches(&refs.Apps[ii]) {
			refs.Apps = append(refs.Apps[:ii], refs.Apps[ii+1:]...)
			changed = true
			break
		}
	}
	if !changed {
		return
	}
	if len(refs.Apps) == 0 {
		s.store.STMDel(stm, key)
	} else {
		s.store.STMPut(stm, &refs)
	}
}
