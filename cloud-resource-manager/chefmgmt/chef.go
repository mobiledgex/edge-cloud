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

package chefmgmt

import (
	"context"
	"fmt"

	"github.com/edgexr/edge-cloud/log"
	"github.com/edgexr/edge-cloud/vault"
)

type ChefAuthKey struct {
	ApiKey        string `json:"apikey"`
	ValidationKey string `json:"validationkey"`
}

func GetChefAuthKeys(ctx context.Context, vaultConfig *vault.Config) (*ChefAuthKey, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "fetch chef auth keys")
	vaultPath := "/secret/data/accounts/chef"
	auth := &ChefAuthKey{}
	err := vault.GetData(vaultConfig, vaultPath, 0, auth)
	if err != nil {
		return nil, fmt.Errorf("Unable to find chef auth keys from vault path %s, %v", vaultPath, err)
	}
	if auth.ApiKey == "" {
		return nil, fmt.Errorf("Unable to find chef API key")
	}
	return auth, nil
}
