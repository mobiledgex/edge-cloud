package dind

import (
	"fmt"
	"os"
	"strconv"
	"time"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8s"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type DindCluster struct {
	ClusterName string
	ClusterID   int
	MasterAddr  string
	KContext    string
}

func (s *Platform) CreateCluster(clusterInst *edgeproto.ClusterInst) error {
	var err error

	clusterName := clusterInst.Key.ClusterKey.Name
	log.DebugLog(log.DebugLevelMexos, "creating local dind cluster", "clusterName", clusterName)

	kconfName := k8s.GetKconfName(clusterInst)
	if err = s.CreateDINDCluster(clusterName, kconfName); err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "created dind", "name", clusterName)
	return nil
}

func (s *Platform) DeleteCluster(clusterInst *edgeproto.ClusterInst) error {
	return s.DeleteDINDCluster(clusterInst.Key.ClusterKey.Name)
}

//CreateDINDCluster creates kubernetes cluster on local mac
func (s *Platform) CreateDINDCluster(clusterName, kconfName string) error {
	cluster, found := s.Clusters[clusterName]
	if found {
		return fmt.Errorf("ERROR - Cluster %s already exists (%v)", clusterName, *cluster)
	}
	s.nextClusterID++
	clusterID := s.nextClusterID
	os.Setenv("DIND_LABEL", clusterName)
	os.Setenv("CLUSTER_ID", GetClusterID(clusterID))
	cluster = &DindCluster{
		ClusterName: clusterName,
		ClusterID:   clusterID,
		KContext:    "dind-" + clusterName + "-" + GetClusterID(clusterID),
		MasterAddr:  "10.192." + GetClusterID(clusterID) + ".2",
	}
	log.DebugLog(log.DebugLevelMexos, "CreateDINDCluster via dind-cluster-v1.13.sh", "name", clusterName, "clusterid", clusterID)

	out, err := sh.Command("dind-cluster-v1.13.sh", "up").Command("tee", "/tmp/dind.log").CombinedOutput()
	if err != nil {
		return fmt.Errorf("ERROR creating Dind Cluster: [%s] %v", out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "Finished CreateDINDCluster", "name", clusterName)
	//race condition exists where the config file is not ready until just after the cluster create is done
	time.Sleep(3 * time.Second)

	//now set the k8s config
	out, err = sh.Command("kubectl", "config", "use-context", cluster.KContext).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ERROR setting kube config context: [%s] %v", out, err)
	}
	//copy kubeconfig locally
	log.DebugLog(log.DebugLevelMexos, "locally copying kubeconfig", "kconfName", kconfName)
	home := os.Getenv("HOME")
	out, err = sh.Command("cp", home+"/.kube/config", kconfName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	// add cluster to cluster map
	s.Clusters[clusterName] = cluster
	return nil
}

//DeleteDINDCluster creates kubernetes cluster on local mac
func (s *Platform) DeleteDINDCluster(name string) error {
	cluster, found := s.Clusters[name]
	if !found {
		return fmt.Errorf("ERROR - Cluster %s doesn't exists", name)
	}
	os.Setenv("DIND_LABEL", cluster.ClusterName)
	os.Setenv("CLUSTER_ID", GetClusterID(cluster.ClusterID))
	log.DebugLog(log.DebugLevelMexos, "DeleteDINDCluster", "name", name)

	out, err := sh.Command("dind-cluster-v1.13.sh", "clean").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "Finished dind-cluster-v1.13.sh clean", "name", name, "out", out)
	// Delete the entry from the dindClusters
	delete(s.Clusters, name)

	/* network is already deleted by the clean
	netname := GetDockerNetworkName(name)
	log.DebugLog(log.DebugLevelMexos, "removing docker network", "netname", netname, "out", out)
	out, err = sh.Command("docker", "network", "rm", netname).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v", out, err)
	}
	fmt.Printf("ran command docker network rm for network: %s.  Result: %s", netname, out)
	*/
	return nil
}

func GetClusterID(id int) string {
	return strconv.Itoa(id)
}
