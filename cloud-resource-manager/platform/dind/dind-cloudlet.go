package dind

import (
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *Platform) CreatePlatform(pf *edgeproto.Platform, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelMexos, "create platform for dind")
	return nil
}

func (s *Platform) DeletePlatform(pf *edgeproto.Platform) error {
	log.DebugLog(log.DebugLevelMexos, "delete platform for dind")
	return nil
}

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, pf *edgeproto.Platform, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelMexos, "create cloudlet for dind")
	updateCallback(edgeproto.UpdateTask, "Creating Cloudlet")

	updateCallback(edgeproto.UpdateTask, "Starting CRMServer")
	err := cloudcommon.StartCRMService(cloudlet, pf)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "dind cloudlet create failed", "err", err)
		return err
	}
	return nil
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet, pf *edgeproto.Platform) error {
	log.DebugLog(log.DebugLevelMexos, "delete cloudlet for dind")
	err := cloudcommon.StopCRMService(cloudlet, pf)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "dind cloudlet delete failed", "err", err)
		return err
	}

	return nil
}
