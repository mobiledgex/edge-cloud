package main

import (
	"context"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

var oldPrometheusControllerApp = edgeproto.App{
	Key:           MEXPrometheusAppKey,
	ImagePath:     "stable/prometheus-operator",
	Deployment:    cloudcommon.AppDeploymentTypeHelm,
	DefaultFlavor: edgeproto.FlavorKey{Name: *appFlavor},
	DelOpt:        edgeproto.DeleteType_AUTO_DELETE,
	InternalPorts: true,
}
var durationShort = "5s"
var durationLong = "45s"

func TestInfra(t *testing.T) {
	// Test duration to string conversion
	oldInterval, _ := time.ParseDuration(durationShort)
	require.Equal(t, durationShort, scrapeIntervalInSeconds(oldInterval), "scrapeIntervalInSeconds test")

	// Test fillConfig scrape interval 5 sec
	err := fillAppConfigs(&oldPrometheusControllerApp, oldInterval)
	require.Nil(t, err, "fillAppConfigs failed")
	// Should be a single Config there
	require.Equal(t, 1, len(oldPrometheusControllerApp.Configs), "Number of configs in app is wrong")
	*scrapeInterval, _ = time.ParseDuration(durationLong)
	newApp, err := getPrometheusAppFromClusterSvc()
	require.Nil(t, err, "getPrometheusAppFromController failed")
	// Check that the fields that are different are correct
	setPrometheusAppDiffFields(&oldPrometheusControllerApp, newApp)
	require.Equal(t, 3, len(newApp.Fields), "Incorrect number of different fields")
	require.Contains(t, newApp.Fields, edgeproto.AppFieldConfigs, "Missing edgeproto.AppFieldConfigs")
	require.Contains(t, newApp.Fields, edgeproto.AppFieldConfigsKind, "Missing edgeproto.AppFieldConfigsKind")
	require.Contains(t, newApp.Fields, edgeproto.AppFieldConfigsConfig, "Missing edgeproto.AppFieldConfigsConfig")

	// Change the image path and check that image path only gets set
	newApp2 := oldPrometheusControllerApp
	newApp2.ImagePath = "newImagePath"
	setPrometheusAppDiffFields(&oldPrometheusControllerApp, &newApp2)
	require.Equal(t, 1, len(newApp2.Fields), "Incorrect number of different fields")
	require.Contains(t, newApp2.Fields, edgeproto.AppFieldImagePath, "Missing edgeproto.AppFieldImagePath")

	// If the apps have both the image and imagePath, make sure everything gets set
	*scrapeInterval, _ = time.ParseDuration(durationLong)
	newApp, err = getPrometheusAppFromClusterSvc()
	require.Nil(t, err, "getPrometheusAppFromController failed")
	newApp.ImagePath = "newImagePath"
	// Check that the fields that are different are correct
	setPrometheusAppDiffFields(&oldPrometheusControllerApp, newApp)
	require.Equal(t, 4, len(newApp.Fields), "Incorrect number of different fields")
	require.Contains(t, newApp.Fields, edgeproto.AppFieldConfigs, "Missing edgeproto.AppFieldConfigs")
	require.Contains(t, newApp.Fields, edgeproto.AppFieldConfigsKind, "Missing edgeproto.AppFieldConfigsKind")
	require.Contains(t, newApp.Fields, edgeproto.AppFieldConfigsConfig, "Missing edgeproto.AppFieldConfigsConfig")
	require.Contains(t, newApp.Fields, edgeproto.AppFieldImagePath, "Missing edgeproto.AppFieldImagePath")
}

// TestAutoScaleT primarily checks that AutoScale template parsing works, because
// otherwise cluster-svc could crash during runtime if template has an issue.
func TestAutoScaleT(t *testing.T) {
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	edgeproto.InitAutoScalePolicyCache(&AutoScalePolicyCache)

	instKey := testutil.ClusterInstData[0].Key

	policy := edgeproto.AutoScalePolicy{}
	policy.Key.Developer = instKey.Developer
	policy.Key.Name = "test-policy"
	policy.MinNodes = 1
	policy.MaxNodes = 5
	policy.ScaleUpCpuThresh = 80
	policy.ScaleDownCpuThresh = 20
	policy.TriggerTimeSec = 60

	AutoScalePolicyCache.Update(ctx, &policy, 0)
	configs, err := getAppInstConfigs(instKey, policy.Key.Name)
	require.Nil(t, err)
	require.Equal(t, 1, len(configs))
}
