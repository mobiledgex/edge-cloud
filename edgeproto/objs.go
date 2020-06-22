package edgeproto

import (
	"errors"
	fmt "fmt"
	"net"
	"sort"
	"strconv"
	strings "strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/util"
	context "golang.org/x/net/context"
)

var AutoScaleMaxNodes uint32 = 10

var minPort uint32 = 1
var maxPort uint32 = 65535

const (
	AppConfigHelmYaml      = "helmCustomizationYaml"
	AppAccessCustomization = "appAccessCustomization"
	AppConfigEnvYaml       = "envVarsYaml"
)

var ValidConfigKinds = map[string]struct{}{
	AppConfigHelmYaml:      struct{}{},
	AppAccessCustomization: struct{}{},
	AppConfigEnvYaml:       struct{}{},
}

// sort each slice by key
func (a *AllData) Sort() {
	sort.Slice(a.AppInstances[:], func(i, j int) bool {
		return a.AppInstances[i].Key.GetKeyString() < a.AppInstances[j].Key.GetKeyString()
	})
	sort.Slice(a.Apps[:], func(i, j int) bool {
		return a.Apps[i].Key.GetKeyString() < a.Apps[j].Key.GetKeyString()
	})
	sort.Slice(a.Cloudlets[:], func(i, j int) bool {
		return a.Cloudlets[i].Key.GetKeyString() < a.Cloudlets[j].Key.GetKeyString()
	})
	sort.Slice(a.OperatorCodes[:], func(i, j int) bool {
		return a.OperatorCodes[i].GetKey().GetKeyString() < a.OperatorCodes[j].GetKey().GetKeyString()
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
	sort.Slice(a.CloudletPools[:], func(i, j int) bool {
		return a.CloudletPools[i].Key.GetKeyString() < a.CloudletPools[j].Key.GetKeyString()
	})
	sort.Slice(a.CloudletPoolMembers[:], func(i, j int) bool {
		return a.CloudletPoolMembers[i].GetKeyString() < a.CloudletPoolMembers[j].GetKeyString()
	})
	sort.Slice(a.AutoScalePolicies[:], func(i, j int) bool {
		return a.AutoScalePolicies[i].Key.GetKeyString() < a.AutoScalePolicies[j].Key.GetKeyString()
	})
	sort.Slice(a.AutoProvPolicies[:], func(i, j int) bool {
		return a.AutoProvPolicies[i].Key.GetKeyString() < a.AutoProvPolicies[j].Key.GetKeyString()
	})
	sort.Slice(a.PrivacyPolicies[:], func(i, j int) bool {
		return a.PrivacyPolicies[i].Key.GetKeyString() < a.PrivacyPolicies[j].Key.GetKeyString()
	})
	sort.Slice(a.AutoProvPolicyCloudlets[:], func(i, j int) bool {
		if a.AutoProvPolicyCloudlets[i].Key.GetKeyString() == a.AutoProvPolicyCloudlets[j].Key.GetKeyString() {
			return a.AutoProvPolicyCloudlets[i].CloudletKey.GetKeyString() < a.AutoProvPolicyCloudlets[j].CloudletKey.GetKeyString()
		}
		return a.AutoProvPolicyCloudlets[i].Key.GetKeyString() < a.AutoProvPolicyCloudlets[j].Key.GetKeyString()
	})
	sort.Slice(a.ResTagTables[:], func(i, j int) bool {
		return a.ResTagTables[i].Key.GetKeyString() < a.ResTagTables[j].Key.GetKeyString()
	})
	sort.Slice(a.AppInstRefs[:], func(i, j int) bool {
		return a.AppInstRefs[i].Key.GetKeyString() < a.AppInstRefs[j].Key.GetKeyString()
	})
}

func (a *NodeData) Sort() {
	sort.Slice(a.Nodes[:], func(i, j int) bool {
		// ignore name for sorting because it is ignored for comparison
		ikey := a.Nodes[i].Key
		ikey.Name = ""
		jkey := a.Nodes[j].Key
		jkey.Name = ""
		return ikey.GetKeyString() < jkey.GetKeyString()
	})
}

// Validate functions to validate user input

func (key *OperatorCodeKey) ValidateKey() error {
	if key.GetKeyString() == "" {
		return errors.New("No code specified")
	}
	return nil
}

func (s *OperatorCode) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
		return err
	}
	if s.Organization == "" {
		return errors.New("No organization specified")
	}
	return nil
}

func (key *ClusterKey) ValidateKey() error {
	if !util.ValidKubernetesName(key.Name) {
		return errors.New("Invalid cluster name")
	}
	return nil
}

func (key *ClusterInstKey) ValidateKey() error {
	if err := key.ClusterKey.ValidateKey(); err != nil {
		return err
	}
	if err := key.CloudletKey.ValidateKey(); err != nil {
		return err
	}
	return nil
}

