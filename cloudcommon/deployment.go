package cloudcommon

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/deploygen"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	yaml "github.com/mobiledgex/yaml/v2"
	v1 "k8s.io/api/core/v1"
)

var AppDeploymentTypeKubernetes = "kubernetes"
var AppDeploymentTypeVM = "vm"
var AppDeploymentTypeHelm = "helm"
var AppDeploymentTypeDocker = "docker"

var ValidDeployments = []string{
	AppDeploymentTypeKubernetes,
	AppDeploymentTypeVM,
	AppDeploymentTypeHelm,
	AppDeploymentTypeDocker,
}

type DockerManifest struct {
	DockerComposeFiles []string
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
	case edgeproto.ImageType_IMAGE_TYPE_DOCKER:
		if deployment == AppDeploymentTypeKubernetes || deployment == AppDeploymentTypeDocker {
			return true
		}
	case edgeproto.ImageType_IMAGE_TYPE_QCOW:
		if deployment == AppDeploymentTypeVM {
			return true
		}
	case edgeproto.ImageType_IMAGE_TYPE_HELM:
		if deployment == AppDeploymentTypeHelm {
			return true
		}
	}
	return false
}

func IsValidDeploymentManifest(appDeploymentType, command, manifest string, ports []dme.AppPort) error {
	if appDeploymentType == AppDeploymentTypeVM {
		if command != "" {
			return fmt.Errorf("both deploymentmanifest and command cannot be used together for VM based deployment")
		}
		if strings.HasPrefix(manifest, "#cloud-config") {
			return nil
		}
		return fmt.Errorf("only cloud-init script support, must start with '#cloud-config'")
	} else if appDeploymentType == AppDeploymentTypeKubernetes {
		objs, _, err := DecodeK8SYaml(manifest)
		if err != nil {
			return fmt.Errorf("parse kubernetes deployment yaml failed, %v", err)
		}
		// check that any ports specified on App are part of manifest
		objPorts := make(map[string]struct{})
		for _, obj := range objs {
			ksvc, ok := obj.(*v1.Service)
			if !ok {
				continue
			}
			for _, kp := range ksvc.Spec.Ports {
				appPort := dme.AppPort{}
				appPort.Proto, err = edgeproto.GetLProto(string(kp.Protocol))
				if err != nil {
					log.DebugLog(log.DebugLevelApi, "unrecognized port protocol in kubernetes manifest", "proto", string(kp.Protocol))
					continue
				}
				appPort.InternalPort = int32(kp.TargetPort.IntValue())
				objPorts[appPort.String()] = struct{}{}
			}
		}
		missingPorts := []string{}
		for _, appPort := range ports {
			// http is mapped to tcp
			if appPort.EndPort != 0 {
				// We have a range-port notation on the dme.AppPort
				// while our manifest exhaustively enumerates each as a kubePort
				start := appPort.InternalPort
				end := appPort.EndPort
				for i := start; i <= end; i++ {
					// expand short hand notation to test membership in map
					tp := dme.AppPort{
						Proto:        appPort.Proto,
						InternalPort: int32(i),
						EndPort:      int32(0),
					}
					if appPort.Proto == dme.LProto_L_PROTO_HTTP {
						appPort.Proto = dme.LProto_L_PROTO_TCP
					}

					if _, found := objPorts[tp.String()]; found {
						continue
					}
					protoStr, _ := edgeproto.LProtoStr(appPort.Proto)
					missingPorts = append(missingPorts, fmt.Sprintf("%s:%d", protoStr, tp.InternalPort))
				}
				continue
			}
			if appPort.Proto == dme.LProto_L_PROTO_HTTP {
				appPort.Proto = dme.LProto_L_PROTO_TCP
			}
			if _, found := objPorts[appPort.String()]; found {
				continue
			}
			protoStr, _ := edgeproto.LProtoStr(appPort.Proto)
			missingPorts = append(missingPorts, fmt.Sprintf("%s:%d", protoStr, appPort.InternalPort))
		}
		if len(missingPorts) > 0 {
			return fmt.Errorf("port %s defined in AccessPorts but missing from kubernetes manifest (note http is mapped to tcp)", strings.Join(missingPorts, ","))
		}
	}
	return nil
}

func GetDefaultDeploymentType(imageType edgeproto.ImageType) (string, error) {
	switch imageType {
	case edgeproto.ImageType_IMAGE_TYPE_DOCKER:
		return AppDeploymentTypeKubernetes, nil
	case edgeproto.ImageType_IMAGE_TYPE_QCOW:
		return AppDeploymentTypeVM, nil
	case edgeproto.ImageType_IMAGE_TYPE_HELM:
		return AppDeploymentTypeHelm, nil
	}
	return "", fmt.Errorf("unknown image type %s", imageType)
}

func GetImageTypeForDeployment(deployment string) (edgeproto.ImageType, error) {
	switch deployment {
	case AppDeploymentTypeDocker:
		fallthrough
	case AppDeploymentTypeKubernetes:
		return edgeproto.ImageType_IMAGE_TYPE_DOCKER, nil
	case AppDeploymentTypeHelm:
		return edgeproto.ImageType_IMAGE_TYPE_HELM, nil
	case AppDeploymentTypeVM:
		// could be different formats
		fallthrough
	default:
		return edgeproto.ImageType_IMAGE_TYPE_UNKNOWN, nil
	}
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

func validateRemoteZipManifest(manifest string) error {
	zipfile := "/tmp/temp.zip"
	err := GetRemoteManifestToFile(manifest, zipfile)
	if err != nil {
		return fmt.Errorf("cannot get manifest from %s, %v", manifest, err)
	}
	defer os.Remove(zipfile)
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		return fmt.Errorf("cannot read zipfile from manifest %s, %v", manifest, err)
	}
	defer r.Close()
	foundManifest := false
	var filesInManifest = make(map[string]bool)
	var dm DockerManifest
	for _, f := range r.File {
		filesInManifest[f.Name] = true
		if f.Name == "manifest.yml" {
			foundManifest = true
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("cannot open manifest.yml in zipfile: %v", err)
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(rc)
			rc.Close()
			err = yaml.Unmarshal(buf.Bytes(), &dm)
			if err != nil {
				return fmt.Errorf("unmarshalling manifest.yml: %v", err)
			}
		}
	}
	if !foundManifest {
		return fmt.Errorf("no manifest.yml in zipfile %s", manifest)
	}
	for _, dc := range dm.DockerComposeFiles {
		_, ok := filesInManifest[dc]
		if !ok {
			return fmt.Errorf("docker-compose file specified in manifest but not in zip: %s", dc)
		}
	}
	return nil
}

func GetDeploymentManifest(manifest string) (string, error) {
	// manifest may be remote target or inline json/yaml
	if strings.HasPrefix(manifest, "http://") || strings.HasPrefix(manifest, "https://") {

		if strings.HasSuffix(manifest, ".zip") {
			log.DebugLog(log.DebugLevelApi, "zipfile manifest found", "manifest", manifest)
			return manifest, validateRemoteZipManifest(manifest)
		}
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
	return "", fmt.Errorf("Invalid deployment generator %s", target)
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

func GetRemoteManifestToFile(target string, filename string) error {
	resp, err := http.Get(target)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad response from remote manifest %d", resp.StatusCode)
	}
	manifestBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, manifestBytes, 0644)
}
