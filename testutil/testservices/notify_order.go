// This file is not in the same package as notify to avoid including
// the testing packages in the notify package.
package testservices

import (
	"fmt"
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
	vmPool := reflect.TypeOf((*notify.VMPoolSendMany)(nil))
	gpuDriver := reflect.TypeOf((*notify.GPUDriverSendMany)(nil))
	network := reflect.TypeOf((*notify.NetworkSendMany)(nil))
	cloudlet := reflect.TypeOf((*notify.CloudletSendMany)(nil))
	clusterInst := reflect.TypeOf((*notify.ClusterInstSendMany)(nil))
	app := reflect.TypeOf((*notify.AppSendMany)(nil))
	TrustPolicy := reflect.TypeOf((*notify.TrustPolicySendMany)(nil))
	autoScalePolicy := reflect.TypeOf((*notify.AutoScalePolicySendMany)(nil))
	autoProvPolicy := reflect.TypeOf((*notify.AutoProvPolicySendMany)(nil))
	appInst := reflect.TypeOf((*notify.AppInstSendMany)(nil))
	appInstRefs := reflect.TypeOf((*notify.AppInstRefsSendMany)(nil))

	// Cloudlet dependencies
	if o, found := order[cloudlet]; found {
		CheckDep(t, "cloudlet", order, o, flavor)
		CheckDep(t, "cloudlet", order, o, vmPool)
		CheckDep(t, "cloudlet", order, o, gpuDriver)
	}
	// ClusterInst dependencies
	if o, found := order[clusterInst]; found {
		CheckDep(t, "clusterinst", order, o, flavor)
		CheckDep(t, "clusterinst", order, o, cloudlet)
		CheckDep(t, "clusterinst", order, o, autoScalePolicy)
		CheckDep(t, "clusterinst", order, o, TrustPolicy)
		CheckDep(t, "clusterinst", order, o, network)
	}
	// App dependecies
	if o, found := order[app]; found {
		CheckDep(t, "app", order, o, flavor)
		CheckDep(t, "app", order, o, autoProvPolicy)
	}
	// AppInst dependencies
	if o, found := order[appInst]; found {
		CheckDep(t, "appinst", order, o, flavor)
		CheckDep(t, "appinst", order, o, app)
		CheckDep(t, "appinst", order, o, clusterInst)
		CheckDep(t, "appinst", order, o, TrustPolicy)
	}
	// AppInstRefs dependencies
	if o, found := order[appInstRefs]; found {
		CheckDep(t, "appinstrefs", order, o, app)
		// For auto-prov, AppInsts must be sent before AppInstRefs.
		// This ensures that the health state of AppInsts can be
		// checked when traversing the refs.
		CheckDep(t, "appinstrefs", order, o, appInst)
		// For auto-prov, Cloudlets must be sent before AppInstRefs.
		// This ensures that the health state of Cloudlets can be
		// checked when traversing the refs.
		CheckDep(t, "appinstrefs", order, o, cloudlet)
	}
}

func CheckDep(t *testing.T, typ string, order map[reflect.Type]int, ord int, dep reflect.Type) {
	depOrd, found := order[dep]
	if !found {
		fmt.Printf("Warning: missing %s dep %v\n", typ, dep)
		return
	}
	require.Greater(t, ord, depOrd)
}
