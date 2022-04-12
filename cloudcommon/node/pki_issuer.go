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

package node

// Three intermediate certificates are used to issue certificates to services.
// The global intermediate certificate is used for global services.
// The regional intermediate certificate is used for regional services like
// the Controller.
// The cloudlet intermediate certificate is used for regional services that
// run in partner environments, where we have less control over security.
const (
	CertIssuerGlobal           = "pki-global"
	CertIssuerRegional         = "pki-regional"
	CertIssuerRegionalCloudlet = "pki-regional-cloudlet"
	NoTlsClientIssuer          = ""
)

// To avoid services from one region to be able to talk to services from
// another region, all certs issued from regional intermediate authorities
// must be tagged with the region in the URI SANs, of the form region://<name>.
// Vault issuing roles are used to ensure that services from one region
// can only add a URI SAN from their own region. Regions are verified by
// the custom verification function on the tls config.

type MatchCA struct {
	Issuer             string
	RequireRegionMatch bool
}

func GlobalMatchCA() MatchCA {
	return MatchCA{
		Issuer: CertIssuerGlobal,
	}
}

func AnyRegionalMatchCA() MatchCA {
	return MatchCA{
		Issuer: CertIssuerRegional,
	}
}

func SameRegionalMatchCA() MatchCA {
	return MatchCA{
		Issuer:             CertIssuerRegional,
		RequireRegionMatch: true,
	}
}

func SameRegionalCloudletMatchCA() MatchCA {
	return MatchCA{
		Issuer:             CertIssuerRegionalCloudlet,
		RequireRegionMatch: true,
	}
}
