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
	AppCache             edgeproto.AppCache
	AppInstCache         edgeproto.AppInstCache
	CloudletCache        edgeproto.CloudletCache
	FlavorCache          edgeproto.FlavorCache
	ClusterFlavorCache   edgeproto.ClusterFlavorCache
	ClusterInstCache     edgeproto.ClusterInstCache
	AppInstInfoCache     edgeproto.AppInstInfoCache
	ClusterInstInfoCache edgeproto.ClusterInstInfoCache
	CloudletInfoCache    edgeproto.CloudletInfoCache
}

func NewDummyHandler() *DummyHandler {
	h := &DummyHandler{}
	edgeproto.InitAppCache(&h.AppCache)
	edgeproto.InitAppInstCache(&h.AppInstCache)
	edgeproto.InitCloudletCache(&h.CloudletCache)
	edgeproto.InitAppInstInfoCache(&h.AppInstInfoCache)
	edgeproto.InitClusterInstInfoCache(&h.ClusterInstInfoCache)
	edgeproto.InitCloudletInfoCache(&h.CloudletInfoCache)
	edgeproto.InitFlavorCache(&h.FlavorCache)
	edgeproto.InitClusterFlavorCache(&h.ClusterFlavorCache)
	edgeproto.InitClusterInstCache(&h.ClusterInstCache)
	h.DefaultHandler.SendApp = &h.AppCache
	h.DefaultHandler.RecvApp = &h.AppCache
	h.DefaultHandler.SendAppInst = &h.AppInstCache
	h.DefaultHandler.RecvAppInst = &h.AppInstCache
	h.DefaultHandler.SendCloudlet = &h.CloudletCache
	h.DefaultHandler.RecvCloudlet = &h.CloudletCache
	h.DefaultHandler.SendAppInstInfo = &h.AppInstInfoCache
	h.DefaultHandler.RecvAppInstInfo = &h.AppInstInfoCache
	h.DefaultHandler.SendClusterInstInfo = &h.ClusterInstInfoCache
	h.DefaultHandler.RecvClusterInstInfo = &h.ClusterInstInfoCache
	h.DefaultHandler.SendCloudletInfo = &h.CloudletInfoCache
	h.DefaultHandler.RecvCloudletInfo = &h.CloudletInfoCache
	h.DefaultHandler.SendFlavor = &h.FlavorCache
	h.DefaultHandler.RecvFlavor = &h.FlavorCache
	h.DefaultHandler.SendClusterFlavor = &h.ClusterFlavorCache
	h.DefaultHandler.RecvClusterFlavor = &h.ClusterFlavorCache
	h.DefaultHandler.SendClusterInst = &h.ClusterInstCache
	h.DefaultHandler.RecvClusterInst = &h.ClusterInstCache
	return h
}

func (s *DummyHandler) SetServerCb(mgr *ServerMgr) {
	s.AppCache.SetNotifyCb(mgr.UpdateApp)
	s.AppInstCache.SetNotifyCb(mgr.UpdateAppInst)
	s.CloudletCache.SetNotifyCb(mgr.UpdateCloudlet)
	s.FlavorCache.SetNotifyCb(mgr.UpdateFlavor)
	s.ClusterFlavorCache.SetNotifyCb(mgr.UpdateClusterFlavor)
	s.ClusterInstCache.SetNotifyCb(mgr.UpdateClusterInst)
}

func (s *DummyHandler) SetClientCb(cl *Client) {
	s.AppInstInfoCache.SetNotifyCb(cl.UpdateAppInstInfo)
	s.ClusterInstInfoCache.SetNotifyCb(cl.UpdateClusterInstInfo)
	s.CloudletInfoCache.SetNotifyCb(cl.UpdateCloudletInfo)
}

type CacheType int

const (
	AppType     CacheType = iota
	AppInstType           = iota
	CloudletType
	FlavorType
	ClusterFlavorType
	ClusterInstType
	AppInstInfoType
	ClusterInstInfoType
	CloudletInfoType
)

type WaitForCache interface {
	GetCount() int
}

func (s *DummyHandler) WaitFor(typ CacheType, count int) {
	var cache WaitForCache
	switch typ {
	case AppType:
		cache = &s.AppCache
	case AppInstType:
		cache = &s.AppInstCache
	case CloudletType:
		cache = &s.CloudletCache
	case FlavorType:
		cache = &s.FlavorCache
	case ClusterFlavorType:
		cache = &s.ClusterFlavorCache
	case ClusterInstType:
		cache = &s.ClusterInstCache
	case AppInstInfoType:
		cache = &s.AppInstInfoCache
	case ClusterInstInfoType:
		cache = &s.ClusterInstInfoCache
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

func (s *DummyHandler) WaitForClusterInstInfo(count int) {
	WaitFor(&s.ClusterInstInfoCache, count)
}

func (s *DummyHandler) WaitForCloudletInfo(count int) {
	WaitFor(&s.CloudletInfoCache, count)
}

func (s *DummyHandler) WaitForApps(count int) {
	WaitFor(&s.AppCache, count)
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

func (s *DummyHandler) WaitForClusterFlavors(count int) {
	WaitFor(&s.ClusterFlavorCache, count)
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
