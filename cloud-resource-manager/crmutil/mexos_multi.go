package crmutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/azure"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/gcloud"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

const (
	mexOSKubernetes = "mex-openstack-kubernetes"
	gcloudGKE       = "gcloud-gke"
	azureAKS        = "azure-aks"
)

var yamlMEXCluster = `apiVersion: v1
kind: mex-openstack-kubernetes
resource: kubernetes-cluster
metadata:
  kind: {{.Kind}}
  name: {{.Name}}
  tags: {{.Tags}}
  tenant: {{.Tenant}}
  region: {{.Region}}
  zone: {{.Zone}}
  location: {{.Location}}
spec:
  flags: {{.Flags}}
  flavor: x1.medium
  key: {{.Key}}
  dockerregistry: registry.mobiledgex.net:5000
  rootlb: {{.RootLB}}
  networkscheme: {{.NetworkScheme}}
`

type templateFill struct {
	Name, Kind, Flavor, Tags, Tenant, Region, Zone, DNSZone string
	ImageFlavor, Location, RootLB                           string
	StorageSpec, NetworkScheme, MasterFlavor, Topology      string
	NodeFlavor, Operator, Key, Image, Options               string
	ImageType, AppURI, ProxyPath, PortMap, PathMap          string
	AccessLayer, ExternalNetwork                            string
	ExternalRouter, Flags, KubeManifest                     string
	NumMasters, NumNodes                                    int
}

//MEXClusterCreateClustInst calls MEXClusterCreate with a manifest created from the template
func MEXClusterCreateClustInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst) error {
	//XXX trigger off clusterInst or flavor to pick the right template: mex, aks, gke
	mf, err := getManifestClustInst(rootLB, clusterInst)
	if err != nil {
		return err
	}
	return MEXClusterCreateManifest(mf)
}

func getManifestClustInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst) (*Manifest, error) {
	name := clusterInst.Key.ClusterKey.Name
	flavor := clusterInst.Flavor.Name
	data := templateFill{
		Name:          name,
		Tags:          name + "-tag",
		Tenant:        name + "-tenant",
		Key:           clusterInst.Key.ClusterKey.Name,
		Kind:          "mex-k8s-cluster",
		Region:        "eu-central-1",
		Zone:          "eu-central-1c",
		Location:      "bonn",
		Flavor:        flavor,
		RootLB:        rootLB.Name,
		NetworkScheme: "priv-subnet,mex-k8s-net-1,10.101.X.0/24",
	}
	mf, err := templateUnmarshal(&data, yamlMEXCluster)
	if err != nil {
		return nil, err
	}
	return mf, nil
}

//MEXClusterCreateManifest creates a cluster
func MEXClusterCreateManifest(mf *Manifest) error {
	Debug("creating cluster", "mf", mf)
	switch mf.Kind {
	case mexOSKubernetes:
		//guid, err := mexCreateClusterKubernetes(mf)
		err := mexCreateClusterKubernetes(mf)
		if err != nil {
			return fmt.Errorf("can't create cluster, %v", err)
		}
		//Debug("new guid", "guid", *guid)
		Debug("created kubernetes cluster", "mf", mf)
		return nil
	case gcloudGKE:
		return gcloudCreateGKE(mf)
	case azureAKS:
		return azureCreateAKS(mf)
	default:
		return fmt.Errorf("unknown type %s", mf.Kind)
	}
}

