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

func (s *ResTagTableApi) validateMultiTagInput(in *edgeproto.ResTagTable) error {
	if len(in.Tags) > 1 {
		for i, ctag := range in.Tags {
			if i == len(in.Tags)-1 {
				break
			}
			for _, ttag := range in.Tags[i+1 : len(in.Tags)] {
				if ctag == ttag {
					return fmt.Errorf("Duplicate Tag Found %s in multi-tag input", ctag)
				}
			}
		}
	}
	return nil
}

func (s *ResTagTableApi) AddResTag(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.ResTagTable

	// validate input, don't accept dup tag values in any  multi-tag input
	err := s.validateMultiTagInput(in)
	if err != nil {
		return &edgeproto.Result{}, err
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &tbl) {
			return in.Key.NotFoundError()
		}
		for _, t := range in.Tags {
			// Check tbl we just fetched for dups, could be an empty table
			for _, tt := range tbl.Tags {
				if t == tt {
					return fmt.Errorf("Duplicate Tag Found %s", t)
				}
			}
			tbl.Tags = append(tbl.Tags, t)
		}
		s.store.STMPut(stm, &tbl)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *ResTagTableApi) RemoveResTag(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.ResTagTable

	// validate input, don't accept dup tag values in any  multi-tag input
	err := s.validateMultiTagInput(in)
	if err != nil {
		return &edgeproto.Result{}, err
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &tbl) {
			return in.Key.NotFoundError()
		}
		for _, t := range in.Tags {
			for j, tag := range tbl.Tags {

				if t == tag {
					tbl.Tags[j] = tbl.Tags[len(tbl.Tags)-1]
					tbl.Tags[len(tbl.Tags)-1] = ""
					tbl.Tags = tbl.Tags[:len(tbl.Tags)-1]
				}
			}
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
	// non-nominal corner case: Cloudlet has no resource map at all, node flavor asks for a resource
	// so only hints found in the flavor name itself can be used, currently only resource 'gpu' uses this.
	if resmap == nil {
		// handle any flavor name hints that may exist
		if _, ok := nodeflavor.OptResMap[strings.ToLower(edgeproto.OptResNames_name[int32(edgeproto.OptResNames_GPU)])]; ok {
			if strings.Contains(flavor.Name, "gpu") {
				az, _ := s.findAZmatch("gpu", cli)
				img, _ := s.findImagematch("gpu", cli)
				return az, img, true, nil
			}
		}
		return "", "", false, fmt.Errorf("Clouddlet has no Resource mapping tables")
	}
	// Run the extent of the resource map. If the nodeflavor requests
	// an optional resource, look into that restagtbl for hints to match
	// the given flavorInfo's properities.
	for res, tblkey := range resmap {
		resname := edgeproto.OptResNames_value[strings.ToUpper(res)]
		switch resname {

		case int32(edgeproto.OptResNames_GPU):

			var numgpus int
			var err error
			gpuval := nodeflavor.OptResMap[strings.ToLower(edgeproto.OptResNames_name[resname])]
			if numgpus, err = strconv.Atoi(gpuval); err != nil {
				err = fmt.Errorf("atoi failed for %s", gpuval)
				return "", "", false, err
			}
			if numgpus > 0 {
				if !strings.Contains(flavor.Name, "gpu") {
					tbl, err := s.GetCloudletResourceMap(tblkey)
					if err != nil || tbl == nil {
						// gpu requested, name didn't match and
						// no gpu table, this osFlavor fails
						return "", "", false, err
					}
					for _, tag := range tbl.Tags {
						// PropMap impl:
						for k, v := range flavor.PropMap {
							// continue on first match
							// effectively the same as orig str impl. Next, we'll offer
							// search by property name functionality where tag values become
							// more interesting, and match on either key or value
							if strings.Contains(tag, k) || strings.Contains(tag, v) {
								continue
							}
						}
					}
				}
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
