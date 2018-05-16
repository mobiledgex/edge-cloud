// Start etcd

package main

import (
	"context"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
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
func (e *EtcdClient) Create(key, val string) error {
	if e.client == nil {
		return edgeproto.ObjStoreErrNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), WriteRequestTimeout)
	txn := e.client.Txn(ctx)
	txn = txn.If(clientv3.Compare(clientv3.Version(key), "=", 0))
	txn = txn.Then(clientv3.OpPut(key, val))
	resp, err := txn.Commit()
	cancel()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return edgeproto.ObjStoreErrKeyExists
	}
	util.DebugLog(util.DebugLevelEtcd, "create data", "key", key, "val", val)
	return nil
}

// update fails if key does not exist
func (e *EtcdClient) Update(key, val string, version int64) error {
	if e.client == nil {
		return edgeproto.ObjStoreErrNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), WriteRequestTimeout)
	txn := e.client.Txn(ctx)
	if version == edgeproto.ObjStoreUpdateVersionAny {
		// version 0 means it doesn't exist yet
		txn = txn.If(clientv3.Compare(clientv3.Version(key), "!=", 0))
	} else {
		txn = txn.If(clientv3.Compare(clientv3.Version(key), "=", version))
	}
	txn = txn.Then(clientv3.OpPut(key, val))
	resp, err := txn.Commit()
	cancel()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return edgeproto.ObjStoreErrKeyNotFound
	}
	util.DebugLog(util.DebugLevelEtcd, "update data", "key", key, "val", val)
	return nil
}

func (e *EtcdClient) Delete(key string) error {
	if e.client == nil {
		return edgeproto.ObjStoreErrNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), WriteRequestTimeout)
	_, err := e.client.Delete(ctx, key)
	cancel()
	if err != nil {
		return err
	}
	util.DebugLog(util.DebugLevelEtcd, "delete data", "key", key)
	return nil
}

func (e *EtcdClient) Get(key string) ([]byte, int64, error) {
	if e.client == nil {
		return nil, 0, edgeproto.ObjStoreErrNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), ReadRequestTimeout)
	resp, err := e.client.Get(ctx, key)
	cancel()
	if err != nil {
		return nil, 0, err
	}
	if len(resp.Kvs) == 0 {
		return nil, 0, edgeproto.ObjStoreErrKeyNotFound
	}
	obj := resp.Kvs[0]
	util.DebugLog(util.DebugLevelEtcd, "got data", "key", key, "val", string(obj.Value), "ver", obj.Version)
	return obj.Value, obj.Version, nil
}

// Get records that have the given key prefix
func (e *EtcdClient) List(key string, cb edgeproto.ListCb) error {
	if e.client == nil {
		return edgeproto.ObjStoreErrNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), ReadRequestTimeout)
	resp, err := e.client.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return err
	}
	for _, obj := range resp.Kvs {
		util.DebugLog(util.DebugLevelEtcd, "list data", "key", string(obj.Key), "val", string(obj.Value))
		err = cb(obj.Key, obj.Value)
		if err != nil {
			break
		}
	}
	return err
}
