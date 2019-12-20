package crmutil

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/stretchr/testify/assert"
)

// NOTE - manifests generally are not supported yet!!!!
var testManifest = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: someapplication1-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      run: someapplication1
  template:
    metadata:
      labels:
        run: someapplication1
        target: {{ .CRM.ClusterIp }}:443
    spec:
      volumes:
      imagePullSecrets:
      - name: registry.mobiledgex.net
      containers:
      - name: someapplication1
        image: registry.mobiledgex.net/mobiledgex_AcmeAppCo/someapplication1:1.0
        imagePullPolicy: Always
        ports:
        - containerPort: 80
          protocol: TCP
        - containerPort: 443
          protocol: TCP
`

var testConfigFile = `nfs:
  server: {{ .CRM.ClusterIp }}
  path: /ifs/kubernetes
`

var testClusterIp = "10.1.1.1"

var testManifestResult = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: someapplication1-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      run: someapplication1
  template:
    metadata:
      labels:
        run: someapplication1
        target: 10.1.1.1:443
    spec:
      volumes:
      imagePullSecrets:
      - name: registry.mobiledgex.net
      containers:
      - name: someapplication1
        image: registry.mobiledgex.net/mobiledgex_AcmeAppCo/someapplication1:1.0
        imagePullPolicy: Always
        ports:
        - containerPort: 80
          protocol: TCP
        - containerPort: 443
          protocol: TCP
`

var testConfigFileResult = `nfs:
  server: 10.1.1.1
  path: /ifs/kubernetes
`
var testConfigFileWrongVar = `nfs:
  server: {{ .CRM.ClusterName }}
  path: /ifs/kubernetes
`

func TestCrmDeploymentVars(t *testing.T) {
	deploymentVars := DeploymentReplaceVars{
		CRM: CrmReplaceVars{
			ClusterIp: testClusterIp,
		},
	}
	// positive tests
	val, err := ReplaceDeploymentVars(testConfigFile, &deploymentVars)
	assert.Nil(t, err)
	assert.Equal(t, testConfigFileResult, val)
	val, err = ReplaceDeploymentVars(testManifest, &deploymentVars)
	assert.Nil(t, err)
	assert.Equal(t, testManifestResult, val)
	// Test manifest decode
	_, _, err = cloudcommon.DecodeK8SYaml(val)
	assert.Nil(t, err)

	// configFile with no vars
	val, err = ReplaceDeploymentVars(testConfigFileResult, &deploymentVars)
	assert.Nil(t, err)
	assert.Equal(t, testConfigFileResult, val)

	// error cases
	val, err = ReplaceDeploymentVars(testConfigFileWrongVar, &deploymentVars)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "CRM.ClusterName")
}
