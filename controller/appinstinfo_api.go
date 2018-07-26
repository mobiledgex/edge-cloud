package main

import (
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
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
	if !appInstApi.HasKey(in.GetKey()) {
		fmt.Printf("app inst not found %v\n", in.GetKey())
		return
	}
	// for now assume all fields have been specified
	in.Fields = edgeproto.AppInstInfoAllFields
	s.store.Put(in, nil)
}

func (s *AppInstInfoApi) internalDelete(stm concurrency.STM, key *edgeproto.AppInstKey) {
	s.store.STMDel(stm, key)
}

// Delete for notify never actually deletes the data
func (s *AppInstInfoApi) Delete(in *edgeproto.AppInstInfo, notifyId int64) {
	// no-op
}

func (s *AppInstInfoApi) Flush(notifyId int64) {
	// no-op
}
