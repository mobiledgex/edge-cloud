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

package access

import (
	"context"
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
)

type AppAccessConfig struct {
	DnsOverride         string `yaml:"dnsOverride"`
	LbTlsCertCommonName string `yaml:"lbTlsCertCommonName"`
}

// TLSCert is optionally used for TLS termination
type TLSCert struct {
	CommonName string
	CertString string
	KeyString  string
	TTL        int64
}

func GetAppAccessConfig(ctx context.Context, configs []*edgeproto.ConfigFile, delims string) (*AppAccessConfig, error) {
	deploymentVars, varsFound := ctx.Value(crmutil.DeploymentReplaceVarsKey).(*crmutil.DeploymentReplaceVars)
	var aac AppAccessConfig

	log.SpanLog(ctx, log.DebugLevelInfra, "getAppAccessConfig", "deploymentVars", deploymentVars, "varsFound", varsFound)
	if !varsFound {
		// If no deployment vars were populated, return an empty config
		return &aac, nil
	}
	// Walk the Configs in the App and generate the yaml files from the helm customization ones
	for _, v := range configs {
		if v.Kind == edgeproto.AppAccessCustomization {
			cfg, err := cloudcommon.GetDeploymentManifest(ctx, nil, v.Config)
			if err != nil {
				return nil, err
			}
			// Fill in the Deployment Vars passed as a variable through the context
			cfg, err = crmutil.ReplaceDeploymentVars(cfg, delims, deploymentVars)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "getAppAccessConfig failed to replace CRM variables",
					"config file", v.Config, "DeploymentVars", deploymentVars, "error", err)
				return nil, err
			}
			err = yaml.Unmarshal([]byte(cfg), &aac)
			if err != nil {
				return nil, fmt.Errorf("unable to unmarshall app access config: %s err: %v", cfg, err)
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "Got app access config", "aac", aac)
		}
	}
	return &aac, nil
}
