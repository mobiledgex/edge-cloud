package crmutil

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/rs/xid"

	"github.com/mobiledgex/edge-cloud-infra/openstack-prov/oscliapi"
	"github.com/mobiledgex/edge-cloud-infra/openstack-tenant/agent/cloudflare"
	//"github.com/mobiledgex/edge-cloud/edgeproto"

	valid "github.com/asaskevich/govalidator"
	"github.com/codeskyblue/go-sh"

	"github.com/nanobox-io/golang-ssh"
	"github.com/parnurzeal/gorequest"
	log "gitlab.com/bobbae/logrus"
	//"github.com/fsouza/go-dockerclient"
)

var mexEnv = map[string]string{
	"MEX_DIR":             os.Getenv("HOME") + "/.mobiledgex",
	"MEX_CF_KEY":          os.Getenv("MEX_CF_KEY"),
	"MEX_CF_USER":         os.Getenv("MEX_CF_USER"),
	"MEX_DOCKER_REG_PASS": os.Getenv("MEX_DOCKER_REG_PASS"),
	"MEX_AGENT_PORT":      "18889",
	"MEX_SSH_KEY":         "id_rsa_mobiledgex",
	"MEX_K8S_MASTER":      "mex-k8s-master",
	"MEX_K8S_NODE":        "mex-k8s-node",
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
	log.Debugln("mexEnv", mexEnv)
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
	log.Debugln("new rootLB", rootLBName)
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
	log.Debugln("new rootLB with manifest", mf)
	rootLB, err := NewRootLB(mf.Spec.RootLB)
	if err != nil {
		return nil, err
	}
	rootLB.PlatConf = mf
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
//   So if an app needs an esternal IP, we can't figure out if that is the case.
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
		log.Debugln("invalid mex env")
		IsValidMEXOSEnv = false
		return false
	}
	//XXX do more validation on internal env vars
	IsValidMEXOSEnv = true
	log.Debugln("valid mex env")
	return IsValidMEXOSEnv
}

//AddFlavor adds a new flavor to be kept track of
func AddFlavor(flavor string) error {
	log.Debugln("add flavor", flavor)
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
	log.Debugln("get cluster flavor", flavor)
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

//CreateCluster creates a cluster of nodes. It can take a while, so call from a goroutine.
func mexCreateClusterKubernetes(mf *Manifest) (*string, error) {
	log.Debugln("create kubernetes cluster", mf)
	//XXX TODO update rootLB.Env[] with manifest specified values
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return nil, err
	}
	if mf.Spec.Flavor == "" {
		return nil, fmt.Errorf("empty flavor")
	}
	if err := ValidateFlavor(mf.Spec.Flavor); err != nil {
		return nil, fmt.Errorf("invalid flavor")
	}
	if !strings.Contains(mf.Spec.Flavor, "kubernetes") {
		return nil, fmt.Errorf("unsupported kubernetes flavor", mf.Spec.Flavor)
	}
	cf, err := GetClusterFlavor(mf.Spec.Flavor)
	if err != nil {
		return nil, err
	}
	if cf.NumNodes < 1 {
		return nil, fmt.Errorf("invalid flavor profile, %v", cf)
	}
	//TODO more than one networks
	if mf.Spec.Networks[0].Kind == "" {
		return nil, fmt.Errorf("empty network spec")
	}
	if mf.Spec.Networks[0].Kind != "priv-subnet" {
		return nil, fmt.Errorf("unsupported netSpec kind")
		// XXX for now
	}
	//TODO allow more net types
	//TODO validate CIDR, etc.
	if mf.Metadata.Tags == "" {
		return nil, fmt.Errorf("empty tag")
	}
	if mf.Metadata.Tenant == "" {
		return nil, fmt.Errorf("empty tenant")
	}
	err = ValidateTenant(mf.Metadata.Tenant)
	if err != nil {
		return nil, fmt.Errorf("can't validate tenant, %v", err)
	}
	err = ValidateTags(mf.Metadata.Tags)
	if err != nil {
		return nil, fmt.Errorf("invalid tag, %v", err)
	}
	//XXX should check for quota, permissions, access control, etc. here
	guid := xid.New().String()
	kvmname := fmt.Sprintf("%s-1-%s-%s", mexEnv["MEX_K8S_MASTER"], mf.Metadata.Name, guid)
	err = oscli.CreateMEXKVM(kvmname,
		"k8s-master",
		mf.Spec.Networks[0].Kind+","+mf.Spec.Networks[0].Name+","+mf.Spec.Networks[0].CIDR,
		mf.Metadata.Tags,
		mf.Metadata.Tenant,
		1,
	)
	if err != nil {
		return nil, fmt.Errorf("can't create k8s master, %v", err)
	}
	for i := 1; i <= cf.NumNodes; i++ {
		//construct node name
		kvmnodename := fmt.Sprintf("%s-%d-%s-%s", mexEnv["MEX_K8S_NODE"], i, mf.Metadata.Name, guid)
		err = oscli.CreateMEXKVM(kvmnodename,
			"k8s-node",
			mf.Spec.Networks[0].Kind+","+mf.Spec.Networks[0].Name+","+mf.Spec.Networks[0].CIDR,
			mf.Metadata.Tags,
			mf.Metadata.Tenant,
			i)
		if err != nil {
			return nil, fmt.Errorf("can't create k8s node, %v", err)
		}
	}
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return nil, fmt.Errorf("missing external network in platform config")
	}
	if err = LBAddRoute(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, kvmname); err != nil {
		return nil, err
	}
	if err = oscli.SetServerProperty(kvmname, "mex-flavor="+mf.Spec.Flavor); err != nil {
		return nil, err
	}
	ready := false
	for i := 0; i < 10; i++ {
		ready, err = IsClusterReady(mf, rootLB)
		if err != nil {
			return nil, err
		}
		if ready {
			log.Debugln("kubernetes cluster ready")
			break
		}
		log.Debugln("wating for kubernetes cluster to be ready...")
		time.Sleep(30 * time.Second)
	}
	if !ready {
		return nil, fmt.Errorf("cluster not ready (yet)")
	}
	return &guid, nil
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
		return fmt.Errorf("can't add route to rootLB, %s, %s, %v", cmd, out, err)
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
		log.Warningf("can't delete route at rootLB, %s, %s, %v", cmd, out, err)
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

