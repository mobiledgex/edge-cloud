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

package deploygen

import (
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/util"
)

var KubernetesBasic = "kubernetes-basic"
var MexDeployGenLabel = "mexDeployGen"

var Generators = map[string]func(app *AppSpec) (string, error){
	KubernetesBasic: kubeBasic,
}

type AppSpec struct {
	Name             string          `json:"name"`
	OrgName          string          `json:"orgname"`
	Version          string          `json:"version"`
	ImagePath        string          `json:"imagepath"`
	ImageType        string          `json:"imagetype"`
	Command          string          `json:"command"`
	Config           string          `json:"config"`
	Annotations      string          `json:"annotations"`
	Ports            []util.PortSpec `json:"ports"`
	ScaleWithCluster bool            `json:"scalewithcluster"`
	ImageHost        string          `json:"imagehost"`
}

func NewAppSpec(app *edgeproto.App) (*AppSpec, error) {
	var err error
	out := &AppSpec{
		Name:             app.Key.Name,
		OrgName:          app.Key.Organization,
		Version:          app.Key.Version,
		ImagePath:        app.ImagePath,
		ImageType:        edgeproto.ImageType_name[int32(app.ImageType)],
		Command:          app.Command,
		Annotations:      app.Annotations,
		ScaleWithCluster: app.ScaleWithCluster,
	}
	urlObj, err := util.ImagePathParse(app.ImagePath)
	if err != nil {
		return nil, err
	}
	out.ImageHost = urlObj.Host

	if app.AccessPorts == "" {
		return out, nil
	}
	out.Ports, err = util.ParsePorts(app.AccessPorts)
	if err != nil {
		return nil, err
	}
	return out, nil
}