func azureCreateAKS(mf *Manifest) error {
	var err error
	if err = azure.CreateResourceGroup(mf.Metadata.ResourceGroup, mf.Metadata.Location); err != nil {
		return err
	}
	if err = azure.CreateAKSCluster(mf.Metadata.ResourceGroup, mf.Metadata.Location); err != nil {
		return err
	}
	if err = azure.GetAKSCredentials(mf.Metadata.ResourceGroup, mf.Metadata.Location); err != nil {
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
	if err = gcloud.GetGKECredentials(mf.Metadata.Name); err != nil {
		return err
	}
	return nil
}

//MEXClusterRemoveClustInst calls MEXClusterRemove with a manifest created from the template
func MEXClusterRemoveClustInst(rootLb *MEXRootLB, clusterInst *edgeproto.ClusterInst) error {
	mf, err := getManifestClustInst(rootLb, clusterInst)
	if err != nil {
		return err
	}
	return MEXClusterRemoveManifest(mf)
}

var yamlMEXFlavor = `apiVersion: v1
kind: mex-openstack-kubernetes
resource: kubernetes-cluster
metadata:
  name: {{.Name}}
  tags: {{.Tags}}
  kind: {{.Kind}}
spec:
  flags: {{.Flags}}
  flavor: {{.Name}}
  flavors: 
    - name: {{.Name}}
      nodes: {{.NumNodes}}
      masters: {{.NumMasters}}
      networkscheme: {{.NetworkScheme}}
      masterflavor: {{.MasterFlavor}}
      nodeflavor: {{.NodeFlavor}}
      storagescheme: {{.StorageSpec}}
      topology: {{.Topology}}
`

//MEXAddFlavorClusterInst uses template to fill in details for flavor add request and calls MEXAddFlavor
func MEXAddFlavorClusterInst(flavor *edgeproto.Flavor) error {
	name := flavor.Key.Name
	data := templateFill{
		Name:          name,
		Tags:          name + "-tag",
		RootLB:        "mexlb.tdg.mobiledgex.net",
		Kind:          "k8s-cluster-flavor",
		NumMasters:    1,
		NumNodes:      2,
		NetworkScheme: "priv-subnet,mex-k8s-net-1,10.201.X0/24",
		StorageSpec:   "default",
		NodeFlavor:    "m4.large",
		MasterFlavor:  "m4.large",
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
	Debug("add flavor", "mf", mf)
	//TODO use full manifest and validate against platform data
	return AddFlavor(mf.Spec.Flavor)
}

// TODO DeleteFlavor -- but almost never done

// TODO lookup guid using name

//MEXClusterRemoveManifest removes a cluster
func MEXClusterRemoveManifest(mf *Manifest) error {
	Debug("removing cluster", "mf", mf)
	switch mf.Kind {
	case mexOSKubernetes:
		if err := mexDeleteClusterKubernetes(mf); err != nil {
			return fmt.Errorf("can't remove cluster, %v", err)
		}
		return nil
	case gcloudGKE:
		return gcloud.DeleteGKECluster(mf.Metadata.Name)
	case azureAKS:
		return azure.DeleteAKSCluster(mf.Metadata.ResourceGroup)
	default:
		return fmt.Errorf("unknown type %s", mf.Kind)
	}
}

var yamlMEXPlatform = `apiVersion: v1
kind: mex-openstack-kubernetes
resource: openstack-platform
metadata:
  kind: {{.Kind}}
  name: {{.Name}}
  tags: {{.Tags}}
  tenant: {{.Tenant}}
  operator: {{.Operator}}
  region: {{.Region}}
  zone: {{.Zone}}
  location: {{.Location}}
  openrc: ~/.mobiledgex/openrc
  dnszone: {{.DNSZone}}
spec:
  flags: {{.Flags}}
  key: {{.Key}}
  flavor: {{.Flavor}}
  rootlb: {{.RootLB}}
  externalnetwork: {{.ExternalNetwork}}
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
	mf, err := fillTemplateCloudletKey(rootLB, cloudletKeyStr)
	if err != nil {
		return err
	}
	return MEXPlatformInitManifest(mf)
}

func fillTemplateCloudletKey(rootLB *MEXRootLB, cloudletKeyStr string) (*Manifest, error) {
	Debug("fill template cloudletkeystr", "cloudletkeystr", cloudletKeyStr)
	clk := edgeproto.CloudletKey{}
	err := json.Unmarshal([]byte(cloudletKeyStr), &clk)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal json cloudletkey %s, %v", cloudletKeyStr, err)
	}
	Debug("unmarshalled cloudletkeystr", "cloudletkey", clk)
	name := clk.Name
	operator := clk.OperatorKey.Name
	data := templateFill{
		Name:            name,
		Tags:            name + "-tag",
		Key:             clk.OperatorKey.Name + "," + clk.Name,
		Flavor:          "x1.medium", // XXX
		Operator:        operator,
		Location:        "bonn",
		Region:          "eu-central-1",
		Zone:            "eu-central-1c",
		RootLB:          rootLB.Name,
		Image:           "registry.mobiledgex.net:5000/mobiledgex/mexosagent",
		Kind:            "mex-tdg-openstack-kubernetes",
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
	mf, err := fillTemplateCloudletKey(rootLB, cloudletKeyStr)
	if err != nil {
		return err
	}
	return MEXPlatformCleanManifest(mf)
}

func templateUnmarshal(data *templateFill, yamltext string) (*Manifest, error) {
	//Debug("template unmarshal", "yamltext", string, "data", data)
	tmpl, err := template.New("mex").Parse(yamltext)
	if err != nil {
		return nil, fmt.Errorf("can't create template for, %v", err)
	}
	var outbuffer bytes.Buffer
	err = tmpl.Execute(&outbuffer, data)
	if err != nil {
		//Debug("template data", "data", data)
		return nil, fmt.Errorf("can't execute template, %v", err)
	}
	mf := &Manifest{}
	err = yaml.Unmarshal(outbuffer.Bytes(), mf)
	if err != nil {
		Debug("error yaml unmarshal, templated data", "templated buffer data", outbuffer.String())
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
	Debug("init platform", "mf", mf)
	err := checkEnvironment()
	if err != nil {
		return err
	}
	switch mf.Kind {
	case mexOSKubernetes:
		if err = MEXCheckEnvVars(mf); err != nil {
			return err
		}
		//TODO validate all mf content against platform data
		if err = RunMEXAgentManifest(mf); err != nil {
			return err
		}
	case gcloudGKE:
		return nil //nothing to do
	case azureAKS:
		return nil //nothing to do
	default:
		return fmt.Errorf("unknown type %s", mf.Kind)
	}
	return nil
}

//MEXPlatformCleanManifest cleans up the platform
func MEXPlatformCleanManifest(mf *Manifest) error {
	Debug("clean platform", "mf", mf)
	err := checkEnvironment()
	if err != nil {
		return err
	}
	switch mf.Kind {
	case mexOSKubernetes:
		if err = MEXCheckEnvVars(mf); err != nil {
			return err
		}
		if err = RemoveMEXAgentManifest(mf); err != nil {
			return err
		}
	case gcloudGKE:
		return nil //nothing to do
	case azureAKS:
		return nil
	default:
		return fmt.Errorf("unknown type %s", mf.Kind)
	}
	return nil
}

var yamlMEXAppKubernetes = `apiVersion: v1
kind: mex-openstack-kubernetes
resource: kubernetes-cluster
metadata:
  kind: {{.Kind}}
  name: {{.Name}}
  tags: {{.Tags}}
  tenant: {{.Tenant}}
spec:
  flags: {{.Flags}}
  key: {{.Key}}
  rootlb: {{.RootLB}}
  imagetype: {{.ImageType}}
  image: {{.Image}}
  proxypath: {{.ProxyPath}}
  portmap: {{.PortMap}} 
  pathmap: {{.PathMap}}
  uri: {{.AppURI}}
  kubemanifest: {{.KubeManifest}}
  accesslayer: {{.AccessLayer}}
`
var yamlMEXAppQcow2 = `apiVersion: v1
kind: mex-openstack
resource: openstack
metadata:
  kind: {{.Kind}}
  name: {{.Name}}
  tags: {{.Tags}}
  tenant: {{.Tenant}}
spec:
  flags: {{.Flags}}
  key: {{.Key}}
  rootlb: {{.RootLB}}
  imagetype: {{.ImageType}}
  image: {{.Image}}
  imageflavor: {{.ImageFlavor}}
  proxypath: {{.ProxyPath}}
  portmap: {{.PortMap}} 
  pathmap: {{.PathMap}}
  uri: {{.AppURI}}
  kubemanifest: {{.KubeManifest}}
  accesslayer: {{.AccessLayer}}
  networkscheme: {{.NetworkScheme}}
`

//MEXCreateAppInst calls MEXCreateApp with templated manifest
func MEXCreateAppInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst) error {
	Debug("mex create app inst", "rootlb", rootLB, "clusterinst", clusterInst, "appinst", appInst)
	imageType, ok := edgeproto.ImageType_name[int32(appInst.ImageType)]
	if !ok {
		return fmt.Errorf("cannot find imagetype in map")
	}
	accessLayer, aok := edgeproto.AccessLayer_name[int32(appInst.AccessLayer)]
	if !aok {
		return fmt.Errorf("cannot find accesslayer in map")
	}
	var data templateFill
	var err error
	var mf *Manifest
	switch imageType {
	case "ImageTypeDocker": //XXX assume kubernetes
		data = templateFill{
			Kind:         "mex-app-kubernetes",
			Name:         appInst.Key.AppKey.Name,
			Tags:         appInst.Key.AppKey.Name + "-kubernetes-tag",
			Key:          clusterInst.Key.ClusterKey.Name,
			Tenant:       appInst.Key.AppKey.Name + "-tenant",
			RootLB:       rootLB.Name,
			Image:        appInst.ImagePath,
			ImageType:    imageType,
			ProxyPath:    appInst.Key.AppKey.Name,
			AppURI:       appInst.Uri,
			PortMap:      appInst.MappedPorts,
			PathMap:      appInst.MappedPath,
			AccessLayer:  accessLayer,
			KubeManifest: appInst.ConfigMap,
		}
		mf, err = templateUnmarshal(&data, yamlMEXAppKubernetes)
		if err != nil {
			return err
		}
	case "ImageTypeQCOW":
		data = templateFill{
			Kind:          "mex-app-vm-qcow2",
			Name:          appInst.Key.AppKey.Name,
			Tags:          appInst.Key.AppKey.Name + "-qcow-tag",
			Key:           clusterInst.Key.ClusterKey.Name,
			Tenant:        appInst.Key.AppKey.Name + "-tenant",
			RootLB:        rootLB.Name,
			Image:         appInst.ImagePath,
			ImageFlavor:   appInst.Flavor.Name,
			ImageType:     imageType,
			ProxyPath:     appInst.Key.AppKey.Name,
			AppURI:        appInst.Uri,
			PortMap:       appInst.MappedPorts,
			PathMap:       appInst.MappedPath,
			AccessLayer:   accessLayer,
			NetworkScheme: "external-ip,external-network-shared",
		}
		mf, err = templateUnmarshal(&data, yamlMEXAppQcow2)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown image type %s", imageType)
	}
	return MEXCreateAppManifest(mf)
}

//MEXCreateAppManifest creates app instances on the cluster platform
func MEXCreateAppManifest(mf *Manifest) error {
	Debug("create app from manifest", "mf", mf)
	switch mf.Kind {
	case mexOSKubernetes:
		if mf.Spec.ImageType == "ImageTypeDocker" {
			return CreateKubernetesAppManifest(mf)
		} else if mf.Spec.ImageType == "ImageTypeQCOW" {
			//TODO
			return CreateQCOW2AppManifest(mf)
		} else {
			return fmt.Errorf("unknown image type %s", mf.Spec.ImageType)
		}
	case gcloudGKE:
		return fmt.Errorf("not yet supported, type %s", mf.Kind)
	case azureAKS:
		return fmt.Errorf("not yet supported, type %s", mf.Kind)
	default:
		return fmt.Errorf("unknown type %s", mf.Kind)
	}
}

//MEXKillAppManifest kills app
func MEXKillAppManifest(mf *Manifest) error {
	Debug("delete app", "mf", mf)
	switch mf.Kind {
	case mexOSKubernetes:
		if mf.Spec.ImageType == "ImageTypeDocker" {
			return DestroyKubernetesAppManifest(mf)
		} else if mf.Spec.ImageType == "ImageTypeQCOW" {
			return DestroyQCOW2AppManifest(mf)
		} else {
			return fmt.Errorf("unknown image type %s", mf.Spec.ImageType)
		}
	case gcloudGKE:
		return fmt.Errorf("not yet supported, type %s", mf.Kind)
	case azureAKS:
		return fmt.Errorf("not yet supported, type %s", mf.Kind)
	default:
		return fmt.Errorf("unknown type %s", mf.Kind)
	}
}
