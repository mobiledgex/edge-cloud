package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type VMPoolInfoApi struct {
	sync  *Sync
	store edgeproto.VMPoolInfoStore
	cache edgeproto.VMPoolInfoCache
}

var vmPoolInfoApi = VMPoolInfoApi{}

func InitVMPoolInfoApi(sync *Sync) {
	vmPoolInfoApi.sync = sync
	vmPoolInfoApi.store = edgeproto.NewVMPoolInfoStore(sync.store)
	edgeproto.InitVMPoolInfoCache(&vmPoolInfoApi.cache)
	sync.RegisterCache(&vmPoolInfoApi.cache)
}

func (s *VMPoolInfoApi) Update(ctx context.Context, in *edgeproto.VMPoolInfo, rev int64) {
	s.store.Put(ctx, in, nil)

	log.DebugLog(log.DebugLevelNotify, "Update VMPoolInfo", "in", in)
	vmPool := edgeproto.VMPool{}
	err := vmPoolApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !vmPoolApi.store.STMGet(stm, &in.Key, &vmPool) {
			log.DebugLog(log.DebugLevelNotify, "Missing VMPool", "key", in.Key)
			return nil
		}
		switch in.Action {
		case edgeproto.VMAction_VM_ACTION_ALLOCATE:
			fallthrough
		case edgeproto.VMAction_VM_ACTION_RELEASE:
			if vmPool.Action != edgeproto.VMAction_VM_ACTION_DONE {
				// Some action is being performed on Cloudlet VM Pool
				// TODO Wait for it be done?
				log.DebugLog(log.DebugLevelNotify, "VMPool is busy", "key", in.Key, "action", vmPool.Action)
				return nil
			}
			if in.Action == edgeproto.VMAction_VM_ACTION_ALLOCATE {
				edgeproto.AllocateVMsFromPool(ctx, in, &vmPool)
			} else {
				edgeproto.ReleaseVMsFromPool(ctx, in, &vmPool)
			}
		case edgeproto.VMAction_VM_ACTION_DONE:
			if vmPool.Action == edgeproto.VMAction_VM_ACTION_DONE {
				return nil
			}
			vmPool.Action = edgeproto.VMAction_VM_ACTION_DONE
		default:
			return fmt.Errorf("Invalid Cloudlet VM Action: %s", in.Action)
		}
		vmPoolApi.store.STMPut(stm, &vmPool)
		return nil
	})
	if err != nil {
		log.DebugLog(log.DebugLevelNotify, "VMPoolAction transition error", "err", err)
	}
}

func (s *VMPoolInfoApi) Delete(ctx context.Context, in *edgeproto.VMPoolInfo, rev int64) {
	s.store.Delete(ctx, in, nil)
}

func (s *VMPoolInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *VMPoolInfoApi) Prune(ctx context.Context, keys map[edgeproto.CloudletKey]struct{}) {
	// no-op
}

func (s *VMPoolInfoApi) ShowVMPoolInfo(in *edgeproto.VMPoolInfo, cb edgeproto.VMPoolInfoApi_ShowVMPoolInfoServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.VMPoolInfo) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
