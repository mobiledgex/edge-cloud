package k8smgmt

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

var envVars = `- name: SOME_ENV1
  value: value1
- name: SOME_ENV2
  valueFrom:
    configMapKeyRef:
      key: CloudletName
      name: mexcluster-info
      optional: true
`

func TestEnvVars(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	app := &testutil.AppData[0]
	app.Deployment = cloudcommon.DeploymentTypeKubernetes
	app.DeploymentGenerator = ""
	config := &edgeproto.ConfigFile{
		Kind:   edgeproto.AppConfigEnvYaml,
		Config: envVars,
	}
	app.Configs = append(app.Configs, config)

	// start up http server to serve envVars
	tsEnvVars := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, envVars)
	}))
	defer tsEnvVars.Close()

	// Test Deploymeent manifest with inline EnvVars
	baseMf, err := cloudcommon.GetAppDeploymentManifest(ctx, nil, app)
	require.Nil(t, err)
	envVarsMf, err := MergeEnvVars(ctx, nil, app, baseMf, "")
	require.Nil(t, err)
	// make envVars remote
	app.Configs[0].Config = tsEnvVars.URL
	remoteEnvVars, err := MergeEnvVars(ctx, nil, app, baseMf, "")
	require.Nil(t, err)
	require.Equal(t, envVarsMf, remoteEnvVars)
}
