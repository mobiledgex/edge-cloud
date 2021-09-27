package influxsup

import (
	"fmt"
	"strings"

	influxclient "github.com/influxdata/influxdb/client/v2"
)

// Get a client to access InfluxDB for edge cloud.
// InfluxDB runs with a public cert (unless running locally) with
// user/pass authentication (which may be blank for local testing).
func GetClient(addr, user, pass string) (influxclient.Client, error) {
	if !strings.Contains(addr, "http://") && !strings.Contains(addr, "https://") {
		return nil, fmt.Errorf("InfluxDB client address %s must contain http:// or https://", addr)
	}
	conf := influxclient.HTTPConfig{
		Addr:     addr,
		Username: user,
		Password: pass,
	}
	return influxclient.NewHTTPClient(conf)
}
