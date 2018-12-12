package crmutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/azure"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/gcloud"
	oscli "github.com/mobiledgex/edge-cloud-infra/openstack-prov/oscliapi"
	"github.com/mobiledgex/edge-cloud-infra/openstack-tenant/agent/cloudflare"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

var yamlMEXCluster = `apiVersion: v1
kind: {{.ResourceKind}}
resource: {{.Resource}}
metadata:
  name: {{.Name}}
  tags: {{.Tags}}
  kind: {{.Kind}}
  tenant: {{.Tenant}}
  operator: {{.Operator}}
  region: {{.Region}}
  zone: {{.Zone}}
  location: {{.Location}}
  project: {{.Project}}
  dnszone: {{.DNSZone}}
  resourcegroup: {{.ResourceGroup}}
spec:
  flags: {{.Flags}}
  flavor: {{.Flavor}}
  key: {{.Key}}
  dockerregistry: registry.mobiledgex.net:5000
  rootlb: {{.RootLB}}
  networkscheme: {{.NetworkScheme}}
`

type templateFill struct {
	Name, Kind, Flavor, Tags, Tenant, Region, Zone, DNSZone              string
	ImageFlavor, Location, RootLB, Resource, ResourceKind, ResourceGroup string
	StorageSpec, NetworkScheme, MasterFlavor, Topology                   string
	NodeFlavor, Operator, Key, Image, Options                            string
	ImageType, AppURI, ProxyPath, AgentImage                             string
	ExternalNetwork, Project, AppTemplate                                string
	ExternalRouter, Flags, IpAccess                                      string
	NumMasters, NumNodes                                                 int
	ConfigDetailDeployment, ConfigDetailResources, ConfigDetailManifest  string
	Command                                                              []string
}

//MEXClusterCreateClustInst calls MEXClusterCreate with a manifest created from the template
func MEXClusterCreateClustInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst) error {
	//XXX trigger off clusterInst or flavor to pick the right template: mex, aks, gke
	mf, err := fillClusterTemplateClustInst(rootLB, clusterInst)
	if err != nil {
		return err
	}
	return MEXClusterCreateManifest(mf)
}

func fillClusterTemplateClustInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst) (*Manifest, error) {
	log.DebugLog(log.DebugLevelMexos, "fill cluster template manifest cluster inst", "clustinst", clusterInst)
	if clusterInst.Key.ClusterKey.Name == "" {
		log.DebugLog(log.DebugLevelMexos, "cannot create empty cluster manifest", "clustinst", clusterInst)
		return nil, fmt.Errorf("invalid cluster inst %v", clusterInst)
	}
	if verr := validateClusterKind(clusterInst.Key.CloudletKey.OperatorKey.Name); verr != nil {
		return nil, verr
	}
	data := templateFill{
		ResourceKind:  "cluster",
		Resource:      clusterInst.Flavor.Name,
		Name:          clusterInst.Key.ClusterKey.Name,
		Tags:          clusterInst.Key.ClusterKey.Name + "-tag",
		Tenant:        clusterInst.Key.ClusterKey.Name + "-tenant",
		Operator:      util.K8SSanitize(clusterInst.Key.CloudletKey.OperatorKey.Name),
		Key:           clusterInst.Key.ClusterKey.Name,
		Kind:          "kubernetes",
		ResourceGroup: clusterInst.Key.CloudletKey.Name + "_" + clusterInst.Key.ClusterKey.Name,
		Flavor:        clusterInst.Flavor.Name,
		DNSZone:       "mobiledgex.net",
		RootLB:        rootLB.Name,
		NetworkScheme: "priv-subnet,mex-k8s-net-1,10.101.X.0/24",
	}

	// if these env variables are not set, fall back to the
	// existing defaults based on deployment type(operator name)
	data.Region = os.Getenv("CLOUDLET_REGION")
	data.Zone = os.Getenv("CLOUDLET_ZONE")
	data.Location = os.Getenv("CLOUDLET_LOCATION")

	switch clusterInst.Key.CloudletKey.OperatorKey.Name {
	case "gcp":
		if data.Region == "" {
			data.Region = "us-west1"
		}
		if data.Zone == "" {
			data.Zone = "us-west1-a"
		}
		if data.Location == "" {
			data.Location = "us-west"
		}
		data.Project = "still-entity-201400" // XXX
	case "azure":
		if data.Region == "" {
			data.Region = "centralus"
		}
		if data.Zone == "" {
			data.Zone = "centralus"
		}
		if data.Location == "" {
			data.Location = "centralus"
		}
	default:
		if data.Region == "" {
			data.Region = "eu-central-1"
		}
		if data.Zone == "" {
			data.Zone = "eu-central-1c"
		}
		if data.Location == "" {
			data.Location = "bonn"
		}
	}
	mf, err := templateUnmarshal(&data, yamlMEXCluster)
	if err != nil {
		return nil, err
	}
	return mf, nil
}

//MEXClusterCreateManifest creates a cluster
func MEXClusterCreateManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "creating cluster via manifest")
	switch mf.Metadata.Operator {
	case "gcp":
		return gcloudCreateGKE(mf)
	case "azure":
		return azureCreateAKS(mf)
	default:
		//guid, err := mexCreateClusterKubernetes(mf)
		err := mexCreateClusterKubernetes(mf)
		if err != nil {
			return fmt.Errorf("can't create cluster, %v", err)
		}
		//log.DebugLog(log.DebugLevelMexos, "new guid", "guid", *guid)
		log.DebugLog(log.DebugLevelMexos, "created kubernetes cluster")
		return nil
	}
}

