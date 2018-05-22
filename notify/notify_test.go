package notify

import (
	"os"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

type DummySendHandler struct {
	appInsts  map[edgeproto.AppInstKey]edgeproto.AppInst
	cloudlets map[edgeproto.CloudletKey]edgeproto.Cloudlet
}

func NewDummySendHandler() *DummySendHandler {
	handler := &DummySendHandler{}
	handler.appInsts = make(map[edgeproto.AppInstKey]edgeproto.AppInst)
	handler.cloudlets = make(map[edgeproto.CloudletKey]edgeproto.Cloudlet)
	return handler
}

func (s *DummySendHandler) GetAllAppInstKeys(keys map[edgeproto.AppInstKey]bool) {
	for key, _ := range s.appInsts {
		keys[key] = true
	}
}

func (s *DummySendHandler) GetAppInst(key *edgeproto.AppInstKey, buf *edgeproto.AppInst) bool {
	obj, found := s.appInsts[*key]
	if found {
		*buf = obj
	}
	return found
}

func (s *DummySendHandler) GetAllCloudletKeys(keys map[edgeproto.CloudletKey]bool) {
	for key, _ := range s.cloudlets {
		keys[key] = true
	}
}

func (s *DummySendHandler) GetCloudlet(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool {
	obj, found := s.cloudlets[*key]
	if found {
		*buf = obj
	}
	return found
}

func (s *DummySendHandler) CreateAppInst(in *edgeproto.AppInst) {
	s.appInsts[in.Key] = *in
	UpdateAppInst(&in.Key)
}

func (s *DummySendHandler) DeleteAppInst(in *edgeproto.AppInst) {
	delete(s.appInsts, in.Key)
	UpdateAppInst(&in.Key)
}

func (s *DummySendHandler) CreateCloudlet(in *edgeproto.Cloudlet) {
	s.cloudlets[in.Key] = *in
	UpdateCloudlet(&in.Key)
}

func (s *DummySendHandler) DeleteCloudlet(in *edgeproto.Cloudlet) {
	delete(s.cloudlets, in.Key)
	UpdateCloudlet(&in.Key)
}

type DummyRecvHandler struct {
	appInsts           map[edgeproto.AppInstKey]edgeproto.AppInst
	cloudlets          map[edgeproto.CloudletKey]edgeproto.Cloudlet
	numAppInstUpdates  int
	numCloudletUpdates int
	numUpdates         int
}

func NewDummyRecvHandler() *DummyRecvHandler {
	handler := &DummyRecvHandler{}
	handler.appInsts = make(map[edgeproto.AppInstKey]edgeproto.AppInst)
	handler.cloudlets = make(map[edgeproto.CloudletKey]edgeproto.Cloudlet)
	return handler
}

func (s *DummyRecvHandler) HandleSendAllDone(maps *NotifySendAllMaps) {
	for key, _ := range s.appInsts {
		if _, ok := maps.AppInsts[key]; !ok {
			delete(s.appInsts, key)
		}
	}
	for key, _ := range s.cloudlets {
		if _, ok := maps.Cloudlets[key]; !ok {
			delete(s.cloudlets, key)
		}
	}
}

func (s *DummyRecvHandler) HandleNotice(notice *edgeproto.Notice) error {
	appInst := notice.GetAppInst()
	if appInst != nil {
		if notice.Action == edgeproto.NoticeAction_UPDATE {
			s.appInsts[appInst.Key] = *appInst
		} else if notice.Action == edgeproto.NoticeAction_DELETE {
			delete(s.appInsts, appInst.Key)
		}
		s.numAppInstUpdates++
	}
	cloudlet := notice.GetCloudlet()
	if cloudlet != nil {
		if notice.Action == edgeproto.NoticeAction_UPDATE {
			s.cloudlets[cloudlet.Key] = *cloudlet
		} else if notice.Action == edgeproto.NoticeAction_DELETE {
			delete(s.cloudlets, cloudlet.Key)
		}
		s.numCloudletUpdates++
	}
	s.numUpdates++
	return nil
}

func waitForAppInsts(recv *DummyRecvHandler, count int) {
	for i := 0; i < 10; i++ {
		if len(recv.appInsts) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func waitForCloudlets(recv *DummyRecvHandler, count int) {
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
	os.Remove(unixSocket)
	recv := NewNotifyReceiver("unix", unixSocket, recvHandler)
	go recv.Run()

	// Register the receiver so the controller sends notices to it
	addr := "unix:" + unixSocket
	RegisterReceiver(addr, NotifyTypeMatcher)
	// It takes a little while for the Run thread to start up
	// Wait until it's connected
	waitForConnect(recv, 1)
	assert.Equal(t, 0, len(recvHandler.appInsts), "num appInsts")
	assert.Equal(t, 0, recvHandler.numUpdates, "num updates")
	assert.Equal(t, uint64(1), recv.connectionId, "connection id")
	assert.Equal(t, NotifyVersion, recv.version, "version")

	// Create some app insts which will trigger updates
	sendHandler.CreateAppInst(&testutil.AppInstData[0])
	sendHandler.CreateAppInst(&testutil.AppInstData[1])
	sendHandler.CreateAppInst(&testutil.AppInstData[2])
	waitForAppInsts(recvHandler, 3)
	assert.Equal(t, 3, len(recvHandler.appInsts), "num appInsts")
	assert.Equal(t, 3, recvHandler.numAppInstUpdates, "num app inst updates")
	assert.Equal(t, 3, recvHandler.numUpdates, "num updates")
	stats := GetNotifySenderStats(addr)
	assert.Equal(t, uint64(3), stats.AppInstsSent)

	// Kill connection out from under the code, forcing reconnect
	recv.server.Stop()

	// Create another app inst to trigger update/reconnect
	// All cloudlets and all app insts will be sent again
	sendHandler.CreateAppInst(&testutil.AppInstData[3])
	waitForAppInsts(recvHandler, 4)
	assert.Equal(t, 4, len(recvHandler.appInsts), "num appInsts")
	assert.Equal(t, 7, recvHandler.numUpdates, "num updates")
	stats = GetNotifySenderStats(addr)
	assert.Equal(t, uint64(2), stats.Connects)
	assert.Equal(t, uint64(7), stats.AppInstsSent)

	// Delete an inst
	sendHandler.DeleteAppInst(&testutil.AppInstData[0])
	waitForAppInsts(recvHandler, 3)
	assert.Equal(t, 3, len(recvHandler.appInsts), "num appInsts")
	assert.Equal(t, 8, recvHandler.numUpdates, "num updates")
	assert.Equal(t, 8, recvHandler.numAppInstUpdates, "num app inst updates")
	stats = GetNotifySenderStats(addr)
	assert.Equal(t, uint64(2), stats.Connects)
	assert.Equal(t, uint64(8), stats.AppInstsSent)

	// This time stop receiver, delete an inst, then start the
	// receiver again. The recvHandler remains the same so none of
	// the data/stats changes. This tests that a delete during
	// disconnect is properly accounted for during the handling
	// of the sendall done command by removing the stale entry.
	recv.Stop()
	sendHandler.DeleteAppInst(&testutil.AppInstData[1])
	recv = NewNotifyReceiver("unix", unixSocket, recvHandler)
	go recv.Run()
	waitForAppInsts(recvHandler, 2)
	assert.Equal(t, 2, len(recvHandler.appInsts), "num appInsts")
	assert.Equal(t, 10, recvHandler.numAppInstUpdates, "num app inst updates")
	stats = GetNotifySenderStats(addr)
	assert.Equal(t, uint64(3), stats.Connects)
	assert.Equal(t, uint64(10), stats.AppInstsSent)

	UnregisterReceiver(addr)

	// Now test cloudlets. Use the same receiver, but register it
	// as a cloudlet mananger
	RegisterReceiver(addr, NotifyTypeCloudletMgr)
	waitForConnect(recv, 1)
	assert.Equal(t, uint64(1), recv.connectionId, "connection id")
	sendHandler.CreateCloudlet(&testutil.CloudletData[0])
	sendHandler.CreateCloudlet(&testutil.CloudletData[1])
	waitForCloudlets(recvHandler, 2)
	sendHandler.DeleteCloudlet(&testutil.CloudletData[1])
	waitForCloudlets(recvHandler, 1)
	assert.Equal(t, 1, len(recvHandler.cloudlets), "num cloudlets")
	assert.Equal(t, 3, recvHandler.numCloudletUpdates, "num cloudlet updates")
	stats = GetNotifySenderStats(addr)
	assert.Equal(t, uint64(3), stats.CloudletsSent, "sent cloudlets")

	UnregisterReceiver(addr)
	recv.Stop()
}
