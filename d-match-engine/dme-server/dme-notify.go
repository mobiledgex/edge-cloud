package main

import (
	//"net"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
)

var cloudletID = 5000

type app struct {
	id        uint64
	name      string
	vers      string
	developer string
}

type NotifyHandler struct {
}

func (s *NotifyHandler) HandleSendAllDone(allMaps *notify.AllMaps) {
	util.InfoLog("Handle send all")
}

func (s *NotifyHandler) HandleNotice(notice *edgeproto.NoticeReply) error {
	var app_inst app
	var cloudlet_inst cloudlet
	var appkey *edgeproto.AppInstKey
	var cloudletkey *edgeproto.CloudletKey

	appInst := notice.GetAppInst()
	if appInst != nil {
		if notice.Action == edgeproto.NoticeAction_UPDATE {
			util.InfoLog("notice app inst update", "key", appInst.Key.GetKeyString())
			appkey = &appInst.Key
			app_inst.id = appkey.Id
			app_inst.name = appkey.AppKey.Name
			app_inst.vers = appkey.AppKey.Version
			app_inst.developer = appkey.AppKey.DeveloperKey.Name

			// Todo: Carrier ID needs to be there
			cloudlet_inst.id = uint64(cloudletID)
			cloudletID = cloudletID + 1
			cloudletkey = &appkey.CloudletKey
			cloudlet_inst.carrier = cloudletkey.Name
			cloudlet_inst.location = appInst.CloudletLoc
			//cloudlet_inst.accessIp = net.ParseIP(appInst,Ip)
			cloudlet_inst.accessIp = appInst.Ip

			// Add it to the app-cloudlet-inst table
			add_app(&app_inst, &cloudlet_inst)
		} else if notice.Action == edgeproto.NoticeAction_DELETE {
			util.InfoLog("notice app inst delete", "key", appInst.Key.GetKeyString())
		}
	}
	return nil
}