func azureCreateAKS(mf *Manifest) error {
	var err error
	if err = azure.CreateResourceGroup(mf.Metadata.ResourceGroup, mf.Metadata.Location); err != nil {
		return err
	}
	if err = azure.CreateAKSCluster(mf.Metadata.ResourceGroup, mf.Metadata.Name); err != nil {
		return err
	}
	//race condition exists where the config file is not ready until just after the cluster create is done
	time.Sleep(3 * time.Second)
	saveKubeconfig()
	if err = azure.GetAKSCredentials(mf.Metadata.ResourceGroup, mf.Metadata.Name); err != nil {
		return err
	}
	kconf, err := getKconf(mf, false)
	if err != nil {
		return fmt.Errorf("cannot get kconf, %v, %v, %v", mf, kconf, err)
	}
	if err = copyFile(defaultKubeconfig(), kconf); err != nil {
		return fmt.Errorf("can't copy %s, %v", defaultKubeconfig(), err)
	}
	log.DebugLog(log.DebugLevelMexos, "created aks", "name", mf.Spec.Key)
	return nil
}

func defaultKubeconfig() string {
	return os.Getenv("HOME") + "/.kube/config"
}

func copyFile(src string, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func gcloudCreateGKE(mf *Manifest) error {
	var err error
	if err = gcloud.SetProject(mf.Metadata.Project); err != nil {
		return err
	}
	if err = gcloud.SetZone(mf.Metadata.Zone); err != nil {
		return err
	}
	if err = gcloud.CreateGKECluster(mf.Metadata.Name); err != nil {
		return err
	}
	//race condition exists where the config file is not ready until just after the cluster create is done
	time.Sleep(3 * time.Second)
	saveKubeconfig()
	if err = gcloud.GetGKECredentials(mf.Metadata.Name); err != nil {
		return err
	}
	kconf, err := getKconf(mf, false)
	if err != nil {
		return fmt.Errorf("cannot get kconf, %v, %v, %v", mf, kconf, err)
	}
	if err = copyFile(defaultKubeconfig(), kconf); err != nil {
		return fmt.Errorf("can't copy %s, %v", defaultKubeconfig(), err)
	}
	log.DebugLog(log.
		DebugLevelMexos, "created gke", "name", mf.Spec.Key)
	return nil
}

func saveKubeconfig() {
	kc := defaultKubeconfig()
	if err := os.Rename(kc, kc+".save"); err != nil {
		log.DebugLog(log.DebugLevelMexos, "can't rename", "name", kc, "error", err)
	}
}

//MEXClusterRemoveClustInst calls MEXClusterRemove with a manifest created from the template
func MEXClusterRemoveClustInst(rootLb *MEXRootLB, clusterInst *edgeproto.ClusterInst) error {
	mf, err := fillClusterTemplateClustInst(rootLb, clusterInst)
	if err != nil {
		return err
	}
	return MEXClusterRemoveManifest(mf)
}

var yamlMEXFlavor = `apiVersion: v1
kind: {{.ResourceKind}}
resource: {{.Resource}}
metadata:
  name: {{.Name}}
  tags: {{.Tags}}
  kind: {{.Kind}}
spec:
  flags: {{.Flags}}
  flavor: {{.Name}}
  flavordetail: 
     name: {{.Name}}
     nodes: {{.NumNodes}}
     masters: {{.NumMasters}}
     networkscheme: {{.NetworkScheme}}
     masterflavor: {{.MasterFlavor}}
     nodeflavor: {{.NodeFlavor}}
     storagescheme: {{.StorageSpec}}
     topology: {{.Topology}}
`

//MEXAddFlavorClusterInst uses template to fill in details for flavor add request and calls MEXAddFlavor
// XXX "adding" flavor is not supported. This is based on wrong understanding. Platform providers
//  give us a set of platform flavors. For example, kvm and neutron flavors like m4.large, m4.xlarge.
//  We do not have control over these platform flavors. We do not "create" them; the are given to us.
//  Similarly, our "cluster" flavor, which is our own invention, tries to map a collection of these
//  platform flavors (e.g. m4.large, neutron flavors like loadbalancer (not available to us yet)) into
//  a synthetic flavor -- x1.medium, x1.large, ... for our own abstraction at the "cluster" layer.
//  Cluster is a collection of things, like compute nodes, networking, etc. which require its own "flavor".
//  However, these cluster flavors are static, based on platform flavors we do not control. We cannot
//  create these flavors.  Per cloudlet, a set of possible cluster flavors are abstracted and made
//  available. They can be retrieved from CRM and examined. But they cannot be "created" as they
//  are real-world resources that are given, by the providers.  They should not be created dynamically
//  either.  Even though openstack allows various platform creation, they are at the "operator" level.
//  A platform provider has ability to customize their offering of real-world resources.  The end user
//  of a platform (e.g. openstack user, like us) do not have ability to create flavors on platforms.
//  Similarly, gcp or azure user do not have ability to create flavors. They are used, as is, and users
//  are supposed to list and examine available flavors.  So the concept here of "creating" flavors is flawed
//  and cannot be reasonably implemented. The code is here for legacy reasons.
func MEXAddFlavorClusterInst(flavor *edgeproto.ClusterFlavor) error {
	log.DebugLog(log.DebugLevelMexos, "adding cluster inst flavor", "flavor", flavor)

	if flavor.Key.Name == "" {
		log.DebugLog(log.DebugLevelMexos, "cannot add empty cluster inst flavor", "flavor", flavor)
		return fmt.Errorf("will not add empty cluster inst %v", flavor)
	}
	data := templateFill{
		ResourceKind:  "flavor",
		Resource:      flavor.Key.Name,
		Name:          flavor.Key.Name,
		Tags:          flavor.Key.Name + "-tag",
		Kind:          "mex-k8s-cluster",
		Flags:         flavor.Key.Name + "-flags",
		NumNodes:      int(flavor.NumNodes),
		NumMasters:    int(flavor.NumMasters),
		NetworkScheme: "priv-subnet,mex-k8s-net-1,10.201.X0/24",
		MasterFlavor:  flavor.MasterFlavor.Name,
		NodeFlavor:    flavor.NodeFlavor.Name,
		StorageSpec:   "default",
		Topology:      "type-1",
	}
	mf, err := templateUnmarshal(&data, yamlMEXFlavor)
	if err != nil {
		return err
	}
	return MEXAddFlavor(mf)
}

//MEXAddFlavor adds flavor using manifest
func MEXAddFlavor(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "add mex flavor")
	//TODO use full manifest and validate against platform data
	return AddFlavorManifest(mf)
}

