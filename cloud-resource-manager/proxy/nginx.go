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

func InitL7Proxy(ctx context.Context, client pc.PlatformClient, ops ...Op) error {
	out, err := client.Output("docker ps --format '{{.Names}}'")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == NginxL7Name {
			// already running
			return nil
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

func CreateNginxProxy(ctx context.Context, client pc.PlatformClient, name, listenIP, destIP string, ports []dme.AppPort, ops ...Op) error {
	// check to see whether nginx or envoy is needed (or both)
	envoyNeeded, nginxNeeded := CheckProtocols(name, ports)
	if envoyNeeded {
		err := CreateEnvoyProxy(ctx, client, name, listenIP, destIP, ports, ops...)
		if err != nil {
			return fmt.Errorf("Create Envoy Proxy failed, %v", err)
		}
	}
	if !nginxNeeded {
		return nil
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "create nginx", "name", name, "listenIP", listenIP, "destIP", destIP, "ports", ports)
	opts := Options{}
	opts.Apply(ops)

	out, err := client.Output("pwd")
	if err != nil {
		return err
	}
	pwd := strings.TrimSpace(string(out))

	dir := pwd + "/nginx/" + name
	log.SpanLog(ctx, log.DebugLevelMexos, "nginx remote dir", "name", name, "dir", dir)
	l7dir := pwd + "/" + NginxL7Dir

	err = pc.Run(client, "mkdir -p "+dir)
	if err != nil {
		return err
	}

	useTLS := false
	err = pc.Run(client, "ls cert.pem && ls key.pem")
	if err == nil {
		useTLS = true
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "nginx certs check",
		"name", name, "useTLS", useTLS)

	errlogFile := dir + "/err.log"
	err = pc.Run(client, "touch "+errlogFile)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos,
			"nginx %s can't create file %s", name, errlogFile)
		return err
	}
	accesslogFile := dir + "/access.log"
	err = pc.Run(client, "touch "+accesslogFile)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos,
			"nginx %s can't create file %s", name, accesslogFile)
		return err
	}
	nconfName := dir + "/nginx.conf"
	err = createNginxConf(ctx, client, nconfName, name, l7dir, listenIP, destIP, ports, useTLS)
	if err != nil {
		return fmt.Errorf("create nginx.conf failed, %v", err)
	}

	cmdArgs := []string{"run", "-d", "-l edge-cloud", "--restart=unless-stopped", "--name", "nginx"+name}
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
	if useTLS {
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
	log.SpanLog(ctx, log.DebugLevelMexos, "nginx docker command", "name", name,
		"cmd", cmd)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't create nginx container %s, %s, %v", name, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "created nginx container", "name",  "nginx"+name)
	return nil
}

func createNginxConf(ctx context.Context, client pc.PlatformClient, confname, name, l7dir, listenIP, backendIP string, ports []dme.AppPort, useTLS bool) error {
	spec := ProxySpec{
		Name:       name,
		UseTLS:     useTLS,
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
				ListenIP: listenIP,
				ListenPort:       p.PublicPort,
				BackendIP:     backendIP,
				BackendPort:       p.InternalPort,
				PathPrefix: p.PathPrefix,
			}
			httpPorts = append(httpPorts, httpPort)
			continue
		case dme.LProto_L_PROTO_TCP:
			continue // have envoy handle the tcp stuff
		case dme.LProto_L_PROTO_UDP:
			udpPort := UDPSpecDetail{
				ListenIP: listenIP,
				ListenPort:       p.PublicPort,
				BackendIP:     backendIP,
				BackendPort:       p.InternalPort,
				ConcurrentConnsPerIP: udpconns,
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

	log.SpanLog(ctx, log.DebugLevelMexos, "create nginx conf", "name", name)
	buf := bytes.Buffer{}
	err = nginxConfT.Execute(&buf, &spec)
	if err != nil {
		return err
	}
	err = pc.WriteFile(client, confname, buf.String(), "nginx.conf", pc.NoSudo)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "write nginx.conf failed",
			"name", name, "err", err)
		return err
	}

	if len(httpPorts) > 0 {
		// add L7 config to L7 nginx instance
		log.SpanLog(ctx, log.DebugLevelMexos, "create L7 nginx conf", "name", name)
		buf := bytes.Buffer{}
		err = nginxL7ConfT.Execute(&buf, httpPorts)
		if err != nil {
			return err
		}
		err = pc.WriteFile(client, l7dir+"/"+name+".conf", buf.String(), "nginx L7 conf", pc.NoSudo)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMexos,
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

func reloadNginxL7(client pc.PlatformClient) error {
	err := pc.Run(client, "docker exec "+NginxL7Name+" nginx -s reload")
	if err != nil {
		log.DebugLog(log.DebugLevelMexos,
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
	UseTLS     bool // To be removed
	MetricPort int32
	Cert       *access.TLSCert
}

type TCPSpecDetail struct {
	ListenIP     string
	ListenPort   int32
	BackendIP     string
	BackendPort      int32
	ConcurrentConns uint64
}

type UDPSpecDetail struct {
	ListenIP     string
	ListenPort   int32
	BackendIP     string
	BackendPort      int32
	ConcurrentConnsPerIP uint64
}

type HTTPSpecDetail struct {
	ListenIP     string
	ListenPort   int32
	BackendIP     string
	BackendPort      int32
	PathPrefix string
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
        listen {{.L7Port}}{{if .UseTLS}} ssl{{end}};
{{- if .UseTLS}}
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
		listen {{.ListenPort}} udp;
		proxy_pass {{.BackendIP}}:{{.ListenPort}};
	}
	{{- end}}
}
{{- end}}
`

var nginxL7Conf = `
{{- range .}}
location /{{.PathPrefix}}/ {
	proxy_pass http://{{.BackendIP}}/;
}
{{- end}}
`

func DeleteNginxProxy(ctx context.Context, client pc.PlatformClient, name string) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "delete nginx", "name", name)
	out, err := client.Output("docker kill " +  "nginx"+name)
	deleteContainer := false
	if err == nil {
		deleteContainer = true
	} else {
		if strings.Contains(string(out), "No such container") {
			log.SpanLog(ctx, log.DebugLevelMexos,
				"nginx LB container already gone", "name",  "nginx"+name)
		} else {
			return fmt.Errorf("can't delete nginx container %s, %s, %v", name, out, err)
		}
	}
	l7conf := NginxL7Dir + "/" + name + ".conf"
	out, err = client.Output("rm " + l7conf)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "delete nginx L7 conf",
			"name", name, "l7conf", l7conf, "out", out, "err", err)
	}
	nginxDir := "nginx/" + name
	out, err = client.Output("rm -rf " + nginxDir)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "delete nginx dir", "name", name, "dir", nginxDir, "out", out, "err", err)
	}
	if deleteContainer {
		out, err = client.Output("docker rm " +  "nginx"+name)
		if err != nil && !strings.Contains(string(out), "No such container") {
			return fmt.Errorf("can't remove nginx container %s, %s, %v", name, out, err)
		}
	}
	reloadNginxL7(client)

	log.SpanLog(ctx, log.DebugLevelMexos, "deleted nginx", "name",  "nginx"+name)
	DeleteEnvoyProxy(ctx, client, name)
	return nil
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
