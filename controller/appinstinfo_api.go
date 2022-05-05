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

type AppInstInfoApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.AppInstInfoStore
}

func NewAppInstInfoApi(sync *Sync, all *AllApis) *AppInstInfoApi {
	appInstInfoApi := AppInstInfoApi{}
	appInstInfoApi.all = all
	appInstInfoApi.sync = sync
	appInstInfoApi.store = edgeproto.NewAppInstInfoStore(sync.store)
	return &appInstInfoApi
}

func (s *AppInstInfoApi) Update(ctx context.Context, in *edgeproto.AppInstInfo, rev int64) {
	s.all.appInstApi.UpdateFromInfo(ctx, in)
}

func (s *AppInstInfoApi) Delete(ctx context.Context, in *edgeproto.AppInstInfo, rev int64) {
	// for backwards compatibility
	s.all.appInstApi.DeleteFromInfo(ctx, in)
}

func (s *AppInstInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *AppInstInfoApi) Prune(ctx context.Context, keys map[edgeproto.AppInstKey]struct{}) {
	// no-op
}
