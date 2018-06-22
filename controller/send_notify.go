package main

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

type ControllerNotifier struct {
	*AppInstApi
	*CloudletApi
	appInstInfo  AppInstInfo
	cloudletInfo CloudletInfo
}

func NewControllerNotifier(appInstApi *AppInstApi, cloudletApi *CloudletApi) *ControllerNotifier {
	n := ControllerNotifier{}
	n.AppInstApi = appInstApi
	n.CloudletApi = cloudletApi
	n.appInstInfo.Init()
	n.cloudletInfo.Init()
	return &n
}

func (s *ControllerNotifier) GetAllAppInstKeys(keys map[edgeproto.AppInstKey]struct{}) {
	s.AppInstApi.GetAllKeys(keys)
}

func (s *ControllerNotifier) GetAllCloudletKeys(keys map[edgeproto.CloudletKey]struct{}) {
	s.CloudletApi.GetAllKeys(keys)
}

func (s *ControllerNotifier) HandleNotice(notice *edgeproto.NoticeRequest) {
	a := notice.GetAppInstInfo()
	if a != nil {
		s.appInstInfo.Process(a)
	}
	c := notice.GetCloudletInfo()
	if c != nil {
		s.cloudletInfo.Process(c)
	}
}

type AppInstInfo struct {
	table map[edgeproto.AppInstKey]*edgeproto.AppInstInfo
	mux   util.Mutex
}

func (s *AppInstInfo) Init() {
	s.table = make(map[edgeproto.AppInstKey]*edgeproto.AppInstInfo)
}

func (s *AppInstInfo) Process(info *edgeproto.AppInstInfo) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.table[info.Key] = info
}

type CloudletInfo struct {
	table map[edgeproto.CloudletKey]*edgeproto.CloudletInfo
	mux   util.Mutex
}

func (s *CloudletInfo) Init() {
	s.table = make(map[edgeproto.CloudletKey]*edgeproto.CloudletInfo)
}

func (s *CloudletInfo) Process(info *edgeproto.CloudletInfo) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.table[info.Key] = info
}
