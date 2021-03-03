package xind

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/dockermgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type ClusterManager interface {
	GetMasterIp(ctx context.Context, names *k8smgmt.KubeNames) (string, error)
	GetDockerNetworkName(ctx context.Context, names *k8smgmt.KubeNames) (string, error)
}

func (s *Xind) CreateAppInstNoPatch(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) (reterr error) {
	client, err := s.GetClient(ctx)
	if err != nil {
		return err
	}
	DeploymentType := app.Deployment
	// Support for local docker appInst
	if DeploymentType == cloudcommon.DeploymentTypeDocker {
		log.SpanLog(ctx, log.DebugLevelInfra, "run docker create app")
		err = dockermgmt.CreateAppInstLocal(client, app, appInst)
		if err != nil {
			return fmt.Errorf("CreateAppInst error for docker %v", err)
		}
		return nil
	}
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}
	masterIP, err := s.clusterManager.GetMasterIp(ctx, names)
	if err != nil {
		return err
	}
	network, err := s.clusterManager.GetDockerNetworkName(ctx, names)
	if err != nil {
		return err
	}

	if len(appInst.MappedPorts) > 0 {
		proxyName := dockermgmt.GetContainerName(&app.Key)
		log.SpanLog(ctx, log.DebugLevelInfra, "Add Proxy", "ports", appInst.MappedPorts, "masterIP", masterIP, "network", network)
		err = proxy.CreateNginxProxy(ctx, client,
			proxyName,
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
		defer func() {
			if reterr == nil {
				return
			}
			undoerr := proxy.DeleteNginxProxy(ctx, client, proxyName)
			log.SpanLog(ctx, log.DebugLevelInfra, "Undo CreateNginxProxy", "undoerr", undoerr)
		}()
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
			if err != nil {
				undoerr := k8smgmt.DeleteAppInst(ctx, client, names, app, appInst)
				log.SpanLog(ctx, log.DebugLevelInfra, "Undo CreateAppInst", "undoerr", undoerr)
			}
		}
	} else if DeploymentType == cloudcommon.DeploymentTypeHelm {
		err = k8smgmt.CreateHelmAppInst(ctx, client, names, clusterInst, app, appInst)
	} else {
		err = fmt.Errorf("invalid deployment type %s", DeploymentType)
	}
	if err != nil {
		return err
	}
	return nil
}

func (s *Xind) CreateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) (reterr error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateAppInst")

	err := s.CreateAppInstNoPatch(ctx, clusterInst, app, appInst, flavor, updateCallback)
	if err != nil {
		return err
	}
	// patch service IP
	err = s.patchServiceIp(ctx, clusterInst, app, appInst)
	if err != nil {
		undoerr := s.DeleteAppInst(ctx, clusterInst, app, appInst, updateCallback)
		log.SpanLog(ctx, log.DebugLevelInfo, "Undo CreateAppInst", "undoerr", undoerr)
		return err
	}
	return nil

}

func (s *Xind) DeleteAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteAppInst")

	var err error
	client, err := s.GetClient(ctx)
	if err != nil {
		return err
	}
	DeploymentType := app.Deployment
	// Support for local docker appInst
	if DeploymentType == cloudcommon.DeploymentTypeDocker {
		log.SpanLog(ctx, log.DebugLevelInfra, "run docker delete app")
		err = dockermgmt.DeleteAppInst(ctx, nil, client, app, appInst)
		if err != nil {
			return fmt.Errorf("DeleteAppInst error for docker %v", err)
		}
		return nil
	}
	// Now for helm and k8s apps
	log.SpanLog(ctx, log.DebugLevelInfra, "run kubectl delete app")
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}

	if DeploymentType == cloudcommon.DeploymentTypeKubernetes {
		err = k8smgmt.DeleteAppInst(ctx, client, names, app, appInst)
	} else if DeploymentType == cloudcommon.DeploymentTypeHelm {
		err = k8smgmt.DeleteHelmAppInst(ctx, client, names, clusterInst)
	} else {
		err = fmt.Errorf("invalid deployment type %s", DeploymentType)
	}
	if err != nil {
		return err
	}

	if len(appInst.MappedPorts) > 0 {
		log.SpanLog(ctx, log.DebugLevelInfra, "DeleteNginxProxy for xind")
		if err = proxy.DeleteNginxProxy(ctx, client, dockermgmt.GetContainerName(&app.Key)); err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "cannot delete proxy", "name", names.AppName)
			return err
		}
	}
	return nil
}

func (s *Xind) UpdateAppInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateAppInst")
	client, err := s.GetClient(ctx)
	if err != nil {
		return err
	}
	DeploymentType := app.Deployment

	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}
	masterIP, err := s.clusterManager.GetMasterIp(ctx, names)
	if err != nil {
		return err
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
		return k8smgmt.UpdateAppInst(ctx, nil, client, names, app, appInst)
	} else if DeploymentType == cloudcommon.DeploymentTypeHelm {
		return k8smgmt.UpdateHelmAppInst(ctx, client, names, app, appInst)
	}
	return fmt.Errorf("UpdateAppInst not supported for deployment: %s", DeploymentType)
}

func (s *Xind) GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
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

func (s *Xind) patchXindSevice(ctx context.Context, kubeNames *k8smgmt.KubeNames, ipaddr string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Patch XIND service", "kubeNames", kubeNames, "ipaddr", ipaddr)

	client, err := s.GetClient(ctx)
	if err != nil {
		return err
	}

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

func (s *Xind) GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return k8smgmt.GetContainerCommand(ctx, clusterInst, app, appInst, req)
}

func (s *Xind) GetConsoleUrl(ctx context.Context, app *edgeproto.App) (string, error) {
	return "", nil
}

func (s *Xind) SetPowerState(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return nil
}

func (s *Xind) patchServiceIp(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	client, err := s.GetClient(ctx)
	if err != nil {
		return err
	}
	names, err := k8smgmt.GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return err
	}
	ipaddr, err := s.clusterManager.GetMasterIp(ctx, names)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Patch service", "kubeNames", names, "ipaddr", ipaddr)
	for _, serviceName := range names.ServiceNames {

		cmd := fmt.Sprintf(`%s kubectl patch svc %s -p '{"spec":{"externalIPs":["%s"]}}'`, names.KconfEnv, serviceName, ipaddr)
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
