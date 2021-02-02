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
	ssh "github.com/mobiledgex/golang-ssh"
	yaml "github.com/mobiledgex/yaml/v2"
)

var createZip = "createZip"
var deleteZip = "deleteZip"

var UseInternalPortInContainer = "internalPort"
var UsePublicPortInContainer = "publicPort"

type DockerNetworkingMode string

var DockerHostMode DockerNetworkingMode = "hostMode"
var DockerBridgeMode DockerNetworkingMode = "bridgeMode"

type DockerOptions struct {
	ForceImagePull bool
}

type DockerReqOp func(do *DockerOptions) error

func WithForceImagePull(force bool) DockerReqOp {
	return func(d *DockerOptions) error {
		d.ForceImagePull = force
		return nil
	}
}

var EnvoyProxy = "envoy"
var NginxProxy = "nginx"

func GetContainerName(appKey *edgeproto.AppKey) string {
	return util.DNSSanitize(appKey.Name + appKey.Version)
}

// Helper function that generates the ports string for docker command
// Example : "-p 80:80/http -p 7777:7777/tcp"
func GetDockerPortString(ports []dme.AppPort, containerPortType string, proxyMatch, listenIP string) []string {
	var cmdArgs []string
	// ensure envoy and nginx docker commands are only opening the udp ports they are managing, not all of the apps udp ports
	for _, p := range ports {
		if p.Proto == dme.LProto_L_PROTO_UDP {
			if proxyMatch == EnvoyProxy && p.Nginx {
				continue
			} else if proxyMatch == NginxProxy && !p.Nginx {
				continue
			}
		}
		proto, err := edgeproto.LProtoStr(p.Proto)
		if err != nil {
			continue
		}
		publicPortStr := fmt.Sprintf("%d", p.PublicPort)
		if p.EndPort != 0 && p.EndPort != p.PublicPort {
			publicPortStr = fmt.Sprintf("%d-%d", p.PublicPort, p.EndPort)
		}
		containerPort := p.PublicPort
		if containerPortType == UseInternalPortInContainer {
			containerPort = p.InternalPort
		}
		containerPortStr := fmt.Sprintf("%d", containerPort)
		if p.EndPort != 0 && p.EndPort != containerPort {
			containerPortStr = fmt.Sprintf("%d-%d", containerPort, p.EndPort)
		}
		listenIPStr := ""
		if listenIP != "" {
			listenIPStr = listenIP + ":"
		}
		pstr := fmt.Sprintf("%s%s:%s/%s", listenIPStr, publicPortStr, containerPortStr, proto)
		cmdArgs = append(cmdArgs, "-p", pstr)
	}
	return cmdArgs
}

func getDockerComposeFileName(client ssh.Client, app *edgeproto.App, appInst *edgeproto.AppInst) string {
	return util.DNSSanitize("docker-compose-"+app.Key.Name+app.Key.Version) + ".yml"
}

