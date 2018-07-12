package main

import (
	"github.com/mobiledgex/edge-cloud/notify"
)

func NewNotifyHandler() *notify.DefaultHandler {
	handler := notify.DefaultHandler{}

	handler.SendAppInst = &appInstApi
	handler.SendCloudlet = &cloudletApi
	handler.RecvAppInstInfo = &appInstInfoApi
	handler.RecvCloudletInfo = &cloudletInfoApi
	return &handler
}
