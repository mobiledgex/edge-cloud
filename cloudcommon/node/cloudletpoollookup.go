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

package node

import (
	"context"

	"github.com/edgexr/edge-cloud/edgeproto"
)

// CloudletPoolLookup interface used by events to determine if cloudlet
// is in a CloudletPool for proper RBAC marking of events.
type CloudletPoolLookup interface {
	InPool(region string, key edgeproto.CloudletKey) bool
	GetCloudletPoolCache(region string) *edgeproto.CloudletPoolCache
	Dumpable() map[string]interface{}
}

type CloudletPoolCache struct {
	cache           edgeproto.CloudletPoolCache
	PoolsByCloudlet edgeproto.CloudletPoolByCloudletKey
}

func (s *CloudletPoolCache) Init() {
	edgeproto.InitCloudletPoolCache(&s.cache)
	s.PoolsByCloudlet.Init()
	s.cache.AddUpdatedCb(s.updatedPool)
	s.cache.AddDeletedCb(s.deletedPool)
}

func (s *CloudletPoolCache) updatedPool(ctx context.Context, old, new *edgeproto.CloudletPool) {
	s.PoolsByCloudlet.Updated(old, new)
}

func (s *CloudletPoolCache) deletedPool(ctx context.Context, old *edgeproto.CloudletPool) {
	s.PoolsByCloudlet.Deleted(old)
}

func (s *CloudletPoolCache) Dumpable() map[string]interface{} {
	return s.PoolsByCloudlet.Dumpable()
}

func (s *CloudletPoolCache) InPool(region string, key edgeproto.CloudletKey) bool {
	return s.PoolsByCloudlet.HasRef(key)
}

func (s *CloudletPoolCache) GetCloudletPoolCache(region string) *edgeproto.CloudletPoolCache {
	return &s.cache
}