func parseDockerComposeManifest(client ssh.Client, dir string, dm *cloudcommon.DockerManifest) error {
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

func handleDockerZipfile(ctx context.Context, authApi cloudcommon.RegistryAuthApi, client ssh.Client, app *edgeproto.App, appInst *edgeproto.AppInst, action string, opts ...DockerReqOp) error {
	var dockerOpt DockerOptions
	for _, op := range opts {
		if err := op(&dockerOpt); err != nil {
			return err
		}
	}
	dir := util.DockerSanitize(app.Key.Name + app.Key.Organization + app.Key.Version)
	filename := dir + "/manifest.zip"
	log.SpanLog(ctx, log.DebugLevelInfra, "docker zip", "filename", filename, "action", action)
	var dockerComposeCommand string

	if action == createZip {
		dockerComposeCommand = "up -d"

		// create a directory for the app and its files
		err := pc.CreateDir(ctx, client, dir, pc.Overwrite)
		if err != nil {
			return err
		}
		passParams := ""
		auth, err := authApi.GetRegistryAuth(ctx, app.DeploymentManifest)
		if err != nil {
			return err
		}
		if auth != nil {
			switch auth.AuthType {
			case cloudcommon.BasicAuth:
				passParams = fmt.Sprintf("--user %s --password %s", auth.Username, auth.Password)
			case cloudcommon.ApiKeyAuth:
				passParams = fmt.Sprintf(`--header="X-JFrog-Art-Api: %s"`, auth.ApiKey)
			case cloudcommon.NoAuth:
			default:
				log.SpanLog(ctx, log.DebugLevelApi, "warning, cannot get registry credentials from vault - unknown authtype", "authType", auth.AuthType)
			}
		}
		// pull the zipfile
		_, err = client.Output(fmt.Sprintf("wget %s -T 60 -P %s %s", passParams, dir, app.DeploymentManifest))
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "wget err", "err", err)
			return fmt.Errorf("wget of app zipfile failed: %v", err)
		}
		s := strings.Split(app.DeploymentManifest, "/")
		zipfile := s[len(s)-1]
		cmd := "unzip -o -d " + dir + " " + dir + "/" + zipfile
		log.SpanLog(ctx, log.DebugLevelInfra, "running unzip", "cmd", cmd)
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
			log.SpanLog(ctx, log.DebugLevelInfra, "found file", "file", f)
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
		log.SpanLog(ctx, log.DebugLevelInfra, "error in parsing docker manifest", "dir", dir, "err", err)
		// for create this is fatal, for delete keep going and cleanup what we can
		if action == createZip {
			return err
		}
	}
	if len(dm.DockerComposeFiles) == 0 && action == createZip {
		return fmt.Errorf("no docker compose files in manifest: %v", err)
	}
	for _, d := range dm.DockerComposeFiles {
		if action == createZip && dockerOpt.ForceImagePull {
			log.SpanLog(ctx, log.DebugLevelInfra, "forcing image pull", "file", d)
			pullcmd := fmt.Sprintf("docker-compose -f %s/%s %s", dir, d, "pull")
			out, err := client.Output(pullcmd)
			if err != nil {
				return fmt.Errorf("error pulling image for docker-compose file: %s, %s, %v", d, out, err)
			}
		}

		cmd := fmt.Sprintf("docker-compose -f %s/%s %s", dir, d, dockerComposeCommand)
		log.SpanLog(ctx, log.DebugLevelInfra, "running docker-compose", "cmd", cmd)
		out, err := client.Output(cmd)

		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "error running docker compose", "out", out, "err", err)
			// for create this is fatal, for delete keep going and cleanup what we can
			if action == createZip {
				return fmt.Errorf("error running docker compose, %s, %v", out, err)
			}
		}
	}

	//cleanup the directory on delete
	if action == deleteZip {
		log.SpanLog(ctx, log.DebugLevelInfra, "deleting app dir", "dir", dir)
		err := pc.DeleteDir(ctx, client, dir, pc.SudoOn)
		if err != nil {
			return fmt.Errorf("error deleting dir, %v", err)
		}
	}
	return nil

}

//createDockerComposeFile creates a docker compose file and returns the file name
func createDockerComposeFile(client ssh.Client, app *edgeproto.App, appInst *edgeproto.AppInst) (string, error) {
	filename := getDockerComposeFileName(client, app, appInst)
	log.DebugLog(log.DebugLevelInfra, "creating docker compose file", "filename", filename)

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
func CreateAppInstLocal(client ssh.Client, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	image := app.ImagePath
	nameLabelVal := util.DNSSanitize(app.Key.Name)
	versionLabelVal := util.DNSSanitize(app.Key.Version)
	name := GetContainerName(&app.Key)
	cloudlet := util.DockerSanitize(appInst.ClusterInstKey().CloudletKey.Name)
	cluster := util.DockerSanitize(appInst.ClusterInstKey().Organization + "-" + appInst.ClusterInstKey().ClusterKey.Name)
	base_cmd := "docker run "
	if appInst.OptRes == "gpu" {
		base_cmd += "--gpus all"
	}

	if app.DeploymentManifest == "" {
		cmd := fmt.Sprintf("%s -d -l edge-cloud -l cloudlet=%s -l cluster=%s  -l %s=%s -l %s=%s --restart=unless-stopped --name=%s %s %s %s", base_cmd,
			cloudlet, cluster, cloudcommon.MexAppNameLabel, nameLabelVal, cloudcommon.MexAppVersionLabel, versionLabelVal, name,
			strings.Join(GetDockerPortString(appInst.MappedPorts, UseInternalPortInContainer, "", cloudcommon.IPAddrAllInterfaces), " "), image, app.Command)
		log.DebugLog(log.DebugLevelInfra, "running docker run ", "cmd", cmd)

		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running app, %s, %v", out, err)
		}
		log.DebugLog(log.DebugLevelInfra, "done docker run ")
	} else {
		filename, err := createDockerComposeFile(client, app, appInst)
		if err != nil {
			return err
		}
		// TODO - missing a label for the metaAppInst label.
		// There is a feature request in docker for it - https://github.com/docker/compose/issues/6159
		// Once that's merged we can add label here too
		// cmd := fmt.Sprintf("docker-compose -f %s -l %s=%s up -d", filename, cloudcommon.MexAppInstanceLabel, labelVal)
		cmd := fmt.Sprintf("docker-compose -f %s up -d", filename)
		log.DebugLog(log.DebugLevelInfra, "running docker-compose", "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running docker compose up, %s, %v", out, err)
		}
	}
	return nil
}

