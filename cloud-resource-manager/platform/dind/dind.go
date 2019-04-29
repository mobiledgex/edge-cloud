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