//DeleteCluster
func mexDeleteClusterKubernetes(mf *Manifest) error {
	log.Debugln("deleting kubernetes cluster", mf)
	if mf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("missing external network in manifest")
	}
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	name := mf.Metadata.Name
	srvs, err := oscli.ListServers()
	if err != nil {
		return err
	}
	log.Debugln("looking for server", name, "in", srvs)
	force := strings.Contains(mf.Spec.Flags, "force")
	for _, s := range srvs {
		if strings.Contains(s.Name, name) {
			if strings.Contains(s.Name, mexEnv["MEX_K8S_MASTER"]) {
				err = LBRemoveRoute(rootLB.Name, mf.Spec.ExternalNetwork, s.Name)
				if err != nil {
					if !force {
						err = fmt.Errorf("failed to remove route for %s, %v", s.Name, err)
						log.Debugln(err)
						return err
					}
					log.Warningln("failed to remove route for", s.Name, err)
					log.Warningln("forced to continue")
				}
			}
			log.Debugln("delete server", s.Name)
			err = oscli.DeleteServer(s.Name)
			if err != nil {
				if !force {
					log.Debugln(err)
					return err
				}
				log.Warningln("error while deleting server", s.Name, err)
				log.Warningln("forced to continue")
			}
		}
	}
	sns, err := oscli.ListSubnets("")
	if err != nil {
		log.Debugln(err)
		return err
	}
	rn := oscli.GetMEXExternalRouter() //XXX for now
	log.Debugln("removing router", rn, "subnets", sns)
	for _, s := range sns {
		if strings.Contains(s.Name, name) {
			log.Debugln("removing router from subnet", rn, s.Name)
			err := oscli.RemoveRouterSubnet(rn, s.Name)
			if err != nil {
				log.Debugln(err)
				//return err
			}
			err = oscli.DeleteSubnet(s.Name)
			if err != nil {
				log.Debugln("error while deleting subnet", err)
				return err
			}
			break
		}
	}
	//XXX tell agent to remove the route
	return nil
}

