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
	FlavorCache       edgeproto.FlavorCache
	ClusterInstCache  edgeproto.ClusterInstCache
	AppInstInfoCache  edgeproto.AppInstInfoCache
	CloudletInfoCache edgeproto.CloudletInfoCache
}

func NewDummyHandler() *DummyHandler {
	h := &DummyHandler{}
	edgeproto.InitAppInstCache(&h.AppInstCache)
	edgeproto.InitCloudletCache(&h.CloudletCache)
	edgeproto.InitAppInstInfoCache(&h.AppInstInfoCache)
	edgeproto.InitCloudletInfoCache(&h.CloudletInfoCache)
	edgeproto.InitFlavorCache(&h.FlavorCache)
	edgeproto.InitClusterInstCache(&h.ClusterInstCache)
	h.DefaultHandler.SendAppInst = &h.AppInstCache
	h.DefaultHandler.RecvAppInst = &h.AppInstCache
	h.DefaultHandler.SendCloudlet = &h.CloudletCache
	h.DefaultHandler.RecvCloudlet = &h.CloudletCache
	h.DefaultHandler.SendAppInstInfo = &h.AppInstInfoCache
	h.DefaultHandler.RecvAppInstInfo = &h.AppInstInfoCache
	h.DefaultHandler.SendCloudletInfo = &h.CloudletInfoCache
	h.DefaultHandler.RecvCloudletInfo = &h.CloudletInfoCache
	h.DefaultHandler.SendFlavor = &h.FlavorCache
	h.DefaultHandler.RecvFlavor = &h.FlavorCache
	h.DefaultHandler.SendClusterInst = &h.ClusterInstCache
	h.DefaultHandler.RecvClusterInst = &h.ClusterInstCache
	return h
}

func (s *DummyHandler) SetServerCb(mgr *ServerMgr) {
	s.AppInstCache.SetNotifyCb(mgr.UpdateAppInst)
	s.CloudletCache.SetNotifyCb(mgr.UpdateCloudlet)
	s.FlavorCache.SetNotifyCb(mgr.UpdateFlavor)
	s.ClusterInstCache.SetNotifyCb(mgr.UpdateClusterInst)
}

func (s *DummyHandler) SetClientCb(cl *Client) {
	s.AppInstInfoCache.SetNotifyCb(cl.UpdateAppInstInfo)
	s.CloudletInfoCache.SetNotifyCb(cl.UpdateCloudletInfo)
}

type CacheType int

const (
	AppInstType CacheType = iota
	CloudletType
	FlavorType
	ClusterInstType
	AppInstInfoType
	CloudletInfoType
)

type WaitForCache interface {
	GetCount() int
}

func (s *DummyHandler) WaitFor(typ CacheType, count int) {
	var cache WaitForCache
	switch typ {
	case AppInstType:
		cache = &s.AppInstCache
	case CloudletType:
		cache = &s.CloudletCache
	case FlavorType:
		cache = &s.FlavorCache
	case ClusterInstType:
		cache = &s.ClusterInstCache
	case AppInstInfoType:
		cache = &s.AppInstInfoCache
	case CloudletInfoType:
		cache = &s.CloudletInfoCache
	}
	WaitFor(cache, count)
}

func WaitFor(cache WaitForCache, count int) {
	if cache == nil {
		return
	}
	for i := 0; i < 10; i++ {
		if cache.GetCount() == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *DummyHandler) WaitForAppInstInfo(count int) {
	WaitFor(&s.AppInstInfoCache, count)
}

func (s *DummyHandler) WaitForCloudletInfo(count int) {
	WaitFor(&s.CloudletInfoCache, count)
}

func (s *DummyHandler) WaitForAppInsts(count int) {
	WaitFor(&s.AppInstCache, count)
}

func (s *DummyHandler) WaitForCloudlets(count int) {
	WaitFor(&s.CloudletCache, count)
}

func (s *DummyHandler) WaitForFlavors(count int) {
	WaitFor(&s.FlavorCache, count)
}

func (s *DummyHandler) WaitForClusterInsts(count int) {
	WaitFor(&s.ClusterInstCache, count)
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
