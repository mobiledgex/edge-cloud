package main

import (
	"errors"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

type dummyEtcd struct {
	db   map[string]string
	vers map[string]int64
}

func (e *dummyEtcd) Start() error {
	e.db = make(map[string]string)
	e.vers = make(map[string]int64)
	return nil
}

func (e *dummyEtcd) Stop() {
	e.db = nil
}

func (e *dummyEtcd) Create(key, val string) error {
	if e.db == nil {
		return edgeproto.ObjStoreErrNotInitialized
	}
	_, ok := e.db[key]
	if ok {
		return edgeproto.ObjStoreErrKeyExists
	}
	e.db[key] = val
	e.vers[key] = 1
	util.DebugLog(util.DebugLevelEtcd, "Created", "key", key, "val", val)
	return nil
}

func (e *dummyEtcd) Update(key, val string, version int64) error {
	if e.db == nil {
		return edgeproto.ObjStoreErrNotInitialized
	}
	_, ok := e.db[key]
	if !ok {
		return edgeproto.ObjStoreErrKeyNotFound
	}
	ver := e.vers[key]
	if version != edgeproto.ObjStoreUpdateVersionAny && ver != version {
		return errors.New("Invalid version")
	}

	e.db[key] = val
	e.vers[key] = ver + 1
	util.DebugLog(util.DebugLevelEtcd, "Updated", "key", key, "val", val, "ver", ver+1)
	return nil
}

func (e *dummyEtcd) Delete(key string) error {
	if e.db == nil {
		return edgeproto.ObjStoreErrNotInitialized
	}
	delete(e.db, key)
	util.DebugLog(util.DebugLevelEtcd, "Delete", "key", key)
	return nil
}

func (e *dummyEtcd) Get(key string) ([]byte, int64, error) {
	if e.db == nil {
		return nil, 0, edgeproto.ObjStoreErrNotInitialized
	}
	val, ok := e.db[key]
	if !ok {
		return nil, 0, edgeproto.ObjStoreErrKeyNotFound
	}
	ver := e.vers[key]

	util.DebugLog(util.DebugLevelEtcd, "Got", "key", key, "val", val, "ver", ver)
	return ([]byte)(val), ver, nil
}

func (e *dummyEtcd) List(key string, cb edgeproto.ListCb) error {
	if e.db == nil {
		return edgeproto.ObjStoreErrNotInitialized
	}
	for k, v := range e.db {
		if !strings.HasPrefix(k, key) {
			continue
		}
		util.DebugLog(util.DebugLevelEtcd, "List", "key", k, "val", v)
		err := cb([]byte(k), []byte(v))
		if err != nil {
			break
		}
	}
	return nil
}
