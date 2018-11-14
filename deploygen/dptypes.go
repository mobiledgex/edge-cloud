package deploygen

import (
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var KubernetesBasic = "kubernetes-basic"

var Generators = map[string]func(app *AppSpec) (string, error){
	KubernetesBasic: kubeBasic,
}

type PortSpec struct {
	Proto string `json:"proto"`
	Port  string `json:"port"`
}

type AppSpec struct {
	Name        string     `json:"name"`
	DevName     string     `json:"devname"`
	Version     string     `json:"version"`
	ImagePath   string     `json:"imagepath"`
	ImageType   string     `json:"imagetype"`
	Command     string     `json:"command"`
	Config      string     `json:"config"`
	Annotations string     `json:"annotations"`
	Ports       []PortSpec `json:"ports"`
}

func NewAppFromApp(app *edgeproto.App) (*AppSpec, error) {
	out := &AppSpec{
		Name:        app.Key.Name,
		DevName:     app.Key.DeveloperKey.Name,
		Version:     app.Key.Version,
		ImagePath:   app.ImagePath,
		ImageType:   edgeproto.ImageType_name[int32(app.ImageType)],
		Command:     app.Command,
		Annotations: app.Annotations,
	}
	err := setPorts(out, app)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func setPorts(app *AppSpec, in *edgeproto.App) error {
	ports := []PortSpec{}
	pstrs := strings.Split(in.AccessPorts, ",")
	for _, pstr := range pstrs {
		pp := strings.Split(pstr, ":")
		if len(pp) != 2 {
			return fmt.Errorf("invalid AccessPorts format %s", pstr)
		}
		port := PortSpec{
			Proto: pp[0],
			Port:  pp[1],
		}
		ports = append(ports, port)
	}
	app.Ports = ports
	return nil
}
