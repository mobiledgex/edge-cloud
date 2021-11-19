package crmutil

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type CrmHAProcess struct {
	controllerData *ControllerData
}

func (s *CrmHAProcess) ActiveChanged(ctx context.Context, platformActive bool) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ActiveChanged", "platformActive", platformActive)
	if platformActive {
		var cloudletInfo edgeproto.CloudletInfo
		if !s.controllerData.CloudletInfoCache.Get(&s.controllerData.cloudletKey, &cloudletInfo) {
			log.SpanLog(ctx, log.DebugLevelInfra, "failed to find cloudlet info in cache", "cloudletKey", s.controllerData.cloudletKey)
			return fmt.Errorf("Cannot find in cloudlet info in cache for key %s", s.controllerData.cloudletKey.String())
		}
		s.controllerData.UpdateCloudletInfo(ctx, &cloudletInfo)
	}
	if s.controllerData.platform != nil {
		s.controllerData.platform.ActiveChanged(ctx, platformActive)
	} else {
		// possible on first startup
		log.SpanLog(ctx, log.DebugLevelInfra, "CRM HA platform is nil", s.controllerData.cloudletKey)
	}
	return nil
}
