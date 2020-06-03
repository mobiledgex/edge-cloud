package crmutil

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Context keyword for getting the DeploymentVars
var DeploymentReplaceVarsKey = "DeploymentReplaceVarsKey"

// These are CRM-specific variables that can be replaced in th Crm service context
type CrmReplaceVars struct {
	// ClusterIp of the cluster master
	ClusterIp string
	// Cloudlet Name
	CloudletName string
	// Cluster Name
	ClusterName string
	// Cloudlet Organization
	CloudletOrg string
	// App Developer Organization
	AppOrg string
	// DNS zone
	DnsZone string
}

// Any configuration(envVar, configFile, manifest) can require service
// specific information filled in
type DeploymentReplaceVars struct {
	// CRM knows about the actual cluster where app is being deployed
	Deployment CrmReplaceVars
}

func ReplaceDeploymentVars(manifest string, delims string, replaceVars *DeploymentReplaceVars) (s string, err error) {
	if delims == "" {
		delims = "[[ ]]"
	}
	delimiter := strings.Split(delims, " ")
	if len(delimiter) != 2 {
		return "", fmt.Errorf("invalid app template delimiter %s, valid format '<START-DELIM> <END-DELIM>'", delims)
	}
	defer func() {
		// template.Parse panics on error, handling it using recover
		if r := recover(); r != nil {
			err = fmt.Errorf("Error with parsing %v, try changing templatedelimiter in app definition", r)
		}
	}()
	tmpl := template.Must(template.New("varsReplaceTemplate").Delims(delimiter[0], delimiter[1]).Parse(manifest))
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, replaceVars); err != nil {
		return "", err
	}
	return buf.String(), nil
}
