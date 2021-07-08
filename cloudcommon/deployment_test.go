package cloudcommon

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/deploygen"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

var deploymentMF = "some deployment manifest"

func TestDeployment(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
	log.InitTracer(nil)
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
	authApi := &DummyRegistryAuthApi{}
	fmt.Printf("test deployment %s, manifest %s, generator %s\n",
		app.Deployment, app.DeploymentManifest, app.DeploymentGenerator)
	_, err := GetAppDeploymentManifest(ctx, authApi, app)
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

var gpuBaseDeploymentManifest = `apiVersion: v1
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pillimogo-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pillimogo
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: pillimogo
        mex-app: pillimogo-deployment
        mexAppName: pillimogo
        mexAppVersion: "101"
    spec:
      containers:`

var gpuSubManifest = `
      - image: docker.mobiledgex.net/atlanticinc/images/pillimogo10:1.0.1
        imagePullPolicy: Always
        name: pillimogo
        ports:
        - containerPort: 443
          protocol: TCP
        resources:
          limits:
             nvidia.com/gpu: 1`

func TestDeploymentManifest(t *testing.T) {
	var err error

	// gpu flavor with rescount as 1
	flavor := &testutil.FlavorData[4]

	manifestResCnt1 := gpuBaseDeploymentManifest + gpuSubManifest
	err = IsValidDeploymentManifestForFlavor(DeploymentTypeKubernetes, manifestResCnt1, flavor)
	require.Nil(t, err, "valid gpu deployment manifest")

	manifestResCnt2 := gpuBaseDeploymentManifest + gpuSubManifest + gpuSubManifest
	err = IsValidDeploymentManifestForFlavor(DeploymentTypeKubernetes, manifestResCnt2, flavor)
	require.NotNil(t, err, "invalid gpu deployment manifest")
	require.Contains(t, err.Error(), "GPU resource limit (value:2) exceeds flavor specified count 1")

	flavor.OptResMap["gpu"] = "pci:4"
	err = IsValidDeploymentManifestForFlavor(DeploymentTypeKubernetes, manifestResCnt2, flavor)
	require.Nil(t, err, "valid gpu deployment manifest")
}
