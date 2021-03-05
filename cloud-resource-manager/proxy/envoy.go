package proxy

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/dockermgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

var envoyYamlT *template.Template
var sdsYamlT *template.Template

// this is the default value in envoy, for DOS protection
const defaultConcurrentConns uint64 = 1024

func init() {
	envoyYamlT = template.Must(template.New("yaml").Parse(envoyYaml))
	sdsYamlT = template.Must(template.New("yaml").Parse(sdsYaml))
}

func CreateEnvoyProxy(ctx context.Context, client ssh.Client, name, listenIP, backendIP string, ports []dme.AppPort, skipHcPorts string, ops ...Op) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "create envoy", "name", name, "listenIP", listenIP, "backendIP", backendIP, "ports", ports)
	opts := Options{}
	opts.Apply(ops)

	out, err := client.Output("pwd")
	if err != nil {
		return err
	}
	pwd := strings.TrimSpace(string(out))

	dir := pwd + "/envoy/" + name
	log.SpanLog(ctx, log.DebugLevelInfra, "envoy remote dir", "name", name, "dir", dir)
	err = pc.Run(client, "mkdir -p "+dir)
	if err != nil {
		return err
	}
	accesslogFile := dir + "/access.log"
	err = pc.Run(client, "touch "+accesslogFile)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra,
			"envoy %s can't create file %s", name, accesslogFile)
		return err
	}
	eyamlName := dir + "/envoy.yaml"
	syamlName := dir + "/sds.yaml"
	isTLS, err := createEnvoyYaml(ctx, client, dir, name, listenIP, backendIP, ports, skipHcPorts)
	if err != nil {
		return fmt.Errorf("create envoy.yaml failed, %v", err)
	}

	// container name is envoy+name for now to avoid conflicts with the nginx containers
	cmdArgs := []string{"run", "-d", "-l edge-cloud", "--restart=unless-stopped", "--name", "envoy" + name}
	if opts.DockerPublishPorts {
		cmdArgs = append(cmdArgs, dockermgmt.GetDockerPortString(ports, dockermgmt.UsePublicPortInContainer, dockermgmt.EnvoyProxy, listenIP)...)
	}
	if opts.DockerNetwork != "" {
		// For dind, we use the network which the dind cluster is on.
		cmdArgs = append(cmdArgs, "--network", opts.DockerNetwork)
	}
	out, err = client.Output("pwd")
	if err != nil {
		return fmt.Errorf("Unable to get pwd: %v", err)
	}
	certsDir, _, _ := cloudcommon.GetCertsDirAndFiles(string(out))
	if isTLS {
		// use envoy SDS (secret discovery service) to refresh certs
		cmdArgs = append(cmdArgs, "-v", syamlName+":/etc/envoy/sds.yaml")
	}
	cmdArgs = append(cmdArgs, []string{
		"-v", certsDir + ":/etc/envoy/certs",
		"-v", accesslogFile + ":/tmp/access.log",
		"-v", eyamlName + ":/etc/envoy/envoy.yaml"}...)
	if opts.DockerUser != "" {
		cmdArgs = append(cmdArgs, []string{"-u", fmt.Sprintf("%s:%s", opts.DockerUser, opts.DockerUser)}...)
	}
	cmdArgs = append(cmdArgs, "docker.mobiledgex.net/mobiledgex/mobiledgex_public/envoy-with-curl@"+cloudcommon.EnvoyImageDigest)
	cmd := "docker " + strings.Join(cmdArgs, " ")
	log.SpanLog(ctx, log.DebugLevelInfra, "envoy docker command", "name", "envoy"+name,
		"cmd", cmd)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't create envoy container %s, %s, %v", "envoy"+name, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "created envoy container", "name", name)
	return nil
}

// Build a map of individual ports with key struct:
// <proto>:<portnum>
func buildPortsMapFromString(portsString string) (map[string]struct{}, error) {
	portMap := make(map[string]struct{})
	if portsString == "all" {
		return portMap, nil
	}
	ports, err := edgeproto.ParseAppPorts(portsString)
	if err != nil {
		return nil, err
	}
	for _, port := range ports {
		if port.EndPort == 0 {
			port.EndPort = port.InternalPort
		}
		for p := port.InternalPort; p <= port.EndPort; p++ {
			proto, err := edgeproto.LProtoStr(port.Proto)
			if err != nil {
				return nil, err
			}
			key := fmt.Sprintf("%s:%d", proto, p)
			portMap[key] = struct{}{}
		}
	}
	return portMap, nil
}

