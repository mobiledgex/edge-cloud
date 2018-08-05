package crmutil

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"

	"github.com/mobiledgex/edge-cloud-infra/openstack-prov/oscliapi"
	"github.com/mobiledgex/edge-cloud-infra/openstack-tenant/agent/cloudflare"
	//"github.com/mobiledgex/edge-cloud/edgeproto"
	valid "github.com/asaskevich/govalidator"
	"github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/log"

	"github.com/nanobox-io/golang-ssh"
	"github.com/parnurzeal/gorequest"
	//"github.com/fsouza/go-dockerclient"
)

var mexEnv = map[string]string{
	"MEX_DEBUG":           os.Getenv("MEX_DEBUG"),
	"MEX_DIR":             os.Getenv("HOME") + "/.mobiledgex",
	"MEX_CF_KEY":          os.Getenv("MEX_CF_KEY"),
	"MEX_CF_USER":         os.Getenv("MEX_CF_USER"),
	"MEX_DOCKER_REG_PASS": os.Getenv("MEX_DOCKER_REG_PASS"),
	"MEX_AGENT_PORT":      "18889",
	"MEX_SSH_KEY":         "id_rsa_mobiledgex",
	"MEX_K8S_MASTER":      "mex-k8s-master",
	"MEX_K8S_NODE":        "mex-k8s-node",
	"MEX_DOCKER_REGISTRY": "registry.mobiledgex.net:5000",
}

//MEXInit initializes MEX API
func MEXInit() {
	for _, evname := range reflect.ValueOf(mexEnv).MapKeys() {
		evstr := evname.String()
		evval := os.Getenv(evstr)
		if evval != "" {
			mexEnv[evstr] = evval
		}
	}
	log.DebugLog(log.DebugLevelMexos, "mex environment", "mexEnv", mexEnv)
}

//MEXCheckEnvVars sets up environment vars and checks for credentials required for running
func MEXCheckEnvVars(mf *Manifest) error {
	// secrets to be passed via Env var still : MEX_CF_KEY, MEX_CF_USER, MEX_DOCKER_REG_PASS
	// TODO: use `secrets` or `vault`
	for _, evar := range []string{"MEX_CF_KEY", "MEX_CF_USER", "MEX_DOCKER_REG_PASS", "MEX_DIR"} {
		if _, ok := mexEnv[evar]; !ok {
			return fmt.Errorf("missing env var %s", evar)
		}
	}
	//TODO need to allow users to save the environment under platform name inside .mobiledgex
	return nil
}

//MEXRootLB has rootLB data
type MEXRootLB struct {
	Name     string
	PlatConf *Manifest
}

//MEXRootLBMap maps name of rootLB to rootLB instance
var MEXRootLBMap = make(map[string]*MEXRootLB)

//NewRootLB gets a new rootLB instance
func NewRootLB(rootLBName string) (*MEXRootLB, error) {
	log.DebugLog(log.DebugLevelMexos, "new rootLB", "rootLBName", rootLBName)
	curlb, ok := MEXRootLBMap[rootLBName]
	if ok {
		return nil, fmt.Errorf("rootlb exists %v", curlb)
	}
	newRootLB := &MEXRootLB{Name: rootLBName}
	MEXRootLBMap[rootLBName] = newRootLB
	return newRootLB, nil
}

//NewRootLBManifest creates rootLB instance and sets Platform Config with manifest
func NewRootLBManifest(mf *Manifest) (*MEXRootLB, error) {
	log.DebugLog(log.DebugLevelMexos, "new rootLB with manifest", "mf", mf)
	rootLB, err := NewRootLB(mf.Spec.RootLB)
	if err != nil {
		return nil, err
	}
	setPlatConf(rootLB, mf)
	return rootLB, nil
}

//DeleteRootLB to be called by code that called NewRootLB
func DeleteRootLB(rootLBName string) {
	delete(MEXRootLBMap, rootLBName) //no mutex because caller should be serializing New/Delete in a control loop
}

//TODO use versioning on docker image for agent

//XXX ClusterInst seems to have Nodes which is a number.
//   The Nodes should be part of the Cluster flavor.  And there should be Max nodes, and current num of nodes.
//   Because the whole point of k8s and similar other clusters is the ability to expand.
//   Cluster flavor defines what kind of cluster we have available for use.
//   A medium cluster flavor may say "I have three nodes, ..."
//   Can the Node or Flavor change for the ClusterInst?
//   What needs to be done when contents change.
//   The old and new values are supposedly to be passed in the future when Cache is updated.
//   We have to compare old values and new values and figure out what changed.
//   Then act on the changes noticed.
//   There is no indication of type of cluster being created.  So assume k8s.
//   Nor is there any tenant information, so no ability to isolate, identify, account for usage or quota.
//   And no network type information or storage type information.
//   So if an app needs an external IP, we can't figure out if that is the case.
//   Nor is there a way to return the IP address or DNS name. Or even know if it needs a DNS name.
//   No ability to open ports, redirect or set up any kind of reverse proxy control.  etc.

//ClusterFlavor contains definitions of cluster flavor
type ClusterFlavor struct {
	Type           string
	Name           string
	Status         string
	NumNodes       int
	MaxNodes       int
	NumMasterNodes int
	NetworkSpec    string
	StorageSpec    string
	NodeFlavor     ClusterNodeFlavor
	MasterFlavor   ClusterMasterFlavor
	Topology       string
}

//NetworkSpec examples:
// TYPE,NAME,CIDR,OPTIONS,EXTRAS
// "priv-subnet,mex-k8s-net-1,10.201.X.0/24,rp-dns-name"
// "external-ip,external-network-shared,1.2.3.4/8,dhcp"
// "external-ip,external-network-shared,1.2.3.4/8"
// "external-dns,external-network-shared,1.2.3.4/8,dns-name"
// "net-custom-type,some-name,8.8.244.33/16,auto-1"

//StorageSpec examples:
// TYPE,NAME,PARAM,OPTIONS,EXTRAS
//  ceph,internal-ceph-cluster,param1:param2:param3,opt1:opt2,extra1:extra2
//  nfs,nfsv4-internal,param1,opt1,extra1
//  gluster,glusterv3-ext,param1,opt1,extra1
//  postgres-cluster,post-v3,param1,opt1,extra1

//ClusterNodeFlavor contains details of flavor for the node
type ClusterNodeFlavor struct {
	Type string
	Name string
}

//ClusterMasterFlavor contains details of flavor for the master node
type ClusterMasterFlavor struct {
	Type string
	Name string
}

//ValidClusterFlavors lists all valid flavor names
var ValidClusterFlavors = []string{
	"x1.tiny", "x1.medium", "x1.small", "x1.large", "x1.xlarge", "x1.xxlarge",
}

const (
	activeStatus    = "active"
	availableStatus = "available"
	activeService   = "ACTIVE"
)

//AvailableClusterFlavors lists currently available flavors
var AvailableClusterFlavors = []*ClusterFlavor{
	&ClusterFlavor{
		Name:           "x1.medium",
		Type:           "k8s",
		Status:         activeStatus,
		NumNodes:       2,
		NumMasterNodes: 1,
		Topology:       "type-1",
		NetworkSpec:    "priv-subnet,mex-k8s-net-1,10.101.X.0/24",
		StorageSpec:    "default",
		NodeFlavor:     ClusterNodeFlavor{Name: "k8s-medium", Type: "k8s-node"},
		MasterFlavor:   ClusterMasterFlavor{Name: "k8s-medium", Type: "k8s-master"},
	},
}

//IsValidMEXOSEnv caches the validity of the env
var IsValidMEXOSEnv = false

//ValidateMEXOSEnv makes sure the environment is valid for mexos
func ValidateMEXOSEnv(osEnvValid bool) bool {
	if !osEnvValid {
		log.DebugLog(log.DebugLevelMexos, "invalid mex env")
		IsValidMEXOSEnv = false
		return false
	}
	//XXX do more validation on internal env vars
	IsValidMEXOSEnv = true
	log.DebugLog(log.DebugLevelMexos, "valid mex env")
	return IsValidMEXOSEnv
}

//AddFlavor adds a new flavor to be kept track of
func AddFlavor(flavor string) error {
	log.DebugLog(log.DebugLevelMexos, "add flavor", "flavor", flavor)
	if err := ValidateFlavor(flavor); err != nil {
		return fmt.Errorf("invalid flavor")
	}
	for _, f := range AvailableClusterFlavors {
		if flavor == f.Name {
			if f.Status == activeStatus {
				return nil // fmt.Errorf("exists already")
			}
			if f.Status == availableStatus {
				f.Status = activeStatus
				return nil
			}
		}
	}
	nf := ClusterFlavor{Name: flavor}
	nf.Status = activeStatus
	AvailableClusterFlavors = append(AvailableClusterFlavors, &nf)
	//XXX need local database to store this persistently since controller won't
	return nil
}

