package main

import (
	"context"
	"strings"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"google.golang.org/grpc"
)

// Implement notify.RecvAppInstHandler
type AppHandler struct {
}

type AppInstHandler struct {
}

type CloudletHandler struct {
}

type CloudletInfoHandler struct {
}

type FlowRateLimitSettingsHandler struct {
}

type MaxReqsRateLimitSettingsHandler struct {
}

func (s *AppHandler) Update(ctx context.Context, in *edgeproto.App, rev int64) {
	dmecommon.AddApp(ctx, in)
}

func (s *AppHandler) Delete(ctx context.Context, in *edgeproto.App, rev int64) {
	dmecommon.RemoveApp(ctx, in)
}

func (s *AppHandler) Prune(ctx context.Context, keys map[edgeproto.AppKey]struct{}) {
	dmecommon.PruneApps(ctx, keys)
}

func (s *AppHandler) Flush(ctx context.Context, notifyId int64) {}

func (s *AppInstHandler) Update(ctx context.Context, in *edgeproto.AppInst, rev int64) {
	dmecommon.AddAppInst(ctx, in)
}

func (s *AppInstHandler) Delete(ctx context.Context, in *edgeproto.AppInst, rev int64) {
	dmecommon.RemoveAppInst(ctx, in)
	dmecommon.PurgeAppInstClients(ctx, &in.Key)
}

func (s *AppInstHandler) Prune(ctx context.Context, keys map[edgeproto.AppInstKey]struct{}) {
	dmecommon.PruneAppInsts(ctx, keys)
}

func (s *AppInstHandler) Flush(ctx context.Context, notifyId int64) {}

func (s *CloudletHandler) Update(ctx context.Context, in *edgeproto.Cloudlet, rev int64) {
	// * use cloudlet object for maintenance state as this state is used
	//   by controller to avoid end-user interacting with cloudlets for
	//   appinst/clusterinst actions. Refer SetInstMaintenanceStateForCloudlet
	// * use cloudletInfo object for cloudlet state as this correctly gives
	//   information if cloudlet is online or not
	dmecommon.SetInstStateFromCloudlet(ctx, in)
}

func (s *CloudletHandler) Delete(ctx context.Context, in *edgeproto.Cloudlet, rev int64) {
	// If cloudlet object, doesn't exist then delete it from DME refs
	// even if cloudletInfo for the same exists
	dmecommon.DeleteCloudletInfo(ctx, &in.Key)
}

func (s *CloudletHandler) Prune(ctx context.Context, keys map[edgeproto.CloudletKey]struct{}) {
	// If cloudlet object, doesn't exist then delete it from DME refs
	// even if cloudletInfo for the same exists
	dmecommon.PruneCloudlets(ctx, keys)
}

func (s *CloudletHandler) Flush(ctx context.Context, notifyId int64) {}

func (s *CloudletInfoHandler) Update(ctx context.Context, in *edgeproto.CloudletInfo, rev int64) {
	// * use cloudlet object for maintenance state as this state is used
	//   by controller to avoid end-user interacting with cloudlets for
	//   appinst/clusterinst actions. Refer SetInstMaintenanceStateForCloudlet
	// * use cloudletInfo object for cloudlet state as this correctly gives
	//   information if cloudlet is online or not
	dmecommon.SetInstStateFromCloudletInfo(ctx, in)
}

func (s *CloudletInfoHandler) Delete(ctx context.Context, in *edgeproto.CloudletInfo, rev int64) {
	// set cloudlet state for the instance accordingly
	in.State = dme.CloudletState_CLOUDLET_STATE_NOT_PRESENT
	dmecommon.SetInstStateFromCloudletInfo(ctx, in)
}

func (s *CloudletInfoHandler) Prune(ctx context.Context, keys map[edgeproto.CloudletKey]struct{}) {
	// set cloudlet state for all the instances accordingly
	dmecommon.PruneInstsCloudletState(ctx, keys)
}

func (s *CloudletInfoHandler) Flush(ctx context.Context, notifyId int64) {}

func (r *FlowRateLimitSettingsHandler) Update(ctx context.Context, in *edgeproto.FlowRateLimitSettings, rev int64) {
	if in.Key.RateLimitKey.ApiEndpointType == edgeproto.ApiEndpointType_DME {
		// Update RateLimitMgr with updated RateLimitSettings
		dmecommon.RateLimitMgr.UpdateFlowRateLimitSettings(in)
	}
}

func (r *FlowRateLimitSettingsHandler) Delete(ctx context.Context, in *edgeproto.FlowRateLimitSettings, rev int64) {
	if in.Key.RateLimitKey.ApiEndpointType == edgeproto.ApiEndpointType_DME {
		dmecommon.RateLimitMgr.RemoveFlowRateLimitSettings(in.Key)
	}
}

