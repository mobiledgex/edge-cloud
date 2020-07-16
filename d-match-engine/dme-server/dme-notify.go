package main

import (
	"context"
	"strings"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
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

type CloudletInfoHandler struct {
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
	PurgeAppInstClients(ctx, &in.Key)
}

func (s *AppInstHandler) Prune(ctx context.Context, keys map[edgeproto.AppInstKey]struct{}) {
	dmecommon.PruneAppInsts(ctx, keys)
}

func (s *AppInstHandler) Flush(ctx context.Context, notifyId int64) {}

func (s *CloudletInfoHandler) Update(ctx context.Context, in *edgeproto.CloudletInfo, rev int64) {
	dmecommon.SetInstStateForCloudlet(ctx, in)
}

func (s *CloudletInfoHandler) Delete(ctx context.Context, in *edgeproto.CloudletInfo, rev int64) {
	dmecommon.DeleteCloudletInfo(ctx, in)
}

func (s *CloudletInfoHandler) Prune(ctx context.Context, keys map[edgeproto.CloudletKey]struct{}) {
	dmecommon.PruneCloudlets(ctx, keys)
}

func (s *CloudletInfoHandler) Flush(ctx context.Context, notifyId int64) {}

var nodeCache edgeproto.NodeCache
var ClientSender *notify.AppInstClientSend
var appInstClientKeyCache edgeproto.AppInstClientKeyCache
var platformClientsCache edgeproto.DeviceCache

func initNotifyClient(ctx context.Context, addrs string, tlsDialOption grpc.DialOption) *notify.Client {
	edgeproto.InitNodeCache(&nodeCache)
	edgeproto.InitAppInstClientKeyCache(&appInstClientKeyCache)
	edgeproto.InitDeviceCache(&platformClientsCache)
	appInstClientKeyCache.SetUpdatedCb(SendCachedClients)
	notifyClient := notify.NewClient(nodeMgr.Name(), strings.Split(addrs, ","), tlsDialOption)
	notifyClient.RegisterRecv(notify.GlobalSettingsRecv(&dmecommon.Settings, dmecommon.SettingsUpdated))
	notifyClient.RegisterRecv(notify.NewAutoProvPolicyRecv(&dmecommon.AutoProvPolicyHandler{}))
	notifyClient.RegisterRecv(notify.NewOperatorCodeRecv(&dmecommon.DmeAppTbl.OperatorCodes))
	notifyClient.RegisterRecv(notify.NewAppRecv(&AppHandler{}))
	notifyClient.RegisterRecv(notify.NewAppInstRecv(&AppInstHandler{}))
	notifyClient.RegisterRecv(notify.NewClusterInstRecv(&dmecommon.DmeAppTbl.FreeReservableClusterInsts))
	notifyClient.RegisterRecvAppInstClientKeyCache(&appInstClientKeyCache)

	notifyClient.RegisterSendNodeCache(&nodeCache)
	notifyClient.RegisterSendDeviceCache(&platformClientsCache)
	platformClientsCache.SetFlushAll()
	notifyClient.RegisterRecv(notify.NewCloudletInfoRecv(&CloudletInfoHandler{}))
	ClientSender = notify.NewAppInstClientSend()
	notifyClient.RegisterSend(ClientSender)

	log.SpanLog(ctx, log.DebugLevelInfo, "notify client to", "addrs", addrs)
	return notifyClient
}
