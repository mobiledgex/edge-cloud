package main

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
)

type NotifyHandler struct {
}

func (s *NotifyHandler) HandleSendAllDone(allMaps *notify.NotifySendAllMaps) {
	util.InfoLog("Handle send all")
}

func (s *NotifyHandler) HandleNotice(notice *edgeproto.Notice) error {
	appInst := notice.GetAppInst()
	if appInst != nil {
		if notice.Action == edgeproto.NoticeAction_UPDATE {
			util.InfoLog("notice app inst update", "key", appInst.Key.GetKeyString())
		} else if notice.Action == edgeproto.NoticeAction_DELETE {
			util.InfoLog("notice app inst delete", "key", appInst.Key.GetKeyString())
		}
	}
	return nil
}
