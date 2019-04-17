package fake

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type Platform struct{}

func (s *Platform) GetType() string {
	return "fake"
}

func (s *Platform) Init(key *edgeproto.CloudletKey) error {
	log.DebugLog(log.DebugLevelMexos, "running in fake cloudlet mode")
	return nil
}

func (s *Platform) GatherCloudletInfo(info *edgeproto.CloudletInfo) error {
	info.OsMaxRam = 500
	info.OsMaxVcores = 50
	info.OsMaxVolGb = 5000
	info.Flavors = []*edgeproto.FlavorInfo{
		&edgeproto.FlavorInfo{
			Name:  "flavor1",
			Vcpus: uint64(10),
			Ram:   uint64(101024),
			Disk:  uint64(500),
		},
	}
	return nil
}

func (s *Platform) CreateCluster(clusterInst *edgeproto.ClusterInst, flavor *edgeproto.ClusterFlavor) error {
	log.DebugLog(log.DebugLevelMexos, "fake ClusterInst ready")
	return nil
}

func (s *Platform) DeleteCluster(clusterInst *edgeproto.ClusterInst) error {
	log.DebugLog(log.DebugLevelMexos, "fake ClusterInst deleted")
	return nil
}

func (s *Platform) CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	log.DebugLog(log.DebugLevelMexos, "fake AppInst ready")
	return nil
}

func (s *Platform) DeleteAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	log.DebugLog(log.DebugLevelMexos, "fake AppInst deleted")
	return nil
}

func (s *Platform) GetAppInstRuntime(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	return &edgeproto.AppInstRuntime{}, nil
}

func (s *Platform) GetPlatformClient() (pc.PlatformClient, error) {
	return &pc.LocalClient{}, nil
}

func (s *Platform) GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return req.Command, nil
}
