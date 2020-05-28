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
// L7 proxying is handled by a single L7 nginx instance since only
// one instance can bind to port 443.
// All other access to an AppInst is handled by a per-AppInst dedicated
// nginx instance for better isolation.

var NginxL7Dir = "nginxL7"
var NginxL7Name = "nginxL7"

var nginxConfT *template.Template
var nginxL7ConfT *template.Template

// defaultConcurrentConnsPerIP is the default DOS protection setting for connections per source IP
const defaultConcurrentConnsPerIP uint64 = 100

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

func getUDPConcurrentConnectionsPerIP() (uint64, error) {
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
	nginxL7ConfT = template.Must(template.New("l7app").Parse(nginxL7Conf))
}

func InitL7Proxy(ctx context.Context, client ssh.Client, ops ...Op) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitL7Proxy")
	out, err := client.Output("docker ps -a --format '{{.Names}},{{.Status}}'")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Split(line, ",")
		if len(fields) == 2 {
			containername := fields[0]
			status := fields[1]
			if containername == NginxL7Name {
				if strings.HasPrefix(status, "Up") {
					log.SpanLog(ctx, log.DebugLevelInfra, "L7Proxy already running")
					return nil
				}
				log.SpanLog(ctx, log.DebugLevelInfra, "L7Proxy present but not running, remove it", "status", status)
				out, err := client.Output("docker rm -f " + containername)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "unable to remove old L7Proxy", "out", out, "err", err)
					return fmt.Errorf("Unable to remove old L7Proxy")
				}
				break
			}
		}
	}
	listenIP := ""
	backendIP := ""
	return CreateNginxProxy(ctx, client, NginxL7Name, listenIP, backendIP, []dme.AppPort{}, ops...)
}

func CheckProtocols(name string, ports []dme.AppPort) (bool, bool) {
	needEnvoy := false
	needNginx := false
	for _, p := range ports {
		switch p.Proto {
		case dme.LProto_L_PROTO_HTTP:
			needNginx = true
		case dme.LProto_L_PROTO_TCP:
			needEnvoy = true // have envoy handle the tcp stuff
		case dme.LProto_L_PROTO_UDP:
			needNginx = true
		}
	}
	if name == NginxL7Name {
		needNginx = true
	}
	return needEnvoy, needNginx
}

func getNginxContainerName(name string) string {
	if name == NginxL7Name {
		return name
	}
	return "nginx" + name
}