func createEnvoyYaml(ctx context.Context, client ssh.Client, yamldir, name, listenIP, backendIP string, ports []dme.AppPort, skipHcPorts string) (bool, error) {
	var skipHcAll = false
	var skipHcPortsMap map[string]struct{}
	var err error

	spec := ProxySpec{
		Name:       name,
		MetricPort: cloudcommon.ProxyMetricsPort,
		CertName:   cloudcommon.CertName,
	}
	// check skip health check ports
	if skipHcPorts == "all" {
		skipHcAll = true
	} else {
		skipHcPortsMap, err = buildPortsMapFromString(skipHcPorts)
		if err != nil {
			return false, err
		}
	}

	isTLS := false
	for _, p := range ports {
		endPort := p.EndPort
		if endPort == 0 {
			endPort = p.PublicPort
		} else {
			// if we have a port range, the internal ports and external ports must match
			if p.InternalPort != p.PublicPort {
				return false, fmt.Errorf("public and internal ports must match when port range in use")
			}
		}
		// Currently there is no (known) way to put a port range within Envoy.
		// So we create one spec per port when there is a port range in use
		internalPort := p.InternalPort
		for pubPort := p.PublicPort; pubPort <= endPort; pubPort++ {
			switch p.Proto {
			// only support tcp for now
			case dme.LProto_L_PROTO_TCP:
				key := fmt.Sprintf("%s:%d", "tcp", internalPort)
				_, skipHealthCheck := skipHcPortsMap[key]
				tcpPort := TCPSpecDetail{
					ListenPort:  pubPort,
					ListenIP:    listenIP,
					BackendIP:   backendIP,
					BackendPort: internalPort,
					UseTLS:      p.Tls,
					HealthCheck: !skipHcAll && !skipHealthCheck,
				}
				if p.Tls {
					isTLS = true
				}
				tcpconns, err := getTCPConcurrentConnections()
				if err != nil {
					return false, err
				}
				tcpPort.ConcurrentConns = tcpconns
				spec.TCPSpec = append(spec.TCPSpec, &tcpPort)
			case dme.LProto_L_PROTO_UDP:
				if p.Nginx { // defv specified nginx for this port (range)
					continue
				}
				udpPort := UDPSpecDetail{
					ListenPort:  pubPort,
					ListenIP:    listenIP,
					BackendIP:   backendIP,
					BackendPort: internalPort,
				}
				udpconns, err := getUDPConcurrentConnections()
				if err != nil {
					return false, err
				}
				udpPort.ConcurrentConns = udpconns
				spec.UDPSpec = append(spec.UDPSpec, &udpPort)
			}
			internalPort++
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "create envoy yaml", "name", name)
	buf := bytes.Buffer{}
	err = envoyYamlT.Execute(&buf, &spec)
	if err != nil {
		return isTLS, err
	}
	err = pc.WriteFile(client, yamldir+"/envoy.yaml", buf.String(), "envoy.yaml", pc.NoSudo)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "write envoy.yaml failed",
			"name", name, "err", err)
		return isTLS, err
	}
	if isTLS {
		log.SpanLog(ctx, log.DebugLevelInfra, "create sds yaml", "name", name)
		buf := bytes.Buffer{}
		err = sdsYamlT.Execute(&buf, &spec)
		if err != nil {
			return isTLS, err
		}
		err = pc.WriteFile(client, yamldir+"/sds.yaml", buf.String(), "sds.yaml", pc.NoSudo)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "write sds.yaml failed",
				"name", name, "err", err)
			return isTLS, err
		}
	}
	return isTLS, nil
}

