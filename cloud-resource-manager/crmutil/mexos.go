package crmutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/codeskyblue/go-sh"
	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud-infra/openstack-prov/oscliapi"
	"github.com/mobiledgex/edge-cloud-infra/openstack-tenant/agent/cloudflare" //"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/nanobox-io/golang-ssh"
	"github.com/parnurzeal/gorequest"
	//"github.com/fsouza/go-dockerclient"
)

//mexEnv contains backward compatibility and other environment vars.
//The variables with hard-coded defaults here are for backwards compatibilty
var mexEnv = map[string]string{
	"MEX_DEBUG":           os.Getenv("MEX_DEBUG"),
	"MEX_DIR":             os.Getenv("HOME") + "/.mobiledgex",
	"MEX_CF_KEY":          os.Getenv("MEX_CF_KEY"),
	"MEX_CF_USER":         os.Getenv("MEX_CF_USER"),
	"MEX_DOCKER_REG_PASS": os.Getenv("MEX_DOCKER_REG_PASS"),
	"MEX_EXT_NETWORK":     os.Getenv("MEX_EXT_NETWORK"),
	"MEX_REGISTRY_USER":   "mobiledgex",
	"MEX_AGENT_PORT":      "18889",
	"MEX_REGISTRY":        "registry.mobiledgex.net",
	"MEX_DOCKER_REGISTRY": "registry.mobiledgex.net:5000",
	"MEX_K8S_USER":        "bob",                // backward compatibility => root
	"MEX_SSH_KEY":         "id_rsa_mobiledgex",  // backward compatibility => id_rsa_mex
	"MEX_OS_IMAGE":        "mobiledgex-16.04-2", // backward compatibility => mobiledgex
}

//MEXRootLB has rootLB data
type MEXRootLB struct {
	Name     string
	PlatConf *Manifest
}

//MEXRootLBMap maps name of rootLB to rootLB instance
var MEXRootLBMap = make(map[string]*MEXRootLB)

var defaultPrivateNetRange = "10.101.X.0/24"

//AvailableClusterFlavors lists currently available flavors
var AvailableClusterFlavors = []*ClusterFlavor{
	&ClusterFlavor{
		Name:           "x1.tiny",
		Kind:           "mex-cluster-flavor",
		PlatformFlavor: "m4.small",
		Status:         "active",
		NumNodes:       2,
		NumMasterNodes: 1,
		Topology:       "type-1",
		NetworkSpec:    "priv-subnet,mex-k8s-net-1," + defaultPrivateNetRange,
		StorageSpec:    "default",
		NodeFlavor:     ClusterNodeFlavor{Name: "k8s-tiny", Type: "k8s-node"},
		MasterFlavor:   ClusterMasterFlavor{Name: "k8s-tiny", Type: "k8s-master"},
	},
	&ClusterFlavor{
		Name:           "x1.small",
		Kind:           "mex-cluster-flavor",
		PlatformFlavor: "m4.medium",
		Status:         "active",
		NumNodes:       2,
		NumMasterNodes: 1,
		Topology:       "type-1",
		NetworkSpec:    "priv-subnet,mex-k8s-net-1," + defaultPrivateNetRange,
		StorageSpec:    "default",
		NodeFlavor:     ClusterNodeFlavor{Name: "k8s-small", Type: "k8s-node"},
		MasterFlavor:   ClusterMasterFlavor{Name: "k8s-small", Type: "k8s-master"},
	},
	&ClusterFlavor{
		Name:           "x1.medium",
		Kind:           "mex-cluster-flavor",
		PlatformFlavor: "m4.large",
		Status:         "active",
		NumNodes:       2,
		NumMasterNodes: 1,
		Topology:       "type-1",
		NetworkSpec:    "priv-subnet,mex-k8s-net-1," + defaultPrivateNetRange,
		StorageSpec:    "default",
		NodeFlavor:     ClusterNodeFlavor{Name: "k8s-medium", Type: "k8s-node"},
		MasterFlavor:   ClusterMasterFlavor{Name: "k8s-medium", Type: "k8s-master"},
	},
	&ClusterFlavor{
		Name:           "x1.large",
		Kind:           "mex-cluster-flavor",
		PlatformFlavor: "m4.xlarge",
		Status:         "active",
		NumNodes:       2,
		NumMasterNodes: 1,
		Topology:       "type-1",
		NetworkSpec:    "priv-subnet,mex-k8s-net-1," + defaultPrivateNetRange,
		StorageSpec:    "default",
		NodeFlavor:     ClusterNodeFlavor{Name: "k8s-large", Type: "k8s-node"},
		MasterFlavor:   ClusterMasterFlavor{Name: "k8s-large", Type: "k8s-master"},
	},
}

var sshOpts = []string{"StrictHostKeyChecking=no", "UserKnownHostsFile=/dev/null"}

//IsValidMEXOSEnv caches the validity of the env
var IsValidMEXOSEnv = false

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
	Kind           string
	Name           string
	PlatformFlavor string
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
	for _, evar := range []string{"MEX_CF_KEY", "MEX_CF_USER", "MEX_DOCKER_REG_PASS", "MEX_DIR", "MEX_REGISTRY_USER"} {
		if _, ok := mexEnv[evar]; !ok {
			return fmt.Errorf("missing env var %s", evar)
		}
	}
	//original base VM image uses id_rsa_mobiledgex key
	if mexEnv["MEX_OS_IMAGE"] == "mobiledgex-16.04-2" && mexEnv["MEX_SSH_KEY"] != "id_rsa_mobiledgex" {
		return fmt.Errorf("os image %s cannot use key %s", mexEnv["MEX_OS_IMAGE"], mexEnv["MEX_SSH_KEY"])
	}
	//packer VM image uses id_rsa_mex key
	if mexEnv["MEX_OS_IMAGE"] == "mobiledgex" && mexEnv["MEX_SSH_KEY"] != "id_rsa_mex" {
		return fmt.Errorf("os image %s cannot use key %s", mexEnv["MEX_OS_IMAGE"], mexEnv["MEX_SSH_KEY"])
	}
	//TODO need to allow users to save the environment under platform name inside .mobiledgex or Vault
	return nil
}

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
	if rootLB == nil {
		log.DebugLog(log.DebugLevelMexos, "error, newrootlbmanifest, rootLB is null")
	}
	return rootLB, nil
}

//DeleteRootLB to be called by code that called NewRootLB
func DeleteRootLB(rootLBName string) {
	delete(MEXRootLBMap, rootLBName) //no mutex because caller should be serializing New/Delete in a control loop
}

