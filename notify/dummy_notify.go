// Dummy Sender/Receiver for unit testing.
// These are exported because the notify package is meant to be included
// in other processes, so to include these structs in other package's
// unit tests, these test structures must be exported.
package notify

import (
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type DummyHandler struct {
	AppCache             edgeproto.AppCache
	AppInstCache         edgeproto.AppInstCache
	CloudletCache        edgeproto.CloudletCache
	FlavorCache          edgeproto.FlavorCache
	ClusterInstCache     edgeproto.ClusterInstCache
	AppInstInfoCache     edgeproto.AppInstInfoCache
	ClusterInstInfoCache edgeproto.ClusterInstInfoCache
	CloudletInfoCache    edgeproto.CloudletInfoCache
	AlertCache           edgeproto.AlertCache
	NodeCache            edgeproto.NodeCache
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
	edgeproto.InitClusterInstCache(&h.ClusterInstCache)
	edgeproto.InitAlertCache(&h.AlertCache)
	edgeproto.InitNodeCache(&h.NodeCache)
	return h
}

func (s *DummyHandler) RegisterServer(mgr *ServerMgr) {
	mgr.RegisterSendAppCache(&s.AppCache)
	mgr.RegisterSendAppInstCache(&s.AppInstCache)
	mgr.RegisterSendCloudletCache(&s.CloudletCache)
	mgr.RegisterSendCloudletInfoCache(&s.CloudletInfoCache)
	mgr.RegisterSendFlavorCache(&s.FlavorCache)
	mgr.RegisterSendClusterInstCache(&s.ClusterInstCache)
	mgr.RegisterSendAlertCache(&s.AlertCache)

	mgr.RegisterRecvAppInstInfoCache(&s.AppInstInfoCache)
	mgr.RegisterRecvClusterInstInfoCache(&s.ClusterInstInfoCache)
	mgr.RegisterRecvCloudletInfoCache(&s.CloudletInfoCache)
	mgr.RegisterRecvAlertCache(&s.AlertCache)
	mgr.RegisterRecvNodeCache(&s.NodeCache)
}

func (s *DummyHandler) RegisterCRMClient(cl *Client) {
	cl.SetFilterByCloudletKey()
	cl.RegisterSendAppInstInfoCache(&s.AppInstInfoCache)
	cl.RegisterSendClusterInstInfoCache(&s.ClusterInstInfoCache)
	cl.RegisterSendCloudletInfoCache(&s.CloudletInfoCache)
	cl.RegisterSendAlertCache(&s.AlertCache)
	cl.RegisterSendNodeCache(&s.NodeCache)

	cl.RegisterRecvAppCache(&s.AppCache)
	cl.RegisterRecvAppInstCache(&s.AppInstCache)
	cl.RegisterRecvCloudletCache(&s.CloudletCache)
	cl.RegisterRecvFlavorCache(&s.FlavorCache)
	cl.RegisterRecvClusterInstCache(&s.ClusterInstCache)
}

func (s *DummyHandler) RegisterDMEClient(cl *Client) {
	cl.RegisterRecvAppCache(&s.AppCache)
	cl.RegisterRecvAppInstCache(&s.AppInstCache)
}

type CacheType int

const (
	AppType     CacheType = iota
	AppInstType           = iota
	CloudletType
	FlavorType
	ClusterInstType
	AppInstInfoType
	ClusterInstInfoType
	CloudletInfoType
	AlertType
	NodeType
)

type WaitForCache interface {
	GetCount() int
}

func (s *DummyHandler) WaitFor(typ CacheType, count int) {
	cache := s.GetCache(typ)
	WaitFor(cache, count)
}

func (s *DummyHandler) GetCache(typ CacheType) WaitForCache {
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
	case ClusterInstType:
		cache = &s.ClusterInstCache
	case AppInstInfoType:
		cache = &s.AppInstInfoCache
	case ClusterInstInfoType:
		cache = &s.ClusterInstInfoCache
	case CloudletInfoType:
		cache = &s.CloudletInfoCache
	case AlertType:
		cache = &s.AlertCache
	case NodeType:
		cache = &s.NodeCache
	}
	return cache
}

func (c CacheType) String() string {
	switch c {
	case AppType:
		return "AppCache"
	case AppInstType:
		return "AppInstCache"
	case CloudletType:
		return "CloudletCache"
	case FlavorType:
		return "FlavorCache"
	case ClusterInstType:
		return "ClusterInstCache"
	case AppInstInfoType:
		return "AppInstCache"
	case ClusterInstInfoType:
		return "ClusterInstCache"
	case CloudletInfoType:
		return "CloudletInfoCache"
	case AlertType:
		return "AlertCache"
	case NodeType:
		return "NodeCache"
	}
	return "unknown cache type"
}

func WaitFor(cache WaitForCache, count int) {
	if cache == nil {
		return
	}
	for i := 0; i < 50; i++ {
		if cache.GetCount() == count {
			break
		}
		time.Sleep(20 * time.Millisecond)
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

func (s *DummyHandler) WaitForClusterInsts(count int) {
	WaitFor(&s.ClusterInstCache, count)
}

func (s *DummyHandler) WaitForAlerts(count int) {
	s.WaitFor(AlertType, count)
}

func (s *DummyHandler) WaitForCloudletState(key *edgeproto.CloudletKey, state edgeproto.CloudletState, version string) error {
	lastState := edgeproto.CloudletState_CLOUDLET_STATE_UNKNOWN
	for i := 0; i < 100; i++ {
		cloudletInfo := edgeproto.CloudletInfo{}
		if s.CloudletInfoCache.Get(key, &cloudletInfo) {
			if cloudletInfo.State == state {
				if cloudletInfo.Version == version {
					return nil
				}
				return fmt.Errorf("invalid cloudletInfo version: %s, should be %s",
					cloudletInfo.Version,
					version)
			}
			lastState = cloudletInfo.State
		}
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("Unable to get desired cloudletInfo state, actual state %s, desired state %s", lastState, state)
}

func (s *DummyHandler) GetCloudletDetails(key *edgeproto.CloudletKey) (string, int64, error) {
	for _, obj := range s.NodeCache.Objs {
		if obj.Key.NodeType != edgeproto.NodeType_NODE_CRM {
			continue
		}
		if obj.Key.CloudletKey != *key {
			continue
		}
		return obj.ImageVersion, obj.NotifyId, nil
	}
	return "", -1, fmt.Errorf("Unable to find cloudlet in node list")
}

func (s *Client) WaitForConnect(connect uint64) {
	for i := 0; i < 10; i++ {
		if s.sendrecv.stats.Connects == connect {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (mgr *ServerMgr) WaitServerCount(count int) {
	for i := 0; i < 50; i++ {
		mgr.mux.Lock()
		cnt := len(mgr.table)
		mgr.mux.Unlock()
		if cnt == count {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
}
