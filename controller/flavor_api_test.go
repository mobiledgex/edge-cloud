package main

import (
	"context"
	"testing"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestFlavorApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	testinit()

	dummy := dummyEtcd{}
	dummy.Start()
	ctx := log.StartTestSpan(context.Background())
	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalFlavorTest(t, "cud", &flavorApi, testutil.FlavorData)
	aflav := edgeproto.Flavor{}
	aflav.Key.Name = "mex-med-gpu"
	aflav.Ram = 8192
	aflav.Vcpus = 4
	aflav.Disk = 20
	_, err := flavorApi.CreateFlavor(ctx, &aflav)
	require.Nil(t, err, "CreateFlavor")

	aflav.OptResMap["gpu"] = "1"
	_, err = flavorApi.AddFlavorRes(ctx, &aflav)
	require.Nil(t, err, "AddFlavorRes")
	aflav.OptResMap["nas"] = "200" // Units = MB
	_, err = flavorApi.AddFlavorRes(ctx, &aflav)
	require.Nil(t, err, "AddFlavorRes")

	// fetch the flavor by key and ensure 2 entries
	err = flavorApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !flavorApi.store.STMGet(stm, &aflav.Key, &aflav) {
			return objstore.ErrKVStoreKeyNotFound
		}
		return nil
	})
	require.Equal(t, 2, len(aflav.OptResMap), "AddFlavorRes")
	dflav := testutil.FlavorData[3] // this has no map whatever
	delete(aflav.OptResMap, "nas")
	// all that's left in aflav is 'gpu' so that's what should be deleted.
	require.Equal(t, 1, len(aflav.OptResMap), "AddFlavorRes")
	_, err = flavorApi.DelFlavorRes(ctx, &aflav)
	require.Nil(t, err, "DelFlavorRes")

	// fetch the aflav key into dflav to test contents
	err = flavorApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !flavorApi.store.STMGet(stm, &aflav.Key, &dflav) {
			return objstore.ErrKVStoreKeyNotFound
		}
		return nil
	})

	require.Equal(t, 1, len(dflav.OptResMap), "DelFavorRes")
	for key, _ := range dflav.OptResMap {
		require.Equal(t, "nas", key, "DelFavorRes")
	}
	dummy.Stop()
}
