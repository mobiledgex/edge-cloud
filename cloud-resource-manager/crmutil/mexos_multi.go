package crmutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/azure"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/gcloud"
	"github.com/mobiledgex/edge-cloud-infra/openstack-tenant/agent/cloudflare"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

var yamlMEXCluster = `apiVersion: v1
kind: cluster
resource: fmbncisrs101.tacn.detemobil.de:5000/v2.0
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
	Name, Kind, Flavor, Tags, Tenant, Region, Zone, DNSZone                         string
	ImageFlavor, Location, RootLB, ResourceGroup                                    string
	StorageSpec, NetworkScheme, MasterFlavor, Topology                              string
	NodeFlavor, Operator, Key, Image, Options                                       string
	ImageType, AppURI, ProxyPath                                                    string
	ExternalNetwork, Project, AppTemplate                                           string
	ExternalRouter, Flags, IpAccess                                                 string
	NumMasters, NumNodes                                                            int
	ConfigDetailDeployment, ConfigDetailResources, ConfigKind, ConfigDetailManifest string
	Command                                                                         []string
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
		Name:          clusterInst.Key.ClusterKey.Name,
		Tags:          clusterInst.Key.ClusterKey.Name + "-tag",
		Tenant:        clusterInst.Key.ClusterKey.Name + "-tenant",
		Operator:      util.K8SSanitize(clusterInst.Key.CloudletKey.OperatorKey.Name),
		Key:           clusterInst.Key.ClusterKey.Name,
		Kind:          clusterInst.Flavor.Name,
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
	log.DebugLog(log.DebugLevelMexos, "creating cluster via manifest", "mf", mf)
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
		log.DebugLog(log.DebugLevelMexos, "created kubernetes cluster", "mf", mf)
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
	if err = copy(defaultKubeconfig(), kconf); err != nil {
		return fmt.Errorf("can't copy %s, %v", defaultKubeconfig(), err)
	}
	log.DebugLog(log.DebugLevelMexos, "created aks", "name", mf.Spec.Key)
	return nil
}

func defaultKubeconfig() string {
	return os.Getenv("HOME") + "/.kube/config"
}

func copy(src string, dst string) error {
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
	if err = copy(defaultKubeconfig(), kconf); err != nil {
		return fmt.Errorf("can't copy %s, %v", defaultKubeconfig(), err)
	}
	log.DebugLog(log.DebugLevelMexos, "created gke", "name", mf.Spec.Key)
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
kind: flavor
resource: fmbncisrs101.tacn.detemobil.de:5000/v2.0
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
func MEXAddFlavorClusterInst(flavor *edgeproto.ClusterFlavor) error {
	log.DebugLog(log.DebugLevelMexos, "adding cluster inst flavor", "flavor", flavor)

	if flavor.Key.Name == "" {
		log.DebugLog(log.DebugLevelMexos, "cannot add empty cluster inst flavor", "flavor", flavor)
		return fmt.Errorf("will not add empty cluster inst %v", flavor)
	}
	data := templateFill{
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
	log.DebugLog(log.DebugLevelMexos, "add mex flavor", "mf", mf)
	//TODO use full manifest and validate against platform data
	return AddFlavorManifest(mf)
}

// TODO DeleteFlavor -- but almost never done

// TODO lookup guid using name

//MEXClusterRemoveManifest removes a cluster
func MEXClusterRemoveManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "removing cluster", "mf", mf)
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
kind: platform
resource: fmbncisrs101.tacn.detemobil.de:5000/v2.0
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
    image: {{.Image}}
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
	data := templateFill{
		Name:            clk.Name,
		Tags:            clk.Name + "-tag",
		Key:             clk.Name + "-" + util.K8SSanitize(clk.OperatorKey.Name),
		Flavor:          "x1.medium",
		Operator:        util.K8SSanitize(clk.OperatorKey.Name),
		Location:        "bonn",
		Region:          "eu-central-1",
		Zone:            "eu-central-1c",
		RootLB:          rootLB.Name,
		Image:           "registry.mobiledgex.net:5000/mobiledgex/mexosagent",
		Kind:            "mex-platform",
		ExternalNetwork: "external-network-shared",
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
	log.DebugLog(log.DebugLevelMexos, "init platform", "mf", mf)
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
	log.DebugLog(log.DebugLevelMexos, "clean platform", "mf", mf)
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

var yamlMEXAppKubernetes = `apiVersion: v1
kind: application
resource: fmbncisrs101.tacn.detemobil.de:5000/v2.0
metadata:
  kind: {{.Kind}}
  name: {{.Name}}
  tags: {{.Tags}}
  tenant: {{.Tenant}}
  operator: {{.Operator}}
  dnszone: {{.DNSZone}}
