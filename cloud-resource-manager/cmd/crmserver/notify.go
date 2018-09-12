package main

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/notify"
)

//NewNotifyHandler instantiates new notify handler
func NewNotifyHandler(cd *crmutil.ControllerData) *notify.DefaultHandler {
	h := notify.DefaultHandler{}
	h.SendAppInst = &cd.AppInstCache
	h.RecvAppInst = &cd.AppInstCache
	h.SendCloudlet = &cd.CloudletCache
	h.RecvCloudlet = &cd.CloudletCache
	h.SendAppInstInfo = &cd.AppInstInfoCache
	h.RecvAppInstInfo = &cd.AppInstInfoCache
	h.SendClusterInstInfo = &cd.ClusterInstInfoCache
	h.RecvClusterInstInfo = &cd.ClusterInstInfoCache
	h.SendCloudletInfo = &cd.CloudletInfoCache
	h.RecvCloudletInfo = &cd.CloudletInfoCache
	h.RecvFlavor = &cd.FlavorCache
	h.RecvClusterFlavor = &cd.ClusterFlavorCache // todo: auto-gen this
	// or make it easier to add new caches/objects
	h.RecvClusterInst = &cd.ClusterInstCache
	h.SendNode = &cd.NodeCache
	return &h
}
