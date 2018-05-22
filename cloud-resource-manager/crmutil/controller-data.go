package crmutil

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
)

type ControllerData struct {
	cloudlets    map[edgeproto.CloudletKey]*edgeproto.Cloudlet
	cloudletsMux util.Mutex
	appInsts     map[edgeproto.AppInstKey]*edgeproto.AppInst
	appInstsMux  util.Mutex
}

// NewControllerData creates a new instance to track data from the controller
func NewControllerData() *ControllerData {
	cd := &ControllerData{}
	cd.cloudlets = make(map[edgeproto.CloudletKey]*edgeproto.Cloudlet)
	cd.appInsts = make(map[edgeproto.AppInstKey]*edgeproto.AppInst)
	return cd
}

func (cd *ControllerData) UpdateAppInst(inst *edgeproto.AppInst) {
	cd.appInstsMux.Lock()
	defer cd.appInstsMux.Unlock()
	cd.appInsts[inst.Key] = inst
}

func (cd *ControllerData) DeleteAppInst(inst *edgeproto.AppInst) {
	cd.appInstsMux.Lock()
	defer cd.appInstsMux.Unlock()
	delete(cd.appInsts, inst.Key)
}

func (cd *ControllerData) GetAppInst(key *edgeproto.AppInstKey, inst *edgeproto.AppInst) bool {
	cd.appInstsMux.Lock()
	defer cd.appInstsMux.Unlock()
	if val, found := cd.appInsts[*key]; found {
		*inst = *val
		return true
	}
	return false
}

func (cd *ControllerData) UpdateCloudlet(inst *edgeproto.Cloudlet) {
	cd.cloudletsMux.Lock()
	defer cd.cloudletsMux.Unlock()
	cd.cloudlets[inst.Key] = inst
}

func (cd *ControllerData) DeleteCloudlet(inst *edgeproto.Cloudlet) {
	cd.cloudletsMux.Lock()
	defer cd.cloudletsMux.Unlock()
	delete(cd.cloudlets, inst.Key)
}

func (cd *ControllerData) GetCloudlet(key *edgeproto.CloudletKey, inst *edgeproto.Cloudlet) bool {
	cd.cloudletsMux.Lock()
	defer cd.cloudletsMux.Unlock()
	if val, found := cd.cloudlets[*key]; found {
		*inst = *val
		return true
	}
	return false
}

func (cd *ControllerData) HandleNotifyDone(allMaps *notify.NotifySendAllMaps) {
	cd.appInstsMux.Lock()
	for key, val := range cd.appInsts {
		if _, found := allMaps.AppInsts[key]; !found {
			delete(cd.appInsts, val.Key)
		}
	}
	cd.appInstsMux.Unlock()

	cd.cloudletsMux.Lock()
	for key, val := range cd.cloudlets {
		if _, found := allMaps.Cloudlets[key]; !found {
			delete(cd.cloudlets, val.Key)
		}
	}
	cd.cloudletsMux.Unlock()
}
