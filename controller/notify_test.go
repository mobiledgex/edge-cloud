package main

import (
	"context"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/proto"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

type DummyHandler struct {
	appInsts           map[proto.AppInstKey]proto.AppInst
	cloudlets          map[proto.CloudletKey]proto.Cloudlet
	numAppInstUpdates  int
	numCloudletUpdates int
	numUpdates         int
}

func NewDummyHandler() *DummyHandler {
	handler := &DummyHandler{}
	handler.appInsts = make(map[proto.AppInstKey]proto.AppInst)
	handler.cloudlets = make(map[proto.CloudletKey]proto.Cloudlet)
	return handler
}

func (s *DummyHandler) HandleSendAllDone(maps *NotifySendAllMaps) {
	for key, _ := range s.appInsts {
		if _, ok := maps.appInsts[key]; !ok {
			delete(s.appInsts, key)
		}
	}
	for key, _ := range s.cloudlets {
		if _, ok := maps.cloudlets[key]; !ok {
			delete(s.cloudlets, key)
		}
	}
}

func (s *DummyHandler) HandleNotice(notice *proto.Notice) error {
	appInst := notice.GetAppInst()
	if appInst != nil {
		if notice.Action == proto.NoticeAction_UPDATE {
			s.appInsts[appInst.Key] = *appInst
		} else if notice.Action == proto.NoticeAction_DELETE {
			delete(s.appInsts, appInst.Key)
		}
		s.numAppInstUpdates++
	}
	cloudlet := notice.GetCloudlet()
	if cloudlet != nil {
		if notice.Action == proto.NoticeAction_UPDATE {
			s.cloudlets[cloudlet.Key] = *cloudlet
		} else if notice.Action == proto.NoticeAction_DELETE {
			delete(s.cloudlets, cloudlet.Key)
		}
		s.numCloudletUpdates++
	}
	s.numUpdates++
	return nil
}

func waitForAppInsts(recv *DummyHandler, count int) {
	for i := 0; i < 10; i++ {
		if len(recv.appInsts) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func waitForCloudlets(recv *DummyHandler, count int) {
	for i := 0; i < 10; i++ {
		if len(recv.cloudlets) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func waitForConnect(recv *NotifyReceiver, connect uint64) {
	for i := 0; i < 10; i++ {
		if recv.connectionId == connect {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestNotify(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelApi | util.DebugLevelNotify)
	InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	// override retry time
	NotifyRetryTime = 10 * time.Millisecond

	operatorApi := InitOperatorApi(&dummy)
	cloudletApi := InitCloudletApi(&dummy, operatorApi)
	developerApi := InitDeveloperApi(&dummy)
	appApi := InitAppApi(&dummy, developerApi)
	appInstApi := InitAppInstApi(&dummy, appApi, cloudletApi)
	InitNotifySenders(appInstApi, cloudletApi)

	// create supporting data
	ctx := context.TODO()
	for _, obj := range DevData {
		_, err := developerApi.CreateDeveloper(ctx, &obj)
		assert.Nil(t, err, "Create developer")
	}
	for _, obj := range AppData {
		_, err := appApi.CreateApp(ctx, &obj)
		assert.Nil(t, err, "Create app")
	}
	for _, obj := range OperatorData {
		_, err := operatorApi.CreateOperator(ctx, &obj)
		assert.Nil(t, err, "Create operator")
	}
	for _, obj := range CloudletData {
		_, err := cloudletApi.CreateCloudlet(ctx, &obj)
		assert.Nil(t, err, "Create cloudlet")
	}

	// This tests the Sender (controller) sending notices to
	// a Receiver. The controller as initialized above is the
	// sender. The receiver is dummy'd out below.

	// This is the receiver
	handler := NewDummyHandler()
	unixSocket := "/var/tmp/testNotify.socket"
	recv := NewNotifyReceiver("unix", unixSocket, handler)
	go recv.Run()

	// Register the receiver so the controller sends notices to it
	addr := "unix:" + unixSocket
	RegisterReceiver(addr, NotifyTypeMatcher)
	// It takes a little while for the Run thread to start up
	// Wait until it's connected
	waitForConnect(recv, 1)
	assert.Equal(t, 0, len(handler.appInsts), "num appInsts")
	assert.Equal(t, 0, handler.numUpdates, "num updates")
	assert.Equal(t, uint64(1), recv.connectionId, "connection id")
	assert.Equal(t, NotifyVersion, recv.version, "version")

	// Create some app insts which will trigger updates
	_, _ = appInstApi.CreateAppInst(ctx, &AppInstData[0])
	_, _ = appInstApi.CreateAppInst(ctx, &AppInstData[1])
	_, _ = appInstApi.CreateAppInst(ctx, &AppInstData[2])
	waitForAppInsts(handler, 3)
	assert.Equal(t, 3, len(handler.appInsts), "num appInsts")
	assert.Equal(t, 3, handler.numAppInstUpdates, "num app inst updates")
	assert.Equal(t, 3, handler.numUpdates, "num updates")
	stats := GetNotifySenderStats(addr)
	assert.Equal(t, uint64(3), stats.AppInstsSent)

	// Kill connection out from under the code, forcing reconnect
	recv.server.Stop()

	// Create another app inst to trigger update/reconnect
	// All cloudlets and all app insts will be sent again
	_, _ = appInstApi.CreateAppInst(ctx, &AppInstData[3])
	waitForAppInsts(handler, 4)
	assert.Equal(t, 4, len(handler.appInsts), "num appInsts")
	assert.Equal(t, 7, handler.numUpdates, "num updates")
	stats = GetNotifySenderStats(addr)
	assert.Equal(t, uint64(2), stats.Connects)
	assert.Equal(t, uint64(7), stats.AppInstsSent)

	// Delete an inst
	_, _ = appInstApi.DeleteAppInst(ctx, &AppInstData[0])
	waitForAppInsts(handler, 3)
	assert.Equal(t, 3, len(handler.appInsts), "num appInsts")
	assert.Equal(t, 8, handler.numUpdates, "num updates")
	assert.Equal(t, 8, handler.numAppInstUpdates, "num app inst updates")
	stats = GetNotifySenderStats(addr)
	assert.Equal(t, uint64(2), stats.Connects)
	assert.Equal(t, uint64(8), stats.AppInstsSent)

	// This time stop receiver, delete an inst, then start the
	// receiver again. The handler remains the same so none of
	// the data/stats changes. This tests that a delete during
	// disconnect is properly accounted for during the handling
	// of the sendall done command by removing the stale entry.
	recv.Stop()
	_, _ = appInstApi.DeleteAppInst(ctx, &AppInstData[1])
	recv = NewNotifyReceiver("unix", unixSocket, handler)
	go recv.Run()
	waitForAppInsts(handler, 2)
	assert.Equal(t, 2, len(handler.appInsts), "num appInsts")
	assert.Equal(t, 10, handler.numAppInstUpdates, "num app inst updates")
	stats = GetNotifySenderStats(addr)
	assert.Equal(t, uint64(3), stats.Connects)
	assert.Equal(t, uint64(10), stats.AppInstsSent)

	UnregisterReceiver(addr)
	recv.Stop()
	dummy.Stop()
}
