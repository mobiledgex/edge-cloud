package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
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
	modRev   map[string]int64
	watchers map[string]*watcher
	rev      int64
	syncCb   objstore.SyncCb
	mux      util.Mutex
}

func (e *dummyEtcd) Start() error {
	e.db = make(map[string]string)
	e.vers = make(map[string]int64)
	e.modRev = make(map[string]int64)
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
		return 0, objstore.ErrKVStoreNotInitialized
	}
	_, ok := e.db[key]
	if ok {
		return 0, objstore.ErrKVStoreKeyExists
	}
	e.db[key] = val
	e.vers[key] = 1
	e.rev++
	e.modRev[key] = e.rev
	log.DebugLog(log.DebugLevelEtcd, "Created", "key", key, "val", val, "rev", e.rev)
	e.triggerWatcher(objstore.SyncUpdate, key, val, e.rev)
	return e.rev, nil
}

func (e *dummyEtcd) Update(key, val string, version int64) (int64, error) {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.db == nil {
		return 0, objstore.ErrKVStoreNotInitialized
	}
	_, ok := e.db[key]
	if !ok {
		return 0, objstore.ErrKVStoreKeyNotFound
	}
	ver := e.vers[key]
	if version != objstore.ObjStoreUpdateVersionAny && ver != version {
		return 0, errors.New("Invalid version")
	}

	e.db[key] = val
	e.vers[key] = ver + 1
	e.rev++
	e.modRev[key] = e.rev
	log.DebugLog(log.DebugLevelEtcd, "Updated", "key", key, "val", val, "ver", ver+1, "rev", e.rev)
	e.triggerWatcher(objstore.SyncUpdate, key, val, e.rev)
	return e.rev, nil
}

func (e *dummyEtcd) Put(key, val string, ops ...objstore.KVOp) (int64, error) {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.db == nil {
		return 0, objstore.ErrKVStoreNotInitialized
	}
	ver, ok := e.vers[key]
	if !ok {
		ver = 0
	}
	e.db[key] = val
	e.vers[key] = ver + 1
	e.rev++
	e.modRev[key] = e.rev
	log.DebugLog(log.DebugLevelEtcd, "Put", "key", key, "val", val, "ver", ver+1, "rev", e.rev)
	e.triggerWatcher(objstore.SyncUpdate, key, val, e.rev)
	return e.rev, nil
}

func (e *dummyEtcd) Delete(key string) (int64, error) {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.db == nil {
		return 0, objstore.ErrKVStoreNotInitialized
	}
	delete(e.db, key)
	delete(e.vers, key)
	delete(e.modRev, key)
	e.rev++
	log.DebugLog(log.DebugLevelEtcd, "Delete", "key", key, "rev", e.rev)
	e.triggerWatcher(objstore.SyncDelete, key, "", e.rev)
	return e.rev, nil
}

func (e *dummyEtcd) Get(key string) ([]byte, int64, int64, error) {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.db == nil {
		return nil, 0, 0, objstore.ErrKVStoreNotInitialized
	}
	val, ok := e.db[key]
	if !ok {
		return nil, 0, 0, objstore.ErrKVStoreKeyNotFound
	}
	ver := e.vers[key]

	log.DebugLog(log.DebugLevelEtcd, "Got", "key", key, "val", val, "ver", ver, "rev", e.rev)
	return ([]byte)(val), ver, e.modRev[key], nil
}

