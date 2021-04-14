package k8smgmt

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

// this is an initial set of supported helm install options
var validHelmInstallOpts = map[string]struct{}{
	"version":  struct{}{},
	"timeout":  struct{}{},
	"wait":     struct{}{},
	"verify":   struct{}{},
	"username": struct{}{},
}

func getHelmOpts(ctx context.Context, client ssh.Client, appName, delims string, configs []*edgeproto.ConfigFile) (string, error) {
	var ymls []string

	deploymentVars, varsFound := ctx.Value(crmutil.DeploymentReplaceVarsKey).(*crmutil.DeploymentReplaceVars)
	// Walk the Configs in the App and generate the yaml files from the helm customization ones
	for ii, v := range configs {
		if v.Kind == edgeproto.AppConfigHelmYaml {
			// config can either be remote, or local
			cfg, err := cloudcommon.GetDeploymentManifest(ctx, nil, v.Config)
			if err != nil {
				return "", err
			}
			// Fill in the Deployment Vars passed as a variable through the context
			if varsFound {
				cfg, err = crmutil.ReplaceDeploymentVars(cfg, delims, deploymentVars)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "failed to replace Crm variables",
						"config file", v.Config, "DeploymentVars", deploymentVars, "error", err)
					return "", err
				}
			}
			file := fmt.Sprintf("%s%d", appName, ii)
			err = pc.WriteFile(client, file, cfg, v.Kind, pc.NoSudo)
			if err != nil {
				return "", err
			}
			ymls = append(ymls, file)
		}
	}
	return getHelmYamlOpt(ymls), nil
}

// helm chart install options are passed as app annotations.
// Example: "version=1.2.2,wait=true,timeout=60" would result in "--version 1.2.2 --wait --timeout 60"
func getHelmInstallOptsString(annotations string) (string, error) {
	outArr := []string{}
	if annotations == "" {
		return "", nil
	}
	// Prevent possible cross-scripting
	invalidChar := strings.IndexAny(annotations, ";`")
	if invalidChar != -1 {
		return "", fmt.Errorf("\"%c\" not allowed in annotations", annotations[invalidChar])
	}
	opts := strings.Split(annotations, ",")
	for _, v := range opts {
		// split by '='
		nameVal := strings.Split(v, "=")
		if len(nameVal) < 2 {
			return "", fmt.Errorf("Invalid annotations string <%s>", annotations)
		}
		// case of "wait=true", true, should not be passed
		if nameVal[1] == "true" {
			nameVal = nameVal[:1]
		} else {
			// make sure that all strings are quoted
			nameVal[1] = strings.TrimSpace(nameVal[1])
			if _, err := strconv.ParseFloat(nameVal[1], 64); err != nil {
				nameVal[1] = strconv.Quote(nameVal[1])
			}
		}
		nameVal[0] = strings.TrimSpace(nameVal[0])
		// validate that the option is one of the supported ones
		if _, found := validHelmInstallOpts[nameVal[0]]; !found {
			return "", fmt.Errorf("Invalid install option passed <%s>", nameVal[0])
		}
		// prepend '--' to the flag
		nameVal[0] = "--" + nameVal[0]
		outArr = append(outArr, nameVal...)
	}
	return strings.Join(outArr, " "), nil
}

// helm chart repositories are encoded in image path
// There are two types of charts:
//   - standard: "stable/prometheus-operator" which come from the default repo
//   - external: "https://resources.gigaspaces.com/helm-charts:gigaspaces/insightedge"
//      - repo name is "gigaspaces" and path is "https://resources.gigaspaces.com/helm-charts"
func getHelmRepoAndChart(imagePath string) (string, string, error) {
	var chart = ""
	// scheme + host + first part of path gives repo path
	chartUrl, err := url.Parse(imagePath)
	if err != nil {
		return "", "", err
	}
	sepIndex := strings.IndexByte(chartUrl.Path, ':')
	if sepIndex < 0 {
		chart = chartUrl.Path
	} else {
		// split path into path, and chart
		chart = chartUrl.Path[sepIndex+1:]
		chartUrl.Path = chartUrl.Path[0:sepIndex]
	}

	chartParts := strings.Split(chart, "/")
	if len(chartParts) != 2 {
		return "", "", fmt.Errorf("Could not parse the chart: <%s>", imagePath)
	}

	if chartUrl.Hostname() != "" {
		return chartParts[0] + " " + chartUrl.String(), chart, nil
	}
	return "", chart, nil
}

