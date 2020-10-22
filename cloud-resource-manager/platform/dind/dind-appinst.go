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

func (s *Platform) CreateAppInstInternal(ctx context.Context, clusterInst *edgeproto.ClusterInst,
	app *edgeproto.App, appInst *edgeproto.AppInst, names *k8smgmt.KubeNames) error {
	var err error
	client := &pc.LocalClient{}
	DeploymentType := app.Deployment
	// Support for local docker appInst
	if DeploymentType == cloudcommon.DeploymentTypeDocker {
		log.SpanLog(ctx, log.DebugLevelInfra, "run docker create app for dind")
		err = dockermgmt.CreateAppInstLocal(client, app, appInst)
		if err != nil {
			return fmt.Errorf("CreateAppInstLocal error for docker %v", err)
		}
		return nil
	}
	// Now for helm and k8s apps
	log.SpanLog(ctx, log.DebugLevelInfra, "run kubectl create app for dind")
	cluster, err := FindCluster(names.ClusterName)
	if err != nil {
		return err
	}
	masterIP := cluster.MasterAddr
	network := GetDockerNetworkName(cluster)
	// NOTE: for DIND we don't check whether this is internal
	if len(appInst.MappedPorts) > 0 {
		log.SpanLog(ctx, log.DebugLevelInfra, "Add Proxy for dind", "ports", appInst.MappedPorts)
		err = proxy.CreateNginxProxy(ctx, client,
			dockermgmt.GetContainerName(&app.Key),
			cloudcommon.IPAddrAllInterfaces,
			masterIP,
			appInst.MappedPorts,
			app.SkipHcPorts,
			proxy.WithDockerNetwork(network),
			proxy.WithDockerPublishPorts())
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "cannot add proxy", "appName", names.AppName, "ports", appInst.MappedPorts)
			return err
		}
	}

	// Add crm local replace variables
	deploymentVars := crmutil.DeploymentReplaceVars{
		Deployment: crmutil.CrmReplaceVars{
			ClusterIp:    masterIP,
			CloudletName: k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Name),
			ClusterName:  k8smgmt.NormalizeName(clusterInst.Key.ClusterKey.Name),
			CloudletOrg:  k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Organization),
			AppOrg:       k8smgmt.NormalizeName(app.Key.Organization),
		},
	}
	ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

	if DeploymentType == cloudcommon.DeploymentTypeKubernetes {
		err = k8smgmt.CreateAppInst(ctx, nil, client, names, app, appInst)
		if err == nil {
			err = k8smgmt.WaitForAppInst(ctx, client, names, app, k8smgmt.WaitRunning)
		}
	} else if DeploymentType == cloudcommon.DeploymentTypeHelm {
		err = k8smgmt.CreateHelmAppInst(ctx, client, names, clusterInst, app, appInst)
	} else {
		err = fmt.Errorf("invalid deployment type %s for dind", DeploymentType)
	}
	if err != nil {
		proxy.DeleteNginxProxy(ctx, client, dockermgmt.GetContainerName(&app.Key))
		log.SpanLog(ctx, log.DebugLevelInfra, "error creating dind app")
		return err
	}
	return nil
}

func (s *Platform) CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, privacyPolicy *edgeproto.PrivacyPolicy, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "call runKubectlCreateApp for dind")
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}
	if err = s.CreateAppInstInternal(ctx, clusterInst, app, appInst, names); err != nil {
		return err
	}
	cluster, err := FindCluster(names.ClusterName)
	if err != nil {
		s.DeleteAppInst(ctx, clusterInst, app, appInst, updateCallback)
		return err
	}
	masterIP := cluster.MasterAddr
	err = s.patchDindSevice(ctx, names, masterIP)
	if err != nil {
		s.DeleteAppInst(ctx, clusterInst, app, appInst, updateCallback)
		return err
	}
	return nil
}

