package cloudcommon

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

// The following structures describe the yaml structure of the templates we generate for k8s
type MatchLabels struct {
	Run string `yaml:"run"`
}
type SpecSelector struct {
	Labels MatchLabels `yaml:"matchLabels"`
}
type Volume struct {
	Name string `yaml:"name"`
}
type ImagePullSecrets struct {
}
type K8sLocalObject struct {
	Name string `yaml:"name"`
}
type SecretKeyRef struct {
	K8sLocalObject `yaml:",inline"`
	Key            string `yaml:"key"`
	Optional       bool   `yaml:"optional,omitempty"`
}
type ConfigMapKeyRef struct {
	K8sLocalObject `yaml:",inline"`
	Key            string `yaml:"key"`
	Optional       bool   `yaml:"optional,omitempty"`
}
type ContainerEnvValRef struct {
	ConfigMapKey ConfigMapKeyRef `yaml:"configMapKeyRef,omitempty"`
	SecrecKey    SecretKeyRef    `yaml:"secretKeyRef,omitempty"`
}
type ContainerEnvVars struct {
	Name      string             `yaml:"name"`
	Value     string             `yaml:"value,omitempty"`
	ValueFrom ContainerEnvValRef `yaml:"valueFrom,omitempty"`
}
type ContainerPorts struct {
	ContainerPort int32  `yaml:"containerPort"`
	Protocol      string `yaml:"protocol"`
}
type K8sContainer struct {
	Name            string             `yaml:"name"`
	Image           string             `yaml:"image"`
	ImagePullPolicy string             `yaml:"imagePullPolicy"`
	EnvVars         []ContainerEnvVars `yaml:"env,omitempty"`
	Ports           []ContainerPorts   `yaml:"ports"`
	Command         []string           `yaml:"command,omitempty"`
}
type TemplateSpec struct {
	Volumes          []Volume         `yaml:"volumes,omitempty"`
	ImagePullSecrets []K8sLocalObject `yaml:"imagePullSecrets,omitempty"`
	Containers       []K8sContainer   `yaml:"containers"`
}
type EdgeK8sMetadataLabels struct {
	Run string `yaml:"run"`
}
type EdgeK8sMetadata struct {
	Name   string                `yaml:"name,omitempty"`
	Labels EdgeK8sMetadataLabels `yaml:"labels,omitempty"`
}
type DeploymentTemplate struct {
	Metadata EdgeK8sMetadata `yaml:"metadata"`
	Spec     TemplateSpec    `yaml:"spec"`
}
type EdgeK8sDeploymentSpec struct {
	Selector SpecSelector       `yaml:"selector"`
	Replicas int32              `yaml:"replicas"`
	Template DeploymentTemplate `yaml:"template"`
}

type EdgeK8sServicePort struct {
	Name       string `yaml:"name"`
	Protocol   string `yaml:"protocol"`
	Port       int32  `yaml:"port"`
	TargetPort int32  `yaml:"targetPort"`
}
type ServiceSelector struct {
	Run string `yaml:"run"`
}
type EdgeK8sServiceSpec struct {
	Type     string               `yaml:"type"`
	Ports    []EdgeK8sServicePort `yaml:"ports,omitempty"`
	Selector ServiceSelector      `yaml:"selector"`
}
type EdgeK8sManifest struct {
	ApiVersion string          `yaml:"apiVerstion"`
	Kind       string          `yaml:"kind"`
	Metadata   EdgeK8sMetadata `yaml:"metadata"`
	Spec       interface{}     `yaml:"spec"`
}

func DecodeK8SYaml(manifest string) ([]runtime.Object, []*schema.GroupVersionKind, error) {
	files := strings.Split(manifest, "---")
	decode := scheme.Codecs.UniversalDeserializer().Decode
	objs := []runtime.Object{}
	kinds := []*schema.GroupVersionKind{}

	for _, file := range files {
		obj, kind, err := decode([]byte(file), nil, nil)
		if err != nil {
			return nil, nil, err
		}
		objs = append(objs, obj)
		kinds = append(kinds, kind)
	}
	return objs, kinds, nil
}
