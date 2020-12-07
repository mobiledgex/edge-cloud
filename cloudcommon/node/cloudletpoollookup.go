package node

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
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
