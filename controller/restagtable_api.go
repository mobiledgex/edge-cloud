package main

import (
	"context"
	"fmt"
	"sort"

	"strconv"
	"strings"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vmspec"
)

type ResTagTableApi struct {
	sync  *Sync
	store edgeproto.ResTagTableStore
	cache edgeproto.ResTagTableCache
}

var resTagTableApi = ResTagTableApi{}

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

// Routines supporting the mapping used in GetVMSpec
//
func (s *ResTagTableApi) GetCloudletResourceMap(key *edgeproto.ResTagTableKey) (*edgeproto.ResTagTable, error) {

	var ctx context.Context
	tbl, err := resTagTableApi.GetResTagTable(ctx, key)
	return tbl, err
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

func (s *ResTagTableApi) optResLookup(nodeflavor edgeproto.Flavor, flavor edgeproto.FlavorInfo, cl edgeproto.Cloudlet, cli edgeproto.CloudletInfo) (string, string, bool, error) {
	var resmap map[string]*edgeproto.ResTagTableKey = cl.ResTagMap
	var img, az string
	var wildcard bool = false
	// Run the extent of the resource map. If the nodeflavor requests
	// an optional resource, look into that restagtbl for hints to match
	// the given flavorInfo's properities.

	for res, tblkey := range resmap {

		resname := edgeproto.OptResNames_value[strings.ToUpper(res)]
		switch resname {

		case int32(edgeproto.OptResNames_GPU):
			var numgpus, numres int
			var err error
			var count string
			gpuval := nodeflavor.OptResMap[strings.ToLower(edgeproto.OptResNames_name[resname])]
			request := strings.Split(gpuval, ":")
			if len(request) == 1 {
				// should not happen with CLI validation in place
				return "", "", false, fmt.Errorf("invalid optresmap entry encountered flavor %s request %s",
					nodeflavor.Key.Name, gpuval)
			}
			if len(request) == 2 {
				// generic request for res type, no res specifier present
				wildcard = true
				count = request[1]
			} else if len(request) == 3 {
				count = request[2]
			}
			if numgpus, err = strconv.Atoi(count); err != nil {
				return "", "", false, fmt.Errorf("Non-numertic resource count encountered")
			}
			if numgpus > 0 {
				tbl, err := s.GetCloudletResourceMap(tblkey)
				if err != nil || tbl == nil {
					// gpu requested, name didn't match and
					// no gpu table, osFlavor fails
					return "", "", false, err
				}
				// check tags table for a key match
				// If found, take the value of that tag entry  and search our flavors properties map
				// for a match. If found, this is our flavor.
				for tag_key, tag_val := range tbl.Tags {
					if tag_key == request[0] {
						if len(flavor.PropMap) == 0 {
							continue
						}
						var alias []string
						for flav_key, flav_val := range flavor.PropMap {
							// How many resources are supplied by this os flavor?
							alias = strings.Split(flav_val, ":")
							if len(alias) == 2 {
								if numres, err = strconv.Atoi(alias[1]); err != nil {
									return "", "", false, fmt.Errorf("Non-numeric count found in os flavor props")
								}
							} else {
								continue
							}
							if wildcard {
								// we have just the $kind:1 as in vgpu:1
								if flav_key == request[0] && numres >= numgpus {
									goto flavor_found
								}
							} else {
								// we have resource type specifier as in $kind:$alias:N ex: vgpu:Nvidia-63:1
								if strings.Contains(tag_val, flav_val) {
									if numres >= numgpus {
										goto flavor_found
									}
								}
							}
						}
					}
				}
				return "", "", false, fmt.Errorf("no matching tag found for mex flavor  %s\n\n", nodeflavor.Key.Name)

			flavor_found:
				az, _ = s.findAZmatch("gpu", cli)
				img, _ = s.findImagematch("gpu", cli)
				return az, img, true, nil
			} else {
				return "", "", false, fmt.Errorf("No GPU resources requested")
			}
			// Other resources TBI
		case int32(edgeproto.OptResNames_NAS):
			break
		case int32(edgeproto.OptResNames_NIC):
			break
		default:
			log.InfoLog("Unhandled resource", "res", res)
		}
	}
	return "", "", false, nil

}

// GetVMSpec returns the VMCreationAttributes including flavor name and the size of the external volume which is required, if any
func (s *ResTagTableApi) GetVMSpec(nodeflavor edgeproto.Flavor, cl edgeproto.Cloudlet, cli edgeproto.CloudletInfo) (*vmspec.VMCreationSpec, error) {
	var flavorList []*edgeproto.FlavorInfo
	var vmspec vmspec.VMCreationSpec
	var az, img string

	flavorList = cli.Flavors
	log.InfoLog("GetVMSpec with closest flavor available", "flavorList", flavorList, "nodeflavor", nodeflavor)
	sort.Slice(flavorList[:], func(i, j int) bool {
		if flavorList[i].Vcpus < flavorList[j].Vcpus {
			return true
		}
		if flavorList[i].Vcpus > flavorList[j].Vcpus {
			return false
		}
		if flavorList[i].Ram < flavorList[j].Ram {
			return true
		}
		if flavorList[i].Ram > flavorList[j].Ram {
			return false
		}

		return flavorList[i].Disk < flavorList[j].Disk
	})
	for _, flavor := range flavorList {

		if flavor.Vcpus < nodeflavor.Vcpus {
			continue
		}
		if flavor.Ram < nodeflavor.Ram {
			continue
		}
		if flavor.Disk == 0 {
			// flavors of zero disk size mean that the volume is allocated separately
			vmspec.ExternalVolumeSize = nodeflavor.Disk
		} else if flavor.Disk < nodeflavor.Disk {
			continue
		}
		// Good matches for flavor so far, does nodeflavor request an
		// optional resource? If so, the flavor will have a non-nil OptResMap.
		// If any specific resource fails, the flavor is rejected.
		var ok bool
		if nodeflavor.OptResMap != nil {
			if az, img, ok, _ = resTagTableApi.optResLookup(nodeflavor, *flavor, cl, cli); !ok {
				continue
			}
		}
		vmspec.FlavorName = flavor.Name
		vmspec.AvailabilityZone = az
		vmspec.ImageName = img
		log.InfoLog("Found closest flavor", "flavor", flavor, "vmspec", vmspec)

		return &vmspec, nil
	}
	return &vmspec, fmt.Errorf("no suitable platform flavor found for %s, please try a smaller flavor", nodeflavor.Key.Name)
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
	// 4 and 5 allow specific types of resource instances and are also optional.
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
