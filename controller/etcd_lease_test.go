// +build slow
// To run this unit test, run:
// go test -tags=slow
// or:
// go test -run TestEtcdLease -v -tags=slow

package main

import (
	"context"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEtcdLease(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd)
	etcd, err := StartLocalEtcdServer()
	require.Nil(t, err, "Etcd start")
	defer etcd.Stop()

	objStore, err := GetEtcdClientBasic(etcd.Config.ClientUrls)
	require.Nil(t, err, "Etcd client")

	lease, err := objStore.Grant(context.Background(), 1)

	// put key with lease
	key := "key1"
	val := "val1"
	_, err = objStore.Put(key, val, objstore.WithLease(lease))
	assert.Nil(t, err, "put with lease")

	// key should exist right after putting it.
	bval, _, _, err := objStore.Get(key)
	assert.Nil(t, err, "get key")
	assert.Equal(t, val, string(bval), "val check")

	// spawn keepalive to keep key alive
	ctx, cancel := context.WithCancel(context.Background())
	var kperr error
	go func() {
		kperr = objStore.KeepAlive(ctx, lease)
	}()

	// key should still exist
	time.Sleep(6 * time.Second)
	bval, _, _, err = objStore.Get(key)
	assert.Nil(t, err, "get key")
	assert.Equal(t, val, string(bval), "val check")

	// cancel keepalive
	cancel()

	// wait 3 seconds, then key should be revoked
	time.Sleep(3 * time.Second)
	_, _, _, err = objStore.Get(key)
	assert.Equal(t, objstore.ErrKVStoreKeyNotFound, err, "check expired")

	assert.Nil(t, kperr, "keepalive error")
}
