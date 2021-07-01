package cloudcommon

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	yaml "github.com/mobiledgex/yaml/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

const yamlSeparator = "\n---"

func DecodeK8SYaml(manifest string) ([]runtime.Object, []*schema.GroupVersionKind, error) {
	files := strings.Split(manifest, yamlSeparator)
	decode := scheme.Codecs.UniversalDeserializer().Decode
	objs := []runtime.Object{}
	kinds := []*schema.GroupVersionKind{}

	for _, file := range files {
		file = strings.TrimSpace(file)
		if len(file) == 0 {
			continue
		}
		obj, kind, err := decode([]byte(file), nil, nil)
		if err != nil {
			return nil, nil, err
		}
		objs = append(objs, obj)
		kinds = append(kinds, kind)
	}
	return objs, kinds, nil
}

type DockerContainer struct {
	Image string `mapstructure:"image"`
}

func DecodeDockerComposeYaml(manifest string) (map[string]DockerContainer, error) {
	obj := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(manifest), &obj)
	if err != nil {
		return nil, err
	}
	if _, ok := obj["services"]; !ok {
		return nil, fmt.Errorf("unable to find services in docker compose file")
	}
	containers := make(map[string]DockerContainer)
	err = mapstructure.Decode(obj["services"], &containers)
	if err != nil {
		return nil, err
	}
	return containers, nil
}
