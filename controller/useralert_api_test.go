package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestUserAlertApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache, &appInstInfoApi, &clusterInstInfoApi)
	reduceInfoTimeouts(t, ctx)
	InfluxUsageUnitTestSetup(t)
	defer InfluxUsageUnitTestStop()
	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, &gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
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
	userAlert := testutil.UserAlertData[0]
	userAlert.Severity = "invalid"
	_, err := userAlertApi.CreateUserAlert(ctx, &userAlert)
	require.NotNil(t, err, "Invlid severity passed in")

	// Invalid set of conditions for an alert
	userAlert = testutil.UserAlertData[0]
	userAlert.ActiveConnLimit = 10
	_, err = userAlertApi.CreateUserAlert(ctx, &userAlert)
	require.NotNil(t, err, "Both active connections and cpu cannot be set for a user alert")

	// Invalid set of conditions for an alert
	userAlert = testutil.UserAlertData[0]
	userAlert.ActiveConnLimit = 0
	userAlert.CpuLimit = 0
	userAlert.MemLimit = 0
	userAlert.DiskLimit = 0
	_, err = userAlertApi.CreateUserAlert(ctx, &userAlert)
	require.NotNil(t, err, "User Alert should have at least one set value")

	// Invalid set of conditions for an alert
	userAlert = testutil.UserAlertData[0]
	userAlert.CpuLimit = 200
	_, err = userAlertApi.CreateUserAlert(ctx, &userAlert)
	require.NotNil(t, err, "Cpu cannot be >100%")

	// Create user alert with trigger time, that's invalid
	userAlert = testutil.UserAlertData[0]
	userAlert.TriggerTime = 0
	_, err = userAlertApi.CreateUserAlert(ctx, &userAlert)
	require.NotNil(t, err, "Trigger Time should be at least 30s")

	// Delete non-existent user alert
	userAlert = testutil.UserAlertData[0]
	_, err = userAlertApi.DeleteUserAlert(ctx, &userAlert)
	require.NotNil(t, err)
	require.Equal(t, err, userAlert.Key.NotFoundError())

	// Create a user alerts
	testutil.InternalUserAlertTest(t, "cud", &userAlertApi, testutil.UserAlertData)

	// Add alert to app
	appAlert := edgeproto.AppUserDefinedAlert{
		AppKey:           testutil.AppData[0].Key,
		UserDefinedAlert: testutil.UserAlertData[0].Key.Name,
	}
	_, err = appApi.AddAppUserDefinedAlert(ctx, &appAlert)
	require.Nil(t, err)

	// Add non-existent alert to app
	appAlert.UserDefinedAlert = "nonexistent"
	_, err = appApi.AddAppUserDefinedAlert(ctx, &appAlert)
	require.NotNil(t, err, "User Alert Should exist before being added to an app")

	// Remove user alert, that is configured on the app - should fail
	userAlert = testutil.UserAlertData[0]
	_, err = userAlertApi.DeleteUserAlert(ctx, &userAlert)
	require.NotNil(t, err, "Cannot delete alert that's configured on an app")

	// Update user alert - add invalid selection
	userAlert = testutil.UserAlertData[1]
	userAlert.CpuLimit = 30
	userAlert.Fields = []string{edgeproto.UserAlertFieldCpuLimit}
	_, err = userAlertApi.UpdateUserAlert(ctx, &userAlert)
	require.NotNil(t, err, "Should not be allowed to update alert with invalid set of arguments")
	userAlert = testutil.UserAlertData[1]
	userAlert.TriggerTime = 0
	userAlert.Fields = []string{edgeproto.UserAlertFieldTriggerTime}
	_, err = userAlertApi.UpdateUserAlert(ctx, &userAlert)
	require.NotNil(t, err, "Should not be allowed to update alert with invalid trigger time")

	// Update user alert
	userAlert = testutil.UserAlertData[0]
	userAlert.CpuLimit = 90
	_, err = userAlertApi.UpdateUserAlert(ctx, &userAlert)
	require.Nil(t, err)

	// Remove user alert from the app
	appAlert = edgeproto.AppUserDefinedAlert{
		AppKey:           testutil.AppData[0].Key,
		UserDefinedAlert: testutil.UserAlertData[0].Key.Name,
	}
	_, err = appApi.RemoveAppUserDefinedAlert(ctx, &appAlert)
	require.Nil(t, err)

	// Delete all user alert
	userAlert = testutil.UserAlertData[0]
	_, err = userAlertApi.DeleteUserAlert(ctx, &userAlert)
	require.Nil(t, err)
}
