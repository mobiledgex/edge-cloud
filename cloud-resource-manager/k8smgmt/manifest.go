package k8smgmt

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	yaml "github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

const AppConfigEnvYaml = "envVarsYaml"

const MexAppLabel = "mex-app"

func addEnvVars(ctx context.Context, template *v1.PodTemplateSpec, envVars []v1.EnvVar) {
	// walk the containers and append environment variables to each
	for j, _ := range template.Spec.Containers {
		template.Spec.Containers[j].Env = append(template.Spec.Containers[j].Env, envVars...)
	}
}

func addImagePullSecret(ctx context.Context, template *v1.PodTemplateSpec, secretName string) {
	found := false
	for _, s := range template.Spec.ImagePullSecrets {
		if s.Name == secretName {
			found = true
		}
	}
	if !found {
		var newSecret v1.LocalObjectReference
		newSecret.Name = secretName
		log.SpanLog(ctx, log.DebugLevelMexos, "adding imagePullSecret", "secretName", secretName)
		template.Spec.ImagePullSecrets = append(template.Spec.ImagePullSecrets, newSecret)
	}
}

func addMexLabel(meta *metav1.ObjectMeta, label string) {
	// Add a label so we can lookup the pods created by this
	// deployment. Pods names are used for shell access.
	meta.Labels[MexAppLabel] = label
}

// Merge in all the environment variables into
func MergeEnvVars(ctx context.Context, vaultConfig *vault.Config, kubeManifest string, configs []*edgeproto.ConfigFile, imagePullSecret string) (string, error) {
	var envVars []v1.EnvVar
	var files []string
	var err error

	deploymentVars, varsFound := ctx.Value(crmutil.DeploymentReplaceVarsKey).(*crmutil.DeploymentReplaceVars)
	log.SpanLog(ctx, log.DebugLevelMexos, "MergeEnvVars", "kubeManifest", kubeManifest, "imagePullSecret", imagePullSecret)

	// Walk the Configs in the App and get all the environment variables together
	for _, v := range configs {
		if v.Kind == AppConfigEnvYaml {
			var curVars []v1.EnvVar
			cfg := v.Config
			// Fill in the Deployment Vars passed as a variable through the context
			if varsFound {
				cfg, err = crmutil.ReplaceDeploymentVars(cfg, deploymentVars)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelMexos, "failed to replace Crm variables",
						"EnvVars ", v.Config, "DeploymentVars", deploymentVars, "error", err)
					return "", err
				}
			}

			if err1 := yaml.Unmarshal([]byte(cfg), &curVars); err1 != nil {
				log.SpanLog(ctx, log.DebugLevelMexos, "cannot unmarshal env vars", "kind", v.Kind,
					"config", cfg, "error", err1)
			} else {
				envVars = append(envVars, curVars...)
			}
		}
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "Merging environment variables", "envVars", envVars)
	mf, err := cloudcommon.GetDeploymentManifest(ctx, vaultConfig, kubeManifest)
	if err != nil {
		return mf, err
	}
	// Fill in the Deployment Vars passed as a variable through the context
	if varsFound {
		mf, err = crmutil.ReplaceDeploymentVars(mf, deploymentVars)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMexos, "failed to replace Crm variables",
				"manifest", mf, "DeploymentVars", deploymentVars, "error", err)
			return "", err
		}
	}

	//decode the objects so we can find the container objects, where we'll add the env vars
	objs, _, err := cloudcommon.DecodeK8SYaml(mf)
	if err != nil {
		return kubeManifest, fmt.Errorf("invalid kubernetes deployment yaml, %s", err.Error())
	}

	//walk the objects
	for i, _ := range objs {
		switch obj := objs[i].(type) {
		case *appsv1.Deployment:
			addEnvVars(ctx, &obj.Spec.Template, envVars)
			addMexLabel(&obj.Spec.Template.ObjectMeta, obj.ObjectMeta.Name)
			if imagePullSecret != "" {
				addImagePullSecret(ctx, &obj.Spec.Template, imagePullSecret)
			}
		case *appsv1.DaemonSet:
			addEnvVars(ctx, &obj.Spec.Template, envVars)
			addMexLabel(&obj.Spec.Template.ObjectMeta, obj.ObjectMeta.Name)
			if imagePullSecret != "" {
				addImagePullSecret(ctx, &obj.Spec.Template, imagePullSecret)
			}
		}
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
