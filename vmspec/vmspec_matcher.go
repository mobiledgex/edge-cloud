package vmspec

import (
	"fmt"
	"sort"
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

// GetVMSpec returns the VMCreationAttributes including flavor name and the size of the external volume which is required, if any
func GetVMSpec(flavorList []*edgeproto.FlavorInfo, nodeflavor edgeproto.Flavor) (*VMCreationSpec, error) {
	log.InfoLog("GetVMSpec with closest flavor available", "flavorList", flavorList, "nodeflavor", nodeflavor)
	var vmspec VMCreationSpec

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

		if strings.Contains(flavor.Name, "gpu") {
			continue
		}

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
		vmspec.FlavorName = flavor.Name
		log.InfoLog("Found closest flavor", "flavor", flavor, "vmspec", vmspec)
		return &vmspec, nil
	}
	return &vmspec, fmt.Errorf("no suitable platform flavor found for %s, please try a smaller flavor", nodeflavor.Key.Name)
}
