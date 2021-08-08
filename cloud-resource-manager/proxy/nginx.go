package proxy

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/access"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/dockermgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

// Nginx is used to proxy connections from the external network to
// the internal cloudlet clusters. It is not doing any load-balancing.
// Access to an AppInst is handled by a per-AppInst dedicated
// nginx instance for better isolation.

var NginxL7Name = "nginxL7"

var nginxConfT *template.Template

// defaultConcurrentConnsPerIP is the default DOS protection setting for connections per source IP
const defaultConcurrentConnsPerIP uint64 = 100
const defaultWorkerConns int = 1024

// TCP is in envoy, which does not have concurrent connections per IP, but rather
// just concurrent connections overall
func getTCPConcurrentConnections() (uint64, error) {
	var err error
	connStr := os.Getenv("MEX_LB_CONCURRENT_TCP_CONNS")
	conns := defaultConcurrentConns
	if connStr != "" {
		conns, err = strconv.ParseUint(connStr, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	return conns, nil
}

func getUDPConcurrentConnections() (uint64, error) {
	var err error
	connStr := os.Getenv("MEX_LB_CONCURRENT_UDP_CONNS")
	conns := defaultConcurrentConnsPerIP
	if connStr != "" {
		conns, err = strconv.ParseUint(connStr, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	return conns, nil
}

func init() {
	nginxConfT = template.Must(template.New("conf").Parse(nginxConf))
}

// This actually deletes the L7 proxy for backwards compatibility, so it doesnt conflict with tcp:443 apps
func InitL7Proxy(ctx context.Context, client ssh.Client, ops ...Op) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitL7Proxy")

	out, err := client.Output("docker kill " + NginxL7Name)
	log.SpanLog(ctx, log.DebugLevelInfra, "kill nginx result", "out", out, "err", err)

	if err != nil && strings.Contains(string(out), "No such container") {
		return nil
	}
	// container should autoremove on kill but just in case
	out, err = client.Output("docker rm " + NginxL7Name)
	log.SpanLog(ctx, log.DebugLevelInfra, "rm nginx result", "out", out, "err", err)
	if err != nil && strings.Contains(string(out), "No such container") {
		return nil
	}
	return err
}

func CheckProtocols(name string, ports []dme.AppPort) (bool, bool) {
	needEnvoy := false
	needNginx := false
	for _, p := range ports {
		switch p.Proto {
		case dme.LProto_L_PROTO_TCP:
			needEnvoy = true
		case dme.LProto_L_PROTO_UDP:
			if p.Nginx {
				needNginx = true
			} else {
				needEnvoy = true
			}
		}
	}
	return needEnvoy, needNginx
}

func getNginxContainerName(name string) string {
	return "nginx" + name
}

func CreateNginxProxy(ctx context.Context, client ssh.Client, name, listenIP, destIP string, ports []dme.AppPort, skipHcPorts string, ops ...Op) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "CreateNginxProxy", "listenIP", listenIP, "destIP", destIP)
	containerName := getNginxContainerName(name)

	// check to see whether nginx or envoy is needed (or both)
	envoyNeeded, nginxNeeded := CheckProtocols(name, ports)
	if envoyNeeded {
		err := CreateEnvoyProxy(ctx, client, name, listenIP, destIP, ports, skipHcPorts, ops...)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "CreateEnvoyProxy failed ", "err", err)
			return fmt.Errorf("Create Envoy Proxy failed, %v", err)
		}
	}
	if !nginxNeeded {
		return nil
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "create nginx", "name", name, "listenIP", listenIP, "destIP", destIP, "ports", ports)
	opts := Options{}
	opts.Apply(ops)

	out, err := client.Output("pwd")
	if err != nil {
		return err
	}
	pwd := strings.TrimSpace(string(out))

	dir := pwd + "/nginx/" + name
	log.SpanLog(ctx, log.DebugLevelInfra, "nginx remote dir", "name", name, "dir", dir)

	err = pc.Run(client, "mkdir -p "+dir)
	if err != nil {
		return err
	}

	usesTLS := false
	err = pc.Run(client, "ls cert.pem && ls key.pem")
	if err == nil {
		usesTLS = true
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "nginx certs check",
		"name", name, "usesTLS", usesTLS)

	errlogFile := dir + "/err.log"
	err = pc.Run(client, "touch "+errlogFile)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra,
			"nginx %s can't create file %s", name, errlogFile)
		return err
	}
	accesslogFile := dir + "/access.log"
	err = pc.Run(client, "touch "+accesslogFile)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra,
			"nginx %s can't create file %s", name, accesslogFile)
		return err
	}
	nconfName := dir + "/nginx.conf"
	err = createNginxConf(ctx, client, nconfName, name, listenIP, destIP, ports, usesTLS)
	if err != nil {
		return fmt.Errorf("create nginx.conf failed, %v", err)
	}

	cmdArgs := []string{"run", "-d", "-l edge-cloud", "--restart=unless-stopped", "--name", containerName}
	if opts.DockerPublishPorts {
		cmdArgs = append(cmdArgs, dockermgmt.GetDockerPortString(ports, dockermgmt.UsePublicPortInContainer, dockermgmt.NginxProxy, listenIP)...)
	}
	if opts.DockerNetwork != "" {
		// For dind, we use the network which the dind cluster is on.
		cmdArgs = append(cmdArgs, "--network", opts.DockerNetwork)
	}
	if usesTLS {
		cmdArgs = append(cmdArgs, "-v", pwd+"/cert.pem:/etc/ssl/certs/server.crt")
		cmdArgs = append(cmdArgs, "-v", pwd+"/key.pem:/etc/ssl/certs/server.key")
	}
	cmdArgs = append(cmdArgs, []string{
		"-v", dir + ":/var/www/.cache",
		"-v", "/etc/ssl/certs:/etc/ssl/certs",
		"-v", errlogFile + ":/var/log/nginx/error.log",
		"-v", accesslogFile + ":/var/log/nginx/access.log",
		"-v", nconfName + ":/etc/nginx/nginx.conf",
		"docker.mobiledgex.net/mobiledgex/mobiledgex_public/nginx-with-curl"}...)
	cmd := "docker " + strings.Join(cmdArgs, " ")
	log.SpanLog(ctx, log.DebugLevelInfra, "nginx docker command", "containerName", containerName,
		"cmd", cmd)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't create nginx container %s, %s, %v", name, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "created nginx container", "containerName", containerName)
	return nil
}

