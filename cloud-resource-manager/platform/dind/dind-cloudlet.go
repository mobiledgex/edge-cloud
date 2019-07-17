package dind

import (
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelMexos, "create cloudlet for dind")
	updateCallback(edgeproto.UpdateTask, "Creating Cloudlet")

	updateCallback(edgeproto.UpdateTask, "Starting CRMServer")
	err := cloudcommon.StartCRMService(cloudlet)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "dind cloudlet create failed", "err", err)
		return err
	}
	return nil
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error {
	log.DebugLog(log.DebugLevelMexos, "delete cloudlet for dind")
	err := cloudcommon.StopCRMService(cloudlet)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "dind cloudlet delete failed", "err", err)
		return err
	}

	return nil
}
