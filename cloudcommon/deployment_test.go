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

	// kvm with no deployment is ok
	app.Deployment = AppDeploymentTypeKVM
	app.DeploymentManifest = ""
	app.DeploymentGenerator = ""
	testAppDeployment(t, app, true)

	// untested - remote generator

	// negative test - invalid generator
	app.Deployment = AppDeploymentTypeKubernetes
	app.DeploymentManifest = ""
	app.DeploymentGenerator = "invalid"
	testAppDeployment(t, app, false)
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
