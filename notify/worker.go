package notify

import (
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
)

// Notify workers spawn go threads to deal with changes that take time,
// such that we do not hold up the notify recieve thread. At the same
// time, it consolidates multiple change events for the same key,
// such that only the latest state is processed, rather than all
// intermediate states.
// As result, changes for the same key are run in serial, while changes
// for different keys are run in parallel.

type NotifyType int32

const (
	TypeAppInst NotifyType = iota
	TypeAppInstInfo
	TypeCloudlet
	TypeFlavor
	TypeClusterInst
	TypeClusterInstInfo
)

// WorkKey uniquely identifies a key that needs work done on it.
type WorkKey struct {
	typ NotifyType
	key string
}

// WorkArgs carry the data that is passed to the notify callback.
type WorkArgs struct {
	key objstore.ObjKey
	old objstore.Obj
}

type WorkFunc func(key objstore.ObjKey, old objstore.Obj)

// WorkMgr tracks different keys that need work done.
// It spawns separate go threads to do work on different keys,
// but never spawns multiple threads for the same key.
type WorkMgr struct {
	mux sync.Mutex
	// funcs is a map of NotifyCb funcs to call for keys that need work done.
	funcs map[NotifyType]WorkFunc
	// Workers tracks go threads in progress.
	// An entry in the workers map means a go thread is running for the key.
	// The value of the map indicates if work needs to be done.
	// A value of false means the go thread is working, but no more work is
	// needed. A value of true means the go thread is working, and once its
	// done will call the func callback again since more work is needed.
	workers map[WorkKey]bool
	// args keeps track of the arguments to NotifyCb.
	// Because multiple changes may be absorbed into the hash map, we
	// always keep the oldest "old" value.
	args map[WorkKey]*WorkArgs
	// stats
	statThreadsSpawned   int64
	statThreadsCompleted int64
}

func NewWorkMgr() *WorkMgr {
	mgr := WorkMgr{}
	mgr.funcs = make(map[NotifyType]WorkFunc)
	mgr.workers = make(map[WorkKey]bool)
	mgr.args = make(map[WorkKey]*WorkArgs)
	return &mgr
}

func (s *WorkMgr) queueWork(typ NotifyType, key objstore.ObjKey, old objstore.Obj) {
	// Note: WorkKey cannot use the objstore.ObjKey directly, because
	// it's a pointer (because it's an interface). If it's a pointer,
	// the hash computation ends up with different values even if the
	// deferenced key is the same, because the pointer value is different.
	// So we have to convert the key to a string to be able to hash to the
	// the same value.
	wkey := WorkKey{
		typ: typ,
		key: key.GetKeyString(),
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	// if workerFound is true, there is already a go thread for this key
	_, workerFound := s.workers[wkey]
	// set the value to true to indicate work is queued
	s.workers[wkey] = true
	// only update the args if there is no older args already present
	if _, found := s.args[wkey]; !found {
		s.args[wkey] = &WorkArgs{
			key: key,
			old: old,
		}
	}
	if !workerFound {
		s.statThreadsSpawned++
		go s.doWork(wkey)
	}
}

func (s *WorkMgr) doWork(wkey WorkKey) {
	for {
		s.mux.Lock()
		if s.workers[wkey] == false {
			// no more work to do for this key
			delete(s.workers, wkey)
			s.statThreadsCompleted++
			s.mux.Unlock()
			return
		}
		arg := s.args[wkey]
		fn := s.funcs[wkey.typ]
		// set workers value to false to indicate we're doing work,
		// and no more work is queued.
		s.workers[wkey] = false
		// remove args that will be used
		delete(s.args, wkey)
		s.mux.Unlock()

		// do work. The current data should be looked up from the cache
		// via the passed in key.
		fn(arg.key, arg.old)
	}
}

func (s *WorkMgr) AppInstChanged(key *edgeproto.AppInstKey, old *edgeproto.AppInst) {
	s.queueWork(TypeAppInst, key, old)
}

func (s *WorkMgr) ClusterInstChanged(key *edgeproto.ClusterInstKey, old *edgeproto.ClusterInst) {
	s.queueWork(TypeClusterInst, key, old)
}

func (s *WorkMgr) SetChangedCb(typ NotifyType, fn func(key objstore.ObjKey, old objstore.Obj)) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.funcs[typ] = fn
}
