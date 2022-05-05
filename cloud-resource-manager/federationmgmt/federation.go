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

package federationmgmt

import (
	"context"
	"fmt"

	"github.com/edgexr/edge-cloud/log"
	"github.com/edgexr/edge-cloud/vault"
)

type ApiKeyData struct {
	Data string `json:"data"`
}

func getApiKeyVaultPath(fedName string) string {
	return fmt.Sprintf("secret/data/federation/%s", fedName)
}

func GetFederationAPIKey(ctx context.Context, vaultConfig *vault.Config, fedName string) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "fetch federation API key", "fedName", fedName)
	vaultPath := getApiKeyVaultPath(fedName)
	apiKeyData := &ApiKeyData{}
	err := vault.GetData(vaultConfig, vaultPath, 0, apiKeyData)
	if err != nil {
		return "", fmt.Errorf("Unable to find federation API key from vault path %s, %v", vaultPath, err)
	}
	if apiKeyData.Data == "" {
		return "", fmt.Errorf("Unable to find federation API key from vault path %s", vaultPath)
	}
	return apiKeyData.Data, nil
}

func PutAPIKeyToVault(ctx context.Context, vaultConfig *vault.Config, fedName, apiKey string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "upload federation API key to vault", "fedName", fedName)
	vPath := getApiKeyVaultPath(fedName)
	err := vault.PutData(vaultConfig, vPath, &ApiKeyData{Data: apiKey})
	if err != nil {
		return fmt.Errorf("Unable to store partner API key in vault: %s", err)
	}
	return nil
}

func DeleteAPIKeyFromVault(ctx context.Context, vaultConfig *vault.Config, fedName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "delete federation API key rom vault", "fedName", fedName)
	client, err := vaultConfig.Login()
	if err == nil {
		err = vault.DeleteKV(client, getApiKeyVaultPath(fedName))
		if err != nil {
			return fmt.Errorf("Failed to delete API Key from vault %s, %v", fedName, err)
		}
		return nil
	}
	return fmt.Errorf("Failed to login in to vault to delete partner federation API key %s, %v", fedName, err)
}
