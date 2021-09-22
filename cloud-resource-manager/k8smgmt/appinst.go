package k8smgmt

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const WaitDeleted string = "WaitDeleted"
const WaitRunning string = "WaitRunning"

const DefaultNamespace string = "default"

// This is half of the default controller AppInst timeout
var maxWait = 15 * time.Minute

// max time waiting for a load balancer ip
var maxLoadBalancerIPWait = 2 * time.Minute

// How long to wait on create if there are no resources
var createWaitNoResources = 10 * time.Second

var applyManifest = "apply"
var createManifest = "create"

var podStateRegString = "(\\S+)\\s+\\d+\\/\\d+\\s+(\\S+)\\s+\\d+\\s+\\S+"
var podStateReg = regexp.MustCompile(podStateRegString)

func LbServicePortToString(p *v1.ServicePort) string {
	proto := p.Protocol
	port := p.Port
	return edgeproto.ProtoPortToString(string(proto), port)
}

func CheckPodsStatus(ctx context.Context, client ssh.Client, kConfEnv, namespace, selector, waitFor string, startTimer time.Time) (bool, error) {
	done := false
	log.SpanLog(ctx, log.DebugLevelInfra, "check pods status", "namespace", namespace, "selector", selector)
	cmd := fmt.Sprintf("%s kubectl get pods --no-headers -n %s --selector=%s", kConfEnv, namespace, selector)
	out, err := client.Output(cmd)
	if err != nil {
		log.InfoLog("error getting pods", "err", err, "out", out)
		return done, fmt.Errorf("error getting pods: %v", err)
	}
	lines := strings.Split(out, "\n")
	// there are potentially multiple pods in the lines loop, we will quit processing this obj
	// only when they are all up, i.e. no non-
	podCount := 0
	runningCount := 0

	for _, line := range lines {
		if line == "" {
			continue
		}
		// there can be multiple pods, one per line. If all
		// of them are running we can quit the loop
		if podStateReg.MatchString(line) {
			podCount++
			matches := podStateReg.FindStringSubmatch(line)
			podName := matches[1]
			podState := matches[2]
			switch podState {
			case "Running":
				log.SpanLog(ctx, log.DebugLevelInfra, "pod is running", "podName", podName)
				runningCount++
			case "Pending":
				fallthrough
			case "ContainerCreating":
				fallthrough
			case "CreateContainerConfigError": // this can be a transient state for some deployments
				log.SpanLog(ctx, log.DebugLevelInfra, "still waiting for pod", "podName", podName, "state", podState)
			case "Terminating":
				log.SpanLog(ctx, log.DebugLevelInfra, "pod is terminating", "podName", podName, "state", podState)
			default:
				if strings.Contains(podState, "Init") {
					// Init state cannot be matched exactly, e.g. Init:0/2
					log.SpanLog(ctx, log.DebugLevelInfra, "pod in init state", "podName", podName, "state", podState)
				} else {
					// try to find out what error was
					// TODO: pull events and send
					// them back as status updates
					// rather than sending back
					// full "describe" dump
					cmd := fmt.Sprintf("%s kubectl describe pod -n %s --selector=%s", kConfEnv, namespace, selector)
					out, derr := client.Output(cmd)
					if derr == nil {
						return done, fmt.Errorf("Run container failed, pod state: %s - %s", podState, out)
					}
					return done, fmt.Errorf("Pod is unexpected state: %s", podState)
				}
			}
		} else if strings.Contains(line, "No resources found") {
			// If creating, pods may not have taken
			// effect yet. If deleting, may already
			// be removed.
			if waitFor == WaitRunning && time.Since(startTimer) > createWaitNoResources {
				return done, fmt.Errorf("no resources found for %s on create: %s", createWaitNoResources, line)
			}
			break
		} else {
			return done, fmt.Errorf("unable to parse kubectl output: [%s]", line)
		}
	}
	if waitFor == WaitDeleted {
		if podCount == 0 {
			log.SpanLog(ctx, log.DebugLevelInfra, "all pods gone", "selector", selector)
			done = true
		}
	} else {
		if podCount == runningCount {
			log.SpanLog(ctx, log.DebugLevelInfra, "all pods up", "selector", selector)
			done = true
		}
	}
	return done, nil
}

