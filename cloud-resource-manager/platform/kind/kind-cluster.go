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
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// See https://hub.docker.com/r/kindest/node/tags for all available versions
// Use env var KIND_IMAGE to override default below.
var DefaultNodeImage = "kindest/node:v1.16.15"

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
	NumMasters []struct{}
	NumNodes   []struct{}
}

var ConfigTemplate = `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
{{- range .NumMasters}}
- role: control-plane
{{- end}}
{{- range .NumNodes}}
- role: worker
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
	tmpl, err := template.New("config").Parse(ConfigTemplate)
	if err != nil {
		return err
	}
	params := &ConfigParams{
		NumMasters: make([]struct{}, clusterInst.NumMasters),
		NumNodes:   make([]struct{}, clusterInst.NumNodes),
	}
	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, params)
	if err != nil {
		return err
	}
	// copy over config
	configFile := name + "-config.yaml"
	err = pc.WriteFile(client, configFile, buf.String(), "KIND config", pc.NoSudo)
	if err != nil {
		return err
	}
	nodeImage := os.Getenv("KIND_IMAGE")
	if nodeImage == "" {
		nodeImage = DefaultNodeImage
	}
	cmd := fmt.Sprintf("kind create cluster --config=%s --kubeconfig=%s --image=%s --name=%s --wait 300s", configFile, kconf, nodeImage, name)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("failed to run cmd %s, %s, %s", cmd, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "successfully created KIND cluster", "name", name)
	return nil
}

func (s *Platform) DeleteKINDCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	name := k8smgmt.GetK8sNodeNameSuffix(&clusterInst.Key)
	log.SpanLog(ctx, log.DebugLevelInfra, "delete KIND cluster", "name", name)
	client, err := s.Xind.GetClient(ctx)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("kind delete cluster --name=%s", name)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("failed to run cmd %s, %s, %s", cmd, out, err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "successfully delete KIND cluster", "name", name)
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
