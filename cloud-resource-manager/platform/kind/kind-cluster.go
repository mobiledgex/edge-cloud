package kind

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/common/xind"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

// See https://hub.docker.com/r/kindest/node/tags for all available versions
// Use env var KIND_IMAGE to override default below.
var DefaultNodeImage = "kindest/node:v1.17.17"

func (s *Platform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	var err error

	switch clusterInst.Deployment {
	case cloudcommon.DeploymentTypeDocker:
		updateCallback(edgeproto.UpdateTask, "Create done for Docker Cluster on KIND")
		return nil
	case cloudcommon.DeploymentTypeKubernetes:
		updateCallback(edgeproto.UpdateTask, "Create KIND Cluster")
	default:
		return fmt.Errorf("Only K8s and Docker clusters are supported on KIND")
	}
	// Create K8s cluster
	if err = s.CreateKINDCluster(ctx, clusterInst); err != nil {
		return err
	}
	return nil
}

func (s *Platform) UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("update cluster not supported for KIND")
}

func (s *Platform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return s.DeleteKINDCluster(ctx, clusterInst)
}

type ConfigParams struct {
	Image      string
	NumMasters []struct{}
	NumNodes   []struct{}
}

var ConfigTemplate = `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: true
nodes:
{{- range .NumMasters}}
- role: control-plane
  image: {{$.Image}}
{{- end}}
{{- range .NumNodes}}
- role: worker
  image: {{$.Image}}
{{- end}}
`

func (s *Platform) CreateKINDCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	name := k8smgmt.GetK8sNodeNameSuffix(&clusterInst.Key)
	kconf := k8smgmt.GetKconfName(clusterInst)
	log.SpanLog(ctx, log.DebugLevelInfra, "create KIND cluster", "name", name, "kconf", kconf)
	client, err := s.Xind.GetClient(ctx)
	if err != nil {
		return err
	}
	nodeImage := os.Getenv("KIND_IMAGE")
	if nodeImage == "" {
		nodeImage = DefaultNodeImage
	}
	tmpl, err := template.New("config").Parse(ConfigTemplate)
	if err != nil {
		return err
	}
	params := &ConfigParams{
		Image:      nodeImage,
		NumMasters: make([]struct{}, clusterInst.NumMasters),
		NumNodes:   make([]struct{}, clusterInst.NumNodes),
	}
	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, params)
	if err != nil {
		return err
	}
	configFile := name + "-config.yaml"

	// see if cluster already exists with the correct config
	exists, err := ClusterExists(ctx, client, name)
	if err != nil {
		return err
	}
	if exists {
		log.SpanLog(ctx, log.DebugLevelInfra, "cluster exists, checking config", "name", name, "config", configFile)
		cmd := fmt.Sprintf("cat %s", configFile)
		out, err := client.Output(cmd)
		if err == nil && strings.TrimSpace(out) == strings.TrimSpace(buf.String()) {
			log.SpanLog(ctx, log.DebugLevelInfra, "cluster exists and config matches, reusing")
			// unpause nodes if they were paused
			nodes, err := GetClusterContainerNames(ctx, client, name)
			if err != nil {
				return err
			}
			err = xind.UnpauseContainers(ctx, client, nodes)
			if err != nil {
				return err
			}
			// clear out any leftover AppInsts
			err = k8smgmt.ClearCluster(ctx, client, clusterInst)
			if err != nil {
				return err
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "reusing existing KIND cluster", "name", name)
			return nil
		}
		// delete cluster
		log.SpanLog(ctx, log.DebugLevelInfra, "missing or mismatched config, removing existing and then recreating", "name", name)
		cmd = fmt.Sprintf("kind delete cluster --name=%s", name)
		out, err = client.Output(cmd)
		if err != nil {
			return cmdFailed(cmd, out, err)
		}
	}

	// write config
	err = pc.WriteFile(client, configFile, buf.String(), "KIND config", pc.NoSudo)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("kind create cluster --config=%s --kubeconfig=%s --name=%s", configFile, kconf, name)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("failed to run cmd %s, %s, %s", cmd, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "successfully created KIND cluster", "name", name)
	kconfEnv := "KUBECONFIG=" + kconf
	cmd = fmt.Sprintf(`%s kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$(%s kubectl version | base64 | tr -d '\n')"`, kconfEnv, kconfEnv)
	// XXX: in case we decide to use cilium, or in case we need to test
	// against cilium, this is how to install it:
	// cmd = fmt.Sprintf(`%s kubectl create -f https://raw.githubusercontent.com/cilium/cilium/v1.9/install/kubernetes/quick-install.yaml`, kconfEnv)
	log.SpanLog(ctx, log.DebugLevelInfra, "installing weave", "cmd", cmd)
	out, err = client.Output(cmd)
	log.SpanLog(ctx, log.DebugLevelInfra, "weave install result", "out", out, "err", err)
	if err != nil {
		return fmt.Errorf("failed to install weave %s: %s, %s", cmd, out, err)
	}
	err = xind.WaitClusterReady(ctx, client, clusterInst, 300*time.Second)
	if err != nil {
		return err
	}
	return nil
}

