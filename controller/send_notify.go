package main

import (
	"github.com/mobiledgex/edge-cloud/notify"
)

func NewNotifyHandler() *notify.DefaultHandler {
	handler := notify.DefaultHandler{}

	handler.SendFlavor = &flavorApi
	handler.SendClusterInst = &clusterInstApi
	handler.SendAppInst = &appInstApi
	handler.SendCloudlet = &cloudletApi
	handler.RecvAppInstInfo = &appInstInfoApi
	handler.RecvClusterInstInfo = &clusterInstInfoApi
	handler.RecvCloudletInfo = &cloudletInfoApi
	return &handler
}
