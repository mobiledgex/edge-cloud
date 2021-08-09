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
	"github.com/mobiledgex/edge-cloud/objstore"
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

	GPUDriverLicenseConfig = "license.conf"
)

var ValidConfigKinds = map[string]struct{}{
	AppConfigHelmYaml:      struct{}{},
	AppAccessCustomization: struct{}{},
	AppConfigEnvYaml:       struct{}{},
}

var ReservedPlatformPorts = map[string]string{
	"tcp:22":    "Platform inter-node SSH",
	"tcp:20800": "Kubernetes master join server",
}

type WaitStateSpec struct {
	StreamCache *StreamObjCache
	StreamKey   *AppInstKey
}

type WaitStateOps func(wSpec *WaitStateSpec) error

func WithStreamObj(streamCache *StreamObjCache, streamKey *AppInstKey) WaitStateOps {
	return func(wSpec *WaitStateSpec) error {
		wSpec.StreamCache = streamCache
		wSpec.StreamKey = streamKey
		return nil
	}
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
	for i := range a.CloudletInfos {
		sort.Slice(a.CloudletInfos[i].ResourcesSnapshot.ClusterInsts[:], func(ii, jj int) bool {
			return a.CloudletInfos[i].ResourcesSnapshot.ClusterInsts[ii].GetKeyString() < a.CloudletInfos[i].ResourcesSnapshot.ClusterInsts[jj].GetKeyString()
		})
		sort.Slice(a.CloudletInfos[i].ResourcesSnapshot.VmAppInsts[:], func(ii, jj int) bool {
			return a.CloudletInfos[i].ResourcesSnapshot.VmAppInsts[ii].GetKeyString() < a.CloudletInfos[i].ResourcesSnapshot.VmAppInsts[jj].GetKeyString()
		})
	}
	sort.Slice(a.CloudletPools[:], func(i, j int) bool {
		return a.CloudletPools[i].Key.GetKeyString() < a.CloudletPools[j].Key.GetKeyString()
	})
	sort.Slice(a.AutoScalePolicies[:], func(i, j int) bool {
		return a.AutoScalePolicies[i].Key.GetKeyString() < a.AutoScalePolicies[j].Key.GetKeyString()
	})
	sort.Slice(a.AutoProvPolicies[:], func(i, j int) bool {
		return a.AutoProvPolicies[i].Key.GetKeyString() < a.AutoProvPolicies[j].Key.GetKeyString()
	})
	sort.Slice(a.TrustPolicies[:], func(i, j int) bool {
		return a.TrustPolicies[i].Key.GetKeyString() < a.TrustPolicies[j].Key.GetKeyString()
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
	sort.Slice(a.VmPools[:], func(i, j int) bool {
		return a.VmPools[i].Key.GetKeyString() < a.VmPools[j].Key.GetKeyString()
	})
	sort.Slice(a.FlowRateLimitSettings[:], func(i, j int) bool {
		return a.FlowRateLimitSettings[i].Key.GetKeyString() < a.FlowRateLimitSettings[j].Key.GetKeyString()
	})
	sort.Slice(a.MaxReqsRateLimitSettings[:], func(i, j int) bool {
		return a.MaxReqsRateLimitSettings[i].Key.GetKeyString() < a.MaxReqsRateLimitSettings[j].Key.GetKeyString()
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

func (key *VirtualClusterInstKey) ValidateKey() error {
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

func (key *GPUDriverKey) ValidateKey() error {
	if key.Organization != "" && !util.ValidName(key.Organization) {
		return errors.New("Invalid organization name")
	}
	if key.Name == "" {
		return errors.New("Missing gpu driver name")
	}
	if !util.ValidName(key.Name) {
		return errors.New("Invalid gpu driver name")
	}
	return nil
}

func (g *GPUDriverBuild) ValidateName() error {
	if g.Name == "" {
		return errors.New("Missing gpu driver build name")
	}
	if g.Name == GPUDriverLicenseConfig {
		return fmt.Errorf("%s is a reserved name and hence cannot be used as a build name", g.Name)
	}
	if !util.ValidName(g.Name) {
		return fmt.Errorf("Invalid gpu driver build name: %s", g.Name)
	}
	return nil
}

func (g *GPUDriverBuild) Validate() error {
	if err := g.ValidateName(); err != nil {
		return err
	}
	if g.DriverPath == "" {
		return fmt.Errorf("Missing driverpath")
	}
	if g.Md5Sum == "" {
		return fmt.Errorf("Missing md5sum")
	}
	if _, err := util.ImagePathParse(g.DriverPath); err != nil {
		return fmt.Errorf("Invalid driver path(%q): %v", g.DriverPath, err)
	}
	if g.DriverPathCreds != "" {
		out := strings.Split(g.DriverPathCreds, ":")
		if len(out) != 2 {
			return fmt.Errorf("Invalid GPU driver build path credentials, should be in format 'username:password'")
		}
	}
	if g.OperatingSystem == OSType_LINUX && g.KernelVersion == "" {
		return fmt.Errorf("Kernel version is required for Linux build")
	}
	if err := g.ValidateEnums(); err != nil {
		return err
	}
	return nil
}

func (g *GPUDriverBuildMember) Validate() error {
	if err := g.GetKey().ValidateKey(); err != nil {
		return err
	}
	if err := g.Build.Validate(); err != nil {
		return err
	}
	return nil
}

func (g *GPUDriver) Validate(fields map[string]struct{}) error {
	if err := g.GetKey().ValidateKey(); err != nil {
		return err
	}
	if err := g.ValidateEnums(); err != nil {
		return err
	}
	buildNames := make(map[string]struct{})
	for _, build := range g.Builds {
		if err := build.Validate(); err != nil {
			return err
		}
		if _, ok := buildNames[build.Name]; ok {
			return fmt.Errorf("GPU driver build with name %s already exists", build.Name)
		}
		buildNames[build.Name] = struct{}{}
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
		if s.MaintenanceState != dme.MaintenanceState_NORMAL_OPERATION && s.MaintenanceState != dme.MaintenanceState_MAINTENANCE_START && s.MaintenanceState != dme.MaintenanceState_MAINTENANCE_START_NO_FAILOVER {
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

	if _, found := fields[CloudletFieldDefaultResourceAlertThreshold]; found {
		if s.DefaultResourceAlertThreshold < 0 || s.DefaultResourceAlertThreshold > 100 {
			return fmt.Errorf("Invalid resource alert threshold %d specified, valid threshold is in the range of 0 to 100", s.DefaultResourceAlertThreshold)

		}
	}

	for _, resQuota := range s.ResourceQuotas {
		if resQuota.AlertThreshold < 0 || resQuota.AlertThreshold > 100 {
			return fmt.Errorf("Invalid resource quota alert threshold %d specified for %s, valid threshold is in the range of 0 to 100", resQuota.AlertThreshold, resQuota.Name)

		}
	}

	return nil
}

func (s *CloudletInfo) Validate(fields map[string]struct{}) error {
	return nil
}

func (s *CloudletInternal) Validate(fields map[string]struct{}) error {
	return nil
}

func (key *CloudletPoolKey) ValidateKey() error {
	if !util.ValidName(key.Organization) {
		return errors.New("Invalid organization name")
	}
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

func (s *VM) ValidateName() error {
	if s.Name == "" {
		return errors.New("Missing VM name")
	}
	if !util.ValidName(s.Name) {
		return errors.New("Invalid VM name")
	}
	return nil
}

var invalidVMPoolIPs = map[string]struct{}{
	"0.0.0.0":   struct{}{},
	"127.0.0.1": struct{}{},
}

func (s *VM) Validate() error {
	if err := s.ValidateName(); err != nil {
		return err
	}
	if s.NetInfo.ExternalIp != "" {
		if net.ParseIP(s.NetInfo.ExternalIp) == nil {
			return fmt.Errorf("Invalid Address: %s", s.NetInfo.ExternalIp)
		}
		if _, ok := invalidVMPoolIPs[s.NetInfo.ExternalIp]; ok {
			return fmt.Errorf("Invalid Address: %s", s.NetInfo.ExternalIp)
		}
	}
	if s.NetInfo.InternalIp == "" {
		return fmt.Errorf("Missing internal IP for VM: %s", s.Name)
	}
	if net.ParseIP(s.NetInfo.InternalIp) == nil {
		return fmt.Errorf("Invalid Address: %s", s.NetInfo.InternalIp)
	}
	if _, ok := invalidVMPoolIPs[s.NetInfo.InternalIp]; ok {
		return fmt.Errorf("Invalid Address: %s", s.NetInfo.ExternalIp)
	}
	return nil
}

func (key *VMPoolKey) ValidateKey() error {
	if !util.ValidName(key.Organization) {
		return errors.New("Invalid organization name")
	}
	if !util.ValidName(key.Name) {
		return errors.New("Invalid VM pool name")
	}
	return nil
}

func (s *VMPool) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
		return err
	}
	if err := s.ValidateEnums(); err != nil {
		return err
	}
	externalIPMap := make(map[string]struct{})
	internalIPMap := make(map[string]struct{})
	for _, v := range s.Vms {
		if err := v.Validate(); err != nil {
			return err
		}
		if v.NetInfo.ExternalIp != "" {
			if _, ok := externalIPMap[v.NetInfo.ExternalIp]; ok {
				return fmt.Errorf("VM with same external IP %s already exists", v.NetInfo.ExternalIp)
			}
			externalIPMap[v.NetInfo.ExternalIp] = struct{}{}
		}
		if v.NetInfo.InternalIp != "" {
			if _, ok := internalIPMap[v.NetInfo.InternalIp]; ok {
				return fmt.Errorf("VM with same internal IP %s already exists", v.NetInfo.InternalIp)
			}
			internalIPMap[v.NetInfo.InternalIp] = struct{}{}
		}
		if v.State != VMState_VM_FREE && v.State != VMState_VM_FORCE_FREE {
			return errors.New("Invalid VM state, only VmForceFree state is allowed")
		}
	}
	return nil
}

func (s *VMPoolMember) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
		return err
	}
	if err := s.Vm.Validate(); err != nil {
		return err
	}
	return nil
}

func (s *VMPoolInfo) Validate(fields map[string]struct{}) error {
	return nil
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

func (s *AppInstClientKey) ValidateKey() error {
	if s.AppInstKey.Matches(&AppInstKey{}) && s.UniqueId == "" && s.UniqueIdType == "" {
		return fmt.Errorf("At least one of the key fields must be non-empty %v", s)
	}
	return nil
}

func (s *AppInstClientKey) Validate(fields map[string]struct{}) error {
	return s.ValidateKey()
}

func (s *AutoScalePolicy) HasV0Config() bool {
	if s.ScaleUpCpuThresh > 0 || s.ScaleDownCpuThresh > 0 {
		return true
	}
	return false
}

func (s *AutoScalePolicy) HasV1Config() bool {
	if s.TargetCpu > 0 || s.TargetMem > 0 || s.TargetActiveConnections > 0 {
		return true
	}
	return false
}

const DefaultStabilizationWindowSec = 300

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
	if s.HasV0Config() && s.HasV1Config() {
		return errors.New("The new target cpu/mem/active-connections can only be used once the old cpu threshold settings have been disabled (set to 0)")
	}
	if s.HasV0Config() {
		if s.ScaleUpCpuThresh < 0 || s.ScaleUpCpuThresh > 100 {
			return errors.New("Scale up CPU threshold must be between 0 and 100")
		}
		if s.ScaleDownCpuThresh < 0 || s.ScaleDownCpuThresh > 100 {
			return errors.New("Scale down CPU threshold must be between 0 and 100")
		}
		if s.ScaleUpCpuThresh <= s.ScaleDownCpuThresh {
			return fmt.Errorf("Scale down cpu threshold must be less than scale up cpu threshold")
		}
	} else if !s.HasV1Config() {
		return fmt.Errorf("One of target cpu or target mem or target active connections must be specified")
	} else {
		// v1 config
		if s.StabilizationWindowSec == 0 {
			s.StabilizationWindowSec = DefaultStabilizationWindowSec
		}
		if s.TargetCpu < 0 || s.TargetCpu > 100 {
			return fmt.Errorf("Target cpu must be between 0 (disabled) and 100")
		}
		if s.TargetMem < 0 || s.TargetMem > 100 {
			return fmt.Errorf("Target mem must be between 0 (disabled) and 100")
		}
		maxActiveConnections := uint64(1e12)
		if s.TargetActiveConnections < 0 || s.TargetActiveConnections > maxActiveConnections {
			return fmt.Errorf("Target active connections must be between 0 (disabled) and %d", maxActiveConnections)
		}
	}
	if s.MaxNodes <= s.MinNodes {
		return fmt.Errorf("Max nodes must be greater than Min nodes")
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

func (s *TrustPolicy) Validate(fields map[string]struct{}) error {
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
	}
	return 0, fmt.Errorf("Unsupported protocol: %s", s)
}

func LProtoStr(proto dme.LProto) (string, error) {
	switch proto {
	case dme.LProto_L_PROTO_TCP:
		return "tcp", nil
	case dme.LProto_L_PROTO_UDP:
		return "udp", nil
	}
	return "", fmt.Errorf("Invalid proto %d", proto)
}

func L4ProtoStr(proto dme.LProto) (string, error) {
	switch proto {
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

		// loop through to verify we are not using a platform reserved port
		lastPort := endport
		if lastPort == 0 {
			lastPort = baseport
		}
		for pnum := baseport; pnum <= lastPort; pnum++ {
			pstring := fmt.Sprintf("%s:%d", strings.ToLower(portSpec.Proto), pnum)
			desc, reserved := ReservedPlatformPorts[pstring]
			if reserved {
				return nil, fmt.Errorf("App cannot use port %s - reserved for %s", pstring, desc)
			}
		}

		p := dme.AppPort{
			Proto:        proto,
			InternalPort: int32(baseport),
			EndPort:      int32(endport),
			Tls:          portSpec.Tls,
			Nginx:        portSpec.Nginx,
			MaxPktSize:   portSpec.MaxPktSize,
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

func CmpSortFlavorInfo(a *FlavorInfo, b *FlavorInfo) bool {
	return a.Name < b.Name
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
	opts = append(opts, cmpopts.SortSlices(CmpSortFlavorInfo))
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

func GetTags(obj interface{}) map[string]string {
	switch v := obj.(type) {
	case objstore.Obj:
		return v.GetObjKey().GetTags()
	case objstore.ObjKey:
		return v.GetTags()
	default:
		return map[string]string{}
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

func (c *CloudletInfoCache) WaitForState(ctx context.Context, key *CloudletKey, targetState dme.CloudletState, timeout time.Duration) error {
	curState := dme.CloudletState_CLOUDLET_STATE_UNKNOWN
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
			dme.CloudletState_CamelName[int32(targetState)],
			dme.CloudletState_CamelName[int32(curState)])
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

func (s *App) GetAutoProvPolicys() map[PolicyKey]struct{} {
	policies := make(map[PolicyKey]struct{})
	if s.AutoProvPolicy != "" {
		key := PolicyKey{
			Name:         s.AutoProvPolicy,
			Organization: s.Key.Organization,
		}
		policies[key] = struct{}{}
	}
	for _, name := range s.AutoProvPolicies {
		key := PolicyKey{
			Name:         name,
			Organization: s.Key.Organization,
		}
		policies[key] = struct{}{}
	}
	return policies
}

func (s *AutoProvPolicy) GetCloudletKeys() map[CloudletKey]struct{} {
	keys := make(map[CloudletKey]struct{})
	for _, cl := range s.Cloudlets {
		keys[cl.Key] = struct{}{}
	}
	return keys
}

func (s *CloudletPool) GetCloudletKeys() map[CloudletKey]struct{} {
	keys := make(map[CloudletKey]struct{})
	for _, name := range s.Cloudlets {
		key := CloudletKey{
			Organization: s.Key.Organization,
			Name:         name,
		}
		keys[key] = struct{}{}
	}
	return keys
}

func (s *AppInst) GetRealClusterName() string {
	if s.RealClusterName != "" {
		return s.RealClusterName
	}
	return s.Key.ClusterInstKey.ClusterKey.Name
}

// For vanity naming of reservable ClusterInsts, the actual ClusterInst
// name may be on the AppInst object.
func (s *AppInst) ClusterInstKey() *ClusterInstKey {
	return s.Key.ClusterInstKey.Real(s.GetRealClusterName())
}

// Convert VirtualClusterInstKey to a real ClusterInstKey,
// given the real cluster name (may be blank for no aliasing).
func (s *VirtualClusterInstKey) Real(realClusterName string) *ClusterInstKey {
	key := ClusterInstKey{
		ClusterKey:   s.ClusterKey,
		CloudletKey:  s.CloudletKey,
		Organization: s.Organization,
	}
	if realClusterName != "" {
		key.ClusterKey.Name = realClusterName
	}
	return &key
}

// Convert real ClusterInstKey to a VirtualClusterInstKey,
// give the virtual cluster name (may be blank for no aliasing).
func (s *ClusterInstKey) Virtual(virtualName string) *VirtualClusterInstKey {
	key := VirtualClusterInstKey{
		ClusterKey:   s.ClusterKey,
		CloudletKey:  s.CloudletKey,
		Organization: s.Organization,
	}
	if virtualName != "" {
		key.ClusterKey.Name = virtualName
	}
	return &key
}

func (s *ClusterInstRefKey) FromClusterInstKey(key *ClusterInstKey) {
	s.ClusterKey = key.ClusterKey
	s.Organization = key.Organization
}

func (s *ClusterInstKey) FromClusterInstRefKey(key *ClusterInstRefKey, clKey *CloudletKey) {
	s.ClusterKey = key.ClusterKey
	s.Organization = key.Organization
	s.CloudletKey = *clKey
}

func (s *AppInstRefKey) FromAppInstKey(key *AppInstKey) {
	s.AppKey = key.AppKey
	s.ClusterInstKey.ClusterKey = key.ClusterInstKey.ClusterKey
	s.ClusterInstKey.Organization = key.ClusterInstKey.Organization
}

func (s *AppInstKey) FromAppInstRefKey(key *AppInstRefKey, clKey *CloudletKey) {
	s.AppKey = key.AppKey
	s.ClusterInstKey.ClusterKey = key.ClusterInstKey.ClusterKey
	s.ClusterInstKey.Organization = key.ClusterInstKey.Organization
	s.ClusterInstKey.CloudletKey = *clKey
}

func (s *StreamObj) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().ValidateKey(); err != nil {
		return err
	}
	return nil
}

func GetStreamKeyFromClusterInstKey(key *VirtualClusterInstKey) AppInstKey {
	return AppInstKey{
		ClusterInstKey: *key,
	}
}

func GetStreamKeyFromCloudletKey(key *CloudletKey) AppInstKey {
	return AppInstKey{
		ClusterInstKey: VirtualClusterInstKey{
			CloudletKey: *key,
		},
	}
}

// Temporary way to get unique stream key for GPU driver object
// This will be fixed as part of 3rd-party in-memory DB changes
func GetStreamKeyFromGPUDriverKey(key *GPUDriverKey) AppInstKey {
	return AppInstKey{
		ClusterInstKey: VirtualClusterInstKey{
			CloudletKey: CloudletKey{
				Name: key.Name + "_" + key.Organization,
			},
		},
	}
}

func (r *InfraResources) UpdateResources(inRes *InfraResources) (updated bool) {
	if inRes == nil || len(inRes.Vms) == 0 {
		return false
	}
	if len(r.Vms) != len(inRes.Vms) {
		return true
	}
	vmStatusMap := make(map[string]string)
	for _, vmInfo := range r.Vms {
		vmStatusMap[vmInfo.Name] = vmInfo.Status
	}
	for _, vmInfo := range inRes.Vms {
		status, ok := vmStatusMap[vmInfo.Name]
		if !ok {
			return true
		}
		if status != vmInfo.Status {
			return true
		}
	}
	return false
}

func (key *AlertPolicyKey) ValidateKey() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid alert policy name")
	}
	if !util.ValidName(key.Organization) {
		return errors.New("Invalid alert policy organization")
	}
	return nil
}

func (a *AlertPolicy) Validate(fields map[string]struct{}) error {
	if err := a.GetKey().ValidateKey(); err != nil {
		return err
	}
	// Since active connections and other metrics are part
	// of different instances of Prometheus, disallow mixing them
	if a.ActiveConnLimit != 0 {
		if a.CpuUtilizationLimit != 0 || a.MemUtilizationLimit != 0 || a.DiskUtilizationLimit != 0 {
			return errors.New("Active Connection Alerts should not include any other triggers")
		}
	}
	// at least one of the values for alert should be set
	if a.ActiveConnLimit == 0 && a.CpuUtilizationLimit == 0 &&
		a.MemUtilizationLimit == 0 && a.DiskUtilizationLimit == 0 {
		return errors.New("At least one of the measurements for alert should be set")
	}
	// check CPU to be within 0-100 percent
	if a.CpuUtilizationLimit > 100 {
		return errors.New("Cpu utilization limit is percent. Valid values 1-100%")
	}
	// check Memory to be within 0-100 percent
	if a.MemUtilizationLimit > 100 {
		return errors.New("Memory utilization limit is percent. Valid values 1-100%")
	}
	// check Disk to be within 0-100 percent
	if a.DiskUtilizationLimit > 100 {
		return errors.New("Disk utilization limit is percent. Valid values 1-100%")
	}
	return nil
}

// Check if AlertPolicies are different between two apps
func (app *App) AppAlertPoliciesDifferent(other *App) bool {
	alertsDiff := false
	if len(app.AlertPolicies) != len(other.AlertPolicies) {
		alertsDiff = true
	} else {
		oldAlerts := make(map[string]struct{})
		for _, alert := range app.AlertPolicies {
			oldAlerts[alert] = struct{}{}
		}
		for _, alert := range other.AlertPolicies {
			if _, found := oldAlerts[alert]; !found {
				alertsDiff = true
				break
			}
		}
	}
	return alertsDiff
}