//EnableRootLB creates a seed presence node in cloudlet that also becomes first Agent node.
//  It also sets up first basic network router and subnet, ready for running first MEX agent.
func EnableRootLB(mf *Manifest, rootLB *MEXRootLB) error {
	log.Debugln("enable rootlb", rootLB)
	if mf.Spec.ExternalNetwork == "" {
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
		if strings.Contains(s.Name, "mex-lb-") && strings.Contains(s.Name, mf.Metadata.Zone) {
			log.Debugln("found", s)
			found++
		}
	}
	if found == 0 {
		netspec := fmt.Sprintf("external-ip,%s,%s", mf.Spec.ExternalNetwork, mf.Spec.InternalCIDR)
		if strings.Contains(mf.Spec.Options, "dhcp") {
			netspec = netspec + ",dhcp"
		}
		log.Debugln("creating agent node kvm", mf, netspec)
		err := oscli.CreateMEXKVM(rootLB.Name,
			"mex-agent-node", //important, don't change
			netspec,
			mf.Metadata.Tags,
			mf.Metadata.Tenant,
			1)
		if err != nil {
			log.Debugln("error while creating mex kvm", err)
			return err
		}
		log.Debugln("created kvm instance", rootLB.Name)
	} else {
		log.Debugln("using existing kvm instance %s", rootLB.Name)
	}
	return nil
}

//XXX allow creating more than one LB

//GetServerIPAddr gets the server IP
func GetServerIPAddr(networkName, serverName string) (string, error) {
	log.Debugln("get server ip addr", networkName, serverName)
	//sd, err := oscli.GetServerDetails(rootLB)
	sd, err := oscli.GetServerDetails(serverName)
	if err != nil {
		return "", err
	}
	its := strings.Split(sd.Addresses, "=")
	if len(its) != 2 {
		return "", fmt.Errorf("GetServerIPAddr: can't parse server detail addresses, %v, %v", sd, err)
	}
	if its[0] != networkName {
		return "", fmt.Errorf("invalid network name in server detail address, %s", sd.Addresses)
	}
	addr := its[1]
	log.Debugln("got ip addr", addr)
	return addr, nil
}

//CopySSHCredential copies over the ssh credential for mex to LB
func CopySSHCredential(serverName, networkName, userName string) error {
	log.Debugln("copying ssh credentials", serverName, networkName, userName)
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
	auth := ssh.Auth{Keys: []string{mexEnv["MEX_DIR"] + "/id_rsa_mobiledgex"}}
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
	log.Debugln("wait for rootlb", rootLB)
	if mf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("waiting for lb, missing external network in manifest")
	}
	client, err := GetSSHClient(rootLB.Name, mf.Spec.ExternalNetwork, "root")
	if err != nil {
		return err
	}
	running := false
	for i := 0; i < 10; i++ {
		_, err := client.Output("grep done /tmp/mobiledgex.log") //XXX beware of use of word done
		if err == nil {
			log.Debugln("rootlb running", rootLB)
			running = true
			if err := CopySSHCredential(rootLB.Name, mf.Spec.ExternalNetwork, "root"); err != nil {
				return fmt.Errorf("can't copy ssh credential to RootLB, %v", err)
			}
			break
		}
		log.Debugln("wating for rootlb...")
		time.Sleep(30 * time.Second)
	}
	if !running {
		return fmt.Errorf("while creating cluster, timeout waiting for RootLB")
	}
	//XXX just ssh docker commands instead since there is no real benefit for now
	//err = InitDockerMachine(mf,rootLB,addr)
	return nil
}

//InitDockerMachine preps docker-machine for use with docker command
func InitDockerMachine(rootLB, addr string) error {
	log.Debugln("init docker machine", rootLB)
	home := os.Getenv("HOME")
	_, err := sh.Command("docker-machine", "create", "-d", "generic", "--generic-ip-address", addr, "--generic-ssh-key", mexEnv["MEX_DIR"]+"/id_rsa_mobiledgex", "--generic-ssh-user", "bob", rootLB).Output()
	if err != nil {
		return err
	}
	if err := os.Setenv("DOCKER_TLS_VERIFY", "1"); err != nil {
		log.Fatal(err)
	}
	if err := os.Setenv("DOCKER_HOST", "tcp://"+addr+":2376"); err != nil {
		log.Fatal(err)
	}
	if err := os.Setenv("DOCKER_CERT_PATH", home+"/.docker/machine/machines/"+rootLB); err != nil {
		log.Fatal(err)
	}
	if err := os.Setenv("DOCKER_MACHINE_NAME", rootLB); err != nil {
		log.Fatal(err)
	}
	return nil
}