// WaitForAppInst waits for pods to either start or result in an error if WaitRunning specified,
// or if WaitDeleted is specified then wait for them to all disappear.
func WaitForAppInst(ctx context.Context, client ssh.Client, names *KubeNames, app *edgeproto.App, waitFor string) error {
	// wait half as long as the total controller wait time, which includes all tasks
	log.SpanLog(ctx, log.DebugLevelInfra, "waiting for appinst pods", "appName", app.Key.Name, "maxWait", maxWait, "waitFor", waitFor)
	start := time.Now()

	// it might be nicer to pull the state directly rather than parsing it, but the states displayed
	// are a combination of states and reasons, e.g. ErrImagePull is not actually a state, so it's
	// just easier to parse the summarized output from kubectl which combines states and reasons
	objs, _, err := cloudcommon.DecodeK8SYaml(app.DeploymentManifest)
	if err != nil {
		log.InfoLog("unable to decode k8s yaml", "err", err)
		return err
	}
	var name string
	for ii, _ := range objs {
		for {
			name = ""
			namespace := ""
			switch obj := objs[ii].(type) {
			case *appsv1.Deployment:
				name = obj.ObjectMeta.Name
				namespace = obj.ObjectMeta.Namespace
			case *appsv1.DaemonSet:
				name = obj.ObjectMeta.Name
				namespace = obj.ObjectMeta.Namespace
			case *appsv1.StatefulSet:
				name = obj.ObjectMeta.Name
				namespace = obj.ObjectMeta.Namespace
			}
			if name == "" {
				break
			}
			if namespace == "" {
				if names.MultitenantNamespace != "" {
					namespace = names.MultitenantNamespace
				} else if names.VirtualClusterNamespace != "" {
					namespace = names.VirtualClusterNamespace
				} else {
					namespace = DefaultNamespace
				}
			}
			selector := fmt.Sprintf("%s=%s", MexAppLabel, name)
			done, err := CheckPodsStatus(ctx, client, names.KconfEnv, namespace, selector, waitFor, start)
			if err != nil {
				return err
			}
			if done {
				break
			}
			elapsed := time.Since(start)
			if elapsed >= (maxWait) {
				// for now we will return no errors when we time out.  In future we will use some other state or status
				// field to reflect this and employ health checks to track these appinsts
				log.InfoLog("AppInst wait timed out", "appName", app.Key.Name)
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

func UpdateLoadBalancerPortMap(ctx context.Context, client ssh.Client, names *KubeNames, portMap map[string]string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateLoadBalancerPortMap", "names", names)

	services, err := GetServices(ctx, client, names)
	if err != nil {
		return err
	}
	for _, s := range services {
		lbip := ""
		for _, ing := range s.Status.LoadBalancer.Ingress {
			if strings.Contains(ing.IP, "pending") || ing.IP == "" {
				continue
			}
			lbip = ing.IP
			break
		}
		if lbip == "" {
			continue
		}
		ports := s.Spec.Ports
		for _, p := range ports {
			portString := LbServicePortToString(&p)
			portMap[portString] = lbip
		}
	}
	return nil
}

func PopulateAppInstLoadBalancerIps(ctx context.Context, client ssh.Client, names *KubeNames, appinst *edgeproto.AppInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "PopulateAppInstLoadBalancerIps", "appInst", appinst.Key.String(), "maxLoadBalancerIPWait", maxLoadBalancerIPWait)
	appinst.InternalPortToLbIp = make(map[string]string)
	start := time.Now()
	for {
		err := UpdateLoadBalancerPortMap(ctx, client, names, appinst.InternalPortToLbIp)
		if err != nil {
			return err
		}
		allPortsHaveIp := true
		// see if all services have an LB IP and update
		for _, mappedPort := range appinst.MappedPorts {
			portString, err := edgeproto.AppInternalPortToString(&mappedPort)
			if err != nil {
				return err
			}
			lbip, ok := appinst.InternalPortToLbIp[portString]
			if ok {
				log.SpanLog(ctx, log.DebugLevelInfra, "found load balancer ip for port", "portString", portString, "lbip", lbip)
				appinst.InternalPortToLbIp[portString] = lbip
			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "did not find load balancer ip for port", "portString", portString)
				allPortsHaveIp = false
			}
		}
		if allPortsHaveIp {
			log.SpanLog(ctx, log.DebugLevelInfra, "All ports successfully got external IPS")
			return nil
		} else {
			elapsed := time.Since(start)
			if elapsed >= (maxLoadBalancerIPWait) {
				log.SpanLog(ctx, log.DebugLevelInfra, "AppInst service lbip wait timed out", "appInst", appinst.Key.String())
				return fmt.Errorf("Timed out waiting for Load Balancer IPs for appinst")
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "Not all ports have external IPs, wait and try again")
			time.Sleep(time.Second * 1)
		}
	}
}

func getConfigDirName(names *KubeNames) (string, string) {
	dir := names.ClusterName
	if names.MultitenantNamespace != "" {
		dir += "." + names.MultitenantNamespace
	}
	return dir, names.AppName + names.AppOrg + names.AppVersion + ".yaml"
}

func CreateDeveloperDefinedNamespaces(ctx context.Context, client ssh.Client, names *KubeNames) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateDeveloperDefinedNamespaces")
	for _, n := range names.DeveloperDefinedNamespaces {
		if n == DefaultNamespace {
			continue
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "Creating Namespace", "name", n)
		cmd := fmt.Sprintf("kubectl create namespace %s --kubeconfig=%s", n, names.KconfName)
		out, err := client.Output(cmd)
		if err != nil {
			if strings.Contains(out, "AlreadyExists") {
				log.SpanLog(ctx, log.DebugLevelInfra, "namespace already exists")
			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "kubectl create namespace failed", "out", string(out), "err", err)
				return fmt.Errorf("kubectl create namespace failed - %v", err)
			}
		}
	}
	return nil
}

