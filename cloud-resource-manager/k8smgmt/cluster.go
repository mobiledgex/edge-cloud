package k8smgmt

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

func DeleteNodes(ctx context.Context, client ssh.Client, kconfName string, nodes []string) error {
	for _, node := range nodes {
		cmd := fmt.Sprintf("KUBECONFIG=%s kubectl delete node %s", kconfName, node)
		log.SpanLog(ctx, log.DebugLevelInfra, "k8smgmt delete node", "node", node, "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("failed to delete k8s node, %s, %s, %v", cmd, out, err)
		}
	}
	return nil
}

func CleanupClusterConfig(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst) error {
	names, err := GetKubeNames(clusterInst, &edgeproto.App{}, &edgeproto.AppInst{})
	if err != nil {
		return err
	}
	configDir, _ := getConfigDirName(names)
	log.SpanLog(ctx, log.DebugLevelInfra, "CleanupClusterConfig remove dir", "configDir", configDir)
	err = pc.DeleteDir(ctx, client, configDir, pc.NoSudo)
	if err != nil {
		return fmt.Errorf("failed to delete cluster config dir %s: %v", configDir, err)
	}
	return nil
}
