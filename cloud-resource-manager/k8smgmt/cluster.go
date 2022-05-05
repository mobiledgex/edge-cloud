// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8smgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgexr/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/edgexr/edge-cloud/cloudcommon"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type NoScheduleMasterTaintAction string

const NoScheduleMasterTaintAdd NoScheduleMasterTaintAction = "master-noschedule-taint-add"
const NoScheduleMasterTaintRemove NoScheduleMasterTaintAction = "master-noschedule-taint-remove"
const NoScheduleMasterTaintNone NoScheduleMasterTaintAction = "master-noschedule-taint-none"

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

func SetMasterNoscheduleTaint(ctx context.Context, client ssh.Client, masterName string, kubeconfig string, action NoScheduleMasterTaintAction) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetMasterNoscheduleTaint", "masterName", masterName, "action", action)

	var cmd string
	if action == NoScheduleMasterTaintAdd {
		log.SpanLog(ctx, log.DebugLevelInfra, "adding taint to master", "masterName", masterName)
		cmd = fmt.Sprintf("kubectl taint nodes %s node-role.kubernetes.io/master=:NoSchedule --kubeconfig=%s", masterName, kubeconfig)
		out, err := client.Output(cmd)
		if err != nil {
			if strings.Contains(out, "already has node-role.kubernetes.io/master") {
				log.SpanLog(ctx, log.DebugLevelInfra, "master taint already present")
			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "error adding master taint", "out", out, "err", err)
				return fmt.Errorf("Cannot add NoSchedule taint to master, %v", err)

			}
		}
	} else if action == NoScheduleMasterTaintRemove {
		log.SpanLog(ctx, log.DebugLevelInfra, "removing taint from master", "masterName", masterName)
		cmd = fmt.Sprintf("kubectl taint nodes %s node-role.kubernetes.io/master:NoSchedule-  --kubeconfig=%s", masterName, kubeconfig)
		out, err := client.Output(cmd)
		if err != nil {
			if strings.Contains(out, "not found") {
				log.SpanLog(ctx, log.DebugLevelInfra, "master taint already gone")
			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "error removing master taint", "out", out, "err", err)
				return fmt.Errorf("Cannot remove NoSchedule taint from master, %v", err)
			}
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
	kconfname := GetKconfName(clusterInst)
	out, err := client.Output("rm " + kconfname)
	if err != nil && !strings.Contains(out, "No such file or directory") {
		return fmt.Errorf("failed to delete kubeconf %s: %v, %v", kconfname, out, err)
	}
	return nil
}

func ClearCluster(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "clearing cluster", "cluster", clusterInst.Key)
	names, err := GetKubeNames(clusterInst, &edgeproto.App{}, &edgeproto.AppInst{})
	if err != nil {
		return err
	}
	// For a single-tenant cluster, all config will be in one dir
	configDir, _ := getConfigDirName(names)
	if err := ClearClusterConfig(ctx, client, configDir, "", names.KconfEnv); err != nil {
		return err
	}
	// For a multi-tenant cluster, each namespace will have a config dir
	cmd := fmt.Sprintf("%s kubectl get ns -o name", names.KconfEnv)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error getting namespaces, %s: %s, %s", cmd, out, err)
	}
	for _, str := range strings.Split(out, "\n") {
		str = strings.TrimSpace(str)
		str = strings.TrimPrefix(str, "namespace/")
		if strings.HasPrefix(str, "kube-") || str == "default" || str == "" {
			continue
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "cleaning config for namespace", "namespace", str)
		nsNames := *names
		nsNames.MultitenantNamespace = str
		configDir, _ := getConfigDirName(&nsNames)
		err = ClearClusterConfig(ctx, client, configDir, str, names.KconfEnv)
		if err != nil {
			return err
		}
		cmd = fmt.Sprintf("%s kubectl delete ns %s", names.KconfEnv, str)
		log.SpanLog(ctx, log.DebugLevelInfra, "deleting extra namespace", "cmd", cmd)
		out, err = client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error deleting namespace, %s: %s, %s", cmd, out, err)
		}
	}

	// delete all helm installs (and leftover junk)
	cmd = fmt.Sprintf("%s helm ls -q", names.KconfEnv)
	out, err = client.Output(cmd)
	if err != nil {
		if strings.Contains(out, "could not find tiller") {
			// helm not installed
			out = ""
		} else {
			return fmt.Errorf("error listing helm instances, %s: %s, %s", cmd, out, err)
		}
	}
	helmServs := []string{}
	for _, name := range strings.Split(out, "\n") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		cmd = fmt.Sprintf("%s helm delete %s", names.KconfEnv, name)
		log.SpanLog(ctx, log.DebugLevelInfra, "deleting helm install", "cmd", cmd)
		out, err = client.Output(cmd)
		if err != nil && !strings.Contains(out, "not found") {
			return fmt.Errorf("error deleting helm install, %s: %s, %s", cmd, out, err)
		}
		helmServs = append(helmServs, name+"-pr-kubelet")
	}
	// If helm prometheus-operator 7.1.1 was installed, pr-kubelet services will
	// be leftover. Need to delete manually.
	if len(helmServs) > 0 {
		cmd = fmt.Sprintf("%s kubectl delete --ignore-not-found --namespace=kube-system service %s", names.KconfEnv, strings.Join(helmServs, " "))
		log.SpanLog(ctx, log.DebugLevelInfra, "deleting helm services", "cmd", cmd)
		out, err = client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error deleting helm services, %s: %s, %s", cmd, out, err)
		}
	}
	// If helm prometheus-operator was installed, CRDs will be leftover.
	// Need to delete manually.
	cmd = fmt.Sprintf("%s kubectl delete --ignore-not-found customresourcedefinitions prometheuses.monitoring.coreos.com servicemonitors.monitoring.coreos.com podmonitors.monitoring.coreos.com alertmanagers.monitoring.coreos.com alertmanagerconfigs.monitoring.coreos.com prometheusrules.monitoring.coreos.com probes.monitoring.coreos.com thanosrulers.monitoring.coreos.com", names.KconfEnv)
	log.SpanLog(ctx, log.DebugLevelInfra, "deleting prometheus CRDs", "cmd", cmd)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deleting prometheus-operator CRDs, %s: %s, %s", cmd, out, err)
	}
	return nil
}

