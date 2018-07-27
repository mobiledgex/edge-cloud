package crmutil

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"

	"github.com/mobiledgex/edge-cloud-infra/openstack-prov/oscliapi"
	"github.com/mobiledgex/edge-cloud-infra/openstack-tenant/agent/cloudflare"
	"github.com/mobiledgex/edge-cloud/edgeproto"

	valid "github.com/asaskevich/govalidator"
	"github.com/codeskyblue/go-sh"

	"github.com/nanobox-io/golang-ssh"
	"github.com/parnurzeal/gorequest"
	log "gitlab.com/bobbae/logrus"
	//"github.com/fsouza/go-dockerclient"
)

var eRootLBName = "mex-lb-1.mobiledgex.net" //has to be FQDN
var eMEXAgentPort = "18889"
var eMEXZone = "mobiledgex.net"
var eMEXDir = os.Getenv("HOME") + "/.mobiledgex"
var eMEXSSHKey = "id_rsa_mobiledgex"
var eMEXAgentImage = "registry.mobiledgex.net:5000/mobiledgex/mexosagent" //XXX missing vers
var eMEXExternalRouter = "mex-k8s-router-1"
var eMEXExternalNetwork = "external-network-shared"
var eMEXK8SMaster = "mex-k8s-master"
var eMEXK8SNode = "mex-k8s-node"
var eMEXNetwork = "mex-k8s-net-1"
var eCFKey = os.Getenv("MEX_CF_KEY")
var eCFUser = os.Getenv("MEX_CF_USER")
var eMEXDockerRegistry = "registry.mobiledgex.net:5000"
var eMEXDockerRegPass = os.Getenv("MEX_DOCKER_REG_PASS")

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
	activeStatus  = "active"
	activeService = "ACTIVE"
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
		return false
	}

	//XXX do more validation on internal env vars
	IsValidMEXOSEnv = true
	return IsValidMEXOSEnv
}

