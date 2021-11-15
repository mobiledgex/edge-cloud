package crmutil

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
)

type CrmHAProcess struct {
	controllerData *ControllerData
}

func (s *CrmHAProcess) BecomeActiveCallback(ctx context.Context, haRole process.HARole) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "BecomeActiveCallback")
	var cloudletInfo edgeproto.CloudletInfo
	if !s.controllerData.CloudletInfoCache.Get(&s.controllerData.cloudletKey, &cloudletInfo) {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to find cloudlet info in cache", "cloudletKey", s.controllerData.cloudletKey)
		return fmt.Errorf("Cannot find in cloudlet info in cache for key %s", s.controllerData.cloudletKey.String())
	}
	if s.controllerData.platform != nil {
		s.controllerData.platform.BecomeActive(ctx, s.controllerData.highAvailabilityManager.HARole)
	} else {
		// possible on first startup
		log.SpanLog(ctx, log.DebugLevelInfra, "CRM HA platform is nil", s.controllerData.cloudletKey)
	}
	s.controllerData.UpdateCloudletInfo(ctx, &cloudletInfo)
	return nil
}
