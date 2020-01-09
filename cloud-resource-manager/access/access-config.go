package access

import (
	"context"
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
)

const AppAccessCustomization = "appAccessCustomization"

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

func GetAppAccessConfig(ctx context.Context, configs []*edgeproto.ConfigFile) (*AppAccessConfig, error) {
	deploymentVars, varsFound := ctx.Value(crmutil.DeploymentReplaceVarsKey).(*crmutil.DeploymentReplaceVars)
	var aac AppAccessConfig

	log.SpanLog(ctx, log.DebugLevelMexos, "getAppAccessConfig", "deploymentVars", deploymentVars, "varsFound", varsFound)
	if !varsFound {
		return nil, fmt.Errorf("unable to find replacement vars")
	}
	// Walk the Configs in the App and generate the yaml files from the helm customization ones
	for _, v := range configs {
		if v.Kind == AppAccessCustomization {
			cfg := v.Config
			// Fill in the Deployment Vars passed as a variable through the context
			cfg, err := crmutil.ReplaceDeploymentVars(cfg, deploymentVars)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelMexos, "getAppAccessConfig failed to replace CRM variables",
					"config file", v.Config, "DeploymentVars", deploymentVars, "error", err)
				return nil, err
			}
			err = yaml.Unmarshal([]byte(cfg), &aac)
			if err != nil {
				return nil, fmt.Errorf("unable to unmarshall app access config: %s err: %v", cfg, err)
			}
			log.SpanLog(ctx, log.DebugLevelMexos, "Got app access config", "aac", aac)
		}
	}
	return &aac, nil
}
