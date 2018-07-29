package crmutil

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/mobiledgex/edge-cloud/edgeproto"

	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/azure"
	"github.com/mobiledgex/edge-cloud-infra/k8s-prov/gcloud"
	log "gitlab.com/bobbae/logrus"
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
  flavor: {{.Flavor}}
  rootlb: {{.RootLB}}
  networks:
      - name: {{.NetworkName}}
        kind: {{.NetworkKind}}
        cidr: {{.CIDR}}
`

type templateFill struct {
	Name, Kind, Flavor, Tags, Tenant, Region, Zone, DNSZone string
	Location, RootLB, NetworkName, NetworkKind, CIDR        string
	StorageSpec, NetworkSpec, MasterFlavor, Topology        string
	NodeFlavor, Operator, Key, Image, Registry, Options     string
	AppFlavor, ProxyPath, PortMap, PathMap, Kubernetes      string
	AccessLayer, ExternalNetwork, InternalNetwork           string
	InternalCIDR, ExternalRouter, Flags                     string
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
		Name:        name,
		Tags:        name + "-tag",
		Tenant:      name + "-tenant",
		Kind:        "mex-k8s-cluster",
		Region:      "eu-central-1",
		Zone:        "eu-central-1c",
		Location:    "bonn",
		Flavor:      flavor,
		RootLB:      rootLB.Name,
		NetworkName: "mex-k8s-net-1",
		NetworkKind: "priv-subnet",
		CIDR:        "10.101.X.0/24",
	}
	mf, err := templateUnmarshal(&data, yamlMEXCluster)
	if err != nil {
		return nil, err
	}
	return mf, nil
}

//MEXClusterCreateManifest creates a cluster
func MEXClusterCreateManifest(mf *Manifest) error {
	log.Debug("creating cluster", mf)
	switch mf.Kind {
	case mexOSKubernetes:
		guid, err := mexCreateClusterKubernetes(mf)
		if err != nil {
			return fmt.Errorf("can't create cluster, %v", err)
		}
		log.Debugln("new guid", *guid)
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
  nummasters: {{.NumMaster}}
  flavors: 
    - name: {{.Name}}
      nodes: {{.NumNodes}}
      masters: {{.NumMasters}}
      networkscheme: {{.NetworkSpec}}
      masterflavor: {{.MasterFlavor}}
      nodeflavor: {{.NodeFlavor}}
      storagescheme: {{.StorageSpec}}
      topology: {{.Topology}}
`

//MEXAddFlavorClusterInst uses template to fill in details for flavor add request and calls MEXAddFlavor
func MEXAddFlavorClusterInst(flavor *edgeproto.Flavor) error {
	name := flavor.Key.Name
	data := templateFill{
		Name:         name,
		Tags:         name + "-tag",
		RootLB:       "mex-lb-1.mobiledgex.net",
		Kind:         "k8s-cluster-flavor",
		NumMasters:   1,
		NumNodes:     2,
		NetworkSpec:  "priv-subnet,mex-k8s-net-1,10.201.X0/24",
		StorageSpec:  "default",
		NodeFlavor:   "m4.large",
		MasterFlavor: "m4.large",
		Topology:     "type-1",
	}
	mf, err := templateUnmarshal(&data, yamlMEXFlavor)
	if err != nil {
		return err
	}
	return MEXAddFlavor(mf)
}

//MEXAddFlavor adds flavor using manifest
func MEXAddFlavor(mf *Manifest) error {
	log.Debugln("add flavor", mf)
	//TODO use full manifest and validate against platform data
	return AddFlavor(mf.Spec.Flavor)
}

// TODO DeleteFlavor -- but almost never done

// TODO lookup guid using name

//MEXClusterRemoveManifest removes a cluster
func MEXClusterRemoveManifest(mf *Manifest) error {
	log.Debugln("removing cluster", mf)
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
  rootlb: {{.RootLB}}
  dockerregistry: {{.Registry}}
  externalnetwork: {{.ExternalNetwork}}
  internalnetwork: {{.InternalNetwork}}
  internalcidr: {{.InternalCIDR}}
  externalrouter: {{.ExternalRouter}}
  options: {{.Options}}
  agent: 
    image: {{.Image}}
    status: active
  images:
    - name: mobiledgex-16.04-2
      kind: qcow2
      favorite: yes
      osflavor: m4.large
  flavors:
    - name: x1.medium
      cpus: 4
      memory: 8G
      storage: local
      masters: 1
      nodes: 2
      favorite: yes
      networkscheme: mexproxy
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
	clk := edgeproto.CloudletKey{}
	err := json.Unmarshal([]byte(cloudletKeyStr), &clk)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal json cloudletkey %s, %v", cloudletKeyStr, err)
	}
	name := clk.Name
	operator := clk.OperatorKey.Name
	data := templateFill{
		Name:            name,
		Tags:            name + "-tag",
		Key:             cloudletKeyStr,
		Operator:        operator,
		Location:        "bonn",
		Region:          "eu-central-1",
		Zone:            "eu-central-1c",
		RootLB:          rootLB.Name,
		Image:           "registry.mobiledgex.net:5000/mobiledgex/mexosagent",
		Kind:            "mex-tdg-openstack-kubernetes",
		Registry:        "registry.mobiledgex.net:5000",
		ExternalNetwork: "external-network-shared",
		InternalNetwork: "mex-k8s-net-1",
		InternalCIDR:    "10.101.X.X/24",
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
	tmpl, err := template.New("mexflavor").Parse(yamltext)
	if err != nil {
		return nil, fmt.Errorf("can't create template for, %v", err)
	}
	var outbuffer bytes.Buffer
	bufwriter := bufio.NewWriter(&outbuffer)
	err = tmpl.Execute(bufwriter, data)
	if err != nil {
		log.Debugln("template data", data)
		return nil, fmt.Errorf("can't execute template, %v", err)
	}
	mf := &Manifest{}
	err = yaml.Unmarshal(outbuffer.Bytes(), mf)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal templated data, %v", err)
	}
	return mf, nil
}