func createNginxConf(ctx context.Context, client ssh.Client, confname, name, listenIP, backendIP string, ports []dme.AppPort, usesTLS bool) error {
	spec := ProxySpec{
		Name:        name,
		UsesTLS:     usesTLS,
		MetricIP:    listenIP,
		MetricPort:  cloudcommon.ProxyMetricsPort,
		WorkerConns: defaultWorkerConns,
	}
	portCount := 0

	udpconns, err := getUDPConcurrentConnections()
	if err != nil {
		return err
	}
	for _, p := range ports {
		if p.Proto == dme.LProto_L_PROTO_UDP {
			if !p.Nginx { // use envoy
				continue
			}
			udpPort := UDPSpecDetail{
				ListenIP:        listenIP,
				BackendIP:       backendIP,
				BackendPort:     p.InternalPort,
				ConcurrentConns: udpconns,
			}
			endPort := p.EndPort
			if endPort == 0 {
				endPort = p.PublicPort
				portCount = portCount + 1
			} else {
				portCount = int((p.EndPort-p.InternalPort)+1) + portCount
				// if we have a port range, the internal ports and external ports must match
				if p.InternalPort != p.PublicPort {
					return fmt.Errorf("public and internal ports must match when port range in use")
				}
			}
			for pnum := p.PublicPort; pnum <= endPort; pnum++ {
				udpPort.NginxListenPorts = append(udpPort.NginxListenPorts, pnum)
			}
			// if there is more than one listen port, we don't use the backend port as the
			// listen port is used as the backend port in the case of a range
			if len(udpPort.NginxListenPorts) > 1 {
				udpPort.BackendPort = 0
			}
			spec.UDPSpec = append(spec.UDPSpec, &udpPort)
		}
	}
	// need to have more worker connections than ports otherwise nginx will crash
	if portCount > 1000 {
		spec.WorkerConns = int(float64(portCount) * 1.2)
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "create nginx conf", "name", name)
	buf := bytes.Buffer{}
	err = nginxConfT.Execute(&buf, &spec)
	if err != nil {
		return err
	}
	err = pc.WriteFile(client, confname, buf.String(), "nginx.conf", pc.NoSudo)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "write nginx.conf failed",
			"name", name, "err", err)
		return err
	}
	return nil
}

