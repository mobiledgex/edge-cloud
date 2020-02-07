package crmutil

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"

	"github.com/mobiledgex/edge-cloud/log"
)

var RemoteServerNone = ""

// RunCommand runs a command either locally or on the remote server
func RunCommand(ctx context.Context, plat platform.Platform, client pc.PlatformClient, remoteServer string, cmd string) (string, error) {
	log.DebugLog(log.DebugLevelMexos, "RunCommand", "remoteServer", remoteServer, "cmd", cmd)
	if remoteServer == RemoteServerNone {
		out, err := client.Output(cmd)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "cmd failed", "cmd", cmd, "err", err, "out", out)
			return out, fmt.Errorf("command \"%s\" failed, %v", cmd, err)
		}
		return out, err
	}
	out, err := plat.RunRemoteCommand(ctx, client, remoteServer, cmd)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "cmd failed", "cmd", cmd, "err", err, "out", out)
		return out, fmt.Errorf("remote command \"%s\" failed, %v", cmd, err)
	}
	return out, nil
}