//ValidateMEXOSEnv makes sure the environment is valid for mexos
func ValidateMEXOSEnv(osEnvValid bool) bool {
	IsValidMEXOSEnv = false
	if !osEnvValid {
		log.DebugLog(log.DebugLevelMexos, "invalid mex env")
		return false
	}
	IsValidMEXOSEnv = true
	log.DebugLog(log.DebugLevelMexos, "valid mex env")
	return IsValidMEXOSEnv
}

func AddFlavorManifest(mf *Manifest) error {
	_, err := GetClusterFlavor(mf.Spec.Flavor)
	if err != nil {
		return err
	}
	// Adding flavors in platforms cannot be done dynamically. For example, x1.xlarge cannot be
	// implemented in currently DT cloudlets. Controller can learn what flavors available. Not create new ones.
	return nil
}

func GetClusterFlavor(flavor string) (*ClusterFlavor, error) {
	log.DebugLog(log.DebugLevelMexos, "get cluster flavor details", "cluster flavor", flavor)
	for _, af := range AvailableClusterFlavors {
		if af.Name == flavor {
			log.DebugLog(log.DebugLevelMexos, "using cluster flavor", "cluster flavor", af)
			return af, nil
		}
	}
	return nil, fmt.Errorf("unsupported cluster flavor %s", flavor)
}

//mexCreateClusterKubernetes creates a cluster of nodes. It can take a while, so call from a goroutine.
func mexCreateClusterKubernetes(mf *Manifest) error {
	//func mexCreateClusterKubernetes(mf *Manifest) (*string, error) {
	log.DebugLog(log.DebugLevelMexos, "create kubernetes cluster", "mf", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if rootLB == nil {
		return fmt.Errorf("can't create kubernetes cluster, rootLB is null")
	}
	if mf.Spec.Flavor == "" {
		return fmt.Errorf("empty cluster flavor")
	}
	cf, err := GetClusterFlavor(mf.Spec.Flavor)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "invalid platform flavor, can't create cluster", "mf", mf)
		return err
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
	mf.Metadata.Tags += "," + cf.PlatformFlavor
	//TODO add whole manifest yaml->json into stringified property of the kvm instance for later
	//XXX should check for quota, permissions, access control, etc. here
	//guid := xid.New().String()
	//kvmname := fmt.Sprintf("%s-1-%s-%s", "mex-k8s-master", mf.Metadata.Name, guid)
	kvmname := fmt.Sprintf("%s-1-%s", "mex-k8s-master", mf.Metadata.Name)
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
		cf.PlatformFlavor,
	)
	if err != nil {
		return fmt.Errorf("can't create k8s master, %v", err)
	}
	for i := 1; i <= cf.NumNodes; i++ {
		//construct node name
		//kvmnodename := fmt.Sprintf("%s-%d-%s-%s", "mex-k8s-node", i, mf.Metadata.Name, guid)
		kvmnodename := fmt.Sprintf("%s-%d-%s", "mex-k8s-node", i, mf.Metadata.Name)
		err = oscli.CreateMEXKVM(kvmnodename,
			"k8s-node",
			mf.Spec.NetworkScheme,
			mf.Metadata.Tags,
			mf.Metadata.Tenant,
			i,
			cf.PlatformFlavor,
		)
		if err != nil {
			return fmt.Errorf("can't create k8s node, %v", err)
		}
	}
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("missing external network in platform config")
	}
	if err = LBAddRoute(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, kvmname); err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot add route on rootlb")
		//return err
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
		log.DebugLog(log.DebugLevelMexos, "waiting for kubernetes cluster to be ready...")
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

func getRootLB(name string) (*MEXRootLB, error) {
	rootLB, ok := MEXRootLBMap[name]
	if !ok {
		return nil, fmt.Errorf("can't find rootlb %s", name)
	}
	if rootLB == nil {
		log.DebugLog(log.DebugLevelMexos, "getrootlb, rootLB is null")
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
	if rootLB == nil {
		return fmt.Errorf("can't delete kubernetes cluster, rootLB is null")
	}
	name := mf.Metadata.Name
	if name == "" {
		log.DebugLog(log.DebugLevelMexos, "error, empty name", "mf", mf)
		return fmt.Errorf("empty name")
	}
	srvs, err := oscli.ListServers()
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "looking for server", "name", name, "servers", srvs)
	force := strings.Contains(mf.Spec.Flags, "force")
	serverDeleted := false
	for _, s := range srvs {
		if strings.Contains(s.Name, name) {
			if strings.Contains(s.Name, "mex-k8s-master") {
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
			kconfname := fmt.Sprintf("%s.kubeconfig", s.Name[strings.LastIndex(s.Name, "-")+1:])
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
	if rootLB == nil {
		return fmt.Errorf("cannot enable rootLB, rootLB is null")
	}
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
		return fmt.Errorf("missing dns zone in manifest, metadata %v", mf.Metadata)
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
		cf, err := GetClusterFlavor(mf.Spec.Flavor)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "invalid platform flavor, can't create rootLB", "mf", mf, "rootLB", rootLB)
			return fmt.Errorf("cannot create rootLB invalid platform flavor %v", err)
		}
		log.DebugLog(log.DebugLevelMexos, "creating agent node kvm", "mf", mf, "netspec", netspec)
		err = oscli.CreateMEXKVM(rootLB.Name,
			"mex-agent-node", //important, don't change
			netspec,
			mf.Metadata.Tags,
			mf.Metadata.Tenant,
			1,
			cf.PlatformFlavor,
		)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "error while creating mex kvm", "error", err)
			return err
		}
		log.DebugLog(log.DebugLevelMexos, "created kvm instance", "name", rootLB.Name)

		rootLBIPaddr, ierr := GetServerIPAddr(rootLB.PlatConf.Spec.ExternalNetwork, rootLB.Name)
		if ierr != nil {
			log.DebugLog(log.DebugLevelMexos, "cannot get rootlb IP address", "error", ierr)
			return fmt.Errorf("created rootlb but cannot get rootlb IP")
		}
		ports := []int{
			18889, //mexosagent HTTP server
			18888, //mexosagent GRPC server
			443,   //mexosagent reverse proxy HTTPS
			8001,  //kubectl proxy
			6443,  //kubernetes control
			8000,  //mex k8s join token server
		}
		ruleName := oscli.GetDefaultSecurityRule()
		privateNetCIDR := strings.Replace(defaultPrivateNetRange, "X", "0", 1)
		allowedClientCIDR := GetAllowedClientCIDR()
		for _, p := range ports {
			for _, cidr := range []string{rootLBIPaddr + "/32", privateNetCIDR, allowedClientCIDR} {
				err := oscli.AddSecurityRuleCIDR(cidr, "tcp", ruleName, p)
				if err != nil {
					log.DebugLog(log.DebugLevelMexos, "warning, error while adding security rule", "error", err, "cidr", cidr, "rulename", ruleName, "port", p)
				}
			}
		}
		//TODO: removal of security rules. Needs to be done for general resource per VM object.
		//    Add annotation to the running VM. When VM is removed, go through annotations
		//   and undo the resource allocations, like security rules, etc.
	} else {
		log.DebugLog(log.DebugLevelMexos, "re-using existing kvm instance", "name", rootLB.Name)
	}
	log.DebugLog(log.DebugLevelMexos, "done enabling rootlb", "name", rootLB)
	return nil
}

