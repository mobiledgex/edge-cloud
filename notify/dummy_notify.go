// Dummy Sender/Receiver for unit testing.
// These are exported because the notify package is meant to be included
// in other processes, so to include these structs in other package's
// unit tests, these test structures must be exported.
package notify

import (
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type DummyHandler struct {
	DefaultHandler
	AppInstCache      edgeproto.AppInstCache
	CloudletCache     edgeproto.CloudletCache
	AppInstInfoCache  edgeproto.AppInstInfoCache
	CloudletInfoCache edgeproto.CloudletInfoCache
}

func NewDummyHandler() *DummyHandler {
	h := &DummyHandler{}
	edgeproto.InitAppInstCache(&h.AppInstCache)
	edgeproto.InitCloudletCache(&h.CloudletCache)
	edgeproto.InitAppInstInfoCache(&h.AppInstInfoCache)
	edgeproto.InitCloudletInfoCache(&h.CloudletInfoCache)
	h.DefaultHandler.SendAppInst = &h.AppInstCache
	h.DefaultHandler.RecvAppInst = &h.AppInstCache
	h.DefaultHandler.SendCloudlet = &h.CloudletCache
	h.DefaultHandler.RecvCloudlet = &h.CloudletCache
	h.DefaultHandler.SendAppInstInfo = &h.AppInstInfoCache
	h.DefaultHandler.RecvAppInstInfo = &h.AppInstInfoCache
	h.DefaultHandler.SendCloudletInfo = &h.CloudletInfoCache
	h.DefaultHandler.RecvCloudletInfo = &h.CloudletInfoCache
	return h
}

func (s *DummyHandler) SetServerCb(mgr *ServerMgr) {
	s.AppInstCache.SetNotifyCb(mgr.UpdateAppInst)
	s.CloudletCache.SetNotifyCb(mgr.UpdateCloudlet)
}

func (s *DummyHandler) SetClientCb(cl *Client) {
	s.AppInstInfoCache.SetNotifyCb(cl.UpdateAppInstInfo)
	s.CloudletInfoCache.SetNotifyCb(cl.UpdateCloudletInfo)
}

func (s *DummyHandler) WaitForAppInstInfo(count int) {
	for i := 0; i < 10; i++ {
		if len(s.AppInstInfoCache.Objs) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *DummyHandler) WaitForCloudletInfo(count int) {
	for i := 0; i < 10; i++ {
		if len(s.CloudletInfoCache.Objs) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *DummyHandler) WaitForAppInsts(count int) {
	for i := 0; i < 10; i++ {
		if len(s.AppInstCache.Objs) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *DummyHandler) WaitForCloudlets(count int) {
	for i := 0; i < 10; i++ {
		if len(s.CloudletCache.Objs) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *Client) WaitForConnect(connect uint64) {
	for i := 0; i < 10; i++ {
		if s.stats.Connects == connect {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (mgr *ServerMgr) WaitServerCount(count int) {
	for i := 0; i < 10; i++ {
		mgr.mux.Lock()
		cnt := len(mgr.table)
		mgr.mux.Unlock()
		if cnt == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}