//AddFlavor adds a new flavor to be kept track of
func AddFlavor(flavor string) error {
	if err := ValidateFlavor(flavor); err != nil {
		return fmt.Errorf("invalid flavor")
	}

	for _, f := range AvailableClusterFlavors {
		if flavor == f.Name {
			if f.Status == activeStatus {
				return nil // fmt.Errorf("exists already")
			}
			if f.Status == "available" {
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
	for _, f := range AvailableClusterFlavors {
		if flavor == f.Name {
			if f.Status == activeStatus {
				return f, nil
			}
			return nil, fmt.Errorf("flavor exists but status not active")
		}

	}

	return nil, fmt.Errorf("flavor does not exist")
}

//CreateClusterFromClusterInstData uses clusterInst data to create cluster
func CreateClusterFromClusterInstData(rootLB string, c *edgeproto.ClusterInst) error {
	flavor := c.Flavor.Name
	name := c.Key.ClusterKey.Name
	netSpec := ""                 // not available from controller
	tags := c.Key.ClusterKey.Name // not available, just use key as tag
	tenant := ""                  //not avail

	// XXX ClusterKey is the only unique thing here. So for a given key,
	// like pokemon, there can be only one instance of cluster
	// for that.
	//  There is no way for us to return the actual meaningful name
	//  created in the openstack cluster which can identify the instance.
	// So deletion has to happen with ClusterKey only as well.
	return CreateCluster(rootLB, flavor, name, netSpec, tags, tenant)
}

//CreateCluster creates a cluster of nodes. It can take a while, so call from a goroutine.
func CreateCluster(rootLB, flavor, name, netSpec, tags, tenant string) error {
	if flavor == "" {
		return fmt.Errorf("empty flavor")
	}

	if err := ValidateFlavor(flavor); err != nil {
		return fmt.Errorf("invalid flavor")
	}
	//XXX we only support x1.medium for now

	if !IsFlavorSupported(flavor) {
		return fmt.Errorf("unsupported flavor")
	}

	cf, err := GetClusterFlavor(flavor)
	if err != nil {
		return err
	}

	if cf.NumNodes < 1 {
		return fmt.Errorf("invalid flavor profile, %v", cf)
	}

	if netSpec == "" {
		netSpec = cf.NetworkSpec

	}

	ni, err := oscli.ParseNetSpec(netSpec)
	if err != nil {
		return fmt.Errorf("invalid netSpec, %v", err)
	}

	if ni.Kind != "priv-subnet" {
		return fmt.Errorf("unsupported netSpec kind")
		// XXX for now
	}

	//XXX allow more net types

	//TODO validate CIDR, etc.

	if tags == "" {
		tags = "unknown-tag"
	}
	if tenant == "" {
		tenant = "unknown-tenant"
	}

	err = ValidateTenant(tenant)
	if err != nil {
		return fmt.Errorf("can't validate tenant, %v", err)
	}

	err = ValidateTags(tags)
	if err != nil {
		return fmt.Errorf("invalid tag, %v", err)
	}

	//XXX should check for quota, permissions, access control, etc. here
	//       but we don't have sufficient information from above layer

	guid := xid.New()

	//construct master node name
	id := 1
	kvmname := fmt.Sprintf("%s-%d-%s-%s", eMEXK8SMaster, id, name, guid.String())

	err = oscli.CreateMEXKVM(kvmname, "k8s-master", netSpec, tags, tenant, id)
	if err != nil {
		return fmt.Errorf("can't create k8s master, %v", err)
	}

	for i := 1; i <= cf.NumNodes; i++ {
		//construct node name
		kvmnodename := fmt.Sprintf("%s-%d-%s-%s", eMEXK8SNode, i, name, guid.String())

		err = oscli.CreateMEXKVM(kvmnodename, "k8s-node", netSpec, tags, tenant, i)
		if err != nil {
			return fmt.Errorf("can't create k8s master, %v", err)
		}
	}

	//If RootLB is not running yet, then the following will fail.

	if err = LBAddRoute(rootLB, kvmname); err != nil {
		return err
	}

	if err = oscli.SetServerProperty(kvmname, "mex-flavor="+flavor); err != nil {
		return err
	}

	ready := false

	for i := 0; i < 10; i++ {
		ready, err = IsClusterReady(rootLB, kvmname, flavor)
		if err != nil {
			return err
		}
		if ready {
			break
		}

		time.Sleep(30 * time.Second)
	}

	if !ready {
		return fmt.Errorf("cluster not ready (yet)")
	}

	return nil
}

//LBAddRoute adds a route to LB
func LBAddRoute(rootLB, name string) error {
	if rootLB == "" {
		return fmt.Errorf("empty rootLB")
	}
	if name == "" {
		return fmt.Errorf("empty name")
	}
	ap, err := LBGetRoute(rootLB, name)
	if err != nil {
		return err
	}
	if len(ap) != 2 {
		return fmt.Errorf("expected 2 addresses, got %d", len(ap))
	}

	cmd := fmt.Sprintf("ip route add %s via %s dev ens3", ap[0], ap[1])
	client, err := GetSSHClient(rootLB, eMEXExternalNetwork, "root")
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
func LBRemoveRoute(rootLB, name string) error {
	ap, err := LBGetRoute(rootLB, name)
	if err != nil {
		return err
	}
	if len(ap) != 2 {
		return fmt.Errorf("expected 2 addresses, got %d", len(ap))
	}

	cmd := fmt.Sprintf("ip route delete %s via %s dev ens3", ap[0], ap[1])
	client, err := GetSSHClient(rootLB, eMEXExternalNetwork, "root")
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
	return flavor == "x1.medium"
}

//DeleteClusterByName remove cluster by name
func DeleteClusterByName(rootLB, name string) error {
	// The ClusterKey name + random-string (generated at creation time) should be passed here.

	// But controller does not know because it does not receive results from the creation and store them.
	// XXX This means ClusterKey name is the only id. This limits single cluster instance per clusterKey name.

	srvs, err := oscli.ListServers()
	if err != nil {
		return err
	}

	log.Debugln("servers", srvs)
	for _, s := range srvs {
		if strings.Contains(s.Name, name) {
			if strings.Contains(s.Name, eMEXK8SMaster) {
				err = LBRemoveRoute(rootLB, s.Name)
				if err != nil {
					err = fmt.Errorf("failed remove route for %s, %v", s.Name, err)
					log.Debugln(err)
					return err
				}
			}
			log.Debugln("delete server", s.Name)
			err = oscli.DeleteServer(s.Name)
			if err != nil {
				log.Debugln(err)
				return err
			}
		}
	}

	sns, err := oscli.ListSubnets("")
	if err != nil {
		log.Debugln(err)
		return err
	}
	log.Debugln("subnets", sns)

	rn := oscli.GetMEXExternalRouter() //XXX for now
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
				log.Debugln(err)
				return err
			}
			break
		}
	}

	//XXX tell agent to remove the route
	return nil
}

//InitEnvVars sets up environment
func InitEnvVars() {
	dockerRegistry := os.Getenv("MEX_DOCKER_REGISTRY")
	if dockerRegistry != "" {
		eMEXDockerRegistry = dockerRegistry
	}
	extRouter := os.Getenv("MEX_EXT_ROUTER") // mex-k8s-router-1
	if extRouter != "" {
		eMEXExternalRouter = extRouter
	}
	intNetwork := os.Getenv("MEX_NETWORK")
	if intNetwork != "" {
		eMEXNetwork = intNetwork
	}

	extNetwork := os.Getenv("MEX_EXT_NETWORK") // "external-network-shared"
	if extNetwork != "" {
		eMEXExternalNetwork = extNetwork
	}

	domainZone := os.Getenv("MEX_ZONE")
	if domainZone != "" {
		eMEXZone = domainZone
	}

	agentImage := os.Getenv("MEX_AGENT_IMAGE")
	if agentImage != "" {
		eMEXAgentImage = agentImage
	}
	mexDir := os.Getenv("MEX_DIR")
	if mexDir != "" {
		eMEXDir = mexDir
	}
	rootLB := os.Getenv("MEX_ROOT_LB")
	if rootLB != "" {
		eRootLBName = rootLB //XXX
	}
}

//EnableRootLB creates a seed presence node in cloudlet that also becomes first Agent node.
//  It also sets up first basic network router and subnet, ready for running first MEX agent.
func EnableRootLB(rootLB string) error {
	InitEnvVars()

	err := oscli.PrepNetwork()
	if err != nil {
		return err
	}

	sl, err := oscli.ListServers()
	if err != nil {
		return err
	}
	found := 0
	for _, s := range sl {
		if strings.Contains(s.Name, "mex-lb-") && strings.Contains(s.Name, "mobiledgex.net") {
			found++
		}
	}
	if found == 0 {
		//XXX no tenant , no tags -- we don't have the information
		//   Role as mex-agent-node, for now.
		err := oscli.CreateMEXKVM(rootLB, "mex-agent-node", "external-ip,external-network-shared,10.101.X.X/24,dhcp", "", "", 1)
		if err != nil {
			return err
		}
		log.Debugln("created %s", rootLB)
	} else {
		log.Debugln("reusing %s", rootLB)
	}

	return nil
}

//XXX allow creating more than one LB

//GetServerIPAddr gets the server IP
func GetServerIPAddr(networkName, serverName string) (string, error) {
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
	return addr, nil
}

//CopySSHCredential copies over the ssh credential for mex to LB
func CopySSHCredential(serverName, networkName, userName string) error {
	addr, err := GetServerIPAddr(networkName, serverName)
	if err != nil {
		return err
	}

	kf := eMEXDir + "/" + eMEXSSHKey
	out, err := sh.Command("scp", "-o", "StrictHostKeyChecking=no", "-i", kf, kf, "root@"+addr+":").Output()
	if err != nil {
		return fmt.Errorf("can't copy %s to %s, %s, %v", kf, addr, out, err)
	}
	return nil
}

//GetSSHClient returns ssh client handle for the server
func GetSSHClient(serverName, networkName, userName string) (ssh.Client, error) {
	auth := ssh.Auth{Keys: []string{eMEXDir + "/id_rsa_mobiledgex"}}

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

//GetRootLBName returns default rootLB name
func GetRootLBName() string {
	return eRootLBName
}

//WaitForRootLB waits for the RootLB instance to be up and copies of SSH credentials for internal networks.
//  Idempotent, but don't call all the time.
func WaitForRootLB(rootLB string) error {
	client, err := GetSSHClient(rootLB, eMEXExternalNetwork, "root")
	if err != nil {
		return err
	}

	running := false
	for i := 0; i < 10; i++ {
		_, err := client.Output("grep done /tmp/mobiledgex.log") //XXX beware of use of word done
		if err == nil {
			running = true
			if err := CopySSHCredential(rootLB, eMEXExternalNetwork, "root"); err != nil {
				return fmt.Errorf("can't copy ssh credential to RootLB, %v", err)
			}
			break
		}

		time.Sleep(30 * time.Second)
	}

	if !running {
		return fmt.Errorf("while creating cluster, timeout waiting for RootLB")
	}

	//XXX just ssh docker commands instead since there is no real benefit for now
	//err = InitDockerMachine(rootLB,addr)

	return nil
}

//InitDockerMachine preps docker-machine for use with docker command
func InitDockerMachine(rootLB, addr string) error {
	home := os.Getenv("HOME")

	_, err := sh.Command("docker-machine", "create", "-d", "generic", "--generic-ip-address", addr, "--generic-ssh-key", eMEXDir+"/id_rsa_mobiledgex", "--generic-ssh-user", "bob", rootLB).Output()
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

//RunMEXAgent runs the MEX agent on the RootLB. It first registers FQDN to cloudflare domain registry if not already registered.
//   It then obtains certficiates from Letsencrypt, if not done yet.  Then it runs the docker instance of MEX agent
//   on the RootLB. It can be told to manually pull image from docker repository.  This allows upgrading with new image.
//   It uses MEX private docker repository.  If an instance is running already, we don't start another one.
func RunMEXAgent(fqdn string, pull bool) error {
	//fqdn is that of the machine/kvm-instance running the agent
	if !valid.IsDNSName(fqdn) {
		return fmt.Errorf("fqdn %s is not valid", fqdn)
	}

	err := EnableRootLB(fqdn)
	if err != nil {
		return fmt.Errorf("Failed to enable root LB %v", err)
	}

	err = WaitForRootLB(fqdn)
	if err != nil {
		return fmt.Errorf("Error waiting for rootLB %v", err)
	}

	client, err := GetSSHClient(fqdn, eMEXExternalNetwork, "root")
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("docker ps |grep %s", eMEXAgentImage)
	_, err = client.Output(cmd)
	if err == nil {
		//agent docker instance exists
		//XXX check better
		return nil
	}

	if err = ActivateFQDNA(fqdn); err != nil {
		return err
	}

	err = AcquireCertificates(fqdn)
	if err != nil {
		return fmt.Errorf("can't acquire certificate for %s, %v", fqdn, err)
	}

	if eMEXDockerRegPass == "" {
		return fmt.Errorf("empty docker registry pass env var")
	}

	cmd = fmt.Sprintf("echo %s > .docker-pass", eMEXDockerRegPass)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't store docker pass, %s, %v", out, err)
	}

	if pull {
		cmd = fmt.Sprintf("cat .docker-pass| docker login -u mobiledgex --password-stdin %s; docker pull %s; docker run -d --rm --name %s --net=host -v `pwd`:/var/www/.cache -v /etc/ssl/certs:/etc/ssl/certs %s -debug", eMEXAgentImage, eMEXDockerRegistry, fqdn, eMEXAgentImage)
	} else {
		cmd = fmt.Sprintf("cat .docker-pass| docker login -u mobiledgex --password-stdin %s; docker run -d --rm --name %s --net=host -v `pwd`:/var/www/.cache -v /etc/ssl/certs:/etc/ssl/certs %s -debug", eMEXDockerRegistry, fqdn, eMEXAgentImage)
	}

	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error running dockerized agent on RootLB %s, %s, %s, %v", fqdn, cmd, out, err)
	}

	return nil
}

//UpdateMEXAgent upgrades the mex agent
func UpdateMEXAgent(fqdn string) error {
	err := RemoveMEXAgent(fqdn)
	if err != nil {
		return err
	}

	// Force pulling a potentially newer docker image
	return RunMEXAgent(fqdn, true)
}

//RemoveMEXAgent deletes mex agent docker instance
func RemoveMEXAgent(fqdn string) error {
	err := oscli.DeleteServer(fqdn)
	if err != nil {
		return err
	}

	recs, err := cloudflare.GetDNSRecords(eMEXZone)
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

//AcquireCertificates obtains certficates from Letsencrypt over ACME. It should be used carefully. The API calls have quota.
func AcquireCertificates(fqdn string) error {
	if eCFKey == "" {
		return fmt.Errorf("no MEX_CF_KEY")
	}
	if eCFUser == "" {
		return fmt.Errorf("no MEX_CF_USER")
	}

	client, err := GetSSHClient(fqdn, eMEXExternalNetwork, "root")
	if err != nil {
		return fmt.Errorf("can't get ssh client for acme.sh, %v", err)
	}
	fullchain := fqdn + "/fullchain.cer"
	cmd := fmt.Sprintf("ls -a %s", fullchain)
	_, err = client.Output(cmd)
	if err == nil {
		return nil
	}

	cmd = fmt.Sprintf("docker run --rm -e CF_Key=%s -e CF_Email=%s -v `pwd`:/acme.sh --net=host neilpang/acme.sh --issue -d %s --dns dns_cf", eCFKey, eCFUser, fqdn)

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
func ActivateFQDNA(fqdn string) error {
	if eCFKey == "" {
		return fmt.Errorf("no MEX_CF_KEY")
	}
	if eCFUser == "" {
		return fmt.Errorf("no MEX_CF_USER")
	}

	if err := cloudflare.InitAPI(eCFUser, eCFKey); err != nil {
		return fmt.Errorf("cannot init cloudflare api, %v", err)
	}

	dr, err := cloudflare.GetDNSRecords(eMEXZone)
	if err != nil {
		return fmt.Errorf("cannot get dns records for %s, %v", fqdn, err)
	}

	addr, err := GetServerIPAddr(eMEXExternalNetwork, fqdn)

	for _, d := range dr {
		if d.Type == "A" && d.Name == fqdn {
			if d.Content == addr {
				return nil
			}
			log.Warningf("cloudflare A record has different address %v, not %s, it will be overwritten", d, addr)
			if err = cloudflare.DeleteDNSRecord(eMEXZone, d.ID); err != nil {
				return fmt.Errorf("can't delete DNS record for %s, %v", fqdn, err)
			}
			break
		}
	}

	if err != nil {
		return err
	}
	if err := cloudflare.CreateDNSRecord(eMEXZone, fqdn, "A", addr, 1, false); err != nil {
		return fmt.Errorf("can't create DNS record for %s, %v", fqdn, err)
	}

	//once successfully inserted the A record will take a bit of time, but not too long due to fast cloudflare anycast
	time.Sleep(10 * time.Second)

	return nil
}

//nolint
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		err = in.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		err = out.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

//IsClusterReady checks to see if cluster is read, i.e. rootLB is running and active
func IsClusterReady(rootLB, clustername, flavor string) (bool, error) {
	if flavor == "" {
		return false, fmt.Errorf("empty flavor")
	}

	if !IsFlavorSupported(flavor) {
		return false, fmt.Errorf("unsupported flavor")
	}
	cf, err := GetClusterFlavor(flavor)
	if err != nil {
		return false, err
	}

	if cf.NumNodes < 1 {
		return false, fmt.Errorf("invalid flavor profile, %v", cf)
	}

	name, err := FindClusterWithKey(clustername)
	if err != nil {
		return false, fmt.Errorf("can't find cluster with key %s, %v", clustername, err)
	}
	ipaddr, err := FindNodeIP(name)
	if err != nil {
		return false, err
	}
	client, err := GetSSHClient(rootLB, eMEXExternalNetwork, "root")
	if err != nil {
		return false, fmt.Errorf("can't get ssh client for cluser ready check, %v", err)
	}

	cmd := fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %s bob@%s kubectl get nodes | grep Ready |grep -v NotReady| wc -l", eMEXSSHKey, ipaddr)
	out, err := client.Output(cmd)
	if err != nil {
		return false, fmt.Errorf("kubectl fail on %s, %s, %v", name, out, err)
	}

	totalNodes := cf.NumNodes + cf.NumMasterNodes
	tn := fmt.Sprintf("%d", totalNodes)

	if !strings.Contains(out, tn) {
		return false, fmt.Errorf("not ready, %s", out)
	}

	if err := CopyKubeConfig(rootLB, name); err != nil {
		return false, fmt.Errorf("kubeconfig copy failed, %v", err)
	}
	return true, nil
}

//CopyKubeConfig copies over kubeconfig from the cluster
func CopyKubeConfig(rootLB, name string) error {
	ipaddr, err := FindNodeIP(name)
	if err != nil {
		return err
	}
	client, err := GetSSHClient(rootLB, eMEXExternalNetwork, "root")
	if err != nil {
		return fmt.Errorf("can't get ssh client for copying kubeconfig, %v", err)
	}

	cmd := fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %s bob@%s:.kube/config  kubeconfig-%s", eMEXSSHKey, ipaddr, name)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't copy kubeconfig from %s, %s, %v", name, out, err)
	}
	cmd = fmt.Sprintf("cat kubeconfig-%s", name)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("can't cat kubeconfig-%s, %s, %v", name, out, err)
	}
	err = ioutil.WriteFile(eMEXDir+"/kubeconfig-"+name, []byte(out), 0666)
	if err != nil {
		return fmt.Errorf("can't write kubeconfig-%s content,%v", name, err)
	}
	return nil
}

