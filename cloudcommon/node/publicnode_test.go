package node

import (
	"context"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestPublicCertManager(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	api := &cloudcommon.TestPublicCertApi{}
	mgr := NewPublicCertManager("localhost", api)
	_, err := mgr.GetServerTlsConfig(ctx)
	require.Nil(t, err)
	require.Equal(t, 1, api.GetCount)

	// force refresh
	mgr.expiresAt = time.Now()
	mgr.refreshThreshold = time.Hour
	mgr.refreshRetryDelay = time.Millisecond
	mgr.StartRefresh()
	// wait until refresh done
	for ii := 0; ii < 10; ii++ {
		if api.GetCount == 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.Equal(t, 2, api.GetCount)
	mgr.StopRefresh()
}
