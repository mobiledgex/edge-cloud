package main

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstInfoApi struct {
	sync  *Sync
	store edgeproto.AppInstInfoStore
	cache edgeproto.AppInstInfoCache
}

var appInstInfoApi = AppInstInfoApi{}

func InitAppInstInfoApi(sync *Sync) {
	appInstInfoApi.sync = sync
	appInstInfoApi.store = edgeproto.NewAppInstInfoStore(sync.store)
	edgeproto.InitAppInstInfoCache(&appInstInfoApi.cache)
	sync.RegisterCache(&appInstInfoApi.cache)
}

func (s *AppInstInfoApi) ShowAppInstInfo(in *edgeproto.AppInstInfo, cb edgeproto.AppInstInfoApi_ShowAppInstInfoServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInstInfo) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AppInstInfoApi) Update(in *edgeproto.AppInstInfo, notifyId int64) {
	// note: must be applied even if appinst doesn't exist, since this
	// will get called on error on delete.
	in.Fields = edgeproto.AppInstInfoAllFields
	s.store.Put(in, s.sync.syncWait)
}

func (s *AppInstInfoApi) Delete(in *edgeproto.AppInstInfo, notifyId int64) {
	s.store.Delete(in, s.sync.syncWait)
}

func (s *AppInstInfoApi) Flush(notifyId int64) {
	// XXX Set all states to NotConnected? Need to store notifyId in cache.
	// no-op
}