//GetClusterFlavor returns the flavor of the cluster
func GetClusterFlavor(flavor string) (*ClusterFlavor, error) {
	log.DebugLog(log.DebugLevelMexos, "get cluster flavor", "flavor", flavor)
	for _, f := range AvailableClusterFlavors {
		if flavor == f.Name {
			if f.Status == activeStatus {
				return f, nil
			}
			return nil, fmt.Errorf("flavor exists but status not active")
		}
	}
	return nil, fmt.Errorf("flavor does not exist %s", flavor)
}

//mexCreateClusterKubernetes creates a cluster of nodes. It can take a while, so call from a goroutine.
func mexCreateClusterKubernetes(mf *Manifest) error {
	//func mexCreateClusterKubernetes(mf *Manifest) (*string, error) {
	log.DebugLog(log.DebugLevelMexos, "create kubernetes cluster", "mf", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if mf.Spec.Flavor == "" {
		return fmt.Errorf("empty flavor")
	}
	if err = ValidateFlavor(mf.Spec.Flavor); err != nil {
		return fmt.Errorf("invalid flavor")
	}
	cf, err := GetClusterFlavor(mf.Spec.Flavor)
	if err != nil {
		return err
	}
	if cf.NumNodes < 1 {
		return fmt.Errorf("invalid flavor profile, %v", cf)
	}
	//TODO more than one networks
	if mf.Spec.NetworkScheme == "" {
		return fmt.Errorf("empty network spec")
	}
	if !strings.HasPrefix(mf.Spec.NetworkScheme, "priv-subnet") {
		return fmt.Errorf("unsupported netSpec kind %s", mf.Spec.NetworkScheme)
		// XXX for now
	}
	//TODO allow more net types
	//TODO validate CIDR, etc.
	if mf.Metadata.Tags == "" {
		return fmt.Errorf("empty tag")
	}
	if mf.Metadata.Tenant == "" {
		return fmt.Errorf("empty tenant")
	}
	err = ValidateTenant(mf.Metadata.Tenant)
	if err != nil {
		return fmt.Errorf("can't validate tenant, %v", err)
	}
	err = ValidateTags(mf.Metadata.Tags)
	if err != nil {
		return fmt.Errorf("invalid tag, %v", err)
	}
	//TODO add whole manifest yaml->json into stringified property of the kvm instance for later
	//XXX should check for quota, permissions, access control, etc. here
	//guid := xid.New().String()
	//kvmname := fmt.Sprintf("%s-1-%s-%s", mexEnv["MEX_K8S_MASTER"], mf.Metadata.Name, guid)
	kvmname := fmt.Sprintf("%s-1-%s", mexEnv["MEX_K8S_MASTER"], mf.Metadata.Name)
	sd, err := oscli.GetServerDetails(kvmname)
	if err == nil {
		if sd.Name == kvmname {
			log.DebugLog(log.DebugLevelMexos, "k8s master exists", "kvmname", kvmname)
			return nil
		}
	}
	log.DebugLog(log.DebugLevelMexos, "proceed to create k8s master kvm", "kvmname", kvmname)
	err = oscli.CreateMEXKVM(kvmname,
		"k8s-master",
		mf.Spec.NetworkScheme,
		mf.Metadata.Tags,
		mf.Metadata.Tenant,
		1,
	)
	if err != nil {
		return fmt.Errorf("can't create k8s master, %v", err)
	}
	for i := 1; i <= cf.NumNodes; i++ {
		//construct node name
		//kvmnodename := fmt.Sprintf("%s-%d-%s-%s", mexEnv["MEX_K8S_NODE"], i, mf.Metadata.Name, guid)
		kvmnodename := fmt.Sprintf("%s-%d-%s", mexEnv["MEX_K8S_NODE"], i, mf.Metadata.Name)
		err = oscli.CreateMEXKVM(kvmnodename,
			"k8s-node",
			mf.Spec.NetworkScheme,
			mf.Metadata.Tags,
			mf.Metadata.Tenant,
			i)
		if err != nil {
			return fmt.Errorf("can't create k8s node, %v", err)
		}
	}
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("missing external network in platform config")
	}
	if err = LBAddRoute(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, kvmname); err != nil {
		return err
	}
	if err = oscli.SetServerProperty(kvmname, "mex-flavor="+mf.Spec.Flavor); err != nil {
		return err
	}
	ready := false
	for i := 0; i < 10; i++ {
		ready, err = IsClusterReady(mf, rootLB)
		if err != nil {
			return err
		}
		if ready {
			log.DebugLog(log.DebugLevelMexos, "kubernetes cluster ready")
			break
		}
		log.DebugLog(log.DebugLevelMexos, "wating for kubernetes cluster to be ready...")
		time.Sleep(30 * time.Second)
	}
	if !ready {
		return fmt.Errorf("cluster not ready (yet)")
	}
	//return &guid, nil
	return nil
}

//LBAddRoute adds a route to LB
func LBAddRoute(rootLBName, extNet, name string) error {
	if rootLBName == "" {
		return fmt.Errorf("empty rootLB")
	}
	if name == "" {
		return fmt.Errorf("empty name")
	}
	if extNet == "" {
		return fmt.Errorf("empty external network")
	}
	ap, err := LBGetRoute(rootLBName, name)
	if err != nil {
		return err
	}
	if len(ap) != 2 {
		return fmt.Errorf("expected 2 addresses, got %d", len(ap))
	}
	cmd := fmt.Sprintf("ip route add %s via %s dev ens3", ap[0], ap[1])
	client, err := GetSSHClient(rootLBName, extNet, "root")
	if err != nil {
		return err
	}
	out, err := client.Output(cmd)
	if err != nil {
		if strings.Contains(out, "RTNETLINK") && strings.Contains(out, " exists") {
			log.DebugLog(log.DebugLevelMexos, "warning, can't add existing route to rootLB", "cmd", cmd, "out", out, "error", err)
			return nil
		}
		return fmt.Errorf("can't add route to rootlb, %s, %s, %v", cmd, out, err)
	}
	return nil
}

//LBRemoveRoute removes route for LB
func LBRemoveRoute(rootLB, extNet, name string) error {
	ap, err := LBGetRoute(rootLB, name)
	if err != nil {
		return err
	}
	if len(ap) != 2 {
		return fmt.Errorf("expected 2 addresses, got %d", len(ap))
	}
	cmd := fmt.Sprintf("ip route delete %s via %s dev ens3", ap[0], ap[1])
	client, err := GetSSHClient(rootLB, extNet, "root")
	if err != nil {
		return err
	}
	out, err := client.Output(cmd)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "can't delete route at rootLB", "cmd", cmd, "out", out, "error", err)
		//not a fatal error
		return nil
	}
	return nil
}

//GetInternalIP returns IP of the server
func GetInternalIP(name string) (string, error) {
	sd, err := oscli.GetServerDetails(name)
	if err != nil {
		return "", err
	}
	its := strings.Split(sd.Addresses, "=")
	if len(its) != 2 {
		return "", fmt.Errorf("GetInternalIP: can't parse server detail addresses, %v, %v", sd, err)
	}
	return its[1], nil
}

//GetInternalCIDR returns CIDR of server
func GetInternalCIDR(name string) (string, error) {
	addr, err := GetInternalIP(name)
	if err != nil {
		return "", err
	}
	cidr := addr + "/24" // XXX we use this convention of /24 in k8s priv-net
	return cidr, nil
}

//LBGetRoute returns route of LB
func LBGetRoute(rootLB, name string) ([]string, error) {
	cidr, err := GetInternalCIDR(name)
	if err != nil {
		return nil, err
	}
	_, ipn, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("can't parse %s, %v", cidr, err)
	}
	v4 := ipn.IP.To4()
	dn := fmt.Sprintf("%d.%d.%d.0/24", v4[0], v4[1], v4[2])
	rn := oscli.GetMEXExternalRouter()
	rd, err := oscli.GetRouterDetail(rn)
	if err != nil {
		return nil, fmt.Errorf("can't get router detail for %s, %v", rn, err)
	}
	reg, err := oscli.GetRouterDetailExternalGateway(rd)
	if err != nil {
		return nil, fmt.Errorf("can't get router detail external gateway, %v", err)
	}
	if len(reg.ExternalFixedIPs) < 1 {
		return nil, fmt.Errorf("can't get external fixed ips list from router detail external gateway")
	}
	fip := reg.ExternalFixedIPs[0]
	return []string{dn, fip.IPAddress}, nil
}

//ValidateNetSpec parses and validates the netSpec
func ValidateNetSpec(netSpec string) error {
	if netSpec == "" {
		return fmt.Errorf("empty netspec")
	}
	return nil
}

//ValidateTags parses and validates tags
func ValidateTags(tags string) error {
	if tags == "" {
		return fmt.Errorf("empty tags")
	}
	return nil
}

//ValidateTenant parses and validates tenant
func ValidateTenant(tenant string) error {
	if tenant == "" {
		return fmt.Errorf("emtpy tenant")
	}
	return nil
}

//ValidateFlavor parses and validates flavor
func ValidateFlavor(flavor string) error {
	if flavor == "" {
		return fmt.Errorf("empty flavor")
	}
	for _, f := range ValidClusterFlavors {
		if flavor == f {
			return nil
		}
	}
	return fmt.Errorf("invalid flavor")
}

