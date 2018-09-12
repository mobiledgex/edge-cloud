package edgeproto

import (
	"errors"
	"sort"
	strings "strings"

	"github.com/mobiledgex/edge-cloud/util"
)

// contains sets of each applications for yaml marshalling
type ApplicationData struct {
	Operators        []Operator        `yaml:"operators"`
	Cloudlets        []Cloudlet        `yaml:"cloudlets"`
	Flavors          []Flavor          `yaml:"flavors"`
	ClusterFlavors   []ClusterFlavor   `yaml:"clusterflavors"`
	Clusters         []Cluster         `yaml:"clusters"`
	ClusterInsts     []ClusterInst     `yaml:"clusterinsts"`
	Developers       []Developer       `yaml:"developers"`
	Applications     []App             `yaml:"apps"`
	AppInstances     []AppInst         `yaml:"appinstances"`
	CloudletInfos    []CloudletInfo    `yaml:"cloudletinfos"`
	AppInstInfos     []AppInstInfo     `yaml:"appinstinfos"`
	ClusterInstInfos []ClusterInstInfo `yaml:"clusterinstinfos"`
}

// sort each slice by key
func (a *ApplicationData) Sort() {
	sort.Slice(a.AppInstances[:], func(i, j int) bool {
		return a.AppInstances[i].Key.GetKeyString() < a.AppInstances[j].Key.GetKeyString()
	})
	sort.Slice(a.Applications[:], func(i, j int) bool {
		return a.Applications[i].Key.GetKeyString() < a.Applications[j].Key.GetKeyString()
	})
	sort.Slice(a.Cloudlets[:], func(i, j int) bool {
		return a.Cloudlets[i].Key.GetKeyString() < a.Cloudlets[j].Key.GetKeyString()
	})
	sort.Slice(a.Developers[:], func(i, j int) bool {
		return a.Developers[i].Key.GetKeyString() < a.Developers[j].Key.GetKeyString()
	})
	sort.Slice(a.Operators[:], func(i, j int) bool {
		return a.Operators[i].Key.GetKeyString() < a.Operators[j].Key.GetKeyString()
	})
	sort.Slice(a.Clusters[:], func(i, j int) bool {
		return a.Clusters[i].Key.GetKeyString() < a.Clusters[j].Key.GetKeyString()
	})
	sort.Slice(a.ClusterInsts[:], func(i, j int) bool {
		return a.ClusterInsts[i].Key.GetKeyString() < a.ClusterInsts[j].Key.GetKeyString()
	})
	sort.Slice(a.Flavors[:], func(i, j int) bool {
		return a.Flavors[i].Key.GetKeyString() < a.Flavors[j].Key.GetKeyString()
	})
	sort.Slice(a.ClusterFlavors[:], func(i, j int) bool {
		return a.ClusterFlavors[i].Key.GetKeyString() < a.ClusterFlavors[j].Key.GetKeyString()
	})
	sort.Slice(a.CloudletInfos[:], func(i, j int) bool {
		return a.CloudletInfos[i].Key.GetKeyString() < a.CloudletInfos[j].Key.GetKeyString()
	})
	sort.Slice(a.AppInstInfos[:], func(i, j int) bool {
		return a.AppInstInfos[i].Key.GetKeyString() < a.AppInstInfos[j].Key.GetKeyString()
	})
	sort.Slice(a.ClusterInstInfos[:], func(i, j int) bool {
		return a.ClusterInstInfos[i].Key.GetKeyString() < a.ClusterInstInfos[j].Key.GetKeyString()
	})
}

// Validate functions to validate user input

func (key *DeveloperKey) Validate() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid developer name")
	}
	return nil
}

func (s *Developer) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
}

func (key *OperatorKey) Validate() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid operator name")
	}
	return nil
}

func (s *Operator) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
}

func (key *ClusterKey) Validate() error {
	if !util.ValidKubernetesName(key.Name) {
		return errors.New("Invalid cluster name")
	}
	return nil
}

func (s *Cluster) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
}

func (key *ClusterInstKey) Validate() error {
	if err := key.ClusterKey.Validate(); err != nil {
		return err
	}
	if err := key.CloudletKey.Validate(); err != nil {
		return err
	}
	return nil
}

func (s *ClusterInst) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
}

func (key *FlavorKey) Validate() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid flavor name")
	}
	return nil
}

func (s *Flavor) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
}

func (key *ClusterFlavorKey) Validate() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid cluster flavor name")
	}
	return nil
}

func (s *ClusterFlavor) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
}

func (key *AppKey) Validate() error {
	if err := key.DeveloperKey.Validate(); err != nil {
		return err
	}
	if !util.ValidName(key.Name) {
		return errors.New("Invalid app name")
	}
	if !util.ValidName(key.Version) {
		return errors.New("Invalid app version string")
	}
	return nil
}

func (s *App) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
}

func (key *CloudletKey) Validate() error {
	if err := key.OperatorKey.Validate(); err != nil {
		return err
	}
	if !util.ValidName(key.Name) {
		return errors.New("Invalid cloudlet name")
	}
	return nil
}

func (s *Cloudlet) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().Validate(); err != nil {
		return err
	}
	return nil
}

func (s *CloudletInfo) Validate(fields map[string]struct{}) error {
	return nil
}

func (key *AppInstKey) Validate() error {
	if err := key.AppKey.Validate(); err != nil {
		return err
	}
	if err := key.CloudletKey.Validate(); err != nil {
		return err
	}
	if key.Id == 0 {
		return errors.New("AppInst Id cannot be zero")
	}
	return nil
}

func (s *AppInst) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().Validate(); err != nil {
		return err
	}
	return nil
}

func (s *AppInstInfo) Validate(fields map[string]struct{}) error {
	return nil
}

func (s *ClusterInstInfo) Validate(fields map[string]struct{}) error {
	return nil
}

func (s *CloudletRefs) Validate(fields map[string]struct{}) error {
	return nil
}

func (s *ClusterRefs) Validate(fields map[string]struct{}) error {
	return nil
}

func MakeFieldMap(fields []string) map[string]struct{} {
	fmap := make(map[string]struct{})
	if fields == nil {
		return fmap
	}
	for _, set := range fields {
		for {
			fmap[set] = struct{}{}
			idx := strings.LastIndex(set, ".")
			if idx == -1 {
				break
			}
			set = set[:idx]
		}
	}
	return fmap
}

func HasField(fmap map[string]struct{}, field string) bool {
	_, ok := fmap[field]
	return ok
}

func (m *Metric) AddTag(name string, val string) {
	tag := MetricTag{Name: name, Val: val}
	m.Tags = append(m.Tags, &tag)
}

func (m *Metric) AddDoubleVal(name string, dval float64) {
	val := MetricVal{Name: name}
	val.Value = &MetricVal_Dval{Dval: dval}
	m.Vals = append(m.Vals, &val)
}

func (m *Metric) AddIntVal(name string, ival uint64) {
	val := MetricVal{Name: name}
	val.Value = &MetricVal_Ival{Ival: ival}
	m.Vals = append(m.Vals, &val)
}
