package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
)

func TestAppApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// cannot create apps without developer
	ctx := context.TODO()
	for _, obj := range testutil.AppData {
		_, err := appApi.CreateApp(ctx, &obj)
		assert.NotNil(t, err, "Create app without developer")
	}

	// create support data
	testutil.InternalDeveloperCreate(t, &developerApi, testutil.DevData)
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalClusterFlavorCreate(t, &clusterFlavorApi, testutil.ClusterFlavorData)
	testutil.InternalClusterCreate(t, &clusterApi, testutil.ClusterData)

	testutil.InternalAppTest(t, "cud", &appApi, testutil.AppData)

	// Test replicas being set via deploygenConfig (replicas: 5)
	app_mf := appApi.cache.Objs[testutil.AppData[0].Key].DeploymentManifest
	objs, _, err := cloudcommon.DecodeK8SYaml(app_mf)
	assert.Nil(t, err, "Decode K8s deployment manifest file")

	for i, _ := range objs {
		deployment, ok := objs[i].(*appsv1.Deployment)
		if !ok {
			continue
		}
		assert.Equal(t, int32(5), *deployment.Spec.Replicas, "Replicas count is set correctly (5)")
		break
	}

	// Test replicas being set via deploygenConfig (replicas: )
	app_mf = appApi.cache.Objs[testutil.AppData[2].Key].DeploymentManifest
	objs, _, err = cloudcommon.DecodeK8SYaml(app_mf)
	assert.Nil(t, err, "Decode K8s deployment manifest file")

	for i, _ := range objs {
		deployment, ok := objs[i].(*appsv1.Deployment)
		if !ok {
			continue
		}
		assert.Equal(t, int32(1), *deployment.Spec.Replicas, "Replicas count is set correctly (1)")
		break
	}

	dummy.Stop()
}