//RunMEXAgentManifest runs the MEX agent on the RootLB. It first registers FQDN to cloudflare domain registry if not already registered.
//   It then obtains certficiates from Letsencrypt, if not done yet.  Then it runs the docker instance of MEX agent
//   on the RootLB. It can be told to manually pull image from docker repository.  This allows upgrading with new image.
//   It uses MEX private docker repository.  If an instance is running already, we don't start another one.
func RunMEXAgentManifest(mf *Manifest, pull bool) error {
	log.Debugln("run mex agent", mf)
	//XXX TODO update rootLB.Env with manifest
	fqdn := mf.Spec.RootLB
	//fqdn is that of the machine/kvm-instance running the agent
	if !valid.IsDNSName(fqdn) {
		return fmt.Errorf("fqdn %s is not valid", fqdn)
	}
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
	if mf.Spec.DockerRegistry == "" {
		return fmt.Errorf("missing docker registry")
	}
	log.Debugln("record platform config", mf)
	rootLB.PlatConf = mf
	err = EnableRootLB(mf, rootLB)
	if err != nil {
		log.Debugln("can't enable agent", rootLB.Name)
		return fmt.Errorf("Failed to enable root LB %v", err)
	}
	err = WaitForRootLB(mf, rootLB)
	if err != nil {
		log.Debugln("timeout wating for agent to run", rootLB.Name)
		return fmt.Errorf("Error waiting for rootLB %v", err)
	}
	client, err := GetSSHClient(rootLB.Name, mf.Spec.ExternalNetwork, "root")
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("docker ps |grep %s", mf.Spec.Agent.Image)
	_, err = client.Output(cmd)
	if err == nil {
		//agent docker instance exists
		//XXX check better
		log.Debugln("agent docker instance already running")
		return nil
	}
	if err = ActivateFQDNA(mf, rootLB, rootLB.Name); err != nil {
		return err
	}
	log.Debugln("FQDN A record activated", rootLB.Name)
	err = AcquireCertificates(mf, rootLB, rootLB.Name) //fqdn name may be different than rootLB.Name
	if err != nil {
		return fmt.Errorf("can't acquire certificate for %s, %v", rootLB.Name, err)
	}
	log.Debugln("acquired certificates from letsencrypt", rootLB.Name)
	if mexEnv["MEX_DOCKER_REG_PASS"] == "" {
		return fmt.Errorf("empty docker registry pass env var")
	}
	cmd = fmt.Sprintf("echo %s > .docker-pass", mexEnv["MEX_DOCKER_REG_PASS"])
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't store docker pass, %s, %v", out, err)
	}
	log.Debugln("seeded docker registry password")
	dockerinstanceName := fmt.Sprintf("%s-%s", mf.Metadata.Name, rootLB.Name)
	if pull { //XXX needs to be in manifest
		log.Debugln("force pull from docker registry")
		cmd = fmt.Sprintf("cat .docker-pass| docker login -u mobiledgex --password-stdin %s; docker pull %s; docker run -d --rm --name %s --net=host -v `pwd`:/var/www/.cache -v /etc/ssl/certs:/etc/ssl/certs %s -debug", mf.Spec.Agent.Image, mf.Spec.DockerRegistry, dockerinstanceName, mf.Spec.Agent.Image)
	} else {
		cmd = fmt.Sprintf("cat .docker-pass| docker login -u mobiledgex --password-stdin %s; docker run -d --rm --name %s --net=host -v `pwd`:/var/www/.cache -v /etc/ssl/certs:/etc/ssl/certs %s -debug", mf.Spec.DockerRegistry, dockerinstanceName, mf.Spec.Agent.Image)
	}
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error running dockerized agent on RootLB %s, %s, %s, %v", rootLB.Name, cmd, out, err)
	}
	log.Debugln("running dockerized agent")
	return nil
}

//UpdateMEXAgentManifest upgrades the mex agent
func UpdateMEXAgentManifest(mf *Manifest) error {
	log.Debugln("update mex agent", mf)
	err := RemoveMEXAgentManifest(mf)
	if err != nil {
		return err
	}
	// Force pulling a potentially newer docker image
	return RunMEXAgentManifest(mf, true)
}

