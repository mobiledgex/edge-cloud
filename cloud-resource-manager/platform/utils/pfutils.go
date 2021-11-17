package pfutils

import (
	"context"
	"fmt"
	"os"
	"plugin"

	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/dind"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/fake"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/kind"
	"github.com/mobiledgex/edge-cloud/log"
)

var solib = ""
var GetPlatformFunc func(plat string) (pf.Platform, error)

func GetPlatform(ctx context.Context, plat string, setVersionProps func(context.Context, map[string]string)) (pf.Platform, error) {
	// Building plugins is slow, so directly importable
	// platforms are not built as plugins.
	if plat == "PLATFORM_TYPE_DIND" {
		return &dind.Platform{}, nil
	} else if plat == "PLATFORM_TYPE_FAKE" {
		return &fake.Platform{}, nil
	} else if plat == "PLATFORM_TYPE_FAKE_SINGLE_CLUSTER" {
		return &fake.PlatformSingleCluster{}, nil
	} else if plat == "PLATFORM_TYPE_KIND" {
		return &kind.Platform{}, nil
	} else if plat == "PLATFORM_TYPE_FAKE_VM_POOL" {
		return &fake.PlatformVMPool{}, nil
	}
	if GetPlatformFunc == nil {
		plug, err := loadPlugin(ctx)
		if err != nil {
			return nil, err
		}
		sym, err := plug.Lookup("GetPlatform")
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "plugin does not have GetPlatform symbol", "plugin", solib)
			return nil, fmt.Errorf("failed to load plugin for GetPlatform, err: GetPlatform symbol not found")
		}
		getPlatFunc, ok := sym.(func(plat string) (pf.Platform, error))
		if !ok {
			log.SpanLog(ctx, log.DebugLevelInfo, "plugin GetPlatform symbol does not implement func(plat string) (platform.Platform, error)", "plugin", solib)
			return nil, fmt.Errorf("failed to load plugin for platform: %s, err: GetPlatform symbol does not implement func(plat string) (platform.Platform, error)", plat)
		}
		GetPlatformFunc = getPlatFunc
		if setVersionProps != nil {
			pf, err := getPlatFunc(plat)
			if err == nil {
				setVersionProps(ctx, pf.GetVersionProperties())
			}
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "Creating platform")
	return GetPlatformFunc(plat)
}

func GetClusterSvc(ctx context.Context, pluginRequired bool) (pf.ClusterSvc, error) {
	plug, err := loadPlugin(ctx)
	if err != nil {
		if !pluginRequired {
			log.SpanLog(ctx, log.DebugLevelInfo, "plugin not required, ignoring load plugin failure", "err", err)
			err = nil
		}
		return nil, err
	}
	sym, err := plug.Lookup("GetClusterSvc")
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "plugin does not have GetClusterSvc symbol", "plugin", solib)
	}
	getClusterSvcFunc, ok := sym.(func() (pf.ClusterSvc, error))
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfo, "plugin GetClusterSvc symbol does not implement func() (platform.ClusterSvc, error)", "plugin", solib)
		return nil, fmt.Errorf("failed to load plugin %s, err: GetClusterSvc symbol does not implement func() (platform.ClusterSvc, error)", solib)
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "Creating ClusterSvc")
	return getClusterSvcFunc()
}

func loadPlugin(ctx context.Context) (*plugin.Plugin, error) {
	// Load platform from plugin
	if solib == "" {
		solib = os.Getenv("GOPATH") + "/plugins/platforms.so"
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "Loading plugin", "plugin", solib)
	plug, err := plugin.Open(solib)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "failed to load plugin", "plugin", solib, "error", err)
		return nil, fmt.Errorf("failed to load plugin %s, err: %v", solib, err)
	}
	return plug, nil
}
