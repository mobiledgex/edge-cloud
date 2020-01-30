package main

import (
	"context"
	"strings"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

// Implement notify.RecvAppInstHandler
type AppHandler struct {
}

type AppInstHandler struct {
}

type CloudletInfoHandler struct {
}

func (s *AppHandler) Update(ctx context.Context, in *edgeproto.App, rev int64) {
	dmecommon.AddApp(in)
}

func (s *AppHandler) Delete(ctx context.Context, in *edgeproto.App, rev int64) {
	dmecommon.RemoveApp(in)
}

func (s *AppHandler) Prune(ctx context.Context, keys map[edgeproto.AppKey]struct{}) {
	dmecommon.PruneApps(keys)
}

func (s *AppHandler) Flush(ctx context.Context, notifyId int64) {}

func (s *AppInstHandler) Update(ctx context.Context, in *edgeproto.AppInst, rev int64) {
	dmecommon.AddAppInst(in)
}

func (s *AppInstHandler) Delete(ctx context.Context, in *edgeproto.AppInst, rev int64) {
	dmecommon.RemoveAppInst(in)
}

func (s *AppInstHandler) Prune(ctx context.Context, keys map[edgeproto.AppInstKey]struct{}) {
	dmecommon.PruneAppInsts(keys)
}

func (s *AppInstHandler) Flush(ctx context.Context, notifyId int64) {}

func (s *CloudletInfoHandler) Update(ctx context.Context, in *edgeproto.CloudletInfo, rev int64) {
	dmecommon.SetInstStateForCloudlet(in)
}

func (s *CloudletInfoHandler) Delete(ctx context.Context, in *edgeproto.CloudletInfo, rev int64) {
	dmecommon.DeleteCloudletInfo(in)
}

func (s *CloudletInfoHandler) Prune(ctx context.Context, keys map[edgeproto.CloudletKey]struct{}) {
	dmecommon.PruneCloudlets(keys)
}

func (s *CloudletInfoHandler) Flush(ctx context.Context, notifyId int64) {}

var nodeCache edgeproto.NodeCache
var ClientSender *notify.AppInstClientSend

func initNotifyClient(addrs string, tlsCertFile string) *notify.Client {
	edgeproto.InitNodeCache(&nodeCache)
	//	edgeproto.InitAppInstClientCache(&clientCache)
	notifyClient := notify.NewClient(strings.Split(addrs, ","), tlsCertFile)
	notifyClient.RegisterRecv(notify.GlobalSettingsRecv(&dmecommon.Settings, dmecommon.SettingsUpdated))
	notifyClient.RegisterRecv(notify.NewAutoProvPolicyRecv(&dmecommon.AutoProvPolicyHandler{}))
	notifyClient.RegisterRecv(notify.NewOperatorCodeRecv(&dmecommon.DmeAppTbl.OperatorCodes))
	notifyClient.RegisterRecv(notify.NewAppRecv(&AppHandler{}))
	notifyClient.RegisterRecv(notify.NewAppInstRecv(&AppInstHandler{}))
	notifyClient.RegisterRecv(notify.NewClusterInstRecv(&dmecommon.DmeAppTbl.FreeReservableClusterInsts))
	notifyClient.RegisterSendNodeCache(&nodeCache)
	//	notifyClient.RegisterSendAppInstClientCache(&clientCache)
	notifyClient.RegisterRecv(notify.NewCloudletInfoRecv(&CloudletInfoHandler{}))
	ClientSender = notify.NewAppInstClientSend()
	notifyClient.RegisterSend(ClientSender)

	log.InfoLog("notify client to", "addrs", addrs)
	return notifyClient
}