func createOrUpdateAppInst(ctx context.Context, authApi cloudcommon.RegistryAuthApi, client ssh.Client, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst, appInstFlavor *edgeproto.Flavor, action string) error {
	if action == createManifest && names.MultitenantNamespace != "" {
		err := CreateNamespace(ctx, client, names)
		if err != nil {
			return err
		}
	}

	mf, err := cloudcommon.GetDeploymentManifest(ctx, authApi, app.DeploymentManifest)
	if err != nil {
		return err
	}
	if names.MultitenantNamespace != "" {
		// Mulit-tenant cluster, add network policy
		np, err := GetNetworkPolicy(ctx, app, appInst, names)
		if err != nil {
			return err
		}
		mf = AddManifest(mf, np)
	}
	mf, err = MergeEnvVars(ctx, authApi, app, mf, names.ImagePullSecrets, names, appInstFlavor)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to merge env vars", "error", err)
		return fmt.Errorf("error merging environment variables config file: %s", err)
	}
	configDir, configName := getConfigDirName(names)
	err = pc.CreateDir(ctx, client, configDir, pc.NoOverwrite)
	if err != nil {
		return err
	}
	file := configDir + "/" + configName
	log.SpanLog(ctx, log.DebugLevelInfra, "writing config file", "file", file, "kubeManifest", mf)
	err = pc.WriteFile(client, file, mf, "K8s Deployment", pc.NoSudo)
	if err != nil {
		return err
	}
	// Kubernetes provides 3 styles of object management.
	// We use the Declarative Object configuration style, to be able to
	// update and prune.
	// Note that "kubectl create" does NOT fall under this style.
	// Only "apply" and "delete" should be used. All configuration files
	// for an AppInst must be stored in their own directory.

	// Selector selects which objects to consider for pruning.
	// Previously we used "--all", but that ends up deleting extra stuff,
	// especially in the case of multi-tenant clusters. We want to
	// transition to using the config label, but we can't use it for
	// old configs that didn't have it. So we will have to continue
	// to use "--all" for old stuff until those old configs eventually
	// get removed naturally over time.
	selector := "--all"
	if names.MultitenantNamespace != "" {
		selector = fmt.Sprintf("-l %s=%s", ConfigLabel, getConfigLabel(names))
	}
	cmd := fmt.Sprintf("%s kubectl apply -f %s --prune %s", names.KconfEnv, configDir, selector)
	log.SpanLog(ctx, log.DebugLevelInfra, "running kubectl", "action", action, "cmd", cmd)
	out, err := client.Output(cmd)
	if err != nil && strings.Contains(string(out), `pruning nonNamespaced object /v1, Kind=Namespace: namespaces "kube-system" is forbidden: this namespace may not be deleted`) {
		// odd error that occurs on Azure, probably due to some system
		// object they have in their k8s cluster setup. Ignore it
		// since it doesn't affect the other aspects of the apply.
		err = nil
	}
	if err != nil {
		return fmt.Errorf("error running kubectl command %s: %s, %v", cmd, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "done kubectl", "action", action)
	return nil

}

func CreateAppInst(ctx context.Context, authApi cloudcommon.RegistryAuthApi, client ssh.Client, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst, appInstFlavor *edgeproto.Flavor) error {
	return createOrUpdateAppInst(ctx, authApi, client, names, app, appInst, appInstFlavor, createManifest)
}

func UpdateAppInst(ctx context.Context, authApi cloudcommon.RegistryAuthApi, client ssh.Client, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst, appInstFlavor *edgeproto.Flavor) error {
	err := createOrUpdateAppInst(ctx, authApi, client, names, app, appInst, appInstFlavor, applyManifest)
	if err != nil {
		return err
	}
	return WaitForAppInst(ctx, client, names, app, WaitRunning)
}

