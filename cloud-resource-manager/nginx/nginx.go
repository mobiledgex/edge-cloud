package nginx

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

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

func init() {
	nginxConfT = template.Must(template.New("conf").Parse(nginxConf))
	nginxL7ConfT = template.Must(template.New("l7app").Parse(nginxL7Conf))
}

func InitL7Proxy(client pc.PlatformClient, ops ...Op) error {
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
	return CreateNginxProxy(client, NginxL7Name, "", []dme.AppPort{}, ops...)
}

func CreateNginxProxy(client pc.PlatformClient, name, originIP string, ports []dme.AppPort, ops ...Op) error {
	log.DebugLog(log.DebugLevelMexos, "create nginx", "name", name, "originip", originIP, "ports", ports)
	opts := Options{}
	opts.Apply(ops)

	out, err := client.Output("pwd")
	if err != nil {
		return err
	}
	pwd := strings.TrimSpace(string(out))

	dir := pwd + "/nginx/" + name
	log.DebugLog(log.DebugLevelMexos, "nginx remote dir", "name", name, "dir", dir)
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
	log.DebugLog(log.DebugLevelMexos, "nginx certs check",
		"name", name, "useTLS", useTLS)

	errlogFile := dir + "/err.log"
	err = pc.Run(client, "touch "+errlogFile)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos,
			"nginx %s can't create file %s", name, errlogFile)
		return err
	}
	accesslogFile := dir + "/access.log"
	err = pc.Run(client, "touch "+accesslogFile)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos,
			"nginx %s can't create file %s", name, accesslogFile)
		return err
	}
	nconfName := dir + "/nginx.conf"
	err = createNginxConf(client, nconfName, name, l7dir, originIP, ports, useTLS)
	if err != nil {
		return fmt.Errorf("create nginx.conf failed, %v", err)
	}

	cmdArgs := []string{"run", "-d", "--restart=unless-stopped", "--name", name}
	if opts.DockerPublishPorts {
		for _, p := range ports {
			if p.Proto == dme.LProto_L_PROTO_HTTP {
				// L7 is handled by the L7 instance
				continue
			}
			proto := "tcp"
			if p.Proto == dme.LProto_L_PROTO_UDP {
				proto = "udp"
			}
			pstr := fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PublicPort, proto)
			cmdArgs = append(cmdArgs, "-p", pstr)
		}
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
		"nginx"}...)
	cmd := "docker " + strings.Join(cmdArgs, " ")
	log.DebugLog(log.DebugLevelMexos, "nginx docker command", "name", name,
		"cmd", cmd)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't create nginx container %s, %s, %v", name, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "created nginx container", "name", name)
	return nil
}

func createNginxConf(client pc.PlatformClient, confname, name, l7dir, originIP string, ports []dme.AppPort, useTLS bool) error {
	spec := ProxySpec{
		Name:   	name,
		UseTLS: 	useTLS,
		MetricPort: cloudcommon.NginxMetricsPort,
	}
	httpPorts := []HTTPSpecDetail{}

	for _, p := range ports {
		origin := fmt.Sprintf("%s:%d", originIP, p.InternalPort)
		switch p.Proto {
		case dme.LProto_L_PROTO_HTTP:
			httpPort := HTTPSpecDetail{
				Port:       p.PublicPort,
				PathPrefix: p.PathPrefix,
				Origin:     origin,
			}
			httpPorts = append(httpPorts, httpPort)
			continue
		case dme.LProto_L_PROTO_TCP:
			tcpPort := TCPSpecDetail{
				Port:   p.PublicPort,
				Origin: origin,
			}
			spec.TCPSpec = append(spec.TCPSpec, &tcpPort)
			spec.L4 = true
		case dme.LProto_L_PROTO_UDP:
			udpPort := UDPSpecDetail{
				Port:   p.PublicPort,
				Origin: origin,
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

	log.DebugLog(log.DebugLevelMexos, "create nginx conf", "name", name)
	buf := bytes.Buffer{}
	err := nginxConfT.Execute(&buf, &spec)
	if err != nil {
		return err
	}
	err = pc.WriteFile(client, confname, buf.String(), "nginx.conf")
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "write nginx.conf failed",
			"name", name, "err", err)
		return err
	}

	if len(httpPorts) > 0 {
		// add L7 config to L7 nginx instance
		log.DebugLog(log.DebugLevelMexos, "create L7 nginx conf", "name", name)
		buf := bytes.Buffer{}
		err = nginxL7ConfT.Execute(&buf, httpPorts)
		if err != nil {
			return err
		}
		err = pc.WriteFile(client, l7dir+"/"+name+".conf", buf.String(), "nginx L7 conf")
		if err != nil {
			log.DebugLog(log.DebugLevelMexos,
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
	Name    	string
	L4, L7  	bool
	UDPSpec 	[]*UDPSpecDetail
	TCPSpec 	[]*TCPSpecDetail
	L7Port  	int32
	UseTLS  	bool
	MetricPort  int32
}

type TCPSpecDetail struct {
	Port   int32
	Origin string
}

type UDPSpecDetail struct {
	Port   int32
	Origin string
}

type HTTPSpecDetail struct {
	Port       int32
	PathPrefix string
	Origin     string
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
	{{- range .TCPSpec}}
	server {
		listen {{.Port}};
		proxy_pass {{.Origin}};
	}
	{{- end}}
	{{- range .UDPSpec}}
	server {
		listen {{.Port}} udp;
		proxy_pass {{.Origin}};
	}
	{{- end}}
}
http {
	server {
		listen 127.0.0.1:{{.MetricPort}};
		server_name 127.0.0.1:{{.MetricPort}};
	    location /nginx_metrics {
			allow 127.0.0.1;
			deny all;
			stub_status;
    	}
	}
}
{{- end}}
`

var nginxL7Conf = `
{{- range .}}
location /{{.PathPrefix}}/ {
	proxy_pass http://{{.Origin}}/;
}
{{- end}}
`

func DeleteNginxProxy(client pc.PlatformClient, name string) error {
	log.DebugLog(log.DebugLevelMexos, "delete nginx", "name", name)
	out, err := client.Output("docker kill " + name)
	deleteContainer := false
	if err == nil {
		deleteContainer = true
	} else {
		if strings.Contains(string(out), "No such container") {
			log.DebugLog(log.DebugLevelMexos,
				"nginx LB container already gone", "name", name)
		} else {
			return fmt.Errorf("can't delete nginx container %s, %s, %v", name, out, err)
		}
	}
	l7conf := NginxL7Dir + "/" + name + ".conf"
	out, err = client.Output("rm " + l7conf)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "delete nginx L7 conf",
			"name", name, "l7conf", l7conf, "out", out, "err", err)
	}
	nginxDir := "nginx/" + name
	out, err = client.Output("rm -rf " + nginxDir)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "delete nginx dir", "name", name, "dir", nginxDir, "out", out, "err", err)
	}
	if deleteContainer {
		out, err = client.Output("docker rm " + name)
		if err != nil {
			return fmt.Errorf("can't remove nginx container %s, %s, %v", name, out, err)
		}
	}
	reloadNginxL7(client)

	log.DebugLog(log.DebugLevelMexos, "deleted nginx", "name", name)
	return nil
}

type Options struct {
	DockerPublishPorts bool
	DockerNetwork      string
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

func (o *Options) Apply(ops []Op) {
	for _, op := range ops {
		op(o)
	}
}
