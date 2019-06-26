package pfutils

import (
	"os"
	"plugin"

	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/dind"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/fake"
	"github.com/mobiledgex/edge-cloud/log"
)

var solib = ""

func GetPlatform(plat string) (pf.Platform, error) {
	// Building plugins is slow, so directly importable
	// platforms are not built as plugins.
	if plat == "DIND" {
		return &dind.Platform{}, nil
	} else if plat == "FAKE" {
		return &fake.Platform{}, nil
	}

	// Load platform from plugin
	if solib == "" {
		solib = os.Getenv("GOPATH") + "/plugins/platforms.so"
	}
	log.DebugLog(log.DebugLevelMexos, "Loading plugin", "plugin", solib)
	plug, err := plugin.Open(solib)
	if err != nil {
		log.FatalLog("failed to load plugin", "plugin", solib, "error", err)
	}
	sym, err := plug.Lookup("GetPlatform")
	if err != nil {
		log.FatalLog("plugin does not have GetPlatform symbol", "plugin", solib)
	}
	getPlatFunc, ok := sym.(func(plat string) (pf.Platform, error))
	if !ok {
		log.FatalLog("plugin GetPlatform symbol does not implement func(plat string) (platform.Platform, error)", "plugin", solib)
	}
	log.DebugLog(log.DebugLevelMexos, "Creating platform")
	return getPlatFunc(plat)
}
