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
	Name, Kind, Flavor, Tags, Tenant, Region, Zone, DNSZone string
	ImageFlavor, Location, RootLB, ResourceGroup            string
	StorageSpec, NetworkScheme, MasterFlavor, Topology      string
	NodeFlavor, Operator, Key, Image, Options               string
	ImageType, AppURI, ProxyPath, PortMap, PathMap          string
	AccessLayer, ExternalNetwork, Project                   string
	ExternalRouter, Flags, KubeManifest                     string
	NumMasters, NumNodes                                    int
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
		if err != nil {
			return fmt.Errorf("cannot get kconf, %v, %v, %v", mf, kconf, err)
		}
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
spec:
  flags: {{.Flags}}
  key: {{.Key}}
  rootlb: {{.RootLB}}
  image: {{.Image}}
  imagetype: {{.ImageType}}
  imageflavor: {{.ImageFlavor}}
  proxypath: {{.ProxyPath}}
  portmap: {{.PortMap}} 
  pathmap: {{.PathMap}}
  flavor: {{.Flavor}}
  uri: {{.AppURI}}
  kubemanifest: {{.KubeManifest}}
  accesslayer: {{.AccessLayer}}
  networkscheme: {{.NetworkScheme}}
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
spec:
  flags: {{.Flags}}
  key: {{.Key}}
  rootlb: {{.RootLB}}
  image: {{.Image}}
  imagetype: {{.ImageType}}
  imageflavor: {{.ImageFlavor}}
  proxypath: {{.ProxyPath}}
  portmap: {{.PortMap}} 
  pathmap: {{.PathMap}}
  flavor: {{.Flavor}}
  uri: {{.AppURI}}
  kubemanifest: {{.KubeManifest}}
  accesslayer: {{.AccessLayer}}
  networkscheme: {{.NetworkScheme}}