func (e *dummyEtcd) List(key string, cb objstore.ListCb) error {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.db == nil {
		return objstore.ErrKVStoreNotInitialized
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

func (e *dummyEtcd) Rev(key string) int64 {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.modRev[key]
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

func (e *dummyEtcd) Grant(ctx context.Context, ttl int64) (int64, error) {
	return 0, errors.New("unsupported")
}

func (e *dummyEtcd) KeepAlive(ctx context.Context, leaseID int64) error {
	return errors.New("unsupported")
}

// Based on clientv3/concurrency/stm.go
func (e *dummyEtcd) ApplySTM(apply func(concurrency.STM) error) (int64, error) {
	stm := dummySTM{client: e}
	var err error
	var rev int64 = 0
	ii := 0
	for {
		stm.reset()
		err = apply(&stm)
		if err != nil {
			break
		}
		rev, err = e.commit(&stm)
		if err == nil {
			break
		}
		ii++
		if ii > 5 {
			err = errors.New("too many iterations")
			break
		}
	}
	return rev, err
}

func (e *dummyEtcd) commit(stm *dummySTM) (int64, error) {
	// This implements etcd's SerializableSnapshot isolation model,
	// which checks for both read and write conflicts.
	e.mux.Lock()
	defer e.mux.Unlock()
	if len(stm.wset) == 0 {
		return e.rev, nil
	}

	rev := int64(math.MaxInt64 - 1)
	// check that gets have not changed
	for key, resp := range stm.rset {
		if e.modRev[key] != resp.modRev {
			fmt.Printf("rset modRev mismatch %s e.modRev %d resp.modRev %d\n",
				key, e.modRev[key], resp.modRev)
			return 0, errors.New("rset rev mismatch")
		}
		if resp.rev < rev {
			// find the lowest rev among the reads
			// all write keys need to be at this rev
			rev = resp.rev
		}
	}
	// check that no write keys are past the database revision
	// of the first get. If rset is empty, rev will be a huge
	// number so all these checks will pass.
	for key, _ := range stm.wset {
		wrev, ok := e.modRev[key]
		if !ok {
			wrev = 0
		}
		if wrev > rev {
			fmt.Printf("wset rev mismatch %s rev %d wrev %d\n",
				key, rev, wrev)
			return 0, errors.New("wset rev mismatch")
		}
	}
	// commit all changes in one revision
	e.rev++
	for key, val := range stm.wset {
		ver, ok := e.vers[key]
		if !ok {
			ver = 0
		}
		if val == "" {
			// delete
			delete(e.db, key)
			delete(e.vers, key)
			delete(e.modRev, key)
			log.DebugLog(log.DebugLevelEtcd, "Delete",
				"key", key, "rev", e.rev)
			e.triggerWatcher(objstore.SyncDelete, key, "", e.rev)
		} else {
			e.db[key] = val
			e.vers[key] = ver + 1
			e.modRev[key] = e.rev
			log.DebugLog(log.DebugLevelEtcd, "Commit", "key", key,
				"val", val, "ver", ver+1, "rev", e.rev)
			e.triggerWatcher(objstore.SyncUpdate, key, val, e.rev)
		}
	}
	return e.rev, nil
}

type dummyReadResp struct {
	val    string
	modRev int64
	rev    int64
}

type dummySTM struct {
	concurrency.STM
	client       *dummyEtcd
	rset         map[string]*dummyReadResp
	wset         map[string]string
	firstReadRev int64
}

func (d *dummySTM) reset() {
	d.rset = make(map[string]*dummyReadResp)
	d.wset = make(map[string]string)
}

func (d *dummySTM) Get(keys ...string) string {
	key := keys[0]
	if wv, ok := d.wset[key]; ok {
		return wv
	}
	byt, _, modRev, err := d.client.Get(key)
	rev := d.client.rev
	if err != nil {
		byt = make([]byte, 0)
		modRev = 0
	}
	resp := dummyReadResp{
		val:    string(byt),
		rev:    rev,
		modRev: modRev,
	}
	d.rset[key] = &resp
	return string(byt)
}

func (d *dummySTM) Put(key, val string, opts ...v3.OpOption) {
	d.wset[key] = val
}

func (d *dummySTM) Rev(key string) int64 {
	return d.client.Rev(key)
}

func (d *dummySTM) Del(key string) {
	d.wset[key] = ""
}
