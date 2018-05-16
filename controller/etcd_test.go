// test etcd process

package main

import (
	"os"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

func testCalls(t *testing.T, objStore edgeproto.ObjStore) {
	count := 0
	m := make(map[string]string)
	key1 := "1/1/2222222"
	val1 := "app1"
	m["1/0/123456789"] = "value1"
	m[key1] = val1
	m["1/1/123456789"] = "app2"
	m["1/1/12323445"] = "app3"

	// test create
	for key, val := range m {
		err := objStore.Create(key, val)
		assert.Nil(t, err, "Create failed for key %s", key)
	}
	err := objStore.Create(key1, val1)
	assert.Equal(t, edgeproto.ObjStoreErrKeyExists, err, "Create object that already exists")

	// test get and list
	val, vers, err := objStore.Get(key1)
	assert.Nil(t, err, "Get key %s", key1)
	assert.Equal(t, val1, string(val), "Get key %s value", key1)
	assert.EqualValues(t, 1, vers, "version for key %s", key1)
	val, vers, err = objStore.Get("No such key")
	assert.Equal(t, edgeproto.ObjStoreErrKeyNotFound, err, "Get non-existent key")

	count = 0
	err = objStore.List("", func(key, val []byte) error {
		count++
		return nil
	})
	assert.Equal(t, 4, count, "List count")

	// test update
	val2 := "app111"
	err = objStore.Update(key1, val2, 1)
	assert.Nil(t, err, "Update existing object")
	val, vers, err = objStore.Get(key1)
	assert.Nil(t, err, "Get key %s", key1)
	assert.Equal(t, val2, string(val), "Get key %s updated value", key1)
	assert.EqualValues(t, 2, vers, "version for key %s", key1)
	err = objStore.Update(key1, val2, 1)
	assert.NotNil(t, err, "Update with wrong mod value")
	err = objStore.Update(key1, val2, edgeproto.ObjStoreUpdateVersionAny)
	assert.Nil(t, err, "Update any version")
	val, vers, err = objStore.Get(key1)
	assert.Nil(t, err, "Get key %s", key1)
	assert.EqualValues(t, 3, vers, "version for key %s", key1)

	err = objStore.Update("no-such-key", "", 0)
	assert.Equal(t, edgeproto.ObjStoreErrKeyNotFound, err, "Update non-existent key")

	// test delete
	err = objStore.Delete(key1)
	assert.Nil(t, err, "Delete key %s", key1)
	val, _, err = objStore.Get(key1)
	assert.Equal(t, edgeproto.ObjStoreErrKeyNotFound, err, "Get deleted key")
	count = 0
	err = objStore.List("", func(key, val []byte) error {
		count++
		return nil
	})
	assert.Equal(t, 3, count, "List count")
}

func TestEtcdDummy(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd)
	dummy := dummyEtcd{}
	dummy.Start()
	testCalls(t, &dummy)
	dummy.Stop()
}

func TestEtcdReal(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd)
	etcd, err := StartLocalEtcdServer()
	assert.Nil(t, err, "Etcd start")
	if err != nil {
		return
	}
	_, err = os.Stat(etcd.Config.LogFile)
	assert.Nil(t, err, "Stat log file %s", etcd.Config.LogFile)

	objStore, err := GetEtcdClientBasic(etcd.Config.ClientUrls)
	assert.Nil(t, err, "Etcd client")
	if err != nil {
		return
	}
	testCalls(t, objStore)

	etcd.Stop()
	if _, err := os.Stat(etcd.Config.LogFile); !os.IsNotExist(err) {
		t.Errorf("Etcd logfile still present after cleanup: %s", err)
	}
	if _, err := os.Stat(etcd.Config.DataDir); !os.IsNotExist(err) {
		// this indicates etcd process is probably still running
		t.Errorf("testdir still present after cleanup: %s", err)
	}
}
