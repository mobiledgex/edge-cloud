package main

import (
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

// Implement notify.RecvAppInstHandler
type AppHandler struct {
}

type AppInstHandler struct {
}

func (s *AppHandler) Update(in *edgeproto.App, rev int64) {
	addApp(in)
}

func (s *AppHandler) Delete(in *edgeproto.App, rev int64) {
	removeApp(in)
}

func (s *AppHandler) Prune(keys map[edgeproto.AppKey]struct{}) {
	pruneApps(keys)
}

func (s *AppHandler) Flush(notifyId int64) {}

func (s *AppInstHandler) Update(in *edgeproto.AppInst, rev int64) {
	addAppInst(in)
}

func (s *AppInstHandler) Delete(in *edgeproto.AppInst, rev int64) {
	removeAppInst(in)
}

func (s *AppInstHandler) Prune(keys map[edgeproto.AppInstKey]struct{}) {
	pruneAppInsts(keys)
}

func (s *AppInstHandler) Flush(notifyId int64) {}

var nodeCache edgeproto.NodeCache

func initNotifyClient(addrs string, tlsCertFile string) *notify.Client {
	edgeproto.InitNodeCache(&nodeCache)
	notifyClient := notify.NewClient(strings.Split(addrs, ","), tlsCertFile)
	notifyClient.RegisterRecv(notify.NewAppRecv(&AppHandler{}))
	notifyClient.RegisterRecv(notify.NewAppInstRecv(&AppInstHandler{}))
	notifyClient.RegisterSendNodeCache(&nodeCache)
	log.InfoLog("notify client to", "addrs", addrs)
	return notifyClient
}