config:
  kind: {{.ConfigKind}}
  source:
  detail:
    resources: {{.ConfigDetailResources}}
    deployment: {{.ConfigDetailDeployment}}
    manifest: {{.ConfigDetailManifest}}
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

var yamlMEXAppQcow2 = `apiVersion: v1
kind: mex-kvm-application
resource: openstack
metadata:
  kind: {{.Kind}}
  name: {{.Name}}
  tags: {{.Tags}}
  tenant: {{.Tenant}}
  operator: {{.Operator}}
  dnszone: {{.DNSZone}}
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
`

//MEXAppCreateAppInst creates app inst with templated manifest
func MEXAppCreateAppInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst) error {
	log.DebugLog(log.DebugLevelMexos, "mex create app inst", "rootlb", rootLB, "clusterinst", clusterInst, "appinst", appInst)
	mf, err := fillAppTemplate(rootLB, appInst, clusterInst)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "fillAppTemplate error", "error", err)
		return err
	}
	return MEXAppCreateAppManifest(mf)
}

//MEXAppDeleteAppInst deletes app with templated manifest
func MEXAppDeleteAppInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst) error {
	log.DebugLog(log.DebugLevelMexos, "mex delete app inst", "rootlb", rootLB, "clusterinst", clusterInst, "appinst", appInst)
	mf, err := fillAppTemplate(rootLB, appInst, clusterInst)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "fillAppTemplate error", "error", err)
		return err
	}
	return MEXAppDeleteAppManifest(mf)
}

var validDeployments = []string{"kubernetes", "kvm"} // TODO "docker", ...

func isValidDeploymentType(appDeploymentType string) bool {
	for _, d := range validDeployments {
		if appDeploymentType == d {
			return true
		}
	}
	return false
}

