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

type RateLimitSettingsHandler struct {
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

func (r *RateLimitSettingsHandler) Update(ctx context.Context, in *edgeproto.RateLimitSettings, rev int64) {
	if in.Key.ApiEndpointType == edgeproto.ApiEndpointType_DME {
		// Update RateLimitMgr with updated RateLimitSettings
		dmecommon.UnaryRateLimitMgr.UpdateRateLimitSettings(in)
		dmecommon.StreamRateLimitMgr.UpdateRateLimitSettings(in)
	}
}

func (r *RateLimitSettingsHandler) Delete(ctx context.Context, in *edgeproto.RateLimitSettings, rev int64) {
	if in.Key.ApiEndpointType == edgeproto.ApiEndpointType_DME {
		// "Remove" RateLimitSettings from RateLimitMgr (ie. remove settings but keep the key)
		in = &edgeproto.RateLimitSettings{
			Key: in.Key,
		}
		dmecommon.UnaryRateLimitMgr.UpdateRateLimitSettings(in)
		dmecommon.StreamRateLimitMgr.UpdateRateLimitSettings(in)
	}
}

func (r *RateLimitSettingsHandler) Prune(ctx context.Context, keys map[edgeproto.RateLimitSettingsKey]struct{}) {
	for key, _ := range keys {
		if key.ApiEndpointType == edgeproto.ApiEndpointType_DME {
			// "Remove" RateLimitSettings from RateLimitMgr (ie. remove settings but keep the key)
			in := &edgeproto.RateLimitSettings{
				Key: key,
			}
			dmecommon.UnaryRateLimitMgr.UpdateRateLimitSettings(in)
			dmecommon.StreamRateLimitMgr.UpdateRateLimitSettings(in)
		}
	}
}

func (r *RateLimitSettingsHandler) Flush(ctx context.Context, notifyId int64) {}

var nodeCache edgeproto.NodeCache
var rateLimitSettingsCache edgeproto.RateLimitSettingsCache

func initNotifyClient(ctx context.Context, addrs string, tlsDialOption grpc.DialOption, notifyOps ...notify.ClientOp) *notify.Client {
	edgeproto.InitNodeCache(&nodeCache)
	edgeproto.InitAppInstClientKeyCache(&dmecommon.AppInstClientKeyCache)
	edgeproto.InitDeviceCache(&dmecommon.PlatformClientsCache)
	dmecommon.AppInstClientKeyCache.SetUpdatedCb(dmecommon.SendCachedClients)
	edgeproto.InitRateLimitSettingsCache(&rateLimitSettingsCache)
	notifyClient := notify.NewClient(nodeMgr.Name(), strings.Split(addrs, ","), tlsDialOption, notifyOps...)
	notifyClient.RegisterRecv(notify.GlobalSettingsRecv(&dmecommon.Settings, dmecommon.SettingsUpdated))
	notifyClient.RegisterRecv(notify.NewAutoProvPolicyRecv(&dmecommon.AutoProvPolicyHandler{}))
	notifyClient.RegisterRecv(notify.NewOperatorCodeRecv(&dmecommon.DmeAppTbl.OperatorCodes))
	notifyClient.RegisterRecv(notify.NewAppRecv(&AppHandler{}))
	notifyClient.RegisterRecv(notify.NewCloudletRecv(&CloudletHandler{}))
	notifyClient.RegisterRecv(notify.NewAppInstRecv(&AppInstHandler{}))
	notifyClient.RegisterRecv(notify.NewClusterInstRecv(&dmecommon.DmeAppTbl.FreeReservableClusterInsts))
	notifyClient.RegisterRecv(notify.NewRateLimitSettingsRecv(&RateLimitSettingsHandler{}))
	notifyClient.RegisterRecvAppInstClientKeyCache(&dmecommon.AppInstClientKeyCache)

	notifyClient.RegisterSendNodeCache(&nodeCache)
	notifyClient.RegisterSendDeviceCache(&dmecommon.PlatformClientsCache)
	dmecommon.PlatformClientsCache.SetFlushAll()
	notifyClient.RegisterRecv(notify.NewCloudletInfoRecv(&CloudletInfoHandler{}))
	dmecommon.ClientSender = notify.NewAppInstClientSend()
	notifyClient.RegisterSend(dmecommon.ClientSender)
	notifyClient.RegisterSendRateLimitSettingsCache(&rateLimitSettingsCache)

	log.SpanLog(ctx, log.DebugLevelInfo, "notify client to", "addrs", addrs)
	return notifyClient
}
