package xind

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

func PauseContainers(ctx context.Context, client ssh.Client, names []string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "pausing containers", "names", names)
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			continue
		}
		cmd := fmt.Sprintf("docker pause %s", name)
		out, err := client.Output(cmd)
		if err != nil && !strings.Contains(out, "is already paused") {
			return fmt.Errorf("pausing container failed, %s: %s, %s", cmd, out, err)
		}
	}
	return nil
}

func UnpauseContainers(ctx context.Context, client ssh.Client, names []string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "unpausing containers", "names", names)
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			continue
		}
		cmd := fmt.Sprintf("docker unpause %s", name)
		out, err := client.Output(cmd)
		if err != nil && !strings.Contains(out, "is not paused") {
			return fmt.Errorf("unpausing container failed, %s: %s, %s", cmd, out, err)
		}
	}
	return nil
}

var waitDelay = 5 * time.Second

func WaitClusterReady(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst, timeout time.Duration) error {
	names, err := k8smgmt.GetKubeNames(clusterInst, &edgeproto.App{}, &edgeproto.AppInst{})
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "waiting for cluster to be ready", "kconf", names.KconfName)
	startTime := time.Now()
	for {
		if time.Since(startTime) > timeout {
			break
		}
		cmd := fmt.Sprintf("%s kubectl get nodes --no-headers", names.KconfEnv)
		out, err := client.Output(cmd)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "wait cluster ready", "cmd", cmd, "out", out, "err", err)
			time.Sleep(waitDelay)
			continue
		}
		ready := 0
		total := 0
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			state := fields[1]
			if state == "Ready" {
				ready++
			}
			total++
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "wait cluster ready", "kconf", names.KconfName, "ready", ready, "total", total)
		if ready == total {
			return nil
		}
		time.Sleep(waitDelay)
	}
	return fmt.Errorf("timed out waiting for cluster nodes to be ready")
}
