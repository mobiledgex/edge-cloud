package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type CloudletVMPoolInfoApi struct {
	sync  *Sync
	store edgeproto.CloudletVMPoolInfoStore
	cache edgeproto.CloudletVMPoolInfoCache
}

var cloudletVMPoolInfoApi = CloudletVMPoolInfoApi{}

func InitCloudletVMPoolInfoApi(sync *Sync) {
	cloudletVMPoolInfoApi.sync = sync
	cloudletVMPoolInfoApi.store = edgeproto.NewCloudletVMPoolInfoStore(sync.store)
	edgeproto.InitCloudletVMPoolInfoCache(&cloudletVMPoolInfoApi.cache)
	sync.RegisterCache(&cloudletVMPoolInfoApi.cache)
}

func (s *CloudletVMPoolInfoApi) Update(ctx context.Context, in *edgeproto.CloudletVMPoolInfo, rev int64) {
	s.store.Put(ctx, in, nil)

	log.DebugLog(log.DebugLevelNotify, "Update CloudletVMPoolInfo", "in", in)
	cloudletVMPool := edgeproto.CloudletVMPool{}
	err := cloudletVMPoolApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletVMPoolApi.store.STMGet(stm, &in.Key, &cloudletVMPool) {
			log.DebugLog(log.DebugLevelNotify, "Missing CloudletVMPool", "key", in.Key)
			return nil
		}
		switch in.Action {
		case edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE:
			fallthrough
		case edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_RELEASE:
			if cloudletVMPool.Action != edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE {
				// Some action is being performed on Cloudlet VM Pool
				// TODO Wait for it be done?
				log.DebugLog(log.DebugLevelNotify, "CloudletVMPool is busy", "key", in.Key, "action", cloudletVMPool.Action)
				return nil
			}
			if in.Action == edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE {
				edgeproto.AllocateCloudletVMsFromPool(ctx, in, &cloudletVMPool)
			} else {
				edgeproto.ReleaseCloudletVMsFromPool(ctx, in, &cloudletVMPool)
			}
		case edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE:
			if cloudletVMPool.Action == edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE {
				return nil
			}
			cloudletVMPool.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE
		default:
			return fmt.Errorf("Invalid Cloudlet VM Action: %s", in.Action)
		}
		cloudletVMPoolApi.store.STMPut(stm, &cloudletVMPool)
		return nil
	})
	if err != nil {
		log.DebugLog(log.DebugLevelNotify, "CloudletVMPoolAction transition error", "err", err)
	}
}

func (s *CloudletVMPoolInfoApi) Delete(ctx context.Context, in *edgeproto.CloudletVMPoolInfo, rev int64) {
	s.store.Delete(ctx, in, nil)
}

func (s *CloudletVMPoolInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *CloudletVMPoolInfoApi) Prune(ctx context.Context, keys map[edgeproto.CloudletKey]struct{}) {
	// no-op
}

func (s *CloudletVMPoolInfoApi) ShowCloudletVMPoolInfo(in *edgeproto.CloudletVMPoolInfo, cb edgeproto.CloudletVMPoolInfoApi_ShowCloudletVMPoolInfoServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletVMPoolInfo) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