//RemoveMEXAgentManifest deletes mex agent docker instance
func RemoveMEXAgentManifest(mf *Manifest) error {
	log.Debugln("deleting mex agent", mf)
	//XXX delete server!!!
	err := oscli.DeleteServer(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if mf.Metadata.DNSZone == "" {
		return fmt.Errorf("missing dns zone in manifest")
	}
	recs, err := cloudflare.GetDNSRecords(mf.Metadata.DNSZone)
	fqdn := mf.Spec.RootLB
	if err != nil {
		return fmt.Errorf("can not get dns records for %s, %v", fqdn, err)
	}
	for _, rec := range recs {
		if rec.Type == "A" && rec.Name == fqdn {
			err = cloudflare.DeleteDNSRecord(fqdn, rec.ID)
			if err != nil {
				return fmt.Errorf("cannot delete dns record id %s zone %s, %v", rec.ID, fqdn, err)
			}
		}
	}
	//TODO remove mex-k8s  internal nets and router
	return nil
}

//CheckCredentialsCF checks for Cloudflare
func CheckCredentialsCF() error {
	log.Debugln("check for cloudflare credentials")
	for _, envname := range []string{"MEX_CF_KEY", "MEX_CF_USER"} {
		if _, ok := mexEnv[envname]; !ok {
			return fmt.Errorf("no env var for %s", envname)
		}
	}
	return nil
}

//AcquireCertificates obtains certficates from Letsencrypt over ACME. It should be used carefully. The API calls have quota.
func AcquireCertificates(mf *Manifest, rootLB *MEXRootLB, fqdn string) error {
	log.Debugln("acquiring certificates for fqdn", fqdn)
	if mf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("acquire certificate, missing external network in manifest")
	}
	if err := CheckCredentialsCF(); err != nil {
		return err
	}
	client, err := GetSSHClient(fqdn, mf.Spec.ExternalNetwork, "root")
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
		log.Debugln("waiting for letsencrypt...")
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
	return nil
}

//ActivateFQDNA updates and ensures FQDN is registered properly
func ActivateFQDNA(mf *Manifest, rootLB *MEXRootLB, fqdn string) error {
	if mf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("activate fqdn A record, missing external network in manifest")
	}
	if err := CheckCredentialsCF(); err != nil {
		return err
	}
	if err := cloudflare.InitAPI(mexEnv["MEX_CF_USER"], mexEnv["MEX_CF_KEY"]); err != nil {
		return fmt.Errorf("cannot init cloudflare api, %v", err)
	}
	log.Debugln("getting dns record for zone", mf.Metadata.DNSZone)
	dr, err := cloudflare.GetDNSRecords(mf.Metadata.DNSZone)
	if err != nil {
		return fmt.Errorf("cannot get dns records for %s, %v", fqdn, err)
	}
	addr, err := GetServerIPAddr(mf.Spec.ExternalNetwork, fqdn)
	for _, d := range dr {
		if d.Type == "A" && d.Name == fqdn {
			if d.Content == addr {
				return nil
			}
			log.Warningf("cloudflare A record has different address %v, not %s, it will be overwritten", d, addr)
			if err = cloudflare.DeleteDNSRecord(mf.Metadata.DNSZone, d.ID); err != nil {
				return fmt.Errorf("can't delete DNS record for %s, %v", fqdn, err)
			}
			break
		}
	}
	if err != nil {
		log.Debugln("error while talking to cloudflare", err)
		return err
	}
	if err := cloudflare.CreateDNSRecord(mf.Metadata.DNSZone, fqdn, "A", addr, 1, false); err != nil {
		return fmt.Errorf("can't create DNS record for %s, %v", fqdn, err)
	}
	log.Debugln("wating for cloudflare...")
	//once successfully inserted the A record will take a bit of time, but not too long due to fast cloudflare anycast
	time.Sleep(10 * time.Second) //XXX just one time wait
	return nil
}

