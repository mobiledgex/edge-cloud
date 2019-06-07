package deploygen

import (
	"bytes"
	"github.com/mobiledgex/edge-cloud/util"
	"strings"
	"text/template"
)

var MexRegistry = "docker.mobiledgex.net"
var MexRegistrySecret = "mexgitlabsecret"

var kubeLbT *template.Template
var kubeAppDpT *template.Template
var kubeAppDsT *template.Template

func init() {
	kubeLbT = template.Must(template.New("lb").Parse(lbTemplate))
	kubeAppDpT = template.Must(template.New("appdp").Parse(dpTemplate + podTemplate))
	kubeAppDsT = template.Must(template.New("appds").Parse(dsTemplate + podTemplate))
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

func setKubePorts(ports []util.PortSpec) []kubePort {
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

func (g *kubeBasicGen) kubeApp() {
	if g.err != nil {
		return
	}
	var cs []string
	if g.app.Command != "" {
		cs = strings.Split(g.app.Command, " ")
	}
	data := appData{
		Name:           util.DNSSanitize(g.app.Name + "-deployment"),
		DNSName:        util.DNSSanitize(g.app.Name),
		Run:            util.K8SSanitize(g.app.Name),
		Ports:          g.ports,
		ImagePath:      g.app.ImagePath,
		Command:        cs,
		RegistrySecret: MexRegistrySecret,
	}
	buf := bytes.Buffer{}
	if g.app.Scale {
		g.err = kubeAppDsT.Execute(&buf, &data)
	} else {
		g.err = kubeAppDpT.Execute(&buf, &data)
	}
	if g.err != nil {
		return
	}
	g.files = append(g.files, buf.String())
}

type appData struct {
	Name           string
	DNSName        string
	Run            string
	ImagePath      string
	Ports          []kubePort
	Command        []string
	RegistrySecret string
}

var dpTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
spec:
  replicas: 1`

var dsTemplate = `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: {{.Name}}
spec:`

var podTemplate = `
  selector:
    matchLabels:
      run: {{.Run}}
  template:
    metadata:
      labels:
        run: {{.Run}}
    spec:
      volumes:
      imagePullSecrets:
      - name: {{.RegistrySecret}}
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
