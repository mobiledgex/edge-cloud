package cloudcommon

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/deploygen"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	yaml "github.com/mobiledgex/yaml/v2"
	v1 "k8s.io/api/core/v1"
)

var DeploymentTypeKubernetes = "kubernetes"
var DeploymentTypeVM = "vm"
var DeploymentTypeHelm = "helm"
var DeploymentTypeDocker = "docker"

var ValidAppDeployments = []string{
	DeploymentTypeKubernetes,
	DeploymentTypeVM,
	DeploymentTypeHelm,
	DeploymentTypeDocker,
}

var ValidCloudletDeployments = []string{
	DeploymentTypeDocker,
	DeploymentTypeKubernetes,
}

type DockerManifest struct {
	DockerComposeFiles []string
}

func IsValidDeploymentType(DeploymentType string, validDeployments []string) bool {
	for _, d := range validDeployments {
		if DeploymentType == d {
			return true
		}
	}
	return false
}

func IsValidDeploymentForImage(imageType edgeproto.ImageType, deployment string) bool {
	switch imageType {
	case edgeproto.ImageType_IMAGE_TYPE_DOCKER:
		if deployment == DeploymentTypeKubernetes || deployment == DeploymentTypeDocker {
			return true
		}
	case edgeproto.ImageType_IMAGE_TYPE_QCOW:
		if deployment == DeploymentTypeVM {
			return true
		}
	case edgeproto.ImageType_IMAGE_TYPE_HELM:
		if deployment == DeploymentTypeHelm {
			return true
		}
	}
	return false
}

func GetDockerDeployType(manifest string) string {
	if manifest == "" {
		return "docker"
	}
	if strings.HasSuffix(manifest, ".zip") {
		return "docker-compose-zip"
	}
	return "docker-compose"
}

// GetMappedAccessType gets the default access type for the deployment if AccessType_ACCESS_TYPE_DEFAULT_FOR_DEPLOYMENT
// is specified.  It returns an error if the access type is not valid for the deployment
func GetMappedAccessType(accessType edgeproto.AccessType, deployment, deploymentManifest string) (edgeproto.AccessType, error) {
	switch accessType {

	case edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER:
		if deployment == DeploymentTypeKubernetes || deployment == DeploymentTypeHelm || deployment == DeploymentTypeDocker || deployment == DeploymentTypeVM {
			return accessType, nil
		}
	case edgeproto.AccessType_ACCESS_TYPE_DIRECT:
		if deployment == DeploymentTypeVM || deployment == DeploymentTypeDocker {
			return accessType, nil
		}
	case edgeproto.AccessType_ACCESS_TYPE_DEFAULT_FOR_DEPLOYMENT:
		if deployment == DeploymentTypeVM {
			return edgeproto.AccessType_ACCESS_TYPE_DIRECT, nil
		}
		// all others default to LB
		return edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER, nil
	}
	return accessType, fmt.Errorf("Invalid access type for deployment")
}

