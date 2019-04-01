package k8smgmt

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	yaml "gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

const AppConfigEnvYaml = "envVarsYaml"

// Merge in all the environment variables into
func MergeEnvVars(kubeManifest string, configs []*edgeproto.ConfigFile) (string, error) {
	var envVars []v1.EnvVar
	var files []string
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
			}
		}
	}
	//nothing to do if no variables to merge
	if len(envVars) == 0 {
		return kubeManifest, nil
	}
	log.DebugLog(log.DebugLevelMexos, "Merging environment variables", "envVars", envVars)
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
		//walk the containers and append environment variables to each
		for i, _ := range deployment.Spec.Template.Spec.Containers {
			deployment.Spec.Template.Spec.Containers[i].Env =
				append(deployment.Spec.Template.Spec.Containers[i].Env, envVars...)
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
