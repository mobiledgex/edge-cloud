package main

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
)

func NewNotifyHandler() *notify.DefaultHandler {
	handler := notify.DefaultHandler{}
	appInstInfos := edgeproto.NewAppInstInfoCache()
	cloudletInfos := edgeproto.NewCloudletInfoCache()

	handler.SendAppInst = &appInstApi
	handler.SendCloudlet = &cloudletApi
	handler.RecvAppInstInfo = appInstInfos
	handler.RecvCloudletInfo = cloudletInfos
	return &handler
}