func (s *ClusterInst) Validate(fields map[string]struct{}) error {
	return s.GetKey().ValidateKey()
}

func (key *FlavorKey) ValidateKey() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid flavor name")
	}
	return nil
}

func (s *Flavor) Validate(fields map[string]struct{}) error {
	err := s.GetKey().ValidateKey()
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

func (key *AppKey) ValidateKey() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid app name")
	}
	if !util.ValidName(key.Version) {
		return errors.New("Invalid app version string")
	}
	if !util.ValidName(key.Organization) {
		return errors.New("Invalid organization name")
	}
	return nil
}

func validateCustomizationConfigs(configs []*ConfigFile) error {
	for _, cfg := range configs {
		if _, found := ValidConfigKinds[cfg.Kind]; !found {
			return fmt.Errorf("Invalid Config Kind - %s", cfg.Kind)
		}
	}
	return nil
}

func (s *App) Validate(fields map[string]struct{}) error {
	var err error
	if err = s.GetKey().ValidateKey(); err != nil {
		return err
	}
	if err = s.ValidateEnums(); err != nil {
		return err
	}
	if _, found := fields[AppFieldAccessPorts]; found {
		if s.AccessPorts != "" {
			_, err = ParseAppPorts(s.AccessPorts)
			if err != nil {
				return err
			}
		}
	}
	if s.AuthPublicKey != "" {
		_, err = util.ValidatePublicKey(s.AuthPublicKey)
		if err != nil {
			return err
		}
	}
	if s.TemplateDelimiter != "" {
		out := strings.Split(s.TemplateDelimiter, " ")
		if len(out) != 2 {
			return fmt.Errorf("invalid app template delimiter %s, valid format '<START-DELIM> <END-DELIM>'", s.TemplateDelimiter)
		}
	}
	if err = validateCustomizationConfigs(s.Configs); err != nil {
		return err
	}
	return nil
}

func (key *CloudletKey) ValidateKey() error {
	if !util.ValidName(key.Organization) {
		return errors.New("Invalid organization name")
	}
	if !util.ValidName(key.Name) {
		return errors.New("Invalid cloudlet name")
	}
	return nil
}

func (s *Cloudlet) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
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
	if _, found := fields[CloudletFieldMaintenanceState]; found {
		if s.MaintenanceState != MaintenanceState_NORMAL_OPERATION && s.MaintenanceState != MaintenanceState_MAINTENANCE_START && s.MaintenanceState != MaintenanceState_MAINTENANCE_START_NO_FAILOVER {
			return errors.New("Invalid maintenance state, only normal operation and maintenance start states are allowed")
		}
	}
	if s.VmImageVersion != "" {
		if err := util.ValidateImageVersion(s.VmImageVersion); err != nil {
			return err
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

func (key *CloudletPoolKey) ValidateKey() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid Cloudlet Pool name")
	}
	return nil
}

func (s *CloudletPool) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
		return err
	}
	return nil
}

func (key *CloudletPoolMember) ValidateKey() error {
	if err := key.CloudletKey.ValidateKey(); err != nil {
		return err
	}
	if err := key.PoolKey.ValidateKey(); err != nil {
		return err
	}
	return nil
}

func (s *CloudletPoolMember) Validate(fields map[string]struct{}) error {
	return s.ValidateKey()
}

func (key *ResTagTableKey) ValidateKey() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid ResTagTable name")
	}
	return nil
}

func (s *ResTagTable) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
		return err
	}
	return nil
}

func (key *AppInstKey) ValidateKey() error {
	if err := key.AppKey.ValidateKey(); err != nil {
		return err
	}
	if err := key.ClusterInstKey.ValidateKey(); err != nil {
		return err
	}
	return nil
}

func (s *AppInst) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
		return err
	}
	if err := validateCustomizationConfigs(s.Configs); err != nil {
		return err
	}
	return nil
}

func (key *ControllerKey) ValidateKey() error {
	if key.Addr == "" {
		return errors.New("Invalid address")
	}
	return nil
}

func (s *Controller) Validate(fields map[string]struct{}) error {
	return s.GetKey().ValidateKey()
}

func (key *NodeKey) ValidateKey() error {
	if key.Name == "" {
		return errors.New("Invalid node name")
	}
	return key.CloudletKey.ValidateKey()
}

func (s *Node) Validate(fields map[string]struct{}) error {
	return s.GetKey().ValidateKey()
}

func (key *AlertKey) ValidateKey() error {
	if len(string(*key)) == 0 {
		return errors.New("Invalid empty string AlertKey")
	}
	return nil
}