//FindNodeIP finds IP for the given node
func FindNodeIP(name string) (string, error) {
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
	if key == "" {
		return "", fmt.Errorf("empty key")
	}
	srvs, err := oscli.ListServers()
	if err != nil {
		return "", err
	}

	for _, s := range srvs {
		if s.Status == activeService && strings.Contains(s.Name, key) && strings.Contains(s.Name, eMEXK8SMaster) {
			return s.Name, nil
		}
	}
	return "", fmt.Errorf("key %s not found", key)
}

//CreateKubernetesApp instantiates a new kubernetes deployment
func CreateKubernetesApp(rootLB, clustername, deployment, manifest string) error {
	kubeconfig, client, ipaddr, err := ValidateKubernetesParameters(rootLB, clustername, manifest)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("%s kubectl apply -f %s", kubeconfig, manifest)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deploying kubernetes app, %s, %s, %v", cmd, out, err)
	}

	//TODO: use service manifest, or require a single manifest with both deployment and service. Use namespaces.
	cmd = fmt.Sprintf("%s kubectl expose deployment %s-deployment --type LoadBalancer --name %s-service --external-ip %s", kubeconfig, deployment, deployment, ipaddr)

	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error creating service for kubernetes deployment, %s, %s, %v", cmd, out, err)
	}
	cmd = fmt.Sprintf("%s kubectl get svc %-service -o jsonpath='{.spec.ports[0].port}'", kubeconfig, deployment)
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error getting port for kubernetes service, %s, %s, %v", cmd, out, err)
	}

	if _, err := strconv.Atoi(out); err == nil {
		return fmt.Errorf("malformed port from kubectl for svc %s, %s, %v", deployment, out, err)
	}
	origin := fmt.Sprintf("http://%s:%s", ipaddr, out)
	errs := AddPathReverseProxy(rootLB, deployment, origin)
	if errs != nil {
		return fmt.Errorf("Errors adding reverse proxy path, %v", errs)
	}

	return nil
}