func ClearClusterConfig(ctx context.Context, client ssh.Client, configDir, namespace, kconfEnv string) error {
	// if config dir doesn't exist, then there's no config
	cmd := fmt.Sprintf("stat %s", configDir)
	out, err := client.Output(cmd)
	log.SpanLog(ctx, log.DebugLevelInfra, "clear cluster config", "dir", configDir, "out", out, "err", err)
	if err != nil {
		if strings.Contains(out, "No such file or directory") {
			return nil
		}
		return err
	}
	nsArg := ""
	if namespace != "" {
		nsArg = "-n " + namespace
	}
	// delete all AppInsts configs in cluster
	cmd = fmt.Sprintf("%s kubectl delete %s -f %s", kconfEnv, nsArg, configDir)
	log.SpanLog(ctx, log.DebugLevelInfra, "deleting cluster app", "cmd", cmd)
	out, err = client.Output(cmd)
	// bash returns "does not exist", zsh returns "no matches found"
	if err != nil && !strings.Contains(out, "does not exist") && !strings.Contains(out, "no matches found") {
		for _, msg := range strings.Split(out, "\n") {
			msg = strings.TrimSpace(msg)
			if msg == "" || strings.Contains(msg, " deleted") || strings.Contains(msg, "NotFound") {
				continue
			}
			return fmt.Errorf("error deleting cluster apps, %s: %s, %s", cmd, out, err)
		}
	}
	// delete all AppInst config files
	cmd = fmt.Sprintf("rm %s/*.yaml", configDir)
	log.SpanLog(ctx, log.DebugLevelInfra, "deleting all app config files", "cmd", cmd)
	out, err = client.Output(cmd)
	// bash returns "No such file or directory", zsh returns "no matches found"
	if err != nil && !strings.Contains(out, "No such file or directory") && !strings.Contains(out, "no matches found") {
		return fmt.Errorf("error deleting cluster config files, %s: %s, %s", cmd, out, err)
	}
	// remove configDir
	cmd = fmt.Sprintf("rmdir %s", configDir)
	log.SpanLog(ctx, log.DebugLevelInfra, "removing config dir", "cmd", cmd)
	out, err = client.Output(cmd)
	if err != nil && !strings.Contains(out, "Directory not empty") && !strings.Contains(out, "No such file or directory") {
		return fmt.Errorf("error removing config dir, %s: %s, %s", cmd, out, err)
	}
	return nil
}

type Nodes struct {
	ApiVersion string    `json:"apiVersion"`
	Items      []v1.Node `json:"items"`
}

func GetNodeInfos(ctx context.Context, client ssh.Client, kconfEnv string) ([]*edgeproto.NodeInfo, error) {
	cmd := fmt.Sprintf("%s kubectl get nodes --output=json", kconfEnv)
	log.SpanLog(ctx, log.DebugLevelInfra, "GetNodeInfo", "cmd", cmd)
	out, err := client.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("get nodes failed, %s, %v", out, err)
	}
	nodes := Nodes{}
	err = json.Unmarshal([]byte(out), &nodes)
	if err != nil {
		return nil, err
	}
	info := []*edgeproto.NodeInfo{}
	for _, item := range nodes.Items {
		nodeInfo := &edgeproto.NodeInfo{}
		nodeInfo.Name = item.Name
		nodeInfo.Allocatable = make(map[string]*edgeproto.Udec64)
		nodeInfo.Capacity = make(map[string]*edgeproto.Udec64)
		for res, quantity := range item.Status.Allocatable {
			name, dec, err := convertNodeResource(res, quantity)
			if err == nil && name == unsupportedResource {
				continue
			}
			if err != nil {
				return nil, err
			}
			nodeInfo.Allocatable[name] = dec
		}
		for res, quantity := range item.Status.Capacity {
			name, dec, err := convertNodeResource(res, quantity)
			if err == nil && name == unsupportedResource {
				continue
			}
			if err != nil {
				return nil, err
			}
			nodeInfo.Capacity[name] = dec
		}
		info = append(info, nodeInfo)
	}
	return info, nil
}

var unsupportedResource = "unsupported"

func convertNodeResource(res v1.ResourceName, quantity resource.Quantity) (string, *edgeproto.Udec64, error) {
	var name string
	scale := uint64(1)
	switch res {
	case v1.ResourceCPU:
		name = cloudcommon.ResourceVcpus
	case v1.ResourceMemory:
		name = cloudcommon.ResourceRamMb
		scale = 1024 * 1024
	case v1.ResourceEphemeralStorage:
		name = cloudcommon.ResourceDiskGb
		scale = 1024 * 1024 * 1024
	default:
		// unsupported
		return unsupportedResource, nil, nil
	}
	dec, err := QuantityToUdec64(quantity)
	if err != nil {
		return name, nil, fmt.Errorf("Resource %s, %v", name, err)
	}
	if scale != 1 {
		dec.Whole /= scale
	}
	return name, dec, nil
}
