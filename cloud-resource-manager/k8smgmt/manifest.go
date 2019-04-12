package k8smgmt

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	yaml "github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

const (
	AppConfigEnvYaml  = "envVarsYaml"
	K8smgmtConfigVars = "k8smgmtConfig"
)

type K8sMgmtConfig struct {
	ReplicaScalingPolicy string `yaml:"replicaScalingPolicy"`
}

func GetKubeNodeCount(client pc.PlatformClient, names *KubeNames) (int, error) {
	log.DebugLog(log.DebugLevelMexos, "fetching kubectl node count (excluding master)")
	cmd := fmt.Sprintf("%s kubectl get nodes -l node-role.kubernetes.io/master!= --no-headers | wc -l", names.KconfEnv)

	out, err := client.Output(cmd)
	if err != nil {
		return 0, fmt.Errorf("error fetching kubernetes node count, %s, %v", out, err)
	}
	count, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 0, fmt.Errorf("error parsing kubernetes node count, %s, %v", out, err)
	}
	return count, nil
}

// Merge in all the "configs" settings into manifest file
func MergeConfigsVars(client pc.PlatformClient, names *KubeNames, kubeManifest string, configs []*edgeproto.ConfigFile) (string, error) {
	var envVars []v1.EnvVar
	var files []string
	setConfig := false
	replicas := 0

	//quick bail, if nothing to do
	if len(configs) == 0 {
		return kubeManifest, nil
	}

	// Walk the Configs in the App and get all the environment variables together
	for _, v := range configs {
		if v.Kind == AppConfigEnvYaml {
			var curVars []v1.EnvVar
			if err1 := yaml.Unmarshal([]byte(v.Config), &curVars); err1 != nil {
				log.DebugLog(log.DebugLevelMexos, "cannot unmarshal env vars", "kind", v.Kind,
					"config", v.Config, "error", err1)
			} else {
				envVars = append(envVars, curVars...)
				setConfig = true
			}
		} else if v.Kind == K8smgmtConfigVars {
			k8sconfig := K8sMgmtConfig{}
			if err := yaml.Unmarshal([]byte(v.Config), &k8sconfig); err != nil {
				log.DebugLog(log.DebugLevelMexos, "cannot unmarshal k8smgmt config", "kind", v.Kind,
					"config", v.Config, "error", err)
			} else {
				if k8sconfig.ReplicaScalingPolicy == "match-cluster-nodes" {
					replicas, err = GetKubeNodeCount(client, names)
					if err != nil {
						log.DebugLog(log.DebugLevelMexos, "k8mgmt config error", "error", err)
					}
					setConfig = true
				}
			}
		}
	}

	//nothing to do if no variables to merge
	if !setConfig {
		return kubeManifest, nil
	}

	log.DebugLog(log.DebugLevelMexos, "merge configs vars into k8s manifest file")
	mf, err := cloudcommon.GetDeploymentManifest(kubeManifest)
	if err != nil {
		return mf, err
	}
	//decode the objects so we can find the container objects, where we'll add the env vars
	objs, _, err := cloudcommon.DecodeK8SYaml(mf)
	if err != nil {
		return kubeManifest, fmt.Errorf("invalid kubernetes deployment yaml, %s", err.Error())
	}

	//walk the objects
	for i, _ := range objs {
		//make sure we are working on the Deployment object
		deployment, ok := objs[i].(*appsv1.Deployment)
		if !ok {
			continue
		}
		if len(envVars) > 0 {
			log.DebugLog(log.DebugLevelMexos, "Merging environment variables", "envVars", envVars)
			//walk the containers and append environment variables to each
			for i, _ := range deployment.Spec.Template.Spec.Containers {
				deployment.Spec.Template.Spec.Containers[i].Env =
					append(deployment.Spec.Template.Spec.Containers[i].Env, envVars...)
			}
		}

		if replicas > 0 {
			log.DebugLog(log.DebugLevelMexos, "Set replicas to match number of cluster nodes",
				"replicas", replicas)
			nreplicas := int32(replicas)
			deployment.Spec.Replicas = &nreplicas
		}
		//there should only be one deployment object, so break out of the loop
		break
	}
	//marshal the objects back together and return as one string
	printer := &printers.YAMLPrinter{}
	for _, o := range objs {
		buf := bytes.Buffer{}
		err := printer.PrintObj(o, &buf)
		if err != nil {
			return kubeManifest, fmt.Errorf("unable to marshal the k8s objects back together, %s", err.Error())
		} else {
			files = append(files, buf.String())
		}
	}
	mf = strings.Join(files, "---\n")
	return mf, nil
}
