// Start etcd

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/mobiledgex/edge-cloud/proto"
	"github.com/mobiledgex/edge-cloud/util"
)

type EtcdClient struct {
	client clientv3.KV
	config clientv3.Config
}

var (
	WriteRequestTimeout = 10 * time.Second
	ReadRequestTimeout  = 2 * time.Second
)

func GetEtcdClientBasic(clientIP string, clientPort uint) (proto.ObjStore, error) {
	clientUrl := fmt.Sprintf("http://%s:%d", clientIP, clientPort)
	cfg := clientv3.Config{
		Endpoints: []string{clientUrl},
	}
	return GetEtcdClient(&cfg)
}

func GetEtcdClient(cfg *clientv3.Config) (proto.ObjStore, error) {
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

// create fails if key already exists
func (e *EtcdClient) Create(key, val string) error {
	if e.client == nil {
		return proto.ObjStoreErrNotInitialized
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
		return proto.ObjStoreErrKeyExists
	}
	util.DebugLog(util.DebugLevelEtcd, "create data", "key", key, "val", val)
	return nil
}

// update fails if key does not exist
func (e *EtcdClient) Update(key, val string, version int64) error {
	if e.client == nil {
		return proto.ObjStoreErrNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), WriteRequestTimeout)
	txn := e.client.Txn(ctx)
	if version == proto.ObjStoreUpdateVersionAny {
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
		return proto.ObjStoreErrKeyNotFound
	}
	util.DebugLog(util.DebugLevelEtcd, "update data", "key", key, "val", val)
	return nil
}

func (e *EtcdClient) Delete(key string) error {
	if e.client == nil {
		return proto.ObjStoreErrNotInitialized
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
		return nil, 0, proto.ObjStoreErrNotInitialized
	}
	ctx, cancel := context.WithTimeout(context.Background(), ReadRequestTimeout)
	resp, err := e.client.Get(ctx, key)
	cancel()
	if err != nil {
		return nil, 0, err
	}
	if len(resp.Kvs) == 0 {
		return nil, 0, proto.ObjStoreErrKeyNotFound
	}
	obj := resp.Kvs[0]
	util.DebugLog(util.DebugLevelEtcd, "got data", "key", key, "val", string(obj.Value), "ver", obj.Version)
	return obj.Value, obj.Version, nil
}

// Get records that have the given key prefix
func (e *EtcdClient) List(key string, cb proto.ListCb) error {
	if e.client == nil {
		return proto.ObjStoreErrNotInitialized
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