func fillAppTemplate(rootLB *MEXRootLB, appInst *edgeproto.AppInst, clusterInst *edgeproto.ClusterInst) (*Manifest, error) {
	var data templateFill
	var err error
	var mf *Manifest
	log.DebugLog(log.DebugLevelMexos, "fill app template", "appinst", appInst, "clusterInst", clusterInst)
	imageType, ok := edgeproto.ImageType_name[int32(appInst.ImageType)]
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
	config, err := ParseAppInstConfig(appInst.Config)
	if err != nil {
		return nil, fmt.Errorf("error parsing appinst config %s, %v", appInst.Config, err)
	}
	log.DebugLog(log.DebugLevelMexos, "appinst config", "config", config)
	appDeploymentType := ""
	switch imageType {
	case "ImageTypeDocker":
		appDeploymentType = "kubernetes"
	case "ImageTypeQCOW":
		appDeploymentType = "kvm"
	default:
		return nil, fmt.Errorf("unknown image type %s", imageType)
	}
	if config.Kind == "config" {
		appDeploymentType = config.ConfigDetail.Deployment
	}
	if !isValidDeploymentType(appDeploymentType) {
		return nil, fmt.Errorf("invalid deployment type, '%s'", appDeploymentType)
	}
	log.DebugLog(log.DebugLevelMexos, "app deploying", "imageType", imageType, "deploymentType", appDeploymentType)
	switch appDeploymentType {
	case "kubernetes":
		if imageType != "ImageTypeDocker" {
			return nil, fmt.Errorf("invalid image type %s for deployment type %s", imageType, appDeploymentType)
		}
		data = templateFill{
			Kind:                   clusterInst.Flavor.Name,
			Name:                   util.K8SSanitize(appInst.Key.AppKey.Name),
			Tags:                   util.K8SSanitize(appInst.Key.AppKey.Name) + "-kubernetes-tag",
			Key:                    clusterInst.Key.ClusterKey.Name,
			Tenant:                 util.K8SSanitize(appInst.Key.AppKey.Name) + "-tenant",
			DNSZone:                "mobiledgex.net",
			Operator:               util.K8SSanitize(clusterInst.Key.CloudletKey.OperatorKey.Name),
			RootLB:                 rootLB.Name,
			Image:                  appInst.ImagePath,
			ImageType:              imageType,
			ImageFlavor:            appInst.Flavor.Name,
			ProxyPath:              util.K8SSanitize(appInst.Key.AppKey.Name),
			AppURI:                 appInst.Uri,
			IpAccess:               ipAccess,
			AppTemplate:            appInst.AppTemplate,
			ConfigKind:             config.Kind,
			ConfigDetailDeployment: config.ConfigDetail.Deployment,
			ConfigDetailResources:  config.ConfigDetail.Resources,
			ConfigDetailManifest:   config.ConfigDetail.Manifest,
		}
		if config.Kind == "command" {
			data.Command = config.Command
		}
		mf, err = templateUnmarshal(&data, yamlMEXAppKubernetes)
		if err != nil {
			return nil, err
		}
	//case "docker":
	//	if imageType != "ImageTypeDocker" {
	//		return nil, fmt.Errorf("invalid image type %s for deployment type %s", imageType, appDeploymentType)
	//	}
	case "kvm":
		if imageType != "ImageTypeQCOW" {
			return nil, fmt.Errorf("invalid image type %s for deployment type %s", imageType, appDeploymentType)
		}
		data = templateFill{
			Kind:          clusterInst.Flavor.Name,
			Name:          appInst.Key.AppKey.Name,
			Tags:          appInst.Key.AppKey.Name + "-qcow-tag",
			Key:           clusterInst.Key.ClusterKey.Name,
			Tenant:        appInst.Key.AppKey.Name + "-tenant",
			Operator:      clusterInst.Key.CloudletKey.OperatorKey.Name,
			RootLB:        rootLB.Name,
			Image:         appInst.ImagePath,
			ImageFlavor:   appInst.Flavor.Name,
			DNSZone:       "mobiledgex.net",
			ImageType:     imageType,
			ProxyPath:     appInst.Key.AppKey.Name,
			AppURI:        appInst.Uri,
			NetworkScheme: "external-ip,external-network-shared",
		}
		mf, err = templateUnmarshal(&data, yamlMEXAppQcow2)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown image type %s", imageType)
	}
	log.DebugLog(log.DebugLevelMexos, "filled app manifest", "mf", mf)
	err = addPorts(mf, appInst)
	if err != nil {
		return nil, err
	}
	log.DebugLog(log.DebugLevelMexos, "added port to app manifest", "mf", mf)
	return mf, nil
}