type ProxySpec struct {
	Name        string
	UDPSpec     []*UDPSpecDetail
	TCPSpec     []*TCPSpecDetail
	UsesTLS     bool // To be removed
	MetricIP    string
	MetricPort  int32
	CertName    string
	WorkerConns int
}

type TCPSpecDetail struct {
	ListenIP        string
	ListenPort      int32
	BackendIP       string
	BackendPort     int32
	ConcurrentConns uint64
	UseTLS          bool // for port specific TLS termination
	HealthCheck     bool
}

type UDPSpecDetail struct {
	ListenIP         string
	ListenPort       int32
	NginxListenPorts []int32
	BackendIP        string
	BackendPort      int32
	ConcurrentConns  uint64
	MaxPktSize       int64
}

var nginxConf = `
user  nginx;
worker_processes  1;

error_log  /var/log/nginx/error.log warn;
pid        /var/run/nginx.pid;

events {
    worker_connections  {{.WorkerConns}};
}

stream { 
    limit_conn_zone $binary_remote_addr zone=ipaddr:10m;
	{{- range .UDPSpec}}
	server {
		limit_conn ipaddr {{.ConcurrentConns}}; 
		{{range $portnum := .NginxListenPorts}}
		listen {{$portnum}} udp; 
		{{end}}
		{{if eq .BackendPort 0}}
		proxy_pass {{.BackendIP}}:$server_port;
		{{- end}}
		{{if ne .BackendPort 0}}
		proxy_pass {{.BackendIP}}:{{.BackendPort}};
		{{- end}}
	}
	{{- end}}
}
`

func DeleteNginxProxy(ctx context.Context, client ssh.Client, name string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "delete nginx", "name", name)
	containerName := getNginxContainerName(name)
	out, err := client.Output("docker kill " + containerName)
	log.SpanLog(ctx, log.DebugLevelInfra, "kill nginx result", "out", out, "err", err)

	nginxDir := "nginx/" + name
	out, err = client.Output("rm -rf " + nginxDir)
	log.SpanLog(ctx, log.DebugLevelInfra, "delete nginx dir result", "name", name, "dir", nginxDir, "out", out, "err", err)

	out, err = client.Output("docker rm -f " + containerName)
	log.SpanLog(ctx, log.DebugLevelInfra, "rm nginx result", "out", out, "err", err)
	if err != nil && !strings.Contains(string(out), "No such container") {
		// delete the envoy proxy for best effort
		DeleteEnvoyProxy(ctx, client, name)
		return fmt.Errorf("can't remove nginx container %s, %s, %v", name, out, err)
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "deleted nginx", "containerName", containerName)
	return DeleteEnvoyProxy(ctx, client, name)
}

type Options struct {
	DockerPublishPorts bool
	DockerNetwork      string
	Cert               *access.TLSCert
	DockerUser         string
	MetricIP           string
}

type Op func(opts *Options)

func WithDockerNetwork(network string) Op {
	return func(opts *Options) {
		opts.DockerNetwork = network
	}
}

func WithDockerPublishPorts() Op {
	return func(opts *Options) {
		opts.DockerPublishPorts = true
	}
}

func WithTLSCert(cert *access.TLSCert) Op {
	return func(opts *Options) {
		opts.Cert = cert
		opts.Cert.CommonName = strings.Replace(opts.Cert.CommonName, "*", "_", 1)
	}
}

func WithDockerUser(user string) Op {
	return func(opts *Options) {
		opts.DockerUser = user
	}
}

func WithMetricIP(addr string) Op {
	return func(opts *Options) {
		opts.MetricIP = addr
	}
}

func (o *Options) Apply(ops []Op) {
	for _, op := range ops {
		op(o)
	}
}
