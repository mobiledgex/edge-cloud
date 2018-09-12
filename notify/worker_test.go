package notify

import (
	"sync"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	mgr := NewWorkMgr()
	changes := newChanges()

	mgr.SetChangedCb(TypeAppInst, edgeproto.AppInstGenericNotifyCb(changes.appInstChanged))
	mgr.SetChangedCb(TypeClusterInst, edgeproto.ClusterInstGenericNotifyCb(changes.clusterInstChanged))

	// spawn a task. It will wait since go channel is empty.
	mgr.AppInstChanged(&testutil.AppInstData[0].Key, &testutil.AppInstData[0])
	assert.Equal(t, 1, len(mgr.workers))
	<-changes.appInstWorkStarted
	assert.Equal(t, 1, len(changes.curAppInst))
	assert.Equal(t, 0, changes.appInstChanges)
	checkCurAppInst(t, changes, &testutil.AppInstData[0].Key, &testutil.AppInstData[0])

	// create another two tasks with the same key. Because it's the same
	// key, they will get absorbed behind the current blocked task.
	mgr.AppInstChanged(&testutil.AppInstData[0].Key, &testutil.AppInstData[1])
	mgr.AppInstChanged(&testutil.AppInstData[0].Key, &testutil.AppInstData[2])
	assert.Equal(t, 1, len(mgr.workers))
	assert.Equal(t, 1, len(changes.curAppInst))
	assert.Equal(t, 0, changes.appInstChanges)
	checkCurAppInst(t, changes, &testutil.AppInstData[0].Key, &testutil.AppInstData[0])

	// trigger first task to finish. Another task will be run because
	// changes were queued.
	changes.appInstWorkGo <- true
	<-changes.appInstWorkStarted
	<-changes.appInstWorkDone
	assert.Equal(t, 1, len(mgr.workers))
	assert.Equal(t, 1, len(changes.curAppInst))
	assert.Equal(t, 1, changes.appInstChanges)
	// key is same, but old should be first old after blocked task
	checkCurAppInst(t, changes, &testutil.AppInstData[0].Key, &testutil.AppInstData[1])

	// trigger task to finish. Because two tasks were absorbed into one
	// change, no more work is needed.
	changes.appInstWorkGo <- true
	<-changes.appInstWorkDone
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, 0, len(mgr.workers))
	assert.Equal(t, 0, len(changes.curAppInst))
	assert.Equal(t, 2, changes.appInstChanges)

	// reset changes
	changes.appInstChanges = 0

	// Now add multiple keys. They should run in parallel (two workers)
	mgr.AppInstChanged(&testutil.AppInstData[0].Key, &testutil.AppInstData[0])
	mgr.AppInstChanged(&testutil.AppInstData[1].Key, &testutil.AppInstData[1])
	<-changes.appInstWorkStarted
	<-changes.appInstWorkStarted
	assert.Equal(t, 2, len(mgr.workers))
	assert.Equal(t, 2, len(changes.curAppInst))
	assert.Equal(t, 0, changes.appInstChanges)
	changes.appInstWorkGo <- true
	<-changes.appInstWorkDone
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, 1, len(mgr.workers))
	assert.Equal(t, 1, len(changes.curAppInst))
	assert.Equal(t, 1, changes.appInstChanges)
	changes.appInstWorkGo <- true
	<-changes.appInstWorkDone
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, 0, len(mgr.workers))
	assert.Equal(t, 0, len(changes.curAppInst))
	assert.Equal(t, 2, changes.appInstChanges)

	// Add multiple of single key
	mgr.ClusterInstChanged(&testutil.ClusterInstData[0].Key, &testutil.ClusterInstData[0])
	// wait until first thread is spawned
	<-changes.clusterInstWorkStarted
	mgr.ClusterInstChanged(&testutil.ClusterInstData[0].Key, &testutil.ClusterInstData[3])
	mgr.ClusterInstChanged(&testutil.ClusterInstData[0].Key, &testutil.ClusterInstData[1])
	mgr.ClusterInstChanged(&testutil.ClusterInstData[0].Key, &testutil.ClusterInstData[2])
	// Add multiple keys
	mgr.ClusterInstChanged(&testutil.ClusterInstData[1].Key, &testutil.ClusterInstData[1])
	mgr.ClusterInstChanged(&testutil.ClusterInstData[2].Key, &testutil.ClusterInstData[2])
	// There should be 3 threads
	assert.Equal(t, 3, len(mgr.workers))
	<-changes.clusterInstWorkStarted
	<-changes.clusterInstWorkStarted
	assert.Equal(t, 3, len(changes.curClusterInst))
	assert.Equal(t, 0, changes.clusterInstChanges)
	checkCurClusterInst(t, changes, &testutil.ClusterInstData[0].Key, &testutil.ClusterInstData[0])
	checkCurClusterInst(t, changes, &testutil.ClusterInstData[1].Key, &testutil.ClusterInstData[1])
	checkCurClusterInst(t, changes, &testutil.ClusterInstData[2].Key, &testutil.ClusterInstData[2])
	// Trigger all three. Note that multiple changes are behind the
	// first key, so it will run again. So there will be a total of four
	// callbacks called.
	changes.clusterInstWorkGo <- true
	changes.clusterInstWorkGo <- true
	changes.clusterInstWorkGo <- true
	changes.clusterInstWorkGo <- true
	<-changes.clusterInstWorkStarted
	<-changes.clusterInstWorkDone
	<-changes.clusterInstWorkDone
	<-changes.clusterInstWorkDone
	<-changes.clusterInstWorkDone
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, 0, len(mgr.workers))
	assert.Equal(t, 0, len(changes.curClusterInst))
	assert.Equal(t, 4, changes.clusterInstChanges)
}

