// test etcd process

package main

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/stretchr/testify/assert"
)

func expectNewRev(t *testing.T, expRev *int64, checkRev int64) {
	*expRev++
	assert.Equal(t, *expRev, checkRev, "revision")
}

func testCalls(t *testing.T, objStore objstore.ObjStore) {
	count := 0
	m := make(map[string]string)
	key1 := "1/1/2222222"
	val1 := "app1"
	m["1/0/123456789"] = "value1"
	m[key1] = val1
	m["1/1/123456789"] = "app2"
	m["1/1/12323445"] = "app3"
	var expRev int64 = 1
	var rev int64

	syncCheck := NewSyncCheck(t, objStore)
	defer syncCheck.Stop()

	// test create
	for key, val := range m {
		rev, err := objStore.Create(key, val)
		expectNewRev(t, &expRev, rev)
		assert.Nil(t, err, "Create failed for key %s", key)
		assert.Equal(t, expRev, rev, "revision")
		syncCheck.Expect(t, key, val, expRev)
	}
	_, err := objStore.Create(key1, val1)
	assert.Equal(t, objstore.ErrObjStoreKeyExists, err, "Create object that already exists")

	// test get and list
	val, vers, err := objStore.Get(key1)
	assert.Nil(t, err, "Get key %s", key1)
	assert.Equal(t, val1, string(val), "Get key %s value", key1)
	assert.EqualValues(t, 1, vers, "version for key %s", key1)
	val, vers, err = objStore.Get("No such key")
	assert.Equal(t, objstore.ErrObjStoreKeyNotFound, err, "Get non-existent key")

	count = 0
	err = objStore.List("", func(key, val []byte, rev int64) error {
		count++
		return nil
	})
	assert.Equal(t, 4, count, "List count")

	// test update
	val2 := "app111"
	rev, err = objStore.Update(key1, val2, 1)
	expectNewRev(t, &expRev, rev)
	assert.Equal(t, expRev, rev, "revision")
	assert.Nil(t, err, "Update existing object")
	syncCheck.Expect(t, key1, val2, expRev)
	val, vers, err = objStore.Get(key1)
	assert.Nil(t, err, "Get key %s", key1)
	assert.Equal(t, val2, string(val), "Get key %s updated value", key1)
	assert.EqualValues(t, 2, vers, "version for key %s", key1)
	rev, err = objStore.Update(key1, val2, 1)
	assert.NotNil(t, err, "Update with wrong mod value")
	rev, err = objStore.Update(key1, val2, objstore.ObjStoreUpdateVersionAny)
	expectNewRev(t, &expRev, rev)
	assert.Nil(t, err, "Update any version")
	val, vers, err = objStore.Get(key1)
	assert.Nil(t, err, "Get key %s", key1)
	assert.EqualValues(t, 3, vers, "version for key %s", key1)

	rev, err = objStore.Update("no-such-key", "", 0)
	assert.Equal(t, objstore.ErrObjStoreKeyNotFound, err, "Update non-existent key")

	// test delete
	rev, err = objStore.Delete(key1)
	expectNewRev(t, &expRev, rev)
	assert.Nil(t, err, "Delete key %s", key1)
	syncCheck.ExpectNil(t, key1, expRev)
	val, _, err = objStore.Get(key1)
	assert.Equal(t, objstore.ErrObjStoreKeyNotFound, err, "Get deleted key")
	count = 0
	err = objStore.List("", func(key, val []byte, rev int64) error {
		count++
		return nil
	})
	assert.Equal(t, 3, count, "List count")
	assert.Equal(t, 3, len(syncCheck.kv), "sync count")

	// test put
	pkey := "1/foo/adslfk"
	pval := "put value"
	rev, err = objStore.Put(pkey, pval)
	expectNewRev(t, &expRev, rev)
	assert.Nil(t, err, "Put key %s", pkey)
	syncCheck.Expect(t, pkey, pval, expRev)
	val, vers, err = objStore.Get(pkey)
	assert.Nil(t, err, "Get key %s", pkey)
	assert.Equal(t, pval, string(val), "Get key %s value", pkey)
	rev, err = objStore.Put(pkey, pval)
	expectNewRev(t, &expRev, rev)
	assert.Nil(t, err, "Put key %s again", pkey)

	// debug sync
	syncCheck.Dump()
}

func TestEtcdDummy(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd)
	dummy := dummyEtcd{}
	dummy.Start()
	testCalls(t, &dummy)
	dummy.Stop()
}

func TestEtcdReal(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd)
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

type SyncCheck struct {
	kv         map[string]string
	mux        sync.Mutex
	syncCancel context.CancelFunc
	syncList   map[string]struct{}
	rev        int64
}

func NewSyncCheck(t *testing.T, objstore objstore.ObjStore) *SyncCheck {
	sy := SyncCheck{}
	sy.kv = make(map[string]string)

	ctx, cancel := context.WithCancel(context.Background())
	sy.syncCancel = cancel
	go func() {
		err := objstore.Sync(ctx, "", sy.Cb)
		assert.Nil(t, err, "Sync error")
	}()
	return &sy
}

func (s *SyncCheck) Stop() {
	s.syncCancel()
}

func (s *SyncCheck) Cb(data *objstore.SyncCbData) {
	s.mux.Lock()
	defer s.mux.Unlock()
	log.InfoLog("sync check cb", "action", objstore.SyncActionStrs[data.Action], "key", string(data.Key), "val", string(data.Value), "rev", data.Rev)
	switch data.Action {
	case objstore.SyncUpdate:
		s.kv[string(data.Key)] = string(data.Value)
		s.rev = data.Rev
	case objstore.SyncDelete:
		delete(s.kv, string(data.Key))
		s.rev = data.Rev
	case objstore.SyncListStart:
		s.syncList = make(map[string]struct{})
	case objstore.SyncList:
		s.kv[string(data.Key)] = string(data.Value)
		s.syncList[string(data.Key)] = struct{}{}
	case objstore.SyncListEnd:
		for key, _ := range s.kv {
			if _, found := s.syncList[key]; !found {
				delete(s.kv, key)
			}
		}
		s.syncList = nil
		s.rev = data.Rev
	}
}

func (s *SyncCheck) WaitRev(rev int64) {
	for ii := 0; ii < 10; ii++ {
		if s.rev == rev {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	log.InfoLog("Wait rev timed out", "rev", rev)
}

func (s *SyncCheck) Expect(t *testing.T, key, val string, rev int64) {
	s.WaitRev(rev)
	s.mux.Lock()
	defer s.mux.Unlock()
	foundVal, found := s.kv[key]
	assert.Equal(t, rev, s.rev, "rev")
	assert.True(t, found, "find", "key", key, "rev", rev)
	assert.Equal(t, foundVal, val, "find", "key", key, "val", val, "rev", rev)
}

func (s *SyncCheck) ExpectNil(t *testing.T, key string, rev int64) {
	s.WaitRev(rev)
	s.mux.Lock()
	defer s.mux.Unlock()
	_, found := s.kv[key]
	assert.Equal(t, rev, s.rev, "rev")
	assert.False(t, found, "not find", "key", key, "rev", rev)
}

func (s *SyncCheck) Dump() {
	s.mux.Lock()
	defer s.mux.Unlock()
	log.InfoLog("sync check rev", "rev", s.rev)
	for key, val := range s.kv {
		log.InfoLog("sync check kv", "key", key, "val", val)
	}
}
