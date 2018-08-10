package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/util"
)

type Sync struct {
	store      objstore.KVStore
	rev        int64
	mux        util.Mutex
	cond       sync.Cond
	initWait   bool
	syncDone   bool
	syncCancel context.CancelFunc
	caches     map[string]ObjCache
}

type ObjCache interface {
	SyncUpdate(key, val []byte, rev int64)
	SyncDelete(key []byte, rev int64)
	SyncListStart()
	SyncListEnd()
	GetTypeString() string
}

func InitSync(store objstore.KVStore) *Sync {
	sync := Sync{}
	sync.store = store
	sync.initWait = true
	sync.mux.InitCond(&sync.cond)
	sync.caches = make(map[string]ObjCache)
	sync.rev = 1
	return &sync
}

func (s *Sync) RegisterCache(cache ObjCache) {
	s.caches[cache.GetTypeString()] = cache
}

// Watch on all key changes in a single thread.
// Compared to watching on different objects in separate threads,
// this prevents race conditions when objects have dependencies on
// each other, i.e. an update to one object must update other types
// of objects.
func (s *Sync) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.syncCancel = cancel
	go func() {
		prefix := fmt.Sprintf("%d/", objstore.GetRegion())
		err := s.store.Sync(ctx, prefix, s.syncCb)
		if err != nil {
			log.WarnLog("sync failed", "err", err)
		}
		s.syncDone = true
		s.cond.Broadcast()
	}()

	// Wait until the initial show of the sync call is complete.
	// All caches should be updated when done.
	s.mux.Lock()
	defer s.mux.Unlock()
	for s.initWait {
		s.cond.Wait()
	}
}

func (s *Sync) Done() {
	s.syncCancel()
	s.mux.Lock()
	for !s.syncDone {
		s.cond.Wait()
	}
}

func (s *Sync) GetCache(key []byte) (ObjCache, bool) {
	_, typ, _, err := objstore.DbKeyPrefixParse(string(key))
	if err != nil {
		log.WarnLog("Failed to parse db key", "key", key, "err", err)
		return nil, false
	}
	cache, found := s.caches[typ]
	if !found {
		log.DebugLog(log.DebugLevelApi, "No cache for type", "typ", typ)
	}
	return cache, found
}

// syncCb calls into caches to update them with the new data.
// This thread context is the only context in which we can modify the cache
// data, otherwise there could be race conditions against the sync data
// coming from etcd.
func (s *Sync) syncCb(data *objstore.SyncCbData) {
	log.DebugLog(log.DebugLevelApi, "Sync cb", "action", objstore.SyncActionStrs[data.Action], "key", string(data.Key), "value", string(data.Value), "rev", data.Rev)

	s.mux.Lock()
	switch data.Action {
	case objstore.SyncListStart:
		for _, cache := range s.caches {
			cache.SyncListStart()
		}
	case objstore.SyncListEnd:
		for _, cache := range s.caches {
			cache.SyncListEnd()
		}
	case objstore.SyncList:
		fallthrough
	case objstore.SyncUpdate:
		if cache, found := s.GetCache(data.Key); found {
			cache.SyncUpdate(data.Key, data.Value, data.Rev)
		}
		if !data.MoreEvents {
			s.rev = data.Rev
		}
	case objstore.SyncDelete:
		if cache, found := s.GetCache(data.Key); found {
			cache.SyncDelete(data.Key, data.Rev)
		}
		if !data.MoreEvents {
			s.rev = data.Rev
		}
	}

	// notify any threads waiting on cache update to finish
	if s.initWait && data.Action == objstore.SyncListEnd {
		s.initWait = false
	}
	s.cond.Broadcast()
	s.mux.Unlock()
}

// syncWait is used by API calls to wait until data has been updated in cache
func (s *Sync) syncWait(rev int64) {
	s.mux.Lock()
	defer s.mux.Unlock()
	log.DebugLog(log.DebugLevelApi, "syncWait", "cur-rev", s.rev, "wait-rev", rev)
	for s.rev < rev && !s.syncDone {
		s.cond.Wait()
	}
}

func (s *Sync) ApplySTMWait(apply func(concurrency.STM) error) error {
	rev, err := s.store.ApplySTM(apply)
	if err == nil {
		s.syncWait(rev)
	}
	return err
}
