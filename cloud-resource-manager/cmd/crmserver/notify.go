package main

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
)

type NotifyHandler struct {
	cd *crmutil.ControllerData
}

func NewNotifyHandler(cd *crmutil.ControllerData) *NotifyHandler {
	return &NotifyHandler{cd: cd}
}

func (s *NotifyHandler) HandleSendAllDone(allMaps *notify.NotifySendAllMaps) {
	s.cd.HandleNotifyDone(allMaps)
}

func (s *NotifyHandler) HandleNotice(notice *edgeproto.Notice) error {
	appInst := notice.GetAppInst()
	if appInst != nil {
		if notice.Action == edgeproto.NoticeAction_UPDATE {
			s.cd.UpdateAppInst(appInst)
		} else if notice.Action == edgeproto.NoticeAction_DELETE {
			s.cd.DeleteAppInst(appInst)
		}
	}
	cloudlet := notice.GetCloudlet()
	if cloudlet != nil {
		if notice.Action == edgeproto.NoticeAction_UPDATE {
			s.cd.UpdateCloudlet(cloudlet)
		} else if notice.Action == edgeproto.NoticeAction_DELETE {
			s.cd.DeleteCloudlet(cloudlet)
		}
	}
	return nil
}