func CreateNginxProxy(ctx context.Context, client ssh.Client, name, listenIP, destIP string, ports []dme.AppPort, ops ...Op) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "CreateNginxProxy", "listenIP", listenIP, "destIP", destIP)
	containerName := getNginxContainerName(name)

	// check to see whether nginx or envoy is needed (or both)
	envoyNeeded, nginxNeeded := CheckProtocols(name, ports)
	if envoyNeeded {
		err := CreateEnvoyProxy(ctx, client, name, listenIP, destIP, ports, ops...)
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
	l7dir := pwd + "/" + NginxL7Dir

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
	err = createNginxConf(ctx, client, nconfName, name, l7dir, listenIP, destIP, ports, usesTLS)
	if err != nil {
		return fmt.Errorf("create nginx.conf failed, %v", err)
	}

	cmdArgs := []string{"run", "-d", "-l edge-cloud", "--restart=unless-stopped", "--name", containerName}
	if opts.DockerPublishPorts {
		cmdArgs = append(cmdArgs, dockermgmt.GetDockerPortString(ports, dockermgmt.UsePublicPortInContainer, dme.LProto_L_PROTO_UDP, listenIP)...)
		if name == NginxL7Name {
			// Special case. When the L7 nginx instance is created,
			// there are no configs yet for L7. Expose the L7 port manually.
			pstr := fmt.Sprintf("%d:%d/tcp", cloudcommon.RootLBL7Port, cloudcommon.RootLBL7Port)
			cmdArgs = append(cmdArgs, "-p", pstr)
		}
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
		"-v", l7dir + ":/etc/nginx/L7",
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

func createNginxConf(ctx context.Context, client ssh.Client, confname, name, l7dir, listenIP, backendIP string, ports []dme.AppPort, usesTLS bool) error {
	spec := ProxySpec{
		Name:       name,
		UsesTLS:    usesTLS,
		MetricPort: cloudcommon.ProxyMetricsPort,
	}
	httpPorts := []HTTPSpecDetail{}

	udpconns, err := getUDPConcurrentConnectionsPerIP()
	if err != nil {
		return err
	}
	for _, p := range ports {
		switch p.Proto {
		case dme.LProto_L_PROTO_HTTP:
			httpPort := HTTPSpecDetail{
				ListenIP:    listenIP,
				ListenPort:  p.PublicPort,
				BackendIP:   backendIP,
				BackendPort: p.InternalPort,
				PathPrefix:  p.PathPrefix,
			}
			httpPorts = append(httpPorts, httpPort)
			continue
		case dme.LProto_L_PROTO_TCP:
			continue // have envoy handle the tcp stuff
		case dme.LProto_L_PROTO_UDP:
			udpPort := UDPSpecDetail{
				ListenIP:             listenIP,
				BackendIP:            backendIP,
				BackendPort:          p.InternalPort,
				ConcurrentConnsPerIP: udpconns,
			}
			endPort := p.EndPort
			if endPort == 0 {
				endPort = p.PublicPort
			} else {
				// if we have a port range, the internal ports and external ports must match
				if p.InternalPort != p.PublicPort {
					return fmt.Errorf("public and internal ports must match when port range in use")
				}
			}
			for pnum := p.PublicPort; pnum <= endPort; pnum++ {
				udpPort.ListenPorts = append(udpPort.ListenPorts, pnum)
			}
			// if there is more than one listen port, we don't use the backend port as the
			// listen port is used as the backend port in the case of a range
			if len(udpPort.ListenPorts) > 1 {
				udpPort.BackendPort = 0
			}
			spec.UDPSpec = append(spec.UDPSpec, &udpPort)
			spec.L4 = true
		}
	}

	if name == NginxL7Name {
		// this is the L7 nginx instance, generate L7 config
		spec.L4 = false
		spec.L7 = true
		spec.L7Port = cloudcommon.RootLBL7Port
		// this is where all other Apps will store their L7 config
		err := pc.Run(client, "mkdir -p "+l7dir)
		if err != nil {
			return err
		}
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

	if len(httpPorts) > 0 {
		// add L7 config to L7 nginx instance
		log.SpanLog(ctx, log.DebugLevelInfra, "create L7 nginx conf", "name", name)
		buf := bytes.Buffer{}
		err = nginxL7ConfT.Execute(&buf, httpPorts)
		if err != nil {
			return err
		}
		err = pc.WriteFile(client, l7dir+"/"+name+".conf", buf.String(), "nginx L7 conf", pc.NoSudo)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra,
				"write nginx L7 conf failed",
				"name", name, "err", err)
			return err
		}
		err = reloadNginxL7(client)
		if err != nil {
			return err
		}
	}
	return nil
}

func reloadNginxL7(client ssh.Client) error {
	err := pc.Run(client, "docker exec "+NginxL7Name+" nginx -s reload")
	if err != nil {
		log.DebugLog(log.DebugLevelInfra,
			"reload L7 nginx config failed", "err", err)
	}
	return err
}

type ProxySpec struct {
	Name       string
	L4, L7     bool
	UDPSpec    []*UDPSpecDetail
	TCPSpec    []*TCPSpecDetail
	L7Port     int32
	UsesTLS    bool // To be removed
	MetricPort int32
	CertName   string
}

type TCPSpecDetail struct {
	ListenIP        string
	ListenPort      int32
	BackendIP       string
	BackendPort     int32
	ConcurrentConns uint64
	UseTLS          bool // for port specific TLS termination
}

type UDPSpecDetail struct {
	ListenIP             string
	ListenPorts          []int32
	BackendIP            string
	BackendPort          int32
	ConcurrentConnsPerIP uint64
}

type HTTPSpecDetail struct {
	ListenIP    string
	ListenPort  int32
	BackendIP   string
	BackendPort int32
	PathPrefix  string
}

var nginxConf = `
user  nginx;
worker_processes  1;

error_log  /var/log/nginx/error.log warn;
pid        /var/run/nginx.pid;

events {
    worker_connections  1024;
}

{{if .L7 -}}
http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
    access_log  /var/log/nginx/access.log  main;
    keepalive_timeout  65;
    server_tokens off;
    server {
        listen {{.L7Port}}{{if .UsesTLS}} ssl{{end}};
{{- if .UsesTLS}}
        ssl_certificate        /etc/ssl/certs/server.crt;
        ssl_certificate_key    /etc/ssl/certs/server.key;
{{- end}}
        include /etc/nginx/L7/*.conf;
	}
    server {
        listen 127.0.0.1:{{.MetricPort}};
        server_name 127.0.0.1:{{.MetricPort}};
        location /nginx_metrics {
            stub_status;
            allow 127.0.0.1;
            deny all;
		}
	}
}
{{- end}}

{{if .L4 -}}
stream { 
    limit_conn_zone $binary_remote_addr zone=ipaddr:10m;
	{{- range .UDPSpec}}
	server {
		limit_conn ipaddr {{.ConcurrentConnsPerIP}}; 
		{{range $portnum := .ListenPorts}}
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
{{- end}}
`

var nginxL7Conf = `
{{- range .}}
location /{{.PathPrefix}}/ {
	proxy_pass http://{{.BackendIP}}:{{.BackendPort}}/;
}
{{- end}}
`

func DeleteNginxProxy(ctx context.Context, client ssh.Client, name string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "delete nginx", "name", name)
	containerName := getNginxContainerName(name)
	out, err := client.Output("docker kill " + containerName)
	log.SpanLog(ctx, log.DebugLevelInfra, "kill nginx result", "out", out, "err", err)

	l7conf := NginxL7Dir + "/" + name + ".conf"
	out, err = client.Output("rm " + l7conf)
	log.SpanLog(ctx, log.DebugLevelInfra, "delete nginx L7 conf result",
		"name", name, "l7conf", l7conf, "out", out, "err", err)

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

	reloadNginxL7(client)
	log.SpanLog(ctx, log.DebugLevelInfra, "deleted nginx", "containerName", containerName)
	return DeleteEnvoyProxy(ctx, client, name)
}

type Options struct {
	DockerPublishPorts bool
	DockerNetwork      string
	Cert               *access.TLSCert
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

func (o *Options) Apply(ops []Op) {
	for _, op := range ops {
		op(o)
	}
}