func (s *Platform) DeleteKINDCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	name := k8smgmt.GetK8sNodeNameSuffix(&clusterInst.Key)
	log.SpanLog(ctx, log.DebugLevelInfra, "delete KIND cluster", "name", name)
	client, err := s.Xind.GetClient(ctx)
	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "pausing cluster instead of deleting", "name", name)
	// clear out any AppInsts
	err = k8smgmt.ClearCluster(ctx, client, clusterInst)
	if err != nil {
		return err
	}
	// pause nodes
	nodes, err := GetClusterContainerNames(ctx, client, name)
	if err != nil {
		return err
	}
	err = xind.PauseContainers(ctx, client, nodes)
	log.SpanLog(ctx, log.DebugLevelInfra, "successfully paused KIND cluster", "name", name)
	return nil
}

func (s *Platform) GetMasterIp(ctx context.Context, names *k8smgmt.KubeNames) (string, error) {
	masterContainer := names.K8sNodeNameSuffix + "-control-plane"
	client, err := s.Xind.GetClient(ctx)
	if err != nil {
		return "", err
	}
	cmd := fmt.Sprintf("docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' %s", masterContainer)
	out, err := client.Output(cmd)
	if err != nil {
		return "", err
	}
	lines := strings.Split(out, "\n")
	return strings.TrimSpace(lines[0]), nil
}

func (s *Platform) GetDockerNetworkName(ctx context.Context, names *k8smgmt.KubeNames) (string, error) {
	masterContainer := names.K8sNodeNameSuffix + "-control-plane"
	client, err := s.Xind.GetClient(ctx)
	if err != nil {
		return "", err
	}
	cmd := fmt.Sprintf("docker inspect -f '{{.HostConfig.NetworkMode}}' %s", masterContainer)
	out, err := client.Output(cmd)
	if err != nil {
		return "", err
	}
	lines := strings.Split(out, "\n")
	return strings.TrimSpace(lines[0]), nil
}

func cmdFailed(cmd string, out string, err error) error {
	return fmt.Errorf("command failed, %s: %s, %s", cmd, out, err)
}

func GetClusters(ctx context.Context, client ssh.Client) ([]string, error) {
	cmd := "kind get clusters"
	out, err := client.Output(cmd)
	if err != nil {
		return nil, cmdFailed(cmd, out, err)
	}
	if strings.Contains(out, "No kind clusters found") {
		return []string{}, nil
	}
	clusters := []string{}
	for _, name := range strings.Split(out, "\n") {
		name = strings.TrimSpace(name)
		if name != "" {
			clusters = append(clusters, name)
		}
	}
	return clusters, nil
}

func ClusterExists(ctx context.Context, client ssh.Client, name string) (bool, error) {
	cmd := "kind get clusters"
	out, err := client.Output(cmd)
	if err != nil {
		return false, cmdFailed(cmd, out, err)
	}
	for _, n := range strings.Split(out, "\n") {
		if name == n {
			return true, nil
		}
	}
	return false, nil
}

func GetClusterContainerNames(ctx context.Context, client ssh.Client, clusterName string) ([]string, error) {
	cmd := fmt.Sprintf("kind get nodes --name %s", clusterName)
	out, err := client.Output(cmd)
	if err != nil {
		return nil, cmdFailed(cmd, out, err)
	}
	nodes := []string{}
	for _, name := range strings.Split(out, "\n") {
		name = strings.TrimSpace(name)
		if name != "" {
			nodes = append(nodes, name)
		}
	}
	return nodes, nil
}

func (s *Platform) ActiveChanged(ctx context.Context, platformActive bool) {
}