// TODO DeleteFlavor -- but almost never done

// TODO lookup guid using name

//MEXClusterRemoveManifest removes a cluster
func MEXClusterRemoveManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "removing cluster")
	switch mf.Metadata.Operator {
	case "gcp":
		return gcloud.DeleteGKECluster(mf.Metadata.Name)
	case "azure":
		return azure.DeleteAKSCluster(mf.Metadata.ResourceGroup)
	default:
		if err := mexDeleteClusterKubernetes(mf); err != nil {
			return fmt.Errorf("can't remove cluster, %v", err)
		}
		return nil
	}
}

var yamlMEXPlatform = `apiVersion: v1
kind: {{.ResourceKind}}
resource: {{.Resource}}
metadata:
  kind: {{.Kind}}
  name: {{.Name}}
  tags: {{.Tags}}
  tenant: {{.Tenant}}
  region: {{.Region}}
  zone: {{.Zone}}
  location: {{.Location}}
  openrc: ~/.mobiledgex/openrc
  dnszone: {{.DNSZone}}
  operator: {{.Operator}}
spec:
  flags: {{.Flags}}
  flavor: {{.Flavor}}
  rootlb: {{.RootLB}}
  key: {{.Key}}
  dockerregistry: registry.mobiledgex.net:5000
  externalnetwork: {{.ExternalNetwork}}
  networkscheme: {{.NetworkScheme}}
  externalrouter: {{.ExternalRouter}}
  options: {{.Options}}
  agent: 
    image: {{.AgentImage}}
    status: active
`

//XXX how to handle password, keys, ID, etc. in manifest

//MEXPlatformInitCloudletKey calls MEXPlatformInit with templated manifest
func MEXPlatformInitCloudletKey(rootLB *MEXRootLB, cloudletKeyStr string) error {
	//XXX trigger off cloudletKeyStr or flavor to pick the right template: mex, aks, gke
	mf, err := fillPlatformTemplateCloudletKey(rootLB, cloudletKeyStr)
	if err != nil {
		return err
	}
	return MEXPlatformInitManifest(mf)
}

func fillPlatformTemplateCloudletKey(rootLB *MEXRootLB, cloudletKeyStr string) (*Manifest, error) {
	log.DebugLog(log.DebugLevelMexos, "fill template cloudletkeystr", "cloudletkeystr", cloudletKeyStr)
	clk := edgeproto.CloudletKey{}
	err := json.Unmarshal([]byte(cloudletKeyStr), &clk)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal json cloudletkey %s, %v", cloudletKeyStr, err)
	}
	log.DebugLog(log.DebugLevelMexos, "unmarshalled cloudletkeystr", "cloudletkey", clk)
	if clk.Name == "" || clk.OperatorKey.Name == "" {
		log.DebugLog(log.DebugLevelMexos, "will not fill template with invalid cloudletkeystr", "cloudletkeystr", cloudletKeyStr)
		return nil, fmt.Errorf("invalid cloudletkeystr %s", cloudletKeyStr)
	}

	log.DebugLog(log.DebugLevelMexos, "using external network", "extNet", oscli.GetMEXExternalNetwork())

	data := templateFill{
		ResourceKind:    "platform",
		Resource:        util.K8SSanitize(clk.OperatorKey.Name),
		Name:            clk.Name,
		Tags:            clk.Name + "-tag",
		Key:             clk.Name + "-" + util.K8SSanitize(clk.OperatorKey.Name),
		Flavor:          "x1.medium",
		Operator:        util.K8SSanitize(clk.OperatorKey.Name),
		Location:        "bonn",
		Region:          "eu-central-1",
		Zone:            "eu-central-1c",
		RootLB:          rootLB.Name,
		AgentImage:      "registry.mobiledgex.net:5000/mobiledgex/mexosagent",
		Kind:            "mex-platform",
		ExternalNetwork: oscli.GetMEXExternalNetwork(),
		NetworkScheme:   "priv-subnet,mex-k8s-net-1,10.101.X.0/24",
		DNSZone:         "mobiledgex.net",
		ExternalRouter:  "mex-k8s-router-1",
		Options:         "dhcp",
	}
	mf, err := templateUnmarshal(&data, yamlMEXPlatform)
	if err != nil {
		return nil, err
	}
	return mf, nil
}

//MEXPlatformCleanCloudletKey calls MEXPlatformClean with templated manifest
func MEXPlatformCleanCloudletKey(rootLB *MEXRootLB, cloudletKeyStr string) error {
	mf, err := fillPlatformTemplateCloudletKey(rootLB, cloudletKeyStr)
	if err != nil {
		return err
	}
	return MEXPlatformCleanManifest(mf)
}

func templateUnmarshal(data *templateFill, yamltext string) (*Manifest, error) {
	//log.DebugLog(log.DebugLevelMexos, "template unmarshal", "yamltext", string, "data", data)
	tmpl, err := template.New("mex").Parse(yamltext)
	if err != nil {
		return nil, fmt.Errorf("can't create template for, %v", err)
	}
	var outbuffer bytes.Buffer
	err = tmpl.Execute(&outbuffer, data)
	if err != nil {
		//log.DebugLog(log.DebugLevelMexos, "template data", "data", data)
		return nil, fmt.Errorf("can't execute template, %v", err)
	}
	mf := &Manifest{}
	err = yaml.Unmarshal(outbuffer.Bytes(), mf)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "error yaml unmarshal, templated data", "templated buffer data", outbuffer.String())
		return nil, fmt.Errorf("can't unmarshal templated data, %v", err)
	}
	return mf, nil
}

