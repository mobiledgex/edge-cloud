package main

import (
	"context"
	"fmt"

	"strconv"
	"strings"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	//"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vmspec"
)

type ResTagTableApi struct {
	sync  *Sync
	store edgeproto.ResTagTableStore
	cache edgeproto.ResTagTableCache
}

var resTagTableApi = ResTagTableApi{}
var verbose bool = false

func InitResTagTableApi(sync *Sync) {
	resTagTableApi.sync = sync
	resTagTableApi.store = edgeproto.NewResTagTableStore(sync.store)
	edgeproto.InitResTagTableCache(&resTagTableApi.cache)
	sync.RegisterCache(&resTagTableApi.cache)
}

func (s *ResTagTableApi) ValidateResName(in string) (error, bool) {

	// check if the given name is one of our resource enum values
	if _, ok := edgeproto.OptResNames_value[(strings.ToUpper(in))]; !ok {
		var valids []string
		for k, _ := range edgeproto.OptResNames_value {
			valids = append(valids, strings.ToLower(k))
		}
		return fmt.Errorf("Invalid resource name %s found, must be one of %s ", in, valids), false
	}
	return nil, true
}

func (s *ResTagTableApi) CreateResTagTable(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {

	if err := in.Validate(edgeproto.ResTagTableAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}
	in.Key.Name = strings.ToLower(in.Key.Name)
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}
	return &edgeproto.Result{}, err
}

func (s *ResTagTableApi) DeleteResTagTable(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.NotFoundError()
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *ResTagTableApi) GetResTagTable(ctx context.Context, in *edgeproto.ResTagTableKey) (*edgeproto.ResTagTable, error) {
	var tbl edgeproto.ResTagTable
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, in, &tbl) {
			return in.NotFoundError()
		}
		return nil
	})
	return &tbl, err
}

func (s *ResTagTableApi) ShowResTagTable(in *edgeproto.ResTagTable, cb edgeproto.ResTagTableApi_ShowResTagTableServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.ResTagTable) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

// Update misc data, so far the availability zone for any of the optional resources needed.
func (s *ResTagTableApi) UpdateResTagTable(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.ResTagTable
	var err error

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &tbl) {
			return in.Key.NotFoundError()
		}
		tbl.CopyInFields(in)
		s.store.STMPut(stm, &tbl)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *ResTagTableApi) AddResTag(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.ResTagTable

	var err error

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &tbl) {
			return in.Key.NotFoundError()
		}
		if tbl.Tags == nil {
			tbl.Tags = make(map[string]string)
		}
		for k, t := range in.Tags {
			tbl.Tags[k] = t
		}
		s.store.STMPut(stm, &tbl)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *ResTagTableApi) RemoveResTag(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.ResTagTable
	var err error
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &tbl) {
			return in.Key.NotFoundError()
		}
		for k, _ := range in.Tags {
			delete(tbl.Tags, k)
		}
		s.store.STMPut(stm, &tbl)
		return nil
	})

	return &edgeproto.Result{}, err

}

func (s *ResTagTableApi) UsesGpu(ctx context.Context, stm concurrency.STM, flavor edgeproto.FlavorInfo, cl *edgeproto.Cloudlet) bool {

	tbls, _ := s.GetResTablesForCloudlet(ctx, stm, cl)
	resources, rescnt := vmspec.OSFlavorResources(ctx, flavor, tbls)
	if rescnt > 0 {
		if _, ok := resources["gpu"]; ok {
			return true
		}
	}
	return false
}

// GetVMSpec returns the VMCreationAttributes including flavor name and the size of the external volume which is required, if any
func (s *ResTagTableApi) GetVMSpec(ctx context.Context, stm concurrency.STM, nodeflavor edgeproto.Flavor, cl edgeproto.Cloudlet, cli edgeproto.CloudletInfo) (*vmspec.VMCreationSpec, error) {

	tbls, _ := s.GetResTablesForCloudlet(ctx, stm, &cl)
	return vmspec.GetVMSpec(ctx, nodeflavor, cli, tbls)
}

func (s *ResTagTableApi) GetResTablesForCloudlet(ctx context.Context, stm concurrency.STM, cl *edgeproto.Cloudlet) (tables map[string]*edgeproto.ResTagTable, err error) {
	if cl.ResTagMap == nil {
		return nil, fmt.Errorf("Cloudlet %s requests no optional resources", cl.Key.Name)
	}

	tabs := make(map[string]*edgeproto.ResTagTable)
	for k, v := range cl.ResTagMap {
		t := edgeproto.ResTagTable{}
		if resTagTableApi.store.STMGet(stm, v, &t) {
			tabs[k] = &t
		}
	}
	return tabs, nil
}

// Validate CLI input for any Optional Resource Map entries provided with CreateFlavor.
// Any validation of the manditory resource values will be found in flavor_api.go.

func (s *ResTagTableApi) ValidateOptResMapValues(resmap map[string]string) (bool, error) {
	// Currently only gpu resources are supported, but this routine is easily
	// extended to include those, TBI.
	//
	// For GPU resources, when creating a mex flavor, you can specify requests of the form:
	// 1) optresmap=gpu=gpu:N
	// 2) optresmap=gpu=vgpu:N or
	// 3) optresmap=gpu=pci:N
	// 4) optresmap=vgpu:nvidia-63:N
	// 5) optresmap=pci:T4:N
	//
	// Where:
	// 1) indicates we don't care how the resourse is provided, and the first matching os flavor will be used.
	// All other specifiers are optional, and increase specificity of the request.
	//
	// 2) Requests a vGPU resource, of any kind.
	// 3) Requests a dedicated PCI passthru GPU, of any kind.
	//    4 and 5 allow specific types of resource instances and are also optional.
	// 4) optresmap=gpu=vgpu:nvidia-63:1   = specific vgpu type, 1 instance.
	// 5) optresmap=gpu=pci:T4:2           = specific pci passthru, 2 instances.
	//
	// In all cases, a numeric count value is used to map to os flavors that supply > 1 of the given
	// resource. Only flavors that advertise a count >= to that requested should match.
	var err error
	var count string
	for k, v := range resmap {
		if k == "gpu" {
			values := strings.Split(v, ":")
			if len(values) == 1 {
				return false, fmt.Errorf("Missing manditory resource count, ex: optresmap=gpu=gpu:1")
			}
			if values[0] != "pci" && values[0] != "vgpu" && values[0] != "gpu" {
				return false, fmt.Errorf("GPU resource type selector must be one of [gpu, pci, vgpu] found %s", values[0])
			}
			if len(values) == 2 {
				count = values[1]
			} else if len(values) == 3 {
				count = values[2]
			} else {
				return false, fmt.Errorf("Invalid optresmap syntax encountered: ex: optresmap=gpu=gpu:1")
			}
			if _, err = strconv.Atoi(count); err != nil {
				return false, fmt.Errorf("Non-numeric resource count encountered, found %s", values[1])
			}

		} else {
			// if k == "nas" etc
			return false, fmt.Errorf("Only GPU resources currently supported, use optresmap=gpu=$resource:$count found %s", k)
		}
	}
	return true, nil
}
