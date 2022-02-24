package crmutil

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type CrmHAProcess struct {
	controllerData                 *ControllerData
	FinishUpdateCloudletInfoThread chan struct{}
}

func (s *CrmHAProcess) ActiveChangedPreSwitch(ctx context.Context, platformActive bool) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ActiveChangedPreSwitch", "platformActive", platformActive)
	if !platformActive {
		// not supported, CRM should have been killed within HA manager
		log.SpanFromContext(ctx).Finish()
		log.FatalLog("Error: Unexpected CRM transition to inactive", "cloudletKey", s.controllerData.cloudletKey)
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
	log.SpanLog(ctx, log.DebugLevelInfra, "ActiveChangedPostSwitch", "PlatformCommonInitDone", s.controllerData.PlatformCommonInitDone)

	select {
	case s.controllerData.WaitPlatformActive <- true:
	default:
		// this is not expected because the channel should be filled either by transitioning from
		// standby to active, or starting out active. But as there is no transition for the CRM to go
		// active to standby without restarting, the channel should never be filled more than once
		log.SpanFromContext(ctx).Finish()
		log.FatalLog("WaitPlatformActive channel already full")
	}
	return nil
}

func (s *CrmHAProcess) PlatformActiveOnStartup(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfra, "PlatformActiveOnStartup")
	select {
	case s.controllerData.WaitPlatformActive <- true:
	default:
		// this is not expected because the channel should be filled either by transitioning from
		// standby to active, or starting out active. But as there is no transition for the CRM to go
		// active to standby without restarting, the channel should never be filled more than once
		log.SpanFromContext(ctx).Finish()
		log.FatalLog("WaitPlatformActive channel already full")
	}
}

func (s *CrmHAProcess) DumpWatcherFields(ctx context.Context) map[string]interface{} {
	watcherStatus := make(map[string]interface{})
	watcherStatus["Type"] = "CrmHAProcess"
	watcherStatus["PlatformCommonInitDone"] = s.controllerData.PlatformCommonInitDone
	watcherStatus["UpdateHACompatibilityVersion"] = s.controllerData.UpdateHACompatibilityVersion
	watcherStatus["ControllerSyncInProgress"] = s.controllerData.ControllerSyncInProgress
	return watcherStatus
}
