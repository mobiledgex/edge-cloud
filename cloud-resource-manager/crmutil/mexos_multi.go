package crmutil

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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
  flavor: {{.Flavor}}
  rootlb: {{.RootLB}}
  networks:
      - name: {{.NetworkName}}
        kind: {{.NetworkKind}}
        cidr: {{.CIDR}}
`

type templateFill struct {
	Name, Kind, Flavor, Tags, Tenant, Region, Zone     string
	Location, RootLB, NetworkName, NetworkKind, CIDR   string
	StorageSpec, NetworkSpec, MasterFlavor, Topology   string
	NodeFlavor, Operator, Key, Image, Registry         string
	AppFlavor, ProxyPath, PortMap, PathMap, Kubernetes string
	AccessLayer                                        string
	NumMasters, NumNodes                               int
}

//MEXClusterCreateClustInst calls MEXClusterCreate with a manifest created from the template
func MEXClusterCreateClustInst(clusterInst *edgeproto.ClusterInst) error {
	//XXX future: trigger off clusterInst or flavor to pick the right template: mex, aks, gke
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
		RootLB:      "mex-lb-1.mobiledgex.net",
		NetworkName: "mex-k8s-net-1",
		NetworkKind: "priv-subnet",
		CIDR:        "10.101.X.0/24",
	}

	mf, err := templateUnmarshal(&data, yamlMEXCluster)
	if err != nil {
		return err
	}
	return MEXClusterCreate(mf)
}

//MEXGUIDMap is a map of name=>guid for initialized clusters
var MEXGUIDMap = make(map[string]string)

//MEXClusterCreate creates a cluster
func MEXClusterCreate(mf *Manifest) error {
	log.Debug("creating cluster", mf)

	var err error
	var guid *string

	switch mf.Kind {
	case mexOSKubernetes:
		log.Debugln(MEXGUIDMap)
		guidval, ok := MEXGUIDMap[mf.Metadata.Name] //TODO save map on nonvolatile storage
		if ok {
			return fmt.Errorf("%s exists as %s", mf.Metadata.Name, guidval)
		}
		guid, err = CreateCluster(mf.Spec.RootLB, mf.Spec.Flavor, mf.Metadata.Name,
			mf.Spec.Networks[0].Kind+","+mf.Spec.Networks[0].Name+","+mf.Spec.Networks[0].CIDR,
			mf.Metadata.Tags, mf.Metadata.Tenant)
		if err != nil {
			return fmt.Errorf("can't create cluster, %v", err)
		}
		log.Debugln("guid", *guid)
		_, ok = MEXGUIDMap[mf.Metadata.Name]
		if ok {
			log.Warningf("guid %s exists for %s", *guid, mf.Metadata.Name)
		}
		MEXGUIDMap[mf.Metadata.Name] = *guid
		log.Debugln(MEXGUIDMap)
	case gcloudGKE:
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
	case azureAKS:
		if err = azure.CreateResourceGroup(mf.Metadata.ResourceGroup, mf.Metadata.Location); err != nil {
			return err
		}
		if err = azure.CreateAKSCluster(mf.Metadata.ResourceGroup, mf.Metadata.Location); err != nil {
			return err
		}
		if err = azure.GetAKSCredentials(mf.Metadata.ResourceGroup, mf.Metadata.Location); err != nil {
			return err
		}
	}
	log.Debugln("created cluster", mf)
	return err
}

//MEXClusterRemoveClustInst calls MEXClusterCreate with a manifest created from the template
func MEXClusterRemoveClustInst(clusterInst *edgeproto.ClusterInst) error {
	name := clusterInst.Key.ClusterKey.Name
	data := templateFill{
		Name:        name,
		Tags:        name + "-tag",
		Tenant:      name + "-tenant",
		Region:      "eu-central-1",
		Zone:        "eu-central-1c",
		Location:    "bonn",
		Flavor:      "x1.medium",
		RootLB:      "mex-lb-1.mobiledgex.net",
		NetworkName: "mex-k8s-net-1",
		NetworkKind: "priv-subnet",
		CIDR:        "10.101.X.0/24",
	}

	mf, err := templateUnmarshal(&data, yamlMEXCluster)
	if err != nil {
		return err
	}
	return MEXClusterRemove(mf)
}

var yamlMEXFlavor = `apiVersion: v1
kind: mex-openstack-kubernetes
resource: kubernetes-cluster
metadata:
  name: {{.Name}}
  tags: {{.Tags}}
  kind: {{.Kind}}
spec:
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

