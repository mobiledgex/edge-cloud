package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestSettingsApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	testinit()
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	testUpdateMasterNodeFlavor(t, ctx)
	dummy.Stop()
}

func testUpdateMasterNodeFlavor(t *testing.T, ctx context.Context) {
	// Settings has MasterNodeFlavor defaults to ""
	// When you set it to some mex flavor name, that flavor must pre-exist in the DB
	// or update will fail.

	masterFlavor := edgeproto.Flavor{}
	err := settingsApi.initDefaults(ctx)
	require.Nil(t, err, "settingsApi.initDefaults")

	settings := settingsApi.Get()
	if settings.MasterNodeFlavor == "" {
		settings.MasterNodeFlavor = "IDONTEXIST"
		// setup for update
		settings.Fields = make([]string, 0)
		settings.Fields = append(settings.Fields, edgeproto.SettingsFieldMasterNodeFlavor)
		_, err := settingsApi.UpdateSettings(ctx, settings)
		require.Equal(t, "Flavor must preexist",
			err.Error())

		// create a modest flavor
		masterFlavor.Key.Name = "IDOEXIST"
		masterFlavor.Ram = 4096
		masterFlavor.Vcpus = 2
		masterFlavor.Ram = 4096
		masterFlavor.Disk = 40
		_, err = flavorApi.CreateFlavor(ctx, &masterFlavor)
		require.Nil(t, err, "CreateFlavor")

		settings.MasterNodeFlavor = "IDOEXIST"
		_, err = settingsApi.UpdateSettings(ctx, settings)
		require.Nil(t, err, "UpdateSettings")

		// must also be true:
		testsettings := settingsApi.Get()
		require.Equal(t, testsettings.MasterNodeFlavor, masterFlavor.Key.Name)

		// now FlavorDelete should balk at removing flavor "IDOEXIST" until its removed from settings
		_, err = flavorApi.DeleteFlavor(ctx, &masterFlavor)
		require.Equal(t, "Flavor in use by Settings MasterNodeFlavor, change Settings.MasterNodeFlavor first",
			err.Error())

		settings.MasterNodeFlavor = ""
		_, err = settingsApi.UpdateSettings(ctx, settings)
		require.Nil(t, err, "UpdateSettings")
		_, err = flavorApi.DeleteFlavor(ctx, &masterFlavor)
		require.Nil(t, err, "DeleteFlavor")
	} else {
		panic("default setting for MasterNodeFlavor not empty string")
	}
}
