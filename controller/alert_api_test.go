package main

import (
	"context"
	"testing"

	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/vault"
)

func TestAlertApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	testinit()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	for _, alert := range testutil.AlertData {
		alertApi.Update(ctx, &alert, 0)
	}
	testutil.InternalAlertTest(t, "show", &alertApi, testutil.AlertData)

	dummy.Stop()
}

// Set up globals for API unit tests
func testinit() {
	objstore.InitRegion(1)
	tMode := true
	testMode = &tMode
	dockerRegistry := "docker.mobiledgex.net"
	registryFQDN = &dockerRegistry
	vaultConfig, _ = vault.BestConfig("")
	services.events = influxq.NewInfluxQ("events", "user", "pass")
}
