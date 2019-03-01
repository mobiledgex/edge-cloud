package cloudcommon

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
)

// go-cmp Options
var IgnoreAdminRole = cmpopts.AcyclicTransformer("removeAdminRole", func(roles []ormapi.Role) []ormapi.Role {
	// remove automatically created admin role
	newroles := make([]ormapi.Role, 0)
	for _, role := range roles {
		if role.Username == "mexadmin" {
			continue
		}
		newroles = append(newroles, role)
	}
	return newroles
})

var IgnoreAdminUser = cmpopts.AcyclicTransformer("removeAdminUser", func(users []ormapi.User) []ormapi.User {
	// remove automatically created super user
	newusers := make([]ormapi.User, 0)
	for _, user := range users {
		if user.Name == "mexadmin" {
			continue
		}
		newusers = append(newusers, user)
	}
	return newusers
})

var IgnoreAppInstUri = cmpopts.AcyclicTransformer("removeAppInstUri", func(inst edgeproto.AppInst) edgeproto.AppInst {
	// Appinstance URIs usually not provisioned, as they are inherited
	// from the cloudlet. However they are provioned for the default
	// appinst. So we cannot use "nocmp". Remove the URIs for
	// non-defaultCloudlets.
	out := inst
	if out.Key.CloudletKey != DefaultCloudletKey {
		out.Uri = ""
	}
	return out
})
