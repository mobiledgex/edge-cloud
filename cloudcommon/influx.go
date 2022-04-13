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

package cloudcommon

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

type InfluxCreds struct {
	User string
	Pass string
}

func GetInfluxDataAuth(vaultConfig *vault.Config, region string) (*InfluxCreds, error) {
	if vaultConfig.Addr == "" {
		// no vault address, either unit test or no auth needed
		return &InfluxCreds{}, nil
	}
	vaultPath := "/secret/data/" + region + "/accounts/influxdb"
	log.DebugLog(log.DebugLevelApi, "get influxDB credentials ", "vault-path", vaultPath)
	creds := &InfluxCreds{}
	err := vault.GetData(vaultConfig, vaultPath, 0, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to get influxDB credentials for %s, %v", vaultPath, err)
	}
	return creds, nil
}
