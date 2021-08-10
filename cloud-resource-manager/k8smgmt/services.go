package k8smgmt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
	v1 "k8s.io/api/core/v1"
)

type svcItems struct {
	Items []v1.Service `json:"items"`
}

func GetServices(ctx context.Context, client ssh.Client, names *KubeNames) ([]v1.Service, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "get services", "kconf", names.KconfName)
	svcs := svcItems{}
	if names.DeploymentType == cloudcommon.DeploymentTypeDocker {
		// just populate the service names
		for _, sn := range names.ServiceNames {
			item := v1.Service{}
			item.Name = sn
			svcs.Items = append(svcs.Items, item)
		}
		return svcs.Items, nil
	}
	cmd := fmt.Sprintf("%s kubectl get svc -o json -A", names.KconfEnv)
	out, err := client.Output(cmd)
	if err != nil {
		return nil, fmt.Errorf("can not get list of services: %s, %s, %v", cmd, out, err)
	}
	err = json.Unmarshal([]byte(out), &svcs)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot unmarshal svc json", "out", out, "err", err)
		return nil, fmt.Errorf("cannot unmarshal svc json, %s", err.Error())
	}
	return svcs.Items, nil
}
