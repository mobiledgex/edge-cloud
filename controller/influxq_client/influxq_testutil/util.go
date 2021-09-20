package influxq_testutil

import (
	"os"
	"strings"
	"testing"

	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

// Helper function to start influxd
// Must be called before Controller startServices()
// After calling this function, should call defer p.StopLocal() to clean up if return value is non-nil
func StartInfluxd(t *testing.T) *process.Influx {
	// Allow multiple influxDBs to run in parallel
	// from different unit test packages.
	// By default, tests within a package run in serial,
	// but tests from different packages may run in parallel.
	dir, err := os.Getwd()
	require.Nil(t, err)

	parts := strings.Split(dir, "github.com/mobiledgex/")
	require.Equal(t, 2, len(parts), "expect two parts to %v for %s", parts, dir)

	p := &process.Influx{}
	p.Common.Name = "influx-test"
	// addresses are hard-coded per package
	switch parts[1] {
	case "edge-cloud/controller":
		p.HttpAddr = "127.0.0.1:8186"
		p.BindAddr = "127.0.0.1:8187"
	case "edge-cloud/controller/influxq_client":
		p.HttpAddr = "127.0.0.1:8188"
		p.BindAddr = "127.0.0.1:8189"
	default:
		require.True(t, false, "No addresses defined for path %s", parts[1])
	}
	p.DataDir = dir + "/.influxdb"
	logfile := dir + "/influxdb.log"
	log.DebugLog(log.DebugLevelInfo, "Starting new influxDB instance", "pkg", dir, "p", p)
	// start influx
	err = p.StartLocal(logfile, process.WithCleanStartup())
	require.Nil(t, err, "start InfluxDB server")
	return p
}
