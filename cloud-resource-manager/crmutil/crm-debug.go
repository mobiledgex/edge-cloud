package crmutil

import (
	"context"
	"encoding/csv"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

const (
	GetEnvoyVersionCmd = "get-cluster-envoy-version"
	RefreshRootLBCerts = "refresh-rootlb-certs"
	CRMCmd             = "crmcmd"
)

func InitDebug(nodeMgr *node.NodeMgr) {
	nodeMgr.Debug.AddDebugFunc(CRMCmd, runCrmCmd)
}

func runCrmCmd(ctx context.Context, req *edgeproto.DebugRequest) string {
	if req.Args == "" {
		return "please specify shell command in args field"
	}
	rd := csv.NewReader(strings.NewReader(req.Args))
	rd.Comma = ' '
	args, err := rd.Read()
	if err != nil {
		return fmt.Sprintf("failed to split args string, %v", err)
	}
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("exec failed, %v, %s", err, string(out))
	}
	return string(out)
}