func addPorts(mf *Manifest, appInst *edgeproto.AppInst) error {
	for ii, _ := range appInst.MappedPorts {
		port := &appInst.MappedPorts[ii]
		if mf.Spec.Ports == nil {
			mf.Spec.Ports = make([]PortDetail, 0)
		}
		mexproto, ok := dme.LProto_name[int32(port.Proto)]
		if !ok {
			return fmt.Errorf("invalid LProto %d", port.Proto)
		}
		proto := "UDP"
		if port.Proto != dme.LProto_LProtoUDP {
			proto = "TCP"
		}
		p := PortDetail{
			Name:         fmt.Sprintf("%s%d", strings.ToLower(mexproto), port.InternalPort),
			MexProto:     mexproto,
			Proto:        proto,
			InternalPort: int(port.InternalPort),
			PublicPort:   int(port.PublicPort),
			PublicPath:   port.PublicPath,
		}
		mf.Spec.Ports = append(mf.Spec.Ports, p)
	}
	return nil
}

func checkEnvironment() error {
	cfkey := os.Getenv("MEX_CF_KEY")
	if cfkey == "" {
		return fmt.Errorf("missing MEX_CF_KEY")
	}
	cfuser := os.Getenv("MEX_CF_USER")
	if cfuser == "" {
		return fmt.Errorf("missing MEX_CF_USER")
	}
	dockerregpass := os.Getenv("MEX_DOCKER_REG_PASS")
	if dockerregpass == "" {
		return fmt.Errorf("missing MEX_DOCKER_REG_PASS")
	}
	return nil
}

//MEXPlatformInitManifest initializes platform
func MEXPlatformInitManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "init platform")
	err := checkEnvironment()
	if err != nil {
		return err
	}
	switch mf.Metadata.Operator {
	case "gcp":
		return nil //nothing to do
	case "azure":
		return nil //nothing to do
	default:
		if err = MEXCheckEnvVars(mf); err != nil {
			return err
		}
		//TODO validate all mf content against platform data
		if err = RunMEXAgentManifest(mf); err != nil {
			return err
		}
	}
	return nil
}

//MEXPlatformCleanManifest cleans up the platform
func MEXPlatformCleanManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "clean platform")
	err := checkEnvironment()
	if err != nil {
		return err
	}
	switch mf.Metadata.Operator {
	case "gcp":
		return nil //nothing to do
	case "azure":
		return nil
	default:
		if err = MEXCheckEnvVars(mf); err != nil {
			return err
		}
		if err = RemoveMEXAgentManifest(mf); err != nil {
			return err
		}
	}
	return nil
}

var yamlMEXApp = `apiVersion: v1
kind: {{.ResourceKind}}
resource: {{.Resource}}
metadata:
  kind: {{.Kind}}
  name: {{.Name}}
  tags: {{.Tags}}
  tenant: {{.Tenant}}
  operator: {{.Operator}}
  dnszone: {{.DNSZone}}
config:
  kind:
  source:
  detail:
    resources: {{.ConfigDetailResources}}
    deployment: {{.ConfigDetailDeployment}}
spec:
  flags: {{.Flags}}
  key: {{.Key}}
  rootlb: {{.RootLB}}
  image: {{.Image}}
  imagetype: {{.ImageType}}
  imageflavor: {{.ImageFlavor}}
  proxypath: {{.ProxyPath}}
  flavor: {{.Flavor}}
  uri: {{.AppURI}}
  ipaccess: {{.IpAccess}}
  networkscheme: {{.NetworkScheme}}
  kubemanifesttemplate: {{.AppTemplate}}
  command:
{{- range .Command}}
  - {{.}}
{{- end}}
`

//MEXAppCreateAppInst creates app inst with templated manifest
func MEXAppCreateAppInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst, app *edgeproto.App) error {
	log.DebugLog(log.DebugLevelMexos, "mex create app inst", "rootlb", rootLB, "clusterinst", clusterInst, "appinst", appInst)
	mf, err := fillAppTemplate(rootLB, appInst, app, clusterInst)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "fillAppTemplate error", "error", err)
		return err
	}
	return MEXAppCreateAppManifest(mf)
}

//MEXAppDeleteAppInst deletes app with templated manifest
func MEXAppDeleteAppInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst, app *edgeproto.App) error {
	log.DebugLog(log.DebugLevelMexos, "mex delete app inst", "rootlb", rootLB, "clusterinst", clusterInst, "appinst", appInst)
	mf, err := fillAppTemplate(rootLB, appInst, app, clusterInst)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "fillAppTemplate error", "error", err)
		return err
	}
	return MEXAppDeleteAppManifest(mf)
}

