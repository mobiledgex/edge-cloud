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

type CloudletRefsApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.CloudletRefsStore
	cache edgeproto.CloudletRefsCache
}

func NewCloudletRefsApi(sync *Sync, all *AllApis) *CloudletRefsApi {
	cloudletRefsApi := CloudletRefsApi{}
	cloudletRefsApi.all = all
	cloudletRefsApi.sync = sync
	cloudletRefsApi.store = edgeproto.NewCloudletRefsStore(sync.store)
	edgeproto.InitCloudletRefsCache(&cloudletRefsApi.cache)
	sync.RegisterCache(&cloudletRefsApi.cache)
	return &cloudletRefsApi
}

func (s *CloudletRefsApi) Delete(ctx context.Context, key *edgeproto.CloudletKey, wait func(int64)) {
	in := edgeproto.CloudletRefs{Key: *key}
	s.store.Delete(ctx, &in, wait)
}

func (s *CloudletRefsApi) ShowCloudletRefs(in *edgeproto.CloudletRefs, cb edgeproto.CloudletRefsApi_ShowCloudletRefsServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletRefs) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func initCloudletRefs(refs *edgeproto.CloudletRefs, key *edgeproto.CloudletKey) {
	refs.Key = *key
	refs.RootLbPorts = make(map[int32]int32)
}
