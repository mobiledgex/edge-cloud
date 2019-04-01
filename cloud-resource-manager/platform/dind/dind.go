package dind

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type Platform struct {
	Clusters      map[string]*DindCluster
	nextClusterID int
	// TODO: may need a lock around Clusters
}

func (s *Platform) GetType() string {
	return "dind"
}

func (s *Platform) Init(key *edgeproto.CloudletKey) error {
	s.Clusters = make(map[string]*DindCluster)
	return nil
}

func (s *Platform) GatherCloudletInfo(info *edgeproto.CloudletInfo) error {
	return GetLimits(info)
}

func (s *Platform) GetPlatformClient() pc.PlatformClient {
	return &pc.LocalClient{}
}
