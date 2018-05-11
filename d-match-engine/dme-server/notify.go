package main

import (
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/proto"
	"github.com/mobiledgex/edge-cloud/util"
)

type NotifyHandler struct {
}

func (s *NotifyHandler) HandleSendAllDone(allMaps *notify.NotifySendAllMaps) {
	util.InfoLog("Handle send all")
}

func (s *NotifyHandler) HandleNotice(notice *proto.Notice) error {
	appInst := notice.GetAppInst()
	if appInst != nil {
		if notice.Action == proto.NoticeAction_UPDATE {
			util.InfoLog("notice app inst update", "key", appInst.Key.GetKeyString())
		} else if notice.Action == proto.NoticeAction_DELETE {
			util.InfoLog("notice app inst delete", "key", appInst.Key.GetKeyString())
		}
	}
	return nil
}
