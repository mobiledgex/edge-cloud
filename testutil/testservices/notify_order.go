package testservices

import (
	"reflect"
	"testing"

	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/stretchr/testify/require"
)

// Check order dependencies for notify send.
// This encompasses both object dependencies (objects depend on other objects)
// and service-specific dependencies.
func CheckNotifySendOrder(t *testing.T, order map[reflect.Type]int) {
	flavor := reflect.TypeOf((*notify.FlavorSendMany)(nil))
	cloudlet := reflect.TypeOf((*notify.CloudletSendMany)(nil))
	clusterInst := reflect.TypeOf((*notify.ClusterInstSendMany)(nil))
	app := reflect.TypeOf((*notify.AppSendMany)(nil))
	privacyPolicy := reflect.TypeOf((*notify.PrivacyPolicySendMany)(nil))
	autoScalePolicy := reflect.TypeOf((*notify.AutoScalePolicySendMany)(nil))
	autoProvPolicy := reflect.TypeOf((*notify.AutoProvPolicySendMany)(nil))
	appInst := reflect.TypeOf((*notify.AppInstSendMany)(nil))
	appInstRefs := reflect.TypeOf((*notify.AppInstRefsSendMany)(nil))

	// Cloudlet dependencies
	if o, found := order[cloudlet]; found {
		CheckDep(t, order, o, flavor)
	}
	// ClusterInst dependencies
	if o, found := order[clusterInst]; found {
		CheckDep(t, order, o, flavor)
		CheckDep(t, order, o, cloudlet)
		CheckDep(t, order, o, autoScalePolicy)
		CheckDep(t, order, o, privacyPolicy)
	}
	// App dependecies
	if o, found := order[app]; found {
		CheckDep(t, order, o, flavor)
		CheckDep(t, order, o, autoProvPolicy)
	}
	// AppInst dependencies
	if o, found := order[appInst]; found {
		CheckDep(t, order, o, flavor)
		CheckDep(t, order, o, app)
		CheckDep(t, order, o, clusterInst)
		CheckDep(t, order, o, privacyPolicy)
	}
	// AppInstRefs dependencies
	if o, found := order[appInstRefs]; found {
		CheckDep(t, order, o, app)
		// For auto-prov, AppInsts must be sent before AppInstRefs.
		// This ensures that the health state of AppInsts can be
		// checked when traversing the refs.
		CheckDep(t, order, o, appInst)
		// For auto-prov, Cloudlets must be sent before AppInstRefs.
		// This ensures that the health state of Cloudlets can be
		// checked when traversing the refs.
		CheckDep(t, order, o, cloudlet)
	}
}

func CheckDep(t *testing.T, order map[reflect.Type]int, ord int, dep reflect.Type) {
	depOrd, found := order[dep]
	require.True(t, found)
	require.Greater(t, ord, depOrd)
}
