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

package apis

import (
	"fmt"
	"log"
	"os"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"gopkg.in/yaml.v2"
)

type influxData struct {
	Database string
	Cmd      string
}

func RunInfluxAPI(api, influxname, apiFile string, apiFileVars map[string]string, outputDir string) bool {
	log.Printf("Running influx APIs for %s\n", apiFile)

	if apiFile == "" {
		log.Println("Error: Cannot run influx APIs without API file")
		return false
	}

	data := influxData{}
	err := util.ReadYamlFile(apiFile, &data, util.WithVars(apiFileVars))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error in unmarshal for file %s", apiFile)
		os.Exit(1)
	}

	proc := util.GetInflux(influxname)
	cl, err := proc.GetClient()
	if err != nil {
		log.Printf("failed to create new influx client, process %v, %v\n", proc, err)
		return false
	}
	query := client.NewQuery(data.Cmd, data.Database, "")
	resp, err := cl.Query(query)
	if err == nil {
		err = resp.Error()
	}
	if err != nil {
		log.Printf("failed to query influxdb, cmd %s, db %s, err %v\n", data.Cmd, data.Database, err)
		return false
	}
	util.FilterInfluxTime(resp.Results)
	out, err := yaml.Marshal(resp.Results)
	if err != nil {
		log.Printf("failed to marshal influx query result, %v, %v\n", resp.Results, err)
		return false
	}
	truncate := true
	util.PrintToFile("show-commands.yml", outputDir, string(out), truncate)
	return true
}
