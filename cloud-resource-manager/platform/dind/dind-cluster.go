package dind

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type DindCluster struct {
	ClusterName string
	ClusterID   int
	MasterAddr  string
	KContext    string
}

func (s *Platform) CreateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, timeout time.Duration) error {
	var err error

	switch clusterInst.Deployment {
	case cloudcommon.AppDeploymentTypeDocker:
		updateCallback(edgeproto.UpdateTask, "Create done for Docker Cluster on DIND")
		return nil
	case cloudcommon.AppDeploymentTypeKubernetes:
		updateCallback(edgeproto.UpdateTask, "Create DIND Cluster")
	default:
		return fmt.Errorf("Only K8s and Docker clusters are supported on DIND")
	}
	// Create K8s cluster
	clusterName := k8smgmt.NormalizeName(clusterInst.Key.ClusterKey.Name + clusterInst.Key.Developer)
	log.SpanLog(ctx, log.DebugLevelMexos, "creating local dind cluster", "clusterName", clusterName)

	kconfName := k8smgmt.GetKconfName(clusterInst)
	if err = s.CreateDINDCluster(ctx, clusterName, kconfName); err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "created dind", "name", clusterName)
	return nil
}

func (s *Platform) UpdateClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("update cluster not supported for DIND")
}

func (s *Platform) DeleteClusterInst(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {
	return s.DeleteDINDCluster(ctx, clusterInst)
}

//CreateDINDCluster creates kubernetes cluster on local mac
func (s *Platform) CreateDINDCluster(ctx context.Context, clusterName, kconfName string) error {
	clusters, err := GetClusters()
	if err != nil {
		return err
	}
	ids := make(map[int]struct{})
	for _, clust := range clusters {
		if clust.ClusterName == clusterName {
			return fmt.Errorf("ERROR - Cluster %s already exists (%v)", clusterName, clust)
		}
		ids[clust.ClusterID] = struct{}{}
	}
	clusterID := 1
	for {
		if _, found := ids[clusterID]; !found {
			break
		}
		clusterID++
	}
	// if KUBECONFIG is set, then the dind-cluster script will write config
	// to that file instead of ~/.kube/config, which is super confusing.
	// For consistency, make sure KUBECONFIG is not set (it may be pointing
	// to the wrong place).
	os.Unsetenv("KUBECONFIG")
	os.Setenv("DIND_LABEL", clusterName)
	os.Setenv("CLUSTER_ID", GetClusterID(clusterID))
	cluster := NewClusterFor(clusterName, clusterID)
	log.SpanLog(ctx, log.DebugLevelMexos, "CreateDINDCluster", "scriptName", cloudcommon.DindScriptName, "name", clusterName, "clusterid", clusterID)

	out, err := sh.Command(cloudcommon.DindScriptName, "up").Command("tee", "/tmp/dind.log").CombinedOutput()
	if err != nil {
		return fmt.Errorf("ERROR creating Dind Cluster: [%s] %v", out, err)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "Finished CreateDINDCluster", "name", clusterName)
	//race condition exists where the config file is not ready until just after the cluster create is done
	time.Sleep(3 * time.Second)

	//now set the k8s config
	log.SpanLog(ctx, log.DebugLevelMexos, "set config context", "kcontext", cluster.KContext)
	out, err = sh.Command("kubectl", "config", "use-context", cluster.KContext).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ERROR setting kube config context: [%s] %v", string(out), err)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "set config context output", "out", string(out), "err", err)

	//copy kubeconfig locally
	log.SpanLog(ctx, log.DebugLevelMexos, "locally copying kubeconfig", "kconfName", kconfName)
	home := os.Getenv("HOME")
	out, err = sh.Command("cp", home+"/.kube/config", kconfName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	out, err = sh.Command("cat", home+"/.kube/config").CombinedOutput()
	log.SpanLog(ctx, log.DebugLevelMexos, "config file", "home", home, "out", string(out), "err", err)

	// bridge nginxL7 network to this cluster's network
	out, err = sh.Command("docker", "network", "connect",
		GetDockerNetworkName(&cluster), proxy.NginxL7Name).CombinedOutput()
	if err != nil && strings.Contains(string(out), "already exists") {
		err = nil
	}
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "cannot connect nginx network",
			"cluster", cluster, "out", out, "err", err)
		return fmt.Errorf("failed to connect nginxL7 network, %s, %v", out, err)
	}
	return nil
}

//DeleteDINDCluster creates kubernetes cluster on local mac
func (s *Platform) DeleteDINDCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) error {

	clusterName := k8smgmt.NormalizeName(clusterInst.Key.ClusterKey.Name + clusterInst.Key.Developer)
	log.SpanLog(ctx, log.DebugLevelMexos, "DeleteDINDCluster", "clusterName", clusterName)

	if clusterInst.Deployment == cloudcommon.AppDeploymentTypeDocker {
		log.SpanLog(ctx, log.DebugLevelMexos, "No delete required for DIND docker cluster", "clusterName", clusterName)
		return nil
	}
	cluster, err := FindCluster(clusterName)
	if err != nil {
		return fmt.Errorf("ERROR - Cluster %s not found, %v", clusterName, err)
	}

	// disconnect nginxL7 network
	out, err := sh.Command("docker", "network", "disconnect",
		GetDockerNetworkName(cluster), proxy.NginxL7Name).CombinedOutput()
	if err != nil && strings.Contains(string(out), "is not connected") {
		err = nil
	}
	if err != nil {
		return fmt.Errorf("docker network disconnect failed, %s, %v",
			out, err)
	}

	os.Setenv("DIND_LABEL", cluster.ClusterName)
	os.Setenv("CLUSTER_ID", GetClusterID(cluster.ClusterID))
	out, err = sh.Command(cloudcommon.DindScriptName, "clean").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "Finished dind clean", "scriptName", cloudcommon.DindScriptName, "clusterName", clusterName, "out", out)
	return nil
}

func GetClusterID(id int) string {
	return strconv.Itoa(id)
}

func FindCluster(clusterName string) (*DindCluster, error) {
	clusters, err := GetClusters()
	if err != nil {
		return nil, err
	}
	for ii, _ := range clusters {
		if clusters[ii].ClusterName == clusterName {
			return &clusters[ii], nil
		}
	}
	return nil, fmt.Errorf("dind cluster %s not found", clusterName)
}

func GetClusters() ([]DindCluster, error) {
	out, err := sh.Command("docker", "ps", "--format", "{{.Names}}").CombinedOutput()
	if err != nil {
		return nil, err
	}
	clusters := []DindCluster{}
	r, _ := regexp.Compile("kube-master-(\\S+)-(\\d+)")
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if r.MatchString(line) {
			matches := r.FindStringSubmatch(line)
			cname := matches[1]
			cid, err := strconv.Atoi(matches[2])
			if err != nil {
				return nil, fmt.Errorf("Could not parse kube-master id: %s", line)
			}
			clusters = append(clusters, NewClusterFor(cname, cid))
		}
	}
	return clusters, nil
}

func NewClusterFor(clusterName string, id int) DindCluster {
	return DindCluster{
		ClusterName: clusterName,
		ClusterID:   id,
		KContext:    "dind-" + clusterName + "-" + GetClusterID(id),
		MasterAddr:  "10.192." + GetClusterID(id) + ".2",
	}
}

func GetDockerNetworkName(cluster *DindCluster) string {
	return "kubeadm-dind-net-" + cluster.ClusterName + "-" + GetClusterID(cluster.ClusterID)
}
