// Regions
// Each etcd cluster runs in a given region.
// We may have multiple regions world-wide.
// To allow easy expansion of new regions, regions are
// not define programatically. Rather, regions are defined
// by an integer and are translated to a name via mappings
// defined in the etcd database. So adding a new Region is
// as simple as creating a new etcd cluster with a new region
// ID, and adding that region name to the mapping.

package main

import "github.com/mobiledgex/edge-cloud/util"

var (
	myRegion uint32 = 0
)

func InitRegion(region uint32) {
	myRegion = region
}

func GetRegion() uint32 {
	if myRegion == 0 {
		util.FatalLog("Region not initialized")
	}
	return myRegion
}
