package node

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Wrapper around edgeproto.NodeCache to add in the region
// There are three cases:
// NotifyRoot: region is "", region in NodeKey is already set
// Controller: region is set, override region in NodeKey
// CRM/DME: region is "", region is NodeKey is not set, but will get set
// once it goes to controller.

type RegionNodeCache struct {
	edgeproto.NodeCache
	setRegion string
}

func (s *RegionNodeCache) Update(ctx context.Context, in *edgeproto.Node, rev int64) {
	if s.setRegion != "" {
		in.Key.Region = s.setRegion
	}
	s.NodeCache.Update(ctx, in, rev)
}

func (s *RegionNodeCache) Delete(ctx context.Context, in *edgeproto.Node, rev int64) {
	if s.setRegion != "" {
		in.Key.Region = s.setRegion
	}
	s.NodeCache.Delete(ctx, in, rev)
}

func (s *RegionNodeCache) Prune(ctx context.Context, validKeys map[edgeproto.NodeKey]struct{}) {
	if s.setRegion != "" {
		keys := make(map[edgeproto.NodeKey]struct{})
		for k, _ := range validKeys {
			k.Region = s.setRegion
			keys[k] = struct{}{}
		}
		validKeys = keys
	}
	s.NodeCache.Prune(ctx, validKeys)
}

func nodeMatches(key *edgeproto.NodeKey, filter *edgeproto.NodeKey) bool {
	// if region is not set on node, then this is a node below
	// controller in the notify tree that doesn't know what region
	// it is in, so don't filter based on region.
	if key.Region == "" && filter.Region != "" {
		f := *filter
		f.Region = ""
		filter = &f
	}
	return key.Matches(filter, edgeproto.MatchFilter())
}
