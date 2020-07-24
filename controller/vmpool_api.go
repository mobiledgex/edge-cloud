package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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
	if err := in.Validate(nil); err != nil {
		return &edgeproto.Result{}, err
	}

	// Let cloudlet update the pool, if the pool is in use by Cloudlet
	if cloudletApi.UsesVMPool(&in.Key) {
		updateVMs := make(map[string]edgeproto.VM)
		for _, vm := range in.Vms {
			vm.State = edgeproto.VMState_VM_UPDATE
			updateVMs[vm.Name] = vm
		}
		err = s.updateVMPoolInternal(ctx, &in.Key, updateVMs)
	} else {
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
	}
	return &edgeproto.Result{}, err
}

func (s *VMPoolApi) DeleteVMPool(ctx context.Context, in *edgeproto.VMPool) (*edgeproto.Result, error) {
	// Validate if pool is in use by Cloudlet
	if cloudletApi.UsesVMPool(&in.Key) {
		return &edgeproto.Result{}, fmt.Errorf("VMPool in use by Cloudlet")
	}
	return s.store.Delete(ctx, in, s.sync.syncWait)
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

	var err error
	// Let cloudlet update the pool, if the pool is in use by Cloudlet
	if cloudletApi.UsesVMPool(&in.Key) {
		updateVMs := make(map[string]edgeproto.VM)
		updateVMs[in.Vm.Name] = in.Vm
		in.Vm.State = edgeproto.VMState_VM_ADD
		err = s.updateVMPoolInternal(ctx, &in.Key, updateVMs)
	} else {
		err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
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

	}
	return &edgeproto.Result{}, err
}

func (s *VMPoolApi) RemoveVMPoolMember(ctx context.Context, in *edgeproto.VMPoolMember) (*edgeproto.Result, error) {
	var err error
	// Let cloudlet update the pool, if the pool is in use by Cloudlet
	if cloudletApi.UsesVMPool(&in.Key) {
		updateVMs := make(map[string]edgeproto.VM)
		updateVMs[in.Vm.Name] = in.Vm
		in.Vm.State = edgeproto.VMState_VM_REMOVE
		err = s.updateVMPoolInternal(ctx, &in.Key, updateVMs)
	} else {
		err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			cur := edgeproto.VMPool{}
			if !s.store.STMGet(stm, &in.Key, &cur) {
				return in.Key.NotFoundError()
			}
			changed := false
			for ii, vm := range cur.Vms {
				if vm.Name == in.Vm.Name {
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

	}

	return &edgeproto.Result{}, err
}

// Transition states indicate states in which the CRM is still busy.
var UpdateVMPoolTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_UPDATING: struct{}{},
}

func (s *VMPoolApi) updateVMPoolInternal(ctx context.Context, key *edgeproto.VMPoolKey, vms map[string]edgeproto.VM) error {
	if len(vms) == 0 {
		return fmt.Errorf("no VMs specified")
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.VMPool{}
		if !s.store.STMGet(stm, key, &cur) {
			return key.NotFoundError()
		}
		existingVMs := make(map[string]struct{})
		for ii, vm := range cur.Vms {
			existingVMs[vm.Name] = struct{}{}
			updateVM, ok := vms[vm.Name]
			if !ok {
				continue
			}
			if vm.State == edgeproto.VMState_VM_ADD {
				return fmt.Errorf("VM %s already exists as part of VM Pool", vm.Name)
			}
			cur.Vms[ii] = updateVM
		}
		for vmName, vm := range vms {
			if _, ok := existingVMs[vmName]; !ok {
				if vm.State == edgeproto.VMState_VM_UPDATE || vm.State == edgeproto.VMState_VM_REMOVE {
					return fmt.Errorf("VM %s does not exist in the pool", vmName)
				}
			}
		}
		cur.State = edgeproto.TrackedState_UPDATE_REQUESTED
		s.store.STMPut(stm, &cur)
		return nil
	})
	err = vmPoolApi.cache.WaitForState(ctx, key, edgeproto.TrackedState_READY, UpdateVMPoolTransitions, edgeproto.TrackedState_UPDATE_ERROR, settingsApi.Get().UpdateVmPoolTimeout.TimeDuration(), "Updated VM Pool Successfully", nil)
	if err != nil {
		// Update Failed
		return err
	}
	return err
}

func (s *VMPoolApi) UpdateFromInfo(ctx context.Context, in *edgeproto.VMPoolInfo) {
	log.DebugLog(log.DebugLevelApi, "Update VMPool from info", "info", in)
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		vmPool := edgeproto.VMPool{}
		if !s.store.STMGet(stm, &in.Key, &vmPool) {
			// got deleted in the meantime
			return nil
		}
		vmPool.Vms = []edgeproto.VM{}
		for _, infoVM := range in.Vms {
			vmPool.Vms = append(vmPool.Vms, infoVM)
		}
		vmPool.State = in.State
		vmPool.Errors = in.Errors
		s.store.STMPut(stm, &vmPool)
		return nil
	})
}