//XXX InternalPort used for both port and targetport
var kubeManifestSimpleShared = `apiVersion: v1
kind: Service
metadata:
  name: {{.Metadata.Name}}-service
  labels:
    run: {{.Metadata.Name}}
spec:
  type: LoadBalancer
  ports:
{{- range .Spec.Ports}}
  - name: {{.Name}}
    protocol: {{.Proto}}
    port: {{.InternalPort}}
    targetPort: {{.InternalPort}}
{{- end}}
  selector:
    run: {{.Metadata.Name}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Metadata.Name}}-deployment
spec:
  selector:
    matchLabels:
      run: {{.Metadata.Name}}
  replicas: 2
  template:
    metadata:
      labels:
        run: {{.Metadata.Name}}
    spec:
      volumes:
      imagePullSecrets:
      - name: mexregistrysecret
      containers:
      - name: {{.Metadata.Name}}
        image: {{.Spec.Image}}
        imagePullPolicy: Always
        ports:
{{- range .Spec.Ports}}
        - containerPort: {{.InternalPort}}
{{- end}}
        command:
{{- range .Spec.Command}}
        - {{.}}
{{- end}}
`

func ParseAppInstConfigDetail(conf []byte) (*AppInstConfigDetail, error) {
	confDetail := &AppInstConfigDetail{}
	err := json.Unmarshal(conf, confDetail)
	if err != nil {
		return nil, err
	}
	return confDetail, nil
}

func ParseAppInstConfig(configStr string) (*AppInstConfig, error) {
	if configStr == "" {
		return &AppInstConfig{Kind: "command", Source: "", Command: []string{}}, nil
	}
	if strings.HasPrefix(configStr, "{") {
		confDetail := &AppInstConfigDetail{}
		err := json.Unmarshal([]byte(configStr), confDetail)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal json config str, err %v, config `%s`", err, configStr)
		}
		return &AppInstConfig{Kind: "config", Source: configStr, ConfigDetail: *confDetail}, nil
	}
	if strings.HasPrefix(configStr, "http://") ||
		strings.HasPrefix(configStr, "https://") {
		resp, err := http.Get(configStr)
		if err != nil {
			return nil, fmt.Errorf("cannot get config from %s, %v", configStr, err)
		}
		defer resp.Body.Close()
		configBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read config from %s, %v", configStr, err)
		}
		confDetail, err := ParseAppInstConfigDetail(configBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot parse app inst config detail from %s, %v", configStr, err)
		}
		return &AppInstConfig{Kind: "config", Source: configStr, ConfigDetail: *confDetail}, nil
	}
	cmd, err := ParseAppInstConfigCommand(configStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse config command string %s, %v", configStr, err)
	}
	return &AppInstConfig{Kind: "command", Source: configStr, Command: cmd}, nil
}

func ParseAppInstConfigCommand(configStr string) ([]string, error) {
	// XXX  Not sure what Jon wants in terms of format of this string.
	// XXX I am preserving his split string call here.
	return strings.Split(configStr, " "), nil
}

func genKubeManifest(mf *Manifest) (string, error) {
	if mf.Spec.KubeManifestTemplate == "" {
		// Who/What generates the template, and where it comes from
		// is TBD. If no template, use a default simple template.
		if mf.Spec.IpAccess == edgeproto.IpAccess_name[int32(edgeproto.IpAccess_IpAccessShared)] {
			log.DebugLog(log.DebugLevelMexos, "using builtin simple shared kubernetes template")
			mf.Spec.KubeManifestTemplate = kubeManifestSimpleShared
		} else {
			return "", fmt.Errorf("No default template for dedicated IpAccess. Please specify template.")
		}
	}
	// Template string is assumed to be contents of template unless
	// it starts with http:// or file://
	tmplDef := mf.Spec.KubeManifestTemplate
	log.DebugLog(log.DebugLevelMexos, "genenrating kubernetes manifest from template", "template", tmplDef)
	if strings.HasPrefix(mf.Spec.KubeManifestTemplate, "http://") {
		resp, err := http.Get(tmplDef)
		if err != nil {
			return "", fmt.Errorf("can't get http template, %s", err.Error())
		}
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("can't read http template, %s", err.Error())
		}
		tmplDef = string(bytes)
	} else if strings.HasPrefix(mf.Spec.KubeManifestTemplate, "file://") {
		filename := mf.Spec.KubeManifestTemplate[7:]
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			return "", fmt.Errorf("can't read template file, %s", err.Error())
		}
		tmplDef = string(bytes)
	}
	tmpl, err := template.New("kubemanifest").Parse(tmplDef)
	if err != nil {
		return "", fmt.Errorf("unable to compile Kube Manifest Template: %s", err.Error())
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, mf)
	if err != nil {
		return "", fmt.Errorf("unable to generate k8s deployment file: %s", err.Error())
	}
	return buf.String(), err
}

