package dmetest

import (
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

var platosUniqueId = "MEL-ID"
var AtlanticUniqueId = "AtlanticLabs"

// Device - used to test devices that send registered
var DeviceData = []dme.RegisterClientRequest{
	dme.RegisterClientRequest{
		OrgName:      cloudcommon.Organizationplatos,
		AppName:      cloudcommon.PlatosEnablingLayer,
		AppVers:      "1.1",
		UniqueIdType: platosUniqueId,
		UniqueId:     "device1",
	},
	dme.RegisterClientRequest{
		OrgName:      cloudcommon.Organizationplatos,
		AppName:      cloudcommon.PlatosEnablingLayer,
		AppVers:      "2.1",
		UniqueIdType: platosUniqueId,
		UniqueId:     "device2",
	},
	dme.RegisterClientRequest{
		OrgName:      "Atlantic Labs",
		AppName:      "HarryPotter-go",
		AppVers:      "1.0",
		UniqueIdType: AtlanticUniqueId,
		UniqueId:     "device1",
	},
	// Duplicate Register
	dme.RegisterClientRequest{
		OrgName:      cloudcommon.Organizationplatos,
		AppName:      cloudcommon.PlatosEnablingLayer,
		AppVers:      "2.1",
		UniqueIdType: platosUniqueId,
		UniqueId:     "device2",
	},
}
