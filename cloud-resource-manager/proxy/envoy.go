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
	"github.com/mobiledgex/edge-cloud/log"
)

// Envoy is now handling all of our L4 TCP loadbalancing
// Eventually L7 and UDP Proxying support will be added so
// all of our loadbalancing is handled by envoy.
// UDP proxying is currently blocked by: https://github.com/envoyproxy/envoy/issues/492

var envoyYamlT *template.Template

func init() {
	envoyYamlT = template.Must(template.New("yaml").Parse(envoyYaml))
}

func CreateEnvoyProxy(ctx context.Context, client pc.PlatformClient, name, originIP string, ports []dme.AppPort, ops ...Op) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "create envoy", "name", name, "originip", originIP, "ports", ports)
	opts := Options{}
	opts.Apply(ops)

	out, err := client.Output("pwd")
	if err != nil {
		return err
	}
	pwd := strings.TrimSpace(string(out))

	dir := pwd + "/envoy/" + name
	log.SpanLog(ctx, log.DebugLevelMexos, "envoy remote dir", "name", name, "dir", dir)

	err = pc.Run(client, "mkdir -p "+dir)
	if err != nil {
		return err
	}
	accesslogFile := dir + "/access.log"
	err = pc.Run(client, "touch "+accesslogFile)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos,
			"envoy %s can't create file %s", name, accesslogFile)
		return err
	}
	eyamlName := dir + "/envoy.yaml"
	err = createEnvoyYaml(ctx, client, eyamlName, name, originIP, ports)
	if err != nil {
		return fmt.Errorf("create envoy.yaml failed, %v", err)
	}

	// container name is envoy+name for now to avoid conflicts with the nginx containers
	cmdArgs := []string{"run", "-d", "-l edge-cloud", "--restart=unless-stopped", "--name", "envoy" + name}
	if opts.DockerPublishPorts {
		cmdArgs = append(cmdArgs, dockermgmt.GetDockerPortString(ports, dockermgmt.UsePublicPortInContainer)...)
	}
	if opts.DockerNetwork != "" {
		// For dind, we use the network which the dind cluster is on.
		cmdArgs = append(cmdArgs, "--network", opts.DockerNetwork)
	}
	cmdArgs = append(cmdArgs, []string{
		"-v", accesslogFile + ":/var/log/access.log",
		"-v", eyamlName + ":/etc/envoy/envoy.yaml",
		"docker.mobiledgex.net/mobiledgex/mobiledgex_public/envoy-with-curl"}...)
	cmd := "docker " + strings.Join(cmdArgs, " ")
	log.SpanLog(ctx, log.DebugLevelMexos, "envoy docker command", "name", "envoy"+name,
		"cmd", cmd)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't create envoy container %s, %s, %v", "envoy"+name, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "created envoy container", "name", name)
	return nil
}

func createEnvoyYaml(ctx context.Context, client pc.PlatformClient, yamlname, name, originIP string, ports []dme.AppPort) error {
	spec := ProxySpec{
		Name:       name,
		MetricPort: cloudcommon.LBMetricsPort,
	}
	for _, p := range ports {
		switch p.Proto {
		// only support tcp for now
		case dme.LProto_L_PROTO_TCP:
			tcpPort := TCPSpecDetail{
				Port:       p.PublicPort,
				Origin:     originIP,
				OriginPort: p.InternalPort,
			}
			spec.TCPSpec = append(spec.TCPSpec, &tcpPort)
			spec.L4 = true
		}
	}

	log.SpanLog(ctx, log.DebugLevelMexos, "create envoy yaml", "name", name)
	buf := bytes.Buffer{}
	err := envoyYamlT.Execute(&buf, &spec)
	if err != nil {
		return err
	}
	err = pc.WriteFile(client, yamlname, buf.String(), "envoy.yaml", pc.NoSudo)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "write envoy.yaml failed",
			"name", name, "err", err)
		return err
	}
	return nil
}

// TODO: Probably should eventually find a better way to uniquely name clusters other than just by the port theyre getting proxied from
var envoyYaml = `
{{if .L4 -}}
static_resources:
  listeners:
  {{- range .TCPSpec}}
  - address:
      socket_address:
        address: 0.0.0.0
        port_value: {{.Port}}
    filter_chains:
    - filters:
      - name: envoy.tcp_proxy
        config:
          stat_prefix: ingress_tcp
          cluster: backend{{.Port}}
          access_log:
            - name: envoy.file_access_log
              config:
                path: /var/log/access.log
                json_format: {
                  "start_time": "%START_TIME%",
                  "duration": "%DURATION%",
                  "bytes_sent": "%BYTES_SENT%",
                  "bytes_received": "%BYTES_RECEIVED%",
                  "client_address": "%DOWNSTREAM_REMOTE_ADDRESS%",
                  "upstream_cluster": "%UPSTREAM_CLUSTER%"
				}
  {{- end}}
  clusters:
  {{- range .TCPSpec}}
  - name: backend{{.Port}}
    connect_timeout: 0.25s
    type: strict_dns
    lb_policy: round_robin
    hosts:
    - socket_address:
        address: {{.Origin}}
        port_value: {{.OriginPort}}
{{- end}}
admin:
  access_log_path: "/var/log/admin.log"
  address:
    socket_address:
      address: 0.0.0.0
      port_value: {{.MetricPort}}
{{- end}}
`

func DeleteEnvoyProxy(ctx context.Context, client pc.PlatformClient, name string) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "delete envoy", "name", "envoy"+name)
	out, err := client.Output("docker kill " + "envoy" + name)
	deleteContainer := false
	if err == nil {
		deleteContainer = true
	} else {
		if strings.Contains(string(out), "No such container") {
			log.SpanLog(ctx, log.DebugLevelMexos,
				"envoy LB container already gone", "name", "envoy"+name)
		} else {
			return fmt.Errorf("can't delete envoy container %s, %s, %v", name, out, err)
		}
	}
	envoyDir := "envoy/" + name
	out, err = client.Output("rm -rf " + envoyDir)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "delete envoy dir", "name", name, "dir", envoyDir, "out", out, "err", err)
	}
	if deleteContainer {
		out, err = client.Output("docker rm " + "envoy" + name)
		if err != nil && !strings.Contains(string(out), "No such container") {
			return fmt.Errorf("can't remove envoy container %s, %s, %v", "envoy"+name, out, err)
		}
	}

	log.SpanLog(ctx, log.DebugLevelMexos, "deleted envoy", "name", name)
	return nil
}

func GetEnvoyContainerName(name string) string {
	return "envoy" + name
}
