package k8s

import (
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

const AppConfigHelmYaml = "hemlCustomizationYaml"

func CreateHelmAppInst(client pc.PlatformClient, names *KubeNames, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	log.DebugLog(log.DebugLevelMexos, "create kubernetes helm app", "clusterInst", clusterInst, "kubeNames", names)

	// install helm if it's not installed yet
	cmd := fmt.Sprintf("%s helm version", names.KconfEnv)
	out, err := client.Output(cmd)
	if err != nil {
		err = InstallHelm(client, names)
		if err != nil {
			return err
		}
	}

	// Walk the Configs in the App and generate the yaml files from the helm customization ones
	var ymls []string
	for ii, v := range app.Configs {
		if v.Kind == AppConfigHelmYaml {
			file := fmt.Sprintf("%s%d", names.AppName, ii)
			err := pc.WriteFile(client, file, v.Config, v.Kind)
			if err != nil {
				return err
			}
			ymls = append(ymls, file)
		}
	}
	helmOpts := getHelmYamlOpt(ymls)
	log.DebugLog(log.DebugLevelMexos, "Helm options", "helmOpts", helmOpts)
	cmd = fmt.Sprintf("%s helm install %s --name %s %s", names.KconfEnv, names.AppImage, names.AppName, helmOpts)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deploying helm chart, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "applied helm chart")
	return nil
}

func DeleteHelmAppInst(client pc.PlatformClient, names *KubeNames, clusterInst *edgeproto.ClusterInst) error {
	log.DebugLog(log.DebugLevelMexos, "delete kubernetes helm app")
	cmd := fmt.Sprintf("%s helm delete --purge %s", names.KconfEnv, names.AppName)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deleting helm chart, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "removed helm chart")
	return nil
}

func InstallHelm(client pc.PlatformClient, names *KubeNames) error {
	log.DebugLog(log.DebugLevelMexos, "installing helm into cluster", "kconf", names.KconfName)

	// Add service account for tiller
	cmd := fmt.Sprintf("%s kubectl create serviceaccount --namespace kube-system tiller", names.KconfEnv)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error creating tiller service account, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "setting service acct", "kconf", names.KconfName)

	cmd = fmt.Sprintf("%s kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller", names.KconfEnv)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error creating role binding, %s, %s, %v", cmd, out, err)
	}

	cmd = fmt.Sprintf("%s helm init --wait --service-account tiller", names.KconfEnv)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error initializing tiller for app, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "helm tiller initialized")
	return nil
}

// concatenate files with a ',' and prepend '-f'
// Example: ["foo.yaml", "bar.yaml", "foobar.yaml"] ---> "-f foo.yaml,bar.yaml,foobar.yaml"
func getHelmYamlOpt(ymls []string) string {
	// empty string
	if len(ymls) == 0 {
		return ""
	}
	return "-f " + strings.Join(ymls, ",")
}
