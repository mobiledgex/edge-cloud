// Get key for each object type to save to etcd.
// This keeps key generation all in one place so that
// we don't have conflicting keys.

// Etcd database is just lists of key-value pairs, each one an object.
// Key format is region/object-type/object-id.
// Value is a string representation of the object.
// Region is an integer.
// object-type is an integer (or string) - see below.
// object-id is typically uint64.

// Because different regions will have different etcd databases,
// different regions may have overlapping object-ids.
// We should be careful in any algorithms that use objects
// from different regions to include the region id in any
// look-up keys or look-up functions.

package main

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/util"
)

type KeyType int

// Do not change the enum values because they are used in the key string.
// When adding a new enum, be sure to add it to the unit test to
// enforce the correct type to value mapping.
const (
	RegionType    KeyType = 0
	AppType       KeyType = 1
	CloudletType  KeyType = 2
	DeveloperType KeyType = 3
	OperatorType  KeyType = 4
	AppInstType   KeyType = 5
)

var (
	keyTypeStrs = []string{
		"region",
		"app",
		"cloudlet",
		"developer",
		"operator",
		"appinst",
	}
)

func checkRange(val KeyType) {
	if val < 0 || int(val) >= len(keyTypeStrs) {
		util.FatalLog("Key type out of range", "val", val)
	}
}

func (val KeyType) String() string {
	checkRange(val)
	return keyTypeStrs[val]
}

func GetObjStoreKey(keytype KeyType, id string) string {
	return fmt.Sprintf("%d/%d/%s", GetRegion(), keytype, id)
}
