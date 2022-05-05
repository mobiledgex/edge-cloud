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
	"context"

	"github.com/edgexr/edge-cloud/edgeproto"
)

type ClusterInstInfoApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.ClusterInstInfoStore
}

func NewClusterInstInfoApi(sync *Sync, all *AllApis) *ClusterInstInfoApi {
	clusterInstInfoApi := ClusterInstInfoApi{}
	clusterInstInfoApi.all = all
	clusterInstInfoApi.sync = sync
	clusterInstInfoApi.store = edgeproto.NewClusterInstInfoStore(sync.store)
	return &clusterInstInfoApi
}

func (s *ClusterInstInfoApi) Update(ctx context.Context, in *edgeproto.ClusterInstInfo, rev int64) {
	s.all.clusterInstApi.UpdateFromInfo(ctx, in)
}

func (s *ClusterInstInfoApi) Delete(ctx context.Context, in *edgeproto.ClusterInstInfo, rev int64) {
	// for backwards compatibility
	s.all.clusterInstApi.DeleteFromInfo(ctx, in)
}

func (s *ClusterInstInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *ClusterInstInfoApi) Prune(ctx context.Context, keys map[edgeproto.ClusterInstKey]struct{}) {
	// no-op
}
