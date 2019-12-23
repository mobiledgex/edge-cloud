package influxsup

import (
	influxclient "github.com/influxdata/influxdb/client/v2"
)

// Get a client to access InfluxDB for edge cloud.
// InfluxDB runs with a public cert (unless running locally) with
// user/pass authentication (which may be blank for local testing).
func GetClient(addr, user, pass string) (influxclient.Client, error) {
	conf := influxclient.HTTPConfig{
		Addr:     addr,
		Username: user,
		Password: pass,
	}
	return influxclient.NewHTTPClient(conf)
}
