package dockermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	yaml "github.com/mobiledgex/yaml/v2"
)

var createZip = "createZip"
var deleteZip = "deleteZip"

var UseInternalPortInContainer = "internalPort"
var UsePublicPortInContainer = "publicPort"

// Helper function that generates the ports string for docker command
// Example : "-p 80:80/http -p 7777:7777/tcp"
func GetDockerPortString(ports []dme.AppPort, containerPortType string, protoMatch dme.LProto, listenIP string) []string {
	var cmdArgs []string

	log.InfoLog("XXXXXXX ENTER GetDockerPortString", "ports", ports)

	for _, p := range ports {
		if p.Proto == dme.LProto_L_PROTO_HTTP {
			// L7 not allowed for docker
			continue
		}
		// quick fix to deal with the fact that nginx and envoy both try to listen to a superset of ports.
		// remove this when we go to 100% envoy
		if protoMatch != dme.LProto_L_PROTO_UNKNOWN && protoMatch != p.Proto {
			continue
		}
		proto, err := edgeproto.LProtoStr(p.Proto)
		if err != nil {
			continue
		}
		containerPort := p.PublicPort
		if containerPortType == UseInternalPortInContainer {
			containerPort = p.InternalPort
		}
		listenIPStr := ""
		if listenIP != "" {
			listenIPStr = listenIP + ":"
		}
		pstr := fmt.Sprintf("%s%d:%d/%s", listenIPStr, p.PublicPort, containerPort, proto)
		cmdArgs = append(cmdArgs, "-p", pstr)
	}
	log.InfoLog("XXXXXXX LEAVE GetDockerPortString", "cmdArgs", cmdArgs)

	return cmdArgs
}

func getDockerComposeFileName(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) string {
	return util.DNSSanitize("docker-compose-"+app.Key.Name+app.Key.Version) + ".yml"
}

func parseDockerComposeManifest(client pc.PlatformClient, dir string, dm *cloudcommon.DockerManifest) error {
	cmd := fmt.Sprintf("cat %s/%s", dir, "manifest.yml")
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error cat manifest, %s, %v", out, err)
	}
	err = yaml.Unmarshal([]byte(out), &dm)
	if err != nil {
		return fmt.Errorf("unmarshalling manifest.yml: %v", err)
	}
	return nil
}

func handleDockerZipfile(ctx context.Context, client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst, action string) error {
	dir := util.DockerSanitize(app.Key.Name + app.Key.DeveloperKey.Name + app.Key.Version)
	filename := dir + "/manifest.zip"
	log.SpanLog(ctx, log.DebugLevelMexos, "docker zip", "filename", filename, "action", action)
	var dockerComposeCommand string

	if action == createZip {
		dockerComposeCommand = "up -d"

		//create a directory for the app and its files
		output, err := client.Output("mkdir " + dir)
		if err != nil {
			if !strings.Contains(output, "File exists") {
				log.SpanLog(ctx, log.DebugLevelMexos, "mkdir err", "out", output, "err", err)
				return err
			}
		}
		// pull the zipfile
		_, err = client.Output("wget -T 60 -P " + dir + " " + app.DeploymentManifest)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMexos, "wget err", "err", err)
			return fmt.Errorf("wget of app zipfile failed: %v", err)
		}
		s := strings.Split(app.DeploymentManifest, "/")
		zipfile := s[len(s)-1]
		cmd := "unzip -o -d " + dir + " " + dir + "/" + zipfile
		log.SpanLog(ctx, log.DebugLevelMexos, "running unzip", "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error unzipping, %s, %v", out, err)
		}
		// find the files which were extracted
		cmd = "ls -m " + dir
		out, err = client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running ls, %s, %v", out, err)
		}

		manifestFound := false
		files := strings.Split(out, ",")

		for _, f := range files {

			f = strings.TrimSpace(f)
			log.SpanLog(ctx, log.DebugLevelMexos, "found file", "file", f)
			if f == "manifest.yml" {
				manifestFound = true
			}
		}
		if !manifestFound {
			return fmt.Errorf("no manifest.yml file found in zipfile")
		}
	} else {
		// delete
		dockerComposeCommand = "down"
	}
	// parse the yaml manifest and find the compose files
	var dm cloudcommon.DockerManifest
	err := parseDockerComposeManifest(client, dir, &dm)
	if err != nil {
		return err
	}
	if len(dm.DockerComposeFiles) == 0 {
		return fmt.Errorf("no docker compose files in manifest: %v", err)
	}
	for _, d := range dm.DockerComposeFiles {
		cmd := fmt.Sprintf("docker-compose -f %s/%s %s", dir, d, dockerComposeCommand)
		log.SpanLog(ctx, log.DebugLevelMexos, "running docker-compose", "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running docker compose, %s, %v", out, err)
		}
	}

	//cleanup the directory on delete
	if action == deleteZip {
		log.SpanLog(ctx, log.DebugLevelMexos, "deleting app dir", "dir", dir)
		err := pc.DeleteDir(ctx, client, dir)
		if err != nil {
			return fmt.Errorf("error deleting dir, %v", err)
		}
	}
	return nil

}

