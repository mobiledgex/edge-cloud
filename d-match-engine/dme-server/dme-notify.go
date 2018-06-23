package main

import (
	//"net"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
)

type NotifyHandler struct {
}

func (s *NotifyHandler) HandleSendAllDone(allMaps *notify.AllMaps) {
	pruneApps(allMaps)
}

func (s *NotifyHandler) HandleNotice(notice *edgeproto.NoticeReply) error {
	appInst := notice.GetAppInst()
	if appInst != nil {
		if notice.Action == edgeproto.NoticeAction_UPDATE {
			util.DebugLog(util.DebugLevelNotify,
				"notice app inst update",
				"key", appInst.Key.GetKeyString())
			addApp(appInst)
		} else if notice.Action == edgeproto.NoticeAction_DELETE {
			util.DebugLog(util.DebugLevelNotify,
				"notice app inst delete",
				"key", appInst.Key.GetKeyString())
			removeApp(appInst)
		}
	}
	return nil
}

func initNotifyClient(addrs string, handler *NotifyHandler) *notify.Client {
	notifyClient := notify.NewDMEClient(strings.Split(addrs, ","), handler)
	util.InfoLog("notify client to", "addrs", addrs)
	return notifyClient
}
