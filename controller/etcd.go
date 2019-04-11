// Start etcd

package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type EtcdClient struct {
	client *clientv3.Client
	config clientv3.Config
}

var (
	WriteRequestTimeout = 10 * time.Second
	ReadRequestTimeout  = 2 * time.Second
)

func GetEtcdClientBasic(clientUrls string) (*EtcdClient, error) {
	endpoints := strings.Split(clientUrls, ",")
	cfg := clientv3.Config{
		Endpoints: endpoints,
	}
	return GetEtcdClient(&cfg)
}

func GetEtcdClient(cfg *clientv3.Config) (*EtcdClient, error) {
	client, err := clientv3.New(*cfg)
	if err != nil {
		return nil, err
	}
	etcdClient := EtcdClient{
		client: client,
		config: *cfg,
	}
	return &etcdClient, nil
}

// Do a member list call to see if we're connected
func (e *EtcdClient) CheckConnected(tries int, retryTime time.Duration) error {
	var err error
	for ii := 0; ii < tries; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), WriteRequestTimeout)
		_, err = e.client.MemberList(ctx)
		cancel()
		if err == nil {
			return nil
		}
	}
	return err
}

// create fails if key already exists
func (e *EtcdClient) Create(key, val string) (int64, error) {
	if e.client == nil {
		return 0, objstore.ErrKVStoreNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), WriteRequestTimeout)
	txn := e.client.Txn(ctx)
	txn = txn.If(clientv3.Compare(clientv3.Version(key), "=", 0))
	txn = txn.Then(clientv3.OpPut(key, val))
	resp, err := txn.Commit()
	cancel()
	if err != nil {
		return 0, err
	}
	if !resp.Succeeded {
		return 0, objstore.ErrKVStoreKeyExists
	}
	log.DebugLog(log.DebugLevelEtcd, "created data", "key", key, "val", val, "rev", resp.Header.Revision)
	return resp.Header.Revision, nil
}

// update fails if key does not exist
func (e *EtcdClient) Update(key, val string, version int64) (int64, error) {
	if e.client == nil {
		return 0, objstore.ErrKVStoreNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), WriteRequestTimeout)
	txn := e.client.Txn(ctx)
	if version == objstore.ObjStoreUpdateVersionAny {
		// version 0 means it doesn't exist yet
		txn = txn.If(clientv3.Compare(clientv3.Version(key), "!=", 0))
	} else {
		txn = txn.If(clientv3.Compare(clientv3.Version(key), "=", version))
	}
	txn = txn.Then(clientv3.OpPut(key, val))
	resp, err := txn.Commit()
	cancel()
	if err != nil {
		return 0, err
	}
	if !resp.Succeeded {
		return 0, objstore.ErrKVStoreKeyNotFound
	}
	log.DebugLog(log.DebugLevelEtcd, "updated data", "key", key, "val", val, "rev", resp.Header.Revision)
	return resp.Header.Revision, nil
}

func (e *EtcdClient) Delete(key string) (int64, error) {
	if e.client == nil {
		return 0, objstore.ErrKVStoreNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), WriteRequestTimeout)
	resp, err := e.client.Delete(ctx, key)
	cancel()
	if err != nil {
		return 0, err
	}
	if resp.Deleted == 0 {
		return 0, objstore.ErrKVStoreKeyNotFound
	}
	log.DebugLog(log.DebugLevelEtcd, "deleted data", "key", key, "rev", resp.Header.Revision)
	return resp.Header.Revision, nil
}

func (e *EtcdClient) Get(key string) ([]byte, int64, int64, error) {
	if e.client == nil {
		return nil, 0, 0, objstore.ErrKVStoreNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), ReadRequestTimeout)
	resp, err := e.client.Get(ctx, key)
	cancel()
	if err != nil {
		return nil, 0, 0, err
	}
	if len(resp.Kvs) == 0 {
		return nil, 0, 0, objstore.ErrKVStoreKeyNotFound
	}
	obj := resp.Kvs[0]
	log.DebugLog(log.DebugLevelEtcd, "got data", "key", key, "val", string(obj.Value), "ver", obj.Version, "rev", resp.Header.Revision, "create", obj.CreateRevision, "mod", obj.ModRevision, "ver", obj.Version)
	return obj.Value, obj.Version, obj.ModRevision, nil
}

func (e *EtcdClient) Put(key, val string, ops ...objstore.KVOp) (int64, error) {
	if e.client == nil {
		return 0, objstore.ErrKVStoreNotInitialized
	}
	opts := objstore.GetKVOptions(ops)
	etcdOps := make([]clientv3.OpOption, 0)
	if opts.LeaseID != 0 {
		etcdOps = append(etcdOps, clientv3.WithLease(clientv3.LeaseID(opts.LeaseID)))
	}
	ctx, cancel := context.WithTimeout(context.Background(), WriteRequestTimeout)
	resp, err := e.client.Put(ctx, key, val, etcdOps...)
	cancel()
	if err != nil {
		return 0, err
	}
	log.DebugLog(log.DebugLevelEtcd, "put data", "key", key, "val", val, "rev",
		resp.Header.Revision)
	return resp.Header.Revision, nil
}

// Get records that have the given key prefix
func (e *EtcdClient) List(key string, cb objstore.ListCb) error {
	if e.client == nil {
		return objstore.ErrKVStoreNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), ReadRequestTimeout)
	resp, err := e.client.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return err
	}
	for _, obj := range resp.Kvs {
		log.DebugLog(log.DebugLevelEtcd, "list data", "key", string(obj.Key), "val", string(obj.Value), "rev", resp.Header.Revision, "create", obj.CreateRevision, "mod", obj.ModRevision, "ver", obj.Version)
		err = cb(obj.Key, obj.Value, resp.Header.Revision)
		if err != nil {
			break
		}
	}
	return err
}

