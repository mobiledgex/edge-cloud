package process

type Vault struct {
	Common      `yaml:",inline"`
	DmeSecret   string
	McormSecret string
}
type Etcd struct {
	Common         `yaml:",inline"`
	DataDir        string
	PeerAddrs      string
	ClientAddrs    string
	InitialCluster string
}
type Controller struct {
	Common        `yaml:",inline"`
	EtcdAddrs     string
	ApiAddr       string
	HttpAddr      string
	NotifyAddr    string
	TLS           TLSCerts
	ShortTimeouts bool
}
type Dme struct {
	Common      `yaml:",inline"`
	ApiAddr     string
	HttpAddr    string
	NotifyAddrs string
	LocVerUrl   string
	TokSrvUrl   string
	Carrier     string
	CloudletKey string
	VaultAddr   string
	CookieExpr  string
	TLS         TLSCerts
}
type Crm struct {
	Common      `yaml:",inline"`
	ApiAddr     string
	NotifyAddrs string
	CloudletKey string
	Platform    string
	Plugin      string
	TLS         TLSCerts
}
type MC struct {
	Common    `yaml:",inline"`
	Addr      string
	SqlAddr   string
	VaultAddr string
	RolesFile string
	TLS       TLSCerts
}
type Sql struct {
	Common   `yaml:",inline"`
	DataDir  string
	HttpAddr string
	Username string
	Dbname   string
	TLS      TLSCerts
}
type LocApiSim struct {
	Common  `yaml:",inline"`
	Port    int
	Locfile string
	Geofile string
	Country string
}
type TokSrvSim struct {
	Common `yaml:",inline"`
	Port   int
	Token  string
}
type SampleApp struct {
	Common       `yaml:",inline"`
	Exename      string
	Args         []string
	Command      string
	VolumeMounts []string
}
type Influx struct {
	Common   `yaml:",inline"`
	DataDir  string
	HttpAddr string
	Config   string // set during Start
}
type ClusterSvc struct {
	Common      `yaml:",inline"`
	NotifyAddrs string
	CtrlAddrs   string
	PromPorts   string
	InfluxDB    string
	Interval    string
	TLS         TLSCerts
}