//createDockerComposeFile creates a docker compose file and returns the file name
func createDockerComposeFile(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) (string, error) {
	filename := getDockerComposeFileName(client, app, appInst)
	log.DebugLog(log.DebugLevelMexos, "creating docker compose file", "filename", filename)

	err := pc.WriteFile(client, filename, app.DeploymentManifest, "Docker compose file", pc.NoSudo)
	if err != nil {
		log.DebugLog(log.DebugLevelInfo, "Error writing docker compose file", "err", err)
		return "", err
	}
	return filename, nil
}

// Local Docker AppInst create is different due to fact that MacOS doesn't like '--network=host' option.
// Instead on MacOS docker needs to have port mapping  explicity specified with '-p' option.
// As a result we have a separate function specifically for a docker app creation on a MacOS laptop
func CreateAppInstLocal(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	image := app.ImagePath
	name := util.DockerSanitize(app.Key.Name)
	cloudlet := util.DockerSanitize(appInst.Key.ClusterInstKey.CloudletKey.Name)
	cluster := util.DockerSanitize(appInst.Key.ClusterInstKey.Developer + "-" + appInst.Key.ClusterInstKey.ClusterKey.Name)

	if app.DeploymentManifest == "" {
		cmd := fmt.Sprintf("docker run -d -l edge-cloud -l cloudlet=%s -l cluster=%s --restart=unless-stopped --name=%s %s %s %s", cloudlet, cluster, name,
			strings.Join(GetDockerPortString(appInst.MappedPorts, UseInternalPortInContainer, dme.LProto_L_PROTO_UNKNOWN, cloudcommon.IPAddrAllInterfaces), " "), image, app.Command)
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

func CreateAppInst(ctx context.Context, client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	image := app.ImagePath
	name := util.DockerSanitize(app.Key.Name)
	if app.DeploymentManifest == "" {
		cmd := fmt.Sprintf("docker run -d --restart=unless-stopped --network=host --name=%s %s %s", name, image, app.Command)
		if app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
			cmd = fmt.Sprintf("docker run -d -l edge-cloud --restart=unless-stopped --name=%s %s %s %s", name,
				strings.Join(GetDockerPortString(appInst.MappedPorts, UsePublicPortInContainer, dme.LProto_L_PROTO_UNKNOWN, cloudcommon.IPAddrLocalHost), " "), image, app.Command)
		}
		log.SpanLog(ctx, log.DebugLevelMexos, "running docker run ", "cmd", cmd)

		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running app, %s, %v", out, err)
		}
		log.SpanLog(ctx, log.DebugLevelMexos, "done docker run ")
	} else {
		if strings.HasSuffix(app.DeploymentManifest, ".zip") {
			return handleDockerZipfile(ctx, client, app, appInst, createZip)
		}
		filename, err := createDockerComposeFile(client, app, appInst)
		if err != nil {
			return err
		}
		cmd := fmt.Sprintf("docker-compose -f %s up -d", filename)
		log.SpanLog(ctx, log.DebugLevelMexos, "running docker-compose", "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running docker compose up, %s, %v", out, err)
		}
	}
	return nil
}

