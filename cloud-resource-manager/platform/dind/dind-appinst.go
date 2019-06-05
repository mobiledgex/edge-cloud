package dind

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/nginx"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *Platform) CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor) error {
	log.DebugLog(log.DebugLevelMexos, "call runKubectlCreateApp for dind")

	var err error
	client := &pc.LocalClient{}
	appDeploymentType := app.Deployment

	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}

	if len(appInst.MappedPorts) > 0 {
		log.DebugLog(log.DebugLevelMexos, "AddNginxProxy for dind", "ports", appInst.MappedPorts)
		cluster, err := FindCluster(names.ClusterName)
		if err != nil {
			return err
		}
		masterIP := cluster.MasterAddr
		network := GetDockerNetworkName(cluster)
		err = nginx.CreateNginxProxy(client,
			names.AppName,
			masterIP,
			appInst.MappedPorts,
			nginx.WithDockerNetwork(network),
			nginx.WithDockerPublishPorts())
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "cannot add nginx proxy", "appName", names.AppName, "ports", appInst.MappedPorts)
			return err
		}
	}

	if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
		err = k8smgmt.CreateAppInst(client, names, app, appInst)
		if err == nil {
			err = k8smgmt.WaitForAppInst(client, names, app)
		}
	} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
		err = k8smgmt.CreateHelmAppInst(client, names, clusterInst, app, appInst)
	} else {
		err = fmt.Errorf("invalid deployment type %s for dind", appDeploymentType)
	}
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "error creating dind app")
		return err
	}
	return nil
}

func (s *Platform) DeleteAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	log.DebugLog(log.DebugLevelMexos, "run kubectl delete app for dind")

	var err error
	client := &pc.LocalClient{}
	appDeploymentType := app.Deployment

	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}

	if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
		err = k8smgmt.DeleteAppInst(client, names, app, appInst)
	} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
		err = k8smgmt.DeleteHelmAppInst(client, names, clusterInst)
	} else {
		err = fmt.Errorf("invalid deployment type %s for dind", appDeploymentType)
	}
	if err != nil {
		return err
	}

	if len(appInst.MappedPorts) > 0 {
		log.DebugLog(log.DebugLevelMexos, "DeleteNginxProxy for dind")
		if err = nginx.DeleteNginxProxy(client, names.AppName); err != nil {
			log.DebugLog(log.DebugLevelMexos, "cannot delete nginx proxy", "name", names.AppName)
			return err
		}
	}
	return nil
}

func (s *Platform) GetAppInstRuntime(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	client, err := s.GetPlatformClient(clusterInst)
	if err != nil {
		return nil, err
	}
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return nil, err
	}

	return k8smgmt.GetAppInstRuntime(client, names, app, appInst)
}

func (s *Platform) GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return k8smgmt.GetContainerCommand(clusterInst, app, appInst, req)
}