func GetAllowedClientCIDR() string {
	//XXX TODO get real list of allowed clients from remote database or template configuration
	return "0.0.0.0/0"
}

//XXX allow creating more than one LB

//GetServerIPAddr gets the server IP
func GetServerIPAddr(networkName, serverName string) (string, error) {
	//TODO: mexosagent cache
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
	out, err := sh.Command("scp", "-o", sshOpts[0], "-o", sshOpts[1], "-i", kf, kf, "root@"+addr+":").Output()
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
	log.DebugLog(log.DebugLevelMexos, "got ssh client", "addr", addr, "key", auth)
	return client, nil
}

//WaitForRootLB waits for the RootLB instance to be up and copies of SSH credentials for internal networks.
//  Idempotent, but don't call all the time.
func WaitForRootLB(mf *Manifest, rootLB *MEXRootLB) error {
	log.DebugLog(log.DebugLevelMexos, "wait for rootlb", "name", rootLB)
	if rootLB == nil {
		return fmt.Errorf("cannot wait for lb, rootLB is null")
	}

	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("waiting for lb, missing external network in manifest")
	}
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return err
	}
	running := false
	for i := 0; i < 10; i++ {
		log.DebugLog(log.DebugLevelMexos, "waiting for rootlb...")
		_, err := client.Output("grep done /tmp/mobiledgex.log") //XXX beware of use of word done
		if err == nil {
			log.DebugLog(log.DebugLevelMexos, "rootlb is running", "name", rootLB)
			running = true
			if err := CopySSHCredential(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root"); err != nil {
				return fmt.Errorf("can't copy ssh credential to RootLB, %v", err)
			}
			break
		}
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
	if rootLB == nil {
		return fmt.Errorf("cannot set platform config manifest, rootLB is null")
	}
	setPlatConf(rootLB, mf)
	return nil
}

func setPlatConf(rootLB *MEXRootLB, mf *Manifest) {
	log.DebugLog(log.DebugLevelMexos, "rootlb platconf set", "rootlb", rootLB, "platconf", mf)
	if rootLB == nil {
		log.DebugLog(log.DebugLevelMexos, "cannot set platconf, rootLB is null")
	}

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
			log.DebugLog(log.DebugLevelMexos, "server with same name as rootLB exists", "fqdn", fqdn)
			rootLB, err := getRootLB(fqdn)
			if err != nil {
				return fmt.Errorf("cannot find rootlb %s", fqdn)
			}
			//return RunMEXOSAgentContainer(mf, rootLB)
			return RunMEXOSAgentService(mf, rootLB)
		}
	}
	log.DebugLog(log.DebugLevelMexos, "proceed to create mex agent", "fqdn", fqdn)
	rootLB, err := getRootLB(fqdn)
	if err != nil {
		return fmt.Errorf("cannot find rootlb %s", fqdn)
	}
	if rootLB == nil {
		return fmt.Errorf("cannot run mex agent manifest, rootLB is null")
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
		log.DebugLog(log.DebugLevelMexos, "timeout waiting for agent to run", "name", rootLB.Name)
		return fmt.Errorf("Error waiting for rootLB %v", err)
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
	//return RunMEXOSAgentContainer(mf, rootLB)
	return RunMEXOSAgentService(mf, rootLB)
}

func RunMEXOSAgentService(mf *Manifest, rootLB *MEXRootLB) error {
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return err
	}
	for _, act := range []string{"stop", "disable"} {
		out, err := client.Output("systemctl " + act + " mexosagent.service")
		if err != nil {
			log.InfoLog("warning: cannot "+act+" mexosagent.service", "out", out, "err", err)
		}
	}
	log.DebugLog(log.DebugLevelMexos, "copying new mexosagent service")
	for _, dest := range []struct{ path, name string }{
		{"/usr/local/bin/", "mexosagent"},
		{"/lib/systemd/system/", "mexosagent.service"},
	} {
		cmd := fmt.Sprintf("scp -o %s -o %s -i %s %s@%s:files-repo/mobiledgex/%s %s", sshOpts[0], sshOpts[1], mexEnv["MEX_SSH_KEY"], mexEnv["MEX_REGISTRY_USER"], mexEnv["MEX_REGISTRY"], dest.name, dest.path)
		out, err := client.Output(cmd)
		if err != nil {
			log.InfoLog("error: cannot download from registry", "fn", dest.name, "path", dest.path, "error", err, "out", out)
			return err
		}
		out, err = client.Output("chmod a+rx " + dest.path + dest.name)
		if err != nil {
			log.InfoLog("error: cannot chmod", "error", err, "fn", dest.name, "path", dest.path)
			return err
		}
	}
	log.DebugLog(log.DebugLevelMexos, "starting mexosagent.service")
	for _, act := range []string{"enable", "start"} {
		out, err := client.Output("systemctl " + act + " mexosagent.service")
		if err != nil {
			log.InfoLog("warning: cannot "+act+" mexosagent.service", "out", out, "err", err)
		}
	}
	log.DebugLog(log.DebugLevelMexos, "started mexosagent.service")
	return nil
}