func CreateAppInst(ctx context.Context, authApi cloudcommon.RegistryAuthApi, client ssh.Client, app *edgeproto.App, appInst *edgeproto.AppInst, opts ...DockerReqOp) error {
	var dockerOpt DockerOptions
	for _, op := range opts {
		if err := op(&dockerOpt); err != nil {
			return err
		}
	}
	image := app.ImagePath
	nameLabelVal := util.DNSSanitize(app.Key.Name)
	versionLabelVal := util.DNSSanitize(app.Key.Version)
	base_cmd := "docker run "
	if appInst.OptRes == "gpu" {
		base_cmd += "--gpus all"
	}

	if app.DeploymentManifest == "" {
		if dockerOpt.ForceImagePull {
			log.SpanLog(ctx, log.DebugLevelInfra, "forcing image pull", "image", image)
			pullcmd := "docker image pull " + image
			out, err := client.Output(pullcmd)
			if err != nil {
				return fmt.Errorf("error pulling docker image: %s, %s, %v", image, out, err)
			}
		}
		cmd := fmt.Sprintf("%s -d -l %s=%s -l %s=%s --restart=unless-stopped --network=host --name=%s %s %s", base_cmd,
			cloudcommon.MexAppNameLabel, nameLabelVal, cloudcommon.MexAppVersionLabel,
			versionLabelVal, GetContainerName(&app.Key), image, app.Command)
		log.SpanLog(ctx, log.DebugLevelInfra, "running docker run ", "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running docker run, %s, %v", out, err)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "done docker run ")
	} else {
		if strings.HasSuffix(app.DeploymentManifest, ".zip") {
			return handleDockerZipfile(ctx, authApi, client, app, appInst, createZip, opts...)
		}
		filename, err := createDockerComposeFile(client, app, appInst)
		if err != nil {
			return err
		}
		if dockerOpt.ForceImagePull {
			log.SpanLog(ctx, log.DebugLevelInfra, "forcing image pull", "filename", filename)
			pullcmd := fmt.Sprintf("docker-compose -f %s pull", filename)
			out, err := client.Output(pullcmd)
			if err != nil {
				return fmt.Errorf("error pulling image for docker-compose file: %s, %s, %v", filename, out, err)
			}
		}
		// TODO - missing a label for the metaAppInst label.
		// There is a feature request in docker for it - https://github.com/docker/compose/issues/6159
		// Once that's merged we can add label here too
		// cmd := fmt.Sprintf("docker-compose -f %s -l %s=%s up -d", filename, cloudcommon.MexAppInstanceLabel, labelVal)
		cmd := fmt.Sprintf("docker-compose -f %s up -d", filename)
		log.SpanLog(ctx, log.DebugLevelInfra, "running docker-compose", "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error running docker compose up, %s, %v", out, err)
		}
	}
	return nil
}