func fillAppTemplate(rootLB *MEXRootLB, appInst *edgeproto.AppInst, app *edgeproto.App, clusterInst *edgeproto.ClusterInst) (*Manifest, error) {
	var data templateFill
	var err error
	var mf *Manifest
	log.DebugLog(log.DebugLevelMexos, "fill app template", "appinst", appInst, "clusterInst", clusterInst)
	imageType, ok := edgeproto.ImageType_name[int32(app.ImageType)]
	if !ok {
		return nil, fmt.Errorf("cannot find imagetype in map")
	}
	if clusterInst.Flavor.Name == "" {
		return nil, fmt.Errorf("will not fill app template, invalid cluster flavor name")
	}
	if verr := validateClusterKind(clusterInst.Key.CloudletKey.OperatorKey.Name); verr != nil {
		return nil, verr
	}
	if appInst.Key.AppKey.Name == "" {
		return nil, fmt.Errorf("will not fill app template, invalid appkey name")
	}
	ipAccess, ok := edgeproto.IpAccess_name[int32(appInst.IpAccess)]
	if !ok {
		return nil, fmt.Errorf("cannot find ipaccess in map")
	}
	if len(appInst.Key.AppKey.Name) < 3 {
		log.DebugLog(log.DebugLevelMexos, "warning, very short appkey name", "name", appInst.Key.AppKey.Name)
	}
	config, err := cloudcommon.ParseAppConfig(app.Config)
	if err != nil {
		return nil, fmt.Errorf("error parsing appinst config %s, %v", app.Config, err)
	}
	log.DebugLog(log.DebugLevelMexos, "appinst config", "config", config)
	appDeploymentType := app.Deployment
	if err != nil {
		return nil, err
	}
	log.DebugLog(log.DebugLevelMexos, "app deploying", "imageType", imageType, "deploymentType", appDeploymentType)
	if !cloudcommon.IsValidDeploymentForImage(app.ImageType, appDeploymentType) {
		return nil, fmt.Errorf("deployment is not valid for image type")
	}
	switch appDeploymentType {
	case cloudcommon.AppDeploymentTypeKubernetes:
		data = templateFill{
			ResourceKind:           "application",
			Resource:               util.K8SSanitize(appInst.Key.AppKey.Name),
			Kind:                   "kubernetes",
			Name:                   util.K8SSanitize(appInst.Key.AppKey.Name),
			Tags:                   util.K8SSanitize(appInst.Key.AppKey.Name) + "-kubernetes-tag",
			Key:                    clusterInst.Key.ClusterKey.Name,
			Tenant:                 util.K8SSanitize(appInst.Key.AppKey.Name) + "-tenant",
			DNSZone:                "mobiledgex.net",
			Operator:               util.K8SSanitize(clusterInst.Key.CloudletKey.OperatorKey.Name),
			RootLB:                 rootLB.Name,
			Image:                  app.ImagePath,
			ImageType:              imageType,
			ImageFlavor:            appInst.Flavor.Name,
			ProxyPath:              util.K8SSanitize(appInst.Key.AppKey.Name),
			AppURI:                 appInst.Uri,
			IpAccess:               ipAccess,
			ConfigDetailDeployment: app.Deployment,
			ConfigDetailResources:  config.Resources,
			Command:                strings.Split(app.Command, " "),
		}
		mf, err = templateUnmarshal(&data, yamlMEXApp)
		if err != nil {
			return nil, err
		}
	case cloudcommon.AppDeploymentTypeDocker:
		data = templateFill{
			ResourceKind:           "application",
			Resource:               util.K8SSanitize(appInst.Key.AppKey.Name),
			Kind:                   "docker",
			Name:                   util.K8SSanitize(appInst.Key.AppKey.Name),
			Tags:                   util.K8SSanitize(appInst.Key.AppKey.Name) + "-docker-tag",
			Key:                    clusterInst.Key.ClusterKey.Name,
			Tenant:                 util.K8SSanitize(appInst.Key.AppKey.Name) + "-tenant",
			DNSZone:                "mobiledgex.net",
			Operator:               util.K8SSanitize(clusterInst.Key.CloudletKey.OperatorKey.Name),
			RootLB:                 rootLB.Name,
			Image:                  app.ImagePath,
			ImageType:              imageType,
			ImageFlavor:            appInst.Flavor.Name,
			ProxyPath:              util.K8SSanitize(appInst.Key.AppKey.Name),
			AppURI:                 appInst.Uri,
			IpAccess:               ipAccess,
			ConfigDetailDeployment: app.Deployment,
			ConfigDetailResources:  config.Resources,
			Command:                strings.Split(app.Command, " "),
		}
		mf, err = templateUnmarshal(&data, yamlMEXApp)
		if err != nil {
			return nil, err
		}
	case cloudcommon.AppDeploymentTypeKVM:
		data = templateFill{
			ResourceKind:           "application",
			Resource:               appInst.Key.AppKey.Name,
			Kind:                   "kvm",
			Name:                   appInst.Key.AppKey.Name,
			Tags:                   appInst.Key.AppKey.Name + "-qcow-tag",
			Key:                    clusterInst.Key.ClusterKey.Name,
			Tenant:                 appInst.Key.AppKey.Name + "-tenant",
			Operator:               clusterInst.Key.CloudletKey.OperatorKey.Name,
			RootLB:                 rootLB.Name,
			Image:                  app.ImagePath,
			ImageFlavor:            appInst.Flavor.Name,
			DNSZone:                "mobiledgex.net",
			ImageType:              imageType,
			ProxyPath:              appInst.Key.AppKey.Name,
			AppURI:                 appInst.Uri,
			NetworkScheme:          "external-ip," + oscli.GetMEXExternalNetwork(),
			ConfigDetailDeployment: app.Deployment,
			ConfigDetailResources:  config.Resources,
		}
		mf, err = templateUnmarshal(&data, yamlMEXApp)
		if err != nil {
			return nil, err
		}
	case cloudcommon.AppDeploymentTypeHelm:
		data = templateFill{
			ResourceKind:           "application",
			Resource:               appInst.Key.AppKey.Name,
			Kind:                   "helm",
			Name:                   appInst.Key.AppKey.Name,
			Tags:                   appInst.Key.AppKey.Name + "-helm-tag",
			Key:                    clusterInst.Key.ClusterKey.Name,
			Tenant:                 appInst.Key.AppKey.Name + "-tenant",
			Operator:               clusterInst.Key.CloudletKey.OperatorKey.Name,
			RootLB:                 rootLB.Name,
			Image:                  app.ImagePath,
			ImageFlavor:            appInst.Flavor.Name,
			DNSZone:                "mobiledgex.net",
			ImageType:              imageType,
			ProxyPath:              appInst.Key.AppKey.Name,
			AppURI:                 appInst.Uri,
			ConfigDetailDeployment: app.Deployment,
			ConfigDetailResources:  config.Resources,
		}
		mf, err = templateUnmarshal(&data, yamlMEXApp)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown image type %s", imageType)
	}
	mf.Config.ConfigDetail.Manifest = app.DeploymentManifest
	log.DebugLog(log.DebugLevelMexos, "filled app manifest")
	err = addPorts(mf, appInst)
	if err != nil {
		return nil, err
	}
	log.DebugLog(log.DebugLevelMexos, "added port to app manifest")
	return mf, nil
}

