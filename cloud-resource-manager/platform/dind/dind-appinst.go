package dind

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/dockermgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *Platform) CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "call runKubectlCreateApp for dind")

	var err error
	client := &pc.LocalClient{}
	appDeploymentType := app.Deployment
	// Support for local docker appInst
	if appDeploymentType == cloudcommon.AppDeploymentTypeDocker {
		log.SpanLog(ctx, log.DebugLevelMexos, "run docker create app for dind")
		err = dockermgmt.CreateAppInstLocal(client, app, appInst)
		if err != nil {
			return fmt.Errorf("CreateAppInstLocal error for docker %v", err)
		}
		return nil
	}
	// Now for helm and k8s apps
	log.SpanLog(ctx, log.DebugLevelMexos, "run kubectl create app for dind")
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}
	cluster, err := FindCluster(names.ClusterName)
	if err != nil {
		return err
	}
	masterIP := cluster.MasterAddr
	network := GetDockerNetworkName(cluster)
	// NOTE: for DIND we don't check whether this is internal
	if len(appInst.MappedPorts) > 0 {
		log.SpanLog(ctx, log.DebugLevelMexos, "Add Proxy for dind", "ports", appInst.MappedPorts)
		err = proxy.CreateNginxProxy(ctx, client,
			names.AppName,
			masterIP,
			appInst.MappedPorts,
			proxy.WithDockerNetwork(network),
			proxy.WithDockerPublishPorts())
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMexos, "cannot add proxy", "appName", names.AppName, "ports", appInst.MappedPorts)
			return err
		}
	}

	// Add crm local replace variables
	deploymentVars := crmutil.DeploymentReplaceVars{
		Deployment: crmutil.CrmReplaceVars{
			ClusterIp: masterIP,
		},
	}
	ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

	if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
		err = k8smgmt.CreateAppInst(ctx, client, names, app, appInst)
		if err == nil {
			err = k8smgmt.WaitForAppInst(ctx, client, names, app, k8smgmt.WaitRunning)
		}
	} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
		err = k8smgmt.CreateHelmAppInst(ctx, client, names, clusterInst, app, appInst)
	} else {
		err = fmt.Errorf("invalid deployment type %s for dind", appDeploymentType)
	}
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "error creating dind app")
		return err
	}
	return nil
}

func (s *Platform) DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	var err error
	client := &pc.LocalClient{}
	appDeploymentType := app.Deployment
	// Support for local docker appInst
	if appDeploymentType == cloudcommon.AppDeploymentTypeDocker {
		log.SpanLog(ctx, log.DebugLevelMexos, "run docker delete app for dind")
		err = dockermgmt.DeleteAppInst(ctx, client, app, appInst)
		if err != nil {
			return fmt.Errorf("DeleteAppInst error for docker %v", err)
		}
		return nil
	}
	// Now for helm and k8s apps
	log.SpanLog(ctx, log.DebugLevelMexos, "run kubectl delete app for dind")
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}

	if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
		err = k8smgmt.DeleteAppInst(ctx, client, names, app, appInst)
	} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
		err = k8smgmt.DeleteHelmAppInst(ctx, client, names, clusterInst)
	} else {
		err = fmt.Errorf("invalid deployment type %s for dind", appDeploymentType)
	}
	if err != nil {
		return err
	}

	if len(appInst.MappedPorts) > 0 {
		log.SpanLog(ctx, log.DebugLevelMexos, "DeleteNginxProxy for dind")
		if err = proxy.DeleteNginxProxy(ctx, client, names.AppName); err != nil {
			log.SpanLog(ctx, log.DebugLevelMexos, "cannot delete proxy", "name", names.AppName)
			return err
		}
	}
	return nil
}

func (s *Platform) UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {

	log.SpanLog(ctx, log.DebugLevelMexos, "UpdateAppInst for dind")
	client := &pc.LocalClient{}
	appDeploymentType := app.Deployment
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}

	cluster, err := FindCluster(names.ClusterName)
	if err != nil {
		return err
	}
	// Add crm local replace variables
	deploymentVars := crmutil.DeploymentReplaceVars{
		Deployment: crmutil.CrmReplaceVars{
			ClusterIp: cluster.MasterAddr,
		},
	}
	ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

	if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
		return k8smgmt.UpdateAppInst(ctx, client, names, app, appInst)
	} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
		return k8smgmt.UpdateHelmAppInst(ctx, client, names, app, appInst)
	}
	return fmt.Errorf("UpdateAppInst not supported for deployment: %s", appDeploymentType)
}

func (s *Platform) GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	client, err := s.GetPlatformClient(ctx, clusterInst)
	if err != nil {
		return nil, err
	}
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return nil, err
	}

	return k8smgmt.GetAppInstRuntime(ctx, client, names, app, appInst)
}

func (s *Platform) GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return k8smgmt.GetContainerCommand(ctx, clusterInst, app, appInst, req)
}

func (s *Platform) GetConsoleUrl(ctx context.Context, app *edgeproto.App) (string, error) {
	return "", nil
}
