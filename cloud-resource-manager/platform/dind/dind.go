package dind

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/common/xind"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/redundancy"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
	xind.Xind
}

func (s *Platform) InitCommon(ctx context.Context, platformConfig *platform.PlatformConfig, caches *platform.Caches, haMgr *redundancy.HighAvailabilityManager, updateCallback edgeproto.CacheUpdateCallback) error {
	return s.Xind.InitCommon(ctx, platformConfig, caches, s, updateCallback)
}

func (s *Platform) InitHAConditional(ctx context.Context, platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	return s.Xind.InitHAConditional(ctx, platformConfig, updateCallback)
}

func (s *Platform) GetInitHAConditionalCompatibilityVersion(ctx context.Context) string {
	return s.Xind.GetInitHAConditionalCompatibilityVersion(ctx)
}

func (s *Platform) ActiveChanged(ctx context.Context, platformActive bool) error {
	return nil
}
