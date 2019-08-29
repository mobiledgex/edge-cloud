package pfutils

import (
	"context"
	"fmt"
	"os"
	"plugin"

	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/dind"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/fake"
	"github.com/mobiledgex/edge-cloud/log"
)

var solib = ""

func GetPlatform(ctx context.Context, plat string) (pf.Platform, error) {
	// Building plugins is slow, so directly importable
	// platforms are not built as plugins.
	if plat == "PLATFORM_TYPE_DIND" {
		outPlatform := &dind.Platform{}
		outPlatform.SetContext(ctx)
		return outPlatform, nil
	} else if plat == "PLATFORM_TYPE_FAKE" {
		outPlatform := &fake.Platform{}
		outPlatform.SetContext(ctx)
		return outPlatform, nil
	}

	// Load platform from plugin
	if solib == "" {
		solib = os.Getenv("GOPATH") + "/plugins/platforms.so"
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "Loading plugin", "plugin", solib)
	plug, err := plugin.Open(solib)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "failed to load plugin", "plugin", solib, "platform", plat, "error", err)
		return nil, fmt.Errorf("failed to load plugin for platform: %s, err: %v", plat, err)
	}
	sym, err := plug.Lookup("GetPlatform")
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "plugin does not have GetPlatform symbol", "plugin", solib)
		return nil, fmt.Errorf("failed to load plugin for platform: %s, err: GetPlatform symbol not found", plat)
	}
	getPlatFunc, ok := sym.(func(ctx context.Context, plat string) (pf.Platform, error))
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfo, "plugin GetPlatform symbol does not implement func(ctx context.Context, plat string) (platform.Platform, error)", "plugin", solib)
		return nil, fmt.Errorf("failed to load plugin for platform: %s, err: GetPlatform symbol does not implement func(ctx context.Context, plat string) (platform.Platform, error)", plat)
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "Creating platform")

	return getPlatFunc(ctx, plat)
}