//MEXClusterRemove removes a cluster
func MEXClusterRemove(mf *Manifest) error {
	log.Debugln("removing cluster", mf)

	var err error

	switch mf.Kind {
	case mexOSKubernetes:
		log.Debugln(MEXGUIDMap)
		_, ok := MEXGUIDMap[mf.Metadata.Name]
		if !ok {
			return fmt.Errorf("cluster %s does not exist", mf.Metadata.Name)
		}
		if err = DeleteClusterByName(mf.Spec.RootLB, mf.Metadata.Name); err != nil {
			return fmt.Errorf("can't remove cluster, %v", err)
		}
		delete(MEXGUIDMap, mf.Metadata.Name)
		log.Debugln(MEXGUIDMap)
	case gcloudGKE:
		if err = gcloud.DeleteGKECluster(mf.Metadata.Name); err != nil {
			return err
		}
	case azureAKS:
		if err = azure.DeleteAKSCluster(mf.Metadata.ResourceGroup); err != nil {
			return err
		}
	}

	log.Debugln("removed cluster", mf)
	return nil
}

//MEXSetEnvVars sets up environment vars and checks for credentials required for running
func MEXSetEnvVars(mf *Manifest) error {
	// secrets to be passed via Env var still : MEX_CF_KEY, MEX_CF_USER, MEX_DOCKER_REG_PASS
	// TODO: use `secrets` or `vault`

	eCFKey := os.Getenv("MEX_CF_KEY")
	if eCFKey == "" {
		return fmt.Errorf("no MEX_CF_KEY")
	}
	eCFUser := os.Getenv("MEX_CF_USER")
	if eCFUser == "" {
		return fmt.Errorf("no MEX_CF_USER")
	}
	eMEXDockerRegPass := os.Getenv("MEX_DOCKER_REG_PASS")
	if eMEXDockerRegPass == "" {
		return fmt.Errorf("no MEX_DOCKER_REG_PASS")
	}

	if err := os.Setenv("MEX_ROOT_LB", mf.Metadata.Name); err != nil {
		return err
	}
	if err := os.Setenv("MEX_AGENT_IMAGE", mf.Spec.Agent.Image); err != nil {
		return err
	}
	if err := os.Setenv("MEX_ZONE", mf.Metadata.DNSZone); err != nil {
		return err
	}
	if err := os.Setenv("MEX_EXT_NETWORK", mf.Spec.ExternalNetwork); err != nil {
		return err
	}
	if err := os.Setenv("MEX_NETWORK", mf.Spec.InternalNetwork); err != nil {
		return err
	}
	if err := os.Setenv("MEX_EXT_ROUTER", mf.Spec.ExternalRouter); err != nil {
		return err
	}
	if err := os.Setenv("MEX_DOCKER_REGISTRY", mf.Spec.DockerRegistry); err != nil {
		return err
	}

	//TODO need to allow users to save the environment under platform name inside .mobiledgex
	return nil
}

