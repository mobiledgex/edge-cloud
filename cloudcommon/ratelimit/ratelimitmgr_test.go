package ratelimit

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/stretchr/testify/require"
)

func checkNumSettings(t *testing.T, lenFlowSettings, lenMaxSettings int, settings *edgeproto.RateLimitSettings) {
	if lenFlowSettings == 0 && lenMaxSettings == 0 && settings == nil {
		// this is ok
		return
	}
	require.Equal(t, lenFlowSettings, len(settings.FlowSettings))
	require.Equal(t, lenMaxSettings, len(settings.MaxReqsSettings))
}

func TestUpdates(t *testing.T) {
	mgr := NewRateLimitManager(false, 100, 100)
	// add settings
	for _, fset := range dbFlowSettings {
		mgr.UpdateFlowRateLimitSettings(fset)
	}
	for _, mset := range dbMaxSettings {
		mgr.UpdateMaxReqsRateLimitSettings(mset)
	}
	// check that data was populated
	require.NotNil(t, mgr.limitsPerApi)
	require.Equal(t, 2, len(mgr.limitsPerApi))

	global := mgr.limitsPerApi[edgeproto.GlobalApiName]
	require.NotNil(t, global)
	checkNumSettings(t, 2, 1, global.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
	checkNumSettings(t, 2, 1, global.apiEndpointRateLimitSettings.PerIpRateLimitSettings)
	require.Nil(t, global.apiEndpointRateLimitSettings.PerUserRateLimitSettings)
	require.Equal(t, 3, len(global.limitAllRequests.limiters))
	require.Equal(t, 0, len(global.limitsPerIp))
	require.Equal(t, 0, len(global.limitsPerUser))

	verifyLoc := mgr.limitsPerApi["VerifyLocation"]
	require.NotNil(t, verifyLoc)
	checkNumSettings(t, 1, 0, verifyLoc.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
	checkNumSettings(t, 1, 0, verifyLoc.apiEndpointRateLimitSettings.PerIpRateLimitSettings)
	require.Nil(t, verifyLoc.apiEndpointRateLimitSettings.PerUserRateLimitSettings)
	require.Equal(t, 1, len(verifyLoc.limitAllRequests.limiters))
	require.Equal(t, 0, len(verifyLoc.limitsPerIp))
	require.Equal(t, 0, len(verifyLoc.limitsPerUser))

	// remove flow settings
	for _, fset := range dbFlowSettings {
		mgr.RemoveFlowRateLimitSettings(fset.Key)
	}
	global = mgr.limitsPerApi[edgeproto.GlobalApiName]
	require.NotNil(t, global)
	checkNumSettings(t, 0, 1, global.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
	checkNumSettings(t, 0, 1, global.apiEndpointRateLimitSettings.PerIpRateLimitSettings)
	require.Nil(t, global.apiEndpointRateLimitSettings.PerUserRateLimitSettings)

	require.Equal(t, 1, len(global.limitAllRequests.limiters))
	require.Equal(t, 0, len(global.limitsPerIp))
	require.Equal(t, 0, len(global.limitsPerUser))
	verifyLoc = mgr.limitsPerApi["VerifyLocation"]
	require.Nil(t, verifyLoc)

	// remove max settings
	for _, mset := range dbMaxSettings {
		mgr.RemoveMaxReqsRateLimitSettings(mset.Key)
	}
	require.Equal(t, 0, len(mgr.limitsPerApi))
}
