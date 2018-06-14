package main

import "github.com/mobiledgex/edge-cloud/edgeproto"

type ControllerNotifier struct {
	*AppInstApi
	*CloudletApi
}

func NewControllerNotifier(appInstApi *AppInstApi, cloudletApi *CloudletApi) *ControllerNotifier {
	n := ControllerNotifier{}
	n.AppInstApi = appInstApi
	n.CloudletApi = cloudletApi
	return &n
}

func (s *ControllerNotifier) GetAllAppInstKeys(keys map[edgeproto.AppInstKey]struct{}) {
	s.AppInstApi.GetAllKeys(keys)
}

func (s *ControllerNotifier) GetAllCloudletKeys(keys map[edgeproto.CloudletKey]struct{}) {
	s.CloudletApi.GetAllKeys(keys)
}
