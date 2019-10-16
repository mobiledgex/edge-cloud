package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
)

var sendMetric *notify.MetricSend
var sendAlert *notify.AlertSend

// NewNotifyHandler instantiates new notify handler
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
	client.RegisterSendAlertCache(&cd.AlertCache)
}

func initSrvNotify(notifyServer *notify.ServerMgr) {
	notifyServer.RegisterSendAppCache(&controllerData.AppCache)
	notifyServer.RegisterSendClusterInstCache(&controllerData.ClusterInstCache)
	notifyServer.RegisterSendAppInstCache(&controllerData.AppInstCache)
	notifyServer.RegisterRecv(notify.NewMetricRecvMany(&CrmMetricsReceiver{}))
	notifyServer.RegisterRecvAlertCache(&controllerData.AlertCache)
}

type CrmMetricsReceiver struct{}

// forward to controller
func (r *CrmMetricsReceiver) Recv(ctx context.Context, metric *edgeproto.Metric) {
	sendMetric.Update(ctx, metric)
}
