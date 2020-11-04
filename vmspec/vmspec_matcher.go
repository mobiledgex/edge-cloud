package vmspec

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// VMCreationSpec includes the flavor and other aspects needed to instantiate a VM
type VMCreationSpec struct {
	FlavorName         string
	ExternalVolumeSize uint64
	AvailabilityZone   string
	ImageName          string
	PrivacyPolicy      *edgeproto.PrivacyPolicy
	MasterNodeFlavor   string
	FlavorInfo         *edgeproto.FlavorInfo
}

var verbose bool = false

// Routines supporting the mapping used in GetVMSpec
//

func findImagematch(res string, cli edgeproto.CloudletInfo) (string, bool) {
	var img *edgeproto.OSImage
	for _, img = range cli.OsImages {
		if strings.Contains(strings.ToLower(img.Name), res) {
			return img.Name, true
		}
	}
	return "", false
}

func findAZmatch(res string, cli edgeproto.CloudletInfo) (string, bool) {
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

func OSFlavorResources(ctx context.Context, flavor edgeproto.FlavorInfo, tbls map[string]*edgeproto.ResTagTable) (offered map[string]struct{}, count int) {
	var rescnt int
	resources := make(map[string]struct{})

	if len(flavor.PropMap) == 0 {
		// optional resources are defined via os flavor properties
		return resources, 0
	}
	// for all optional resources configured for the given cloudlet
	// tbls is like the map in cl.ResTagMap, but rather than key of target table, it's the table itself
	for res, tbl := range tbls {
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
func match(ctx context.Context, resname string, req string, flavor edgeproto.FlavorInfo, tbl *edgeproto.ResTagTable) (bool, error) {

	var reqcnt, flavcnt int
	var err error
	var count string
	var wildcard bool = false

	if verbose {
		log.SpanLog(ctx, log.DebugLevelApi, "match", "resource", resname, "osflavor", flavor.Name)
	}

	// break request into spec and count
	var request []string
	if strings.Contains(req, ":") {
		request = strings.Split(req, ":")
	} else if strings.Contains(req, "=") {
		// VIO syntax uses =
		request = strings.Split(req, "=")
	}

	if len(request) == 1 {
		// should not happen with CLI validation in place
		if verbose {
			log.SpanLog(ctx, log.DebugLevelApi, "Match fail bad request format", "resource", resname, "request", request)
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
			log.SpanLog(ctx, log.DebugLevelApi, "Match fail Non-numeric resource count", "resource", resname, "request", request)
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
			if strings.Contains(flav_val, ":") {
				alias = strings.Split(flav_val, ":")
			} else if strings.Contains(flav_val, "=") {
				// VIO syntax
				alias = strings.Split(flav_val, "=")
			}
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
						log.SpanLog(ctx, log.DebugLevelApi, "Match: wildcard match", "flavor", flavor.Name, "fkey", flav_key, "tkey", tag_key)
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
func resLookup(ctx context.Context, nodeflavor edgeproto.Flavor, flavor edgeproto.FlavorInfo, cli edgeproto.CloudletInfo, tbls map[string]*edgeproto.ResTagTable) (string, string, bool, error) {
	var img, az string

	nodeResources := make(map[string]struct{})
	for res, request := range nodeflavor.OptResMap {
		if verbose {
			log.SpanLog(ctx, log.DebugLevelApi, "lookup", "resource", res, "request", request, "flavor", flavor.Name)
		}
		tbl := tbls[res]
		if tbl == nil {
			continue
		}
		if ok, err := match(ctx, res, request, flavor, tbl); ok {
			if verbose {
				log.SpanLog(ctx, log.DebugLevelApi, "lookup matched", "flavor", flavor.Name, "resource", res, "request", request)
			}
			nodeResources[res] = struct{}{}
			continue
		} else {
			if verbose {
				log.SpanLog(ctx, log.DebugLevelApi, "lookup-I-match failed", "flavor", nodeflavor.Key.Name, "resource", res, "request", request, "err", err.Error())
			}
			return "", "", false, fmt.Errorf("no matching tag found for mex flavor  %s\n\n", nodeflavor.Key.Name)
		}
	}

	flavorResources, _ := OSFlavorResources(ctx, flavor, tbls)
	if !reflect.DeepEqual(nodeResources, flavorResources) {
		return "", "", false, fmt.Errorf("Flavor %s satifies request, yet provides additional resources not requested", flavor.Name)
	}
	if verbose {
		log.SpanLog(ctx, log.DebugLevelApi, "lookup+", "flavor", flavor.Name)
	}
	az, _ = findAZmatch("gpu", cli)
	img, _ = findImagematch("gpu", cli)
	return az, img, true, nil
}

// GetVMSpec returns the VMCreationAttributes including flavor name and the size of the external volume which is required, if any
func GetVMSpec(ctx context.Context, nodeflavor edgeproto.Flavor, cli edgeproto.CloudletInfo, tbls map[string]*edgeproto.ResTagTable) (*VMCreationSpec, error) {
	var flavorList []*edgeproto.FlavorInfo
	var vmspec VMCreationSpec
	var az, img string

	// If nodeflavor requests an optional resource, and there is no OptResMap in cl (tbls = nil) to support it, don't bother looking.
	if nodeflavor.OptResMap != nil && tbls == nil {
		log.SpanLog(ctx, log.DebugLevelApi, "GetVMSpec no optional resource supported", "cloudlet", cli.Key.Name, "flavor", nodeflavor.Key.Name)
		return nil, fmt.Errorf("Optional resource requested by %s, cloudlet %s supports none", nodeflavor.Key.Name, cli.Key.Name)
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
			if az, img, ok, _ = resLookup(ctx, nodeflavor, *flavor, cli, tbls); !ok {
				continue
			}
		} else {
			// Our mex flavor is not requesting any optional resources. (OptResMap in mex flavor = nil)
			// so to prevent _any_ race condition or absence of cloudlet config, skip any o.s. flavor with
			// "gpu" in its name.
			if strings.Contains(flavor.Name, "gpu") {
				log.SpanLog(ctx, log.DebugLevelApi, "No opt resource requested, skipping gpu ", "flavor", flavor.Name)
				continue
			}
			// Finally, if the os flavor we're about to return happens to be offering an optional resource
			// that was not requested, we need to skip it.
			if _, cnt := OSFlavorResources(ctx, *flavor, tbls); cnt != 0 {
				log.SpanLog(ctx, log.DebugLevelApi, "No opt resource requested, skipping ", "flavor", flavor.Name)
				continue
			}
		}
		vmspec.FlavorName = flavor.Name
		vmspec.AvailabilityZone = az
		vmspec.ImageName = img
		vmspec.FlavorInfo = flavor
		log.SpanLog(ctx, log.DebugLevelApi, "Found closest flavor", "flavor", flavor, "vmspec", vmspec)

		return &vmspec, nil
	}
	return &vmspec, fmt.Errorf("no suitable platform flavor found for %s, please try a smaller flavor", nodeflavor.Key.Name)
}
