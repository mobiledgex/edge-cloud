package node

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// CloudletLookup interface used by events to get the kafka cluster endpoint
// for a cloudlet in order to send events out
type CloudletLookup interface {
	Get(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool
	GetCloudletCache(region string) *edgeproto.CloudletCache
}

type CloudletCache struct {
	cache edgeproto.CloudletCache
}

func (s *CloudletCache) Init() {
	edgeproto.InitCloudletCache(&s.cache)
}

func (s *CloudletCache) Get(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool {
	return s.cache.Get(key, buf)
}

func (s *CloudletCache) GetCloudletCache(region string) *edgeproto.CloudletCache {
	return &s.cache
}
