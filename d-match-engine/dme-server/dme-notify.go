package main

import (
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

var cloudletID = 5000

type NotifyHandler struct {
}

func (s *NotifyHandler) HandleSendAllDone(allMaps *notify.AllMaps) {
	pruneApps(allMaps)
}

func (s *NotifyHandler) HandleNotice(notice *edgeproto.NoticeReply) error {
	appInst := notice.GetAppInst()
	if appInst != nil {
		if notice.Action == edgeproto.NoticeAction_UPDATE {
			log.DebugLog(log.DebugLevelNotify,
				"notice app inst update",
				"key", appInst.Key.GetKeyString())
			addApp(appInst)
		} else if notice.Action == edgeproto.NoticeAction_DELETE {
			log.DebugLog(log.DebugLevelNotify,
				"notice app inst delete",
				"key", appInst.Key.GetKeyString())
			removeApp(appInst)
		}
	}
	return nil
}

func initNotifyClient(addrs string, handler *NotifyHandler) *notify.Client {
	notifyClient := notify.NewDMEClient(strings.Split(addrs, ","), handler)
	log.InfoLog("notify client to", "addrs", addrs)
	return notifyClient
}