//MEXPlatformInitManifest initializes platform
func MEXPlatformInitManifest(mf *Manifest) error {
	log.Debug("init platform", mf)
	var err error
	switch mf.Kind {
	case mexOSKubernetes:
		if err = MEXCheckEnvVars(mf); err != nil {
			return err
		}
		//TODO validate all mf content against platform data
		if err = RunMEXAgentManifest(mf, false); err != nil {
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
	log.Debugln("clean platform", mf)
	var err error
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

var yamlMEXApp = `apiVersion: v1
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
  dockerregistry: {{.Registry}}
  image: {{.Image}}
  proxypath: {{.ProxyPath}}
  portmap: {{.PortMap}} 
  pathmap: {{.PathMap}}
  flavor: {{.AppFlavor}}
  uri: {{.AppUri}}
  kubernetes: {{.Kubernetes}}
  accesslayer: {{.AccessLayer}}
`

//MEXCreateAppInst calls MEXCreateApp with templated manifest
func MEXCreateAppInst(rootLB *MEXRootLB, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst) error {
	imageType, ok := edgeproto.ImageType_name[int32(appInst.ImageType)]
	if !ok {
		return fmt.Errorf("cannot find imagetype in map")
	}
	accessLayer, aok := edgeproto.AccessLayer_name[int32(appInst.AccessLayer)]
	if !aok {
		return fmt.Errorf("cannot find accesslayer in map")
	}
	var data templateFill
	switch imageType {
	case "ImageTypeDocker": //XXX does not distinguish plain docker vs kubernetes deployments
		data = templateFill{
			Kind:        "mex-app-docker",
			Name:        appInst.Key.AppKey.Name,
			Tags:        appInst.Key.AppKey.Name + "-docker-tag",
			Key:         clusterInst.Key.ClusterKey.Name,
			Tenant:      appInst.Key.AppKey.Name + "-tenant",
			RootLB:      rootLB.Name,
			Registry:    "registry.mobiledgex.net:5000",
			Image:       appInst.ImagePath,                           //XXX docker image name?
			ProxyPath:   rootLB.Name + "/" + appInst.Key.AppKey.Name, //appInst.Uri ?
			PortMap:     "80:80",                                     //appInst.MappedPorts,         //XXX format
			PathMap:     appInst.MappedPath,                          //XXX format
			AppFlavor:   appInst.Flavor.Name,                         //XXX not sure what this is
			AccessLayer: accessLayer,
			Kubernetes:  "", // XXX "https://k8s.io/examples/application/deployment.yaml",
		}
	case "ImageTypeQCOW":
		data = templateFill{
			Kind:        "mex-app-virtual-machine",
			Name:        appInst.Key.AppKey.Name,
			Tags:        appInst.Key.AppKey.Name + "-qcow-tag",
			Key:         clusterInst.Key.ClusterKey.Name,
			Tenant:      appInst.Key.AppKey.Name + "-tenant",
			RootLB:      rootLB.Name,
			Registry:    "registry.mobiledgex.net:5000",
			Image:       appInst.ImagePath,                           //XXX qcow image name?
			ProxyPath:   rootLB.Name + "/" + appInst.Key.AppKey.Name, //appInst.Uri ?
			PortMap:     appInst.MappedPorts,                         //XXX format
			PathMap:     appInst.MappedPath,                          //XXX format
			AppFlavor:   appInst.Flavor.Name,                         //XXX not sure what this is
			AccessLayer: accessLayer,
		}
	default:
		return fmt.Errorf("unknown image type")
	}
	//TODO Use non kubernetes data as CRD or ConfigMap
	mf, err := templateUnmarshal(&data, yamlMEXPlatform)
	if err != nil {
		return err
	}
	return MEXCreateAppManifest(mf)
}

//MEXCreateApp creates app instances on the cluster platform
func MEXCreateAppManifest(mf *Manifest) error {
	log.Debugln("create app", mf)
	switch mf.Kind {
	case mexOSKubernetes:
		if strings.Contains(mf.Metadata.Tags, "docker") {
			if mf.Spec.Kubernetes == "" { //plain docker
				return CreateDockerAppManifest(mf)
			}
			return CreateKubernetesAppManifest(mf)
		} else if strings.Contains(mf.Metadata.Tags, "qcow") {
		} else {
			return fmt.Errorf("insufficient tag info")
		}
	case gcloudGKE:
		return fmt.Errorf("not yet supported, type %s", mf.Kind)
	case azureAKS:
		return fmt.Errorf("not yet supported, type %s", mf.Kind)
	default:
		return fmt.Errorf("unknown type %s", mf.Kind)
	}
	return nil
}

//MEXKillAppManifest kills app
func MEXKillAppManifest(mf *Manifest) error {
	log.Debugln("delete app", mf)
	switch mf.Kind {
	case mexOSKubernetes:
		return DestroyDockerAppManifest(mf)
	case gcloudGKE:
		return fmt.Errorf("not yet supported, type %s", mf.Kind)
	case azureAKS:
		return fmt.Errorf("not yet supported, type %s", mf.Kind)
	default:
		return fmt.Errorf("unknown type %s", mf.Kind)
	}
	return nil
}
