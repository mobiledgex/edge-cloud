package edgeproto

import (
	"errors"
	fmt "fmt"
	"sort"
	"strconv"
	strings "strings"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
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
	Nodes            []Node            `yaml:"nodes"`
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
	sort.Slice(a.Nodes[:], func(i, j int) bool {
		return a.Nodes[i].Key.GetKeyString() < a.Nodes[j].Key.GetKeyString()
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
	err := s.GetKey().Validate()
	if err != nil {
		return err
	}
	if _, found := fields[FlavorFieldRam]; found && s.Ram == 0 {
		return errors.New("Ram cannot be 0")
	}
	if _, found := fields[FlavorFieldVcpus]; found && s.Vcpus == 0 {
		return errors.New("Vcpus cannot be 0")
	}
	if _, found := fields[FlavorFieldDisk]; found && s.Disk == 0 {
		return errors.New("Disk cannot be 0")
	}
	return nil
}

func (key *ClusterFlavorKey) Validate() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid cluster flavor name")
	}
	return nil
}

func (s *ClusterFlavor) Validate(fields map[string]struct{}) error {
	err := s.GetKey().Validate()
	if err != nil {
		return err
	}
	if _, found := fields[ClusterFlavorFieldNodeFlavor]; found && s.NodeFlavor.Name == "" {
		return errors.New("Invalid empty string for Node Flavor")
	}
	if _, found := fields[ClusterFlavorFieldMasterFlavor]; found && s.MasterFlavor.Name == "" {
		return errors.New("Invalid empty string for Master Flavor")
	}
	if _, found := fields[ClusterFlavorFieldNumNodes]; found && s.NumNodes == 0 {
		return errors.New("Number of nodes cannot be 0")
	}
	if _, found := fields[ClusterFlavorFieldMaxNodes]; found && s.MaxNodes == 0 {
		return errors.New("Number of maximum nodes cannot be 0")
	}
	if _, found := fields[ClusterFlavorFieldNumMasters]; found && s.NumMasters == 0 {
		return errors.New("Number of master nodes cannot be 0")
	}
	return nil
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
	var err error
	if err = s.GetKey().Validate(); err != nil {
		return err
	}
	if s.ImageType == ImageType_ImageTypeUnknown {
		return errors.New("Please specify Image Type")
	}
	if err = s.ValidateEnums(); err != nil {
		return err
	}
	if s.AccessPorts == "" {
		return errors.New("Please specify access ports")
	}
	if _, err = ParseAppPorts(s.AccessPorts); err != nil {
		return err
	}
	return nil
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
	return nil
}

func (s *AppInst) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().Validate(); err != nil {
		return err
	}
	return nil
}

func (key *ControllerKey) Validate() error {
	if key.Addr == "" {
		return errors.New("Invalid address")
	}
	return nil
}

func (s *Controller) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
}

func (key *NodeKey) Validate() error {
	if key.Name == "" {
		return errors.New("Invalid node name")
	}
	return key.CloudletKey.Validate()
}

func (s *Node) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
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

func GetLProto(s string) (dme.LProto, error) {
	s = strings.ToLower(s)
	switch s {
	case "tcp":
		return dme.LProto_LProtoTCP, nil
	case "udp":
		return dme.LProto_LProtoUDP, nil
	case "http":
		return dme.LProto_LProtoHTTP, nil
	}
	return 0, fmt.Errorf("%s is not a supported Protocol", s)
}

func ParseAppPorts(ports string) ([]dme.AppPort, error) {
	appports := make([]dme.AppPort, 0)
	strs := strings.Split(ports, ",")
	for _, str := range strs {
		vals := strings.Split(str, ":")
		if len(vals) != 2 {
			return nil, fmt.Errorf("Invalid Access Ports format, expected proto:port but was %s", vals[0])
		}
		proto, err := GetLProto(vals[0])
		if err != nil {
			return nil, err
		}
		port, err := strconv.ParseInt(vals[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert port %s to integer: %s", vals[1], err)
		}
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("Port %s out of range", vals[1])
		}
		p := dme.AppPort{
			Proto:        proto,
			InternalPort: int32(port),
		}
		appports = append(appports, p)
	}
	return appports, nil
}