//IsClusterReady checks to see if cluster is read, i.e. rootLB is running and active
func IsClusterReady(mf *Manifest, rootLB *MEXRootLB) (bool, error) {
	log.Debugln("checking if cluster is ready", rootLB)
	cf, err := GetClusterFlavor(rootLB.PlatConf.Spec.Flavor)
	if err != nil {
		return false, err
	}
	if cf.NumNodes < 1 {
		return false, fmt.Errorf("invalid flavor profile, %v", cf)
	}
	name, err := FindClusterWithKey(mf.Metadata.Name)
	if err != nil {
		return false, fmt.Errorf("can't find cluster with key %s, %v", mf.Metadata.Name, err)
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
	log.Debugln("checking host", ipaddr, "for kubernetes nodes")
	cmd := fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %s bob@%s kubectl get nodes | grep Ready |grep -v NotReady| wc -l", mexEnv["MEX_SSH_KEY"], ipaddr)
	out, err := client.Output(cmd)
	if err != nil {
		return false, fmt.Errorf("kubectl fail on %s, %s, %v", name, out, err)
	}
	totalNodes := cf.NumNodes + cf.NumMasterNodes
	tn := fmt.Sprintf("%d", totalNodes)
	if !strings.Contains(out, tn) {
		log.Debugln("kubernetes cluster not ready, %s", out)
		return false, nil
	}
	log.Debugln("cluster ready", out)
	kcpath := mexEnv["MEX_DIR"] + "/" + name + ".kubeconfig"
	if _, err := os.Stat(kcpath); err == nil {
		return true, nil
	}
	if err := CopyKubeConfig(rootLB, name); err != nil {
		return false, fmt.Errorf("kubeconfig copy failed, %v", err)
	}
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
	log.Debugln("copying kubeconfig", name)
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
	err = StartKubectlProxy(rootLB, kconfname)
	if err != nil {
		return err
	}
	return ProcessKubeconfig(rootLB, name, []byte(out))
}

//ProcessKubeconfig validates kubeconfig and saves it and creates a copy for proxy access
func ProcessKubeconfig(rootLB *MEXRootLB, name string, dat []byte) error {
	log.Debugln("process kubeconfig file", name)
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
	log.Debugln("wrote", fullpath)
	kc.Clusters[0].Cluster.Server = "http://" + rootLB.Name + ":8001" //XXX allow for more ports and multiple instances XXX
	dat, err = yaml.Marshal(kc)
	if err != nil {
		return fmt.Errorf("can't marshal kubeconfig proxy edit %s, %v", name, err)
	}
	fullpath = mexEnv["MEX_DIR"] + "/" + name + ".kubeconfig-proxy"
	err = ioutil.WriteFile(fullpath, dat, 0666)
	if err != nil {
		return fmt.Errorf("can't write kubeconfig proxy %s, %v", name, err)
	}
	log.Debugln("wrote", fullpath)
	return nil
}

//FindNodeIP finds IP for the given node
func FindNodeIP(name string) (string, error) {
	log.Debugln("find node ip", name)
	if name == "" {
		return "", fmt.Errorf("empty name")
	}
	srvs, err := oscli.ListServers()
	if err != nil {
		return "", err
	}
	for _, s := range srvs {
		if s.Status == activeService && strings.Contains(s.Name, name) {
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
	log.Debugln("find cluster with key", key)
	if key == "" {
		return "", fmt.Errorf("empty key")
	}
	srvs, err := oscli.ListServers()
	if err != nil {
		return "", err
	}
	for _, s := range srvs {
		if s.Status == activeService && strings.Contains(s.Name, key) && strings.Contains(s.Name, mexEnv["MEX_K8S_MASTER"]) {
			return s.Name, nil
		}
	}
	return "", fmt.Errorf("key %s not found", key)
}

//CreateKubernetesAppManifest instantiates a new kubernetes deployment
func CreateKubernetesAppManifest(mf *Manifest) error {
	log.Debugln("create kubernetes app", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name")
	}
	if mf.Spec.Kubernetes == "" {
		return fmt.Errorf("missing kubernetes spec")
	}
	//TODO: support other URI: file://, nfs://, ftp://, git://, or embedded as base64 string
	if !strings.HasPrefix(mf.Spec.Kubernetes, "http://") &&
		!strings.HasPrefix(mf.Spec.Kubernetes, "https://") {
		return fmt.Errorf("unsupported kubernetes spec %s", mf.Spec.Kubernetes)
	}
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name")
	}
	kubeconfig, client, ipaddr, err := ValidateKubernetesParameters(rootLB, mf.Metadata.Name)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("%s kubectl apply -f %s", kubeconfig, mf.Spec.Kubernetes)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deploying kubernetes app, %s, %s, %v", cmd, out, err)
	}
	//TODO: use service manifest, or require a single manifest with both deployment and service. Use namespaces.
	cmd = fmt.Sprintf("%s kubectl expose deployment %s-deployment --type LoadBalancer --name %s-service --external-ip %s", kubeconfig, mf.Metadata.Name, mf.Metadata.Name, ipaddr)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error creating service for kubernetes deployment, %s, %s, %v", cmd, out, err)
	}
	cmd = fmt.Sprintf("%s kubectl get svc %-service -o jsonpath='{.spec.ports[0].port}'", mf.Spec.Kubernetes, mf.Metadata.Name)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error getting port for kubernetes service, %s, %s, %v", cmd, out, err)
	}
	if _, err := strconv.Atoi(out); err == nil {
		return fmt.Errorf("malformed port from kubectl for svc %s, %s, %v", mf.Metadata.Name, out, err)
	}
	origin := fmt.Sprintf("http://%s:%s", ipaddr, out)
	errs := AddPathReverseProxy(rootLB.Name, mf.Metadata.Name, origin)
	if errs != nil {
		return fmt.Errorf("Errors adding reverse proxy path, %v", errs)
	}
	return nil
}

