package k8s

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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
	log.DebugLog(log.DebugLevelMexos, "done kubectl create")
	return nil
}

func DeleteAppInst(client pc.PlatformClient, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	cmd := fmt.Sprintf("%s kubectl delete -f %s.yaml", names.KconfEnv, names.AppName)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deleting kuberknetes app, %s, %s, %s, %v", names.AppName, cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "deleted deployment", "name", names.AppName)
	return nil
}