func CreateHelmAppInst(ctx context.Context, client ssh.Client, names *KubeNames, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "create kubernetes helm app", "clusterInst", clusterInst, "kubeNames", names)

	// install helm if it's not installed yet
	cmd := fmt.Sprintf("%s helm version", names.KconfEnv)
	out, err := client.Output(cmd)
	if err != nil {
		err = InstallHelm(ctx, client, names)
		if err != nil {
			return err
		}
	}

	// get helm repository config for the app
	helmRepo, chart, err := getHelmRepoAndChart(app.ImagePath)
	if err != nil {
		return err
	}
	// Need to add helm repository first
	if helmRepo != "" {
		cmd = fmt.Sprintf("%s helm repo add %s", names.KconfEnv, helmRepo)
		out, err = client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error adding helm repo, %s, %s, %v", cmd, out, err)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "added helm repository")
	}
	helmArgs, err := getHelmInstallOptsString(app.Annotations)
	if err != nil {
		return err
	}
	configs := append(app.Configs, appInst.Configs...)
	helmOpts, err := getHelmOpts(ctx, client, names.AppName, app.TemplateDelimiter, configs)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Helm options", "helmOpts", helmOpts)
	cmd = fmt.Sprintf("%s helm install %s %s --name %s %s", names.KconfEnv, chart, helmArgs,
		names.HelmAppName, helmOpts)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deploying helm chart, %s, %s, %v", cmd, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "applied helm chart")
	return nil
}

func UpdateHelmAppInst(ctx context.Context, client ssh.Client, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "update kubernetes helm app", "app", app, "kubeNames", names)
	helmArgs, err := getHelmInstallOptsString(app.Annotations)
	if err != nil {
		return err
	}
	configs := append(app.Configs, appInst.Configs...)
	helmOpts, err := getHelmOpts(ctx, client, names.AppName, app.TemplateDelimiter, configs)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Helm options", "helmOpts", helmOpts, "helmArgs", helmArgs)
	cmd := fmt.Sprintf("%s helm upgrade %s %s %s %s", names.KconfEnv, helmArgs, helmOpts, names.HelmAppName, names.AppImage)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error updating helm chart, %s, %s, %v", cmd, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "updated helm chart")
	return nil
}

func DeleteHelmAppInst(ctx context.Context, client ssh.Client, names *KubeNames, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "delete kubernetes helm app")
	cmd := fmt.Sprintf("%s helm delete --purge %s", names.KconfEnv, names.HelmAppName)
	out, err := client.Output(cmd)
	if err != nil {
		if !strings.Contains(out, "not found") {
			return fmt.Errorf("error deleting helm chart, %s, %s, %v", cmd, out, err)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "Unable to find the chart, continue", "cmd", cmd,
			"out", out, "err", err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "removed helm chart")
	return nil
}

func InstallHelm(ctx context.Context, client ssh.Client, names *KubeNames) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "installing helm into cluster", "kconf", names.KconfName)

	// Add service account for tiller
	cmd := fmt.Sprintf("%s kubectl create serviceaccount --namespace kube-system tiller", names.KconfEnv)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error creating tiller service account, %s, %s, %v", cmd, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "setting service acct", "kconf", names.KconfName)

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
	log.SpanLog(ctx, log.DebugLevelInfra, "helm tiller initialized")
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
