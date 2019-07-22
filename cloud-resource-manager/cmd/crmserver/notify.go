package main

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
)

var sendMetric *notify.MetricSend

//NewNotifyHandler instantiates new notify handler
func InitClientNotify(client *notify.Client, cd *crmutil.ControllerData) {
	client.RegisterRecvFlavorCache(&cd.FlavorCache)
	client.RegisterRecvAppCache(&cd.AppCache)
	client.RegisterRecvAppInstCache(&cd.AppInstCache)
	client.RegisterRecvCloudletCache(&cd.CloudletCache)
	client.RegisterRecvClusterInstCache(&cd.ClusterInstCache)
	client.RegisterRecv(notify.NewExecRequestRecv(cd.ExecReqHandler))

	client.RegisterSendCloudletInfoCache(&cd.CloudletInfoCache)
	client.RegisterSendAppInstInfoCache(&cd.AppInstInfoCache)
	client.RegisterSendClusterInstInfoCache(&cd.ClusterInstInfoCache)
	client.RegisterSendNodeCache(&cd.NodeCache)
	client.RegisterSend(cd.ExecReqSend)
	sendMetric = notify.NewMetricSend()
	client.RegisterSend(sendMetric)
}

func initSrvNotify(notifyServer *notify.ServerMgr) {
	notifyServer.RegisterSendClusterInstCache(&controllerData.ClusterInstCache)
	notifyServer.RegisterSendAppInstCache(&controllerData.AppInstCache)
	notifyServer.RegisterRecv(notify.NewMetricRecvMany(&CrmMetricsReceiver{}))
}

type CrmMetricsReceiver struct{}

//just forward to crm
func (cmr *CrmMetricsReceiver) Recv(metric *edgeproto.Metric) {
	sendMetric.Update(metric)
}
