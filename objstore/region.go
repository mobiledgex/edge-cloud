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

// Regions
// Each etcd cluster runs in a given region.
// We may have multiple regions world-wide.
// To allow easy expansion of new regions, regions are
// not define programatically. Rather, regions are defined
// by an integer and are translated to a name via mappings
// defined in the etcd database. So adding a new Region is
// as simple as creating a new etcd cluster with a new region
// ID, and adding that region name to the mapping.

package objstore

import "github.com/mobiledgex/edge-cloud/log"

var (
	myRegion uint32 = 0
)

func InitRegion(region uint32) {
	myRegion = region
}

func GetRegion() uint32 {
	if myRegion == 0 {
		log.FatalLog("Region not initialized")
	}
	return myRegion
}
