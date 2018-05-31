// Dummy Sender/Receiver for unit testing.
// These are exported because the notify package is meant to be included
// in other processes, so to include these structs in other package's
// unit tests, these test structures must be exported.
package notify

import (
	"os"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
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

func (s *DummySendHandler) GetAllAppInstKeys(keys map[edgeproto.AppInstKey]struct{}) {
	for key, _ := range s.appInsts {
		keys[key] = struct{}{}
	}
}

func (s *DummySendHandler) GetAppInst(key *edgeproto.AppInstKey, buf *edgeproto.AppInst) bool {
	obj, found := s.appInsts[*key]
	if found {
		*buf = obj
	}
	return found
}

func (s *DummySendHandler) GetAllCloudletKeys(keys map[edgeproto.CloudletKey]struct{}) {
	for key, _ := range s.cloudlets {
		keys[key] = struct{}{}
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
	AppInsts           map[edgeproto.AppInstKey]edgeproto.AppInst
	Cloudlets          map[edgeproto.CloudletKey]edgeproto.Cloudlet
	NumAppInstUpdates  int
	NumCloudletUpdates int
	NumUpdates         int
	Recv               *NotifyReceiver
}

func NewDummyRecvHandler() *DummyRecvHandler {
	handler := &DummyRecvHandler{}
	handler.AppInsts = make(map[edgeproto.AppInstKey]edgeproto.AppInst)
	handler.Cloudlets = make(map[edgeproto.CloudletKey]edgeproto.Cloudlet)
	return handler
}

func (s *DummyRecvHandler) Start(network, addr string) {
	if network == "unix" {
		os.Remove(addr)
	}
	s.Recv = NewNotifyReceiver(network, addr, s)
	go s.Recv.Run()
}

func (s *DummyRecvHandler) Stop() {
	s.Recv.Stop()
}

func (s *DummyRecvHandler) HandleSendAllDone(maps *NotifySendAllMaps) {
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

func (s *DummyRecvHandler) HandleNotice(notice *edgeproto.Notice) error {
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

func (s *DummyRecvHandler) WaitForAppInsts(count int) {
	for i := 0; i < 10; i++ {
		if len(s.AppInsts) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *DummyRecvHandler) WaitForCloudlets(count int) {
	for i := 0; i < 10; i++ {
		if len(s.Cloudlets) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *DummyRecvHandler) WaitForConnect(connect uint64) {
	for i := 0; i < 10; i++ {
		if s.Recv.GetConnnectionId() == connect {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}
