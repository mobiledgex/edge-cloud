package dmetest

import (
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

var SamsungUniqueId = "MEL-ID"
var NianticUniqueId = "NianticLabs"

// Device - used to test devices that send registered
var DeviceData = []dme.RegisterClientRequest{
	dme.RegisterClientRequest{
		OrgName:      cloudcommon.OrganizationSamsung,
		AppName:      cloudcommon.SamsungEnablingLayer,
		AppVers:      "1.1",
		UniqueIdType: SamsungUniqueId,
		UniqueId:     "device1",
	},
	dme.RegisterClientRequest{
		OrgName:      cloudcommon.OrganizationSamsung,
		AppName:      cloudcommon.SamsungEnablingLayer,
		AppVers:      "2.1",
		UniqueIdType: SamsungUniqueId,
		UniqueId:     "device2",
	},
	dme.RegisterClientRequest{
		OrgName:      "Niantic Labs",
		AppName:      "HarryPotter-go",
		AppVers:      "1.0",
		UniqueIdType: NianticUniqueId,
		UniqueId:     "device1",
	},
	// Duplicate Register
	dme.RegisterClientRequest{
		OrgName:      cloudcommon.OrganizationSamsung,
		AppName:      cloudcommon.SamsungEnablingLayer,
		AppVers:      "2.1",
		UniqueIdType: SamsungUniqueId,
		UniqueId:     "device2",
	},
}
