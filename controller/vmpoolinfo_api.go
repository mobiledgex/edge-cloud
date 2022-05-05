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

type VMPoolInfoApi struct {
	all   *AllApis
	cache edgeproto.VMPoolInfoCache
}

func NewVMPoolInfoApi(sync *Sync, all *AllApis) *VMPoolInfoApi {
	vmPoolInfoApi := VMPoolInfoApi{}
	vmPoolInfoApi.all = all
	edgeproto.InitVMPoolInfoCache(&vmPoolInfoApi.cache)
	return &vmPoolInfoApi
}

func (s *VMPoolInfoApi) Update(ctx context.Context, in *edgeproto.VMPoolInfo, rev int64) {
	s.all.vmPoolApi.UpdateFromInfo(ctx, in)
}

func (s *VMPoolInfoApi) Delete(ctx context.Context, in *edgeproto.VMPoolInfo, rev int64) {
	// no-op
}

func (s *VMPoolInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *VMPoolInfoApi) Prune(ctx context.Context, keys map[edgeproto.VMPoolKey]struct{}) {
	// no-op
}