func (s *Alert) Validate(fields map[string]struct{}) error {
	return s.GetKey().ValidateKey()
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

func (s *AppInstRefs) Validate(fields map[string]struct{}) error {
	return nil
}

func (key *PolicyKey) ValidateKey() error {
	if err := util.ValidObjName(key.Organization); err != nil {
		errstring := err.Error()
		// lowercase the first letter of the error message
		errstring = strings.ToLower(string(errstring[0])) + errstring[1:len(errstring)]
		return fmt.Errorf("Invalid organization, " + errstring)
	}
	if key.Name == "" {
		return errors.New("Policy name cannot be empty")
	}
	return nil
}

func (s *AppInstClientKey) Validate(fields map[string]struct{}) error {
	return s.Key.ValidateKey()
}

// Validate fields. Note that specified fields is ignored, so this function
// must be used only in the context when all fields are present (i.e. after
// CopyInFields for an update).
func (s *AutoScalePolicy) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
		return err
	}
	if s.MaxNodes > AutoScaleMaxNodes {
		return fmt.Errorf("Max nodes cannot exceed %d", AutoScaleMaxNodes)
	}
	if s.MinNodes < 1 {
		// Taint on master is not updated during UpdateClusterInst
		// when going up/down from 0, so min supported is 1.
		return errors.New("Min nodes cannot be less than 1")
	}
	if s.ScaleUpCpuThresh < 0 || s.ScaleUpCpuThresh > 100 {
		return errors.New("Scale up CPU threshold must be between 0 and 100")
	}
	if s.ScaleDownCpuThresh < 0 || s.ScaleDownCpuThresh > 100 {
		return errors.New("Scale down CPU threshold must be between 0 and 100")
	}
	if s.MaxNodes <= s.MinNodes {
		return fmt.Errorf("Max nodes must be greater than Min nodes")
	}
	if s.ScaleUpCpuThresh <= s.ScaleDownCpuThresh {
		return fmt.Errorf("Scale down cpu threshold must be less than scale up cpu threshold")
	}
	return nil
}

func (s *AutoProvPolicy) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
		return err
	}
	if s.MinActiveInstances > s.MaxInstances && s.MaxInstances != 0 {
		return fmt.Errorf("Minimum active instances cannot be larger than Maximum Instances")
	}
	if s.MinActiveInstances == 0 && s.DeployClientCount == 0 {
		return fmt.Errorf("One of deploy client count and minimum active instances must be specified")
	}
	return nil
}

func (s *AutoProvInfo) Validate(fields map[string]struct{}) error {
	return nil
}

func (s *PrivacyPolicy) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
		return err
	}
	for _, o := range s.OutboundSecurityRules {
		if o.Protocol != "tcp" && o.Protocol != "udp" && o.Protocol != "icmp" {
			return fmt.Errorf("Protocol must be one of: (tcp,udp,icmp)")
		}
		if o.Protocol == "icmp" {
			if o.PortRangeMin != 0 || o.PortRangeMax != 0 {
				return fmt.Errorf("Port range must be empty for icmp")
			}
		} else {
			if o.PortRangeMin < minPort || o.PortRangeMin > maxPort {
				return fmt.Errorf("Invalid min port range: %d", o.PortRangeMin)
			}
			if o.PortRangeMax > maxPort {
				return fmt.Errorf("Invalid max port range: %d", o.PortRangeMax)
			}
			if o.PortRangeMin > o.PortRangeMax {
				return fmt.Errorf("Min port range: %d cannot be higher than max: %d", o.PortRangeMin, o.PortRangeMax)
			}
		}
		_, _, err := net.ParseCIDR(o.RemoteCidr)
		if err != nil {
			return err
		}
	}
	return nil
}

// Always valid
func (s *DeviceReport) Validate(fields map[string]struct{}) error {
	return nil
}

func (key *DeviceKey) ValidateKey() error {
	if key.UniqueId == "" || key.UniqueIdType == "" {
		return errors.New("Device id cannot be empty")
	}
	return nil
}
func (s *Device) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
		return err
	}
	// TODO - we might want to validate timestamp in the future
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

func (m *Metric) AddBoolVal(name string, bval bool) {
	val := MetricVal{Name: name}
	val.Value = &MetricVal_Bval{Bval: bval}
	m.Vals = append(m.Vals, &val)
}

func (m *Metric) AddStringVal(name string, sval string) {
	val := MetricVal{Name: name}
	val.Value = &MetricVal_Sval{Sval: sval}
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
	if ports == "" {
		return appports, nil
	}

	portSpecs, err := util.ParsePorts(ports)

	if err != nil {
		return nil, err
	}

	var proto dme.LProto
	var baseport int64
	var endport int64

	for _, portSpec := range portSpecs {
		proto, err = GetLProto(portSpec.Proto)
		if err != nil {
			return nil, err
		}
		baseport, err = strconv.ParseInt(portSpec.Port, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("unable to convert port range base value")
		}
		endport, err = strconv.ParseInt(portSpec.EndPort, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("unable to convert port range end value")
		}

		p := dme.AppPort{
			Proto:        proto,
			InternalPort: int32(baseport),
			EndPort:      int32(endport),
			Tls:          portSpec.Tls,
		}

		appports = append(appports, p)
	}
	return appports, nil
}