//IsFlavorSupported checks whether flavor is supported currently
func IsFlavorSupported(flavor string) bool {
	//XXX we only support x1.medium for now
	return flavor == "x1.medium"
}

func getRootLB(name string) (*MEXRootLB, error) {
	rootLB, ok := MEXRootLBMap[name]
	if !ok {
		return nil, fmt.Errorf("can't find rootlb %s", name)
	}
	return rootLB, nil
}

//mexDeleteClusterKubernetes deletes kubernetes cluster
func mexDeleteClusterKubernetes(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "deleting kubernetes cluster", "mf", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	name := mf.Metadata.Name
	srvs, err := oscli.ListServers()
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "looking for server", "name", name, "servers", srvs)
	force := strings.Contains(mf.Spec.Flags, "force")
	serverDeleted := false
	for _, s := range srvs {
		if strings.Contains(s.Name, name) {
			if strings.Contains(s.Name, mexEnv["MEX_K8S_MASTER"]) {
				err = LBRemoveRoute(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, s.Name)
				if err != nil {
					if !force {
						err = fmt.Errorf("failed to remove route for %s, %v", s.Name, err)
						log.DebugLog(log.DebugLevelMexos, "failed to remove route", "name", s.Name, "error", err)
						return err
					}
					log.DebugLog(log.DebugLevelMexos, "forced to continue, failed to remove route ", "name", s.Name, "error", err)
				}
			}
			log.DebugLog(log.DebugLevelMexos, "delete kubernetes server", "name", s.Name)
			err = oscli.DeleteServer(s.Name)
			if err != nil {
				if !force {
					log.DebugLog(log.DebugLevelMexos, "delete server fail", "error", err)
					return err
				}
				log.DebugLog(log.DebugLevelMexos, "forced to continue, error while deleting server", "server", s.Name, "error", err)
			}
			serverDeleted = true
			kconfname := fmt.Sprintf("%s.kubeconfig", s.Name)
			fullpath := mexEnv["MEX_DIR"] + "/" + kconfname
			rerr := os.Remove(fullpath)
			if rerr != nil {
				log.DebugLog(log.DebugLevelMexos, "error can't remove file", "name", fullpath, "error", rerr)
			}
			fullpath += "-proxy"
			rerr = os.Remove(fullpath)
			if rerr != nil {
				log.DebugLog(log.DebugLevelMexos, "error can't remove file", "name", fullpath, "error", rerr)
			}
		}
	}
	if !serverDeleted {
		log.DebugLog(log.DebugLevelMexos, "server not found", "name", name)
		return fmt.Errorf("no server with name %s", name)
	}
	sns, err := oscli.ListSubnets("")
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "can't list subnets", "error", err)
		return err
	}
	rn := oscli.GetMEXExternalRouter() //XXX for now
	log.DebugLog(log.DebugLevelMexos, "removing router from subnet", "routername", rn, "subnets", sns)
	for _, s := range sns {
		if strings.Contains(s.Name, name) {
			log.DebugLog(log.DebugLevelMexos, "removing router from subnet", "router", rn, "subnet", s.Name)
			rerr := oscli.RemoveRouterSubnet(rn, s.Name)
			if rerr != nil {
				log.DebugLog(log.DebugLevelMexos, "not fatal, continue, can't remove router from subnet", "error", rerr)
				//return rerr
			}
			err = oscli.DeleteSubnet(s.Name)
			if err != nil {
				log.DebugLog(log.DebugLevelMexos, "error while deleting subnet", "error", err)
				return err
			}
			break
		}
	}
	//XXX tell agent to remove the route
	//XXX remove kubectl proxy instance
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return err
	}
	cmd := "ps wwh -C kubectl -o pid,args"
	out, err := client.Output(cmd)
	if err != nil || out == "" {
		return nil // no kubectl running
	}
	lines := strings.Split(out, "\n")
	for _, ln := range lines {
		pidnum := parseKCPid(ln, mf.Spec.Key)
		if pidnum == 0 {
			continue
		}
		cmd = fmt.Sprintf("kill -9 %d", pidnum)
		out, err = client.Output(cmd)
		if err != nil {
			log.InfoLog("error killing kubectl proxy", "command", cmd, "out", out, "error", err)
		} else {
			log.DebugLog(log.DebugLevelMexos, "killed kubectl proxy", "line", ln, "cmd", cmd)
		}
		return nil
	}
	return nil
}

//EnableRootLB creates a seed presence node in cloudlet that also becomes first Agent node.
//  It also sets up first basic network router and subnet, ready for running first MEX agent.
func EnableRootLB(mf *Manifest, rootLB *MEXRootLB) error {
	log.DebugLog(log.DebugLevelMexos, "enable rootlb", "name", rootLB)
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("enable rootlb, missing external network in manifest")
	}
	err := oscli.PrepNetwork()
	if err != nil {
		return err
	}
	sl, err := oscli.ListServers()
	if err != nil {
		return err
	}
	if mf.Metadata.DNSZone == "" {
		return fmt.Errorf("missing dns zone in manifest")
	}
	found := 0
	for _, s := range sl {
		if s.Name == rootLB.Name {
			log.DebugLog(log.DebugLevelMexos, "found existing rootlb", "server", s)
			found++
		}
	}
	if found == 0 {
		log.DebugLog(log.DebugLevelMexos, "not found existing server", "name", rootLB.Name)
		netspec := fmt.Sprintf("external-ip,%s", rootLB.PlatConf.Spec.ExternalNetwork)
		if strings.Contains(mf.Spec.Options, "dhcp") {
			netspec = netspec + ",dhcp"
		}
		log.DebugLog(log.DebugLevelMexos, "creating agent node kvm", "mf", mf, "netspec", netspec)
		err := oscli.CreateMEXKVM(rootLB.Name,
			"mex-agent-node", //important, don't change
			netspec,
			mf.Metadata.Tags,
			mf.Metadata.Tenant,
			1)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "error while creating mex kvm", "error", err)
			return err
		}
		log.DebugLog(log.DebugLevelMexos, "created kvm instance", "name", rootLB.Name)
	} else {
		log.DebugLog(log.DebugLevelMexos, "re-using existing kvm instance", "name", rootLB.Name)
	}
	log.DebugLog(log.DebugLevelMexos, "done enabling rootlb", "name", rootLB)
	return nil
}

//XXX allow creating more than one LB

//GetServerIPAddr gets the server IP
func GetServerIPAddr(networkName, serverName string) (string, error) {
	log.DebugLog(log.DebugLevelMexos, "get server ip addr", "networkname", networkName, "servername", serverName)
	//sd, err := oscli.GetServerDetails(rootLB)
	sd, err := oscli.GetServerDetails(serverName)
	if err != nil {
		return "", err
	}
	its := strings.Split(sd.Addresses, "=")
	if len(its) != 2 {
		its = strings.Split(sd.Addresses, ";")
		foundaddr := ""
		if len(its) > 1 {
			for _, it := range its {
				sits := strings.Split(it, "=")
				if len(sits) == 2 {
					if strings.Contains(sits[0], "mex-k8s-net") {
						continue
					}
					if strings.TrimSpace(sits[0]) == networkName { // XXX
						foundaddr = sits[1]
						break
					}
				}
			}
		}
		if foundaddr != "" {
			log.DebugLog(log.DebugLevelMexos, "retrieved server ipaddr", "ipaddr", foundaddr, "netname", networkName, "servername", serverName)
			return foundaddr, nil
		}
		return "", fmt.Errorf("GetServerIPAddr: can't parse server detail addresses, %v, %v", sd, err)
	}
	if its[0] != networkName {
		return "", fmt.Errorf("invalid network name in server detail address, %s", sd.Addresses)
	}
	addr := its[1]
	log.DebugLog(log.DebugLevelMexos, "got server ip addr", "ipaddr", addr, "netname", networkName, "servername", serverName)
	return addr, nil
}

//CopySSHCredential copies over the ssh credential for mex to LB
func CopySSHCredential(serverName, networkName, userName string) error {
	log.DebugLog(log.DebugLevelMexos, "copying ssh credentials", "server", serverName, "network", networkName, "user", userName)
	addr, err := GetServerIPAddr(networkName, serverName)
	if err != nil {
		return err
	}
	kf := mexEnv["MEX_DIR"] + "/" + mexEnv["MEX_SSH_KEY"]
	out, err := sh.Command("scp", "-o", "StrictHostKeyChecking=no", "-i", kf, kf, "root@"+addr+":").Output()
	if err != nil {
		return fmt.Errorf("can't copy %s to %s, %s, %v", kf, addr, out, err)
	}
	return nil
}

//GetSSHClient returns ssh client handle for the server
func GetSSHClient(serverName, networkName, userName string) (ssh.Client, error) {
	auth := ssh.Auth{Keys: []string{mexEnv["MEX_DIR"] + "/" + mexEnv["MEX_SSH_KEY"]}}
	addr, err := GetServerIPAddr(networkName, serverName)
	if err != nil {
		return nil, err
	}
	client, err := ssh.NewNativeClient(userName, addr, "SSH-2.0-mobiledgex-ssh-client-1.0", 22, &auth, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get ssh client, %v", err)
	}
	return client, nil
}

