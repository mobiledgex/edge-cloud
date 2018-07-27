package crmutil

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
}

type NetworkType struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	CIDR    string `json:"cidr"`
	Options string `json:"options"`
	Extra   string `json:"extra"`
}

type AgentType struct {
	Image  string `json:"image"`
	Status string `json:"status"`
}

type ImageType struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Favorite string `json:"favorite"`
	OSFlavor string `json:"osflavor"`
}

type FlavorType struct {
	Name          string `json:"name"`
	CPUs          int    `json:"cpus"`
	Memory        string `json:"memory"`
	Storage       string `json:"storage"`
	Masters       int    `json:"masters"`
	Nodes         int    `json:"nodes"`
	Favorite      string `json:"favorite"`
	NetworkScheme string `json:"networkscheme"`
}

type SpecType struct {
	Flavor          string        `json:"flavor"`
	RootLB          string        `json:"rootlb"`
	Networks        []NetworkType `json:"networks"`
	Agent           AgentType     `json:"agent"`
	Images          []ImageType   `json:"images"`
	Flavors         []FlavorType  `json:"flavors"`
	DockerRegistry  string        `json:"dockerregistry"`
	ExternalNetwork string        `json:"externalnetwork"`
	InternalNetwork string        `json:"internalnetwork"`
	InternalCIDR    string        `json:"internalcidr"`
	ExternalRouter  string        `json:"externalrouter"`
}

type Manifest struct {
	APIVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Resource   string       `json:"resource"`
	Metadata   MetadataType `json:"metadata"`
	Spec       SpecType     `json:"spec"`
}
