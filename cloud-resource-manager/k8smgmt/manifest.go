// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8smgmt

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/edgexr/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/edgexr/edge-cloud/cloudcommon"
	"github.com/edgexr/edge-cloud/deploygen"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/log"
	"github.com/edgexr/edge-cloud/util"
	yaml "github.com/mobiledgex/yaml/v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

const MexAppLabel = "mex-app"
const ConfigLabel = "config"

// TestReplacementVars are used to syntax check app envvars
var TestReplacementVars = crmutil.DeploymentReplaceVars{
	Deployment: crmutil.CrmReplaceVars{
		ClusterIp:    "99.99.99.99",
		CloudletName: "dummyCloudlet",
		ClusterName:  "dummyCluster",
		CloudletOrg:  "dummyCloudletOrg",
		AppOrg:       "dummyAppOrg",
		DnsZone:      "dummy.net",
	},
}

func addEnvVars(ctx context.Context, template *v1.PodTemplateSpec, envVars []v1.EnvVar) {
	// walk the containers and append environment variables to each
	for j, _ := range template.Spec.Containers {
		template.Spec.Containers[j].Env = append(template.Spec.Containers[j].Env, envVars...)
	}
}

func addImagePullSecret(ctx context.Context, template *v1.PodTemplateSpec, secretNames []string) {
	for _, secretName := range secretNames {
		found := false
		for _, s := range template.Spec.ImagePullSecrets {
			if s.Name == secretName {
				found = true
				break
			}
		}
		if !found {
			var newSecret v1.LocalObjectReference
			newSecret.Name = secretName
			log.SpanLog(ctx, log.DebugLevelInfra, "adding imagePullSecret", "secretName", secretName)
			template.Spec.ImagePullSecrets = append(template.Spec.ImagePullSecrets, newSecret)
		}
	}
}

func addMexLabel(meta *metav1.ObjectMeta, label string) {
	// Add a label so we can lookup the pods created by this
	// deployment. Pods names are used for shell access.
	meta.Labels[MexAppLabel] = label
}

// Add app details to the deployment as labels
// these labels will be picked up by Pormetheus and added to the metrics
func addAppInstLabels(meta *metav1.ObjectMeta, app *edgeproto.App) {
	meta.Labels[cloudcommon.MexAppNameLabel] = util.DNSSanitize(app.Key.Name)
	meta.Labels[cloudcommon.MexAppVersionLabel] = util.DNSSanitize(app.Key.Version)
}

// The config label marks all objects that are part of config files in the
// config dir. We use this with apply --prune -l config=configlabel to
// only prune objects that were created with the config label, and are no
// longer present in the configDir files.
// Only objects that are created via files in the configDir should have
// the config label. Typically this would be all the AppInsts in the
// Cluster (or namespace for multi-tenant clusters).
func getConfigLabel(names *KubeNames) string {
	if names.MultitenantNamespace != "" {
		return names.MultitenantNamespace
	}
	return names.ClusterName
}

func addResourceLimits(ctx context.Context, template *v1.PodTemplateSpec, config *edgeproto.ServerlessConfig) error {
	// This assumes there's only one container.
	// Kubernetes does not give a way to specify resource limits per pod.
	// It's either per container, or per namespace.
	cpu, err := resource.ParseQuantity(config.Vcpus.DecString())
	if err != nil {
		return err
	}
	mem, err := resource.ParseQuantity(fmt.Sprintf("%dMi", config.Ram))
	if err != nil {
		return err
	}
	for j, _ := range template.Spec.Containers {
		resources := &template.Spec.Containers[j].Resources
		resources.Limits = v1.ResourceList{}
		resources.Limits[v1.ResourceCPU] = cpu
		resources.Limits[v1.ResourceMemory] = mem
		resources.Requests = v1.ResourceList{}
		resources.Requests[v1.ResourceCPU] = cpu
		resources.Requests[v1.ResourceMemory] = mem
	}
	return nil
}

func GetAppEnvVars(ctx context.Context, app *edgeproto.App, authApi cloudcommon.RegistryAuthApi, deploymentVars *crmutil.DeploymentReplaceVars) (*[]v1.EnvVar, error) {
	var envVars []v1.EnvVar
	for _, v := range app.Configs {
		if v.Kind == edgeproto.AppConfigEnvYaml {
			var curVars []v1.EnvVar
			cfg, err := cloudcommon.GetDeploymentManifest(ctx, authApi, v.Config)
			if err != nil {
				return nil, err
			}
			if deploymentVars != nil {
				cfg, err = crmutil.ReplaceDeploymentVars(cfg, app.TemplateDelimiter, deploymentVars)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "failed to replace Crm variables",
						"EnvVars ", v.Config, "DeploymentVars", deploymentVars, "error", err)
					return nil, err
				}
			}
			err = yaml.Unmarshal([]byte(cfg), &curVars)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "cannot unmarshal env vars", "kind", v.Kind,
					"config", cfg, "error", err)
				return nil, fmt.Errorf("cannot unmarshal env vars: %s - %v", cfg, err)
			} else {
				envVars = append(envVars, curVars...)
			}
		}
	}
	return &envVars, nil
}