//MEXAppCreateAppManifest creates app instances on the cluster platform
func MEXAppCreateAppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "create app from manifest")
	appDeploymentType := mf.Config.ConfigDetail.Deployment
	log.DebugLog(log.DebugLevelMexos, "app deployment", "imageType", mf.Spec.ImageType, "deploymentType", appDeploymentType, "config", mf.Config)
	kubeManifest, err := cloudcommon.GetDeploymentManifest(mf.Config.ConfigDetail.Manifest)
	if err != nil {
		return err
	}
	switch mf.Metadata.Operator {
	case "gcp":
		fallthrough
	case "azure":
		if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
			return runKubectlCreateApp(mf, kubeManifest)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeKVM {
			return fmt.Errorf("not yet supported")
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
			return fmt.Errorf("not yet supported")
		}
		return fmt.Errorf("unknown deployment type %s", appDeploymentType)
	default:
		if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
			return CreateKubernetesAppManifest(mf, kubeManifest)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeKVM {
			return CreateQCOW2AppManifest(mf)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
			return CreateHelmAppManifest(mf)
		}
		return fmt.Errorf("unknown deployment type %s", appDeploymentType)
	}
}

func getKconf(mf *Manifest, createIfMissing bool) (string, error) {
	name := mexEnv["MEX_DIR"] + "/" + mf.Spec.Key + ".kubeconfig"
	log.DebugLog(log.DebugLevelMexos, "get kubeconfig name", "name", name)

	if createIfMissing {
		if _, err := os.Stat(name); os.IsNotExist(err) {
			// if kubeconfig does not exist, optionally create it.  It is possible it was
			// created on a different container or we had a restart of the container
			log.DebugLog(log.DebugLevelMexos, "creating missing kconf file", "name", name)
			switch mf.Metadata.Operator {
			case "gcp":
				if err = gcloud.GetGKECredentials(mf.Metadata.Name); err != nil {
					return "", fmt.Errorf("unable to get GKE credentials %v", err)
				}
				if err = copyFile(defaultKubeconfig(), name); err != nil {
					return "", fmt.Errorf("can't copy %s, %v", defaultKubeconfig(), err)
				}
			case "azure":
				if err = azure.GetAKSCredentials(mf.Metadata.ResourceGroup, mf.Metadata.Name); err != nil {
					return "", fmt.Errorf("unable to get AKS credentials %v", err)
				}
				if err = copyFile(defaultKubeconfig(), name); err != nil {
					return "", fmt.Errorf("can't copy %s, %v", defaultKubeconfig(), err)
				}
			}
		}
	}
	return name, nil
}

type ingressItem struct {
	IP string `json:"ip"`
}

type loadBalancerItem struct {
	Ingresses []ingressItem `json:"ingress"`
}

type statusItem struct {
	LoadBalancer loadBalancerItem `json:"loadBalancer"`
}

type metadataItem struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	CreationTimestamp string `json:"creationTimestamp"`
	ResourceVersion   string `json:"resourceVersion"`
	UID               string `json:"uid"`
}

type svcItem struct {
	APIVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Metadata   metadataItem `json:"metadata"`
	Spec       interface{}  `json:"spec"`
	Status     statusItem   `json:"status"`
}

type svcItems struct {
	Items []svcItem `json:"items"`
}

func runKubectlCreateApp(mf *Manifest, kubeManifest string) error {
	log.DebugLog(log.DebugLevelMexos, "run kubectl create app", "kubeManifest", kubeManifest)
	kconf, err := getKconf(mf, false)
	if err != nil {
		return fmt.Errorf("error creating app due to kconf %v, %v", mf, err)
	}
	out, err := sh.Command("kubectl", "create", "secret", "docker-registry", "mexregistrysecret", "--docker-server="+mexEnv["MEX_DOCKER_REGISTRY"], "--docker-username=mobiledgex", "--docker-password="+mexEnv["MEX_DOCKER_REG_PASS"], "--docker-email=docker@mobiledgex.com", "--kubeconfig="+kconf).CombinedOutput()
	if err != nil {
		if !strings.Contains(string(out), "AlreadyExists") {
			return fmt.Errorf("can't add docker registry secret, %s, %v", out, err)
		}
	}
	kfile := mf.Metadata.Name + ".yaml"
	err = writeKubeManifest(kubeManifest, kfile)
	if err != nil {
		return err
	}
	defer os.Remove(kfile)

	out, err = sh.Command("kubectl", "create", "-f", kfile, "--kubeconfig="+kconf).Output()
	if err != nil {
		return fmt.Errorf("error creating app, %s, %v, %v", out, err, mf)
	}
	err = createAppDNS(mf, kconf)
	if err != nil {
		return fmt.Errorf("error creating dns entry for app, %v, %v", err, mf)
	}
	return nil
}

func isDomainName(s string) bool {
	l := len(s)
	if l == 0 || l > 254 || l == 254 && s[l-1] != '.' {
		return false
	}

	last := byte('.')
	ok := false // Ok once we've seen a letter.
	partlen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			ok = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}
	if last == '-' || partlen > 63 {
		return false
	}

	return ok
}

