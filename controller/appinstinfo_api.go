package main

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstInfoApi struct {
	sync  *Sync
	store edgeproto.AppInstInfoStore
}

var appInstInfoApi = AppInstInfoApi{}

func InitAppInstInfoApi(sync *Sync) {
	appInstInfoApi.sync = sync
	appInstInfoApi.store = edgeproto.NewAppInstInfoStore(sync.store)
}

func (s *AppInstInfoApi) Update(in *edgeproto.AppInstInfo, rev int64) {
	appInstApi.UpdateFromInfo(in)
}

func (s *AppInstInfoApi) Delete(in *edgeproto.AppInstInfo, rev int64) {
	appInstApi.DeleteFromInfo(in)
}

func (s *AppInstInfoApi) Flush(notifyId int64) {
	// no-op
}
