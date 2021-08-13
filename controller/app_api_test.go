package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestAppApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	testinit()
	cplookup := &node.CloudletPoolCache{}
	cplookup.Init()
	nodeMgr.CloudletPoolLookup = cplookup

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// cannot create apps without developer
	ctx := log.StartTestSpan(context.Background())
	for _, obj := range testutil.AppData {
		_, err := appApi.CreateApp(ctx, &obj)
		require.NotNil(t, err, "Create app without developer")
	}

	// create support data
	testutil.InternalAutoProvPolicyCreate(t, &autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)

	testutil.InternalAppTest(t, "cud", &appApi, testutil.AppData)

	// update should validate ports
	upapp := testutil.AppData[3]
	upapp.AccessPorts = "tcp:0"
	upapp.Fields = []string{edgeproto.AppFieldAccessPorts}
	_, err := appApi.UpdateApp(ctx, &upapp)
	require.NotNil(t, err, "Update app with port 0")
	require.Contains(t, err.Error(), "App ports out of range")

	// update should also validate skipHcPorts
	upapp = testutil.AppData[3]
	upapp.SkipHcPorts = "tcp:8080"
	upapp.Fields = []string{edgeproto.AppFieldSkipHcPorts}
	_, err = appApi.UpdateApp(ctx, &upapp)
	require.Nil(t, err, "Update app with SkipHcPort 8080")
	obj := testutil.AppData[3]
	_, err = appApi.DeleteApp(ctx, &obj)
	require.Nil(t, err)

	// validateSkipHcPorts
	obj = testutil.AppData[2]
	obj.SkipHcPorts = "udp:11111"
	obj.Fields = []string{edgeproto.AppFieldSkipHcPorts}
	_, err = appApi.UpdateApp(ctx, &obj)
	require.NotNil(t, err, "update App with udp skipHcPort")
	require.Contains(t, err.Error(), "Protocol L_PROTO_UDP unsupported for healthchecks")

	obj = testutil.AppData[2]
	obj.SkipHcPorts = "tcp:444"
	obj.Fields = []string{edgeproto.AppFieldSkipHcPorts}
	_, err = appApi.UpdateApp(ctx, &obj)
	require.NotNil(t, err, "Update App with skipHcPort not in AccessPorts")
	require.Contains(t, err.Error(), "skipHcPort 444 not found in accessPorts")

	obj = testutil.AppData[8]
	obj.SkipHcPorts = "tcp:5000-5004"
	obj.Fields = []string{edgeproto.AppFieldSkipHcPorts}
	_, err = appApi.UpdateApp(ctx, &obj)
	require.NotNil(t, err, "Update App with skipHcPort range not in AccessPorts")
	require.Contains(t, err.Error(), "skipHcPort 5003 not found in accessPorts")

	obj = testutil.AppData[8]
	obj.SkipHcPorts = "tcp:5000-5002"
	obj.Fields = []string{edgeproto.AppFieldSkipHcPorts}
	_, err = appApi.UpdateApp(ctx, &obj)
	require.Nil(t, err, "Update App with skipHcPort range")

	// image path is optional for docker deployments if
	// deployment manifest is specified.
	app := edgeproto.App{
		Key: edgeproto.AppKey{
			Organization: "org",
			Name:         "someapp",
			Version:      "1.0.1",
		},
		ImageType:          edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:        "tcp:445,udp:1212",
		Deployment:         "docker", // avoid trying to parse k8s manifest
		DeploymentManifest: "some manifest",
		DefaultFlavor:      testutil.FlavorData[2].Key,
	}
	_, err = appApi.CreateApp(ctx, &app)
	require.Nil(t, err, "Create app with deployment manifest")
	checkApp := edgeproto.App{}
	found := appApi.Get(&app.Key, &checkApp)
	require.True(t, found, "found app")
	require.Equal(t, "", checkApp.ImagePath, "image path empty")
	_, err = appApi.DeleteApp(ctx, &app)
	require.Nil(t, err)

	// manifest must be empty if deployment is helm
	app.Deployment = cloudcommon.DeploymentTypeHelm
	app.DeploymentManifest = testK8SManifest1
	app.ImageType = edgeproto.ImageType_IMAGE_TYPE_HELM
	_, err = appApi.CreateApp(ctx, &app)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Manifest is not used for Helm deployments")
	// check that creation passes with empty manifest
	app.DeploymentManifest = ""
	_, err = appApi.CreateApp(ctx, &app)
	require.Nil(t, err)
	_, err = appApi.DeleteApp(ctx, &app)
	require.Nil(t, err)

	// user-specified manifest parsing/consistency/checking
	app.Deployment = "kubernetes"
	app.DeploymentManifest = testK8SManifest1
	app.ImageType = edgeproto.ImageType_IMAGE_TYPE_DOCKER
	app.AccessPorts = "tcp:80"
	_, err = appApi.CreateApp(ctx, &app)
	require.Nil(t, err)
	_, err = appApi.DeleteApp(ctx, &app)
	require.Nil(t, err)

	// empty config check (edgecloud-3993)
	app.Configs = []*edgeproto.ConfigFile{
		&edgeproto.ConfigFile{
			Kind: edgeproto.AppConfigEnvYaml,
		},
	}
	_, err = appApi.CreateApp(ctx, &app)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Empty config for config kind")

	// test Updating of the ports with a manifest k8s. Manifest should be cleared
	k8sApp := testutil.AppData[2]
	// clean up previous instance first
	_, err = appApi.DeleteApp(ctx, &k8sApp)
	require.Nil(t, err)

	k8sApp = testutil.AppData[2]
	k8sApp.Deployment = cloudcommon.DeploymentTypeKubernetes
	k8sApp.DeploymentManifest = testK8SManifest1
	k8sApp.AccessPorts = "tcp:80"
	_, err = appApi.CreateApp(ctx, &k8sApp)
	require.Nil(t, err)
	// Update ports with a manifest and verify it requires an update to the manifest
	k8sApp.AccessPorts = "tcp:80,tcp:81"
	k8sApp.Fields = []string{edgeproto.AppFieldAccessPorts}
	_, err = appApi.UpdateApp(ctx, &k8sApp)
	require.NotNil(t, err, "k8s app with manifest should complain about the manifest")
	require.Contains(t, "kubernetes manifest which was previously specified must be provided again when changing access ports",
		err.Error())

	vmApp := testutil.AppData[3]
	vmApp.Deployment = cloudcommon.DeploymentTypeVM
	vmApp.DeploymentManifest = testVmManifest
	vmApp.AccessPorts = "tcp:80"
	_, err = appApi.CreateApp(ctx, &vmApp)
	require.Nil(t, err)
	vmApp.AccessPorts = "tcp:80,tcp:81"
	vmApp.Fields = []string{edgeproto.AppFieldAccessPorts}
	// Update of the VM app with a manifest and make sure that manifest is retained
	_, err = appApi.UpdateApp(ctx, &vmApp)
	require.Nil(t, err, "Vm app should be updated with no error")
	storedApp := edgeproto.App{}
	found = appApi.Get(vmApp.GetKey(), &storedApp)
	require.True(t, found, "VM app should still be in etcd after update")
	require.Equal(t, testVmManifest, storedApp.DeploymentManifest, "Deployment manifest should not be affected by access port update")

	// accessports with `maxpktsize`
	app.Key.Name = "k8sapp"
	app.Deployment = "kubernetes"
	app.AccessPorts = "tcp:888,udp:1999:maxpktsize=1500"
	app.DeploymentManifest = ""
	app.Configs = nil
	app.ImageType = edgeproto.ImageType_IMAGE_TYPE_DOCKER
	_, err = appApi.CreateApp(ctx, &app)
	require.Nil(t, err, "Create app with maxpktsize")
	_, err = appApi.DeleteApp(ctx, &app)
	require.Nil(t, err)
	app.AccessPorts = "tcp:888,tcp:1999:maxpktsize=1500"
	// maxpktsize is not valid config for TCP port
	_, err = appApi.CreateApp(ctx, &app)
	require.NotNil(t, err, "Create app with maxpktsize fails")

	app.Key.Name = "dockapp"
	app.Deployment = "docker"
	app.AccessPorts = "tcp:888,udp:1999:maxpktsize=1500"
	app.ImageType = edgeproto.ImageType_IMAGE_TYPE_DOCKER
	_, err = appApi.CreateApp(ctx, &app)
	require.Nil(t, err, "Create app with maxpktsize")
	_, err = appApi.DeleteApp(ctx, &app)
	require.Nil(t, err)
	app.AccessPorts = "tcp:888,udp:1999:maxpktsize=1500000"
	// maxpktsize should be less than equal 50000
	_, err = appApi.CreateApp(ctx, &app)
	require.NotNil(t, err, "Create app with maxpktsize fails")

	dummy.Stop()
}

