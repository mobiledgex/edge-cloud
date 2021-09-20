package main

import (
	"context"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestAlertPolicyApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()
	defer testfinish()

	dummy := dummyEtcd{}
	dummy.Start()
	defer dummy.Stop()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache, &appInstInfoApi, &clusterInstInfoApi)
	reduceInfoTimeouts(t, ctx)
	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, &gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData())
	insertCloudletInfo(ctx, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, &autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, &autoScalePolicyApi, testutil.AutoScalePolicyData)
	testutil.InternalAppCreate(t, &appApi, testutil.AppData)
	testutil.InternalClusterInstCreate(t, &clusterInstApi, testutil.ClusterInstData)
	testutil.InternalCloudletRefsTest(t, "show", &cloudletRefsApi, testutil.CloudletRefsData)
	clusterInstCnt := len(clusterInstApi.cache.Objs)
	require.Equal(t, len(testutil.ClusterInstData), clusterInstCnt)
	testutil.InternalAppInstCreate(t, &appInstApi, testutil.AppInstData)

	// Invalid severity
	userAlert := testutil.AlertPolicyData[0]
	userAlert.Severity = "invalid"
	_, err := userAlertApi.CreateAlertPolicy(ctx, &userAlert)
	require.NotNil(t, err, "Invlid severity passed in")

	// Invalid set of conditions for an alert
	userAlert = testutil.AlertPolicyData[0]
	userAlert.ActiveConnLimit = 10
	_, err = userAlertApi.CreateAlertPolicy(ctx, &userAlert)
	require.NotNil(t, err, "Both active connections and cpu cannot be set for a user alert")

	// Invalid set of conditions for an alert
	userAlert = testutil.AlertPolicyData[0]
	userAlert.ActiveConnLimit = 0
	userAlert.CpuUtilizationLimit = 0
	userAlert.MemUtilizationLimit = 0
	userAlert.DiskUtilizationLimit = 0
	_, err = userAlertApi.CreateAlertPolicy(ctx, &userAlert)
	require.NotNil(t, err, "User Alert should have at least one set value")

	// Invalid set of conditions for an alert
	userAlert = testutil.AlertPolicyData[0]
	userAlert.CpuUtilizationLimit = 200
	_, err = userAlertApi.CreateAlertPolicy(ctx, &userAlert)
	require.NotNil(t, err, "Cpu cannot be >100%")

	// Create user alert with trigger time, that's invalid
	userAlert = testutil.AlertPolicyData[0]
	userAlert.TriggerTime = 0
	_, err = userAlertApi.CreateAlertPolicy(ctx, &userAlert)
	require.NotNil(t, err, "Trigger Time should be at least 30s")

	// Create user alert with trigger time, that's invalid
	userAlert = testutil.AlertPolicyData[0]
	userAlert.TriggerTime = edgeproto.Duration(80 * time.Hour)
	_, err = userAlertApi.CreateAlertPolicy(ctx, &userAlert)
	require.NotNil(t, err, "Trigger Time cannot exceed 3 days")

	// Delete non-existent user alert
	userAlert = testutil.AlertPolicyData[0]
	_, err = userAlertApi.DeleteAlertPolicy(ctx, &userAlert)
	require.NotNil(t, err)
	require.Equal(t, err, userAlert.Key.NotFoundError())

	// Create a user alerts
	testutil.InternalAlertPolicyTest(t, "cud", &userAlertApi, testutil.AlertPolicyData)

	// Add alert to app
	appAlert := edgeproto.AppAlertPolicy{
		AppKey:      testutil.AppData[0].Key,
		AlertPolicy: testutil.AlertPolicyData[0].Key.Name,
	}
	_, err = appApi.AddAppAlertPolicy(ctx, &appAlert)
	require.Nil(t, err)

	// Add non-existent alert to app
	appAlert.AlertPolicy = "nonexistent"
	_, err = appApi.AddAppAlertPolicy(ctx, &appAlert)
	require.NotNil(t, err, "User Alert Should exist before being added to an app")

	// remove non-existent alert from app
	appAlert.AlertPolicy = "nonexistent"
	_, err = appApi.RemoveAppAlertPolicy(ctx, &appAlert)
	require.NotNil(t, err, "User Alert Should exist on the app to be removed")

	// Remove user alert, that is configured on the app - should fail
	userAlert = testutil.AlertPolicyData[0]
	_, err = userAlertApi.DeleteAlertPolicy(ctx, &userAlert)
	require.NotNil(t, err, "Cannot delete alert that's configured on an app")

	// Update user alert - add invalid selection
	userAlert = testutil.AlertPolicyData[1]
	userAlert.CpuUtilizationLimit = 30
	userAlert.Fields = []string{edgeproto.AlertPolicyFieldCpuUtilizationLimit}
	_, err = userAlertApi.UpdateAlertPolicy(ctx, &userAlert)
	require.NotNil(t, err, "Should not be allowed to update alert with invalid set of arguments")
	userAlert = testutil.AlertPolicyData[1]
	userAlert.TriggerTime = 0
	userAlert.Fields = []string{edgeproto.AlertPolicyFieldTriggerTime}
	_, err = userAlertApi.UpdateAlertPolicy(ctx, &userAlert)
	require.NotNil(t, err, "Should not be allowed to update alert with invalid trigger time")

	// Update user alert
	userAlert = testutil.AlertPolicyData[0]
	userAlert.CpuUtilizationLimit = 90
	_, err = userAlertApi.UpdateAlertPolicy(ctx, &userAlert)
	require.Nil(t, err)

	// Remove user alert from the app
	appAlert = edgeproto.AppAlertPolicy{
		AppKey:      testutil.AppData[0].Key,
		AlertPolicy: testutil.AlertPolicyData[0].Key.Name,
	}
	_, err = appApi.RemoveAppAlertPolicy(ctx, &appAlert)
	require.Nil(t, err)

	// Delete all user alert
	userAlert = testutil.AlertPolicyData[0]
	_, err = userAlertApi.DeleteAlertPolicy(ctx, &userAlert)
	require.Nil(t, err)
}