func retrieveKubeManifest(manifestSpec string) (string, error) {
	if strings.HasPrefix(manifestSpec, "http://") ||
		strings.HasPrefix(manifestSpec, "https://") {
		resp, err := http.Get(manifestSpec)
		if err != nil {
			return "", fmt.Errorf("cannot get manifest from %s, %v", manifestSpec, err)
		}
		defer resp.Body.Close()
		manifestBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("cannot read manifest from %s, %v", manifestSpec, err)
		}
		return string(manifestBytes), nil
	}
	return "", fmt.Errorf("invalid manifest location %s", manifestSpec)
}

func getAppDeploymentType(mf *Manifest) (string, error) {
	appDeploymentType := ""
	if mf.Config.Kind == "config" {
		appDeploymentType = mf.Config.ConfigDetail.Deployment
	} else {
		switch mf.Spec.ImageType {
		case "ImageTypeDocker":
			appDeploymentType = "kubernetes"
		case "ImageTypeQCOW":
			appDeploymentType = "kvm"
		default:
			return "", fmt.Errorf("unknown spec.image type '%s'", mf.Spec.ImageType)
		}
	}
	if !isValidDeploymentType(appDeploymentType) {
		return "", fmt.Errorf("invalid deployment type, '%s'", appDeploymentType)
	}
	return appDeploymentType, nil
}

func getKubeManifest(mf *Manifest, appDeploymentType string) (string, error) {
	var err error
	kubeManifest := ""
	if appDeploymentType == "kubernetes" {
		if mf.Config.ConfigDetail.Manifest == "" {
			// Generate deployment file
			log.DebugLog(log.DebugLevelMexos, "genenrating kubernetes manifest from template")
			kubeManifest, err = genKubeManifest(mf)
			if err != nil {
				log.DebugLog(log.DebugLevelMexos, "error generating kubernetes manifest from template", "error", err)
				return "", err
			}
			log.DebugLog(log.DebugLevelMexos, "generated kubernetes manifest from builtin template")
		} else {
			kubeManifest, err = retrieveKubeManifest(mf.Config.ConfigDetail.Manifest)
			if err != nil {
				log.DebugLog(log.DebugLevelMexos, "error retrieving kubernetes manifest", "source", mf.Config.ConfigDetail.Manifest, "error", err)
				return "", err
			}
			log.DebugLog(log.DebugLevelMexos, "retrieved kubernetes manifest", "source", mf.Config.ConfigDetail.Manifest)
		}
	}
	if kubeManifest == "" {
		log.DebugLog(log.DebugLevelMexos, "error, cannot retrieve, or generate; empty kube manifest")
		return "", fmt.Errorf("empty kubemanifest")
	}
	return kubeManifest, nil
}

//MEXAppCreateAppManifest creates app instances on the cluster platform
func MEXAppCreateAppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "create app from manifest", "mf", mf)
	appDeploymentType, err := getAppDeploymentType(mf)
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "app deployment", "imageType", mf.Spec.ImageType, "deploymentType", appDeploymentType, "config", mf.Config)
	kubeManifest, err := getKubeManifest(mf, appDeploymentType)
	if err != nil {
		return err
	}
	switch mf.Metadata.Operator {
	case "gcp":
		fallthrough
	case "azure":
		if appDeploymentType == "kubernetes" {
			return runKubectlCreateApp(mf, kubeManifest)
		} else if appDeploymentType == "kvm" {
			return fmt.Errorf("not yet supported")
		}
		return fmt.Errorf("unknown deployment type %s", appDeploymentType)
	default:
		if appDeploymentType == "kubernetes" {
			return CreateKubernetesAppManifest(mf, kubeManifest)
		} else if appDeploymentType == "kvm" {
			return CreateQCOW2AppManifest(mf)
		}
		return fmt.Errorf("unknown deployment type %s", appDeploymentType)
	}
}

