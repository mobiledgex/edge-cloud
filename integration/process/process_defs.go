package process

import "os/exec"

type Vault struct {
	Common    `yaml:",inline"`
	DmeSecret string
	cmd       *exec.Cmd
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
	Common                `yaml:",inline"`
	EtcdAddrs             string
	ApiAddr               string
	HttpAddr              string
	NotifyAddr            string
	NotifyRootAddrs       string
	NotifyParentAddrs     string
	VaultAddr             string
	InfluxAddr            string
	TLS                   TLSCerts
	cmd                   *exec.Cmd
	TestMode              bool
	RegistryFQDN          string
	ArtifactoryFQDN       string
	CloudletRegistryPath  string
	VersionTag            string
	CloudletVMImagePath   string
	CheckpointingInterval string
}
type Dme struct {
	Common      `yaml:",inline"`
	ApiAddr     string
	HttpAddr    string
	NotifyAddrs string
	LocVerUrl   string
	TokSrvUrl   string
	QosPosUrl   string
	Carrier     string
	CloudletKey string
	VaultAddr   string
	CookieExpr  string
	TLS         TLSCerts
	cmd         *exec.Cmd
}
type Crm struct {
	Common              `yaml:",inline"`
	NotifyAddrs         string
	NotifySrvAddr       string
	CloudletKey         string
	Platform            string
	Plugin              string
	TLS                 TLSCerts
	cmd                 *exec.Cmd
	VaultAddr           string
	PhysicalName        string
	TestMode            bool
	Span                string
	CleanupMode         bool
	ContainerVersion    string
	VMImageVersion      string
	PackageVersion      string
	CloudletVMImagePath string
	Region              string
	CommercialCerts     bool
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
	Config   string // set during Start
	TLS      TLSCerts
	Auth     LocalAuth
	cmd      *exec.Cmd
}
type ClusterSvc struct {
	Common         `yaml:",inline"`
	NotifyAddrs    string
	CtrlAddrs      string
	PromPorts      string
	InfluxDB       string
	Interval       string
	PluginRequired bool
	TLS            TLSCerts
	cmd            *exec.Cmd
}
type Jaeger struct {
	Common `yaml:",inline"`
	TLS    TLSCerts
	cmd    *exec.Cmd
}
type Traefik struct {
	Common `yaml:",inline"`
	TLS    TLSCerts
	cmd    *exec.Cmd
}
type NotifyRoot struct {
	Common `yaml:",inline"`
	TLS    TLSCerts
	cmd    *exec.Cmd
}