func RunMEXOSAgentContainer(mf *Manifest, rootLB *MEXRootLB) error {
	if mexEnv["MEX_DOCKER_REG_PASS"] == "" {
		return fmt.Errorf("empty docker registry pass env var")
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
	cmd = fmt.Sprintf("echo %s > .docker-pass", mexEnv["MEX_DOCKER_REG_PASS"])
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't store docker pass, %s, %v", out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "seeded docker registry password")
	dockerinstanceName := fmt.Sprintf("%s-%s", mf.Metadata.Name, rootLB.Name)
	if mf.Spec.DockerRegistry == "" {
		log.DebugLog(log.DebugLevelMexos, "warning, empty docker registry spec, using default.")
		mf.Spec.DockerRegistry = mexEnv["MEX_DOCKER_REGISTRY"]
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
	log.DebugLog(log.DebugLevelMexos, "now running dockerized mexosagent")
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
		return fmt.Errorf("missing dns zone in manifest, metadata %v", mf.Metadata)
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

func checkPEMFile(fn string) error {
	content, err := ioutil.ReadFile(fn)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "can't read downloaded pem file", "name", fn, "error", err)
		return fmt.Errorf("can't read downloaded pem file %s, %v", fn, err)
	}
	contstr := string(content)
	if strings.Contains(contstr, "404 Not Found") {
		log.DebugLog(log.DebugLevelMexos, "404 not found in pem file", "name", fn)
		return fmt.Errorf("registry does not have the pem file %s", fn)
	}
	if !strings.HasPrefix(contstr, "-----BEGIN") {
		log.DebugLog(log.DebugLevelMexos, "does not look like pem file", "name", fn)
		return fmt.Errorf("does not look like pem file %s", fn)
	}
	return nil
}

//AcquireCertificates obtains certficates from Letsencrypt over ACME. It should be used carefully. The API calls have quota.
func AcquireCertificates(mf *Manifest, rootLB *MEXRootLB, fqdn string) error {
	log.DebugLog(log.DebugLevelMexos, "acquiring certificates for FQDN", "FQDN", fqdn)
	if rootLB == nil {
		return fmt.Errorf("cannot acquire certs, rootLB is null")
	}
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("acquire certificate, missing external network in manifest")
	}
	if cerr := CheckCredentialsCF(); cerr != nil {
		return cerr
	}
	kf := mexEnv["MEX_DIR"] + "/" + mexEnv["MEX_SSH_KEY"]
	srcfile := fmt.Sprintf("%s@%s:files-repo/certs/%s/fullchain.cer", mexEnv["MEX_REGISTRY_USER"], mexEnv["MEX_REGISTRY"], fqdn)
	dkey := fmt.Sprintf("%s/%s.key", fqdn, fqdn)
	certfile := "cert.pem" //XXX better file location
	keyfile := "key.pem"   //XXX better location
	out, err := sh.Command("scp", "-o", sshOpts[0], "-o", sshOpts[1], "-i", kf, srcfile, certfile).Output()
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "warning, failed to get cached cert file", "src", srcfile, "cert", certfile, "error", err, "out", out)
	} else if checkPEMFile(certfile) == nil {
		srcfile = fmt.Sprintf("%s@%s:files-repo/certs/%s", mexEnv["MEX_REGISTRY_USER"], mexEnv["MEX_REGISTRY"], dkey)
		out, err = sh.Command("scp", "-o", sshOpts[0], "-o", sshOpts[1], "-i", kf, srcfile, keyfile).Output()
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "warning, failed to get cached key file", "src", srcfile, "cert", certfile, "error", err, "out", out)
		} else if checkPEMFile(keyfile) == nil {
			//because Letsencrypt complains if we get certs repeated for the same fqdn
			log.DebugLog(log.DebugLevelMexos, "got cached certs from registry", "FQDN", fqdn)
			addr, ierr := GetServerIPAddr(rootLB.PlatConf.Spec.ExternalNetwork, fqdn) //XXX should just use fqdn but paranoid
			if ierr != nil {
				log.DebugLog(log.DebugLevelMexos, "failed to get server ip addr", "FQDN", fqdn, "error", ierr)
				return ierr
			}
			for _, fn := range []string{certfile, keyfile} {
				out, oerr := sh.Command("scp", "-o", sshOpts[0], "-o", sshOpts[1], "-i", kf, fn, "root@"+addr+":").Output()
				if oerr != nil {
					return fmt.Errorf("can't copy %s to %s, %v, %v", fn, addr, oerr, out)
				}
				log.DebugLog(log.DebugLevelMexos, "copied", "fn", fn, "addr", addr)
			}
			log.DebugLog(log.DebugLevelMexos, "using cached cert and key", "FQDN", fqdn)
			return nil
		}
	}
	log.DebugLog(log.DebugLevelMexos, "did not get cached cert and key files, will try to acquire new cert")
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
	res, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error running acme.sh docker, %s, %v", res, err)
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
	for _, d := range []struct{ src, dest string }{
		{fullchain, certfile},
		{dkey, keyfile},
	} {
		cmd = fmt.Sprintf("cp %s %s", d.src, d.dest)
		res, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("fail to copy %s to %s on %s, %v, %v", d.src, d.dest, fqdn, err, res)
		}
	}
	cmd = fmt.Sprintf("scp -o %s -o %s -i %s -r %s %s@%s:files-repo/certs", sshOpts[0], sshOpts[1], mexEnv["MEX_SSH_KEY"], fqdn, mexEnv["MEX_REGISTRY_USER"], mexEnv["MEX_REGISTRY"]) // XXX
	res, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("failed to upload certs for %s to %s, %v, %v", fqdn, mexEnv["MEX_REGISTRY"], err, res)
	}
	log.DebugLog(log.DebugLevelMexos, "saved acquired cert and key", "FQDN", fqdn)
	return nil
}

