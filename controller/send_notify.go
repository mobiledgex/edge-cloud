package main

import "github.com/mobiledgex/edge-cloud/proto"

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

func (s *ControllerNotifier) GetAllAppInstKeys(keys map[proto.AppInstKey]bool) {
	s.AppInstApi.GetAllKeys(keys)
}

func (s *ControllerNotifier) GetAllCloudletKeys(keys map[proto.CloudletKey]bool) {
	s.CloudletApi.GetAllKeys(keys)
}
