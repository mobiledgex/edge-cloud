package crmutil

import (
	"context"
	"encoding/json"
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
	var cloudlet edgeproto.Cloudlet

	if !s.controllerData.CloudletInfoCache.Get(&s.controllerData.cloudletKey, &cloudletInfo) {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to find cloudlet info in cache", "cloudletKey", s.controllerData.cloudletKey)
		return fmt.Errorf("cannot find in cloudlet info in cache for key %s", s.controllerData.cloudletKey.String())
	}
	if !s.controllerData.CloudletCache.Get(&s.controllerData.cloudletKey, &cloudlet) {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to find cloudlet in cache", "cloudletKey", s.controllerData.cloudletKey)
		return fmt.Errorf("cannot find in cloudlet info in cache for key %s", s.controllerData.cloudletKey.String())
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "ActiveChangedPostSwitch", "cloudletInfo", cloudletInfo, "cloudlet", cloudlet, "PlatformInitComplete", s.controllerData.PlatformInitComplete)

	// if the platform is already initialized, copy the cloudlet info saved from the previously active unit. If not then the cloudletInfo will be rebuilt
	if s.controllerData.PlatformInitComplete {
		val, err := s.controllerData.highAvailabilityManager.GetValue(ctx, "cloudletInfo")
		if err != nil {
			return err
		}
		if val == "" {
			log.SpanLog(ctx, log.DebugLevelInfra, "no existing cloudlet info found")
		} else {
			err = json.Unmarshal([]byte(val), &cloudletInfo)
			if err != nil {
				return fmt.Errorf("cloudletInfo unmarshal err - %v", err)
			}
		}
	}

	cloudletInfo.ActiveCrmInstance = s.controllerData.highAvailabilityManager.HARole
	s.controllerData.UpdateCloudletInfo(ctx, &cloudletInfo)
	return nil
}
