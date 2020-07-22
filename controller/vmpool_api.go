package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type VMPoolApi struct {
	sync  *Sync
	store edgeproto.VMPoolStore
	cache edgeproto.VMPoolCache
}

var vmPoolApi = VMPoolApi{}

func InitVMPoolApi(sync *Sync) {
	vmPoolApi.sync = sync
	vmPoolApi.store = edgeproto.NewVMPoolStore(sync.store)
	edgeproto.InitVMPoolCache(&vmPoolApi.cache)
	sync.RegisterCache(&vmPoolApi.cache)
}

func (s *VMPoolApi) CreateVMPool(ctx context.Context, in *edgeproto.VMPool) (*edgeproto.Result, error) {
	if err := in.Validate(nil); err != nil {
		return &edgeproto.Result{}, err
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *VMPoolApi) UpdateVMPool(ctx context.Context, in *edgeproto.VMPool) (*edgeproto.Result, error) {
	err := in.ValidateUpdateFields()
	if err != nil {
		return &edgeproto.Result{}, err
	}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.VMPool{}
		changed := 0
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed = cur.CopyInFields(in)
		if err := cur.Validate(nil); err != nil {
			return err
		}
		if changed == 0 {
			return nil
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *VMPoolApi) DeleteVMPool(ctx context.Context, in *edgeproto.VMPool) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.VMPool{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		for _, vm := range cur.Vms {
			if vm.State == edgeproto.VMState_VM_IN_USE {
				return fmt.Errorf("Cloudlet VM %s is in-use", vm.Name)
			}
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *VMPoolApi) ShowVMPool(in *edgeproto.VMPool, cb edgeproto.VMPoolApi_ShowVMPoolServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.VMPool) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *VMPoolApi) AddVMPoolMember(ctx context.Context, in *edgeproto.VMPoolMember) (*edgeproto.Result, error) {
	if err := in.Validate(nil); err != nil {
		return &edgeproto.Result{}, err
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.VMPool{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		poolMember := edgeproto.VMPoolMember{}
		poolMember.Key = in.Key
		for ii, _ := range cur.Vms {
			if cur.Vms[ii].Name == in.Vm.Name {
				return fmt.Errorf("Cloudlet VM with same name already exists as part of Cloudlet VM Pool")
			}
		}
		cur.Vms = append(cur.Vms, in.Vm)
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *VMPoolApi) RemoveVMPoolMember(ctx context.Context, in *edgeproto.VMPoolMember) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.VMPool{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed := false
		for ii, vm := range cur.Vms {
			if vm.Name == in.Vm.Name {
				if vm.State == edgeproto.VMState_VM_IN_USE {
					return fmt.Errorf("Cloudlet VM is in use and hence cannot be removed")
				}
				cur.Vms = append(cur.Vms[:ii], cur.Vms[ii+1:]...)
				changed = true
				break
			}
		}
		if !changed {
			return nil
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}