var testInvalidUrlHelmCfg = "http://invalidUrl"
var testValidYmlHelmCfg = `nfs:
  path: /share
  server: [[ .Deployment.ClusterIp ]]
storageClass:
  name: standard
  defaultClass: true
`

func TestValidateAppConfigs(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// valid config
	configs := []*edgeproto.ConfigFile{
		&edgeproto.ConfigFile{
			Kind:   edgeproto.AppConfigHelmYaml,
			Config: testValidYmlHelmCfg,
		},
	}
	err := validateAppConfigsForDeployment(ctx, configs, cloudcommon.DeploymentTypeHelm)
	require.Nil(t, err)

	// invalid url
	configs = []*edgeproto.ConfigFile{
		&edgeproto.ConfigFile{
			Kind:   edgeproto.AppConfigHelmYaml,
			Config: testInvalidUrlHelmCfg,
		},
	}
	err = validateAppConfigsForDeployment(ctx, configs, cloudcommon.DeploymentTypeHelm)
	require.NotNil(t, err)
}

var testVmManifest = `#cloud-config vmManifest`

var testK8SManifest1 = `---
# Source: cornav/templates/gh-configmap.yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cornav-graphhopper-cm
data:
  config.yml: "..."
---
# Source: cornav/templates/gh-init-configmap.yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cornav-graphhopper-init-cm
data:
  osm.sh: "..."
---
# Source: cornav/templates/gh-pvc.yml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: gh-data-pvc
spec:
  accessModes:
  - ReadWriteMany
  resources:
    requests:
      storage: 500Mi
  storageClassName: nfs-client
  volumeMode: Filesystem
---
# Source: cornav/templates/gh-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: cornav-graphhopper
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8989
    protocol: TCP
    name: http
  selector:
    app: cornav-graphhopper
---
# Source: cornav/templates/gh-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cornav-graphhopper
  labels:
    app: cornav-graphhopper
spec:
  selector:
    matchLabels:
      app: cornav-graphhopper
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: cornav-graphhopper
    spec:
      imagePullSecrets:
        - name: regcred
      securityContext:
        runAsUser: 1000
        runAsGroup: 2000
        fsGroup: 2000
      containers:
      - name: cornav-graphhopper
        image: "graphhopper/graphhopper:latest"
        ports:
        - name: http
          containerPort: 8989
          protocol: TCP
        volumeMounts:
        - name: gh-data
          mountPath: /data
        - name: config
          mountPath: /config
        resources:
          limits:
            cpu: 2000m
            memory: 2048Mi
          requests:
            cpu: 1000m
            memory: 1024Mi
      initContainers:
      - name: cornav-init-graphhopper
        image: thomseddon/utils
        env:
        - name: HTTP_PROXY
          value: http://gif-ccs-001.iavgroup.local:3128
        - name: HTTPS_PROXY
          value: http://gif-ccs-001.iavgroup.local:3128
        volumeMounts:
        - mountPath: /data
          name: gh-data
        - mountPath: /init
          name: init-script
        command: ["/init/osm.sh", "-i", "/data/europe_germany_brandenburg.pbf"]
        resources:
          limits:
            cpu: 100m
            memory: 128Mi
          requests:
            cpu: 100m
            memory: 128Mi
      volumes:
        - name: gh-data
          persistentVolumeClaim:
            claimName: gh-data-pvc
        - name: config
          configMap:
            name: cornav-graphhopper-cm
        - name: init-script
          configMap:
            name: cornav-graphhopper-init-cm
            defaultMode: 0777
`
