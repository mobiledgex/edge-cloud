package dockermgmt

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func CreateAppInst(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	image := app.ImagePath
	name := app.Key.Name
	cmd := fmt.Sprintf("docker run -d --restart=unless-stopped --network=host --name=%s %s %s", name, image, app.Command)
	log.DebugLog(log.DebugLevelMexos, "running docker run ", "cmd", cmd)

	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error running app, %s, %v", out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "done docker run ")
	return nil
}

func DeleteAppInst(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	name := app.Key.Name
	cmd := fmt.Sprintf("docker stop %s", name)
	log.DebugLog(log.DebugLevelMexos, "running docker stop ", "cmd", cmd)

	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error stopping docker app, %s, %v", out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "done docker stop")

	cmd = fmt.Sprintf("docker rm %s", name)
	log.DebugLog(log.DebugLevelMexos, "running docker rm ", "cmd", cmd)

	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error removing docker app, %s, %v", out, err)
	}
	return nil
}

func GetAppInstRuntime(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	rt := &edgeproto.AppInstRuntime{}
	rt.ContainerIds = make([]string, 0)
	// currently just one container identified by the appinst uri
	name := app.Key.Name
	rt.ContainerIds = append(rt.ContainerIds, name)
	return rt, nil
}

func GetContainerCommand(clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	// If no container specified, pick the first one in the AppInst.
	// Note that for docker we currently expect just one
	if req.ContainerId == "" {
		if appInst.RuntimeInfo.ContainerIds == nil ||
			len(appInst.RuntimeInfo.ContainerIds) == 0 {
			return "", fmt.Errorf("no containers to run command in")
		}
		req.ContainerId = appInst.RuntimeInfo.ContainerIds[0]
	}
	cmdStr := fmt.Sprintf("docker exec -it %s -- %s", req.ContainerId, req.Command)
	return cmdStr, nil
}