var yamlMEXPlatform = `apiVersion: v1
kind: mex-openstack-kubernetes
resource: openstack-platform
metadata:
  kind: {{.Kind}}
  name: {{.Name}}
  rootlb: {{.RootLB}}
  tags: {{.Tags}}
  tenant: {{.Tenant}}
  operator: {{.Operator}}
  region: {{.Region}}
  zone: {{.Zone}}
  location: {{.Location}}
  openrc: ~/.mobiledgex/openrc
  dnszone: mobiledgex.net
spec:
  key: {{.Key}}
  dockerregistry: registry.mobiledgex.net:5000
  externalnetwork: external-network-shared
  internalnetwork: mex-k8s-net-1
  internalcidr: 10.101.101.0/24
  externalrouter: mex-k8s-router-1
  agent: 
    image: registry.mobiledgex.net:5000/mobiledgex/mexosagent
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

//MEXPlatformInitCloudletKey calls MEXPlatformInit with templated manifest
func MEXPlatformInitCloudletKey(rootLB, cloudletKeyStr string) error {
	//XXX fugure: trigger off cloudletKeyStr or flavor to pick the right template: mex, aks, gke
	clk := edgeproto.CloudletKey{}
	err := json.Unmarshal([]byte(cloudletKeyStr), &clk)
	if err != nil {
		return fmt.Errorf("can't unmarshal json cloudletkey %s, %v", cloudletKeyStr, err)
	}
	name := clk.Name
	operator := clk.OperatorKey.Name
	data := templateFill{
		Name:     name,
		Tags:     name + "-tag",
		Key:      cloudletKeyStr,
		Operator: operator,
		Location: "bonn",
		Region:   "eu-central-1",
		Zone:     "eu-central-1c",
		RootLB:   "mex-lb-1.mobiledgex.net",
		Kind:     "mex-tdg-openstack-kubernetes",
	}

	mf, err := templateUnmarshal(&data, yamlMEXPlatform)
	if err != nil {
		return err
	}
	return MEXPlatformInit(mf)
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
		log.Debugln(data)
		return nil, fmt.Errorf("can't execute template, %v", err)
	}
	mf := &Manifest{}
	err = yaml.Unmarshal(outbuffer.Bytes(), mf)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal templated data, %v", err)
	}

	return mf, nil
}

//MEXPlatformInit initializes platform
func MEXPlatformInit(mf *Manifest) error {
	log.Debug("init platform", mf)

	var err error

	switch mf.Kind {
	case mexOSKubernetes:
		err = MEXSetEnvVars(mf)
		if err != nil {
			return err
		}
		//TODO validate all mf content against platform data
		if err = RunMEXAgent(mf.Spec.RootLB, mf.Spec.Key, false); err != nil {
			return err
		}
	case gcloudGKE:
	case azureAKS:
	}
	return nil
}

//MEXPlatformClean cleans up the platform
func MEXPlatformClean(mf *Manifest) error {
	log.Debugln("clean platform", mf)

	var err error

	switch mf.Kind {
	case mexOSKubernetes:
		err = MEXSetEnvVars(mf)
		if err != nil {
			return err
		}
		if err = RemoveMEXAgent(mf.Metadata.Name); err != nil {
			return err
		}
	case gcloudGKE:
	case azureAKS:
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

func MEXCreateAppInst(clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst) error {
	appname := appInst.Key.AppKey.Name
	rootLB := GetRootLBName()
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
	case "ImageTypeDocker": //XXX does not distinguish plain docker vs kubernetes
		data = templateFill{
			Kind:        "mex-app-docker",
			Name:        appname,
			Tags:        appname + "-docker-tag",
			Key:         clusterInst.Key.ClusterKey.Name,
			Tenant:      appname + "-tenant",
			RootLB:      rootLB,
			Registry:    "registry.mobiledgex.net:5000",
			Image:       appInst.ImagePath,      //XXX docker image name?
			ProxyPath:   rootLB + "/" + appname, //appInst.Uri ?
			PortMap:     appInst.MappedPorts,    //XXX format
			PathMap:     appInst.MappedPath,     //XXX format
			AppFlavor:   appInst.Flavor.Name,    //XXX not sure what this is
			AccessLayer: accessLayer,
			Kubernetes:  "https://k8s.io/examples/application/deployment.yaml",
		}
	case "ImageTypeQCOW":
		data = templateFill{
			Kind:        "mex-app-virtual-machine",
			Name:        appname,
			Tags:        appname + "-qcow-tag",
			Key:         clusterInst.Key.ClusterKey.Name,
			Tenant:      appname + "-tenant",
			RootLB:      rootLB,
			Registry:    "registry.mobiledgex.net:5000",
			Image:       appInst.ImagePath,      //XXX qcow image name?
			ProxyPath:   rootLB + "/" + appname, //appInst.Uri ?
			PortMap:     appInst.MappedPorts,    //XXX format
			PathMap:     appInst.MappedPath,     //XXX format
			AppFlavor:   appInst.Flavor.Name,    //XXX not sure what this is
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
	return MEXCreateApp(mf)
}

func MEXCreateApp(mf *Manifest) error {
	log.Debugln("create app", mf)

	switch mf.Kind {
	case mexOSKubernetes:
		if strings.Contains(mf.Metadata.Tags, "docker") {
			if mf.Spec.Kubernetes == "" { //plain docker
				return CreateDockerApp(mf.Spec.RootLB,
					mf.Metadata.Name,
					mf.Spec.Key,
					mf.Spec.Flavor,
					mf.Spec.DockerRegistry,
					mf.Spec.ProxyPath,
					mf.Spec.Image,
					mf.Spec.PortMap,
					mf.Spec.PathMap,
					mf.Spec.AccessLayer,
				)
			} else {
				return CreateKubernetesApp(mf.Spec.RootLB,
					mf.Spec.Key,
					mf.Metadata.Name,
					mf.Spec.Kubernetes,
				)
			}
		} else if strings.Contains(mf.Metadata.Tags, "qcow") {
		} else {
			return fmt.Errorf("insufficient tag info")
		}
	case gcloudGKE:
	case azureAKS:
	}
	return nil
}