func DeleteAppInst(ctx context.Context, authApi cloudcommon.RegistryAuthApi, client ssh.Client, app *edgeproto.App, appInst *edgeproto.AppInst) error {

	if app.DeploymentManifest == "" {
		name := GetContainerName(&app.Key)
		cmd := fmt.Sprintf("docker stop %s", name)

		log.SpanLog(ctx, log.DebugLevelInfra, "running docker stop ", "cmd", cmd)
		removeContainer := true
		out, err := client.Output(cmd)

		if err != nil {
			if strings.Contains(out, "No such container") {
				log.SpanLog(ctx, log.DebugLevelInfra, "container already removed", "cmd", cmd)
				removeContainer = false
			} else {
				return fmt.Errorf("error stopping docker app, %s, %v", out, err)
			}
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "done docker stop", "out", out, "err", err)

		if removeContainer {
			cmd = fmt.Sprintf("docker rm %s", name)
			log.SpanLog(ctx, log.DebugLevelInfra, "running docker rm ", "cmd", cmd)
			out, err := client.Output(cmd)
			if err != nil {
				return fmt.Errorf("error removing docker app, %s, %v", out, err)
			}
		}
	} else {
		if strings.HasSuffix(app.DeploymentManifest, ".zip") {
			return handleDockerZipfile(ctx, authApi, client, app, appInst, deleteZip)
		}
		filename := getDockerComposeFileName(client, app, appInst)
		cmd := fmt.Sprintf("docker-compose -f %s down", filename)
		log.SpanLog(ctx, log.DebugLevelInfra, "running docker-compose", "cmd", cmd)
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

func UpdateAppInst(ctx context.Context, authApi cloudcommon.RegistryAuthApi, client ssh.Client, app *edgeproto.App, appInst *edgeproto.AppInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateAppInst", "appkey", app.Key, "ImagePath", app.ImagePath)

	err := DeleteAppInst(ctx, authApi, client, app, appInst)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "DeleteAppInst failed, proceeding with create", "appkey", app.Key, "err", err)
	}
	return CreateAppInst(ctx, authApi, client, app, appInst, WithForceImagePull(true))
}

func appendContainerIdsFromDockerComposeImages(client ssh.Client, dockerComposeFile string, rt *edgeproto.AppInstRuntime) error {
	cmd := fmt.Sprintf("docker-compose -f %s images", dockerComposeFile)
	log.DebugLog(log.DebugLevelInfra, "running docker-compose", "cmd", cmd)
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

func GetAppInstRuntime(ctx context.Context, client ssh.Client, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	rt := &edgeproto.AppInstRuntime{}
	rt.ContainerIds = make([]string, 0)

	// try to get the container names from the runtime environment
	cmd := `docker ps --format "{{.Names}}"`
	out, err := client.Output(cmd)
	if err == nil {
		for _, name := range strings.Split(out, "\n") {
			name = strings.TrimSpace(name)
			rt.ContainerIds = append(rt.ContainerIds, name)
		}
		return rt, nil
	} else {
		log.SpanLog(ctx, log.DebugLevelInfo, "GetAppInstRuntime cmd failed", "cmd", cmd, "err", err)
	}

	// get the expected names if couldn't get it from the runtime
	if app.DeploymentManifest == "" {
		//  just one container identified by the appinst uri
		name := GetContainerName(&app.Key)
		rt.ContainerIds = append(rt.ContainerIds, name)
	} else {
		if strings.HasSuffix(app.DeploymentManifest, ".zip") {

			var dm cloudcommon.DockerManifest
			dir := util.DockerSanitize(app.Key.Name + app.Key.Organization + app.Key.Version)
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
			return "", fmt.Errorf("no containers found for AppInst, please specify one")
		}
		for _, name := range appInst.RuntimeInfo.ContainerIds {
			// prefer non-nginx/envoy container
			if !strings.HasPrefix(name, "nginx") && !strings.HasPrefix(name, "envoy") {
				req.ContainerId = name
				break
			}
		}
		if req.ContainerId == "" {
			req.ContainerId = appInst.RuntimeInfo.ContainerIds[0]
		}
	}
	if req.Cmd != nil {
		cmdStr := fmt.Sprintf("docker exec -it %s %s", req.ContainerId, req.Cmd.Command)
		return cmdStr, nil
	}
	if req.Log != nil {
		cmdStr := "docker logs "
		if req.Log.Since != "" {
			cmdStr += fmt.Sprintf("--since %s ", req.Log.Since)
		}
		if req.Log.Tail != 0 {
			cmdStr += fmt.Sprintf("--tail %d ", req.Log.Tail)
		}
		if req.Log.Timestamps {
			cmdStr += "--timestamps "
		}
		if req.Log.Follow {
			cmdStr += "--follow "
		}
		cmdStr += req.ContainerId
		return cmdStr, nil
	}
	return "", fmt.Errorf("no command or log specified with exec request")
}
