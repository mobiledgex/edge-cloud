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
	"context"

	edgetls "github.com/edgexr/edge-cloud/tls"
	"github.com/edgexr/edge-cloud/vault"
)

// GetPublicCertApi abstracts the way the public cert is retrieved.
// Certain services, like DME running on a Cloudlet, may need to connect
// to the controller to get a public cert from Vault.
type GetPublicCertApi interface {
	GetPublicCert(ctx context.Context, commonName string) (*vault.PublicCert, error)
}

// VaultPublicCertApi implements GetPublicCertApi by connecting directly to Vault.
type VaultPublicCertApi struct {
	VaultConfig *vault.Config
}

func (s *VaultPublicCertApi) GetPublicCert(ctx context.Context, commonName string) (*vault.PublicCert, error) {
	return vault.GetPublicCert(s.VaultConfig, commonName)
}

// TestPublicCertApi implements GetPublicCertApi for unit/e2e testing
type TestPublicCertApi struct {
	GetCount int
}

func (s *TestPublicCertApi) GetPublicCert(ctx context.Context, commonName string) (*vault.PublicCert, error) {
	cert := &vault.PublicCert{}
	cert.Cert = edgetls.LocalTestCert
	cert.Key = edgetls.LocalTestKey
	// 24 hours in seconds
	cert.TTL = 24 * 3600
	s.GetCount++
	return cert, nil
}
