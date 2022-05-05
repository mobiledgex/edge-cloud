// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crmutil

import (
	"context"

	"github.com/edgexr/edge-cloud/cloudcommon/node"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/notify"
)

var sendMetric *notify.MetricSend
var sendAlert *notify.AlertSend

// NewNotifyHandler instantiates new notify handler
func InitClientNotify(client *notify.Client, nodeMgr *node.NodeMgr, cd *ControllerData) {
	client.RegisterRecvSettingsCache(&cd.SettingsCache)
	client.RegisterRecvFlavorCache(&cd.FlavorCache)
	client.RegisterRecvAppCache(&cd.AppCache)
	client.RegisterRecvAppInstCache(&cd.AppInstCache)
	client.RegisterRecvCloudletCache(cd.CloudletCache)
	client.RegisterRecvVMPoolCache(&cd.VMPoolCache)
	client.RegisterRecvClusterInstCache(&cd.ClusterInstCache)
	client.RegisterRecv(notify.NewExecRequestRecv(cd.ExecReqHandler))
	client.RegisterRecvResTagTableCache(&cd.ResTagTableCache)
	client.RegisterRecvGPUDriverCache(&cd.GPUDriverCache)
	client.RegisterRecvNetworkCache(&cd.NetworkCache)
	client.RegisterSendCloudletInfoCache(&cd.CloudletInfoCache)
	client.RegisterSendVMPoolInfoCache(&cd.VMPoolInfoCache)
	client.RegisterSendAppInstInfoCache(&cd.AppInstInfoCache)
	client.RegisterSendClusterInstInfoCache(&cd.ClusterInstInfoCache)
	client.RegisterSend(cd.ExecReqSend)
	sendMetric = notify.NewMetricSend()
	client.RegisterSend(sendMetric)
	client.RegisterSendAlertCache(&cd.AlertCache)
	client.RegisterRecvTrustPolicyCache(&cd.TrustPolicyCache)
	client.RegisterRecvTrustPolicyExceptionCache(&cd.TrustPolicyExceptionCache)
	client.RegisterRecvAutoProvPolicyCache(&cd.AutoProvPolicyCache)
	client.RegisterRecvAutoScalePolicyCache(&cd.AutoScalePolicyCache)
	client.RegisterRecvAlertPolicyCache(&cd.AlertPolicyCache)
	client.RegisterSendAllRecv(cd)
	nodeMgr.RegisterClient(client)
}

func InitSrvNotify(notifyServer *notify.ServerMgr, nodeMgr *node.NodeMgr, controllerData *ControllerData) {
	notifyServer.RegisterSendSettingsCache(&controllerData.SettingsCache)
	notifyServer.RegisterSendFlavorCache(&controllerData.FlavorCache)
	notifyServer.RegisterSendVMPoolCache(&controllerData.VMPoolCache)
	notifyServer.RegisterSendVMPoolInfoCache(&controllerData.VMPoolInfoCache)
	notifyServer.RegisterSendCloudletCache(controllerData.CloudletCache)
	notifyServer.RegisterSendCloudletInternalCache(&controllerData.CloudletInternalCache)
	notifyServer.RegisterSendAutoProvPolicyCache(&controllerData.AutoProvPolicyCache)
	notifyServer.RegisterSendAutoScalePolicyCache(&controllerData.AutoScalePolicyCache)
	notifyServer.RegisterSendAlertPolicyCache(&controllerData.AlertPolicyCache)
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