func DeleteAppInst(ctx context.Context, client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) error {

	if app.DeploymentManifest == "" {
		name := util.DockerSanitize(app.Key.Name)
		cmd := fmt.Sprintf("docker stop %s", name)
		log.SpanLog(ctx, log.DebugLevelMexos, "running docker stop ", "cmd", cmd)

		removeContainer := true
		out, err := client.Output(cmd)
		if err != nil {
			if strings.Contains(out, "No such container") {
				log.SpanLog(ctx, log.DebugLevelMexos, "container already removed", "cmd", cmd)
				removeContainer = false
			} else {
				return fmt.Errorf("error stopping docker app, %s, %v", out, err)
			}
		}
		log.SpanLog(ctx, log.DebugLevelMexos, "done docker stop")

		if removeContainer {
			cmd = fmt.Sprintf("docker rm %s", name)
			log.SpanLog(ctx, log.DebugLevelMexos, "running docker rm ", "cmd", cmd)

			out, err = client.Output(cmd)
			if err != nil {
				return fmt.Errorf("error removing docker app, %s, %v", out, err)
			}
		}
	} else {
		if strings.HasSuffix(app.DeploymentManifest, ".zip") {
			return handleDockerZipfile(ctx, client, app, appInst, deleteZip)
		}
		filename := getDockerComposeFileName(client, app, appInst)
		cmd := fmt.Sprintf("docker-compose -f %s down", filename)
		log.SpanLog(ctx, log.DebugLevelMexos, "running docker-compose", "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running docker-compose down, %s, %v", out, err)
		}
		err = pc.DeleteFile(client, filename)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "unable to delete file", "filename", filename, "err", err)
		}
	}

	return nil
}

func UpdateAppInst(ctx context.Context, client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "UpdateAppInst", "appkey", app.Key, "ImagePath", app.ImagePath)

	err := DeleteAppInst(ctx, client, app, appInst)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "DeleteAppInst failed, proceeding with create", "appkey", app.Key, "err", err)
	}
	return CreateAppInst(ctx, client, app, appInst)
}

func appendContainerIdsFromDockerComposeImages(client pc.PlatformClient, dockerComposeFile string, rt *edgeproto.AppInstRuntime) error {
	cmd := fmt.Sprintf("docker-compose -f %s images", dockerComposeFile)
	log.DebugLog(log.DebugLevelMexos, "running docker-compose", "cmd", cmd)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error running docker compose images, %s, %v", out, err)
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		fs := strings.Fields(line)
		if len(fs) == 6 && fs[0] != "Container" {
			rt.ContainerIds = append(rt.ContainerIds, fs[0])
		}
	}
	return nil
}

func GetAppInstRuntime(client pc.PlatformClient, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	rt := &edgeproto.AppInstRuntime{}
	rt.ContainerIds = make([]string, 0)
	if app.DeploymentManifest == "" {
		//  just one container identified by the appinst uri
		name := util.DockerSanitize(app.Key.Name)
		rt.ContainerIds = append(rt.ContainerIds, name)
	} else {
		if strings.HasSuffix(app.DeploymentManifest, ".zip") {

			var dm cloudcommon.DockerManifest
			dir := util.DockerSanitize(app.Key.Name + app.Key.DeveloperKey.Name + app.Key.Version)
			err := parseDockerComposeManifest(client, dir, &dm)
			if err != nil {
				return rt, err
			}
			for _, d := range dm.DockerComposeFiles {
				err := appendContainerIdsFromDockerComposeImages(client, dir+"/"+d, rt)
				if err != nil {
					return rt, err
				}
			}
		} else {
			filename := getDockerComposeFileName(client, app, appInst)
			err := appendContainerIdsFromDockerComposeImages(client, filename, rt)
			if err != nil {
				return rt, err
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
	cmdStr := fmt.Sprintf("docker exec -it %s %s", req.ContainerId, req.Command)
	return cmdStr, nil
}