//TODO DeleteKubernetesApp

//TODO `helm` support

//ValidateKubernetesParameters checks the kubernetes parameters and kubeconfig settings
func ValidateKubernetesParameters(rootLB, clustername, manifest string) (string, ssh.Client, string, error) {
	client, err := GetSSHClient(rootLB, eMEXExternalNetwork, "root")
	if err != nil {
		return "", nil, "", err
	}

	//TODO: support other URI: file://, nfs://, ftp://, git://, or embedded as base64 string
	if manifest != "" &&
		!strings.HasPrefix(manifest, "http://") &&
		!strings.HasPrefix(manifest, "https://") {
		return "", nil, "", fmt.Errorf("unsupported manifest")
	}

	name, err := FindClusterWithKey(clustername)
	if err != nil {
		return "", nil, "", fmt.Errorf("can't find cluster with key %s, %v", clustername, err)
	}
	ipaddr, err := FindNodeIP(name)
	if err != nil {
		return "", nil, "", err
	}
	kubeconfig := fmt.Sprintf("KUBECONFIG=kubeconfig-%s", name)

	return kubeconfig, client, ipaddr, nil
}

//KubernetesApplyManifest does `apply` on the manifest yaml
func KubernetesApplyManifest(rootLB, clustername, manifest string) error {
	kubeconfig, client, _, err := ValidateKubernetesParameters(rootLB, clustername, manifest)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("%s kubectl apply -f %s", kubeconfig, manifest)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error applying kubernetes manifest, %s, %s, %v", cmd, out, err)
	}

	return nil
}