//ActivateFQDNA updates and ensures FQDN is registered properly
func ActivateFQDNA(mf *Manifest, rootLB *MEXRootLB, fqdn string) error {
	if rootLB == nil {
		return fmt.Errorf("cannot activate certs, rootLB is null")
	}

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
	log.DebugLog(log.DebugLevelMexos, "waiting for cloudflare...")
	//once successfully inserted the A record will take a bit of time, but not too long due to fast cloudflare anycast
	//err = WaitforDNSRegistration(fqdn)
	//if err != nil {
	//	return err
	//}
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
	if rootLB == nil {
		return false, fmt.Errorf("cannot check if cluster is ready, rootLB is null")
	}
	cf, err := GetClusterFlavor(mf.Spec.Flavor)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "invalid cluster flavor, can't check if cluster is ready", "mf", mf, "rootLB", rootLB)
		return false, err
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
	log.DebugLog(log.DebugLevelMexos, "checking master k8s node for available nodes", "ipaddr", ipaddr)
	cmd := fmt.Sprintf("ssh -o %s -o %s -i %s %s@%s kubectl get nodes -o json", sshOpts[0], sshOpts[1], mexEnv["MEX_SSH_KEY"], mexEnv["MEX_K8S_USER"], ipaddr)
	out, err := client.Output(cmd)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "error checking for kubernetes nodes", "out", out, "err", err)
		return false, nil //This is intentional
	}
	gitems := &genericItems{}
	err = json.Unmarshal([]byte(out), gitems)
	if err != nil {
		return false, fmt.Errorf("failed to json unmarshal kubectl get nodes output, %v, %v", err, out)
	}
	log.DebugLog(log.DebugLevelMexos, "kubectl reports nodes", "numnodes", len(gitems.Items))
	if len(gitems.Items) < (cf.NumNodes + cf.NumMasterNodes) {
		log.DebugLog(log.DebugLevelMexos, "kubernetes cluster not ready", "log", out)
		return false, nil
	}
	log.DebugLog(log.DebugLevelMexos, "cluster nodes", "numnodes", cf.NumNodes, "nummasters", cf.NumMasterNodes)
	kcpath := mexEnv["MEX_DIR"] + "/" + name[strings.LastIndex(name, "-")+1:] + ".kubeconfig"
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
	if rootLB == nil {
		return fmt.Errorf("cannot copy kubeconfig, rootLB is null")
	}
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
	kconfname := fmt.Sprintf("%s.kubeconfig", name[strings.LastIndex(name, "-")+1:])
	cmd := fmt.Sprintf("scp -o %s -o %s -i %s %s@%s:.kube/config %s", sshOpts[0], sshOpts[1], mexEnv["MEX_SSH_KEY"], mexEnv["MEX_K8S_USER"], ipaddr, kconfname)
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
	if rootLB == nil {
		return fmt.Errorf("cannot process kubeconfig, rootLB is null")
	}
	kc := &clusterKubeconfig{}
	err := yaml.Unmarshal(dat, kc)
	if err != nil {
		return fmt.Errorf("can't unmarshal kubeconfig %s, %v", name, err)
	}
	if len(kc.Clusters) < 1 {
		return fmt.Errorf("insufficient clusters info in kubeconfig %s", name)
	}
	//log.Debugln("kubeconfig", kc)
	//TODO: the kconfname should include more details, location, tags, to be distinct from other clusters in other regions
	kconfname := fmt.Sprintf("%s.kubeconfig", name[strings.LastIndex(name, "-")+1:])
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
	fullpath = fullpath + "-proxy"
	err = ioutil.WriteFile(fullpath, dat, 0666)
	if err != nil {
		return fmt.Errorf("can't write kubeconfig proxy %s, %v", fullpath, err)
	}
	log.DebugLog(log.DebugLevelMexos, "kubeconfig-proxy file saved", "file", fullpath)
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
		if s.Status == "ACTIVE" && s.Name == name {
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
	for _, s := range srvs {
		if s.Status == "ACTIVE" && strings.HasSuffix(s.Name, key) && strings.HasPrefix(s.Name, "mex-k8s-master") {
			log.DebugLog(log.DebugLevelMexos, "find cluster with key", "key", key, "found", s.Name)
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

func addSecurityRules(rootLB *MEXRootLB, mf *Manifest, kp *kubeParam) error {
	rootLBIPaddr, err := GetServerIPAddr(rootLB.PlatConf.Spec.ExternalNetwork, rootLB.Name)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot get rootlb IP address", "error", err)
		return fmt.Errorf("cannot deploy kubernetes app, cannot get rootlb IP")
	}
	sr := oscli.GetDefaultSecurityRule()
	allowedClientCIDR := GetAllowedClientCIDR()
	for _, port := range mf.Spec.Ports {
		for _, sec := range []struct {
			addr string
			port int
		}{
			{rootLBIPaddr + "/32", port.PublicPort},
			{kp.ipaddr + "/32", port.PublicPort},
			{allowedClientCIDR, port.PublicPort},
			{rootLBIPaddr + "/32", port.InternalPort},
			{kp.ipaddr + "/32", port.InternalPort},
			{allowedClientCIDR, port.InternalPort},
		} {
			err := oscli.AddSecurityRuleCIDR(sec.addr, strings.ToLower(port.Proto), sr, sec.port)
			if err != nil {
				log.DebugLog(log.DebugLevelMexos, "warning, error while adding security rule", "cidr", sec.addr, "securityrule", sr, "port", sec.port)
			}
		}
	}
	if len(mf.Spec.Ports) > 0 {
		err = AddNginxProxy(rootLB.Name, mf.Metadata.Name, kp.ipaddr, mf.Spec.Ports)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "cannot add nginx proxy", "name", mf.Metadata.Name, "ports", mf.Spec.Ports)
			return err
		}
	}
	log.DebugLog(log.DebugLevelMexos, "added nginx proxy", "name", mf.Metadata.Name, "ports", mf.Spec.Ports)
	return nil
}

func deleteSecurityRules(rootLB *MEXRootLB, mf *Manifest, kp *kubeParam) error {
	log.DebugLog(log.DebugLevelMexos, "delete spec ports", "ports", mf.Spec.Ports)
	err := DeleteNginxProxy(rootLB.Name, mf.Metadata.Name)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot delete nginx proxy", "name", mf.Metadata.Name, "rootlb", rootLB.Name, "error", err)
	}

	// TODO - implement the clean up of security rules
	return nil
}

func addDNSRecords(rootLB *MEXRootLB, mf *Manifest, kp *kubeParam) error {
	rootLBIPaddr, err := GetServerIPAddr(rootLB.PlatConf.Spec.ExternalNetwork, rootLB.Name)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot get rootlb IP address", "error", err)
		return fmt.Errorf("cannot deploy kubernetes app, cannot get rootlb IP")
	}
	cmd := fmt.Sprintf("%s kubectl get svc -o json", kp.kubeconfig)
	out, err := kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can not get list of services, %s, %v", out, err)
	}
	svcs := &svcItems{}
	err = json.Unmarshal([]byte(out), svcs)
	if err != nil {
		return fmt.Errorf("can not unmarshal svc json, %v", err)
	}
	if err := cloudflare.InitAPI(mexEnv["MEX_CF_USER"], mexEnv["MEX_CF_KEY"]); err != nil {
		return fmt.Errorf("cannot init cloudflare api, %v", err)
	}
	recs, err := cloudflare.GetDNSRecords(mf.Metadata.DNSZone)
	if err != nil {
		return fmt.Errorf("error getting dns records for %s, %v", mf.Metadata.DNSZone, err)
	}
	fqdnBase := uri2fqdn(mf.Spec.URI)
	for _, item := range svcs.Items {
		if !strings.HasPrefix(item.Metadata.Name, mf.Metadata.Name) {
			continue
		}
		cmd = fmt.Sprintf(`%s kubectl patch svc %s -p '{"spec":{"externalIPs":["%s"]}}'`, kp.kubeconfig, item.Metadata.Name, kp.ipaddr)
		out, err = kp.client.Output(cmd)
		if err != nil {
			return fmt.Errorf("error patching for kubernetes service, %s, %s, %v", cmd, out, err)
		}
		log.DebugLog(log.DebugLevelMexos, "patched externalIPs on service", "service", item.Metadata.Name, "externalIPs", kp.ipaddr)
		fqdn := cloudcommon.ServiceFQDN(item.Metadata.Name, fqdnBase)
		for _, rec := range recs {
			if rec.Type == "A" && rec.Name == fqdn {
				if err := cloudflare.DeleteDNSRecord(mf.Metadata.DNSZone, rec.ID); err != nil {
					return fmt.Errorf("cannot delete existing DNS record %v, %v", rec, err)
				}
				log.DebugLog(log.DebugLevelMexos, "deleted DNS record", "name", fqdn)
			}
		}
		if err := cloudflare.CreateDNSRecord(mf.Metadata.DNSZone, fqdn, "A", rootLBIPaddr, 1, false); err != nil {
			return fmt.Errorf("can't create DNS record for %s,%s, %v", fqdn, rootLBIPaddr, err)
		}
		log.DebugLog(log.DebugLevelMexos, "created DNS record", "name", fqdn, "addr", rootLBIPaddr)
	}
	return nil
}

