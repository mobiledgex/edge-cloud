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
var AppDeploymentTypeVM = "vm"
var AppDeploymentTypeHelm = "helm"
var AppDeploymentTypeDockerSwarm = "docker-swarm"

var ValidDeployments = []string{
	AppDeploymentTypeKubernetes,
	AppDeploymentTypeVM,
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
		if deployment == AppDeploymentTypeVM {
			return true
		}
	case edgeproto.ImageType_ImageTypeUnknown:
		if deployment == AppDeploymentTypeHelm {
			return true
		}
	}
	return false
}

func IsValidDeploymentManifest(appDeploymentType string, manifest string) error {
	if appDeploymentType == AppDeploymentTypeVM {
		if strings.HasPrefix(manifest, "#cloud-config") {
			return nil
		}
		return fmt.Errorf("Only cloud-init script support, must start with '#cloud-config'")
	}
	return nil
}

func GetDefaultDeploymentType(imageType edgeproto.ImageType) (string, error) {
	switch imageType {
	case edgeproto.ImageType_ImageTypeDocker:
		return AppDeploymentTypeKubernetes, nil
	case edgeproto.ImageType_ImageTypeQCOW:
		return AppDeploymentTypeVM, nil
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
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("bad response from remote manifest %d", resp.StatusCode)
	}
	manifestBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(manifestBytes), nil
}
