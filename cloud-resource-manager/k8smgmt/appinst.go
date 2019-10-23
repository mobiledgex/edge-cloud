package k8smgmt

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	appsv1 "k8s.io/api/apps/v1"
)

const WaitDeleted string = "WaitDeleted"
const WaitRunning string = "WaitRunning"

// This is half of the default controller AppInst timeout
var maxWait = 15 * time.Minute

var applyManifest = "apply"
var createManifest = "create"

// WaitForAppInst waits for pods to either start or result in an error if WaitRunning specified,
// or if WaitDeleted is specified then wait for them to all disappear.
func WaitForAppInst(client pc.PlatformClient, names *KubeNames, app *edgeproto.App, waitFor string) error {
	// wait half as long as the total controller wait time, which includes all tasks
	log.DebugLog(log.DebugLevelMexos, "waiting for appinst pods", "appName", app.Key.Name, "maxWait", maxWait, "waitFor", waitFor)
	start := time.Now()

	// it might be nicer to pull the state directly rather than parsing it, but the states displayed
	// are a combination of states and reasons, e.g. ErrImagePull is not actually a state, so it's
	// just easier to parse the summarized output from kubectl which combines states and reasons
	r, _ := regexp.Compile("(\\S+)\\s+\\d+\\/\\d+\\s+(\\w+)\\s+\\d+\\s+\\S+")
	objs, _, err := cloudcommon.DecodeK8SYaml(app.DeploymentManifest)
	if err != nil {
		log.InfoLog("unable to decode k8s yaml", "err", err)
		return err
	}
	for ii, _ := range objs {
		for {
			deployment, ok := objs[ii].(*appsv1.Deployment)
			if ok {
				name := deployment.ObjectMeta.Name
				log.DebugLog(log.DebugLevelMexos, "get pods", "name", name)

				cmd := fmt.Sprintf("%s kubectl get pods --no-headers --selector=%s=%s", names.KconfEnv, MexAppLabel, name)
				out, err := client.Output(cmd)
				if err != nil {
					log.InfoLog("error getting pods", "err", err, "out", out)
					return fmt.Errorf("error getting pods: %v", err)
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
					if r.MatchString(line) {
						podCount++
						matches := r.FindStringSubmatch(line)
						podName := matches[1]
						podState := matches[2]
						switch podState {
						case "Running":
							log.DebugLog(log.DebugLevelMexos, "pod is running", "podName", podName)
							runningCount++
						case "Pending":
							fallthrough
						case "ContainerCreating":
							log.DebugLog(log.DebugLevelMexos, "still waiting for pod", "podName", podName, "state", podState)
						case "Terminating":
							log.DebugLog(log.DebugLevelMexos, "pod is terminating", "podName", podName, "state", podState)
						default:
							// try to find out what error was
							// TODO: pull events and send
							// them back as status updates
							// rather than sending back
							// full "describe" dump
							cmd := fmt.Sprintf("%s kubectl describe pod --selector=%s=%s", names.KconfEnv, MexAppLabel, name)
							out, derr := client.Output(cmd)
							if derr == nil {
								return fmt.Errorf("Run container failed: %s", out)
							}
							return fmt.Errorf("Pod is unexpected state: %s", podState)
						}
					} else {
						if waitFor == WaitDeleted && strings.Contains(line, "No resources found") {
							break
						}
						return fmt.Errorf("unable to parse kubectl output: [%s]", line)
					}
				}
				if waitFor == WaitDeleted {
					if podCount == 0 {
						log.DebugLog(log.DebugLevelMexos, "all pods gone", "deployment name", deployment.ObjectMeta.Name)
						break
					}
				} else {
					if podCount == runningCount {
						log.DebugLog(log.DebugLevelMexos, "all pods up", "deployment name", deployment.ObjectMeta.Name)
						break
					}
				}
				elapsed := time.Since(start)
				if elapsed >= (maxWait) {
					// for now we will return no errors when we time out.  In future we will use some other state or status
					// field to reflect this and employ health checks to track these appinsts
					log.InfoLog("AppInst wait timed out", "appName", app.Key.Name)
					break
				}
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}
	}
	return nil
}

func createOrUpdateAppInst(client pc.PlatformClient, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst, action string) error {
	mf, err := cloudcommon.GetDeploymentManifest(app.DeploymentManifest)
	if err != nil {
		return err
	}
	mf, err = MergeEnvVars(mf, app.Configs, app.ImagePath)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "failed to merge env vars", "error", err)
	}
	log.DebugLog(log.DebugLevelMexos, "writing config file", "kubeManifest", mf)
	file := names.AppName + names.AppRevision + ".yaml"
	err = pc.WriteFile(client, file, mf, "K8s Deployment", pc.NoSudo)
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "running kubectl", "action", action, "file", file)
	cmd := fmt.Sprintf("%s kubectl %s -f %s", names.KconfEnv, action, file)

	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error running kubectl %s command %s, %v", action, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "done kubectl", "action", action)
	return nil

}

func CreateAppInst(client pc.PlatformClient, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	return createOrUpdateAppInst(client, names, app, appInst, createManifest)
}

func UpdateAppInst(client pc.PlatformClient, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	err := createOrUpdateAppInst(client, names, app, appInst, applyManifest)
	if err != nil {
		return err
	}
	return WaitForAppInst(client, names, app, WaitRunning)
}

func DeleteAppInst(client pc.PlatformClient, names *KubeNames, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	log.DebugLog(log.DebugLevelMexos, "deleting app", "name", names.AppName)
	file := names.AppName + names.AppRevision + ".yaml"
	cmd := fmt.Sprintf("%s kubectl delete -f %s", names.KconfEnv, file)
	out, err := client.Output(cmd)
	if err != nil {
		if strings.Contains(string(out), "not found") {
			log.DebugLog(log.DebugLevelMexos, "app not found, cannot delete", "name", names.AppName)
		} else {
			return fmt.Errorf("error deleting kuberknetes app, %s, %s, %s, %v", names.AppName, cmd, out, err)
		}
	}
	log.DebugLog(log.DebugLevelMexos, "deleted deployment", "name", names.AppName)
	//Note wait for deletion of appinst can be done here in a generic place, but wait for creation is split
	// out in each platform specific task so that we can optimize the time taken for create by allowing the
	// wait to be run in parallel with other tasks
	return WaitForAppInst(client, names, app, WaitDeleted)
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