//WaitForRootLB waits for the RootLB instance to be up and copies of SSH credentials for internal networks.
//  Idempotent, but don't call all the time.
func WaitForRootLB(mf *Manifest, rootLB *MEXRootLB) error {
	log.DebugLog(log.DebugLevelMexos, "wait for rootlb", "name", rootLB)
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("waiting for lb, missing external network in manifest")
	}
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return err
	}
	running := false
	for i := 0; i < 10; i++ {
		_, err := client.Output("grep done /tmp/mobiledgex.log") //XXX beware of use of word done
		if err == nil {
			log.DebugLog(log.DebugLevelMexos, "rootlb is running", "name", rootLB)
			running = true
			if err := CopySSHCredential(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root"); err != nil {
				return fmt.Errorf("can't copy ssh credential to RootLB, %v", err)
			}
			break
		}
		log.DebugLog(log.DebugLevelMexos, "wating for rootlb...")
		time.Sleep(30 * time.Second)
	}
	if !running {
		return fmt.Errorf("while creating cluster, timeout waiting for RootLB")
	}
	log.DebugLog(log.DebugLevelMexos, "done waiting for rootlb", "name", rootLB)
	return nil
}

func setPlatConfManifest(mf *Manifest) error {
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	setPlatConf(rootLB, mf)
	return nil
}

func setPlatConf(rootLB *MEXRootLB, mf *Manifest) {
	log.DebugLog(log.DebugLevelMexos, "rootlb platconf set", "rootlb", rootLB, "platconf", mf)
	rootLB.PlatConf = mf
}

//RunMEXAgentManifest runs the MEX agent on the RootLB. It first registers FQDN to cloudflare domain registry if not already registered.
//   It then obtains certficiates from Letsencrypt, if not done yet.  Then it runs the docker instance of MEX agent
//   on the RootLB. It can be told to manually pull image from docker repository.  This allows upgrading with new image.
//   It uses MEX private docker repository.  If an instance is running already, we don't start another one.
func RunMEXAgentManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "run mex agent", "mf", mf)
	fqdn := mf.Spec.RootLB
	//fqdn is that of the machine/kvm-instance running the agent
	if !valid.IsDNSName(fqdn) {
		return fmt.Errorf("fqdn %s is not valid", fqdn)
	}
	if err := setPlatConfManifest(mf); err != nil {
		return fmt.Errorf("can't set plat conf, %v", err)
	}
	sd, err := oscli.GetServerDetails(fqdn)
	if err == nil {
		if sd.Name == fqdn {
			log.DebugLog(log.DebugLevelMexos, "mex agent exists", "fqdn", fqdn)
			return nil
		}
	}
	log.DebugLog(log.DebugLevelMexos, "proceed to create mex agent", "fqdn", fqdn)
	rootLB, err := getRootLB(fqdn)
	if err != nil {
		return fmt.Errorf("cannot find rootlb %s", fqdn)
	}
	if mf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("missing external network")
	}
	if mf.Spec.Agent.Image == "" {
		return fmt.Errorf("missing agent image")
	}
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name")
	}
	log.DebugLog(log.DebugLevelMexos, "record platform config", "mf", mf, "rootLB", rootLB)
	err = EnableRootLB(mf, rootLB)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "can't enable agent", "name", rootLB.Name)
		return fmt.Errorf("Failed to enable root LB %v", err)
	}
	err = WaitForRootLB(mf, rootLB)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "timeout wating for agent to run", "name", rootLB.Name)
		return fmt.Errorf("Error waiting for rootLB %v", err)
	}
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return err
	}
	//XXX rewrite this with --format {{.Names}}
	cmd := fmt.Sprintf("docker ps --filter ancestor=%s --format {{.Names}}", mf.Spec.Agent.Image)
	out, err := client.Output(cmd)
	if err == nil && strings.Contains(out, rootLB.Name) {
		//agent docker instance exists
		//XXX check better
		log.DebugLog(log.DebugLevelMexos, "agent docker instance already running")
		return nil
	}
	if err = ActivateFQDNA(mf, rootLB, rootLB.Name); err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "FQDN A record activated", "name", rootLB.Name)
	err = AcquireCertificates(mf, rootLB, rootLB.Name) //fqdn name may be different than rootLB.Name
	if err != nil {
		return fmt.Errorf("can't acquire certificate for %s, %v", rootLB.Name, err)
	}
	log.DebugLog(log.DebugLevelMexos, "acquired certificates from letsencrypt", "name", rootLB.Name)
	if mexEnv["MEX_DOCKER_REG_PASS"] == "" {
		return fmt.Errorf("empty docker registry pass env var")
	}
	cmd = fmt.Sprintf("echo %s > .docker-pass", mexEnv["MEX_DOCKER_REG_PASS"])
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't store docker pass, %s, %v", out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "seeded docker registry password")
	dockerinstanceName := fmt.Sprintf("%s-%s", mf.Metadata.Name, rootLB.Name)
	if mf.Spec.DockerRegistry == "" {
		log.DebugLog(log.DebugLevelMexos, "warning, empty docker registry spec, using default.")
		mf.Spec.DockerRegistry = "registry.mobiledgex.net:5000" // XXX
	}
	cmd = fmt.Sprintf("cat .docker-pass| docker login -u mobiledgex --password-stdin %s", mf.Spec.DockerRegistry)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error docker login at %s, %s, %s, %v", rootLB.Name, cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "docker login ok")
	cmd = fmt.Sprintf("docker pull %s", mf.Spec.Agent.Image) //probably redundant
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error pulling docker image at %s, %s, %s, %v", rootLB.Name, cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "pulled agent image ok")
	cmd = fmt.Sprintf("docker run -d --rm --name %s --net=host -v `pwd`:/var/www/.cache -v /etc/ssl/certs:/etc/ssl/certs %s -debug", dockerinstanceName, mf.Spec.Agent.Image)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error running dockerized agent on RootLB %s, %s, %s, %v", rootLB.Name, cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "now running dockerized agent")
	return nil
}

//UpdateMEXAgentManifest upgrades the mex agent
func UpdateMEXAgentManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "update mex agent", "mf", mf)
	err := RemoveMEXAgentManifest(mf)
	if err != nil {
		return err
	}
	// Force pulling a potentially newer docker image
	return RunMEXAgentManifest(mf)
}

//RemoveMEXAgentManifest deletes mex agent docker instance
func RemoveMEXAgentManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "deleting mex agent", "mf", mf)
	//XXX we are deleting server kvm!!!
	err := oscli.DeleteServer(mf.Spec.RootLB)
	force := strings.Contains(mf.Spec.Flags, "force")
	if err != nil {
		if !force {
			return err
		}
		log.DebugLog(log.DebugLevelMexos, "forced to continue, deleting mex agent error", "error", err, "rootLB", mf.Spec.RootLB)
	}
	log.DebugLog(log.DebugLevelMexos, "removed rootlb", "name", mf.Spec.RootLB)
	if mf.Metadata.DNSZone == "" {
		return fmt.Errorf("missing dns zone in manifest")
	}
	if cerr := cloudflare.InitAPI(mexEnv["MEX_CF_USER"], mexEnv["MEX_CF_KEY"]); cerr != nil {
		return fmt.Errorf("cannot init cloudflare api, %v", cerr)
	}
	recs, derr := cloudflare.GetDNSRecords(mf.Metadata.DNSZone)
	fqdn := mf.Spec.RootLB
	if derr != nil {
		return fmt.Errorf("can not get dns records for %s, %v", fqdn, derr)
	}
	for _, rec := range recs {
		if rec.Type == "A" && rec.Name == fqdn {
			err = cloudflare.DeleteDNSRecord(mf.Metadata.DNSZone, rec.ID)
			if err != nil {
				return fmt.Errorf("cannot delete dns record id %s Zone %s, %v", rec.ID, mf.Metadata.DNSZone, err)
			}
		}
	}
	log.DebugLog(log.DebugLevelMexos, "removed DNS A record", "FQDN", fqdn)
	//TODO remove mex-k8s  internal nets and router
	return nil
}

//CheckCredentialsCF checks for Cloudflare
func CheckCredentialsCF() error {
	log.DebugLog(log.DebugLevelMexos, "check for cloudflare credentials")
	for _, envname := range []string{"MEX_CF_KEY", "MEX_CF_USER"} {
		if _, ok := mexEnv[envname]; !ok {
			return fmt.Errorf("no env var for %s", envname)
		}
	}
	return nil
}

func downloadFile(url, fname string) error {
	log.DebugLog(log.DebugLevelMexos, "attempt to retrieve file from url", "file", fname, "URL", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("cannot get file from URL %s, %v", url, err)
	}
	fileout, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("can't create file %s,%v", fname, err)
	}
	_, err = io.Copy(fileout, resp.Body)
	if err != nil {
		return fmt.Errorf("can't write file %s, %v", fname, err)
	}
	resp.Body.Close() // nolint
	fileout.Close()   // nolint
	log.DebugLog(log.DebugLevelMexos, "retrieved file from URL", "URL", url, "file", fname)
	return nil
}

