package dind

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/nginx"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
}

func (s *Platform) GetType() string {
	return "dind"
}

func (s *Platform) Init(platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	// set up L7 load balancer
	client, err := s.GetPlatformClient(nil)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Setting up Nginx L7 Proxy")
	err = nginx.InitL7Proxy(client, nginx.WithDockerPublishPorts())
	if err != nil {
		return err
	}
	return nil
}

func (s *Platform) GatherCloudletInfo(info *edgeproto.CloudletInfo) error {
	err := GetLimits(info)
	if err != nil {
		return err
	}
	info.Flavors = []*edgeproto.FlavorInfo{
		&edgeproto.FlavorInfo{
			Name:  "DINDFlavor",
			Vcpus: uint64(info.OsMaxVcores),
			Ram:   uint64(info.OsMaxRam),
			Disk:  uint64(500),
		},
	}
	return nil
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return &pc.LocalClient{}, nil
}
