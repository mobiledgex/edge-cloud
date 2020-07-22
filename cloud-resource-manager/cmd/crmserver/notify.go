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
	client.RegisterRecvSettingsCache(&cd.SettingsCache)
	client.RegisterRecvFlavorCache(&cd.FlavorCache)
	client.RegisterRecvAppCache(&cd.AppCache)
	client.RegisterRecvAppInstCache(&cd.AppInstCache)
	client.RegisterRecvCloudletCache(&cd.CloudletCache)
	client.RegisterRecvClusterInstCache(&cd.ClusterInstCache)
	client.RegisterRecv(notify.NewExecRequestRecv(cd.ExecReqHandler))
	client.RegisterRecvResTagTableCache(&cd.ResTagTableCache)
	client.RegisterSendCloudletInfoCache(&cd.CloudletInfoCache)
	client.RegisterSendAppInstInfoCache(&cd.AppInstInfoCache)
	client.RegisterSendClusterInstInfoCache(&cd.ClusterInstInfoCache)
	client.RegisterSend(cd.ExecReqSend)
	sendMetric = notify.NewMetricSend()
	client.RegisterSend(sendMetric)
	client.RegisterSendAlertCache(&cd.AlertCache)
	client.RegisterRecvPrivacyPolicyCache(&cd.PrivacyPolicyCache)
	client.RegisterRecvAutoProvPolicyCache(&cd.AutoProvPolicyCache)
	client.RegisterSendAllRecv(cd)
	nodeMgr.RegisterClient(client)
}

func initSrvNotify(notifyServer *notify.ServerMgr) {
	notifyServer.RegisterSendSettingsCache(&controllerData.SettingsCache)
	notifyServer.RegisterSendCloudletCache(&controllerData.CloudletCache)
	notifyServer.RegisterSendAutoProvPolicyCache(&controllerData.AutoProvPolicyCache)
	notifyServer.RegisterSendAppCache(&controllerData.AppCache)
	notifyServer.RegisterSendClusterInstCache(&controllerData.ClusterInstCache)
	notifyServer.RegisterSendAppInstCache(&controllerData.AppInstCache)

	notifyServer.RegisterRecv(notify.NewMetricRecvMany(&CrmMetricsReceiver{}))
	notifyServer.RegisterRecvAlertCache(&controllerData.AlertCache)
	// Dummy CloudletInfoCache receiver to avoid sending
	// cloudletInfo updates to controller from Shepherd
	var DummyCloudletInfoRecvCache edgeproto.CloudletInfoCache
	edgeproto.InitCloudletInfoCache(&DummyCloudletInfoRecvCache)
	notifyServer.RegisterRecvCloudletInfoCache(&DummyCloudletInfoRecvCache)
	nodeMgr.RegisterServer(notifyServer)
}

type CrmMetricsReceiver struct{}

// forward to controller
func (r *CrmMetricsReceiver) RecvMetric(ctx context.Context, metric *edgeproto.Metric) {
	sendMetric.Update(ctx, metric)
}
