package cloudcommon

import (
	"archive/zip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/deploygen"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	yaml "github.com/mobiledgex/yaml/v2"
	"github.com/stretchr/testify/require"
)

var deploymentMF = "some deployment manifest"

func TestDeployment(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	app := &testutil.AppData[0]

	// start up http server to serve deployment manifest
	tsManifest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, deploymentMF)
	}))
	defer tsManifest.Close()

	// base case - unspecified - default to kubernetes
	app.Deployment = DeploymentTypeKubernetes
	app.DeploymentManifest = ""
	app.DeploymentGenerator = ""
	testAppDeployment(ctx, t, app, true)

	// specify manifest inline
	app.DeploymentManifest = deploymentMF
	testAppDeployment(ctx, t, app, true)

	// specific remote manifest
	app.DeploymentManifest = tsManifest.URL
	testAppDeployment(ctx, t, app, true)

	// locally specified generator
	app.DeploymentManifest = ""
	app.DeploymentGenerator = deploygen.KubernetesBasic
	testAppDeployment(ctx, t, app, true)

	// Docker image type for Kubernetes deployment
	app.ImageType = edgeproto.ImageType_IMAGE_TYPE_DOCKER
	testValidImageDeployment(t, app, true)

	// vm with no deployment is ok
	app.Deployment = DeploymentTypeVM
	app.DeploymentManifest = ""
	app.DeploymentGenerator = ""
	testAppDeployment(ctx, t, app, true)

	// untested - remote generator

	// QCOW image type for VM deployment
	app.ImageType = edgeproto.ImageType_IMAGE_TYPE_QCOW
	testValidImageDeployment(t, app, true)

	// helm with no manifest
	app.Deployment = DeploymentTypeHelm
	app.DeploymentManifest = ""
	app.DeploymentGenerator = ""
	testAppDeployment(ctx, t, app, true)

	// No image type for Helm deployment
	app.ImageType = edgeproto.ImageType_IMAGE_TYPE_HELM
	testValidImageDeployment(t, app, true)

	// negative test - invalid image type for helm deployment
	app.ImageType = edgeproto.ImageType_IMAGE_TYPE_UNKNOWN
	testValidImageDeployment(t, app, false)

	// negative test - invalid generator
	app.Deployment = DeploymentTypeKubernetes
	app.DeploymentManifest = ""
	app.DeploymentGenerator = "invalid"
	testAppDeployment(ctx, t, app, false)

	// negative test - invalid image type
	app.ImageType = edgeproto.ImageType_IMAGE_TYPE_QCOW
	testValidImageDeployment(t, app, false)
}

func testAppDeployment(ctx context.Context, t *testing.T, app *edgeproto.App, valid bool) {
	fmt.Printf("test deployment %s, manifest %s, generator %s\n",
		app.Deployment, app.DeploymentManifest, app.DeploymentGenerator)
	_, err := GetAppDeploymentManifest(ctx, nil, app)
	if valid {
		require.Nil(t, err)
	} else {
		require.NotNil(t, err)
	}
}

func testValidImageDeployment(t *testing.T, app *edgeproto.App, valid bool) {
	fmt.Printf("test deployment %s, image %s\n", app.Deployment, app.ImageType)
	v := IsValidDeploymentForImage(app.ImageType, app.Deployment)
	if valid {
		require.True(t, v)
	} else {
		require.False(t, v)
	}

}

func TestTimeout(t *testing.T) {
	var val time.Duration

	oneG := 1073741824
	val = GetTimeout(0)
	require.Equal(t, val, 15*time.Minute)
	val = GetTimeout(5 * oneG)
	require.Equal(t, val, 15*time.Minute)
	val = GetTimeout(10 * oneG)
	require.Equal(t, val, 20*time.Minute)
}

func zipManifests() error {
	// Get a Buffer to Write To
	zipfile := "/tmp/test.zip"
	outFile, err := os.Create(zipfile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(outFile)

	content, err := ioutil.ReadFile("manifest.yml")
	if err != nil {
		return fmt.Errorf("unable to open %s manifest file: %v", "manifest.yml", err)
	}
	var dm DockerManifest
	err = yaml.Unmarshal([]byte(content), &dm)
	if err != nil {
		return fmt.Errorf("unmarshalling manifest.yml: %v", err)
	}
	zipFiles := []string{"manifest.yml"}
	zipFiles = append(zipFiles, dm.DockerComposeFiles...)
	for _, fileName := range zipFiles {
		f, err := w.Create(fileName)
		if err != nil {
			return err
		}
		content, err = ioutil.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("unable to open %s file: %v", fileName, err)
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			return err
		}
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

func TestRemoteZipManifests(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	err := zipManifests()
	require.Nil(t, err)

	zipContainers, err := GetRemoteZipDockerManifests(ctx, nil, "", "", NoDownload)
	fmt.Println(zipContainers)
	require.Nil(t, err)
}