func CmpSortDebugReply(a DebugReply, b DebugReply) bool {
	// e2e tests ignore Name for comparison, so name cannot
	// be used to sort.
	aKey := a.Node
	aKey.Name = ""
	bKey := b.Node
	bKey.Name = ""
	return aKey.GetKeyString() < bKey.GetKeyString()
}

func IgnoreTaggedFields(taglist string) []cmp.Option {
	opts := []cmp.Option{}
	opts = append(opts, IgnoreAppFields(taglist))
	opts = append(opts, IgnoreAppInstFields(taglist))
	opts = append(opts, IgnoreAppInstInfoFields(taglist))
	opts = append(opts, IgnoreClusterInstFields(taglist))
	opts = append(opts, IgnoreClusterInstInfoFields(taglist))
	opts = append(opts, IgnoreCloudletFields(taglist))
	opts = append(opts, IgnoreCloudletInfoFields(taglist))
	opts = append(opts, IgnoreNodeFields(taglist))
	return opts
}

func CmpSortSlices() []cmp.Option {
	opts := []cmp.Option{}
	opts = append(opts, cmpopts.SortSlices(CmpSortApp))
	opts = append(opts, cmpopts.SortSlices(CmpSortAppInst))
	opts = append(opts, cmpopts.SortSlices(CmpSortCloudlet))
	opts = append(opts, cmpopts.SortSlices(CmpSortOperatorCode))
	opts = append(opts, cmpopts.SortSlices(CmpSortClusterInst))
	opts = append(opts, cmpopts.SortSlices(CmpSortFlavor))
	opts = append(opts, cmpopts.SortSlices(CmpSortCloudletInfo))
	opts = append(opts, cmpopts.SortSlices(CmpSortAppInstInfo))
	opts = append(opts, cmpopts.SortSlices(CmpSortClusterInstInfo))
	opts = append(opts, cmpopts.SortSlices(CmpSortNode))
	opts = append(opts, cmpopts.SortSlices(CmpSortCloudletPool))
	opts = append(opts, cmpopts.SortSlices(CmpSortCloudletPoolMember))
	opts = append(opts, cmpopts.SortSlices(CmpSortAutoScalePolicy))
	opts = append(opts, cmpopts.SortSlices(CmpSortResTagTable))
	opts = append(opts, cmpopts.SortSlices(CmpSortAppInstRefs))
	return opts
}

func GetOrg(obj interface{}) string {
	switch v := obj.(type) {
	case *OperatorCode:
		return v.Organization
	case *Cloudlet:
		return v.Key.Organization
	case *ClusterInst:
		return v.Key.Organization
	case *App:
		return v.Key.Organization
	case *AppInst:
		return v.Key.AppKey.Organization
	default:
		return "mobiledgex"
	}
}

func (c *ClusterInstCache) UsesOrg(org string) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, cd := range c.Objs {
		val := cd.Obj
		if val.Key.Organization == org || val.Key.CloudletKey.Organization == org || (val.Reservable && val.ReservedBy == org) {
			return true
		}
	}
	return false
}

func (c *CloudletInfoCache) WaitForState(ctx context.Context, key *CloudletKey, targetState CloudletState, timeout time.Duration) error {
	curState := CloudletState_CLOUDLET_STATE_UNKNOWN
	done := make(chan bool, 1)

	checkState := func(key *CloudletKey) {
		info := CloudletInfo{}
		if c.Get(key, &info) {
			curState = info.State
		}
		if curState == targetState {
			done <- true
		}
	}

	cancel := c.WatchKey(key, func(ctx context.Context) {
		checkState(key)
	})
	defer cancel()

	// After setting up watch, check current state,
	// as it may have already changed to target state.
	checkState(key)

	select {
	case <-done:
	case <-time.After(timeout):
		return fmt.Errorf("Timed out; expected state %s buf is %s",
			CloudletState_CamelName[int32(targetState)],
			CloudletState_CamelName[int32(curState)])
	}
	return nil
}

func (s *App) GetAutoProvPolicies() map[string]struct{} {
	policies := make(map[string]struct{})
	if s.AutoProvPolicy != "" {
		policies[s.AutoProvPolicy] = struct{}{}
	}
	for _, name := range s.AutoProvPolicies {
		policies[name] = struct{}{}
	}
	return policies
}