func validateURI(uri string) error {
	if isDomainName(uri) {
		return nil
	}
	fqdn := uri2fqdn(uri)
	if isDomainName(fqdn) {
		return nil
	}
	return fmt.Errorf("URI %s is not a valid domain name", uri)
}

func createAppDNS(mf *Manifest, kconf string) error {
	if mf.Metadata.Operator != "gcp" && mf.Metadata.Operator != "azure" {
		return fmt.Errorf("error, invalid code path")
	}
	if err := CheckCredentialsCF(); err != nil {
		return err
	}
	if err := cloudflare.InitAPI(mexEnv["MEX_CF_USER"], mexEnv["MEX_CF_KEY"]); err != nil {
		return fmt.Errorf("cannot init cloudflare api, %v", err)
	}
	if mf.Spec.URI == "" {
		return fmt.Errorf("URI not specified %v", mf)
	}
	err := validateURI(mf.Spec.URI)
	if err != nil {
		return err
	}
	if mf.Metadata.DNSZone == "" {
		return fmt.Errorf("missing DNS zone, metadata %v", mf.Metadata)
	}
	serviceNames, err := getSvcNames(mf.Metadata.Name, kconf)
	if err != nil {
		return err
	}
	if len(serviceNames) < 1 {
		return fmt.Errorf("no service names starting with %s", mf.Metadata.Name)
	}
	recs, derr := cloudflare.GetDNSRecords(mf.Metadata.DNSZone)
	if derr != nil {
		return fmt.Errorf("error getting dns records for %s, %v", mf.Metadata.DNSZone, err)
	}
	fqdnBase := uri2fqdn(mf.Spec.URI)
	for _, sn := range serviceNames {
		externalIP, err := getSvcExternalIP(sn, kconf)
		if err != nil {
			return err
		}
		fqdn := cloudcommon.ServiceFQDN(sn, fqdnBase)
		for _, rec := range recs {
			if rec.Type == "A" && rec.Name == fqdn {
				if err := cloudflare.DeleteDNSRecord(mf.Metadata.DNSZone, rec.ID); err != nil {
					return fmt.Errorf("cannot delete existing DNS record %v, %v", rec, err)
				}
				log.DebugLog(log.DebugLevelMexos, "deleted DNS record", "name", fqdn)
			}
		}
		if err := cloudflare.CreateDNSRecord(mf.Metadata.DNSZone, fqdn, "A", externalIP, 1, false); err != nil {
			return fmt.Errorf("can't create DNS record for %s,%s, %v", fqdn, externalIP, err)
		}
		//log.DebugLog(log.DebugLevelMexos, "waiting for DNS record to be created on cloudflare...")
		//err = WaitforDNSRegistration(fqdn)
		//if err != nil {
		//	return err
		//}
		log.DebugLog(log.DebugLevelMexos, "registered DNS name, may still need to wait for propagation", "name", fqdn, "externalIP", externalIP)
	}
	return nil
}

func deleteAppDNS(mf *Manifest, kconf string) error {
	if mf.Metadata.Operator != "gcp" && mf.Metadata.Operator != "azure" {
		return fmt.Errorf("error, invalid code path")
	}
	if err := CheckCredentialsCF(); err != nil {
		return err
	}
	if err := cloudflare.InitAPI(mexEnv["MEX_CF_USER"], mexEnv["MEX_CF_KEY"]); err != nil {
		return fmt.Errorf("cannot init cloudflare api, %v", err)
	}
	if mf.Spec.URI == "" {
		return fmt.Errorf("URI not specified %v", mf)
	}
	err := validateURI(mf.Spec.URI)
	if err != nil {
		return err
	}
	if mf.Metadata.DNSZone == "" {
		return fmt.Errorf("missing DNS zone, metadata %v", mf.Metadata)
	}
	serviceNames, err := getSvcNames(mf.Metadata.Name, kconf)
	if err != nil {
		return err
	}
	if len(serviceNames) < 1 {
		return fmt.Errorf("no service names starting with %s", mf.Metadata.Name)
	}
	recs, derr := cloudflare.GetDNSRecords(mf.Metadata.DNSZone)
	if derr != nil {
		return fmt.Errorf("cannot get dns records for dns zone %s, error %v", mf.Metadata.DNSZone, err)
	}
	fqdnBase := uri2fqdn(mf.Spec.URI)
	for _, sn := range serviceNames {
		fqdn := cloudcommon.ServiceFQDN(sn, fqdnBase)
		for _, rec := range recs {
			if rec.Type == "A" && rec.Name == fqdn {
				if err := cloudflare.DeleteDNSRecord(mf.Metadata.DNSZone, rec.ID); err != nil {
					return fmt.Errorf("cannot delete existing DNS record %v, %v", rec, err)
				}
			}
		}
		log.DebugLog(log.DebugLevelMexos, "deleted DNS name", "name", fqdn)
	}
	return nil
}

func uri2fqdn(uri string) string {
	fqdn := strings.Replace(uri, "http://", "", 1)
	fqdn = strings.Replace(fqdn, "https://", "", 1)
	//XXX assumes no trailing elements
	return fqdn
}

func getSvcNames(name string, kconf string) ([]string, error) {
	out, err := sh.Command("kubectl", "get", "svc", "--kubeconfig="+kconf, "-o", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("error getting svc %s, %s, %v", name, out, err)
	}
	svcs := &svcItems{}
	err = json.Unmarshal(out, svcs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling svc json, %v", err)
	}
	var serviceNames []string
	for _, item := range svcs.Items {
		if strings.HasPrefix(item.Metadata.Name, name) {
			serviceNames = append(serviceNames, item.Metadata.Name)
		}
	}
	log.DebugLog(log.DebugLevelMexos, "service names", "names", serviceNames)
	return serviceNames, nil
}

