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

func (s *CrmHAProcess) ActiveChangedPreSwitch(ctx context.Context, platformActive bool) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ActiveChangedPreSwitch", "platformActive", platformActive)
	if !platformActive {
		// not supported, CRM should have been killed within HA manager
		log.FatalLog("Error: Unexpected CRM transition to inactive", "cloudletKey", s.controllerData.cloudletKey)
	}
	if s.controllerData.platform != nil {
		s.controllerData.platform.ActiveChanged(ctx, platformActive)
	} else {
		return fmt.Errorf("CRM HA platform is nil - %v", s.controllerData.cloudletKey)
	}
	return nil
}

func (s *CrmHAProcess) ActiveChangedPostSwitch(ctx context.Context, platformActive bool) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ActiveChangedPostSwitch", "platformActive", platformActive)
	var cloudletInfo edgeproto.CloudletInfo
	if !s.controllerData.CloudletInfoCache.Get(&s.controllerData.cloudletKey, &cloudletInfo) {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to find cloudlet info in cache", "cloudletKey", s.controllerData.cloudletKey)
		return fmt.Errorf("cannot find in cloudlet info in cache for key %s", s.controllerData.cloudletKey.String())
	}
	s.controllerData.UpdateCloudletInfo(ctx, &cloudletInfo)
	return nil
}
