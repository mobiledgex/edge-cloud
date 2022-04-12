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

package vault

import "fmt"

func SignSSHKey(config *Config, publicKey string) (string, error) {
	data := map[string]interface{}{
		"public_key": publicKey,
	}
	client, err := config.Login()
	if err != nil {
		return "", err
	}
	ssh := client.SSH()
	secret, err := ssh.SignKey("machine", data)
	if err != nil {
		return "", err
	}
	signedKey, ok := secret.Data["signed_key"]
	if !ok {
		return "", fmt.Errorf("failed to get signed key from vault: %v", secret)
	}
	signedKeyStr, ok := signedKey.(string)
	if !ok {
		return "", fmt.Errorf("invalid signed key from vault: %v", signedKey)
	}
	return signedKeyStr, nil
}

type MEXKey struct {
	PrivateKey string `mapstructure:"private_key"`
	PublicKey  string `mapstructure:"public_key"`
}
