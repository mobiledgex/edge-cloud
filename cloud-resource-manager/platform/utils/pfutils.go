// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pfutils

import (
	"context"
	"fmt"
	"os"
	"plugin"
	"strings"

	pf "github.com/edgexr/edge-cloud/cloud-resource-manager/platform"
	"github.com/edgexr/edge-cloud/cloud-resource-manager/platform/dind"
	"github.com/edgexr/edge-cloud/cloud-resource-manager/platform/fake"
	"github.com/edgexr/edge-cloud/cloud-resource-manager/platform/kind"
	"github.com/edgexr/edge-cloud/cloudcommon"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/log"
	"github.com/edgexr/edge-cloud/util"
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

// GetAppInstId returns a string for this AppInst that is likely to be
// unique within the region. It does not guarantee uniqueness.
// The delimiter '.' is removed from the AppInstId so that it can be used
// to append further strings to this ID to build derived unique names.
// Salt can be used by the caller to add an extra field if needed
// to ensure uniqueness. In all cases, any requirements for uniqueness
// must be guaranteed by the caller. Name sanitization for the platform is performed
func GetAppInstId(ctx context.Context, appInst *edgeproto.AppInst, app *edgeproto.App, salt string, platformType edgeproto.PlatformType) (string, error) {
	fields := []string{}

	cloudletPlatform, err := GetPlatform(ctx, platformType.String(), nil)
	if err != nil {
		return "", err
	}
	appName := util.DNSSanitize(appInst.Key.AppKey.Name)
	dev := util.DNSSanitize(appInst.Key.AppKey.Organization)
	ver := util.DNSSanitize(appInst.Key.AppKey.Version)
	appId := fmt.Sprintf("%s%s%s", dev, appName, ver)
	fields = append(fields, appId)

	if cloudcommon.IsClusterInstReqd(app) {
		cluster := util.DNSSanitize(appInst.Key.ClusterInstKey.ClusterKey.Name)
		fields = append(fields, cluster)
	}

	loc := util.DNSSanitize(appInst.Key.ClusterInstKey.CloudletKey.Name)
	fields = append(fields, loc)

	oper := util.DNSSanitize(appInst.Key.ClusterInstKey.CloudletKey.Organization)
	fields = append(fields, oper)

	if salt != "" {
		salt = util.DNSSanitize(salt)
		fields = append(fields, salt)
	}
	return cloudletPlatform.NameSanitize(strings.Join(fields, "-")), nil
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