//TODO DeleteKubernetesApp

//TODO `helm` support

//ValidateKubernetesParameters checks the kubernetes parameters and kubeconfig settings
func ValidateKubernetesParameters(rootLB *MEXRootLB, clustName string) (string, ssh.Client, string, error) {
	log.Debugln("validate kubernetes parameters", rootLB, clustName)
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
	log.Debugln("apply kubernetes manifest", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return err
	}
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name")
	}
	kubeconfig, client, _, err := ValidateKubernetesParameters(rootLB, mf.Metadata.Name)
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("%s kubectl apply -f %s", kubeconfig, mf.Spec.Kubernetes)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error applying kubernetes manifest, %s, %s, %v", cmd, out, err)
	}
	return nil
}

//CreateKubernetesNamespaceManifest creates a new namespace in kubernetes
func CreateKubernetesNamespaceManifest(mf *Manifest) error {
	log.Debugln("create kubernetes namespace", mf)
	err := KubernetesApplyManifest(mf)
	if err != nil {
		return fmt.Errorf("error applying kubernetes namespace manifest, %v", err)
	}
	return nil
}

//TODO DeleteKubernetesNamespace

//TODO create configmap and secret to use private docker repos. Kubernetes way of doing this very complicated.

//TODO allow configmap creation from files

//SetKubernetesConfigmapValues sets a key-value in kubernetes configmap
func SetKubernetesConfigmapValues(rootLBName string, clustername string, configname string, keyvalues ...string) error {
	log.Debugln("set configmap values", rootLBName, clustername, configname)
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
	log.Debugln("get kubernetes configmap", rootLBName, clustername, configname)
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

//CreateDockerAppManifest creates an app stricly just plain docker, not kubernetes
func CreateDockerAppManifest(mf *Manifest) error {
	log.Debugln("create docker app", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return fmt.Errorf("can't locate rootLB %s", mf.Spec.RootLB)
	}
	if mf.Spec.DockerRegistry == "" {
		return fmt.Errorf("missing docker registry")
	}
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name")
	}
	if !strings.Contains(mf.Spec.Flavor, "docker") {
		return fmt.Errorf("unsupported docker flavor %s", mf.Spec.Flavor)
	}
	if mf.Spec.Key == "" {
		return fmt.Errorf("empty cluster name")
	}
	if mf.Spec.Image == "" {
		return fmt.Errorf("empty image path")
	}
	if mf.Spec.ProxyPath == "" {
		return fmt.Errorf("empty proxy path")
	}
	if !strings.Contains(mf.Spec.ProxyPath, rootLB.Name) {
		return fmt.Errorf("invalid proxy path %s", mf.Spec.ProxyPath)
	}
	pis := strings.Split(mf.Spec.ProxyPath, "/")
	if len(pis) != 2 {
		return fmt.Errorf("malformed proxy path %s", mf.Spec.ProxyPath)
	}
	proxypath := pis[1]
	if proxypath == "" {
		return fmt.Errorf("empty proxy path after parsing")
	}
	ready, err := IsClusterReady(mf, rootLB)
	if !ready {
		return err
	}
	//XXX what is AccessLayerL4L7
	if !strings.Contains(mf.Spec.AccessLayer, "L7") { // XXX for now
		return fmt.Errorf("access layer %s not supported (yet)", mf.Spec.AccessLayer)
	}
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("create docker app, missing external network in platform config")
	}
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return err
	}
	var origin string
	//TODO XXX unknown format of the mapped ports string
	if mf.Spec.PortMap != "" {
		if !strings.Contains(mf.Spec.PortMap, ":") {
			return fmt.Errorf("invalid port map string %s", mf.Spec.PortMap)
		}
		ix := strings.Index(mf.Spec.PortMap, ":")
		origin = "http://localhost" + mf.Spec.PortMap[ix:] //XXX temporary until ports string is sorted out.
	}
	//dockerPort:= mf.Spec.PortMap[:ix]
	//TODO mf.Spec.PortMap and mf.Spec.PathMap
	var cmd string
	log.Debugln("attempt to use docker registry", mf.Spec.DockerRegistry)
	switch mf.Spec.DockerRegistry {
	case "docker.io":
		cmd = fmt.Sprintf("docker run -d --rm --name %s --net=host %s", mf.Metadata.Name, mf.Spec.Image) //XXX net=host
	case "registry.mobiledgex.net:5000":
		cmd = fmt.Sprintf("cat .docker-pass| docker login -u mobiledgex --password-stdin %s; docker run -d --rm --name %s --net=host %s", mf.Spec.DockerRegistry, mf.Metadata.Name, mf.Spec.Image)
	default:
		return fmt.Errorf("unsupported registry %s", mf.Spec.DockerRegistry)
	}
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error running docker app, %s, %s, %v", cmd, out, err)
	}
	if origin != "" {
		errs := AddPathReverseProxy(rootLB.Name, proxypath, origin)
		if errs != nil {
			return fmt.Errorf("Errors adding reverse proxy path, %v", errs)
		}
	}
	return nil
}

