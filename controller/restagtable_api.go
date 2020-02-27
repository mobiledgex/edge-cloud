package main

import (
	"context"
	"fmt"
	"reflect"
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

// Check the match for any given request 'req' for resource 'resname' in OS flavor 'flavor'.
func (s *ResTagTableApi) match(ctx context.Context, stm concurrency.STM, resname string, req string, flavor edgeproto.FlavorInfo, cl edgeproto.Cloudlet) (bool, error) {

	var reqcnt, flavcnt int
	var err error
	var count string
	var wildcard bool = false

	if verbose {
		log.SpanLog(ctx, log.DebugLevelApi, "match", "resource", resname, "osflavor", flavor.Name)
	}

	// Get the res tag table key for this resource, if any
	tblkey := cl.ResTagMap[resname]
	if tblkey == nil {
		if verbose {
			log.SpanLog(ctx, log.DebugLevelApi, "Match fail no tbl key", "Cloudlet", cl.Key.Name, "resource", resname)
		}
		// no key = no table = no match possible
		return false, fmt.Errorf("cloudlet %s no tbl key for %s", cl.Key.Name, resname)
	}

	// fetch the res tag table
	tbl, err := s.GetCloudletResourceMap(ctx, stm, tblkey)
	if err != nil || tbl == nil {
		if verbose {
			log.SpanLog(ctx, log.DebugLevelApi, "Match fail No res tag tbl", "name", tblkey.Name, "resource", resname)
		}
		return false, fmt.Errorf("cloudlet %s no res tag tbl named %s for resource %s", cl.Key.Name, tblkey.Name, resname)
	}

	// break request into spec and count
	request := strings.Split(req, ":")
	if len(request) == 1 {
		// should not happen with CLI validation in place
		if verbose {
			log.SpanLog(ctx, log.DebugLevelApi, "Match fail bad request format", "cloudlet", cl.Key.Name, "resource", resname, "request", request)
		}
		// XXX in all cases?
		return false, fmt.Errorf("invalid optresmap request %s", request)
	}
	if len(request) == 2 {
		// generic request for res type, no res specifier present
		wildcard = true
		count = request[1]
	} else if len(request) == 3 {
		count = request[2]
	}
	if reqcnt, err = strconv.Atoi(count); err != nil {
		if verbose {
			log.SpanLog(ctx, log.DebugLevelApi, "Match fail Non-numeric resource count", "cloudlet", cl.Key.Name, "resource", resname, "request", request)
		}
		return false, fmt.Errorf("Match fail: resource count %s request %s resource %s ", count, request, resname)
	}
	if reqcnt == 0 {
		// auto convert to 1? XXX
		if verbose {
			log.SpanLog(ctx, log.DebugLevelApi, "Match fail resource request count zero for", "request", request)
		}
		return false, fmt.Errorf("No %s resource count for request %s", resname, request)
	}

	// Finally, run the available tags looking for match
	for tag_key, tag_val := range tbl.Tags {
		var alias []string
		for flav_key, flav_val := range flavor.PropMap {
			// How many resources are supplied by this os flavor?
			alias = strings.Split(flav_val, ":")
			if len(alias) == 2 {
				if flavcnt, err = strconv.Atoi(alias[1]); err != nil {
					if verbose {
						log.SpanLog(ctx, log.DebugLevelApi, "Match fail Non-numeric count found in OS", "flavor", flavor.Name, "alias", alias)
					}
					return false, fmt.Errorf("Non-numeric count found in os flavor props for %s", flavor.Name)
				}
			} else {
				if verbose {
					log.SpanLog(ctx, log.DebugLevelApi, "Match skipping", "flavor", flavor.Name, "prop key", flav_key, "val", flav_val, "len", len(alias))
				}
				continue
			}
			if wildcard {
				// we have just the $kind:1 as in gpu=gpu:1
				if strings.Contains(flav_key, tag_key) && flavcnt >= reqcnt {
					if verbose {
						log.SpanLog(ctx, log.DebugLevelApi, "Match: wildcard", "flavor", flavor.Name, "fkey", flav_key, "tkey", tag_key)
					}
					return true, nil
				}
			} else {
				if request[0] == tag_key {
					if strings.Contains(flav_key, tag_key) {
						if strings.Contains(flav_val, tag_val) && flavcnt >= reqcnt {
							if verbose {
								log.SpanLog(ctx, log.DebugLevelApi, "Match:", "flavor", flavor.Name, "fkey", flav_key, "fval", flav_val, "tval", tag_val)
							}
							return true, nil
						}
					}
				}
			}
		}
	}
	if verbose {
		log.SpanLog(ctx, log.DebugLevelApi, "Match fail: exhausted", "resource", resname, "flavor", flavor.Name)
	}
	return false, fmt.Errorf("No match found for flavor %s", flavor.Name)
}

// For all  optional resources requested by nodeflavor, check if flavor can satisfy them. We know the nominal resources requested
// by nodeflavor are satisfied by flavor already.
func (s *ResTagTableApi) resLookup(ctx context.Context, stm concurrency.STM, nodeflavor edgeproto.Flavor, flavor edgeproto.FlavorInfo, cl edgeproto.Cloudlet, cli edgeproto.CloudletInfo) (string, string, bool, error) {
	var img, az string

	nodeResources := make(map[string]struct{})
	for res, request := range nodeflavor.OptResMap {
		if verbose {
			log.SpanLog(ctx, log.DebugLevelApi, "lookup", "resource", res, "request", request, "flavor", flavor.Name)
		}

		if ok, err := s.match(ctx, stm, res, request, flavor, cl); ok {
			if verbose {
				log.SpanLog(ctx, log.DebugLevelApi, "lookup", "flavor", flavor.Name, "resource", res, "request", request)
			}
			nodeResources[res] = struct{}{}
			continue
		} else {
			if verbose {
				log.SpanLog(ctx, log.DebugLevelApi, "lookup fail", "flavor", nodeflavor.Key.Name, "resource", res, "request", request, "err", err.Error())
			}
			return "", "", false, fmt.Errorf("no matching tag found for mex flavor  %s\n\n", nodeflavor.Key.Name)
		}
	}

	flavorResources, _ := s.osFlavorResources(ctx, stm, flavor, cl)
	if !reflect.DeepEqual(nodeResources, flavorResources) {
		return "", "", false, fmt.Errorf("Flavor %s satifies request, yet provides additional resources not requested", flavor.Name)
	}
	if verbose {
		log.SpanLog(ctx, log.DebugLevelApi, "lookup+", "flavor", flavor.Name)
	}
	az, _ = s.findAZmatch("gpu", cli)
	img, _ = s.findImagematch("gpu", cli)
	return az, img, true, nil
}

// GetVMSpec returns the VMCreationAttributes including flavor name and the size of the external volume which is required, if any
func (s *ResTagTableApi) GetVMSpec(ctx context.Context, stm concurrency.STM, nodeflavor edgeproto.Flavor, cl edgeproto.Cloudlet, cli edgeproto.CloudletInfo) (*vmspec.VMCreationSpec, error) {
	var flavorList []*edgeproto.FlavorInfo
	var vmspec vmspec.VMCreationSpec
	var az, img string

	// If nodeflavor requests an optional resource, and there is no OptResMap in cl to support it, don't bother looking.
	if nodeflavor.OptResMap != nil && cl.ResTagMap == nil {
		log.SpanLog(ctx, log.DebugLevelApi, "GetVMSpec no optional resource supported", "cloudlet", cl.Key.Name, "flavor", nodeflavor.Key.Name)
		return nil, fmt.Errorf("Optional resource requested by %s , cloudlet %s supports none", nodeflavor.Key.Name, cl.Key.Name)
	}

	flavorList = cli.Flavors
	log.SpanLog(ctx, log.DebugLevelApi, "GetVMSpec with closest flavor available", "flavorList", flavorList, "nodeflavor", nodeflavor)

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
			if az, img, ok, _ = s.resLookup(ctx, stm, nodeflavor, *flavor, cl, cli); !ok {
				continue
			}
		} else {
			// Finally, if the os flavor we're about to return happens to be offering an optional resource
			// that was not requested, we need to skip it.
			if _, cnt := s.osFlavorResources(ctx, stm, *flavor, cl); cnt != 0 {
				continue
			}
		}
		vmspec.FlavorName = flavor.Name
		vmspec.AvailabilityZone = az
		vmspec.ImageName = img
		log.SpanLog(ctx, log.DebugLevelApi, "Found closest flavor", "flavor", flavor, "vmspec", vmspec)

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
