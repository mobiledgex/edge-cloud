package k8smgmt

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

func DeleteNodes(ctx context.Context, client ssh.Client, kconfName string, nodes []string) error {
	for _, node := range nodes {
		cmd := fmt.Sprintf("KUBECONFIG=%s kubectl delete node %s", kconfName, node)
		log.SpanLog(ctx, log.DebugLevelMexos, "k8smgmt delete node", "node", node, "cmd", cmd)
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("failed to delete k8s node, %s, %s, %v", cmd, out, err)
		}
	}
	return nil
}