func checkCurAppInst(t *testing.T, c *changes, expkey *edgeproto.AppInstKey, expold *edgeproto.AppInst) {
	c.mux.Lock()
	old, found := c.curAppInst[*expkey]
	c.mux.Unlock()
	assert.True(t, found, "key should be in current workers %v", expkey)
	assert.Equal(t, expold, old)
}

func checkCurClusterInst(t *testing.T, c *changes, expkey *edgeproto.ClusterInstKey, expold *edgeproto.ClusterInst) {
	old, found := c.curClusterInst[*expkey]
	assert.True(t, found, "key should be in current workers %v", expkey)
	assert.Equal(t, expold, old)
}

type changes struct {
	appInstChanges         int
	clusterInstChanges     int
	appInstWorkStarted     chan bool
	clusterInstWorkStarted chan bool
	appInstWorkDone        chan bool
	clusterInstWorkDone    chan bool
	appInstWorkGo          chan bool
	clusterInstWorkGo      chan bool
	curAppInst             map[edgeproto.AppInstKey]*edgeproto.AppInst
	curClusterInst         map[edgeproto.ClusterInstKey]*edgeproto.ClusterInst
	mux                    sync.Mutex
}

func newChanges() *changes {
	c := changes{}
	c.appInstWorkDone = make(chan bool, 5)
	c.clusterInstWorkDone = make(chan bool, 5)
	c.appInstWorkStarted = make(chan bool, 5)
	c.clusterInstWorkStarted = make(chan bool, 5)
	c.appInstWorkGo = make(chan bool, 5)
	c.clusterInstWorkGo = make(chan bool, 5)
	c.curAppInst = make(map[edgeproto.AppInstKey]*edgeproto.AppInst)
	c.curClusterInst = make(map[edgeproto.ClusterInstKey]*edgeproto.ClusterInst)
	return &c
}

func (c *changes) appInstChanged(key *edgeproto.AppInstKey, old *edgeproto.AppInst) {
	c.mux.Lock()
	c.curAppInst[*key] = old
	c.mux.Unlock()

	c.appInstWorkStarted <- true
	<-c.appInstWorkGo

	c.mux.Lock()
	c.appInstChanges++
	delete(c.curAppInst, *key)
	c.mux.Unlock()
	c.appInstWorkDone <- true
}

func (c *changes) clusterInstChanged(key *edgeproto.ClusterInstKey, old *edgeproto.ClusterInst) {
	c.mux.Lock()
	c.curClusterInst[*key] = old
	c.mux.Unlock()

	c.clusterInstWorkStarted <- true
	<-c.clusterInstWorkGo

	c.mux.Lock()
	c.clusterInstChanges++
	delete(c.curClusterInst, *key)
	c.mux.Unlock()
	c.clusterInstWorkDone <- true
}
