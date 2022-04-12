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
	"github.com/mobiledgex/edge-cloud/env"
)

type Config struct {
	Addr string
	Auth Auth
}

func BestConfig(addr string, ops ...BestOp) (*Config, error) {
	// default config
	cfg := &Config{
		Addr: addr,
		Auth: &NoAuth{},
	}
	if addr == "" {
		// no vault specified
		return cfg, nil
	}
	auth, err := BestAuth(ops...)
	if err != nil {
		return cfg, err
	}
	cfg.Auth = auth
	return cfg, nil
}

func NewConfig(addr string, auth Auth) *Config {
	return &Config{
		Addr: addr,
		Auth: auth,
	}
}

func NewAppRoleConfig(addr, roleID, secretID string) *Config {
	return NewConfig(addr, NewAppRoleAuth(roleID, secretID))
}

func (s *Config) Login() (*api.Client, error) {
	if s.Auth == nil {
		return nil, fmt.Errorf("No vault Auth specified")
	}
	client, err := NewClient(s.Addr)
	if err != nil {
		return nil, err
	}
	err = s.Auth.Login(client)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewClient(addr string) (*api.Client, error) {
	client, err := api.NewClient(nil)
	if err != nil {
		return nil, err
	}
	err = client.SetAddress(addr)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type BestOptions struct {
	env env.Env
}

type BestOp func(opts *BestOptions)

func WithEnv(env env.Env) BestOp {
	return func(opts *BestOptions) { opts.env = env }
}

func WithEnvMap(vars map[string]string) BestOp {
	env := env.EnvMap{Vars: vars}
	return WithEnv(&env)
}

func ApplyOps(ops ...BestOp) *BestOptions {
	opts := BestOptions{}
	for _, op := range ops {
		op(&opts)
	}
	if opts.env == nil {
		opts.env = &env.EnvOS{}
	}
	return &opts
}
