package main

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/notify"
)

//NewNotifyHandler instantiates new notify handler
func InitNotify(client *notify.Client, cd *crmutil.ControllerData) {
	client.RegisterRecvFlavorCache(&cd.FlavorCache)
	client.RegisterRecvClusterFlavorCache(&cd.ClusterFlavorCache)
	client.RegisterRecvAppCache(&cd.AppCache)
	client.RegisterRecvAppInstCache(&cd.AppInstCache)
	client.RegisterRecvCloudletCache(&cd.CloudletCache)
	client.RegisterRecvClusterInstCache(&cd.ClusterInstCache)

	client.RegisterSendCloudletInfoCache(&cd.CloudletInfoCache)
	client.RegisterSendAppInstInfoCache(&cd.AppInstInfoCache)
	client.RegisterSendClusterInstInfoCache(&cd.ClusterInstInfoCache)
	client.RegisterSendNodeCache(&cd.NodeCache)
}
