package main

import (
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstRefsApi struct {
	sync  *Sync
	store edgeproto.AppInstRefsStore
	cache edgeproto.AppInstRefsCache
}

var appInstRefsApi = AppInstRefsApi{}

func InitAppInstRefsApi(sync *Sync) {
	appInstRefsApi.sync = sync
	appInstRefsApi.store = edgeproto.NewAppInstRefsStore(sync.store)
	edgeproto.InitAppInstRefsCache(&appInstRefsApi.cache)
	sync.RegisterCache(&appInstRefsApi.cache)
}

func (s *AppInstRefsApi) ShowAppInstRefs(in *edgeproto.AppInstRefs, cb edgeproto.AppInstRefsApi_ShowAppInstRefsServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInstRefs) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AppInstRefsApi) addRef(stm concurrency.STM, key *edgeproto.AppInstKey) {
	refs := edgeproto.AppInstRefs{}
	if !s.store.STMGet(stm, &key.AppKey, &refs) {
		refs.Key = key.AppKey
		refs.Insts = make(map[string]uint32)
	}
	refs.Insts[key.GetKeyString()] = 1
	s.store.STMPut(stm, &refs)
}

func (s *AppInstRefsApi) removeRef(stm concurrency.STM, key *edgeproto.AppInstKey) {
	refs := edgeproto.AppInstRefs{}
	if !s.store.STMGet(stm, &key.AppKey, &refs) {
		return
	}
	delete(refs.Insts, key.GetKeyString())
	if len(refs.Insts) == 0 {
		s.store.STMDel(stm, &key.AppKey)
	} else {
		s.store.STMPut(stm, &refs)
	}
}