func deleteDNSRecords(rootLB *MEXRootLB, mf *Manifest, kp *kubeParam) error {
	cmd := fmt.Sprintf("%s kubectl get svc -o json", kp.kubeconfig)
	out, err := kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can not get list of services, %s, %v", out, err)
	}
	svcs := &svcItems{}
	err = json.Unmarshal([]byte(out), svcs)
	if err != nil {
		return fmt.Errorf("can not unmarshal svc json, %v", err)
	}
	if cerr := cloudflare.InitAPI(mexEnv["MEX_CF_USER"], mexEnv["MEX_CF_KEY"]); cerr != nil {
		return fmt.Errorf("cannot init cloudflare api, %v", cerr)
	}
	recs, derr := cloudflare.GetDNSRecords(mf.Metadata.DNSZone)
	if derr != nil {
		return fmt.Errorf("error getting dns records for %s, %v", mf.Metadata.DNSZone, derr)
	}
	fqdnBase := uri2fqdn(mf.Spec.URI)
	//FIXME use k8s manifest file to delete the whole services and deployments
	for _, item := range svcs.Items {
		if !strings.HasPrefix(item.Metadata.Name, mf.Metadata.Name) {
			continue
		}
		cmd := fmt.Sprintf("%s kubectl delete service %s", kp.kubeconfig, item.Metadata.Name)
		out, err := kp.client.Output(cmd)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "error deleting kubernetes service", "name", item.Metadata.Name, "cmd", cmd, "out", out, "err", err)
		} else {
			log.DebugLog(log.DebugLevelMexos, "deleted service", "name", item.Metadata.Name)
		}
		fqdn := cloudcommon.ServiceFQDN(item.Metadata.Name, fqdnBase)
		for _, rec := range recs {
			if rec.Type == "A" && rec.Name == fqdn {
				if err := cloudflare.DeleteDNSRecord(mf.Metadata.DNSZone, rec.ID); err != nil {
					return fmt.Errorf("cannot delete existing DNS record %v, %v", rec, err)
				}
				log.DebugLog(log.DebugLevelMexos, "deleted DNS record", "name", fqdn)
			}
		}
	}
	cmd = fmt.Sprintf("%s kubectl delete deploy %s", kp.kubeconfig, mf.Metadata.Name+"-deployment")
	out, err = kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deleting kubernetes deployment, %s, %s, %v", cmd, out, err)
	}

	return nil
}
func validateCommon(mf *Manifest) error {
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name for the deployment")
	}
	if mf.Spec.Key == "" {
		return fmt.Errorf("empty cluster name")
	}
	if mf.Spec.Image == "" {
		return fmt.Errorf("empty image")
	}
	if mf.Spec.ProxyPath == "" {
		return fmt.Errorf("empty proxy path")
	}
	if mf.Metadata.DNSZone == "" {
		return fmt.Errorf("missing DNS zone, metadata %v", mf.Metadata)
	}
	return nil
}

func DeleteHelmAppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "delete kubernetes helm app", "mf", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if rootLB == nil {
		return fmt.Errorf("cannot delete helm app, rootLB is null")
	}
	if err = validateCommon(mf); err != nil {
		return err
	}
	kp, err := ValidateKubernetesParameters(rootLB, mf.Spec.Key)
	if err != nil {
		return err
	}
	// remove DNS entries
	if err = deleteDNSRecords(rootLB, mf, kp); err != nil {
		return err
	}
	// remove Security rules
	if err = deleteSecurityRules(rootLB, mf, kp); err != nil {
		return err
	}

	cmd := fmt.Sprintf("%s helm delete %s", kp.kubeconfig, mf.Metadata.Name)
	out, err := kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deleting helm chart, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "removed helm chart")
	return nil
}

func CreateHelmAppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "create kubernetes helm app", "mf", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if rootLB == nil {
		return fmt.Errorf("cannot create helm app, rootLB is null")
	}
	if err = validateCommon(mf); err != nil {
		return err
	}
	kp, err := ValidateKubernetesParameters(rootLB, mf.Spec.Key)
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "will launch app into cluster", "kubeconfig", kp.kubeconfig, "ipaddr", kp.ipaddr)

	cmd := fmt.Sprintf("%s helm init --wait", kp.kubeconfig)
	out, err := kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error initializing tiller for app, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "helm tiller initialized")

	cmd = fmt.Sprintf("%s helm install %s --name %s", kp.kubeconfig, mf.Spec.Image, mf.Metadata.Name)
	out, err = kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deploying helm chart, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "applied helm chart")
	// Add security rules
	if err = addSecurityRules(rootLB, mf, kp); err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot create security rules", "error", err)
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "add spec ports", "ports", mf.Spec.Ports)
	// Add DNS Zone
	if err = addDNSRecords(rootLB, mf, kp); err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot add DNS entries", "error", err)
		return err
	}
	return nil
}

