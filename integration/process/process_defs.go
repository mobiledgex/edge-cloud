package process

import "os/exec"

type Vault struct {
	Common     `yaml:",inline"`
	DmeSecret  string
	Regions    string
	VaultDatas []VaultData
	cmd        *exec.Cmd
}

type VaultData struct {
	Path string
	Data map[string]string
}

type Etcd struct {
	Common         `yaml:",inline"`
	DataDir        string
	PeerAddrs      string
	ClientAddrs    string
	InitialCluster string
	cmd            *exec.Cmd
}
type Controller struct {
	Common               `yaml:",inline"`
	NodeCommon           `yaml:",inline"`
	EtcdAddrs            string
	ApiAddr              string
	HttpAddr             string
	NotifyAddr           string
	NotifyRootAddrs      string
	NotifyParentAddrs    string
	EdgeTurnAddr         string
	InfluxAddr           string
	Region               string
	cmd                  *exec.Cmd
	TestMode             bool
	RegistryFQDN         string
	ArtifactoryFQDN      string
	CloudletRegistryPath string
	VersionTag           string
	CloudletVMImagePath  string
	CheckpointInterval   string
	AppDNSRoot           string
	ChefServerPath       string
}
type Dme struct {
	Common      `yaml:",inline"`
	NodeCommon  `yaml:",inline"`
	ApiAddr     string
	HttpAddr    string
	NotifyAddrs string
	LocVerUrl   string
	TokSrvUrl   string
	QosPosUrl   string
	Carrier     string
	CloudletKey string
	CookieExpr  string
	Region      string
	cmd         *exec.Cmd
}
type Crm struct {
	Common              `yaml:",inline"`
	NodeCommon          `yaml:",inline"`
	NotifyAddrs         string
	NotifySrvAddr       string
	CloudletKey         string
	Platform            string
	Plugin              string
	cmd                 *exec.Cmd
	PhysicalName        string
	TestMode            bool
	Span                string
	ContainerVersion    string
	VMImageVersion      string
	CloudletVMImagePath string
	Region              string
	CommercialCerts     bool
	AppDNSRoot          string
	ChefServerPath      string
	CacheDir            string
}
type LocApiSim struct {
	Common  `yaml:",inline"`
	Port    int
	Locfile string
	Geofile string
	Country string
	cmd     *exec.Cmd
}
type TokSrvSim struct {
	Common `yaml:",inline"`
	Port   int
	Token  string
	cmd    *exec.Cmd
}
type SampleApp struct {
	Common       `yaml:",inline"`
	Exename      string
	Args         []string
	Command      string
	VolumeMounts []string
	cmd          *exec.Cmd
}
type Influx struct {
	Common   `yaml:",inline"`
	DataDir  string
	HttpAddr string
	BindAddr string
	Config   string // set during Start
	TLS      TLSCerts
	Auth     LocalAuth
	cmd      *exec.Cmd
}
type ClusterSvc struct {
	Common         `yaml:",inline"`
	NodeCommon     `yaml:",inline"`
	NotifyAddrs    string
	CtrlAddrs      string
	PromPorts      string
	InfluxDB       string
	Interval       string
	Region         string
	PluginRequired bool
	cmd            *exec.Cmd
}
type DockerGeneric struct {
	Common        `yaml:",inline"`
	Links         []string
	DockerEnvVars map[string]string
	TLS           TLSCerts
	cmd           *exec.Cmd
}
type Jaeger struct {
	DockerGeneric `yaml:",inline"`
}
type ElasticSearch struct {
	DockerGeneric `yaml:",inline"`
	Type          string
}
type Traefik struct {
	Common `yaml:",inline"`
	TLS    TLSCerts
	cmd    *exec.Cmd
}
type NotifyRoot struct {
	Common     `yaml:",inline"`
	NodeCommon `yaml:",inline"`
	cmd        *exec.Cmd
}
type EdgeTurn struct {
	Common     `yaml:",inline"`
	NodeCommon `yaml:",inline"`
	cmd        *exec.Cmd
	ListenAddr string
	ProxyAddr  string
	Region     string
	TestMode   bool
}
