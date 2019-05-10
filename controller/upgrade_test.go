package main

import (
	"fmt"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

// Walk testutils data and populate objStore
func buildDbFromTestData(objStore objstore.KVStore, funcName string) error {
	if _, ok := testutil.PreUpgradeData[funcName]; !ok {
		return fmt.Errorf("No data to build from for %s", funcName)
	}
	for _, kv := range testutil.PreUpgradeData[funcName] {
		if _, err := objStore.Put(kv.Key, kv.Val); err != nil {
			return err
		}
	}
	return nil
}

// walk testutils data and see if the entries exist in the objstore
func compareDbToExpected(objStore objstore.KVStore, funcName string) error {
	var objCount int
	var testKVs []testutil.KVPair
	var ok bool

	// TODO - rewrite the below to use the files instead of testdata
	if testKVs, ok = testutil.PostUpgradeData[funcName]; !ok {
		return fmt.Errorf("No data to check for %s", funcName)
	}
	// TODO - testdata is to be a map of maps, so check all values in this walk as well
	err := objStore.List("", func(key, val []byte, rev int64) error {
		objCount++
		return nil
	})
	if err != nil {
		return err
	}
	if objCount != len(testKVs) {
		return fmt.Errorf("Number of objects in the etcd db[%d] doesn't match the number of expected objects[%d]\n",
			objCount, len(testKVs))
	}
	for _, kv := range testKVs {
		val, _, _, err := objStore.Get(kv.Key)
		if err != nil {
			return err
		}
		if string(val) != kv.Val {
			return fmt.Errorf("Values don't match for the key <%s> - should be: <%s> found: <%s>",
				kv.Key, kv.Val, string(val))
		}
	}
	return nil
}

// Run each upgrade function after populating dummy etcd with test data.
// Verify that the resulting content in etcd matches expected
func TestAllUpgradeFuncs(t *testing.T) {
	objStore := dummyEtcd{}
	for ii, fn := range edgeproto.VersionHash_UpgradeFuncs {
		if fn == nil {
			continue
		}
		objStore.Start()
		err := buildDbFromTestData(&objStore, edgeproto.VersionHash_UpgradeFuncNames[ii])
		assert.Nil(t, err, "Unable to build db from testData")
		err = edgeproto.RunSingleUpgrade(&objStore, fn)
		assert.Nil(t, err, "Upgrade failed")
		err = compareDbToExpected(&objStore, edgeproto.VersionHash_UpgradeFuncNames[ii])
		assert.Nil(t, err, "Unexpected result from upgrade function")
		// Stop it, so it's re-created again
		objStore.Stop()
	}
}