`

//MEXAppCreateAppInst creates app inst with templated manifest
func MEXAppCreateAppInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst) error {
	log.DebugLog(log.DebugLevelMexos, "mex create app inst", "rootlb", rootLB, "clusterinst", clusterInst, "appinst", appInst)
	mf, err := fillAppTemplate(rootLB, appInst, clusterInst)
	if err != nil {
		return err
	}
	return MEXAppCreateAppManifest(mf)
}

//MEXAppDeleteAppInst deletes app with templated manifest
func MEXAppDeleteAppInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst) error {
	log.DebugLog(log.DebugLevelMexos, "mex delete app inst", "rootlb", rootLB, "clusterinst", clusterInst, "appinst", appInst)
	mf, err := fillAppTemplate(rootLB, appInst, clusterInst)
	if err != nil {
		return err
	}
	return MEXAppDeleteAppManifest(mf)
}

func fillAppTemplate(rootLB *MEXRootLB, appInst *edgeproto.AppInst, clusterInst *edgeproto.ClusterInst) (*Manifest, error) {
	var data templateFill
	var err error
	var mf *Manifest
	imageType, ok := edgeproto.ImageType_name[int32(appInst.ImageType)]
	if !ok {
		return nil, fmt.Errorf("cannot find imagetype in map")
	}
	accessLayer, aok := edgeproto.AccessLayer_name[int32(appInst.AccessLayer)]
	if !aok {
		return nil, fmt.Errorf("cannot find accesslayer in map")
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
	if len(appInst.Key.AppKey.Name) < 3 {
		log.DebugLog(log.DebugLevelMexos, "warning, very short appkey name", "name", appInst.Key.AppKey.Name)
	}
	switch imageType {
	case "ImageTypeDocker": //XXX assume kubernetes
		data = templateFill{
			Kind:        clusterInst.Flavor.Name,
			Name:        util.K8SSanitize(appInst.Key.AppKey.Name),
			Tags:        util.K8SSanitize(appInst.Key.AppKey.Name) + "-kubernetes-tag",
			Key:         clusterInst.Key.ClusterKey.Name,
			Tenant:      util.K8SSanitize(appInst.Key.AppKey.Name) + "-tenant",
			Operator:    util.K8SSanitize(clusterInst.Key.CloudletKey.OperatorKey.Name),
			RootLB:      rootLB.Name,
			Image:       appInst.ImagePath,
			ImageType:   imageType,
			ImageFlavor: appInst.Flavor.Name,
			ProxyPath:   util.K8SSanitize(appInst.Key.AppKey.Name),
			AppURI:      appInst.Uri,
			//PortMap:      appInst.MappedPorts,
			PathMap:      appInst.MappedPath,
			AccessLayer:  accessLayer,
			KubeManifest: appInst.Config, // 'Config' not the same as Kubernetes ConfigMap. Here, used by controller to send kubemanifest
		}
		mf, err = templateUnmarshal(&data, yamlMEXAppKubernetes)
		if err != nil {
			return nil, err
		}
	case "ImageTypeQCOW":
		data = templateFill{
			Kind:        clusterInst.Flavor.Name,
			Name:        appInst.Key.AppKey.Name,
			Tags:        appInst.Key.AppKey.Name + "-qcow-tag",
			Key:         clusterInst.Key.ClusterKey.Name,
			Tenant:      appInst.Key.AppKey.Name + "-tenant",
			Operator:    clusterInst.Key.CloudletKey.OperatorKey.Name,
			RootLB:      rootLB.Name,
			Image:       appInst.ImagePath,
			ImageFlavor: appInst.Flavor.Name,
			ImageType:   imageType,
			ProxyPath:   appInst.Key.AppKey.Name,
			AppURI:      appInst.Uri,
			//PortMap:       appInst.MappedPorts,
			PathMap:       appInst.MappedPath,
			AccessLayer:   accessLayer,
			NetworkScheme: "external-ip,external-network-shared",
		}
		mf, err = templateUnmarshal(&data, yamlMEXAppQcow2)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown image type %s", imageType)
	}
	return mf, nil
}

//MEXAppCreateAppManifest creates app instances on the cluster platform
func MEXAppCreateAppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "create app from manifest", "mf", mf)
	switch mf.Metadata.Operator {
	case "gcp":
		if mf.Spec.ImageType == "ImageTypeDocker" {
			return runKubectlCreateApp(mf)
		} else if mf.Spec.ImageType == "ImageTypeQCOW" { // XXX gcp requires raw
			return fmt.Errorf("not yet supported")
		} else {
			return fmt.Errorf("unknown image type %s", mf.Spec.ImageType)
		}
	case "azure":
		if mf.Spec.ImageType == "ImageTypeDocker" {
			return runKubectlCreateApp(mf)
		} else if mf.Spec.ImageType == "ImageTypeQCOW" { // XXX azure requires vhd
			return fmt.Errorf("not yet supported")
		} else {
			return fmt.Errorf("unknown image type %s", mf.Spec.ImageType)
		}
	default:
		if mf.Spec.ImageType == "ImageTypeDocker" {
			return CreateKubernetesAppManifest(mf)
		} else if mf.Spec.ImageType == "ImageTypeQCOW" {
			return CreateQCOW2AppManifest(mf)
		} else {
			return fmt.Errorf("unknown image type %s", mf.Spec.ImageType)
		}
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

func runKubectlCreateApp(mf *Manifest) error {
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

	out, err = sh.Command("kubectl", "create", "-f", mf.Spec.KubeManifest, "--kubeconfig="+kconf).Output()
	if err != nil {
		return fmt.Errorf("error creating app, %s, %v, %v", out, mf, err)
	}
	return nil
}

//MEXAppDeleteManifest kills app
func MEXAppDeleteAppManifest(mf *Manifest) error {
	log.DebugLog(log.DebugLevelMexos, "delete app", "mf", mf)
	switch mf.Metadata.Operator {
	case "gcp":
		if mf.Spec.ImageType == "ImageTypeDocker" {
			return runKubectlDeleteApp(mf)
		} else if mf.Spec.ImageType == "ImageTypeQCOW" {
			return fmt.Errorf("not yet supported")
		} else {
			return fmt.Errorf("unknown image type %s", mf.Spec.ImageType)
		}
	case "azure":
		if mf.Spec.ImageType == "ImageTypeDocker" {
			return runKubectlDeleteApp(mf)
		} else if mf.Spec.ImageType == "ImageTypeQCOW" {
			return fmt.Errorf("not yet supported")
		} else {
			return fmt.Errorf("unknown image type %s", mf.Spec.ImageType)
		}
	default:
		if mf.Spec.ImageType == "ImageTypeDocker" {
			return DestroyKubernetesAppManifest(mf)
		} else if mf.Spec.ImageType == "ImageTypeQCOW" {
			return DestroyQCOW2AppManifest(mf)
		}
		return fmt.Errorf("unknown image type %s", mf.Spec.ImageType)
	}
}

func runKubectlDeleteApp(mf *Manifest) error {
	kconf, err := getKconf(mf, false)
	if err != nil {
		return fmt.Errorf("error deleting app due to kconf,  %v, %v", mf, err)
	}
	out, err := sh.Command("kubectl", "delete", "-f", mf.Spec.KubeManifest, "--kubeconfig="+kconf).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error deleting app, %s, %v, %v", out, mf, err)
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