func IsValidDeploymentManifest(DeploymentType, command, manifest string, ports []dme.AppPort) error {
	if DeploymentType == DeploymentTypeVM {
		if command != "" {
			return fmt.Errorf("both deploymentmanifest and command cannot be used together for VM based deployment")
		}
		if strings.HasPrefix(manifest, "#cloud-config") {
			return nil
		}
		return fmt.Errorf("only cloud-init script support, must start with '#cloud-config'")
	} else if DeploymentType == DeploymentTypeKubernetes {
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
				if strings.HasSuffix(kp.Name, "tls") {
					appPort.Tls = true
				}
				objPorts[appPort.String()] = struct{}{}
			}
		}
		missingPorts := []string{}
		for _, appPort := range ports {
			// http is mapped to tcp
			if appPort.EndPort != 0 {
				// We have a range-port notation on the dme.AppPort
				// while our manifest exhaustively enumerates each as a kubePort
				if appPort.Proto == dme.LProto_L_PROTO_HTTP {
					return fmt.Errorf("Port range not allowed for HTTP")
				}
				start := appPort.InternalPort
				end := appPort.EndPort
				for i := start; i <= end; i++ {
					// expand short hand notation to test membership in map
					tp := dme.AppPort{
						Proto:        appPort.Proto,
						InternalPort: int32(i),
						EndPort:      int32(0),
						Tls:          appPort.Tls,
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
		return DeploymentTypeKubernetes, nil
	case edgeproto.ImageType_IMAGE_TYPE_QCOW:
		return DeploymentTypeVM, nil
	case edgeproto.ImageType_IMAGE_TYPE_HELM:
		return DeploymentTypeHelm, nil
	}
	return "", fmt.Errorf("unknown image type %s", imageType)
}

func GetImageTypeForDeployment(deployment string) (edgeproto.ImageType, error) {
	switch deployment {
	case DeploymentTypeDocker:
		fallthrough
	case DeploymentTypeKubernetes:
		return edgeproto.ImageType_IMAGE_TYPE_DOCKER, nil
	case DeploymentTypeHelm:
		return edgeproto.ImageType_IMAGE_TYPE_HELM, nil
	case DeploymentTypeVM:
		// could be different formats
		fallthrough
	default:
		return edgeproto.ImageType_IMAGE_TYPE_UNKNOWN, nil
	}
}

// GetAppDeploymentManifest gets the deployment-specific manifest.
func GetAppDeploymentManifest(ctx context.Context, vaultConfig *vault.Config, app *edgeproto.App) (string, error) {
	if app.DeploymentManifest != "" {
		return GetDeploymentManifest(ctx, vaultConfig, app.DeploymentManifest)
	} else if app.DeploymentGenerator != "" {
		return GenerateManifest(app)
	} else if app.Deployment == DeploymentTypeKubernetes {
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

func validateRemoteZipManifest(ctx context.Context, vaultConfig *vault.Config, manifest string) error {
	zipfile := "/tmp/temp.zip"
	err := GetRemoteManifestToFile(ctx, vaultConfig, manifest, zipfile)
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

func GetDeploymentManifest(ctx context.Context, vaultConfig *vault.Config, manifest string) (string, error) {
	// manifest may be remote target or inline json/yaml
	if strings.HasPrefix(manifest, "http://") || strings.HasPrefix(manifest, "https://") {

		if strings.HasSuffix(manifest, ".zip") {
			log.DebugLog(log.DebugLevelApi, "zipfile manifest found", "manifest", manifest)
			return manifest, validateRemoteZipManifest(ctx, vaultConfig, manifest)
		}
		mf, err := GetRemoteManifest(ctx, vaultConfig, manifest)
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

func GetRemoteManifest(ctx context.Context, vaultConfig *vault.Config, target string) (string, error) {
	var content string
	err := DownloadFile(ctx, vaultConfig, target, "", &content)
	if err != nil {
		return "", err
	}
	return content, nil
}

func GetRemoteManifestToFile(ctx context.Context, vaultConfig *vault.Config, target string, filename string) error {
	return DownloadFile(ctx, vaultConfig, target, filename, nil)
}

// 5GB = 10minutes
func GetTimeout(cLen int) time.Duration {
	fileSizeInGB := float64(cLen) / (1024.0 * 1024.0 * 1024.0)
	timeoutUnit := int(math.Ceil(fileSizeInGB / 5.0))
	if fileSizeInGB > 5 {
		return time.Duration(timeoutUnit) * 10 * time.Minute
	}
	return 15 * time.Minute
}

func DownloadFile(ctx context.Context, vaultConfig *vault.Config, fileUrlPath string, filePath string, content *string) error {
	var reqConfig *RequestConfig

	log.SpanLog(ctx, log.DebugLevelApi, "attempt to download file", "file-url", fileUrlPath)

	// Adjust request timeout based on File Size
	//  - Timeout is increased by 10min for every 5GB
	//  - If less than 5GB, then use default timeout
	resp, err := SendHTTPReq(ctx, "HEAD", fileUrlPath, vaultConfig, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	contentLength := resp.Header.Get("Content-Length")
	cLen, err := strconv.Atoi(contentLength)
	if err == nil && cLen > 0 {
		timeout := GetTimeout(cLen)
		if timeout > 0 {
			reqConfig = &RequestConfig{
				Timeout: timeout,
			}
			log.SpanLog(ctx, log.DebugLevelApi, "increased request timeout", "file-url", fileUrlPath, "timeout", timeout.String())
		}
	}

	resp, err = SendHTTPReq(ctx, "GET", fileUrlPath, vaultConfig, reqConfig)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	if filePath != "" {
		// Create the file
		out, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to download file %v", err)
		}
	}

	if content != nil {
		contentBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		*content = string(contentBytes)
	}

	return nil
}