//TODO docker logs

//AddPathReverseProxy adds a new route to origin on the reverse proxy
func AddPathReverseProxy(rootLBName, path, origin string) []error {
	log.Debugln("add path to reverse proxy", rootLBName, path, origin)
	if path == "" {
		return []error{fmt.Errorf("empty path")}
	}
	if origin == "" {
		return []error{fmt.Errorf("empty origin")}
	}
	request := gorequest.New()
	maURI := fmt.Sprintf("http://%s:%s/v1/proxy", rootLBName, mexEnv["MEX_AGENT_PORT"])
	// The L7 reverse proxy terminates TLS at the RootLB and uses path routing to get to the service at a IP:port
	pl := fmt.Sprintf(`{ "message": "add", "proxies": [ { "path": "/%s", "origin": "%s" } ] }`, path, origin)
	resp, body, errs := request.Post(maURI).Set("Content-Type", "application/json").Send(pl).End()
	if errs != nil {
		return errs
	}
	if strings.Contains(body, "OK") {
		return nil
	}
	errs = append(errs, fmt.Errorf("resp %v, body %s", resp, body))
	return errs
}

//StartKubectlProxy starts kubectl proxy on the rootLB to handle kubectl commands remotely.
//  To be called after copying over the kubeconfig file from cluster to rootLB.
func StartKubectlProxy(rootLB *MEXRootLB, kubeconfig string) error {
	log.Debugln("start kubectl proxy", rootLB, kubeconfig)
	if rootLB.PlatConf.Spec.ExternalNetwork == "" {
		return fmt.Errorf("start kubectl proxy, missing external network in platform config")
	}
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return err
	}
	//XXX --port for multiple kubectl proxy
	cmd := fmt.Sprintf("kubectl proxy  --accept-hosts='.*' --address='0.0.0.0' --kubeconfig=%s ", kubeconfig)
	//Use .Start() because we don't want to hang
	cl1, cl2, err := client.Start(cmd)
	if err != nil {
		return fmt.Errorf("error running kubectl proxy, %s,  %v", cmd, err)
	}
	cl1.Close() //nolint
	cl2.Close() //nolint
	cmd = fmt.Sprintf("ps ax |grep %s", kubeconfig)
	for i := 0; i < 5; i++ {
		//verify
		out, outerr := client.Output(cmd)
		if outerr == nil {
			log.Debugln("kubectl proxy confirmed running", out)
			return nil
		}
		log.Debugln("waiting for kubectl proxy...")
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("timeout error verifying kubectl proxy")
}

//DestroyDockerAppManifest  kills an app stricly just plain docker, not kubernetes
func DestroyDockerAppManifest(mf *Manifest) error {
	log.Debugln("kill docker app", mf)
	rootLB, err := getRootLB(mf.Spec.RootLB)
	if err != nil {
		return fmt.Errorf("can't locate rootLB %s", mf.Spec.RootLB)
	}
	if mf.Metadata.Name == "" {
		return fmt.Errorf("missing name")
	}
	client, err := GetSSHClient(rootLB.Name, rootLB.PlatConf.Spec.ExternalNetwork, "root")
	if err != nil {
		return err
	}
	//var origin string
	cmd := fmt.Sprintf("docker kill %s", mf.Metadata.Name)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error killing docker app, %s, %s, %v", cmd, out, err)
	}
	//TODO remove reverse proxy if it was added. need to tag and signal this fact.
	return nil
}