//AcquireCertificates obtains certficates from Letsencrypt over ACME. It should be used carefully. The API calls have quota.
func AcquireCertificates(mf *Manifest, rootLB *MEXRootLB, fqdn string) error {
	log.DebugLog(log.DebugLevelMexos, "acquiring certificates for FQDN", "FQDN", fqdn)
	certRegistryURL := "http://registry.mobiledgex.net:8080/certs/" + fqdn //XXX parameterize this
	url := certRegistryURL + "/fullchain.cer"                              // some clients require fullchain cert
	err := downloadFile(url, "cert.pem")                                   //XXX better file location
	if err == nil {
		url = certRegistryURL + "/" + fqdn + ".key"
		err = downloadFile(url, "key.pem")
		if err == nil {
			//because Letsencrypt complains if we get certs repeated for the same fqdn
			log.DebugLog(log.DebugLevelMexos, "got cached certs from registry", "FQDN", fqdn)
			addr, ierr := GetServerIPAddr(rootLB.PlatConf.Spec.ExternalNetwork, fqdn) //XXX should just use fqdn but paranoid
			if ierr != nil {
				log.DebugLog(log.DebugLevelMexos, "failed to get server ip addr", "FQDN", fqdn, "error", ierr)
				return ierr
			}
			kf := mexEnv["MEX_DIR"] + "/" + mexEnv["MEX_SSH_KEY"]
			out, oerr := sh.Command("scp", "-o", "StrictHostKeyChecking=no", "-i", kf, "cert.pem", "root@"+addr+":").Output()
			if oerr != nil {
				return fmt.Errorf("can't copy %s to %s, %s, %v", kf, addr, out, oerr)
			}
			out, err = sh.Command("scp", "-o", "StrictHostKeyChecking=no", "-i", kf, "key.pem", "root@"+addr+":").Output()
			if err != nil {
				return fmt.Errorf("can't copy %s to %s, %s, %v", kf, addr, out, err)
			}
			log.DebugLog(log.DebugLevelMexos, "copied cert and key", "FQDN", fqdn)
			return nil
		}
		return fmt.Errorf("only cert file exists at %s, %v", certRegistryURL, err)
	}
	log.DebugLog(log.DebugLevelMexos, "did not get cached certs")
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("acquire certificate, missing external network in manifest")
	}
	if cerr := CheckCredentialsCF(); cerr != nil {
		return cerr
	}
	client, err := GetSSHClient(fqdn, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return fmt.Errorf("can't get ssh client for acme.sh, %v", err)
	}
	fullchain := fqdn + "/fullchain.cer"
	cmd := fmt.Sprintf("ls -a %s", fullchain)
	_, err = client.Output(cmd)
	if err == nil {
		return nil
	}
	cmd = fmt.Sprintf("docker run --rm -e CF_Key=%s -e CF_Email=%s -v `pwd`:/acme.sh --net=host neilpang/acme.sh --issue -d %s --dns dns_cf", mexEnv["MEX_CF_KEY"], mexEnv["MEX_CF_USER"], fqdn)

	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error running acme.sh docker, %s, %v", out, err)
	}
	cmd = fmt.Sprintf("ls -a %s", fullchain)
	success := false
	for i := 0; i < 10; i++ {
		_, err = client.Output(cmd)
		if err == nil {
			success = true
			break
		}
		log.DebugLog(log.DebugLevelMexos, "waiting for letsencrypt...")
		time.Sleep(30 * time.Second) // ACME takes minimum 200 seconds
	}
	if !success {
		return fmt.Errorf("timeout waiting for ACME")
	}
	cmd = fmt.Sprintf("cp %s cert.pem", fullchain)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't copy to cert.pem on %s, %s, %v", fqdn, out, err)
	}
	cmd = fmt.Sprintf("cp %s/%s.key key.pem", fqdn, fqdn)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't copy to key.pem on %s, %s, %v", fqdn, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "copied over cert and key", "FQDN", fqdn)
	return nil
}

//ActivateFQDNA updates and ensures FQDN is registered properly
func ActivateFQDNA(mf *Manifest, rootLB *MEXRootLB, fqdn string) error {
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("activate fqdn A record, missing external network in manifest")
	}
	if err := CheckCredentialsCF(); err != nil {
		return err
	}
	if err := cloudflare.InitAPI(mexEnv["MEX_CF_USER"], mexEnv["MEX_CF_KEY"]); err != nil {
		return fmt.Errorf("cannot init cloudflare api, %v", err)
	}
	log.DebugLog(log.DebugLevelMexos, "getting dns record for zone", "DNSZone", mf.Metadata.DNSZone)
	dr, err := cloudflare.GetDNSRecords(mf.Metadata.DNSZone)
	if err != nil {
		return fmt.Errorf("cannot get dns records for %s, %v", fqdn, err)
	}
	addr, err := GetServerIPAddr(rootLB.PlatConf.Spec.ExternalNetwork, fqdn)
	for _, d := range dr {
		if d.Type == "A" && d.Name == fqdn {
			if d.Content == addr {
				log.DebugLog(log.DebugLevelMexos, "existing A record", "FQDN", fqdn, "addr", addr)
				return nil
			}
			log.DebugLog(log.DebugLevelMexos, "cloudflare A record has different address, it will be overwritten", "existing", d, "addr", addr)
			if err = cloudflare.DeleteDNSRecord(mf.Metadata.DNSZone, d.ID); err != nil {
				return fmt.Errorf("can't delete DNS record for %s, %v", fqdn, err)
			}
			break
		}
	}
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "error while talking to cloudflare", "error", err)
		return err
	}
	if err := cloudflare.CreateDNSRecord(mf.Metadata.DNSZone, fqdn, "A", addr, 1, false); err != nil {
		return fmt.Errorf("can't create DNS record for %s, %v", fqdn, err)
	}
	log.DebugLog(log.DebugLevelMexos, "wating for cloudflare...")
	//once successfully inserted the A record will take a bit of time, but not too long due to fast cloudflare anycast
	time.Sleep(10 * time.Second) //XXX verify by doing a DNS lookup
	return nil
}

type genericItem struct {
	Item interface{} `json:"items"`
}

type genericItems struct {
	Items []genericItem `json:"items"`
}

//IsClusterReady checks to see if cluster is read, i.e. rootLB is running and active
func IsClusterReady(mf *Manifest, rootLB *MEXRootLB) (bool, error) {
	log.DebugLog(log.DebugLevelMexos, "checking if cluster is ready", "rootLB", rootLB, "platconf", *rootLB.PlatConf)
	cf, err := GetClusterFlavor(rootLB.PlatConf.Spec.Flavor)
	if err != nil {
		return false, err
	}
	if cf.NumNodes < 1 {
		return false, fmt.Errorf("invalid flavor profile, %v", cf)
	}
	name, err := FindClusterWithKey(mf.Spec.Key)
	if err != nil {
		return false, fmt.Errorf("can't find cluster with key %s, %v", mf.Spec.Key, err)
	}
	ipaddr, err := FindNodeIP(name)
	if err != nil {
		return false, err
	}
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return false, fmt.Errorf("is cluster ready, missing external network in platform config")
	}
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return false, fmt.Errorf("can't get ssh client for cluser ready check, %v", err)
	}
	log.DebugLog(log.DebugLevelMexos, "checking host for kubernetes nodes", "ipaddr", ipaddr)
	cmd := fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %s bob@%s kubectl get nodes -o json", mexEnv["MEX_SSH_KEY"], ipaddr)
	out, err := client.Output(cmd)
	if err != nil {
		return false, nil //This is intentional
	}
	gitems := &genericItems{}
	err = json.Unmarshal([]byte(out), gitems)
	if err != nil {
		return false, fmt.Errorf("failed to json unmarshal kubectl get nodes output, %v", err)
	}
	log.DebugLog(log.DebugLevelMexos, "kubectl reports nodes", "numnodes", len(gitems.Items))
	if len(gitems.Items) < (cf.NumNodes + cf.NumMasterNodes) {
		log.DebugLog(log.DebugLevelMexos, "kubernetes cluster not ready", "log", out)
		return false, nil
	}
	log.DebugLog(log.DebugLevelMexos, "cluster nodes", "numnodes", cf.NumNodes, "nummasters", cf.NumMasterNodes)
	kcpath := mexEnv["MEX_DIR"] + "/" + name + ".kubeconfig"
	//if _, err := os.Stat(kcpath); err == nil {
	//	log.DebugLog(log.DebugLevelMexos, "kubeconfig file exists, will not overwrite", "name", kcpath)
	//	return true, nil
	//}
	if err := CopyKubeConfig(rootLB, name); err != nil {
		return false, fmt.Errorf("kubeconfig copy failed, %v", err)
	}
	log.DebugLog(log.DebugLevelMexos, "cluster ready.", "KUBECONFIG", kcpath)
	return true, nil
}

type clusterDetailKc struct {
	CertificateAuthorityData string `json:"certificate-authority-data"`
	Server                   string `json:"server"`
}

type clusterKc struct {
	Name    string          `json:"name"`
	Cluster clusterDetailKc `json:"cluster"`
}

