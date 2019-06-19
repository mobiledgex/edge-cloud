package dockermgmt

import (
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

func getDockerComposeFileName(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) string {
	return util.DNSSanitize("docker-compose-"+app.Key.Name+app.Key.Version) + ".yml"
}

//createDockerComposeFile creates a docker compose file and returns the file name
func createDockerComposeFile(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) (string, error) {
	filename := getDockerComposeFileName(client, app, appInst)
	log.DebugLog(log.DebugLevelMexos, "creating docker compose file", "filename", filename)

	err := pc.WriteFile(client, filename, app.DeploymentManifest, "Docker compose file")
	if err != nil {
		log.InfoLog("Error writing docker compose file", "err", err)
		return "", err
	}
	return filename, nil
}

func CreateAppInst(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	image := app.ImagePath
	name := app.Key.Name
	if app.DeploymentManifest == "" {
		cmd := fmt.Sprintf("docker run -d --restart=unless-stopped --network=host--name=%s %s %s", name, image, app.Command)
		log.DebugLog(log.DebugLevelMexos, "running docker run ", "cmd", cmd)

		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running app, %s, %v", out, err)
		}
		log.DebugLog(log.DebugLevelMexos, "done docker run ")
	} else {
		filename, err := createDockerComposeFile(client, app, appInst)
		if err != nil {
			return err
		}
		cmd := fmt.Sprintf("docker-compose -f %s up -d", filename)
		log.DebugLog(log.DebugLevelMexos, "running docker-compose", "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running docker compose up, %s, %v", out, err)
		}
	}
	return nil
}

func DeleteAppInst(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) error {

	if app.DeploymentManifest == "" {
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
	} else {
		filename := getDockerComposeFileName(client, app, appInst)
		cmd := fmt.Sprintf("docker-compose -f %s down", filename)
		log.DebugLog(log.DebugLevelMexos, "running docker-compose", "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running docker-compose down, %s, %v", out, err)
		}
		err = pc.DeleteFile(client, filename)
		if err != nil {
			log.InfoLog("unable to delete file", "filename", filename, "err", err)
		}
	}

	return nil
}

func GetAppInstRuntime(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	rt := &edgeproto.AppInstRuntime{}
	rt.ContainerIds = make([]string, 0)
	if app.DeploymentManifest == "" {
		//  just one container identified by the appinst uri
		name := app.Key.Name
		rt.ContainerIds = append(rt.ContainerIds, name)
	} else {
		filename := getDockerComposeFileName(client, app, appInst)
		cmd := fmt.Sprintf("docker-compose -f %s images", filename)
		log.DebugLog(log.DebugLevelMexos, "running docker-compose", "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return rt, fmt.Errorf("error running docker compose images, %s, %v", out, err)
		}
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			fs := strings.Fields(line)
			if len(fs) == 6 && fs[0] != "Container" {
				rt.ContainerIds = append(rt.ContainerIds, fs[0])
			}
		}
	}
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
