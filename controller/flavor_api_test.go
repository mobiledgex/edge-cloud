package main

import (
	"context"

	"testing"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestFlavorApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	testinit()
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalFlavorTest(t, "cud", &flavorApi, testutil.FlavorData)
	testMasterFlavor(t, ctx)
	dummy.Stop()
}

func testMasterFlavor(t *testing.T, ctx context.Context) {
	// We optionally maintain one generic modestly sized flavor for use
	// by the MasterNode of a nominal k8s cluster where numnodes (workers)
	// > 0 such that we don't run client workloads on that master. We can therefore
	// use a flavor size sufficent for that purpose only.
	// This mex flavor is created by the mexadmin when setting up a cloudlet that offers
	// optional resources that should not be requested by the master node.
	// The Name of the this flavor is stored in settings.MasterNodeFlavor, and in cases
	// of clusterInst creation per above, the name stored in settings will be looked up and
	// expected to exist. If not, the given nodeflavor in create cluster inst is used as was
	// prior to EC-1767
	var err error

	// ensure the master node default flavor is created, using the default value
	// of settings.MasterNodeFlavor
	cl := testutil.CloudletData[1]
	var cli edgeproto.CloudletInfo = testutil.CloudletInfoData[0]
	settings := settingsApi.Get()
	masterFlavor := edgeproto.Flavor{}
	flavorKey := edgeproto.FlavorKey{}
	flavorKey.Name = settings.MasterNodeFlavor

	err = cloudletApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !flavorApi.store.STMGet(stm, &flavorKey, &masterFlavor) {
			// create the missing flavor
			masterFlavor.Key.Name = "MasterNodeFlavor"
			masterFlavor.Vcpus = 2
			masterFlavor.Disk = 40
			masterFlavor.Ram = 4096
			_, err = flavorApi.CreateFlavor(ctx, &masterFlavor)
			require.Nil(t, err, "Create Master Node Flavor")
		}

		vmspec, err := resTagTableApi.GetVMSpec(ctx, stm, masterFlavor, cl, cli)
		require.Nil(t, err, "GetVmSpec masterNodeFlavor")
		require.Equal(t, "flavor.medium", vmspec.FlavorName)

		return nil
	})
}
