package crmutil

//MetadataType has metadata
type MetadataType struct {
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
	Key           string `json:"key"`
}

//NetworkType has network data
type NetworkType struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	CIDR    string `json:"cidr"`
	Options string `json:"options"`
	Extra   string `json:"extra"`
}

//AgentType has data on agent
type AgentType struct {
	Image  string `json:"image"`
	Status string `json:"status"`
}

//ImageType has data on image
type ImageType struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Favorite string `json:"favorite"`
	OSFlavor string `json:"osflavor"`
}

//FlavorType has data on flavor
type FlavorType struct {
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

//SpecType holds spec block
type SpecType struct {
	Flavor          string        `json:"flavor"`
	RootLB          string        `json:"rootlb"`
	DockerRegistry  string        `json:"dockerregistry"`
	ExternalNetwork string        `json:"externalnetwork"`
	InternalNetwork string        `json:"internalnetwork"`
	InternalCIDR    string        `json:"internalcidr"`
	ExternalRouter  string        `json:"externalrouter"`
	Networks        []NetworkType `json:"networks"`
	Images          []ImageType   `json:"images"`
	Flavors         []FlavorType  `json:"flavors"`
	Agent           AgentType     `json:"agent"`
}

//Manifest is general container for the manifest yaml used by `mex`
type Manifest struct {
	APIVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Resource   string       `json:"resource"`
	Metadata   MetadataType `json:"metadata"`
	Spec       SpecType     `json:"spec"`
}
