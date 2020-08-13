package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestAppApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	testinit()
	log.InitTracer("")
	defer log.FinishTracer()

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

	obj := testutil.AppData[3]
	_, err = appApi.DeleteApp(ctx, &obj)
	require.Nil(t, err)

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

	// user-specified manifest parsing/consistency/checking
	app.Deployment = "kubernetes"
	app.DeploymentManifest = testK8SManifest1
	app.AccessPorts = "tcp:80"
	_, err = appApi.CreateApp(ctx, &app)
	require.Nil(t, err)
	_, err = appApi.DeleteApp(ctx, &app)
	require.Nil(t, err)

	dummy.Stop()
}

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