func (s *Platform) DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	var err error
	client := &pc.LocalClient{}
	DeploymentType := app.Deployment
	// Support for local docker appInst
	if DeploymentType == cloudcommon.DeploymentTypeDocker {
		log.SpanLog(ctx, log.DebugLevelInfra, "run docker delete app for dind")
		err = dockermgmt.DeleteAppInst(ctx, nil, client, app, appInst)
		if err != nil {
			return fmt.Errorf("DeleteAppInst error for docker %v", err)
		}
		return nil
	}
	// Now for helm and k8s apps
	log.SpanLog(ctx, log.DebugLevelInfra, "run kubectl delete app for dind")
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}

	if DeploymentType == cloudcommon.DeploymentTypeKubernetes {
		err = k8smgmt.DeleteAppInst(ctx, client, names, app, appInst)
	} else if DeploymentType == cloudcommon.DeploymentTypeHelm {
		err = k8smgmt.DeleteHelmAppInst(ctx, client, names, clusterInst)
	} else {
		err = fmt.Errorf("invalid deployment type %s for dind", DeploymentType)
	}
	if err != nil {
		return err
	}

	if len(appInst.MappedPorts) > 0 {
		log.SpanLog(ctx, log.DebugLevelInfra, "DeleteNginxProxy for dind")
		if err = proxy.DeleteNginxProxy(ctx, client, dockermgmt.GetContainerName(&app.Key)); err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "cannot delete proxy", "name", names.AppName)
			return err
		}
	}
	return nil
}

func (s *Platform) UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateAppInst for dind")
	client := &pc.LocalClient{}
	DeploymentType := app.Deployment
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
			ClusterIp:    cluster.MasterAddr,
			CloudletName: k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Name),
			ClusterName:  k8smgmt.NormalizeName(clusterInst.Key.ClusterKey.Name),
			CloudletOrg:  k8smgmt.NormalizeName(clusterInst.Key.CloudletKey.Organization),
			AppOrg:       k8smgmt.NormalizeName(app.Key.Organization),
		},
	}
	ctx = context.WithValue(ctx, crmutil.DeploymentReplaceVarsKey, &deploymentVars)

	if DeploymentType == cloudcommon.DeploymentTypeKubernetes {
		return k8smgmt.UpdateAppInst(ctx, nil, client, names, app, appInst)
	} else if DeploymentType == cloudcommon.DeploymentTypeHelm {
		return k8smgmt.UpdateHelmAppInst(ctx, client, names, app, appInst)
	}
	return fmt.Errorf("UpdateAppInst not supported for deployment: %s", DeploymentType)
}

func (s *Platform) GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	clientType := cloudcommon.GetAppClientType(app)
	client, err := s.GetClusterPlatformClient(ctx, clusterInst, clientType)
	if err != nil {
		return nil, err
	}
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return nil, err
	}

	return k8smgmt.GetAppInstRuntime(ctx, client, names, app, appInst)
}

func (s *Platform) patchDindSevice(ctx context.Context, kubeNames *k8smgmt.KubeNames, ipaddr string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Patch DIND service", "kubeNames", kubeNames, "ipaddr", ipaddr)

	client := &pc.LocalClient{}

	for _, serviceName := range kubeNames.ServiceNames {

		cmd := fmt.Sprintf(`%s kubectl patch svc %s -p '{"spec":{"externalIPs":["%s"]}}'`, kubeNames.KconfEnv, serviceName, ipaddr)
		out, err := client.Output(cmd)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "patch svc failed",
				"servicename", serviceName, "out", out, "err", err)
			return fmt.Errorf("error patching for kubernetes service, %s, %s, %v", cmd, out, err)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "patched externalIPs on service", "service", serviceName, "externalIPs", ipaddr)
	}
	return nil
}

func (s *Platform) GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return k8smgmt.GetContainerCommand(ctx, clusterInst, app, appInst, req)
}

func (s *Platform) GetConsoleUrl(ctx context.Context, app *edgeproto.App) (string, error) {
	return "", nil
}

func (s *Platform) SetPowerState(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return nil
}