//CreateKubernetesAppManifest instantiates a new kubernetes deployment
func CreateKubernetesAppManifest(mf *Manifest, kubeManifest string) error {
	log.DebugLog(log.DebugLevelMexos, "create kubernetes app", "mf", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if rootLB == nil {
		return fmt.Errorf("cannot create kubernetes app manifest, rootLB is null")
	}
	if err = validateCommon(mf); err != nil {
		return err
	}
	if mf.Spec.URI == "" { //XXX TODO register to the DNS registry for public IP app,controller needs to tell us which kind of app
		return fmt.Errorf("empty app URI")
	}
	kp, err := ValidateKubernetesParameters(rootLB, mf.Spec.Key)
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "will launch app into cluster", "kubeconfig", kp.kubeconfig, "ipaddr", kp.ipaddr)
	var cmd string
	if mexEnv["MEX_DOCKER_REG_PASS"] == "" {
		return fmt.Errorf("empty docker registry password environment variable")
	}
	//TODO: mexosagent should cache
	var out string
	cmd = fmt.Sprintf("echo %s > .docker-pass", mexEnv["MEX_DOCKER_REG_PASS"])
	out, err = kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't store docker password, %s, %v", out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "stored docker password")
	cmd = fmt.Sprintf("scp -o %s -o %s -i %s .docker-pass %s:", sshOpts[0], sshOpts[1], mexEnv["MEX_SSH_KEY"], kp.ipaddr)
	out, err = kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't copy docker password to k8s-master, %s, %v", out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "copied over docker password")
	cmd = fmt.Sprintf("ssh -o %s -o %s -i %s %s 'cat .docker-pass| docker login -u mobiledgex --password-stdin %s'", sshOpts[0], sshOpts[1], mexEnv["MEX_SSH_KEY"], kp.ipaddr, mexEnv["MEX_DOCKER_REGISTRY"])
	out, err = kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't docker login on k8s-master to %s, %s, %v", mexEnv["MEX_DOCKER_REGISTRY"], out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "docker login ok")
	cmd = fmt.Sprintf("%s kubectl create secret docker-registry mexregistrysecret --docker-server=%s --docker-username=mobiledgex --docker-password=%s --docker-email=docker@mobiledgex.com", kp.kubeconfig, mexEnv["MEX_DOCKER_REGISTRY"], mexEnv["MEX_DOCKER_REG_PASS"])
	out, err = kp.client.Output(cmd)
	if err != nil {
		if strings.Contains(out, "AlreadyExists") {
			log.DebugLog(log.DebugLevelMexos, "secret already exists")
		} else {
			return fmt.Errorf("error creating mexregistrysecret, %s, %s, %v", cmd, out, err)
		}
	}
	log.DebugLog(log.DebugLevelMexos, "created mexregistrysecret docker secret")
	cmd = fmt.Sprintf("cat <<'EOF'> %s.yaml \n%s\nEOF", mf.Metadata.Name, kubeManifest)
	out, err = kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error writing KubeManifest, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "wrote Kube Manifest file")
	cmd = fmt.Sprintf("%s kubectl create -f %s.yaml", kp.kubeconfig, mf.Metadata.Name)
	out, err = kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deploying kubernetes app, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "applied kubernetes manifest")
	// Add security rules
	if err = addSecurityRules(rootLB, mf, kp); err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot create security rules", "error", err)
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "add spec ports", "ports", mf.Spec.Ports)
	// Add DNS Zone
	if err = addDNSRecords(rootLB, mf, kp); err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot add DNS entries", "error", err)
		return err
	}
	return nil
}

type kubeParam struct {
	kubeconfig string
	client     ssh.Client
	ipaddr     string
}

//ValidateKubernetesParameters checks the kubernetes parameters and kubeconfig settings
func ValidateKubernetesParameters(rootLB *MEXRootLB, clustName string) (*kubeParam, error) {
	log.DebugLog(log.DebugLevelMexos, "validate kubernetes parameters rootLB", "rootLB", rootLB, "cluster", clustName)
	if rootLB == nil {
		return nil, fmt.Errorf("cannot validate kubernetes parameters, rootLB is null")
	}
	if rootLB.PlatConf == nil {
		return nil, fmt.Errorf("validate kubernetes parameters, missing platform config")
	}
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return nil, fmt.Errorf("validate kubernetes parameters, missing external network in platform config")
	}
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return nil, err
	}
	name, err := FindClusterWithKey(clustName)
	if err != nil {
		return nil, fmt.Errorf("can't find cluster with key %s, %v", clustName, err)
	}
	ipaddr, err := FindNodeIP(name)
	if err != nil {
		return nil, err
	}
	kubeconfig := fmt.Sprintf("KUBECONFIG=%s.kubeconfig", name[strings.LastIndex(name, "-")+1:])
	return &kubeParam{kubeconfig, client, ipaddr}, nil
}

//KubernetesApplyManifest does `apply` on the manifest yaml
func KubernetesApplyManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "apply kubernetes manifest", "mf", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if rootLB == nil {
		return fmt.Errorf("cannot apply kubernetes manifest, rootLB is null")
	}
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name")
	}
	kp, err := ValidateKubernetesParameters(rootLB, mf.Spec.Key)
	if err != nil {
		return err
	}
	kubeManifest := mf.Config.ConfigDetail.Manifest
	cmd := fmt.Sprintf("cat <<'EOF'> %s \n%s\nEOF", mf.Metadata.Name, kubeManifest)
	out, err := kp.client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error writing deployment, %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "wrote deployment file")
	cmd = fmt.Sprintf("%s kubectl apply -f %s", kp.kubeconfig, mf.Metadata.Name)
	out, err = kp.client.Output(cmd)
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
	if rootLB == nil {
		return fmt.Errorf("cannot set kubeconfig map values, rootLB is null")
	}
	kp, err := ValidateKubernetesParameters(rootLB, clustername)
	if err != nil {
		return err
	}
	//TODO support namespace
	cmd := fmt.Sprintf("%s kubectl create configmap %s ", kp.kubeconfig, configname)
	for _, kv := range keyvalues {
		items := strings.Split(kv, "=")
		if len(items) != 2 {
			return fmt.Errorf("malformed key=value pair, %s", kv)
		}
		cmd = cmd + " --from-literal=" + kv
	}
	out, err := kp.client.Output(cmd)
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
	if rootLB == nil {
		return "", fmt.Errorf("cannot get kubeconfigmap yaml, rootLB is null")
	}
	kp, err := ValidateKubernetesParameters(rootLB, clustername)
	if err != nil {
		return "", err
	}
	//TODO support namespace
	cmd := fmt.Sprintf("%s kubectl get configmap %s -o yaml", kp.kubeconfig, configname)
	out, err := kp.client.Output(cmd)
	if err != nil {
		return "", fmt.Errorf("error getting configmap yaml, %s, %s, %v", cmd, out, err)
	}
	return out, nil
}

func AddNginxProxy(rootLBName, name, ipaddr string, ports []PortDetail) error {
	log.DebugLog(log.DebugLevelMexos, "add nginx proxy", "name", name, "ports", ports)

	request := gorequest.New()
	npURI := fmt.Sprintf("http://%s:%s/v1/nginx", rootLBName, mexEnv["MEX_AGENT_PORT"])
	pl, err := FormNginxProxyRequest(ports, ipaddr, name)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot form nginx proxy request")
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "nginx proxy add request post", "request", *pl)
	resp, body, errs := request.Post(npURI).Set("Content-Type", "application/json").Send(pl).End()
	if errs != nil {
		return fmt.Errorf("error, can't request nginx proxy add, %v", errs)
	}
	if strings.Contains(body, "OK") {
		log.DebugLog(log.DebugLevelMexos, "added nginx proxy OK")
		return nil
	}
	log.DebugLog(log.DebugLevelMexos, "warning, error while adding nginx proxy", "resp", resp, "body", body)
	return fmt.Errorf("cannot add nginx proxy, resp %v", resp)
}

