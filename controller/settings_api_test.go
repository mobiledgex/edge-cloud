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

package main

import (
	"context"
	"testing"

	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestSettingsApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	testSvcs := testinit(ctx, t)
	defer testfinish(testSvcs)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()

	testUpdateMasterNodeFlavor(t, ctx, apis)
	dummy.Stop()
}

func testUpdateMasterNodeFlavor(t *testing.T, ctx context.Context, apis *AllApis) {
	// Settings has MasterNodeFlavor defaults to ""
	// When you set it to some mex flavor name, that flavor must pre-exist in the DB
	// or update will fail.

	masterFlavor := edgeproto.Flavor{}
	err := apis.settingsApi.initDefaults(ctx)
	require.Nil(t, err, "settingsApi.initDefaults")

	settings := apis.settingsApi.Get()
	if settings.MasterNodeFlavor == "" {
		settings.MasterNodeFlavor = "IDONTEXIST"
		// setup for update
		settings.Fields = make([]string, 0)
		settings.Fields = append(settings.Fields, edgeproto.SettingsFieldMasterNodeFlavor)
		_, err := apis.settingsApi.UpdateSettings(ctx, settings)
		require.Equal(t, "Flavor must preexist",
			err.Error())

		// create a modest flavor
		masterFlavor.Key.Name = "IDOEXIST"
		masterFlavor.Ram = 4096
		masterFlavor.Vcpus = 2
		masterFlavor.Ram = 4096
		masterFlavor.Disk = 40
		_, err = apis.flavorApi.CreateFlavor(ctx, &masterFlavor)
		require.Nil(t, err, "CreateFlavor")

		settings.MasterNodeFlavor = "IDOEXIST"
		_, err = apis.settingsApi.UpdateSettings(ctx, settings)
		require.Nil(t, err, "UpdateSettings")

		// must also be true:
		testsettings := apis.settingsApi.Get()
		require.Equal(t, testsettings.MasterNodeFlavor, masterFlavor.Key.Name)

		// now FlavorDelete should balk at removing flavor "IDOEXIST" until its removed from settings
		_, err = apis.flavorApi.DeleteFlavor(ctx, &masterFlavor)
		require.Equal(t, "Flavor in use by Settings MasterNodeFlavor, change Settings.MasterNodeFlavor first",
			err.Error())

		settings.MasterNodeFlavor = ""
		_, err = apis.settingsApi.UpdateSettings(ctx, settings)
		require.Nil(t, err, "UpdateSettings")
		_, err = apis.flavorApi.DeleteFlavor(ctx, &masterFlavor)
		require.Nil(t, err, "DeleteFlavor")
	} else {
		panic("default setting for MasterNodeFlavor not empty string")
	}
}
