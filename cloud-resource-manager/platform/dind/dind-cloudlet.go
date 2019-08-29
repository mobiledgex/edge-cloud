package dind

import (
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(s.ctx, log.DebugLevelMexos, "create cloudlet for dind")
	updateCallback(edgeproto.UpdateTask, "Creating Cloudlet")

	updateCallback(edgeproto.UpdateTask, "Starting CRMServer")
	err := cloudcommon.StartCRMService(s.ctx, cloudlet, pfConfig)
	if err != nil {
		log.SpanLog(s.ctx, log.DebugLevelMexos, "dind cloudlet create failed", "err", err)
		return err
	}
	return nil
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error {
	log.SpanLog(s.ctx, log.DebugLevelMexos, "delete cloudlet for dind")
	err := cloudcommon.StopCRMService(s.ctx, cloudlet)
	if err != nil {
		log.SpanLog(s.ctx, log.DebugLevelMexos, "dind cloudlet delete failed", "err", err)
		return err
	}

	return nil
}