// Sync is used to sync a cache with the etcd database.
// This is needed by the controllers to sync with each other.
// This could also be used by DMEs and CRMs to sync with the etcd database
// directly, rather than via the controller.
// The pros of syncing directly:
// 1. Etcd has built in failover given multiple etcd db (or proxy) endpoints.
// It can failover if one disappears or is manually removed.
// 2. Etcd has a history of changes, so that on reconnect, we do not have
// resend all records (assuming the history has not been compacted).
// Given the possible large distance between controllers/etcd and dme/crm
// instances, random disconnects may be somewhat common. Not having to send
// all records on reconnect will help a lot.
// 3. The controllers need to watch anyway to keep in sync with each other,
// so we already need to write this for the controller. Writing another
// algorithm to sync data from controller to dme/crm somewhat duplicates this code.
// The cons of syncing directly:
// 1. The DMEs/CRMs can't actually connect to etcd directly, as it does not scale.
// We need to run an etcd grpc-proxy on each controller that somewhat duplicates
// what the controller is doing in terms of caching data and handling queries.
// 2. The grpc-proxy is not a true cache. It only caches some results.
// 3. The controller becomes dependent on etcd for communicating with
// DMEs and CRMs
// 4. The grpc-proxy is a thin-ish client. If there is no cache hit, it forwards
// consolidated requests to the etcd server. So there is a tradeoff between
// latency vs # of requests to Etcd (if no cache hit). It appears that
// the cache defaults to 2048 entries but can be adjusted.
// 5. The grpc-proxy broadcasts replies to clients in series. So if 100 clients
// were watching on the proxy for the same data, when the upstream data changes,
// that data would be sent back to each one in serial. So it may not be as
// performant as our own cache.
func (e *EtcdClient) Sync(ctx context.Context, key string, cb objstore.SyncCb) error {
	refresh := true
	done := false
	// we keep track of the revision so that we don't miss any changes
	// between a refresh and a subsequent watch.
	watchRev := int64(1)

	var err error = nil
	data := objstore.SyncCbData{}
	for !done {
		if refresh {
			data.Action = objstore.SyncListStart
			data.Key = nil
			data.Value = nil
			data.Rev = 0
			cb(&data)

			data.Action = objstore.SyncList
			err = e.List(key, func(key, val []byte, rev int64) error {
				data.Key = key
				data.Value = val
				data.Rev = rev
				log.DebugLog(log.DebugLevelEtcd, "sync list data", "key", string(key), "val", string(val), "rev", rev)
				cb(&data)
				watchRev = rev
				return nil
			})
			data.Action = objstore.SyncListEnd
			data.Key = nil
			data.Value = nil
			cb(&data)
			refresh = false
		}
		err = nil
		ch := e.client.Watch(ctx, key, clientv3.WithPrefix(), clientv3.WithRev(watchRev))
		for {
			resp, ok := <-ch
			if !ok {
				// channel closed
				done = true
				break
			}
			if resp.Err() != nil {
				err = resp.Err()
				break
			}
			for ii, event := range resp.Events {
				if event.Type == mvccpb.PUT {
					data.Action = objstore.SyncUpdate
				} else {
					data.Action = objstore.SyncDelete
				}
				data.Key = event.Kv.Key
				data.Value = event.Kv.Value
				data.Rev = resp.Header.Revision
				watchRev = resp.Header.Revision
				if ii == len(resp.Events)-1 {
					data.MoreEvents = false
				} else {
					data.MoreEvents = true
				}
				log.DebugLog(log.DebugLevelEtcd, "watch data", "key", string(data.Key), "val", string(data.Value), "rev", data.Rev, "more-events", data.MoreEvents)
				cb(&data)
			}
		}
		if err == rpctypes.ErrCompacted {
			// history does not exist. Grab all the keys
			// regardless of revision and then run the
			// watch again on the latest revision
			refresh = true
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (e *EtcdClient) ApplySTM(apply func(concurrency.STM) error) (int64, error) {
	resp, err := concurrency.NewSTM(e.client, apply)
	if err != nil {
		return 0, err
	}
	return resp.Header.Revision, nil
}

func (e *EtcdClient) Grant(ctx context.Context, ttl int64) (int64, error) {
	resp, err := e.client.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}
	if resp.Error != "" {
		return 0, errors.New(resp.Error)
	}
	return int64(resp.ID), nil
}

func (e *EtcdClient) KeepAlive(ctx context.Context, leaseID int64) error {
	for {
		ch, err := e.client.KeepAlive(ctx, clientv3.LeaseID(leaseID))
		if err != nil {
			return err
		}
		closed := false
		for {
			select {
			case <-ctx.Done():
				return nil
			case _, ok := <-ch:
				if !ok {
					closed = true
				}
			}
			if closed {
				break
			}
		}
		if closed {
			break
		}
	}
	return nil
}

func (e *EtcdClient) Version() (string, error) {
	key := objstore.DbKeyPrefixString("Version")
	val, _, _, err := e.Get(key)
	if err != nil {
		if strings.Contains(err.Error(), objstore.ErrKVStoreKeyNotFound.Error()) {
			return "", nil
		}
		return "", err
	}
	return string(val), nil
}
