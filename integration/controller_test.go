package integration

import (
	"context"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/integration/setups"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

// This tests the synchronization of the database between
// controller processes via etcd watch calls.
func testControllerSync(t *testing.T, setup *process.ProcessSetup) {
	numproc := 3
	process.RequireEtcdCount(t, setup, numproc)
	process.RequireControllerCount(t, setup, numproc)
	process.ResetEtcds(t, setup, numproc)
	process.StartEtcds(t, setup, numproc)
	process.StartControllers(t, setup, numproc, process.WithDebug("etcd,api,notify"))
	connectTimeout := 4 * time.Second
	ctrlApis := process.ConnectControllerAPIs(t, setup, numproc, connectTimeout)

	devApis := make([]edgeproto.DeveloperApiClient, numproc)
	for ii := 0; ii < numproc; ii++ {
		devApis[ii] = edgeproto.NewDeveloperApiClient(ctrlApis[ii])
	}

	var err error
	retries := 10
	interval := 10 * time.Millisecond

	// create developers and see that they show up on the other controllers
	for _, dev := range testutil.DevData {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err = devApis[0].CreateDeveloper(ctx, &dev)
		cancel()
		assert.Nil(t, err, "create dev "+dev.Key.GetKeyString())
		for ii := 0; ii < numproc; ii++ {
			testutil.WaitAssertFoundDeveloper(t, devApis[ii], &dev, retries, interval)
		}
	}

	// restart a controller and make sure it resyncs
	ctrlApis[0].Close()
	setup.Controllers[0].Stop()
	setup.Controllers[0].Start("", process.WithDebug("etcd,api,notify"))
	ctrlApis[0], err = setup.Controllers[0].ConnectAPI(connectTimeout)
	assert.Nil(t, err, "reconnect to controller 0")
	devApis[0] = edgeproto.NewDeveloperApiClient(ctrlApis[0])
	for _, dev := range testutil.DevData {
		testutil.WaitAssertFoundDeveloper(t, devApis[0], &dev, retries, interval)
	}

	// update a developer and make sure it syncs
	testutil.DevData[0].Email = "updated email"
	testutil.DevData[0].Fields = []string{edgeproto.DeveloperFieldEmail}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	_, err = devApis[0].UpdateDeveloper(ctx, &testutil.DevData[0])
	cancel()
	testutil.DevData[0].Fields = nil
	for ii := 0; ii < numproc; ii++ {
		testutil.WaitAssertFoundDeveloper(t, devApis[ii], &testutil.DevData[0], retries, interval)
	}

	// delete a developer and make sure it propagates
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	_, err = devApis[0].DeleteDeveloper(ctx, &testutil.DevData[0])
	cancel()
	for ii := 0; ii < numproc; ii++ {
		testutil.WaitAssertNotFoundDeveloper(t, devApis[ii], &testutil.DevData[0], retries, interval)
	}

	process.StopControllers(setup, numproc)
	process.StopEtcds(setup, numproc)
}

func TestControllerSyncLocal(t *testing.T) {
	// The LocalCtrlSync setup has controllers connecting to
	// separate etcd instances, so that synchronization must
	// happen via controllerA -> etcdA -> etcdB -> controllerB
	testControllerSync(t, &setups.LocalCtrlSync)
}

func TestControllerSyncBasic(t *testing.T) {
	testControllerSync(t, &setups.LocalBasic)
}