func getKconf(mf *Manifest, createIfMissing bool) (string, error) {
	name := mexEnv["MEX_DIR"] + "/" + mf.Spec.Key + ".kubeconfig"
	log.DebugLog(log.DebugLevelMexos, "get kubeconfig name", "mf", mf, "name", name)

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
				if err = copy(defaultKubeconfig(), name); err != nil {
					return "", fmt.Errorf("can't copy %s, %v", defaultKubeconfig(), err)
				}
			case "azure":
				if err = azure.GetAKSCredentials(mf.Metadata.ResourceGroup, mf.Metadata.Name); err != nil {
					return "", fmt.Errorf("unable to get AKS credentials %v", err)
				}
				if err = copy(defaultKubeconfig(), name); err != nil {
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

	fqdnBase := uri2fqdn(mf.Spec.URI)
	for _, sn := range serviceNames {
		externalIP, err := getSvcExternalIP(sn, kconf)
		if err != nil {
			return err
		}
		fqdn := sn + "." + fqdnBase
		//TODO: if there is a DNS record left over from previous runs, clean it up before adding new record
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
	svcName := name + "-service"
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
		log.DebugLog(log.DebugLevelMexos, "getting exteralIP, examine list of services", "name", name, "svcs", svcs)
		for _, item := range svcs.Items {
			if item.Metadata.Name != svcName { // FIXME name may be appname-service or appname-tcp-service or appname-udp-service
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
	log.DebugLog(log.DebugLevelMexos, "delete app with manifest", "mf", mf)
	appDeploymentType, err := getAppDeploymentType(mf)
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "app delete", "imageType", mf.Spec.ImageType, "deploymentType", appDeploymentType, "config", mf.Config)
	kubeManifest, err := getKubeManifest(mf, appDeploymentType)
	if err != nil {
		return err
	}
	switch mf.Metadata.Operator {
	case "gcp":
		fallthrough
	case "azure":
		if appDeploymentType == "kubernetes" {
			return runKubectlDeleteApp(mf, kubeManifest)
		} else if appDeploymentType == "kvm" {
			return fmt.Errorf("not yet supported")
		}
		return fmt.Errorf("unknown image type %s", appDeploymentType)
	default:
		if appDeploymentType == "kubernetes" {
			return DeleteKubernetesAppManifest(mf, kubeManifest)
		} else if appDeploymentType == "kvm" {
			return DeleteQCOW2AppManifest(mf)
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

	out, err := sh.Command("kubectl", "delete", "-f", kfile, "--kubeconfig="+kconf).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error deleting app, %s, %v, %v", out, mf, err)
	}
	if mf.Metadata.DNSZone == "" {
		return fmt.Errorf("missing dns zone, metadata %v", mf.Metadata)
	}
	fqdn := uri2fqdn(mf.Spec.URI)
	dr, err := cloudflare.GetDNSRecords(mf.Metadata.DNSZone)
	if err != nil {
		return fmt.Errorf("cannot get dns records for %s, %s, %v", mf.Metadata.DNSZone, fqdn, err)
	}
	for _, d := range dr {
		if d.Type == "A" && d.Name == fqdn {
			if err := cloudflare.DeleteDNSRecord(mf.Metadata.DNSZone, d.ID); err != nil {
				return fmt.Errorf("cannot delete DNS record, %v", d)
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
