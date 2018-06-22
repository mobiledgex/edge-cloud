// Dummy Sender/Receiver for unit testing.
// These are exported because the notify package is meant to be included
// in other processes, so to include these structs in other package's
// unit tests, these test structures must be exported.
package notify

import (
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type DummyServerHandler struct {
	AppInsts      map[edgeproto.AppInstKey]edgeproto.AppInst
	AppInstsInfo  map[edgeproto.AppInstKey]edgeproto.AppInstInfo
	Cloudlets     map[edgeproto.CloudletKey]edgeproto.Cloudlet
	CloudletsInfo map[edgeproto.CloudletKey]edgeproto.CloudletInfo
}

func NewDummyServerHandler() *DummyServerHandler {
	handler := &DummyServerHandler{}
	handler.AppInsts = make(map[edgeproto.AppInstKey]edgeproto.AppInst)
	handler.AppInstsInfo = make(map[edgeproto.AppInstKey]edgeproto.AppInstInfo)
	handler.Cloudlets = make(map[edgeproto.CloudletKey]edgeproto.Cloudlet)
	handler.CloudletsInfo = make(map[edgeproto.CloudletKey]edgeproto.CloudletInfo)
	return handler
}

func (s *DummyServerHandler) GetAllAppInstKeys(keys map[edgeproto.AppInstKey]struct{}) {
	for key, _ := range s.AppInsts {
		keys[key] = struct{}{}
	}
}

func (s *DummyServerHandler) GetAppInst(key *edgeproto.AppInstKey, buf *edgeproto.AppInst) bool {
	obj, found := s.AppInsts[*key]
	if found {
		*buf = obj
	}
	return found
}

func (s *DummyServerHandler) GetAllCloudletKeys(keys map[edgeproto.CloudletKey]struct{}) {
	for key, _ := range s.Cloudlets {
		keys[key] = struct{}{}
	}
}

func (s *DummyServerHandler) GetCloudlet(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool {
	obj, found := s.Cloudlets[*key]
	if found {
		*buf = obj
	}
	return found
}

func (s *DummyServerHandler) HandleNotice(notice *edgeproto.NoticeRequest) {
	a := notice.GetAppInstInfo()
	if a != nil {
		s.AppInstsInfo[a.Key] = *a
	}
	c := notice.GetCloudletInfo()
	if c != nil {
		s.CloudletsInfo[c.Key] = *c
	}
}

func (s *DummyServerHandler) CreateAppInst(in *edgeproto.AppInst) {
	s.AppInsts[in.Key] = *in
	UpdateAppInst(&in.Key)
}

func (s *DummyServerHandler) DeleteAppInst(in *edgeproto.AppInst) {
	delete(s.AppInsts, in.Key)
	UpdateAppInst(&in.Key)
}

func (s *DummyServerHandler) CreateCloudlet(in *edgeproto.Cloudlet) {
	s.Cloudlets[in.Key] = *in
	UpdateCloudlet(&in.Key)
}

func (s *DummyServerHandler) DeleteCloudlet(in *edgeproto.Cloudlet) {
	delete(s.Cloudlets, in.Key)
	UpdateCloudlet(&in.Key)
}

func (s *DummyServerHandler) WaitForAppInstInfo(count int) {
	for i := 0; i < 10; i++ {
		if len(s.AppInstsInfo) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *DummyServerHandler) WaitForCloudletInfo(count int) {
	for i := 0; i < 10; i++ {
		if len(s.CloudletsInfo) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

type DummyClientHandler struct {
	AppInsts           map[edgeproto.AppInstKey]edgeproto.AppInst
	Cloudlets          map[edgeproto.CloudletKey]edgeproto.Cloudlet
	NumAppInstUpdates  int
	NumCloudletUpdates int
	NumUpdates         int
}

func NewDummyClientHandler() *DummyClientHandler {
	handler := &DummyClientHandler{}
	handler.AppInsts = make(map[edgeproto.AppInstKey]edgeproto.AppInst)
	handler.Cloudlets = make(map[edgeproto.CloudletKey]edgeproto.Cloudlet)
	return handler
}

func (s *DummyClientHandler) HandleSendAllDone(maps *AllMaps) {
	for key, _ := range s.AppInsts {
		if _, ok := maps.AppInsts[key]; !ok {
			delete(s.AppInsts, key)
		}
	}
	for key, _ := range s.Cloudlets {
		if _, ok := maps.Cloudlets[key]; !ok {
			delete(s.Cloudlets, key)
		}
	}
}

func (s *DummyClientHandler) HandleNotice(notice *edgeproto.NoticeReply) error {
	appInst := notice.GetAppInst()
	if appInst != nil {
		if notice.Action == edgeproto.NoticeAction_UPDATE {
			s.AppInsts[appInst.Key] = *appInst
		} else if notice.Action == edgeproto.NoticeAction_DELETE {
			delete(s.AppInsts, appInst.Key)
		}
		s.NumAppInstUpdates++
	}
	cloudlet := notice.GetCloudlet()
	if cloudlet != nil {
		if notice.Action == edgeproto.NoticeAction_UPDATE {
			s.Cloudlets[cloudlet.Key] = *cloudlet
		} else if notice.Action == edgeproto.NoticeAction_DELETE {
			delete(s.Cloudlets, cloudlet.Key)
		}
		s.NumCloudletUpdates++
	}
	s.NumUpdates++
	return nil
}

func (s *DummyClientHandler) WaitForAppInsts(count int) {
	for i := 0; i < 10; i++ {
		if len(s.AppInsts) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *DummyClientHandler) WaitForCloudlets(count int) {
	for i := 0; i < 10; i++ {
		if len(s.Cloudlets) == count {
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

func WaitServerCount(count int) {
	for i := 0; i < 10; i++ {
		serverMgr.mux.Lock()
		cnt := len(serverMgr.table)
		serverMgr.mux.Unlock()
		if cnt == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}
