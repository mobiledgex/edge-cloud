// test etcd process

package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/stretchr/testify/assert"
)

func expectNewRev(t *testing.T, expRev *int64, checkRev int64) {
	*expRev++
	assert.Equal(t, *expRev, checkRev, "revision")
}

func testCalls(t *testing.T, objStore objstore.KVStore) {
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
	var err error
	ctx := context.Background()

	syncCheck := NewSyncCheck(t, objStore)
	defer syncCheck.Stop()

	// check what happens if no put is called
	// no change is made to database, function returns current revision.
	rev, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
		stm.Get("foo")
		stm.Get("bar")
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, expRev, rev, "revision")

	// test create
	for key, val := range m {
		rev, err := objStore.Create(ctx, key, val)
		expectNewRev(t, &expRev, rev)
		assert.Nil(t, err, "Create failed for key %s", key)
		assert.Equal(t, expRev, rev, "revision")
		syncCheck.Expect(t, key, val, expRev)
	}
	_, err = objStore.Create(ctx, key1, val1)
	assert.Equal(t, objstore.ExistsError(key1), err, "Create object that already exists")

	// test get and list
	val, vers, _, err := objStore.Get(key1)
	assert.Nil(t, err, "Get key %s", key1)
	assert.Equal(t, val1, string(val), "Get key %s value", key1)
	assert.EqualValues(t, 1, vers, "version for key %s", key1)
	val, vers, _, err = objStore.Get("No such key")
	assert.Equal(t, objstore.NotFoundError("No such key"), err, "Get non-existent key")

	count = 0
	err = objStore.List("", func(key, val []byte, rev, modRev int64) error {
		count++
		return nil
	})
	assert.Equal(t, 4, count, "List count")

	// test update
	val2 := "app111"
	rev, err = objStore.Update(ctx, key1, val2, 1)
	expectNewRev(t, &expRev, rev)
	assert.Equal(t, expRev, rev, "revision")
	assert.Nil(t, err, "Update existing object")
	syncCheck.Expect(t, key1, val2, expRev)
	val, vers, _, err = objStore.Get(key1)
	assert.Nil(t, err, "Get key %s", key1)
	assert.Equal(t, val2, string(val), "Get key %s updated value", key1)
	assert.EqualValues(t, 2, vers, "version for key %s", key1)
	rev, err = objStore.Update(ctx, key1, val2, 1)
	assert.NotNil(t, err, "Update with wrong mod value")
	rev, err = objStore.Update(ctx, key1, val2, objstore.ObjStoreUpdateVersionAny)
	expectNewRev(t, &expRev, rev)
	assert.Nil(t, err, "Update any version")
	val, vers, _, err = objStore.Get(key1)
	assert.Nil(t, err, "Get key %s", key1)
	assert.EqualValues(t, 3, vers, "version for key %s", key1)

	rev, err = objStore.Update(ctx, "no-such-key", "", 0)
	assert.Equal(t, objstore.NotFoundError("no-such-key"), err, "Update non-existent key")

	// test delete
	rev, err = objStore.Delete(ctx, key1)
	expectNewRev(t, &expRev, rev)
	assert.Nil(t, err, "Delete key %s", key1)
	syncCheck.ExpectNil(t, key1, expRev)
	val, _, _, err = objStore.Get(key1)
	assert.Equal(t, objstore.NotFoundError(key1), err, "Get deleted key")
	count = 0
	err = objStore.List("", func(key, val []byte, rev, modRev int64) error {
		count++
		return nil
	})
	assert.Equal(t, 3, count, "List count")
	assert.Equal(t, 3, len(syncCheck.kv), "sync count")

	// test put
	pkey := "1/foo/adslfk"
	pval := "put value"
	rev, err = objStore.Put(ctx, pkey, pval)
	expectNewRev(t, &expRev, rev)
	assert.Nil(t, err, "Put key %s", pkey)
	syncCheck.Expect(t, pkey, pval, expRev)
	val, vers, _, err = objStore.Get(pkey)
	assert.Nil(t, err, "Get key %s", pkey)
	assert.Equal(t, pval, string(val), "Get key %s value", pkey)
	rev, err = objStore.Put(ctx, pkey, pval)
	expectNewRev(t, &expRev, rev)
	assert.Nil(t, err, "Put key %s again", pkey)

	// debug sync
	syncCheck.Dump()

	fmt.Println("***** test STM ******")
	k0 := "create/key"
	v0 := "create val"
	k1 := "1/App/someapp"
	v1 := "someapp value"
	k2 := "1/App/anotherapp"
	v2 := "anotherapp value"
	ii := 0

	// This tests that doing a Get that returns "" followed by a Put is
	// equivalent to a "Create", which requires the key does not exist.
	//
	// After doing the stm.Get, the function inteferes with itself
	// by directly putting the KV pair (not via STM so bypasses the STM).
	// The subsequent stm.Put should be restrained by the revision (0)
	// of the stm.Get, such that during commit, it will fail.
	// This will trigger a retry, which will run the function again.
	// On retry, the stm.Get check will return "already exists".
	// If the stm.Get + stm.Put was not equivalent to a create, then
	// the put would have succeeded, and the apply would have succeeded
	// and the number of tries (ii) would just be 1.
	rev, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
		ii++
		if stm.Get(k0) != "" {
			return errors.New("already exists")
		}

		fmt.Println("non-stm put interference")
		objStore.Put(ctx, k0, v0)
		expRev++

		stm.Put(k0, v0)
		return nil
	})
	assert.NotNil(t, err)
	assert.Equal(t, 2, ii)

	// This tests that doing a Get that returns non-"" followed by
	// a Put is equivalent to an "Update", which requires the key exist.
	//
	// After doing the stm.Get, the function interferes with itself
	// by directly deleting the KV pair (not via STM so bypasses the STM).
	// The subsequent stm.Put should be restrained by the revision id (1)
	// of the stm.Get, such that during commit, it will fail.
	// This will trigger a retry, which will run the function again.
	// On retry, the stm.Get check will fail with "not found".
	// If the stm.Get + stm.Put was not equivalent to an update, then
	// the put would have succeeded, and the apply would have succeeded
	// and the number of tries (ii) would just be 1.
	// Note that kv pair already exists from previous test before this
	// function starts.
	ii = 0
	rev, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
		ii++
		if stm.Get(k0) == "" {
			return errors.New("not found")
		}

		fmt.Println("non-stm delete interference")
		objStore.Delete(ctx, k0)
		expRev++

		stm.Put(k0, v0)
		return nil
	})
	assert.NotNil(t, err)
	assert.Equal(t, 2, ii)

	// test create of both at the same time.
	rev, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
		if stm.Get(k1) != "" || stm.Get(k2) != "" {
			return errors.New("already exists")
		}
		stm.Put(k1, v1)
		stm.Put(k2, v2)
		return nil
	})
	assert.Nil(t, err)
	expectNewRev(t, &expRev, rev)
	val, _, _, err = objStore.Get(k1)
	assert.Nil(t, err)
	assert.Equal(t, v1, string(val))
	val, _, _, err = objStore.Get(k2)
	assert.Nil(t, err)
	assert.Equal(t, v2, string(val))

	// check that create when it already exists fails
	rev, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
		if stm.Get(k1) != "" || stm.Get(k2) != "" {
			return errors.New("already exists")
		}
		stm.Put(k1, v1)
		stm.Put(k2, v2)
		return nil
	})
	assert.NotNil(t, err)

	// run update
	newval := "some new value"
	rev, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
		if stm.Get(k1) == "" {
			return errors.New("does not exist")
		}
		stm.Put(k1, newval)
		return nil
	})
	assert.Nil(t, err)
	expectNewRev(t, &expRev, rev)
	val, _, _, err = objStore.Get(k1)
	assert.Nil(t, err)
	assert.Equal(t, newval, string(val))

	// check delete
	rev, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
		if stm.Get(k1) == "" || stm.Get(k2) == "" {
			return errors.New("keys do not exist")
		}
		stm.Del(k1)
		stm.Del(k2)
		return nil
	})
	assert.Nil(t, err)
	expectNewRev(t, &expRev, rev)
	// err is "Key not found" because they were deleted
	val, _, _, err = objStore.Get(k1)
	assert.NotNil(t, err)
	val, _, _, err = objStore.Get(k2)
	assert.NotNil(t, err)

	// check all-or-nothing.
	rev, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
		stm.Put(k1, v1)
		if true {
			return errors.New("error out")
		}
		stm.Put(k2, v2)
		return nil
	})
	// neither key should exist
	val, _, _, err = objStore.Get(k1)
	assert.NotNil(t, err)
	val, _, _, err = objStore.Get(k2)
	assert.NotNil(t, err)

	// check what happens if no put is called
	// no change is made to database, function returns current revision.
	rev, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
		stm.Get(k1)
		stm.Get(k2)
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, expRev, rev, "revision")

	// check that get after put succeeds
	rev, err = objStore.ApplySTM(ctx, func(stm concurrency.STM) error {
		if stm.Get(k1) != "" || stm.Get(k2) != "" {
			return errors.New("already exists")
		}
		stm.Put(k1, v1)
		// since k1 was put, this next check should pass
		if stm.Get(k1) == "" {
			return errors.New("put but not found")
		}
		stm.Put(k2, v2)
		return nil
	})
	assert.Nil(t, err)
	expectNewRev(t, &expRev, rev)
	val, _, _, err = objStore.Get(k1)
	assert.Nil(t, err)
	assert.Equal(t, v1, string(val))
	val, _, _, err = objStore.Get(k2)
	assert.Nil(t, err)
	assert.Equal(t, v2, string(val))
}

func TestEtcdDummy(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd)
	dummy := dummyEtcd{}
	dummy.Start()
	testCalls(t, &dummy)
	dummy.Stop()
}

type SyncCheck struct {
	kv         map[string]string
	mux        sync.Mutex
	syncCancel context.CancelFunc
	syncList   map[string]struct{}
	rev        int64
}

func NewSyncCheck(t *testing.T, objstore objstore.KVStore) *SyncCheck {
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

func (s *SyncCheck) Cb(ctx context.Context, data *objstore.SyncCbData) {
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