type clusterKcContextDetail struct {
	Cluster string `json:"cluster"`
	User    string `json:"user"`
}

type clusterKcContext struct {
	Name    string                 `json:"name"`
	Context clusterKcContextDetail `json:"context"`
}

type clusterKcUserDetail struct {
	ClientCertificateData string `json:"client-certificate-data"`
	ClientKeyData         string `json:"client-key-data"`
}

type clusterKcUser struct {
	Name string              `json:"name"`
	User clusterKcUserDetail `json:"user"`
}

type clusterKubeconfig struct {
	APIVersion     string             `json:"apiVersion"`
	Kind           string             `json:"kind"`
	CurrentContext string             `json:"current-context"`
	Users          []clusterKcUser    `json:"users"`
	Clusters       []clusterKc        `json:"clusters"`
	Contexts       []clusterKcContext `json:"contexts"`
	//XXX Missing preferences
}

//CopyKubeConfig copies over kubeconfig from the cluster
func CopyKubeConfig(rootLB *MEXRootLB, name string) error {
	log.DebugLog(log.DebugLevelMexos, "copying kubeconfig", "name", name)
	ipaddr, err := FindNodeIP(name)
	if err != nil {
		return err
	}
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("copy kube config, missing external network in platform config")
	}
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return fmt.Errorf("can't get ssh client for copying kubeconfig, %v", err)
	}
	kconfname := fmt.Sprintf("kubeconfig-%s", name)
	cmd := fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %s bob@%s:.kube/config %s", mexEnv["MEX_SSH_KEY"], ipaddr, kconfname)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't copy kubeconfig from %s, %s, %v", name, out, err)
	}
	cmd = fmt.Sprintf("cat %s", kconfname)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't cat %s, %s, %v", kconfname, out, err)
	}
	port, serr := StartKubectlProxy(rootLB, kconfname)
	if serr != nil {
		return serr
	}
	return ProcessKubeconfig(rootLB, name, port, []byte(out))
}

//ProcessKubeconfig validates kubeconfig and saves it and creates a copy for proxy access
func ProcessKubeconfig(rootLB *MEXRootLB, name string, port int, dat []byte) error {
	log.DebugLog(log.DebugLevelMexos, "process kubeconfig file", "name", name)
	kc := &clusterKubeconfig{}
	err := yaml.Unmarshal(dat, kc)
	if err != nil {
		return fmt.Errorf("can't unmarshal kubeconfig %s, %v", name, err)
	}
	if len(kc.Clusters) < 1 {
		return fmt.Errorf("insufficient clusters info in kubeconfig %s", name)
	}
	//log.Debugln("kubeconfig", kc)
	kconfname := fmt.Sprintf("%s.kubeconfig", name)
	fullpath := mexEnv["MEX_DIR"] + "/" + kconfname
	err = ioutil.WriteFile(fullpath, dat, 0666)
	if err != nil {
		return fmt.Errorf("can't write kubeconfig %s content,%v", name, err)
	}
	log.DebugLog(log.DebugLevelMexos, "wrote kubeconfig", "file", fullpath)
	kc.Clusters[0].Cluster.Server = fmt.Sprintf("http://%s:%d", rootLB.Name, port)
	dat, err = yaml.Marshal(kc)
	if err != nil {
		return fmt.Errorf("can't marshal kubeconfig proxy edit %s, %v", name, err)
	}
	fullpath = mexEnv["MEX_DIR"] + "/" + name + ".kubeconfig-proxy"
	err = ioutil.WriteFile(fullpath, dat, 0666)
	if err != nil {
		return fmt.Errorf("can't write kubeconfig proxy %s, %v", name, err)
	}
	log.DebugLog(log.DebugLevelMexos, "kubeconfig wrote", "file", fullpath)
	return nil
}

//FindNodeIP finds IP for the given node
func FindNodeIP(name string) (string, error) {
	log.DebugLog(log.DebugLevelMexos, "find node ip", "name", name)
	if name == "" {
		return "", fmt.Errorf("empty name")
	}
	srvs, err := oscli.ListServers()
	if err != nil {
		return "", err
	}
	for _, s := range srvs {
		if s.Status == activeService && s.Name == name {
			ipaddr, err := GetInternalIP(s.Name)
			if err != nil {
				return "", fmt.Errorf("can't get IP for %s, %v", s.Name, err)
			}
			return ipaddr, nil
		}
	}
	return "", fmt.Errorf("node %s not found", name)
}

//FindClusterWithKey finds cluster given a key string
func FindClusterWithKey(key string) (string, error) {
	log.DebugLog(log.DebugLevelMexos, "find cluster with key", "key", key)
	if key == "" {
		return "", fmt.Errorf("empty key")
	}
	srvs, err := oscli.ListServers()
	if err != nil {
		return "", err
	}
	keyMaster := mexEnv["MEX_K8S_MASTER"] + "-1-" + key // XXX 1
	for _, s := range srvs {
		if s.Status == activeService && s.Name == keyMaster {
			return s.Name, nil
		}
	}
	return "", fmt.Errorf("key %s not found", key)
}

type ksaPort struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Port       int    `json:"port"`
	NodePort   int    `json:"nodePort"`
	TargetPort int    `json:"targetPort"`
}
type ksaSpec struct {
	Ports []ksaPort `json:"ports"`
}
type kubernetesServiceAbbrev struct {
	Spec ksaSpec `json:"spec"`
}

//CreateKubernetesAppManifest instantiates a new kubernetes deployment
func CreateKubernetesAppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "create kubernetes app", "mf", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if mf.Spec.KubeManifest == "" {
		return fmt.Errorf("missing kubernetes spec")
	}
	if mf.Spec.URI == "" { //XXX TODO register to the DNS registry for public IP app,controller needs to tell us which kind of app
		return fmt.Errorf("empty app URI")
	}
	//TODO: support other URI: file://, nfs://, ftp://, git://, or embedded as base64 string
	if !strings.HasPrefix(mf.Spec.KubeManifest, "http://") &&
		!strings.HasPrefix(mf.Spec.KubeManifest, "https://") {
		return fmt.Errorf("unsupported kubernetes spec %s", mf.Spec.KubeManifest)
	}
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name for kubernetes deployment")
	}
	//if !strings.Contains(mf.Spec.Flavor, "kubernetes") {
	//	return fmt.Errorf("unsupported kubernetes flavor %s", mf.Spec.Flavor)
	//}
	if mf.Spec.Key == "" {
		return fmt.Errorf("empty kubernetes cluster name")
	}
	if mf.Spec.Image == "" {
		return fmt.Errorf("empty kubernetes image")
	}
	if mf.Spec.ProxyPath == "" {
		return fmt.Errorf("empty kubernetes proxy path")
	}
	kubeconfig, client, ipaddr, err := ValidateKubernetesParameters(rootLB, mf.Spec.Key)
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "will launch app into existing cluster", "kubeconfig", kubeconfig, "ipaddr", ipaddr)
	var cmd string
	if mexEnv["MEX_DOCKER_REG_PASS"] == "" {
		return fmt.Errorf("empty docker registry password environment variable")
	}
	cmd = fmt.Sprintf("echo %s > .docker-pass", mexEnv["MEX_DOCKER_REG_PASS"])
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't store docker password, %s, %v", out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "stored docker password")
	cmd = fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %s .docker-pass %s:", mexEnv["MEX_SSH_KEY"], ipaddr)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't copy docker password to k8s-master, %s, %v", out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "copied over docker password")
	cmd = fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %s %s 'cat .docker-pass| docker login -u mobiledgex --password-stdin %s'", mexEnv["MEX_SSH_KEY"], ipaddr, mexEnv["MEX_DOCKER_REGISTRY"])
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't docker login on k8s-master to %s, %s, %v", mexEnv["MEX_DOCKER_REGISTRY"], out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "docker login ok")
	cmd = fmt.Sprintf("%s kubectl create secret docker-registry mexregistrysecret --docker-server=%s --docker-username=mobiledgex --docker-password=%s --docker-email=docker@mobiledgex.com", kubeconfig, mexEnv["MEX_DOCKER_REGISTRY"], mexEnv["MEX_DOCKER_REG_PASS"])
	out, err = client.Output(cmd)
	if err != nil {
		if strings.Contains(out, "AlreadyExists") {
			log.DebugLog(log.DebugLevelMexos, "secret already exists")
		} else {
			return fmt.Errorf("error creating mexregistrysecret, %s, %s, %v", cmd, out, err)
		}
	}
	log.DebugLog(log.DebugLevelMexos, "created mexregistrysecret docker secret")
	cmd = fmt.Sprintf("%s kubectl create -f %s", kubeconfig, mf.Spec.KubeManifest)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deploying kubernetes app, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "applied kubernetes manifest", "kubemanifest", mf.Spec.KubeManifest)
	cmd = fmt.Sprintf("%s kubectl get svc %s-service -o json", kubeconfig, mf.Metadata.Name)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error getting port for kubernetes service, %s, %s, %v", cmd, out, err)
	}
	svcData := &kubernetesServiceAbbrev{}
	err = json.Unmarshal([]byte(out), svcData)
	if err != nil {
		return fmt.Errorf("can't unmarshall kubernetes service data, %v", err)
	}
	log.DebugLog(log.DebugLevelMexos, "got ports for kubernetes service", "kubemanifest", mf.Spec.KubeManifest, "ports", svcData.Spec.Ports)
	for _, port := range svcData.Spec.Ports {
		origin := fmt.Sprintf("http://%s:%d", ipaddr, port.Port)
		proxypath := mf.Spec.ProxyPath + port.Name
		errs := AddPathReverseProxy(rootLB.Name, proxypath, origin)
		if errs != nil {
			errmsg := fmt.Sprintf("%v", errs)
			if strings.Contains(errmsg, "exists") {
				log.DebugLog(log.DebugLevelMexos, "rproxy path already exists", "path", proxypath, "origin", origin)
			} else {
				return fmt.Errorf("Errors adding reverse proxy path, %v", errs)
			}
		}
	}
	cmd = fmt.Sprintf(`%s kubectl patch svc %s-service -p '{"spec":{"externalIPs":["%s"]}}'`, kubeconfig, mf.Metadata.Name, ipaddr)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error patching for kubernetes service, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "patched externalIPs on service", "service", mf.Metadata.Name, "externalIPs", ipaddr)
	return nil
}

