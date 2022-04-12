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
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

const NoRegion = ""

// CloudletLookup interface used by events to get the kafka cluster endpoint
// for a cloudlet in order to send events out
type CloudletLookup interface {
	GetCloudlet(region string, key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool
	GetCloudletCache(region string) *edgeproto.CloudletCache
}

type CloudletCache struct {
	cache edgeproto.CloudletCache
}

func (s *CloudletCache) Init() {
	edgeproto.InitCloudletCache(&s.cache)
}

func (s *CloudletCache) GetCloudlet(region string, key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool {
	return s.cache.Get(key, buf)
}

func (s *CloudletCache) GetCloudletCache(region string) *edgeproto.CloudletCache {
	return &s.cache
}
