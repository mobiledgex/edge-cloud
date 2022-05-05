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
	"net/url"

	"github.com/hashicorp/vault/api"
	"github.com/edgexr/edge-cloud/log"
)

type LdapAuth struct {
	username string
	password string
}

// LDAP Auth authenticates with Vault via the LDAP server
// configured for Vault.

func NewLdapAuth(username, password string) *LdapAuth {
	return &LdapAuth{
		username: username,
		password: password,
	}
}

func (s *LdapAuth) Type() string {
	return "ldap"
}

func (s *LdapAuth) Login(client *api.Client) error {
	if s.password == "" {
		server := client.Address() + "/ldap"
		log.DebugLog(log.DebugLevelInfo, "ldap.vault secret lookup", "account", s.username, "server", server)
		password, err := FindKeychainSecret(s.username, server)
		if err != nil {
			return err
		}
		if password == "" {
			return fmt.Errorf("empty password for keychain entry account %s server %s", s.username, server)
		}
		s.password = password
	}
	data := map[string]interface{}{
		"password": s.password,
	}

	// enforce https
	u, err := url.Parse(client.Address())
	if err != nil {
		return fmt.Errorf("unable to parse vault address %s", client.Address())
	}
	if u.Scheme != "https" {
		return fmt.Errorf("vault address (%s) must use https for gitlab auth", client.Address())
	}

	resp, err := client.Logical().Write("/auth/ldap/login/"+s.username, data)
	if err != nil {
		return err
	}
	if resp.Auth == nil {
		return fmt.Errorf("no auth info returned")
	}
	client.SetToken(resp.Auth.ClientToken)
	return nil
}
