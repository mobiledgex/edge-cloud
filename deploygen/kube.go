package deploygen

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"

	"github.com/mobiledgex/edge-cloud/util"
)

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
	Tls       bool
	Nginx     bool
}

func kubeBasic(app *AppSpec) (string, error) {
	gen := kubeBasicGen{
		app:   app,
		files: []string{},
		ports: setKubePorts(app.Ports),
	}
	gen.kubeLb([]string{"tcp"})
	gen.kubeLb([]string{"udp"})
	gen.kubeApp()
	if gen.err != nil {
		return "", gen.err
	}
	return strings.Join(gen.files, "---\n"), nil
}

func appID(app *AppSpec) string {
	return app.Name + app.Version
}

// Translate PortSpec objects to the KubePort representation.
// util.PortSpec supports port range short hand, as does dme.AppPort.
// KubePort does not, not yet anyway. So for now, to
// preserve the LBs current notion of reality, we detect the
// presence of a port range, and if found exhaustive enumerate
// each port in range with a suitable KubePort object.
//
func setKubePorts(ports []util.PortSpec) []kubePort {
	kports := []kubePort{}
	var kp kubePort
	for _, port := range ports {
		endPort, _ := strconv.ParseInt(port.EndPort, 10, 32)
		if endPort != 0 { // PortSpec port-range short hand notation,
			// exhaustively enumerate each as a kp
			start, _ := strconv.ParseInt(port.Port, 10, 32) // we sanitized in objs.go
			end, _ := strconv.ParseInt(port.EndPort, 10, 32)
			for i := start; i <= end; i++ {
				p := strconv.Itoa(int(i))
				kp = kubePort{
					Proto: strings.ToLower(port.Proto),
					Port:  p,
				}
				switch port.Proto {
				case "tcp":
					kp.KubeProto = "TCP"
				case "udp":
					kp.KubeProto = "UDP"
				}
				kp.Tls = port.Tls
				kp.Nginx = port.Nginx
				kports = append(kports, kp)
			}
		} else {
			// nominal non-range
			kp = kubePort{
				Proto: strings.ToLower(port.Proto),
				Port:  port.Port,
			}
			switch port.Proto {
			case "tcp":
				kp.KubeProto = "TCP"
			case "udp":
				kp.KubeProto = "UDP"
			}
			kp.Tls = port.Tls
			kp.Nginx = port.Nginx
			kports = append(kports, kp)
		}
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
		Name:  util.DNSSanitize(appID(g.app) + "-" + protos[0]),
		Run:   util.K8SSanitize(appID(g.app)),
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
  - name: {{.Proto}}{{.Port}}{{if .Tls}}tls{{end}}{{if .Nginx}}nginx{{end}}
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
	registrySecret := g.app.ImageHost
	hostname := strings.Split(g.app.ImageHost, ":")
	if len(hostname) > 1 {
		// docker-registry secret name use for
		// "kubectl create secret docker-registry" cannot include ":"
		registrySecret = hostname[0] + "-" + hostname[1]
	}
	data := appData{
		Name:           util.DNSSanitize(appID(g.app) + "-deployment"),
		DNSName:        util.DNSSanitize(appID(g.app)),
		Run:            util.K8SSanitize(appID(g.app)),
		Ports:          g.ports,
		ImagePath:      g.app.ImagePath,
		Command:        cs,
		RegistrySecret: registrySecret,
		MexDeployGen:   MexDeployGenLabel,
	}
	buf := bytes.Buffer{}
	if g.app.ScaleWithCluster {
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
	MexDeployGen   string
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
        {{.MexDeployGen}}: kubernetes-basic
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
