package dind

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *Platform) CreateAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, names *k8smgmt.KubeNames) error {
	log.DebugLog(log.DebugLevelMexos, "call runKubectlCreateApp for dind")

	var err error
	client := &pc.LocalClient{}
	appDeploymentType := app.Deployment

	if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
		err = k8smgmt.CreateAppInst(client, names, app, appInst)
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

func (s *Platform) DeleteAppInst(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, names *k8smgmt.KubeNames) error {
	log.DebugLog(log.DebugLevelMexos, "run kubectl delete app for dind")

	var err error
	client := &pc.LocalClient{}
	appDeploymentType := app.Deployment

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
	return nil
}
