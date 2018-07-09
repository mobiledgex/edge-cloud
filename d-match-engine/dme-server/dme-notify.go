package main

import (
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

// Implement notify.RecvAppInstHandler
type AppInstHandler struct {
}

func (s *AppInstHandler) Update(in *edgeproto.AppInst, rev int64) {
	addApp(in)
}

func (s *AppInstHandler) Delete(in *edgeproto.AppInst, rev int64) {
	removeApp(in)
}

func (s *AppInstHandler) Prune(keys map[edgeproto.AppInstKey]struct{}) {
	pruneApps(keys)
}

func NewNotifyHandler() *notify.DefaultHandler {
	handler := notify.DefaultHandler{}
	handler.RecvAppInst = &AppInstHandler{}
	return &handler
}

func initNotifyClient(addrs string) *notify.Client {
	handler := NewNotifyHandler()
	notifyClient := notify.NewDMEClient(strings.Split(addrs, ","), handler)
	log.InfoLog("notify client to", "addrs", addrs)
	return notifyClient
}
