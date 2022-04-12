// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
