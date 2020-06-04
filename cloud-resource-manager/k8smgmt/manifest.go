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
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/mobiledgex/edge-cloud/vault"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

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
		log.SpanLog(ctx, log.DebugLevelInfra, "adding imagePullSecret", "secretName", secretName)
		template.Spec.ImagePullSecrets = append(template.Spec.ImagePullSecrets, newSecret)
	}
}

func addMexLabel(meta *metav1.ObjectMeta, label string) {
	// Add a label so we can lookup the pods created by this
	// deployment. Pods names are used for shell access.
	meta.Labels[MexAppLabel] = label
}

// Add app details to the deployment as labela
// these labels will be picked up by Pormetheus and added to the metrics
func addAppInstLabels(meta *metav1.ObjectMeta, app *edgeproto.App) {
	meta.Labels[cloudcommon.MexAppNameLabel] = util.DNSSanitize(app.Key.Name)
	meta.Labels[cloudcommon.MexAppVersionLabel] = util.DNSSanitize(app.Key.Version)
}

// Merge in all the environment variables into
func MergeEnvVars(ctx context.Context, vaultConfig *vault.Config, app *edgeproto.App, kubeManifest string, imagePullSecret string) (string, error) {
	var envVars []v1.EnvVar
	var files []string
	var err error

	deploymentVars, varsFound := ctx.Value(crmutil.DeploymentReplaceVarsKey).(*crmutil.DeploymentReplaceVars)
	log.SpanLog(ctx, log.DebugLevelInfra, "MergeEnvVars", "kubeManifest", kubeManifest, "imagePullSecret", imagePullSecret)

	// Walk the Configs in the App and get all the environment variables together
	for _, v := range app.Configs {
		if v.Kind == edgeproto.AppConfigEnvYaml {
			var curVars []v1.EnvVar
			// config can either be remote, or local
			cfg, err := cloudcommon.GetDeploymentManifest(ctx, vaultConfig, v.Config)
			if err != nil {
				return "", err
			}
			// Fill in the Deployment Vars passed as a variable through the context
			if varsFound {
				cfg, err = crmutil.ReplaceDeploymentVars(cfg, app.TemplateDelimiter, deploymentVars)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "failed to replace Crm variables",
						"EnvVars ", v.Config, "DeploymentVars", deploymentVars, "error", err)
					return "", err
				}
			}

			if err1 := yaml.Unmarshal([]byte(cfg), &curVars); err1 != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "cannot unmarshal env vars", "kind", v.Kind,
					"config", cfg, "error", err1)
			} else {
				envVars = append(envVars, curVars...)
			}
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Merging environment variables", "envVars", envVars)
	mf, err := cloudcommon.GetDeploymentManifest(ctx, vaultConfig, kubeManifest)
	if err != nil {
		return mf, err
	}
	// Fill in the Deployment Vars passed as a variable through the context
	if varsFound {
		mf, err = crmutil.ReplaceDeploymentVars(mf, app.TemplateDelimiter, deploymentVars)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "failed to replace Crm variables",
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
			// Add labels for all the appKey data
			addAppInstLabels(&obj.Spec.Template.ObjectMeta, app)
			if imagePullSecret != "" {
				addImagePullSecret(ctx, &obj.Spec.Template, imagePullSecret)
			}
		case *appsv1.DaemonSet:
			addEnvVars(ctx, &obj.Spec.Template, envVars)
			addMexLabel(&obj.Spec.Template.ObjectMeta, obj.ObjectMeta.Name)
			addAppInstLabels(&obj.Spec.Template.ObjectMeta, app)
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
