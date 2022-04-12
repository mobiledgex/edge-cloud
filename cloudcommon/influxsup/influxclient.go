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