func DeleteAppInst(ctx context.Context, client ssh.Client, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	configDir, configName := getConfigDirName(names)
	file := configDir + "/" + configName
	cmd := fmt.Sprintf("%s kubectl delete -f %s", names.KconfEnv, file)
	log.SpanLog(ctx, log.DebugLevelInfra, "deleting app", "name", names.AppName, "cmd", cmd)
	out, err := client.Output(cmd)
	if err != nil {
		if strings.Contains(string(out), "not found") {
			log.SpanLog(ctx, log.DebugLevelInfra, "app not found, cannot delete", "name", names.AppName)
		} else {
			return fmt.Errorf("error deleting kuberknetes app, %s, %s, %s, %v", names.AppName, cmd, out, err)
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "deleted deployment", "name", names.AppName)
	// remove manifest file since directory contains all AppInst manifests for
	// the ClusterInst.
	log.SpanLog(ctx, log.DebugLevelInfra, "remove app manifest", "name", names.AppName, "file", file)
	out, err = client.Output("rm " + file)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "error deleting manifest", "file", file, "out", string(out), "err", err)
	}
	//Note wait for deletion of appinst can be done here in a generic place, but wait for creation is split
	// out in each platform specific task so that we can optimize the time taken for create by allowing the
	// wait to be run in parallel with other tasks
	err = WaitForAppInst(ctx, client, names, app, WaitDeleted)
	if err != nil {
		return err
	}
	if names.MultitenantNamespace != "" {
		// clean up namespace
		if err = DeleteNamespace(ctx, client, names); err != nil {
			return err
		}
	}
	return nil
}

func GetAppInstRuntime(ctx context.Context, client ssh.Client, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	rt := &edgeproto.AppInstRuntime{}
	rt.ContainerIds = make([]string, 0)

	objs, _, err := cloudcommon.DecodeK8SYaml(app.DeploymentManifest)
	if err != nil {
		return nil, err
	}
	var name string
	for ii, _ := range objs {
		name = ""
		namespace := ""
		switch obj := objs[ii].(type) {
		case *appsv1.Deployment:
			name = obj.ObjectMeta.Name
			namespace = obj.ObjectMeta.Namespace
		case *appsv1.DaemonSet:
			name = obj.ObjectMeta.Name
			namespace = obj.ObjectMeta.Namespace
		case *appsv1.StatefulSet:
			name = obj.ObjectMeta.Name
			namespace = obj.ObjectMeta.Namespace
		}
		if name == "" {
			continue
		}
		if namespace == "" {
			if names.MultitenantNamespace != "" {
				namespace = names.MultitenantNamespace
			} else if names.VirtualClusterNamespace != "" {
				namespace = names.VirtualClusterNamespace
			} else {
				namespace = DefaultNamespace
			}
		}
		// Get list of all running pods.
		// NOTE: Parsing status from json output doesn't give correct value as observed with kubectl version 1.18
		//       Hence, look at table output and then get list of running pods and use this to fetch container names
		cmd := fmt.Sprintf("%s kubectl get pods -n %s --no-headers --sort-by=.metadata.name --selector=%s=%s "+
			"| awk '{if ($3 == \"Running\") print $1}'",
			names.KconfEnv, namespace, MexAppLabel, name)
		out, err := client.Output(cmd)
		if err != nil {
			return nil, fmt.Errorf("error getting kubernetes pods, %s, %s, %s", cmd, out, err.Error())
		}
		podNames := strings.Split(out, "\n")
		for _, podName := range podNames {
			podName := strings.TrimSpace(podName)
			if podName == "" {
				continue
			}
			// Returns list of pods and its containers in the format: "<PodName>/<ContainerName>"
			cmd = fmt.Sprintf("%s kubectl get pod %s -n %s -o json | jq -r '.spec.containers[] | .name'",
				names.KconfEnv, podName, namespace)
			out, err = client.Output(cmd)
			if err != nil {
				return nil, fmt.Errorf("error getting kubernetes pod %q containers, %s, %s, %s", podName, cmd, out, err.Error())
			}
			contNames := strings.Split(out, "\n")
			for _, contName := range contNames {
				contName = strings.TrimSpace(contName)
				if contName == "" {
					continue
				}
				rt.ContainerIds = append(rt.ContainerIds, namespace+"/"+podName+"/"+contName)
			}
		}
	}

	return rt, nil
}

func GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetContainerCommand", "app", app, "containerId", req.ContainerId)

	// If no container specified, pick the first one in the AppInst.
	// Note that some deployments may not require a container id.
	if req.ContainerId == "" {
		if appInst.RuntimeInfo.ContainerIds == nil ||
			len(appInst.RuntimeInfo.ContainerIds) == 0 {
			return "", fmt.Errorf("no containers to run command in")
		}
		req.ContainerId = appInst.RuntimeInfo.ContainerIds[0]
	}
	podName := ""
	containerName := ""
	namespace := DefaultNamespace
	parts := strings.Split(req.ContainerId, "/")
	if len(parts) == 1 {
		// old way
		podName = parts[0]
	} else if len(parts) == 2 {
		// new way
		podName = parts[0]
		containerName = parts[1]
	} else if len(parts) == 3 {
		// namespace also included
		namespace = parts[0]
		podName = parts[1]
		containerName = parts[2]
	} else {
		return "", fmt.Errorf("invalid containerID, expected to be of format <namespace>/<PodName>/<ContainerName>")
	}
	names, err := GetKubeNames(clusterInst, app, appInst)
	if err != nil {
		return "", fmt.Errorf("failed to get kube names, %v", err)
	}
	if req.Cmd != nil {
		containerCmd := ""
		if containerName != "" {
			containerCmd = fmt.Sprintf("-c %s ", containerName)
		}
		cmdStr := fmt.Sprintf("%s kubectl exec -n %s -it %s%s -- %s",
			names.KconfEnv, namespace, containerCmd, podName, req.Cmd.Command)
		return cmdStr, nil
	}
	if req.Log != nil {
		cmdStr := fmt.Sprintf("%s kubectl logs -n %s ", names.KconfEnv, namespace)
		if req.Log.Since != "" {
			_, perr := time.ParseDuration(req.Log.Since)
			if perr == nil {
				cmdStr += fmt.Sprintf("--since=%s ", req.Log.Since)
			} else {
				cmdStr += fmt.Sprintf("--since-time=%s ", req.Log.Since)
			}
		}
		if req.Log.Tail != 0 {
			cmdStr += fmt.Sprintf("--tail=%d ", req.Log.Tail)
		}
		if req.Log.Timestamps {
			cmdStr += "--timestamps=true "
		}
		if req.Log.Follow {
			cmdStr += "-f "
		}
		cmdStr += podName
		if containerName != "" {
			cmdStr += " -c " + containerName
		} else {
			cmdStr += " --all-containers"
		}
		return cmdStr, nil
	}
	return "", fmt.Errorf("no command or log specified with the exec request")
}

var namespaceTemplate = template.Must(template.New("namespace").Parse(`apiVersion: v1
kind: Namespace
metadata:
  name: {{.MultitenantNamespace}}
  labels:
    name: {{.MultitenantNamespace}}
`))

func CreateNamespace(ctx context.Context, client ssh.Client, names *KubeNames) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "creating namespace", "name", names.MultitenantNamespace)
	buf := bytes.Buffer{}
	err := namespaceTemplate.Execute(&buf, names)
	if err != nil {
		return err
	}
	file := names.MultitenantNamespace + ".yaml"
	err = pc.WriteFile(client, file, buf.String(), "namespace", pc.NoSudo)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("kubectl create -f %s --kubeconfig=%s", file, names.BaseKconfName)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("Error in creating namespace: %s - %v", out, err)
	}
	// copy the kubeconfig
	log.SpanLog(ctx, log.DebugLevelInfra, "create new kubeconfig for cluster namespace", "clusterKubeConfig", names.BaseKconfName, "namespaceKubeConfig", names.KconfName)
	err = pc.CopyFile(client, names.BaseKconfName, names.KconfName)
	if err != nil {
		return fmt.Errorf("Failed to create new kubeconfig: %v", err)
	}
	// set the new kubeconfig to use the namespace
	cmd = fmt.Sprintf("KUBECONFIG=%s kubectl config set-context --current --namespace=%s", names.KconfName, names.MultitenantNamespace)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("Error in setting new namespace context: %s - %v", out, err)
	}
	return nil
}

func DeleteNamespace(ctx context.Context, client ssh.Client, names *KubeNames) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "deleting namespace", "name", names.MultitenantNamespace)
	cmd := fmt.Sprintf("kubectl delete namespace %s --kubeconfig=%s", names.MultitenantNamespace, names.BaseKconfName)
	out, err := client.Output(cmd)
	if err != nil {
		if !strings.Contains(out, "not found") {
			return fmt.Errorf("Error in deleting namespace: %s - %v", out, err)
		}
	}
	// delete namespaced kconf
	err = pc.DeleteFile(client, names.KconfName)
	if err != nil {
		// just log the error
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to clean up namespaced kconf", "err", err)
	}
	return nil
}