// TODO: Probably should eventually find a better way to uniquely name clusters other than just by the port theyre getting proxied from
var envoyYaml = `
node:
  id: {{.Name}}
  cluster: {{.Name}}
static_resources:
  listeners:
  {{- range .TCPSpec}}
  - address:
      socket_address:
        address: 0.0.0.0
        port_value: {{.ListenPort}}
    filter_chains:
    - filters:
      - name: envoy.tcp_proxy
        config:
          stat_prefix: ingress_tcp
          cluster: backend{{.BackendPort}}
          access_log:
            - name: envoy.file_access_log
              config:
                path: /tmp/access.log
                json_format: {
                  "start_time": "%START_TIME%",
                  "duration": "%DURATION%",
                  "bytes_sent": "%BYTES_SENT%",
                  "bytes_received": "%BYTES_RECEIVED%",
                  "client_address": "%DOWNSTREAM_REMOTE_ADDRESS%",
                  "upstream_cluster": "%UPSTREAM_CLUSTER%"
                }
      {{if .UseTLS -}}
      transport_socket:
        name: "envoy.transport_sockets.tls"
        typed_config:
          "@type": "type.googleapis.com/envoy.api.v2.auth.DownstreamTlsContext"
          common_tls_context:
            tls_certificate_sds_secret_configs:
                sds_config:
                    path: /etc/envoy/sds.yaml
      {{- end}}
  {{- end}}
  {{- range .UDPSpec}}
  - name: udp_listener_{{.ListenPort}}
    address:
      socket_address:
        protocol: UDP
        address: 0.0.0.0
        port_value: {{.ListenPort}}
    listener_filters:
      name: envoy.filters.udp_listener.udp_proxy
      typed_config:
        '@type': type.googleapis.com/envoy.extensions.filters.udp.udp_proxy.v3.UdpProxyConfig
        stat_prefix: downstream{{.BackendPort}}
        cluster: udp_backend{{.BackendPort}}
    reuse_port: true
  {{- end}}
  clusters:
  {{- range .TCPSpec}}
  - name: backend{{.BackendPort}}
    connect_timeout: 0.25s
    type: strict_dns
    circuit_breakers:
        thresholds:
            max_connections: {{.ConcurrentConns}}
    lb_policy: round_robin
    hosts:
    - socket_address:
        address: {{.BackendIP}}
        port_value: {{.BackendPort}}
    {{if .HealthCheck -}}
    health_checks:
      - timeout: 1s
        interval: 5s
        interval_jitter: 1s
        unhealthy_threshold: 3
        healthy_threshold: 3
        tcp_health_check: {}
        no_traffic_interval: 5s
    {{- end}}
{{- end}}
{{- range .UDPSpec}}
  - name: udp_backend{{.BackendPort}}
    connect_timeout: 0.25s
    type: STRICT_DNS
    circuit_breakers:
      thresholds:
        max_connections: {{.ConcurrentConns}}
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: udp_backend{{.BackendPort}}
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: {{.BackendIP}}
                port_value: {{.BackendPort}}
{{- end}}
admin:
  access_log_path: "/tmp/admin.log"
  address:
    socket_address:
      address: 0.0.0.0
      port_value: {{.MetricPort}}
`

var sdsYaml = `
resources:
- "@type": "type.googleapis.com/envoy.api.v2.auth.Secret"
  tls_certificate:
    certificate_chain:
      filename: "/etc/envoy/certs/{{$.CertName}}.crt"
    private_key:
      filename: "/etc/envoy/certs/{{$.CertName}}.key"
`

func DeleteEnvoyProxy(ctx context.Context, client ssh.Client, name string) error {
	containerName := "envoy" + name

	log.SpanLog(ctx, log.DebugLevelInfra, "delete envoy", "name", containerName)
	out, err := client.Output("docker kill " + containerName)
	log.SpanLog(ctx, log.DebugLevelInfra, "kill envoy result", "out", out, "err", err)

	envoyDir := "envoy/" + name
	out, err = client.Output("rm -rf " + envoyDir)
	log.SpanLog(ctx, log.DebugLevelInfra, "delete envoy dir", "name", name, "dir", envoyDir, "out", out, "err", err)

	out, err = client.Output("docker rm -f " + "envoy" + name)
	log.SpanLog(ctx, log.DebugLevelInfra, "rm envoy result", "out", out, "err", err)
	if err != nil && !strings.Contains(string(out), "No such container") {
		// delete the envoy proxy anyway
		return fmt.Errorf("can't remove envoy container %s, %s, %v", name, out, err)
	}
	return nil
}

func GetEnvoyContainerName(name string) string {
	return "envoy" + name
}
