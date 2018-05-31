package notify

import (
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

func TestNotify(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelNotify)

	// override retry time
	NotifyRetryTime = 10 * time.Millisecond

	// This tests the sender sending notices to
	// a Receiver. The controller as initialized above is the
	// sender. The receiver is dummy'd out below.

	// Set up the sender
	sendHandler := NewDummySendHandler()
	InitNotifySenders(sendHandler)

	// This is the receiver
	recvHandler := NewDummyRecvHandler()
	unixSocket := "/var/tmp/testNotify.socket"
	recvHandler.Start("unix", unixSocket)

	// Register the receiver so the controller sends notices to it
	addr := "unix:" + unixSocket
	RegisterReceiver(addr, NotifyTypeMatcher)
	// It takes a little while for the Run thread to start up
	// Wait until it's connected
	recvHandler.WaitForConnect(1)
	assert.Equal(t, 0, len(recvHandler.AppInsts), "num appInsts")
	assert.Equal(t, 0, recvHandler.NumUpdates, "num updates")
	assert.Equal(t, uint64(1), recvHandler.Recv.connectionId, "connection id")
	assert.Equal(t, NotifyVersion, recvHandler.Recv.version, "version")

	// Create some app insts which will trigger updates
	sendHandler.CreateAppInst(&testutil.AppInstData[0])
	sendHandler.CreateAppInst(&testutil.AppInstData[1])
	sendHandler.CreateAppInst(&testutil.AppInstData[2])
	recvHandler.WaitForAppInsts(3)
	assert.Equal(t, 3, len(recvHandler.AppInsts), "num appInsts")
	assert.Equal(t, 3, recvHandler.NumAppInstUpdates, "num app inst updates")
	assert.Equal(t, 3, recvHandler.NumUpdates, "num updates")
	stats := GetNotifySenderStats(addr)
	assert.Equal(t, uint64(3), stats.AppInstsSent)

	// Kill connection out from under the code, forcing reconnect
	recvHandler.Recv.server.Stop()

	// Create another app inst to trigger update/reconnect
	// All cloudlets and all app insts will be sent again
	sendHandler.CreateAppInst(&testutil.AppInstData[3])
	recvHandler.WaitForAppInsts(4)
	assert.Equal(t, 4, len(recvHandler.AppInsts), "num appInsts")
	assert.Equal(t, 7, recvHandler.NumUpdates, "num updates")
	stats = GetNotifySenderStats(addr)
	assert.Equal(t, uint64(2), stats.Connects)
	assert.Equal(t, uint64(7), stats.AppInstsSent)

	// Delete an inst
	sendHandler.DeleteAppInst(&testutil.AppInstData[0])
	recvHandler.WaitForAppInsts(3)
	assert.Equal(t, 3, len(recvHandler.AppInsts), "num appInsts")
	assert.Equal(t, 8, recvHandler.NumUpdates, "num updates")
	assert.Equal(t, 8, recvHandler.NumAppInstUpdates, "num app inst updates")
	stats = GetNotifySenderStats(addr)
	assert.Equal(t, uint64(2), stats.Connects)
	assert.Equal(t, uint64(8), stats.AppInstsSent)

	// This time stop receiver, delete an inst, then start the
	// receiver again. The recvHandler remains the same so none of
	// the data/stats changes. This tests that a delete during
	// disconnect is properly accounted for during the handling
	// of the sendall done command by removing the stale entry.
	recvHandler.Recv.Stop()
	sendHandler.DeleteAppInst(&testutil.AppInstData[1])
	recvHandler.Start("unix", unixSocket)
	recvHandler.WaitForAppInsts(2)
	assert.Equal(t, 2, len(recvHandler.AppInsts), "num appInsts")
	assert.Equal(t, 10, recvHandler.NumAppInstUpdates, "num app inst updates")
	stats = GetNotifySenderStats(addr)
	assert.Equal(t, uint64(3), stats.Connects)
	assert.Equal(t, uint64(10), stats.AppInstsSent)

	UnregisterReceiver(addr)

	// Now test cloudlets. Use the same receiver, but register it
	// as a cloudlet mananger
	RegisterReceiver(addr, NotifyTypeCloudletMgr)
	recvHandler.WaitForConnect(1)
	assert.Equal(t, uint64(1), recvHandler.Recv.connectionId, "connection id")
	sendHandler.CreateCloudlet(&testutil.CloudletData[0])
	sendHandler.CreateCloudlet(&testutil.CloudletData[1])
	recvHandler.WaitForCloudlets(2)
	sendHandler.DeleteCloudlet(&testutil.CloudletData[1])
	recvHandler.WaitForCloudlets(1)
	assert.Equal(t, 1, len(recvHandler.Cloudlets), "num cloudlets")
	assert.Equal(t, 3, recvHandler.NumCloudletUpdates, "num cloudlet updates")
	stats = GetNotifySenderStats(addr)
	assert.Equal(t, uint64(3), stats.CloudletsSent, "sent cloudlets")

	UnregisterReceiver(addr)
	recvHandler.Stop()
}
