package fake

import (
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	ut "github.com/mobiledgex/edge-cloud/setup-env/util"
	"github.com/mobiledgex/edge-cloud/util"
)

type Platform struct {
}

func (s *Platform) GetType() string {
	return "fake"
}

func (s *Platform) Init(platformConfig *platform.PlatformConfig) error {
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

func (s *Platform) UpdateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("update cluster not supported for fake cloudlets")
}
func (s *Platform) CreateClusterInst(clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "First Create Task")
	updateCallback(edgeproto.UpdateTask, "Second Create Task")
	log.DebugLog(log.DebugLevelMexos, "fake ClusterInst ready")
	return nil
}

func (s *Platform) DeleteClusterInst(clusterInst *edgeproto.ClusterInst) error {
	log.DebugLog(log.DebugLevelMexos, "fake ClusterInst deleted")
	return nil
}

func (s *Platform) CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "Creating App Inst")
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

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return &pc.LocalClient{}, nil
}

func (s *Platform) GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return req.Command, nil
}

func getCrmProc(cloudlet *edgeproto.Cloudlet) (*process.Crm, error) {
	return &process.Crm{
		ApiAddr:     fmt.Sprintf("0.0.0.0:%d", cloudlet.BindPort),
		NotifyAddrs: fmt.Sprintf("%s:37001,%s:37002", cloudlet.ControllerAddr, cloudlet.ControllerAddr),
		CloudletKey: fmt.Sprintf("{\"operator_key\":{\"name\":\"%s\"},\"name\":\"%s\"}", cloudlet.Key.OperatorKey.Name, cloudlet.Key.Name),
		Platform:    cloudlet.Platform,
		Common: process.Common{
			Hostname: "127.0.0.1",
		},
		TLS: process.TLSCerts{
			// ServerCert: "/Users/ashishjain/go/src/github.com/mobiledgex/edge-cloud/tls/out/mex-server.crt",
			ServerCert: "tls/out/mex-server.crt",
		},
	}, nil
}

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, updateCallback edgeproto.CacheUpdateCallback) error {
	log.DebugLog(log.DebugLevelMexos, "create fake Cloudlet", "key", cloudlet.Key)
	updateCallback(edgeproto.UpdateTask, "Creating Cloudlet")

	crmProc, err := getCrmProc(cloudlet)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "fake Cloudlet failed", "err", err)
		return err
	}
	opts := []process.StartOp{}
	opts = append(opts, process.WithDebug("mexos"))

	err = crmProc.StartLocal("/tmp/e2e_test_out/"+util.DNSSanitize(cloudlet.Key.Name)+".log", opts...)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "fake Cloudlet failed", "err", err)
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Cloudlet created successfully")
	return nil
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error {
	log.DebugLog(log.DebugLevelMexos, "delete fake Cloudlet")
	crmProc, err := getCrmProc(cloudlet)
	if err != nil {
		return err
	}
	maxwait := 5 * time.Second

	crmProc.StopLocal()
	c := make(chan string)
	go ut.KillProcessesByName(crmProc.GetExeName(), maxwait, crmProc.LookupArgs(), c)
	log.DebugLog(log.DebugLevelMexos, "fake Cloudlet deleted", "msg", <-c)

	return nil
}