//CreateKubernetesNamespace creates a new namespace in kubernetes
func CreateKubernetesNamespace(rootLB, clustername, manifest string) error {
	err := KubernetesApplyManifest(rootLB, clustername, manifest)
	if err != nil {
		return fmt.Errorf("error applying kubernetes namespace manifest, %v", err)
	}

	return nil
}

//TODO DeleteKubernetesNamespace

//TODO create configmap and secret to use private docker repos. Kubernetes way of doing this very complicated.

//TODO allow configmap creation from files

//SetKubernetesConfigmapValues sets a key-value in kubernetes configmap
func SetKubernetesConfigmapValues(rootLB, clustername, configname string, keyvalues ...string) error {
	kubeconfig, client, _, err := ValidateKubernetesParameters(rootLB, clustername, "")
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
func GetKubernetesConfigmapYAML(rootLB, clustername, configname string) (string, error) {
	kubeconfig, client, _, err := ValidateKubernetesParameters(rootLB, clustername, "")
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

//TODO support https://github.com/bitnami-labs/kubewatch

//CreateDockerApp creates an app stricly just plain docker, not kubernetes
func CreateDockerApp(rootLB, appname, clustername, flavorname, registryname, uri, imagename, mports, mpaths, accesslayer string) error {
	if appname == "" {
		return fmt.Errorf("emptyh app name")
	}

	if !IsFlavorSupported(flavorname) {
		return fmt.Errorf("unsupported flavor")
	}

	if clustername == "" {
		return fmt.Errorf("empty cluster name")
	}
	if imagename == "" {
		return fmt.Errorf("empty image path")
	}

	proxypath := ""
	if uri != "" {
		//XXX Not sure what Uri is supposed to contain.  But it seems the only place
		//  to hold URL that can be used to set up reverse proxy at a publicly accessible
		//  DNS pointed Internet node which acts as a path based router for internally
		//  deployed services.
		if !strings.HasPrefix(uri, rootLB) {
			return fmt.Errorf("invalid uri %s", uri)
		}
		pis := strings.Split(uri, "/")
		if len(pis) != 2 {
			return fmt.Errorf("malformed uri %s", uri)
		}
		proxypath = pis[1]
	}

	ready, err := IsClusterReady(rootLB, clustername, flavorname)
	if !ready {
		return err
	}

	if accesslayer == "unknown" {
		accesslayer = "L7" //XXX default to L7
	}

	if accesslayer != "L7" { // XXX for now
		return fmt.Errorf("access layer %s not supported (yet)", accesslayer)
	}

	client, err := GetSSHClient(rootLB, eMEXExternalNetwork, "root")
	if err != nil {
		return err
	}

	origin := ""
	if mports != "" {
		//TODO format of the mapped ports string, multiple ports can be specified, but only one ports string.
		//   XXX for now we just use host net.
		origin = "http://localhost:80" //XXX temporary until ports string is sorted out.
		//XXX ideally the caller should fill origin out and know how to supply proper information.
	}

	if mpaths != "" { //nolint
		// TODO format of mapped path string. Multiple -v can be specified, but there is only one mpath string
	}

	cmd := ""
	switch registryname {
	case "docker.io":
		cmd = fmt.Sprintf("docker run -d --rm --name %s --net=host %s", appname, imagename)
	case eMEXDockerRegistry:
		cmd = fmt.Sprintf("cat .docker-pass| docker login -u mobiledgex --password-stdin %s; docker run -d --rm --name %s --net=host %s", eMEXDockerRegistry, appname, imagename)
	default:
		return fmt.Errorf("unsupported registry %s", registryname)
	}

	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error running docker app, %s, %s, %v", cmd, out, err)
	}
	if accesslayer == "L7" && proxypath != "" {
		errs := AddPathReverseProxy(rootLB, proxypath, origin)
		if errs != nil {
			return fmt.Errorf("Errors adding reverse proxy path, %v", errs)
		}
	}

	return nil
}

//TODO docker logs

//AddPathReverseProxy adds a new route to origin on the reverse proxy
func AddPathReverseProxy(rootLB, path, origin string) []error {
	if path == "" {
		return []error{fmt.Errorf("empty path")}
	}
	if origin == "" {
		return []error{fmt.Errorf("empty origin")}
	}

	request := gorequest.New()

	maURI := fmt.Sprintf("http://%s:%s/v1/proxy", rootLB, eMEXAgentPort)

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
