package cloudcommon

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mobiledgex/edge-cloud/deploygen"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

var deploymentMF = "some deployment manifest"

func TestDeployment(t *testing.T) {
	app := &testutil.AppData[0]

	// start up http server to serve deployment manifest
	tsManifest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, deploymentMF)
	}))
	defer tsManifest.Close()

	// base case - unspecified - default to kubernetes
	app.Deployment = AppDeploymentTypeKubernetes
	app.DeploymentManifest = ""
	app.DeploymentGenerator = ""
	testAppDeployment(t, app, true)

	// specify manifest inline
	app.DeploymentManifest = deploymentMF
	testAppDeployment(t, app, true)

	// specific remote manifest
	app.DeploymentManifest = tsManifest.URL
	testAppDeployment(t, app, true)

	// locally specified generator
	app.DeploymentManifest = ""
	app.DeploymentGenerator = deploygen.KubernetesBasic
	testAppDeployment(t, app, true)

	// Docker image type for Kubernetes deployment
	app.ImageType = edgeproto.ImageType_ImageTypeDocker
	testValidImageDeployment(t, app, true)

	// kvm with no deployment is ok
	app.Deployment = AppDeploymentTypeKVM
	app.DeploymentManifest = ""
	app.DeploymentGenerator = ""
	testAppDeployment(t, app, true)

	// untested - remote generator

	// QCOW image type for KVM deployment
	app.ImageType = edgeproto.ImageType_ImageTypeQCOW
	testValidImageDeployment(t, app, true)

	// helm with no manifest
	app.Deployment = AppDeploymentTypeHelm
	app.DeploymentManifest = ""
	app.DeploymentGenerator = ""
	testAppDeployment(t, app, true)

	// No image type for Helm deployment
	app.ImageType = edgeproto.ImageType_ImageTypeUnknown
	testValidImageDeployment(t, app, true)

	// negative test - invalid image type for helm deployment
	app.ImageType = edgeproto.ImageType_ImageTypeDocker
	testValidImageDeployment(t, app, false)

	// negative test - invalid generator
	app.Deployment = AppDeploymentTypeKubernetes
	app.DeploymentManifest = ""
	app.DeploymentGenerator = "invalid"
	testAppDeployment(t, app, false)

	// negative test - invalid image type
	app.ImageType = edgeproto.ImageType_ImageTypeQCOW
	testValidImageDeployment(t, app, false)
}

func testAppDeployment(t *testing.T, app *edgeproto.App, valid bool) {
	fmt.Printf("test deployment %s, manifest %s, generator %s\n",
		app.Deployment, app.DeploymentManifest, app.DeploymentGenerator)
	_, err := GetAppDeploymentManifest(app)
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
