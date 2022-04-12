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