// Merge in all the environment variables into
func MergeEnvVars(ctx context.Context, authApi cloudcommon.RegistryAuthApi, app *edgeproto.App, kubeManifest string, imagePullSecrets []string, names *KubeNames, appInstFlavor *edgeproto.Flavor) (string, error) {
	var files []string
	log.SpanLog(ctx, log.DebugLevelInfra, "MergeEnvVars", "kubeManifest", kubeManifest)

	deploymentVars, varsFound := ctx.Value(crmutil.DeploymentReplaceVarsKey).(*crmutil.DeploymentReplaceVars)
	log.SpanLog(ctx, log.DebugLevelInfra, "MergeEnvVars", "deploymentVars", deploymentVars, "varsFound", varsFound)
	envVars, err := GetAppEnvVars(ctx, app, authApi, deploymentVars)
	if err != nil {
		return "", err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Merging environment variables", "envVars", envVars)
	mf, err := cloudcommon.GetDeploymentManifest(ctx, authApi, kubeManifest)
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
		log.SpanLog(ctx, log.DebugLevelInfra, "Success to replace Crm variables",
			"manifest", mf, "DeploymentVars", deploymentVars)
	}

	//decode the objects so we can find the container objects, where we'll add the env vars
	objs, _, err := cloudcommon.DecodeK8SYaml(mf)
	if err != nil {
		return kubeManifest, fmt.Errorf("invalid kubernetes deployment yaml, %s", err.Error())
	}

	//walk the objects
	var template *v1.PodTemplateSpec
	var name string
	for i, _ := range objs {
		// convert obj to generic metaObject to set labels on every obj
		var metaObj metav1.Object
		if obj, ok := objs[i].(metav1.Object); ok {
			metaObj = obj
		} else if acc, ok := objs[i].(metav1.ObjectMetaAccessor); ok {
			metaObj = acc.GetObjectMeta()
		}
		if metaObj != nil {
			labels := metaObj.GetLabels()
			if labels == nil {
				labels = make(map[string]string)
			}
			// config label is used to mark for pruning
			labels[ConfigLabel] = getConfigLabel(names)
			metaObj.SetLabels(labels)
		}

		template = nil
		name = ""
		switch obj := objs[i].(type) {
		case *appsv1.Deployment:
			template = &obj.Spec.Template
			name = obj.ObjectMeta.Name
			obj.Spec.Replicas = getDefaultReplicas(app, names, *obj.Spec.Replicas)
		case *appsv1.DaemonSet:
			template = &obj.Spec.Template
			name = obj.ObjectMeta.Name
		case *appsv1.StatefulSet:
			template = &obj.Spec.Template
			name = obj.ObjectMeta.Name
			obj.Spec.Replicas = getDefaultReplicas(app, names, *obj.Spec.Replicas)
		}
		if template == nil {
			continue
		}
		addEnvVars(ctx, template, *envVars)
		addMexLabel(&template.ObjectMeta, name)
		// Add labels for all the appKey data
		addAppInstLabels(&template.ObjectMeta, app)
		if imagePullSecrets != nil {
			addImagePullSecret(ctx, template, imagePullSecrets)
		}
		if names.MultitenantNamespace != "" && app.ServerlessConfig != nil {
			err := addResourceLimits(ctx, template, app.ServerlessConfig)
			if err != nil {
				return "", err
			}
		}
		isGpuFlavor, gpuCount := cloudcommon.IsGPUFlavor(appInstFlavor)
		if isGpuFlavor {
			// just set GPU resource limit for manifest generated by our
			// deployment generator. For custom deployment manifest, developer
			// will specify GPU resource limit if they need it
			if _, ok := template.ObjectMeta.Labels[deploygen.MexDeployGenLabel]; ok {
				gpuCountQty, err := resource.ParseQuantity(fmt.Sprintf("%d", gpuCount))
				if err != nil {
					return "", err
				}
				for j, _ := range template.Spec.Containers {
					// This assumes just one container, as for deploygen
					// generated manifest will only have one container
					resources := &template.Spec.Containers[j].Resources
					if len(resources.Limits) == 0 {
						resources.Limits = v1.ResourceList{}
					}
					resName := v1.ResourceName(cloudcommon.GPUResourceLimitName)
					resources.Limits[resName] = gpuCountQty
				}
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

func AddManifest(mf, addmf string) string {
	if strings.TrimSpace(addmf) == "" {
		return mf
	}
	return mf + "---\n" + addmf
}

func getDefaultReplicas(app *edgeproto.App, names *KubeNames, curReplica int32) *int32 {
	val := int32(1)
	if curReplica != 0 {
		val = curReplica
	}
	if names.MultitenantNamespace != "" && app.ServerlessConfig != nil {
		val = int32(app.ServerlessConfig.MinReplicas)
	}
	return &val
}
