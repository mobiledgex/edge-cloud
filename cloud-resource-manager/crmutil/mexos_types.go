package crmutil

//MetadataDetail has metadata
type MetadataDetail struct {
	Name          string `json:"name"`
	Tags          string `json:"tags"`
	Tenant        string `json:"tenant"`
	Region        string `json:"region"`
	Zone          string `json:"zone"`
	Location      string `json:"location"`
	Project       string `json:"project"`
	ResourceGroup string `json:"resourcegroup"`
	OpenRC        string `json:"openrc"`
	DNSZone       string `json:"dnszone"`
	Kind          string `json:"kind"`
	Operator      string `json:"operator"`
}

//NetworkDetail has network data
type NetworkDetail struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	CIDR    string `json:"cidr"`
	Options string `json:"options"`
	Extra   string `json:"extra"`
}

//AgentDetail has data on agent
type AgentDetail struct {
	Image  string `json:"image"`
	Status string `json:"status"`
}

//FlavorDetail has data on flavor
type FlavorDetail struct {
	Name          string `json:"name"`
	Favorite      string `json:"favorite"`
	Memory        string `json:"memory"`
	Topology      string `json:"topology"`
	NodeFlavor    string `json:"nodeflavor"`
	MasterFlavor  string `json:"masterflavor"`
	NetworkScheme string `json:"networkscheme"`
	Storage       string `json:"storage"`
	StorageScheme string `json:"storagescheme"`
	CPUs          int    `json:"cpus"`
	Masters       int    `json:"masters"`
	Nodes         int    `json:"nodes"`
}

type FlavorDetailInfo struct {
	Name          string `json:"name"`
	Nodes         int    `json:"nodes"`
	Masters       int    `json:"masters"`
	NetworkScheme string `json:"networkscheme"`
	MasterFlavor  string `json:"masterflavor"`
	NodeFlavor    string `json:"nodeflavor"`
	StorageScheme string `json:"storagescheme"`
	Topology      string `json:"topology"`
}

type PortDetail struct {
	Name         string `json:"name"`
	MexProto     string `json:"mexproto"`
	Proto        string `json:"proto"`
	InternalPort int    `json:"internalport"`
	PublicPort   int    `json:"publicport"`
	PublicPath   string `json:"publicpath"`
}

//SpecDetail holds spec block
type SpecDetail struct {
	Flavor               string           `json:"flavor"` // appInst flavor?
	FlavorDetail         FlavorDetailInfo `json:"flavordetail"`
	Flags                string           `json:"flags"`
	RootLB               string           `json:"rootlb"`
	Image                string           `json:"image"`
	ImageFlavor          string           `json:"imageflavor"`
	ImageType            string           `json:"imagetype"`
	DockerRegistry       string           `json:"dockerregistry"`
	ExternalNetwork      string           `json:"externalnetwork"`
	ExternalRouter       string           `json:"externalrouter"`
	Options              string           `json:"options"`
	ProxyPath            string           `json:"proxypath"`
	Ports                []PortDetail     `json:"ports"`
	Command              []string         `json:"command"`
	IpAccess             string           `json:"ipaccess"`
	URI                  string           `json:"uri"`
	Key                  string           `json:"key"`
	KubeManifestTemplate string           `json:"kubemanifesttemplate"`
	NetworkScheme        string           `json:"networkscheme"`
	Agent                AgentDetail      `json:"agent"`
}

//Manifest is general container for the manifest yaml used by `mex`
type Manifest struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Resource   string         `json:"resource"`
	Metadata   MetadataDetail `json:"metadata"`
	Spec       SpecDetail     `json:"spec"`
}