func getSvcExternalIP(name string, kconf string) (string, error) {
	log.DebugLog(log.DebugLevelMexos, "get service external IP", "name", name)
	externalIP := ""
	var out []byte
	var err error
	//wait for Load Balancer to assign external IP address. It takes a variable amount of time.
	for i := 0; i < 100; i++ {
		out, err = sh.Command("kubectl", "get", "svc", "--kubeconfig="+kconf, "-o", "json").Output()
		if err != nil {
			return "", fmt.Errorf("error getting svc %s, %s, %v", name, out, err)
		}
		svcs := &svcItems{}
		err = json.Unmarshal(out, svcs)
		if err != nil {
			return "", fmt.Errorf("error unmarshalling svc json, %v", err)
		}
		log.DebugLog(log.DebugLevelMexos, "getting externalIP, examine list of services", "name", name, "svcs", svcs)
		for _, item := range svcs.Items {
			log.DebugLog(log.DebugLevelMexos, "svc item", "item", item, "name", name)
			if item.Metadata.Name != name {
				continue
			}
			for _, ingress := range item.Status.LoadBalancer.Ingresses {
				if ingress.IP != "" {
					externalIP = ingress.IP
					log.DebugLog(log.DebugLevelMexos, "got externaIP for app", "externalIP", externalIP)
					return externalIP, nil
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
	if externalIP == "" {
		return "", fmt.Errorf("timed out trying to get externalIP")
	}
	return externalIP, nil
}

//MEXAppDeleteManifest kills app
func MEXAppDeleteAppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "delete app with manifest")
	appDeploymentType := mf.Config.ConfigDetail.Deployment
	log.DebugLog(log.DebugLevelMexos, "app delete", "imageType", mf.Spec.ImageType, "deploymentType", appDeploymentType, "config", mf.Config)
	kubeManifest, err := cloudcommon.GetDeploymentManifest(mf.Config.ConfigDetail.Manifest)
	if err != nil {
		return err
	}
	switch mf.Metadata.Operator {
	case "gcp":
		fallthrough
	case "azure":
		if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
			return runKubectlDeleteApp(mf, kubeManifest)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeKVM {
			return fmt.Errorf("not yet supported")
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
			return fmt.Errorf("not yet supported")
		}
		return fmt.Errorf("unknown image type %s", appDeploymentType)
	default:
		if appDeploymentType == cloudcommon.AppDeploymentTypeKubernetes {
			return DeleteKubernetesAppManifest(mf, kubeManifest)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeKVM {
			return DeleteQCOW2AppManifest(mf)
		} else if appDeploymentType == cloudcommon.AppDeploymentTypeHelm {
			return DeleteHelmAppManifest(mf)
		}
		return fmt.Errorf("unknown image type %s", mf.Spec.ImageType)
	}
}

func runKubectlDeleteApp(mf *Manifest, kubeManifest string) error {
	if err := CheckCredentialsCF(); err != nil {
		return err
	}
	if err := cloudflare.InitAPI(mexEnv["MEX_CF_USER"], mexEnv["MEX_CF_KEY"]); err != nil {
		return fmt.Errorf("cannot init cloudflare api, %v", err)
	}
	kconf, err := getKconf(mf, false)
	if err != nil {
		return fmt.Errorf("error deleting app due to kconf,  %v, %v", mf, err)
	}
	kfile := mf.Metadata.Name + ".yaml"
	err = writeKubeManifest(kubeManifest, kfile)
	if err != nil {
		return err
	}
	defer os.Remove(kfile)
	serviceNames, err := getSvcNames(mf.Metadata.Name, kconf)
	if err != nil {
		return err
	}
	if len(serviceNames) < 1 {
		return fmt.Errorf("no service names starting with %s", mf.Metadata.Name)
	}
	out, err := sh.Command("kubectl", "delete", "-f", kfile, "--kubeconfig="+kconf).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error deleting app, %s, %v, %v", out, mf, err)
	}
	if mf.Metadata.DNSZone == "" {
		return fmt.Errorf("missing dns zone, metadata %v", mf.Metadata)
	}
	fqdnBase := uri2fqdn(mf.Spec.URI)
	dr, err := cloudflare.GetDNSRecords(mf.Metadata.DNSZone)
	if err != nil {
		return fmt.Errorf("cannot get dns records for %s, %v", mf.Metadata.DNSZone, err)
	}
	for _, sn := range serviceNames {
		fqdn := cloudcommon.ServiceFQDN(sn, fqdnBase)
		for _, d := range dr {
			if d.Type == "A" && d.Name == fqdn {
				if err := cloudflare.DeleteDNSRecord(mf.Metadata.DNSZone, d.ID); err != nil {
					return fmt.Errorf("cannot delete DNS record, %v", d)
				}
				log.DebugLog(log.DebugLevelMexos, "deleted DNS record", "name", fqdn)
			}
		}
	}
	return nil
}

func validateClusterKind(kind string) error {
	log.DebugLog(log.DebugLevelMexos, "cluster kind", "kind", kind)
	for _, k := range []string{"gcp", "azure"} {
		if kind == k {
			return nil
		}
	}
	if strings.HasPrefix(kind, "mex-") {
		return nil
	}
	log.DebugLog(log.DebugLevelMexos, "warning, cluster kind, operator has no mex- prefix", "kind", kind)
	return nil
}

func writeKubeManifest(kubeManifest string, filename string) error {
	outFile, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open k8s deployment file %s: %s", filename, err.Error())
	}
	_, err = outFile.WriteString(kubeManifest)
	if err != nil {
		outFile.Close()
		os.Remove(filename)
		return fmt.Errorf("unable to write k8s deployment file %s: %s", filename, err.Error())
	}
	outFile.Sync()
	outFile.Close()
	return nil
}
