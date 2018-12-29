package cloudcommon

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mobiledgex/edge-cloud/deploygen"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var AppDeploymentTypeKubernetes = "kubernetes"
var AppDeploymentTypeKVM = "kvm"
var AppDeploymentTypeHelm = "helm"
var AppDeploymentTypeDockerSwarm = "docker-swarm"

var ValidDeployments = []string{
	AppDeploymentTypeKubernetes,
	AppDeploymentTypeKVM,
	AppDeploymentTypeHelm,
	AppDeploymentTypeDockerSwarm,
}

func IsValidDeploymentType(appDeploymentType string) bool {
	for _, d := range ValidDeployments {
		if appDeploymentType == d {
			return true
		}
	}
	return false
}

func IsValidDeploymentForImage(imageType edgeproto.ImageType, deployment string) bool {
	switch imageType {
	case edgeproto.ImageType_ImageTypeDocker:
		if deployment == AppDeploymentTypeKubernetes { // also later docker
			return true
		}
	case edgeproto.ImageType_ImageTypeQCOW:
		if deployment == AppDeploymentTypeKVM {
			return true
		}
	case edgeproto.ImageType_ImageTypeUnknown:
		if deployment == AppDeploymentTypeHelm {
			return true
		}
	}
	return false
}

func GetDefaultDeploymentType(imageType edgeproto.ImageType) (string, error) {
	switch imageType {
	case edgeproto.ImageType_ImageTypeDocker:
		return AppDeploymentTypeKubernetes, nil
	case edgeproto.ImageType_ImageTypeQCOW:
		return AppDeploymentTypeKVM, nil
	}
	return "", fmt.Errorf("unknown image type %s", imageType)
}

// GetAppDeploymentManifest gets the deployment-specific manifest.
func GetAppDeploymentManifest(app *edgeproto.App) (string, error) {
	if app.DeploymentManifest != "" {
		return GetDeploymentManifest(app.DeploymentManifest)
	} else if app.DeploymentGenerator != "" {
		return GenerateManifest(app)
	} else if app.Deployment == AppDeploymentTypeKubernetes {
		// kubernetes requires a deployment yaml. Use default generator.
		app.DeploymentGenerator = deploygen.KubernetesBasic
		str, err := GenerateManifest(app)
		if err != nil {
			return "", fmt.Errorf("failed to use default deployment generator %s, %s", app.Deployment, err.Error())
		}
		return str, nil
	}
	// no manifest specified
	return "", nil
}

func GetDeploymentManifest(manifest string) (string, error) {
	// manifest may be remote target or inline json/yaml
	if strings.HasPrefix(manifest, "http://") || strings.HasPrefix(manifest, "https://") {
		mf, err := GetRemoteManifest(manifest)
		if err != nil {
			return "", fmt.Errorf("cannot get manifest from %s, %v", manifest, err)
		}
		return mf, nil
	}
	// inline manifest
	return manifest, nil
}

func GenerateManifest(app *edgeproto.App) (string, error) {
	target := app.DeploymentGenerator
	if target == "" {
		return "", fmt.Errorf("no deployment generator specified")
	}
	// generator may be remote target or generator name
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return deploygen.SendReq(target, app)
	} else if _, ok := deploygen.Generators[target]; ok {
		return deploygen.RunGen(target, app)
	}
	return "", fmt.Errorf("invalid deployment generator %s", target)
}

func GetRemoteManifest(target string) (string, error) {
	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	manifestBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(manifestBytes), nil
}
