package k8smgmt

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type networkPolicyPort struct {
	Protocol string
	Port     int32
	EndPort  int32
}

type networkPolicyArgs struct {
	Namespace      string
	ConfigLabelKey string
	ConfigLabelVal string
	Ports          []networkPolicyPort
}

// This network policy blocks all ingress for the matching pods.
// The matching pods are all pods within the namespace, due to the
// empty podSelector.
// Ingress is then allowed by the "from" rules. The first "from"
// rule allows access to any port from any pods in the same namespace.
// The second "from" rule allows access to public ports from any source.
var k8sNetworkPolicyTemplate = template.Must(template.New("networkpolicy").Parse(`apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: networkpolicy-{{.Namespace}}
  namespace: {{.Namespace}}
spec:
  podSelector:
    matchLabels:
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: {{.Namespace}}
{{- if .Ports}}
  - from:
    - ipBlock:
        cidr: 0.0.0.0/0
    ports:
{{- range .Ports}}
    - protocol: {{.Protocol}}
      port: {{.Port}}
{{- if .EndPort}}
      endPort: {{.EndPort}}
{{- end}}
{{- end}}
{{- end}}
`))

func GetNetworkPolicy(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, names *KubeNames) (string, error) {
	if names.MultitenantNamespace == "" {
		return "", fmt.Errorf("NetworkPolicy only valid for namespaced instances")
	}
	args := networkPolicyArgs{}
	args.Namespace = names.MultitenantNamespace
	//args.ConfigLabelKey = ConfigLabel
	//args.ConfigLabelVal = getConfigLabel(names)

	for _, port := range appInst.MappedPorts {
		npp := networkPolicyPort{}
		if port.Proto == dme.LProto_L_PROTO_TCP {
			npp.Protocol = "TCP"
		} else if port.Proto == dme.LProto_L_PROTO_UDP {
			npp.Protocol = "UDP"
		} else {
			continue
		}
		npp.Port = port.InternalPort
		npp.EndPort = port.EndPort
		args.Ports = append(args.Ports, npp)
	}

	buf := bytes.Buffer{}
	err := k8sNetworkPolicyTemplate.Execute(&buf, &args)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
