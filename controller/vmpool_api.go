package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type VMPoolApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.VMPoolStore
	cache edgeproto.VMPoolCache
}

func NewVMPoolApi(sync *Sync, all *AllApis) *VMPoolApi {
	vmPoolApi := VMPoolApi{}
	vmPoolApi.all = all
	vmPoolApi.sync = sync
	vmPoolApi.store = edgeproto.NewVMPoolStore(sync.store)
	edgeproto.InitVMPoolCache(&vmPoolApi.cache)
	sync.RegisterCache(&vmPoolApi.cache)
	return &vmPoolApi
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

	cctx := DefCallContext()
	cctx.SetOverride(&in.CrmOverride)

	// Let cloudlet update the pool, if the pool is in use by Cloudlet
	if s.all.cloudletApi.UsesVMPool(&in.Key) && !ignoreCRM(cctx) {
		updateVMs := make(map[string]edgeproto.VM)
		for _, vm := range in.Vms {
			if vm.State != edgeproto.VMState_VM_FORCE_FREE {
				vm.State = edgeproto.VMState_VM_UPDATE
			}
			updateVMs[vm.Name] = vm
		}
		err = s.updateVMPoolInternal(cctx, ctx, &in.Key, updateVMs)
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
			for ii, vm := range cur.Vms {
				if vm.State == edgeproto.VMState_VM_FORCE_FREE {
					cur.Vms[ii].State = edgeproto.VMState_VM_FREE
					changed += 1
				}
			}
			if changed == 0 {
				return nil
			}
			cur.State = edgeproto.TrackedState_READY
			s.store.STMPut(stm, &cur)
			return nil
		})
	}
	return &edgeproto.Result{}, err
}

func (s *VMPoolApi) DeleteVMPool(ctx context.Context, in *edgeproto.VMPool) (*edgeproto.Result, error) {
	if err := in.Key.ValidateKey(); err != nil {
		return &edgeproto.Result{}, err
	}
	// Validate if pool is in use by Cloudlet
	if s.all.cloudletApi.UsesVMPool(&in.Key) {
		return &edgeproto.Result{}, fmt.Errorf("VM pool in use by Cloudlet")
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.NotFoundError()
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

	cctx := DefCallContext()
	cctx.SetOverride(&in.CrmOverride)

	var err error
	// Let cloudlet update the pool, if the pool is in use by Cloudlet
	if s.all.cloudletApi.UsesVMPool(&in.Key) && !ignoreCRM(cctx) {
		updateVMs := make(map[string]edgeproto.VM)
		in.Vm.State = edgeproto.VMState_VM_ADD
		updateVMs[in.Vm.Name] = in.Vm
		err = s.updateVMPoolInternal(cctx, ctx, &in.Key, updateVMs)
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
					return fmt.Errorf("VM with same name already exists as part of VM pool")
				}
				err := validateVMNetInfo(&cur.Vms[ii], &in.Vm)
				if err != nil {
					return err
				}
			}
			cur.Vms = append(cur.Vms, in.Vm)
			cur.State = edgeproto.TrackedState_READY
			s.store.STMPut(stm, &cur)
			return nil
		})
	}
	return &edgeproto.Result{}, err
}

func (s *VMPoolApi) RemoveVMPoolMember(ctx context.Context, in *edgeproto.VMPoolMember) (*edgeproto.Result, error) {
	if err := in.Key.ValidateKey(); err != nil {
		return &edgeproto.Result{}, err
	}
	if err := in.Vm.ValidateName(); err != nil {
		return &edgeproto.Result{}, err
	}

	var err error

	cctx := DefCallContext()
	cctx.SetOverride(&in.CrmOverride)

	// Let cloudlet update the pool, if the pool is in use by Cloudlet
	if s.all.cloudletApi.UsesVMPool(&in.Key) && !ignoreCRM(cctx) {
		updateVMs := make(map[string]edgeproto.VM)
		in.Vm.State = edgeproto.VMState_VM_REMOVE
		updateVMs[in.Vm.Name] = in.Vm
		err = s.updateVMPoolInternal(cctx, ctx, &in.Key, updateVMs)
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
			cur.State = edgeproto.TrackedState_READY
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

func validateVMNetInfo(vmCur, vmNew *edgeproto.VM) error {
	if vmCur.NetInfo.ExternalIp != "" {
		if vmCur.NetInfo.ExternalIp == vmNew.NetInfo.ExternalIp {
			return fmt.Errorf("VM with same external IP already exists as part of VM pool")
		}
	}
	if vmCur.NetInfo.InternalIp != "" {
		if vmCur.NetInfo.InternalIp == vmNew.NetInfo.InternalIp {
			return fmt.Errorf("VM with same internal IP already exists as part of VM pool")
		}
	}
	return nil
}

func (s *VMPoolApi) updateVMPoolInternal(cctx *CallContext, ctx context.Context, key *edgeproto.VMPoolKey, vms map[string]edgeproto.VM) error {
	if len(vms) == 0 {
		return fmt.Errorf("no VMs specified")
	}
	log.SpanLog(ctx, log.DebugLevelApi, "UpdateVMPoolInternal", "key", key, "vms", vms)

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.VMPool{}
		if !s.store.STMGet(stm, key, &cur) {
			return key.NotFoundError()
		}
		if cur.State == edgeproto.TrackedState_UPDATE_REQUESTED && !ignoreTransient(cctx, cur.State) {
			return fmt.Errorf("Action already in progress, please try again later")
		}
		for ii, vm := range cur.Vms {
			for updateVMName, updateVM := range vms {
				if vm.Name == updateVMName {
					if updateVM.State == edgeproto.VMState_VM_ADD {
						return fmt.Errorf("VM %s already exists as part of VM pool", vm.Name)
					}
				} else {
					if updateVM.State == edgeproto.VMState_VM_ADD || updateVM.State == edgeproto.VMState_VM_UPDATE {
						err := validateVMNetInfo(&vm, &updateVM)
						if err != nil {
							return err
						}
					}
				}
			}
			updateVM, ok := vms[vm.Name]
			if !ok {
				continue
			}
			cur.Vms[ii] = updateVM
			delete(vms, vm.Name)
		}
		for vmName, vm := range vms {
			if vm.State == edgeproto.VMState_VM_REMOVE {
				return fmt.Errorf("VM %s does not exist in the pool", vmName)
			}
			cur.Vms = append(cur.Vms, vm)
		}
		log.SpanLog(ctx, log.DebugLevelApi, "Update VMPool", "newPool", cur)
		if !ignoreCRM(cctx) {
			cur.State = edgeproto.TrackedState_UPDATE_REQUESTED
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	if err != nil {
		return err
	}
	if ignoreCRM(cctx) {
		return nil
	}
	err = s.cache.WaitForState(ctx, key, edgeproto.TrackedState_READY, UpdateVMPoolTransitions, edgeproto.TrackedState_UPDATE_ERROR, s.all.settingsApi.Get().UpdateVmPoolTimeout.TimeDuration(), "Updated VM Pool Successfully", nil)
	// State state back to Unknown & Error to nil, as user is notified about the error, if any
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.VMPool{}
		if !s.store.STMGet(stm, key, &cur) {
			return key.NotFoundError()
		}
		cur.State = edgeproto.TrackedState_TRACKED_STATE_UNKNOWN
		cur.Errors = nil
		s.store.STMPut(stm, &cur)
		return nil
	})
	return err
}

func (s *VMPoolApi) UpdateFromInfo(ctx context.Context, in *edgeproto.VMPoolInfo) {
	log.SpanLog(ctx, log.DebugLevelApi, "Update VM pool from info", "info", in)
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