func (r *FlowRateLimitSettingsHandler) Prune(ctx context.Context, keys map[edgeproto.FlowRateLimitSettingsKey]struct{}) {
	dmecommon.RateLimitMgr.PruneFlowRateLimitSettings(keys)
}

func (r *FlowRateLimitSettingsHandler) Flush(ctx context.Context, notifyId int64) {}

func (r *MaxReqsRateLimitSettingsHandler) Update(ctx context.Context, in *edgeproto.MaxReqsRateLimitSettings, rev int64) {
	if in.Key.RateLimitKey.ApiEndpointType == edgeproto.ApiEndpointType_DME {
		// Update RateLimitMgr with updated RateLimitSettings
		dmecommon.RateLimitMgr.UpdateMaxReqsRateLimitSettings(in)
	}
}

func (r *MaxReqsRateLimitSettingsHandler) Delete(ctx context.Context, in *edgeproto.MaxReqsRateLimitSettings, rev int64) {
	if in.Key.RateLimitKey.ApiEndpointType == edgeproto.ApiEndpointType_DME {
		dmecommon.RateLimitMgr.RemoveMaxReqsRateLimitSettings(in.Key)
	}
}

func (r *MaxReqsRateLimitSettingsHandler) Prune(ctx context.Context, keys map[edgeproto.MaxReqsRateLimitSettingsKey]struct{}) {
	dmecommon.RateLimitMgr.PruneMaxReqsRateLimitSettings(keys)
}

func (r *MaxReqsRateLimitSettingsHandler) Flush(ctx context.Context, notifyId int64) {}

var nodeCache edgeproto.NodeCache
var flowRateLimitSettingsCache edgeproto.FlowRateLimitSettingsCache
var maxReqsRateLimitSettingsCache edgeproto.MaxReqsRateLimitSettingsCache

func initNotifyClient(ctx context.Context, addrs string, tlsDialOption grpc.DialOption, notifyOps ...notify.ClientOp) *notify.Client {
	edgeproto.InitNodeCache(&nodeCache)
	edgeproto.InitAppInstClientKeyCache(&dmecommon.AppInstClientKeyCache)
	edgeproto.InitDeviceCache(&dmecommon.PlatformClientsCache)
	dmecommon.AppInstClientKeyCache.SetUpdatedCb(dmecommon.SendCachedClients)
	edgeproto.InitFlowRateLimitSettingsCache(&flowRateLimitSettingsCache)
	edgeproto.InitMaxReqsRateLimitSettingsCache(&maxReqsRateLimitSettingsCache)
	notifyClient := notify.NewClient(nodeMgr.Name(), strings.Split(addrs, ","), tlsDialOption, notifyOps...)
	notifyClient.RegisterRecv(notify.GlobalSettingsRecv(&dmecommon.Settings, dmecommon.SettingsUpdated))
	notifyClient.RegisterRecv(notify.NewAutoProvPolicyRecv(&dmecommon.AutoProvPolicyHandler{}))
	notifyClient.RegisterRecv(notify.NewOperatorCodeRecv(&dmecommon.DmeAppTbl.OperatorCodes))
	notifyClient.RegisterRecv(notify.NewAppRecv(&AppHandler{}))
	notifyClient.RegisterRecv(notify.NewCloudletRecv(&CloudletHandler{}))
	notifyClient.RegisterRecv(notify.NewAppInstRecv(&AppInstHandler{}))
	notifyClient.RegisterRecv(notify.NewClusterInstRecv(&dmecommon.DmeAppTbl.FreeReservableClusterInsts))
	notifyClient.RegisterRecv(notify.NewFlowRateLimitSettingsRecv(&FlowRateLimitSettingsHandler{}))
	notifyClient.RegisterRecv(notify.NewMaxReqsRateLimitSettingsRecv(&MaxReqsRateLimitSettingsHandler{}))
	notifyClient.RegisterRecvAppInstClientKeyCache(&dmecommon.AppInstClientKeyCache)

	notifyClient.RegisterSendNodeCache(&nodeCache)
	notifyClient.RegisterSendDeviceCache(&dmecommon.PlatformClientsCache)
	dmecommon.PlatformClientsCache.SetFlushAll()
	notifyClient.RegisterRecv(notify.NewCloudletInfoRecv(&CloudletInfoHandler{}))
	dmecommon.ClientSender = notify.NewAppInstClientSend()
	notifyClient.RegisterSend(dmecommon.ClientSender)

	log.SpanLog(ctx, log.DebugLevelInfo, "notify client to", "addrs", addrs)
	return notifyClient
}
