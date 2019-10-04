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
	// If the cloudlet went offline we need to prevent clients from being directed to this cloudlet
	if in.State == edgeproto.CloudletState_CLOUDLET_STATE_OFFLINE {
		dmecommon.SetInstStateForCloudlet(&in.Key, edgeproto.TrackedState_TRACKED_STATE_UNKNOWN)
	} else if in.State == edgeproto.CloudletState_CLOUDLET_STATE_READY {
		// re-enable the AppInstances on that cloudlet
		dmecommon.SetInstStateForCloudlet(&in.Key, edgeproto.TrackedState_READY)
	}
}

func (s *CloudletInfoHandler) Delete(ctx context.Context, in *edgeproto.CloudletInfo, rev int64) {}

func (s *CloudletInfoHandler) Prune(ctx context.Context, keys map[edgeproto.CloudletKey]struct{}) {}

func (s *CloudletInfoHandler) Flush(ctx context.Context, notifyId int64) {}

var nodeCache edgeproto.NodeCache

func initNotifyClient(addrs string, tlsCertFile string) *notify.Client {
	edgeproto.InitNodeCache(&nodeCache)
	notifyClient := notify.NewClient(strings.Split(addrs, ","), tlsCertFile)
	notifyClient.RegisterRecv(notify.NewAppRecv(&AppHandler{}))
	notifyClient.RegisterRecv(notify.NewAppInstRecv(&AppInstHandler{}))
	notifyClient.RegisterSendNodeCache(&nodeCache)
	notifyClient.RegisterRecv(notify.NewCloudletInfoRecv(&CloudletInfoHandler{}))
	log.InfoLog("notify client to", "addrs", addrs)
	return notifyClient
}
