package deploygen

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
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