//TODO `helm` support

//ValidateKubernetesParameters checks the kubernetes parameters and kubeconfig settings
func ValidateKubernetesParameters(rootLB *MEXRootLB, clustName string) (string, ssh.Client, string, error) {
	log.DebugLog(log.DebugLevelMexos, "validate kubernetes parameters rootLB", "rootLB", rootLB, "cluster", clustName)
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return "", nil, "", fmt.Errorf("validate kubernetes parameters, missing external network in platform config")
	}
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return "", nil, "", err
	}
	name, err := FindClusterWithKey(clustName)
	if err != nil {
		return "", nil, "", fmt.Errorf("can't find cluster with key %s, %v", clustName, err)
	}
	ipaddr, err := FindNodeIP(name)
	if err != nil {
		return "", nil, "", err
	}
	kubeconfig := fmt.Sprintf("KUBECONFIG=kubeconfig-%s", name)
	return kubeconfig, client, ipaddr, nil
}

//KubernetesApplyManifest does `apply` on the manifest yaml
func KubernetesApplyManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "apply kubernetes manifest", "mf", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name")
	}
	kubeconfig, client, _, err := ValidateKubernetesParameters(rootLB, mf.Spec.Key)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("%s kubectl apply -f %s", kubeconfig, mf.Spec.KubeManifest)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error applying kubernetes manifest, %s, %s, %v", cmd, out, err)
	}
	return nil
}

//CreateKubernetesNamespaceManifest creates a new namespace in kubernetes
func CreateKubernetesNamespaceManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "create kubernetes namespace", "mf", mf)
	err := KubernetesApplyManifest(mf)
	if err != nil {
		return fmt.Errorf("error applying kubernetes namespace manifest, %v", err)
	}
	return nil
}

//TODO DeleteKubernetesNamespace

//TODO allow configmap creation from files

//SetKubernetesConfigmapValues sets a key-value in kubernetes configmap
func SetKubernetesConfigmapValues(rootLBName string, clustername string, configname string, keyvalues ...string) error {
	log.DebugLog(log.DebugLevelMexos, "set configmap values", "rootlbname", rootLBName, "clustername", clustername, "configname", configname)
	rootLB, err := getRootLB(rootLBName)
	if err != nil {
		return err
	}
	kubeconfig, client, _, err := ValidateKubernetesParameters(rootLB, clustername)
	if err != nil {
		return err
	}
	//TODO support namespace
	cmd := fmt.Sprintf("%s kubectl create configmap %s ", kubeconfig, configname)
	for _, kv := range keyvalues {
		items := strings.Split(kv, "=")
		if len(items) != 2 {
			return fmt.Errorf("malformed key=value pair, %s", kv)
		}
		cmd = cmd + " --from-literal=" + kv
	}
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error setting key/values to  kubernetes configmap, %s, %s, %v", cmd, out, err)
	}
	return nil
}

//TODO
//func GetKubernetesConfigmapValues(rootLB, clustername, configname string) (map[string]string, error) {
//}

//GetKubernetesConfigmapYAML returns yaml reprentation of the key-values
func GetKubernetesConfigmapYAML(rootLBName string, clustername, configname string) (string, error) {
	log.DebugLog(log.DebugLevelMexos, "get kubernetes configmap", "rootlbname", rootLBName, "clustername", clustername, "configname", configname)
	rootLB, err := getRootLB(rootLBName)
	if err != nil {
		return "", err
	}
	kubeconfig, client, _, err := ValidateKubernetesParameters(rootLB, clustername)
	if err != nil {
		return "", err
	}
	//TODO support namespace
	cmd := fmt.Sprintf("%s kubectl get configmap %s -o yaml", kubeconfig, configname)
	out, err := client.Output(cmd)
	if err != nil {
		return "", fmt.Errorf("error getting configmap yaml, %s, %s, %v", cmd, out, err)
	}
	return out, nil
}

//AddPathReverseProxy adds a new route to origin on the reverse proxy
func AddPathReverseProxy(rootLBName, path, origin string) []error {
	log.DebugLog(log.DebugLevelMexos, "add path to reverse proxy", "rootlbname", rootLBName, "path", path, "origin", origin)
	if path == "" {
		return []error{fmt.Errorf("empty path")}
	}
	if origin == "" {
		return []error{fmt.Errorf("empty origin")}
	}
	request := gorequest.New()
	maURI := fmt.Sprintf("http://%s:%s/v1/proxy", rootLBName, mexEnv["MEX_AGENT_PORT"])
	// The L7 reverse proxy terminates TLS at the RootLB and uses path routing to get to the service at a IP:port
	pl := fmt.Sprintf(`{ "message": "add", "proxies": [ { "path": "/%s/*catchall", "origin": "%s" } ] }`, path, origin)
	resp, body, errs := request.Post(maURI).Set("Content-Type", "application/json").Send(pl).End()
	if errs != nil {
		return errs
	}
	if strings.Contains(body, "OK") {
		log.DebugLog(log.DebugLevelMexos, "added path to revproxy")
		return nil
	}
	errs = append(errs, fmt.Errorf("resp %v, body %s", resp, body))
	return errs
}

//StartKubectlProxy starts kubectl proxy on the rootLB to handle kubectl commands remotely.
//  To be called after copying over the kubeconfig file from cluster to rootLB.
func StartKubectlProxy(rootLB *MEXRootLB, kubeconfig string) (int, error) {
	log.DebugLog(log.DebugLevelMexos, "start kubectl proxy", "rootlb", rootLB, "kubeconfig", kubeconfig)
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return 0, fmt.Errorf("start kubectl proxy, missing external network in platform config")
	}
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return 0, err
	}
	maxPort := 8000
	cmd := "ps wwh -C kubectl -o args"
	out, err := client.Output(cmd)
	if err == nil && out != "" {
		lines := strings.Split(out, "\n")
		for _, ln := range lines {
			portnum := parseKCPort(ln)
			if portnum > maxPort {
				maxPort = portnum
			}
		}
	}
	maxPort++
	log.DebugLog(log.DebugLevelMexos, "Port for kubectl proxy", "maxport", maxPort)
	cmd = fmt.Sprintf("kubectl proxy  --port %d --accept-hosts='.*' --address='0.0.0.0' --kubeconfig=%s ", maxPort, kubeconfig)
	//Use .Start() because we don't want to hang
	cl1, cl2, err := client.Start(cmd)
	if err != nil {
		return 0, fmt.Errorf("error running kubectl proxy, %s,  %v", cmd, err)
	}
	cl1.Close() //nolint
	cl2.Close() //nolint
	cmd = "ps wwh -C kubectl -o args"
	for i := 0; i < 5; i++ {
		//verify
		out, outerr := client.Output(cmd)
		if outerr == nil {
			if out == "" {
				continue
			}
			lines := strings.Split(out, "\n")
			for _, ln := range lines {
				if parseKCPort(ln) == maxPort {
					log.DebugLog(log.DebugLevelMexos, "kubectl confirmed running with port", "port", maxPort)
				}
			}
			return maxPort, nil
		}
		log.DebugLog(log.DebugLevelMexos, "waiting for kubectl proxy...")
		time.Sleep(3 * time.Second)
	}
	return 0, fmt.Errorf("timeout error verifying kubectl proxy")
}

