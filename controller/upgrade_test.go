package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/stretchr/testify/assert"
)

var upgradeTestFileLocation = "./upgrade_testfiles"
var upgradeTestFilePreSuffix = "_pre.etcd"
var upgradeTestFilePostSuffix = "_post.etcd"

// Walk testutils data and populate objStore
func buildDbFromTestData(objStore objstore.KVStore, funcName string) error {
	var key, val string

	filename := upgradeTestFileLocation + "/" + funcName + upgradeTestFilePreSuffix
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Unable to find preupgrade testdata file at %s", filename)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for {
		if !scanner.Scan() {
			break
		}
		key = scanner.Text()
		if !scanner.Scan() {
			return fmt.Errorf("Improper formatted preupgrade .etcd file - Unmatched key, without a value.")
		}
		val = scanner.Text()
		if _, err := objStore.Put(key, val); err != nil {
			return err
		}
	}
	return nil
}

// walk testutils data and see if the entries exist in the objstore
func compareDbToExpected(objStore objstore.KVStore, funcName string) error {
	var dbObjCount, fileObjCount int

	var key, val string

	filename := upgradeTestFileLocation + "/" + funcName + upgradeTestFilePostSuffix
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Unable to find postupgrade testdata file at %s", filename)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for {
		if !scanner.Scan() {
			break
		}
		key = scanner.Text()
		if !scanner.Scan() {
			return fmt.Errorf("Improper formatted postupgrade .etcd file - Unmatched key, without a value.")
		}
		val = scanner.Text()
		dbVal, _, _, err := objStore.Get(key)
		if err != nil {
			return fmt.Errorf("Unable to get value for key[%s], %v", key, err)
		}
		// data may be in json format or non-json string
		compareDone, err := compareJson(funcName, key, val, string(dbVal))
		if !compareDone {
			err = compareString(funcName, key, val, string(dbVal))
		}
		if err != nil {
			return err
		}
		fileObjCount++
	}
	// count objects in etcd
	err = objStore.List("", func(key, val []byte, rev int64) error {
		dbObjCount++
		return nil
	})
	if err != nil {
		return err
	}
	if fileObjCount != dbObjCount {
		return fmt.Errorf("Number of objects in the etcd db[%d] doesn't match the number of expected objects[%d]\n",
			dbObjCount, fileObjCount)
	}
	return nil
}

func compareJson(funcName, key, expected, actual string) (bool, error) {
	expectedMap := make(map[string]interface{})
	actualMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(expected), &expectedMap)
	if err != nil {
		return false, fmt.Errorf("Unmarshal failed, %v, %s\n", err, expected)
	}
	err = json.Unmarshal([]byte(actual), &actualMap)
	if err != nil {
		return false, fmt.Errorf("Unmarshal failed, %v, %s\n", err, actual)
	}
	if !cmp.Equal(expectedMap, actualMap) {
		fmt.Printf("[%s] comparsion fail for key: %s\n", funcName, key)
		fmt.Printf("expected vs actual:\n")
		fmt.Printf(cmp.Diff(expectedMap, actualMap))
		return true, fmt.Errorf("Values don't match for the key, upgradeFunc: %s", funcName)
	}
	return true, nil
}

func compareString(funcName, key, expected, actual string) error {
	if expected != actual {
		fmt.Printf("[%s] values don't match for the key: %s\n", funcName, key)
		fmt.Printf("[%s] expected: \n%s\n", funcName, expected)
		fmt.Printf("[%s] actual: \n%s\n", funcName, actual)
		return fmt.Errorf("Values don't match for the key, upgradeFunc: %s", funcName)
	}
	return nil
}

// Run each upgrade function after populating dummy etcd with test data.
// Verify that the resulting content in etcd matches expected
func TestAllUpgradeFuncs(t *testing.T) {
	objStore := dummyEtcd{}
	objstore.InitRegion(1)
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