func FormNginxProxyRequest(ports []PortDetail, ipaddr string, name string) (*string, error) {
	portstrs := []string{}
	for _, p := range ports {
		switch p.MexProto {
		case "LProtoHTTP":
			portstrs = append(portstrs,
				fmt.Sprintf(`{"mexproto":"%s", "external": "%d", "internal": "%d", "origin":"%s:%d", "path":"/%s"}`,
					p.MexProto, p.PublicPort, p.InternalPort, ipaddr, p.InternalPort, p.PublicPath))
		case "LProtoTCP":
			portstrs = append(portstrs,
				fmt.Sprintf(`{"mexproto":"%s", "external": "%d", "origin": "%s:%d"}`,
					p.MexProto, p.PublicPort, ipaddr, p.InternalPort))
		case "LProtoUDP":
			portstrs = append(portstrs,
				fmt.Sprintf(`{"mexproto":"%s", "external": "%d", "origin": "%s:%d"}`,
					p.MexProto, p.PublicPort, ipaddr, p.InternalPort))
		default:
			log.DebugLog(log.DebugLevelMexos, "invalid mexproto", "port", p)
		}
	}
	portspec := ""
	for i, ps := range portstrs {
		if i == 0 {
			portspec += ps
		} else {
			portspec += "," + ps
		}

	}
	pl := fmt.Sprintf(`{ "message":"add", "name": "%s" , "ports": %s }`, name, "["+portspec+"]")
	return &pl, nil
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
	if rootLB == nil {
		return 0, fmt.Errorf("cannot kubectl proxy, rootLB is null")
	}
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
	err = oscli.AddSecurityRuleCIDR(GetAllowedClientCIDR(), "tcp", oscli.GetDefaultSecurityRule(), maxPort)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "warning, error while adding external ingress security rule for kubeproxy", "error", err, "port", maxPort)
	}
	log.DebugLog(log.DebugLevelMexos, "added external ingress security rule for kubeproxy", "port", maxPort)
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

func DeleteKubernetesAppManifest(mf *Manifest, kubeManifest string) error {
	log.DebugLog(log.DebugLevelMexos, "delete kubernetes app", "mf", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if rootLB == nil {
		return fmt.Errorf("cannot remove kubernetes app manifest, rootLB is null")
	}
	if mf.Spec.URI == "" { //XXX TODO register to the DNS registry for public IP app,controller needs to tell us which kind of app
		return fmt.Errorf("empty app URI")
	}
	//TODO: support other URI: file://, nfs://, ftp://, git://, or embedded as base64 string
	//if !strings.Contains(mf.Spec.Flavor, "kubernetes") {
	//	return fmt.Errorf("unsupported kubernetes flavor %s", mf.Spec.Flavor)
	//}
	if err = validateCommon(mf); err != nil {
		return err
	}
	kp, err := ValidateKubernetesParameters(rootLB, mf.Spec.Key)
	if err != nil {
		return err
	}
	// Clean up security rules and nginx proxy
	if err = deleteSecurityRules(rootLB, mf, kp); err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot clean up security rules", "name", mf.Metadata.Name, "rootlb", rootLB.Name, "error", err)
		return err
	}
	// Clean up DNS entries
	if err = deleteDNSRecords(rootLB, mf, kp); err != nil {
		log.DebugLog(log.DebugLevelMexos, "cannot clean up DNS entries", "name", mf.Metadata.Name, "rootlb", rootLB.Name, "error", err)
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "deleted deployment", "name", mf.Metadata.Name)
	return nil
}

//DeletePathReverseProxy Deletes a new route to origin on the reverse proxy
func DeletePathReverseProxy(rootLBName, path, origin string) []error {
	log.DebugLog(log.DebugLevelMexos, "delete path reverse proxy", "path", path, "origin", origin)
	//TODO
	return nil
}

func DeleteNginxProxy(rootLBName, name string) error {
	log.DebugLog(log.DebugLevelMexos, "add nginx proxy", "name", name)
	request := gorequest.New()
	npURI := fmt.Sprintf("http://%s:%s/v1/nginx", rootLBName, mexEnv["MEX_AGENT_PORT"])
	pl := fmt.Sprintf(`{"message":"delete","name":"%s"}`, name)
	log.DebugLog(log.DebugLevelMexos, "nginx proxy add request post", "request", pl)
	resp, body, errs := request.Post(npURI).Set("Content-Type", "application/json").Send(pl).End()
	if errs != nil {
		return fmt.Errorf("error, can't request nginx proxy delete, %v", errs)
	}
	if strings.Contains(body, "OK") {
		log.DebugLog(log.DebugLevelMexos, "deleted nginx proxy OK")
		return nil
	}
	log.DebugLog(log.DebugLevelMexos, "error while deleting nginx proxy", "resp", resp, "body", body)
	return fmt.Errorf("cannot delete nginx proxy, resp %v", resp)
}

//CreateQCOW2AppManifest creates qcow2 app
func CreateQCOW2AppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "create qcow2 vm based app", "mf", mf)
	//TODO: support other URI: file://, nfs://, ftp://, git://, or embedded as base64 string
	if !strings.HasPrefix(mf.Spec.Image, "http://") &&
		!strings.HasPrefix(mf.Spec.Image, "https://") {
		return fmt.Errorf("unsupported qcow2 image spec %s", mf.Spec.Image)
	}
	if !strings.Contains(mf.Spec.Flavor, "qcow2") {
		return fmt.Errorf("unsupported qcow2 flavor %s", mf.Spec.Flavor)
	}
	if err := validateCommon(mf); err != nil {
		return err
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

func DeleteQCOW2AppManifest(mf *Manifest) error {
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

func LookupDNS(name string) (string, error) {
	ips, err := net.LookupIP(name)
	if err != nil {
		return "", fmt.Errorf("DNS lookup error, %s, %v", name, err)
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no DNS records, %s", name)
	}
	for _, ip := range ips {
		return ip.String(), nil //XXX return only first one
	}
	return "", fmt.Errorf("no IP in DNS record for %s", name)
}

var dnsRegisterRetryDelay time.Duration = 3 * time.Second

func WaitforDNSRegistration(name string) error {
	var ipa string
	var err error

	for i := 0; i < 100; i++ {
		ipa, err = LookupDNS(name)
		if err == nil && ipa != "" {
			return nil
		}
		time.Sleep(dnsRegisterRetryDelay)
	}
	log.DebugLog(log.DebugLevelMexos, "DNS lookup timed out", "name", name)
	return fmt.Errorf("error, timed out while looking up DNS for name %s", name)
}
