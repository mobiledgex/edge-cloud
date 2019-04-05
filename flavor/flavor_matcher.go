package flavor

import (
	"fmt"
	"sort"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func GetClosestFlavor(flavorList []*edgeproto.FlavorInfo, nodeflavor edgeproto.Flavor) (string, error) {
	log.InfoLog("Get closest flavor available", "flavorList", flavorList, "nodeflavor", nodeflavor)

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
		if flavor.Disk < nodeflavor.Disk {
			continue
		}
		log.InfoLog("Found closest flavor", "flavor", flavor)
		return flavor.Name, nil
	}
	return "", fmt.Errorf("no suitable platform flavor found for %s, please try a smaller flavor", nodeflavor.Key.Name)
}
