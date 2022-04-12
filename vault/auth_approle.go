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

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

type AppRoleAuth struct {
	roleID   string
	secretID string
}

func NewAppRoleAuth(roleID, secretID string) *AppRoleAuth {
	auth := AppRoleAuth{
		roleID:   roleID,
		secretID: secretID,
	}
	return &auth
}

func (s *AppRoleAuth) Login(client *api.Client) error {
	data := map[string]interface{}{
		"role_id":   s.roleID,
		"secret_id": s.secretID,
	}
	resp, err := client.Logical().Write("auth/approle/login", data)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("Empty response from Vault for approle login, possible 404 not found")
	}
	if resp.Auth == nil {
		return fmt.Errorf("no auth info returned")
	}
	client.SetToken(resp.Auth.ClientToken)
	return nil

}

func (s *AppRoleAuth) Type() string {
	return "approle"
}
