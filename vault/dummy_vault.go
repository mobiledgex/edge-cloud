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
	"net/http"
	"net/http/httptest"

	"github.com/hashicorp/vault/api"
)

// DummServer for unit testing responds to all requests with empty data.
func DummyServer() (*httptest.Server, *Config) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"data": {}}`)
	}))
	return server, NewConfig(server.URL, &NoAuth{})
}

// NoAuth skips any auth. It is used for unit testing against a fake httptest server.
type NoAuth struct{}

func (s *NoAuth) Login(client *api.Client) error {
	return nil
}

func (s *NoAuth) Type() string {
	return "none"
}
