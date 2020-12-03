package dind

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

func (s *Platform) CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, caches *platform.Caches, acccessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "create cloudlet for dind")
	updateCallback(edgeproto.UpdateTask, "Creating Cloudlet")

	updateCallback(edgeproto.UpdateTask, "Starting CRMServer")
	err := cloudcommon.StartCRMService(ctx, cloudlet, pfConfig)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "dind cloudlet create failed", "err", err)
		return err
	}
	return nil
}

func (s *Platform) UpdateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelInfra, "update dind Cloudlet", "key", cloudlet.Key)
	return nil
}

func (s *Platform) DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, caches *platform.Caches, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "delete cloudlet for dind")
	updateCallback(edgeproto.UpdateTask, "Deleting Cloudlet")
	updateCallback(edgeproto.UpdateTask, "Stopping CRMServer")
	err := cloudcommon.StopCRMService(ctx, cloudlet)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "dind cloudlet delete failed", "err", err)
		return err
	}

	return nil
}

func (s *Platform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Saving cloudlet access vars", "cloudletName", cloudlet.Key.Name)
	return nil
}

func (s *Platform) DeleteCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Deleting cloudlet access vars", "cloudletName", cloudlet.Key.Name)
	return nil
}

func (s *Platform) GetCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config) (map[string]string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "Get cloudlet access vars", "cloudletName", cloudlet.Key.Name)
	return map[string]string{}, nil
}

func (s *Platform) SyncControllerCache(ctx context.Context, caches *platform.Caches, cloudletState edgeproto.CloudletState) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SyncControllerData", "state", cloudletState)
	return nil
}

func (s *Platform) GetCloudletManifest(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, accessApi platform.AccessApi, flavor *edgeproto.Flavor, caches *platform.Caches) (*edgeproto.CloudletManifest, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "Get cloudlet manifest", "cloudletName", cloudlet.Key.Name)
	return &edgeproto.CloudletManifest{Manifest: "dind manifest"}, nil
}

func (s *Platform) VerifyVMs(ctx context.Context, vms []edgeproto.VM) error {
	return nil
}

func (s *Platform) GetRestrictedCloudletStatus(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	return nil
}
