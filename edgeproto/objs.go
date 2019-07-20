package edgeproto

import (
	"errors"
	fmt "fmt"
	"sort"
	"strconv"
	strings "strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/util"
)

//TODO - need to move out Errors into a separate package
var ErrEdgeApiFlavorNotFound = errors.New("Specified flavor not found")
var ErrEdgeApiAppNotFound = errors.New("Specified app not found")

// contains sets of each applications for yaml marshalling
type ApplicationData struct {
	Operators        []Operator        `yaml:"operators"`
	Cloudlets        []Cloudlet        `yaml:"cloudlets"`
	Flavors          []Flavor          `yaml:"flavors"`
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
	sort.Slice(a.ClusterInsts[:], func(i, j int) bool {
		return a.ClusterInsts[i].Key.GetKeyString() < a.ClusterInsts[j].Key.GetKeyString()
	})
	sort.Slice(a.Flavors[:], func(i, j int) bool {
		return a.Flavors[i].Key.GetKeyString() < a.Flavors[j].Key.GetKeyString()
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

func (key *AppKey) Validate() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid app name")
	}
	if !util.ValidName(key.Version) {
		return errors.New("Invalid app version string")
	}
	if err := key.DeveloperKey.Validate(); err != nil {
		return err
	}
	return nil
}

func (s *App) Validate(fields map[string]struct{}) error {
	var err error
	if err = s.GetKey().Validate(); err != nil {
		return err
	}
	if err = s.ValidateEnums(); err != nil {
		return err
	}
	if _, found := fields[AppFieldAccessPorts]; found {
		if s.AccessPorts != "" {
			if _, err = ParseAppPorts(s.AccessPorts); found && err != nil {
				return err
			}
		}
	}
	if s.AuthPublicKey != "" {
		_, err := util.ValidatePublicKey(s.AuthPublicKey)
		if err != nil {
			return err
		}
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
	if _, found := fields[CloudletFieldLocationLatitude]; found {
		if !util.IsLatitudeValid(s.Location.Latitude) {
			return errors.New("Invalid latitude value")
		}
	}
	if _, found := fields[CloudletFieldLocationLongitude]; found {
		if !util.IsLongitudeValid(s.Location.Longitude) {
			return errors.New("Invalid longitude value")
		}
	}
	if err := s.ValidateEnums(); err != nil {
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
	if err := key.ClusterInstKey.Validate(); err != nil {
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

func GetFields(fmap map[string]struct{}) []string {
	var fields []string

	for k, _ := range fmap {
		fields = append(fields, k)
	}

	return fields
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
		return dme.LProto_L_PROTO_TCP, nil
	case "udp":
		return dme.LProto_L_PROTO_UDP, nil
	case "http":
		return dme.LProto_L_PROTO_HTTP, nil
	}
	return 0, fmt.Errorf("%s is not a supported Protocol", s)
}

func LProtoStr(proto dme.LProto) (string, error) {
	switch proto {
	case dme.LProto_L_PROTO_TCP:
		return "tcp", nil
	case dme.LProto_L_PROTO_UDP:
		return "udp", nil
	case dme.LProto_L_PROTO_HTTP:
		return "http", nil
	}
	return "", fmt.Errorf("Invalid proto %d", proto)
}

func L4ProtoStr(proto dme.LProto) (string, error) {
	switch proto {
	case dme.LProto_L_PROTO_HTTP:
		fallthrough
	case dme.LProto_L_PROTO_TCP:
		return "tcp", nil
	case dme.LProto_L_PROTO_UDP:
		return "udp", nil
	}
	return "", fmt.Errorf("Invalid proto %d", proto)
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

func IgnoreTaggedFields(taglist string) []cmp.Option {
	opts := []cmp.Option{}
	opts = append(opts, IgnoreAppFields(taglist))
	opts = append(opts, IgnoreAppInstFields(taglist))
	opts = append(opts, IgnoreAppInstInfoFields(taglist))
	opts = append(opts, IgnoreClusterInstFields(taglist))
	opts = append(opts, IgnoreClusterInstInfoFields(taglist))
	return opts
}

func CmpSortSlices() []cmp.Option {
	opts := []cmp.Option{}
	opts = append(opts, cmpopts.SortSlices(CmpSortApp))
	opts = append(opts, cmpopts.SortSlices(CmpSortAppInst))
	opts = append(opts, cmpopts.SortSlices(CmpSortCloudlet))
	opts = append(opts, cmpopts.SortSlices(CmpSortDeveloper))
	opts = append(opts, cmpopts.SortSlices(CmpSortOperator))
	opts = append(opts, cmpopts.SortSlices(CmpSortClusterInst))
	opts = append(opts, cmpopts.SortSlices(CmpSortFlavor))
	opts = append(opts, cmpopts.SortSlices(CmpSortCloudletInfo))
	opts = append(opts, cmpopts.SortSlices(CmpSortAppInstInfo))
	opts = append(opts, cmpopts.SortSlices(CmpSortClusterInstInfo))
	opts = append(opts, cmpopts.SortSlices(CmpSortNode))
	return opts
}
