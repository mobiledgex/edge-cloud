package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/stretchr/testify/require"
)

var upgradeTestFileLocation = "./upgrade_testfiles"
var upgradeTestFilePreSuffix = "_pre.etcd"
var upgradeTestFilePostSuffix = "_post.etcd"

// Walk testutils data and populate objStore
func buildDbFromTestData(objStore objstore.KVStore, funcName string) error {
	var key, val string
	ctx := context.Background()

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
		if _, err := objStore.Put(ctx, key, val); err != nil {
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
	fileExpected, err := os.Create(upgradeTestFileLocation + "/" + funcName + "_expected.etcd")
	if err != nil {
		return err
	}
	err = objStore.List("", func(key, val []byte, rev, modRev int64) error {
		fileExpected.WriteString(string(key) + "\n")
		fileExpected.WriteString(string(val) + "\n")
		dbObjCount++
		return nil
	})
	if err != nil {
		return err
	}
	fileExpected.Close()
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
	log.SetDebugLevel(log.DebugLevelUpgrade | log.DebugLevelApi)
	objStore := dummyEtcd{}
	objstore.InitRegion(1)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	for ii, fn := range edgeproto.VersionHash_UpgradeFuncs {
		if fn == nil {
			continue
		}
		objStore.Start()
		err := buildDbFromTestData(&objStore, edgeproto.VersionHash_UpgradeFuncNames[ii])
		require.Nil(t, err, "Unable to build db from testData")
		err = edgeproto.RunSingleUpgrade(ctx, &objStore, fn)
		require.Nil(t, err, "Upgrade failed")
		err = compareDbToExpected(&objStore, edgeproto.VersionHash_UpgradeFuncNames[ii])
		require.Nil(t, err, "Unexpected result from upgrade function(%s)", edgeproto.VersionHash_UpgradeFuncNames[ii])
		// Stop it, so it's re-created again
		objStore.Stop()
	}
	//manually test a failure of checkHttpPorts upgrade
	objStore.Start()
	err := buildDbFromTestData(&objStore, "CheckForHttpPortsFail")
	require.Nil(t, err, "Unable to build db from testData")
	err = edgeproto.RunSingleUpgrade(ctx, &objStore, edgeproto.CheckForHttpPorts)
	require.NotNil(t, err, "Upgrade did not fail")
	objStore.Stop()
}