//DestroyKubernetesAppManifest destroys kubernetes app
func DestroyKubernetesAppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "delete kubernetes app", "mf", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if mf.Spec.KubeManifest == "" {
		return fmt.Errorf("missing kubernetes spec")
	}
	if mf.Spec.URI == "" { //XXX TODO register to the DNS registry for public IP app,controller needs to tell us which kind of app
		return fmt.Errorf("empty app URI")
	}
	//TODO: support other URI: file://, nfs://, ftp://, git://, or embedded as base64 string
	if !strings.HasPrefix(mf.Spec.KubeManifest, "http://") &&
		!strings.HasPrefix(mf.Spec.KubeManifest, "https://") {
		return fmt.Errorf("unsupported kubernetes spec %s", mf.Spec.KubeManifest)
	}
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name for kubernetes deployment")
	}
	if !strings.Contains(mf.Spec.Flavor, "kubernetes") {
		return fmt.Errorf("unsupported kubernetes flavor %s", mf.Spec.Flavor)
	}
	if mf.Spec.Key == "" {
		return fmt.Errorf("empty kubernetes cluster name")
	}
	if mf.Spec.Image == "" {
		return fmt.Errorf("empty kubernetes image")
	}
	if mf.Spec.ProxyPath == "" {
		return fmt.Errorf("empty kubernetes proxy path")
	}
	kubeconfig, client, ipaddr, err := ValidateKubernetesParameters(rootLB, mf.Spec.Key)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("%s kubectl get svc %s-service -o json", kubeconfig, mf.Metadata.Name)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error getting port for kubernetes service, %s, %s, %v", cmd, out, err)
	}
	svcData := &kubernetesServiceAbbrev{}
	err = json.Unmarshal([]byte(out), svcData)
	if err != nil {
		return fmt.Errorf("can't unmarshall kubernetes service data, %v", err)
	}
	log.DebugLog(log.DebugLevelMexos, "got ports for kubernetes service", "kubemanifest", mf.Spec.KubeManifest, "ports", svcData.Spec.Ports)
	for _, port := range svcData.Spec.Ports {
		origin := fmt.Sprintf("http://%s:%d", ipaddr, port.Port)
		proxypath := mf.Spec.ProxyPath + port.Name
		errs := DeletePathReverseProxy(rootLB.Name, proxypath, origin)
		if errs != nil {
			return fmt.Errorf("Errors deleting reverse proxy path, %v", errs)
		}
	}
	cmd = fmt.Sprintf("%s kubectl delete service %s", kubeconfig, mf.Metadata.Name+"-service")
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deleting kubernetes service, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "deleted service", "name", mf.Metadata.Name)
	cmd = fmt.Sprintf("%s kubectl delete deploy %s", kubeconfig, mf.Metadata.Name+"-deployment")
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deleting kubernetes deployment, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "deleted deployment", "name", mf.Metadata.Name)
	return nil
}

//DeletePathReverseProxy Deletes a new route to origin on the reverse proxy
func DeletePathReverseProxy(rootLBName, path, origin string) []error {
	//TODO
	return nil
}

//CreateQCOW2AppManifest creates qcow2 app
func CreateQCOW2AppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "create qcow2 vm based app", "mf", mf)
	//TODO: support other URI: file://, nfs://, ftp://, git://, or embedded as base64 string
	if !strings.HasPrefix(mf.Spec.Image, "http://") &&
		!strings.HasPrefix(mf.Spec.Image, "https://") {
		return fmt.Errorf("unsupported qcow2 image spec %s", mf.Spec.Image)
	}
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name for qcow2 deployment")
	}
	if !strings.Contains(mf.Spec.Flavor, "qcow2") {
		return fmt.Errorf("unsupported qcow2 flavor %s", mf.Spec.Flavor)
	}
	if mf.Spec.Key == "" {
		return fmt.Errorf("empty qcow2 cluster name")
	}
	if mf.Spec.Image == "" {
		return fmt.Errorf("empty qcow2 image")
	}
	if mf.Spec.ProxyPath == "" {
		return fmt.Errorf("empty qcow2 proxy path")
	}
	savedQcowName := mf.Metadata.Name + ".qcow2" // XXX somewhere safe instead
	alreadyExist := false
	images, err := oscli.ListImages()
	if err != nil {
		return fmt.Errorf("cannot list openstack images, %v", err)
	}
	for _, img := range images {
		if img.Name == mf.Metadata.Name && img.Status == "active" {
			log.DebugLog(log.DebugLevelMexos, "warning, glance has image already", "name", mf.Metadata.Name)
			if !strings.Contains(mf.Spec.Flags, "force") {
				alreadyExist = true
			} else {
				log.DebugLog(log.DebugLevelMexos, "forced to download image again. delete existing glance image")
				if ierr := oscli.DeleteImage(mf.Metadata.Name); ierr != nil {
					return fmt.Errorf("error deleting glance image %s, %v", mf.Metadata.Name, ierr)
				}
			}
		}
	}
	if !alreadyExist {
		log.DebugLog(log.DebugLevelMexos, "getting qcow2 image", "image", mf.Spec.Image, "name", savedQcowName)
		out, cerr := sh.Command("curl", "-s", "-o", savedQcowName, mf.Spec.Image).Output()
		if cerr != nil {
			return fmt.Errorf("error retrieving qcow image, %s, %s, %v", savedQcowName, out, cerr)
		}
		finfo, serr := os.Stat(savedQcowName)
		if serr != nil {
			if os.IsNotExist(serr) {
				return fmt.Errorf("downloaded qcow2 file %s does not exist, %v", savedQcowName, serr)
			}
			return fmt.Errorf("error looking for downloaded qcow2 file %v", serr)
		}
		if finfo.Size() < 1000 { //too small
			return fmt.Errorf("invalid downloaded qcow2 file %s", savedQcowName)
		}
		log.DebugLog(log.DebugLevelMexos, "qcow2 image being created", "image", mf.Spec.Image, "name", savedQcowName)
		//openstack image create ubuntu-16.04-mex --disk-format qcow2 --container-format bare --public  --file /var/tmp/ubuntu-16.04.qcow2
		err = oscli.CreateImage(mf.Metadata.Name, savedQcowName)
		if err != nil {
			return fmt.Errorf("cannot create openstack glance image instance from %s, %v", savedQcowName, err)
		}
		log.DebugLog(log.DebugLevelMexos, "saved qcow image to glance", "name", mf.Metadata.Name)
		found := false
		for i := 0; i < 10; i++ {
			images, ierr := oscli.ListImages()
			if ierr != nil {
				return fmt.Errorf("error while getting list of qcow2 glance images, %v", ierr)
			}
			for _, img := range images {
				if img.Name == mf.Metadata.Name && img.Status == "active" {
					found = true
					break
				}
			}
			if found {
				break
			}
			log.DebugLog(log.DebugLevelMexos, "waiting for the image to become active", "name", mf.Metadata.Name)
			time.Sleep(2 * time.Second)
		}
		if !found {
			return fmt.Errorf("timed out waiting for glance to activate the qcow2 image %s", mf.Metadata.Name)
		}
	}
	log.DebugLog(log.DebugLevelMexos, "qcow image is active in glance", "name", mf.Metadata.Name)
	if !strings.HasPrefix(mf.Spec.NetworkScheme, "external-ip,") { //XXX for now
		return fmt.Errorf("invalid network scheme for qcow2 kvm app, %s", mf.Spec.NetworkScheme)
	}
	items := strings.Split(mf.Spec.NetworkScheme, ",")
	if len(items) < 2 {
		return fmt.Errorf("can't find external network name in %s", mf.Spec.NetworkScheme)
	}
	extNetwork := items[1]
	opts := &oscli.ServerOpt{
		Name:   mf.Metadata.Name,
		Image:  mf.Metadata.Name,
		Flavor: mf.Spec.ImageFlavor,
		NetIDs: []string{extNetwork},
	}
	//TODO properties
	//TODO userdata
	log.DebugLog(log.DebugLevelMexos, "calling create openstack kvm server", "opts", opts)
	err = oscli.CreateServer(opts)
	if err != nil {
		return fmt.Errorf("can't create openstack kvm server instance %v, %v", opts, err)
	}
	log.DebugLog(log.DebugLevelMexos, "created openstack kvm server", "opts", opts)
	return nil
}

//DestroyQCOW2AppManifest destroys qcow2 app
func DestroyQCOW2AppManifest(mf *Manifest) error {
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name, no openstack kvm to delete")
	}
	if err := oscli.DeleteServer(mf.Metadata.Name); err != nil {
		return fmt.Errorf("cannot delete openstack kvm %s, %v", mf.Metadata.Name, err)
	}
	return nil
}

func parseKCPort(ln string) int {
	if !strings.Contains(ln, "kubectl") {
		return 0
	}
	if !strings.Contains(ln, "--port") {
		return 0
	}
	var a, b, c, port string
	n, serr := fmt.Sscanf(ln, "%s %s %s %s", &a, &b, &c, &port)
	if serr != nil {
		return 0
	}
	if n != 4 {
		return 0
	}
	portnum, aerr := strconv.Atoi(port)
	if aerr != nil {
		return 0
	}
	return portnum
}

func parseKCPid(ln string, key string) int {
	ln = strings.TrimSpace(ln)
	if !strings.Contains(ln, "kubectl") {
		return 0
	}
	if !strings.HasSuffix(ln, key) {
		return 0
	}
	var pid string
	n, serr := fmt.Sscanf(ln, "%s", &pid)
	if serr != nil {
		return 0
	}
	if n != 1 {
		return 0
	}
	pidnum, aerr := strconv.Atoi(pid)
	if aerr != nil {
		return 0
	}
	return pidnum
}
