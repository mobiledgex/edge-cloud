package influxq_testutil

import (
	"os/exec"
	"testing"

	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/stretchr/testify/require"
)

// Helper function to start influxd
// Must be called before Controller startServices()
// After calling this function, should call defer p.StopLocal() to clean up if return value is non-nil
func StartInfluxd(t *testing.T, addr string) *process.Influx {
	// start influxd if not already running
	_, err := exec.Command("sh", "-c", "pgrep -x influxd").Output()
	if err != nil {
		p := &process.Influx{}
		p.Common.Name = "influx-test"
		p.HttpAddr = addr
		p.DataDir = "/var/tmp/.influxdb"
		// start influx
		err = p.StartLocal("/var/tmp/influxdb.log", process.WithCleanStartup())
		require.Nil(t, err, "start InfluxDB server")
		return p
	}
	return nil
}
