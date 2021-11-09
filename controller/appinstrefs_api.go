package main

import (
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstRefsApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.AppInstRefsStore
	cache edgeproto.AppInstRefsCache
}

func NewAppInstRefsApi(sync *Sync, all *AllApis) *AppInstRefsApi {
	appInstRefsApi := AppInstRefsApi{}
	appInstRefsApi.all = all
	appInstRefsApi.sync = sync
	appInstRefsApi.store = edgeproto.NewAppInstRefsStore(sync.store)
	edgeproto.InitAppInstRefsCache(&appInstRefsApi.cache)
	sync.RegisterCache(&appInstRefsApi.cache)
	return &appInstRefsApi
}

func (s *AppInstRefsApi) ShowAppInstRefs(in *edgeproto.AppInstRefs, cb edgeproto.AppInstRefsApi_ShowAppInstRefsServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInstRefs) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AppInstRefsApi) createRef(stm concurrency.STM, key *edgeproto.AppKey) {
	refs := edgeproto.AppInstRefs{}
	refs.Key = *key
	refs.Insts = make(map[string]uint32)
	refs.DeleteRequestedInsts = make(map[string]uint32)
	s.store.STMPut(stm, &refs)
}

func (s *AppInstRefsApi) deleteRef(stm concurrency.STM, key *edgeproto.AppKey) {
	s.store.STMDel(stm, key)
}

func (s *AppInstRefsApi) addRef(stm concurrency.STM, key *edgeproto.AppInstKey) {
	refs := edgeproto.AppInstRefs{}
	if !s.store.STMGet(stm, &key.AppKey, &refs) {
		refs.Key = key.AppKey
		refs.Insts = make(map[string]uint32)
		refs.DeleteRequestedInsts = make(map[string]uint32)
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
	delete(refs.DeleteRequestedInsts, key.GetKeyString())
	s.store.STMPut(stm, &refs)
}

func (s *AppInstRefsApi) addDeleteRequestedRef(stm concurrency.STM, key *edgeproto.AppInstKey) {
	refs := edgeproto.AppInstRefs{}
	if !s.store.STMGet(stm, &key.AppKey, &refs) {
		refs.Key = key.AppKey
		refs.Insts = make(map[string]uint32)
		refs.DeleteRequestedInsts = make(map[string]uint32)
	}
	refs.DeleteRequestedInsts[key.GetKeyString()] = 1
	s.store.STMPut(stm, &refs)
}

func (s *AppInstRefsApi) removeDeleteRequestedRef(stm concurrency.STM, key *edgeproto.AppInstKey) {
	refs := edgeproto.AppInstRefs{}
	if !s.store.STMGet(stm, &key.AppKey, &refs) {
		return
	}
	delete(refs.DeleteRequestedInsts, key.GetKeyString())
	s.store.STMPut(stm, &refs)
}
