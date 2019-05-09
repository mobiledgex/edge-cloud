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
	for _, kv := range testutil.PreUpgradeData[funcName] {
		if _, err := objStore.Put(kv.Key, kv.Val); err != nil {
			return err
		}
	}
	return nil
}

// walk testutils data and see if the entries exist in the objstore
func compareDbToExpected(objStore objstore.KVStore, funcName string) error {
	for _, kv := range testutil.PostUpgradeData[funcName] {
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
