package crmutil

import (
	"bytes"
	"text/template"
)

// Context keyword for getting the DeploymentVars
var DeploymentReplaceVarsKey = "DeploymentReplaceVarsKey"

// These are CRM-specific variables that can be replaced in th Crm service context
type CrmReplaceVars struct {
	// ClusterIp of the cluster
	ClusterIp string
}

// Any configuration(envVar, configFile, manifest) can require service
// specific information filled in
type DeploymentReplaceVars struct {
	// CRM knows about the actual cluster where app is being deployed
	Deployment CrmReplaceVars
}

func ReplaceDeploymentVars(manifest string, replaceVars *DeploymentReplaceVars) (string, error) {
	tmpl := template.Must(template.New("varsReplaceTemplate").Delims("[[", "]]").Parse(manifest))
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, replaceVars); err != nil {
		return "", err
	}
	return buf.String(), nil
}
