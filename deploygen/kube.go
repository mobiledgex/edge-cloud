package deploygen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	yaml "gopkg.in/yaml.v2"
)

var kubeLbT *template.Template
var kubeAppT *template.Template

func init() {
	kubeLbT = template.Must(template.New("lb").Parse(lbTemplate))
	kubeAppT = template.Must(template.New("app").Parse(appTemplate))
}

type kubeBasicGen struct {
	app   *AppSpec
	err   error
	files []string
	ports []kubePort
}

type kubePort struct {
	Proto     string
	KubeProto string
	Port      string
}

func kubeBasic(app *AppSpec) (string, error) {
	gen := kubeBasicGen{
		app:   app,
		files: []string{},
		ports: setKubePorts(app.Ports),
	}
	gen.kubeLb([]string{"tcp", "http"})
	gen.kubeLb([]string{"udp"})
	gen.kubeApp()
	if gen.err != nil {
		return "", gen.err
	}
	return strings.Join(gen.files, "---\n"), nil

}

func setKubePorts(ports []PortSpec) []kubePort {
	kports := []kubePort{}
	for _, port := range ports {
		kp := kubePort{
			Proto: strings.ToLower(port.Proto),
			Port:  port.Port,
		}
		switch port.Proto {
		case "tcp":
			fallthrough
		case "http":
			kp.KubeProto = "TCP"
		case "udp":
			kp.KubeProto = "UDP"
		}
		kports = append(kports, kp)
	}
	return kports
}

// Kubernetes load balancers don't support mixed protocols
// on load balancers, so we generate a service only for
// ports of the matching protocol.
func (g *kubeBasicGen) kubeLb(protos []string) {
	if g.err != nil {
		return
	}
	lb := lbData{
		Name:  util.DNSSanitize(g.app.Name + "-" + protos[0]),
		Run:   util.K8SSanitize(g.app.Name),
		Ports: []kubePort{},
	}
	for _, port := range g.ports {
		for ii, _ := range protos {
			if port.Proto == protos[ii] {
				lb.Ports = append(lb.Ports, port)
				break
			}
		}
	}
	if len(lb.Ports) == 0 {
		return
	}
	buf := bytes.Buffer{}
	g.err = kubeLbT.Execute(&buf, &lb)
	if g.err != nil {
		return
	}
	g.files = append(g.files, buf.String())
}

type lbData struct {
	Name  string
	Run   string
	Ports []kubePort
}

var lbTemplate = `apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
  labels:
    run: {{.Run}}
spec:
  type: LoadBalancer
  ports:
{{- range .Ports}}
  - name: {{.Proto}}{{.Port}}
    protocol: {{.KubeProto}}
    port: {{.Port}}
    targetPort: {{.Port}}
{{- end}}
  selector:
    run: {{.Run}}
`

type ConfigVars struct {
	Replicas int
}

const (
	DeploygenConfigVars = "deploygenConfig"
	DefaultReplicas     = 1
)

func UnmarshalDeploygenConfig(configs []*edgeproto.ConfigFile) (ConfigVars, error) {
	configVars := ConfigVars{}

	// By default, Replicas is set to 1
	configVars.Replicas = DefaultReplicas

	for _, v := range configs {
		if v.Kind == DeploygenConfigVars {
			if err := yaml.Unmarshal([]byte(v.Config), &configVars); err != nil {
				err = fmt.Errorf("cannot unmarshal config vars, kind %s, "+
					"config %s, error %s", v.Kind, v.Config, err)
				return configVars, err
			}
			if configVars.Replicas == 0 {
				configVars.Replicas = DefaultReplicas
			}
		}
	}
	return configVars, nil
}

func (g *kubeBasicGen) kubeApp() {
	if g.err != nil {
		return
	}
	var cs []string
	if g.app.Command != "" {
		cs = strings.Split(g.app.Command, " ")
	}

	configVars, err := UnmarshalDeploygenConfig(g.app.Configs)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "Unmarshal error", "error", err)
		return
	}

	data := appData{
		Name:      util.DNSSanitize(g.app.Name + "-deployment"),
		DNSName:   util.DNSSanitize(g.app.Name),
		Run:       util.K8SSanitize(g.app.Name),
		Ports:     g.ports,
		ImagePath: g.app.ImagePath,
		Command:   cs,
		Replicas:  configVars.Replicas,
	}

	buf := bytes.Buffer{}
	g.err = kubeAppT.Execute(&buf, &data)
	if g.err != nil {
		return
	}
	g.files = append(g.files, buf.String())
}

type appData struct {
	Name      string
	DNSName   string
	Run       string
	ImagePath string
	Ports     []kubePort
	Command   []string
	Replicas  int
}

var appTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
spec:
  selector:
    matchLabels:
      run: {{.Run}}
  replicas: {{.Replicas}}
  template:
    metadata:
      labels:
        run: {{.Run}}
    spec:
      volumes:
      imagePullSecrets:
      - name: mexregistrysecret
      containers:
      - name: {{.DNSName}}
        image: {{.ImagePath}}
        imagePullPolicy: Always
        ports:
{{- range .Ports}}
        - containerPort: {{.Port}}
          protocol: {{.KubeProto}}
{{- end}}
{{- if .Command}}
        command:
{{- range .Command}}
        - "{{.}}"
{{- end}}
{{- end}}
`
