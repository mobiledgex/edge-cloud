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
	"runtime"

	"github.com/hashicorp/vault/api"
)

type Auth interface {
	// Login to vault and set the vault token on the client
	Login(client *api.Client) error
	// Return auth type
	Type() string
}

// BestAuth determines the best auth to use based on the environment.
func BestAuth(ops ...BestOp) (Auth, error) {
	opts := ApplyOps(ops...)

	roleID := opts.env.Getenv("VAULT_ROLE_ID")
	secretID := opts.env.Getenv("VAULT_SECRET_ID")
	if roleID != "" && secretID != "" {
		return NewAppRoleAuth(roleID, secretID), nil
	}
	githubID := opts.env.Getenv("GITHUB_ID")
	if runtime.GOOS == "darwin" && githubID != "" {
		return NewGithubAuth(githubID), nil
	}
	token := opts.env.Getenv("VAULT_TOKEN")
	if token != "" {
		return NewTokenAuth(token), nil
	}
	ldapID := opts.env.Getenv("LDAP_ID")
	if ldapID != "" {
		return NewLdapAuth(ldapID, opts.env.Getenv("LDAP_PASS")), nil
	}
	return nil, fmt.Errorf("No appropriate Vault auth found, please set VAULT_ROLE_ID and VAULT_SECRET_ID for approle auth, GITHUB_ID for github token auth, or VAULT_TOKEN for token auth, or LDAP_ID for LDAP auth.")
}
