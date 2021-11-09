package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vmspec"
)

type ResTagTableApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.ResTagTableStore
	cache edgeproto.ResTagTableCache
}

var verbose bool = false

func NewResTagTableApi(sync *Sync, all *AllApis) *ResTagTableApi {
	resTagTableApi := ResTagTableApi{}
	resTagTableApi.all = all
	resTagTableApi.sync = sync
	resTagTableApi.store = edgeproto.NewResTagTableStore(sync.store)
	edgeproto.InitResTagTableCache(&resTagTableApi.cache)
	sync.RegisterCache(&resTagTableApi.cache)
	return &resTagTableApi
}

func (s *ResTagTableApi) ValidateResName(ctx context.Context, in string) (error, bool) {
	// check if the given name is one of our resource enum values
	if _, ok := edgeproto.OptResNames_value[(strings.ToUpper(in))]; !ok {
		var valids []string
		for k, _ := range edgeproto.OptResNames_value {
			log.SpanLog(ctx, log.DebugLevelApi, "ValidateResName", "next valid resname", k)
			valids = append(valids, strings.ToLower(k))
		}
		return fmt.Errorf("Invalid resource name %s found, must be one of %s ", in, valids), false
	}
	return nil, true
}

func (s *ResTagTableApi) CreateResTagTable(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {

	if err := in.Validate(edgeproto.ResTagTableAllFieldsMap); err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "CreateResTagTable in.Validate failed all Fields map")
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

// Routines supporting the mapping used in GetVMSpec
//
func (s *ResTagTableApi) GetCloudletResourceMap(ctx context.Context, stm concurrency.STM, key *edgeproto.ResTagTableKey) (*edgeproto.ResTagTable, error) {

	tbl := edgeproto.ResTagTable{}
	if !s.store.STMGet(stm, key, &tbl) {
		return nil, key.NotFoundError()
	}
	return &tbl, nil
}

func (s *ResTagTableApi) findImagematch(res string, cli edgeproto.CloudletInfo) (string, bool) {
	var img *edgeproto.OSImage
	for _, img = range cli.OsImages {
		if strings.Contains(strings.ToLower(img.Name), res) {
			return img.Name, true
		}
	}
	return "", false
}

func (s *ResTagTableApi) findAZmatch(res string, cli edgeproto.CloudletInfo) (string, bool) {
	var az *edgeproto.OSAZone
	for _, az = range cli.AvailabilityZones {
		if strings.Contains(strings.ToLower(az.Name), res) {
			return az.Name, true
		}
	}
	return "", false
}

// Irrespective of any requesting mex flavor, do we think this OS flavor offers any optional resources, given the current cloudlet's mappings?
// Return count and resource type values discovered in flavor.
func (s *ResTagTableApi) osFlavorResources(ctx context.Context, stm concurrency.STM, flavor edgeproto.FlavorInfo, cl edgeproto.Cloudlet) (offered map[string]struct{}, count int) {
	var rescnt int
	resources := make(map[string]struct{})

	if len(flavor.PropMap) == 0 {
		// optional resources are defined via os flavor properties
		return resources, 0
	}
	if cl.ResTagMap == nil {
		// given cloudlet has no resource mappings currently
		log.SpanLog(ctx, log.DebugLevelApi, "No OptResMap for", "cloudlet", cl.Key.Name)
		return resources, 0
	}
	// for all optional resources configured for the given cloudlet
	for res, key := range cl.ResTagMap {
		tbl, err := s.GetCloudletResourceMap(ctx, stm, key)
		if err != nil || tbl == nil {
			if verbose {
				log.SpanLog(ctx, log.DebugLevelApi, "no tbl found", "resource", res, "cloudlet", cl.Key.Name)
			}
			continue
		}
		// look in flavor.PropMap for hints
		for _, flav_val := range flavor.PropMap {
			for _, val := range tbl.Tags {
				if strings.Contains(flav_val, val) {
					if verbose {
						log.SpanLog(ctx, log.DebugLevelApi, "match", "flavor", flavor.Name, "prop", flav_val, "val", val)
					}
					resources[res] = struct{}{}
					rescnt++
				}
			}
		}
	}
	return resources, rescnt
}

func (s *ResTagTableApi) UsesGpu(ctx context.Context, stm concurrency.STM, flavor edgeproto.FlavorInfo, cl edgeproto.Cloudlet) bool {
	resources, rescnt := s.osFlavorResources(ctx, stm, flavor, cl)
	if rescnt > 0 {
		if _, ok := resources["gpu"]; ok {
			return true
		}
	}
	return false
}

// GetVMSpec returns the VMCreationAttributes including flavor name and the size of the external volume which is required, if any
func (s *ResTagTableApi) GetVMSpec(ctx context.Context, stm concurrency.STM, nodeflavor edgeproto.Flavor, cl edgeproto.Cloudlet, cli edgeproto.CloudletInfo) (*vmspec.VMCreationSpec, error) {
	// for those platforms with no concept of a quantized set of resources (flavors)
	// return a VMCreationSpec  based on the our meta-flavor resource request.
	if len(cli.Flavors) == 0 {
		spec := vmspec.VMCreationSpec{
			FlavorName: nodeflavor.Key.Name,
			FlavorInfo: &edgeproto.FlavorInfo{
				Ram:   nodeflavor.Ram,
				Name:  nodeflavor.Key.Name,
				Disk:  nodeflavor.Disk,
				Vcpus: nodeflavor.Vcpus,
			},
		}
		log.SpanLog(ctx, log.DebugLevelApi, "GetVMSpec platform has no native flavors returning mex flavor for", "platform", cl.PlatformType, "as", spec)
		return &spec, nil
	}

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
		if s.store.STMGet(stm, v, &t) {
			tabs[k] = &t
		}
	}
	log.SpanLog(ctx, log.DebugLevelApi, "GetResTablesForCloudlet", "tbl count", len(tabs))
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
	// 4) optresmap=gpu=vgpu:nvidia-63:N
	// 5) optresmap=gpu=pci:T4:2
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
	for k, v := range resmap {
		if k == "gpu" {
			_, err = cloudcommon.ParseGPUResourceCount(v)
			if err != nil {
				return false, err
			}
		} else {
			// if k == "nas" etc
			return false, fmt.Errorf("Only GPU resources currently supported, use optresmap=gpu=$resource:[$specifier:]$count found %s", k)
		}
	}
	return true, nil
}

func (s *ResTagTableApi) AddGpuResourceHintIfNeeded(ctx context.Context, stm concurrency.STM, spec *vmspec.VMCreationSpec, cloudlet edgeproto.Cloudlet) string {

	if s.UsesGpu(ctx, stm, *spec.FlavorInfo, cloudlet) {
		log.SpanLog(ctx, log.DebugLevelApi, "add hint using gpu on", "platform", cloudlet.PlatformType, "flavor", spec.FlavorName)
		return "gpu"
	} else {
		if strings.Contains(spec.FlavorInfo.Name, "gpu") {
			log.SpanLog(ctx, log.DebugLevelApi, "add hint using gpu on", "platform", cloudlet.PlatformType, "flavor", spec.FlavorName)
			return "gpu"
		}
	}
	return ""
}
