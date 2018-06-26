package main

import (
	"context"
	"errors"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/util"
)

type watcher struct {
	cb objstore.SyncCb
}

type dummyEtcd struct {
	db       map[string]string
	vers     map[string]int64
	watchers map[string]*watcher
	rev      int64
	syncCb   objstore.SyncCb
	mux      util.Mutex
}

func (e *dummyEtcd) Start() error {
	e.db = make(map[string]string)
	e.vers = make(map[string]int64)
	e.watchers = make(map[string]*watcher)
	e.rev = 1
	return nil
}

func (e *dummyEtcd) Stop() {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.db = nil
	e.vers = nil
}

func (e *dummyEtcd) Create(key, val string) (int64, error) {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.db == nil {
		return 0, objstore.ErrObjStoreNotInitialized
	}
	_, ok := e.db[key]
	if ok {
		return 0, objstore.ErrObjStoreKeyExists
	}
	e.db[key] = val
	e.vers[key] = 1
	e.rev++
	log.DebugLog(log.DebugLevelEtcd, "Created", "key", key, "val", val, "rev", e.rev)
	e.triggerWatcher(objstore.SyncUpdate, key, val, e.rev)
	return e.rev, nil
}

func (e *dummyEtcd) Update(key, val string, version int64) (int64, error) {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.db == nil {
		return 0, objstore.ErrObjStoreNotInitialized
	}
	_, ok := e.db[key]
	if !ok {
		return 0, objstore.ErrObjStoreKeyNotFound
	}
	ver := e.vers[key]
	if version != objstore.ObjStoreUpdateVersionAny && ver != version {
		return 0, errors.New("Invalid version")
	}

	e.db[key] = val
	e.vers[key] = ver + 1
	e.rev++
	log.DebugLog(log.DebugLevelEtcd, "Updated", "key", key, "val", val, "ver", ver+1, "rev", e.rev)
	e.triggerWatcher(objstore.SyncUpdate, key, val, e.rev)
	return e.rev, nil
}

func (e *dummyEtcd) Delete(key string) (int64, error) {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.db == nil {
		return 0, objstore.ErrObjStoreNotInitialized
	}
	delete(e.db, key)
	e.rev++
	log.DebugLog(log.DebugLevelEtcd, "Delete", "key", key, "rev", e.rev)
	e.triggerWatcher(objstore.SyncDelete, key, "", e.rev)
	return e.rev, nil
}

func (e *dummyEtcd) Get(key string) ([]byte, int64, error) {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.db == nil {
		return nil, 0, objstore.ErrObjStoreNotInitialized
	}
	val, ok := e.db[key]
	if !ok {
		return nil, 0, objstore.ErrObjStoreKeyNotFound
	}
	ver := e.vers[key]

	log.DebugLog(log.DebugLevelEtcd, "Got", "key", key, "val", val, "ver", ver, "rev", e.rev)
	return ([]byte)(val), ver, nil
}

func (e *dummyEtcd) List(key string, cb objstore.ListCb) error {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.db == nil {
		return objstore.ErrObjStoreNotInitialized
	}
	for k, v := range e.db {
		if !strings.HasPrefix(k, key) {
			continue
		}
		log.DebugLog(log.DebugLevelEtcd, "List", "key", k, "val", v, "rev", e.rev)
		err := cb([]byte(k), []byte(v), e.rev)
		if err != nil {
			break
		}
	}
	return nil
}

func (e *dummyEtcd) Sync(ctx context.Context, prefix string, cb objstore.SyncCb) error {
	e.mux.Lock()
	watch := watcher{
		cb: cb,
	}
	e.watchers[prefix] = &watch

	// initial callback of data
	data := objstore.SyncCbData{}
	data.Action = objstore.SyncListStart
	data.Rev = 0
	cb(&data)
	for key, val := range e.db {
		if strings.HasPrefix(key, prefix) {
			log.DebugLog(log.DebugLevelEtcd, "sync list data", "key", key, "val", val, "rev", e.rev)
			data.Action = objstore.SyncList
			data.Key = []byte(key)
			data.Value = []byte(val)
			data.Rev = e.rev
			cb(&data)
		}
	}
	data.Action = objstore.SyncListEnd
	data.Key = nil
	data.Value = nil
	cb(&data)

	e.mux.Unlock()
	<-ctx.Done()
	e.mux.Lock()

	delete(e.watchers, prefix)
	e.mux.Unlock()
	return nil
}

func (e *dummyEtcd) triggerWatcher(action objstore.SyncCbAction, key, val string, rev int64) {
	for prefix, watch := range e.watchers {
		if strings.HasPrefix(key, prefix) {
			data := objstore.SyncCbData{
				Action: action,
				Key:    []byte(key),
				Value:  []byte(val),
				Rev:    rev,
			}
			log.DebugLog(log.DebugLevelEtcd, "watch data", "key", key, "val", val, "rev", rev)
			watch.cb(&data)
		}
	}
}
