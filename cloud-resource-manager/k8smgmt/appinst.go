package k8smgmt

import (
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	appsv1 "k8s.io/api/apps/v1"
)

func CreateAppInst(client pc.PlatformClient, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	mf, err := cloudcommon.GetDeploymentManifest(app.DeploymentManifest)
	if err != nil {
		return err
	}
	mf, err = MergeEnvVars(mf, app.Configs)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "failed to merge env vars", "error", err)
	}
	log.DebugLog(log.DebugLevelMexos, "writing config file", "kubeManifest", mf)
	file := names.AppName + ".yaml"
	err = pc.WriteFile(client, file, mf, "K8s Deployment")
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "running kubectl create ", "file", file)
	cmd := fmt.Sprintf("%s kubectl create -f %s", names.KconfEnv, file)

	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deploying kubernetes app, %s, %v", out, err)
	}

	// TODO: check pod status. For example, above command may
	// succeed but pod state may be ErrImagePull if it can't access the image.

	log.DebugLog(log.DebugLevelMexos, "done kubectl create")
	return nil
}

func DeleteAppInst(client pc.PlatformClient, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	log.DebugLog(log.DebugLevelMexos, "deleting app", "name", names.AppName)
	cmd := fmt.Sprintf("%s kubectl delete -f %s.yaml", names.KconfEnv, names.AppName)
	out, err := client.Output(cmd)
	if err != nil {
		if strings.Contains(string(out), "not found") {
			log.DebugLog(log.DebugLevelMexos, "app not found, cannot delete", "name", names.AppName)
		} else {
			return fmt.Errorf("error deleting kuberknetes app, %s, %s, %s, %v", names.AppName, cmd, out, err)
		}
	}
	log.DebugLog(log.DebugLevelMexos, "deleted deployment", "name", names.AppName)
	return nil
}

func GetAppInstRuntime(client pc.PlatformClient, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	rt := &edgeproto.AppInstRuntime{}
	rt.ContainerIds = make([]string, 0)

	objs, _, err := cloudcommon.DecodeK8SYaml(app.DeploymentManifest)
	if err != nil {
		return nil, err
	}
	for ii, _ := range objs {
		deployment, ok := objs[ii].(*appsv1.Deployment)
		if ok {
			name := deployment.ObjectMeta.Name
			cmd := fmt.Sprintf("%s kubectl get pods -o custom-columns=NAME:.metadata.name --sort-by=.metadata.name --no-headers --selector=%s=%s", names.KconfEnv, MexAppLabel, name)
			out, err := client.Output(cmd)
			if err != nil {
				return nil, fmt.Errorf("error getting kubernetes pods, %s, %s, %s", cmd, out, err.Error())
			}
			lines := strings.Split(out, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				rt.ContainerIds = append(rt.ContainerIds, strings.TrimSpace(line))
			}
		}
	}

	return rt, nil
}

func GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	// If no container specified, pick the first one in the AppInst.
	// Note that some deployments may not require a container id.
	if req.ContainerId == "" {
		if appInst.RuntimeInfo.ContainerIds == nil ||
			len(appInst.RuntimeInfo.ContainerIds) == 0 {
			return "", fmt.Errorf("no containers to run command in")
		}
		req.ContainerId = appInst.RuntimeInfo.ContainerIds[0]
	}
	names, err := GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return "", fmt.Errorf("failed to get kube names, %v", err)
	}
	cmdStr := fmt.Sprintf("%s kubectl exec -it %s -- %s",
		names.KconfEnv, req.ContainerId, req.Command)
	return cmdStr, nil
}
