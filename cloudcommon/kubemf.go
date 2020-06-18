package cloudcommon

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

func DecodeK8SYaml(manifest string) ([]runtime.Object, []*schema.GroupVersionKind, error) {
	files := strings.Split(manifest, "---")
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
