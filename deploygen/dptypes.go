package deploygen

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

var KubernetesBasic = "kubernetes-basic"

var Generators = map[string]func(app *AppSpec) (string, error){
	KubernetesBasic: kubeBasic,
}

type AppSpec struct {
	Name             string          `json:"name"`
	DevName          string          `json:"devname"`
	Version          string          `json:"version"`
	ImagePath        string          `json:"imagepath"`
	ImageType        string          `json:"imagetype"`
	Command          string          `json:"command"`
	Config           string          `json:"config"`
	Annotations      string          `json:"annotations"`
	Ports            []util.PortSpec `json:"ports"`
	ScaleWithCluster bool            `json:"scalewithcluster"`
}

func NewAppSpec(app *edgeproto.App) (*AppSpec, error) {
	var err error
	out := &AppSpec{
		Name:             app.Key.Name,
		DevName:          app.Key.DeveloperKey.Name,
		Version:          app.Key.Version,
		ImagePath:        app.ImagePath,
		ImageType:        edgeproto.ImageType_name[int32(app.ImageType)],
		Command:          app.Command,
		Annotations:      app.Annotations,
		ScaleWithCluster: app.ScaleWithCluster,
	}
	if app.AccessPorts == "" {
		return out, nil
	}
	out.Ports, err = util.ParsePorts(app.AccessPorts)
	if err != nil {
		return nil, err
	}
	return out, nil
}
