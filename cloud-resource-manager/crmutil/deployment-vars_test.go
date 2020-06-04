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
        target: [[ .Deployment.ClusterIp ]]:443
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
  server: [[ .Deployment.ClusterIp ]]
  path: /ifs/kubernetes
`

var testClusterIp = "10.1.1.1"
var testCloudletName = "TestCloudlet"
var testClusterName = "TestCluster"
var testDeveloperOrg = "AcmeAppCo"
var testDnsZone = "mobiledgex-test.net"

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
  server: [[ .Deployment.OperatorName ]]
  path: /ifs/kubernetes
`

var testAppAccessConfig = `
dnsOverride: "*.[[.Deployment.AppOrg]]-[[.Deployment.ClusterName]]-[[.Deployment.CloudletName]].[[.Deployment.DnsZone]]"
lbTlsCertCommonName: ""*.[[.Deployment.AppOrg]]-[[.Deployment.ClusterName]]-[[.Deployment.CloudletName]].[[.Deployment.DnsZone]]"`

var testAppAccessConfigResult = `
dnsOverride: "*.AcmeAppCo-TestCluster-TestCloudlet.mobiledgex-test.net"
lbTlsCertCommonName: ""*.AcmeAppCo-TestCluster-TestCloudlet.mobiledgex-test.net"`

var testInvalidTemplate = `if [[ ! -s \"$OSM_FILE\" ]];`

var testAppAccessConfigDelims = `
dnsOverride: "*.{{.Deployment.AppOrg}}-{{.Deployment.ClusterName}}-{{.Deployment.CloudletName}}.{{.Deployment.DnsZone}}"
lbTlsCertCommonName: ""*.{{.Deployment.AppOrg}}-{{.Deployment.ClusterName}}-{{.Deployment.CloudletName}}.{{.Deployment.DnsZone}}"`

func TestCrmDeploymentVars(t *testing.T) {
	deploymentVars := DeploymentReplaceVars{
		Deployment: CrmReplaceVars{
			ClusterIp:    testClusterIp,
			CloudletName: testCloudletName,
			ClusterName:  testClusterName,
			AppOrg:       testDeveloperOrg,
			DnsZone:      testDnsZone,
		},
	}
	delims := ""
	// positive tests
	val, err := ReplaceDeploymentVars(testConfigFile, delims, &deploymentVars)
	assert.Nil(t, err)
	assert.Equal(t, testConfigFileResult, val)
	val, err = ReplaceDeploymentVars(testManifest, delims, &deploymentVars)
	assert.Nil(t, err)
	assert.Equal(t, testManifestResult, val)
	// Test manifest decode
	_, _, err = cloudcommon.DecodeK8SYaml(val)
	assert.Nil(t, err)
	// App Access Config test
	val, err = ReplaceDeploymentVars(testAppAccessConfig, delims, &deploymentVars)
	assert.Nil(t, err)
	assert.Equal(t, testAppAccessConfigResult, val)

	// configFile with no vars
	val, err = ReplaceDeploymentVars(testConfigFileResult, delims, &deploymentVars)
	assert.Nil(t, err)
	assert.Equal(t, testConfigFileResult, val)

	// error cases
	val, err = ReplaceDeploymentVars(testConfigFileWrongVar, delims, &deploymentVars)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Deployment.OperatorName")

	// Invalid template
	_, err = ReplaceDeploymentVars(testInvalidTemplate, delims, &deploymentVars)
	assert.NotNil(t, err)

	// App Access Config test with different delimiter
	delims = "{{ }}"
	val, err = ReplaceDeploymentVars(testAppAccessConfigDelims, delims, &deploymentVars)
	assert.Nil(t, err)
	assert.Equal(t, testAppAccessConfigResult, val)

	// Invalid delimiter
	delims = "{{ }} {{"
	_, err = ReplaceDeploymentVars(testAppAccessConfigDelims, delims, &deploymentVars)
	assert.NotNil(t, err)
}
