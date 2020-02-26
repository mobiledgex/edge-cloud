package process

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	influxclient "github.com/influxdata/influxdb/client/v2"
	"github.com/mobiledgex/edge-cloud/cloudcommon/influxsup"
	mextls "github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"
)

// Local processes all run in the same global namespace, using different
// tcp ports to communicate with each other.

type TLSCerts struct {
	ServerCert string
	ServerKey  string
	CACert     string
	ClientCert string
	ApiCert    string
	ApiKey     string
}

type LocalAuth struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

var InfluxCredsFile = "/tmp/influx.json"

// EtcdLocal

func (p *Etcd) StartLocal(logfile string, opts ...StartOp) error {
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.CleanStartup {
		if err := p.ResetData(); err != nil {
			return err
		}
	}

	args := []string{"--name", p.Name, "--data-dir", p.DataDir, "--listen-peer-urls", p.PeerAddrs, "--listen-client-urls", p.ClientAddrs, "--advertise-client-urls", p.ClientAddrs, "--initial-advertise-peer-urls", p.PeerAddrs, "--initial-cluster", p.InitialCluster}

	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, nil, logfile)
	return err
}

func (p *Etcd) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Etcd) GetExeName() string { return "etcd" }

func (p *Etcd) LookupArgs() string { return "--name " + p.Name }

func (p *Etcd) ResetData() error {
	return os.RemoveAll(p.DataDir)
}

// ControllerLocal

func (p *Controller) StartLocal(logfile string, opts ...StartOp) error {
	args := []string{"--etcdUrls", p.EtcdAddrs, "--notifyAddr", p.NotifyAddr}
	if p.ApiAddr != "" {
		args = append(args, "--apiAddr")
		args = append(args, p.ApiAddr)
	}
	if p.HttpAddr != "" {
		args = append(args, "--httpAddr")
		args = append(args, p.HttpAddr)
	}
	if p.TLS.ServerCert != "" {
		args = append(args, "--tls")
		args = append(args, p.TLS.ServerCert)
	}
	if p.InfluxAddr != "" {
		args = append(args, "--influxAddr")
		args = append(args, p.InfluxAddr)
	}
	if p.VaultAddr != "" {
		args = append(args, "--vaultAddr")
		args = append(args, p.VaultAddr)
	}
	if p.RegistryFQDN != "" {
		args = append(args, "--registryFQDN")
		args = append(args, p.RegistryFQDN)
	}
	if p.ArtifactoryFQDN != "" {
		args = append(args, "--artifactoryFQDN")
		args = append(args, p.ArtifactoryFQDN)
	}
	if p.CloudletRegistryPath != "" {
		args = append(args, "--cloudletRegistryPath")
		args = append(args, p.CloudletRegistryPath)
	}
	if p.CloudletVMImagePath != "" {
		args = append(args, "--cloudletVMImagePath")
		args = append(args, p.CloudletVMImagePath)
	}
	if p.NotifyParentAddrs != "" {
		args = append(args, "--notifyParentAddrs")
		args = append(args, p.NotifyParentAddrs)
	}
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	if p.TestMode {
		args = append(args, "-testMode")
	}
	if p.VersionTag != "" {
		args = append(args, "--versionTag")
		args = append(args, p.VersionTag)
	}

	var envs []string
	if options.RolesFile != "" {
		dat, err := ioutil.ReadFile(options.RolesFile)
		if err != nil {
			return err
		}
		roles := VaultRoles{}
		err = yaml.Unmarshal(dat, &roles)
		if err != nil {
			return err
		}
		envs = []string{
			fmt.Sprintf("VAULT_ROLE_ID=%s", roles.CtrlRoleID),
			fmt.Sprintf("VAULT_SECRET_ID=%s", roles.CtrlSecretID),
		}
		log.Printf("controller envs: %v\n", envs)
	}

	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, envs, logfile)
	return err
}

func (p *Controller) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Controller) GetExeName() string { return "controller" }

func (p *Controller) LookupArgs() string { return "--apiAddr " + p.ApiAddr }

func getRestClientImpl(timeout time.Duration, addr string, tlsConfig *tls.Config) (*http.Client, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}
	return client, nil
}

func connectAPIImpl(timeout time.Duration, apiaddr string, tlsConfig *tls.Config) (*grpc.ClientConn, error) {
	// Wait for service to be ready to connect.
	// Note: using grpc WithBlock() takes about a second longer
	// than doing the retry connect below so requires a larger timeout.
	startTimeMs := time.Now().UnixNano() / int64(time.Millisecond)
	maxTimeMs := int64(timeout/time.Millisecond) + startTimeMs
	wait := 20 * time.Millisecond
	for {
		_, err := net.Dial("tcp", apiaddr)
		currTimeMs := time.Now().UnixNano() / int64(time.Millisecond)

		if currTimeMs > maxTimeMs {
			err := errors.New("Timeout in connection to " + apiaddr)
			log.Printf("Error: %v\n", err)
			return nil, err
		}
		if err == nil {
			break
		}
		timeout -= wait
		time.Sleep(wait)
	}
	conn, err := grpc.Dial(apiaddr, mextls.GetGrpcDialOption(tlsConfig))
	return conn, err
}

func (p *Controller) ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error) {
	tlsConfig, err := mextls.GetTLSClientConfig(p.ApiAddr, p.TLS.ClientCert, "", false)
	if err != nil {
		return nil, err
	}
	return connectAPIImpl(timeout, p.ApiAddr, tlsConfig)
}

// DmeLocal

func (p *Dme) StartLocal(logfile string, opts ...StartOp) error {
	args := []string{"--notifyAddrs", p.NotifyAddrs}
	if p.ApiAddr != "" {
		args = append(args, "--apiAddr")
		args = append(args, p.ApiAddr)
	}
	if p.HttpAddr != "" {
		args = append(args, "--httpAddr")
		args = append(args, p.HttpAddr)
	}
	if p.LocVerUrl != "" {
		args = append(args, "--locverurl")
		args = append(args, p.LocVerUrl)
	}
	if p.TokSrvUrl != "" {
		args = append(args, "--toksrvurl")
		args = append(args, p.TokSrvUrl)
	}
	if p.QosPosUrl != "" {
		args = append(args, "--qosposurl")
		args = append(args, p.QosPosUrl)
	}
	if p.Carrier != "" {
		args = append(args, "--carrier")
		args = append(args, p.Carrier)
	}
	if p.CloudletKey != "" {
		args = append(args, "--cloudletKey")
		args = append(args, p.CloudletKey)
	}
	if p.TLS.ServerCert != "" {
		args = append(args, "--tls")
		args = append(args, p.TLS.ServerCert)
	}
	if p.TLS.ServerCert != "" && p.TLS.ServerKey != "" {
		if p.TLS.ApiCert != "" {
			args = append(args, "--tlsApiCertFile", p.TLS.ApiCert)
			args = append(args, "--tlsApiKeyFile", p.TLS.ApiKey)
		} else {
			args = append(args, "--tlsApiCertFile", p.TLS.ServerCert)
			args = append(args, "--tlsApiKeyFile", p.TLS.ServerKey)
		}
	}
	if p.VaultAddr != "" {
		args = append(args, "--vaultAddr")
		args = append(args, p.VaultAddr)
	}
	if p.CookieExpr != "" {
		args = append(args, "--cookieExpiration")
		args = append(args, p.CookieExpr)
	}
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	var envs []string
	if options.RolesFile != "" {
		dat, err := ioutil.ReadFile(options.RolesFile)
		if err != nil {
			return err
		}
		roles := VaultRoles{}
		err = yaml.Unmarshal(dat, &roles)
		if err != nil {
			return err
		}
		envs = []string{
			fmt.Sprintf("VAULT_ROLE_ID=%s", roles.DmeRoleID),
			fmt.Sprintf("VAULT_SECRET_ID=%s", roles.DmeSecretID),
		}
		log.Printf("dme envs: %v\n", envs)
	}

	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, envs, logfile)
	return err
}

func (p *Dme) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Dme) GetExeName() string { return "dme-server" }

func (p *Dme) LookupArgs() string { return "--apiAddr " + p.ApiAddr }

func (p *Dme) ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error) {
	return connectAPIImpl(timeout, p.ApiAddr, p.getTlsConfig())
}

func (p *Dme) GetRestClient(timeout time.Duration) (*http.Client, error) {
	return getRestClientImpl(timeout, p.HttpAddr, p.getTlsConfig())
}

func (p *Dme) getTlsConfig() *tls.Config {
	if p.TLS.ServerCert != "" && p.TLS.ServerKey != "" {
		// ServerAuth TLS. For real clients, they'll use
		// their built-in trusted CAs to verify the cert.
		// Since we're using self-signed certs here, add
		// our CA to the cert pool.
		certPool, err := mextls.GetClientCertPool(p.TLS.ServerCert, "")
		if err != nil {
			log.Printf("GetClientCertPool failed, %v\n", err)
			return nil
		}
		config := &tls.Config{
			RootCAs: certPool,
		}
		return config
	}
	// no TLS
	return nil
}

// CrmLocal

func (p *Crm) GetArgs(opts ...StartOp) []string {
	args := []string{"--notifyAddrs", p.NotifyAddrs}
	if p.NotifySrvAddr != "" {
		args = append(args, "--notifySrvAddr")
		args = append(args, p.NotifySrvAddr)
	}
	if p.CloudletKey != "" {
		args = append(args, "--cloudletKey")
		args = append(args, p.CloudletKey)
	}
	if p.TLS.ServerCert != "" {
		args = append(args, "--tls")
		args = append(args, p.TLS.ServerCert)
	}
	if p.Name != "" {
		args = append(args, "--hostname")
		args = append(args, p.Name)
	}
	if p.Platform != "" {
		args = append(args, "--platform")
		args = append(args, p.Platform)
	}
	if p.Plugin != "" {
		args = append(args, "--plugin")
		args = append(args, p.Plugin)
	}
	if p.VaultAddr != "" {
		args = append(args, "--vaultAddr")
		args = append(args, p.VaultAddr)
	}
	if p.PhysicalName != "" {
		args = append(args, "--physicalName")
		args = append(args, p.PhysicalName)
	}
	if p.Span != "" {
		args = append(args, "--span")
		args = append(args, p.Span)
	}
	if p.TestMode {
		args = append(args, "-testMode")
	}
	if p.CleanupMode {
		args = append(args, "-cleanupMode")
	}
	if p.ContainerVersion != "" {
		args = append(args, "--containerVersion")
		args = append(args, p.ContainerVersion)
	}
	if p.CloudletVMImagePath != "" {
		args = append(args, "--cloudletVMImagePath")
		args = append(args, p.CloudletVMImagePath)
	}
	if p.VMImageVersion != "" {
		args = append(args, "--vmImageVersion")
		args = append(args, p.VMImageVersion)
	}
	if p.PackageVersion != "" {
		args = append(args, "--packageVersion")
		args = append(args, p.PackageVersion)
	}
	if p.Region != "" {
		args = append(args, "--region")
		args = append(args, p.Region)
	}
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	return args
}

func (p *Crm) StartLocal(logfile string, opts ...StartOp) error {
	var err error

	args := p.GetArgs(opts...)
	envVars := []string{}
	for k, v := range p.GetEnvVars() {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, envVars, logfile)
	return err
}

func (p *Crm) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Crm) Wait() error {
	return p.cmd.Wait()
}

func (p *Crm) GetExeName() string { return "crmserver" }

func (p *Crm) LookupArgs() string { return "--cloudletKey " + p.CloudletKey }

func (p *Crm) String(opts ...StartOp) string {
	cmd_str := p.GetExeName()
	args := p.GetArgs(opts...)
	key := true
	for _, v := range args {
		if key {
			cmd_str += " " + v
			key = false
		} else {
			cmd_str += " '" + v + "'"
			key = true
		}
	}
	return cmd_str
}

// InfluxLocal

func (p *Influx) StartLocal(logfile string, opts ...StartOp) error {
	var prefix string
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.CleanStartup {
		if err := p.ResetData(); err != nil {
			return err
		}
	}

	configFile, err := influxsup.SetupInflux(p.DataDir,
		influxsup.WithSeverCert(p.TLS.ServerCert), influxsup.WithSeverCertKey(p.TLS.ServerKey), influxsup.WithAuth(p.Auth.User != ""))
	if err != nil {
		return err
	}
	p.Config = configFile
	args := []string{"-config", configFile}
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, nil, logfile)
	if err != nil {
		return err
	}

	// if auth is enabled we need to create default user
	if p.Auth.User != "" {
		time.Sleep(5 * time.Second)
		if p.TLS.ServerCert != "" {
			prefix = "https://" + p.HttpAddr
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		} else {
			prefix = "http://" + p.HttpAddr
		}

		resource := "/query"
		data := url.Values{}
		data.Set("q", "CREATE USER "+p.Auth.User+" WITH PASSWORD '"+p.Auth.Pass+"' WITH ALL PRIVILEGES")
		u, _ := url.ParseRequestURI(prefix)
		u.Path = resource
		u.RawQuery = data.Encode()
		urlStr := fmt.Sprintf("%v", u)
		client := &http.Client{}
		r, _ := http.NewRequest("POST", urlStr, nil)

		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
		fmt.Printf("Query: %s\n", urlStr)
		_, err := client.Do(r)
		if err != nil {
			p.StopLocal()
			return err
		}
	}
	// create auth file for Vault
	creds_json, err := json.Marshal(p.Auth)
	if err != nil {
		p.StopLocal()
		return err
	}
	return ioutil.WriteFile(InfluxCredsFile, creds_json, 0644)
}

func (p *Influx) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Influx) GetExeName() string { return "influxd" }

func (p *Influx) LookupArgs() string { return "" }

func (p *Influx) ResetData() error {
	return os.RemoveAll(p.DataDir)
}

func (p *Influx) GetClient() (influxclient.Client, error) {
	return influxsup.GetClient(p.HttpAddr, p.Auth.User, p.Auth.Pass)
}

// ClusterSvc process

func (p *ClusterSvc) StartLocal(logfile string, opts ...StartOp) error {
	args := []string{"--notifyAddrs", p.NotifyAddrs}
	if p.CtrlAddrs != "" {
		args = append(args, "--ctrlAddrs")
		args = append(args, p.CtrlAddrs)
	}
	if p.TLS.ServerCert != "" {
		args = append(args, "--tls")
		args = append(args, p.TLS.ServerCert)
	}
	if p.PromPorts != "" {
		args = append(args, "--prometheus-ports")
		args = append(args, p.PromPorts)
	}
	if p.InfluxDB != "" {
		args = append(args, "--influxdb")
		args = append(args, p.InfluxDB)
	}
	if p.Interval != "" {
		args = append(args, "--scrapeInterval")
		args = append(args, p.Interval)
	}
	if p.PluginRequired {
		args = append(args, "--pluginRequired")
	}
	args = append(args, "--hostname", p.Name)
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	// Append extra args convert from [arg1=val1, arg2] into ["-arg1", "val1", "-arg2"]
	if len(options.ExtraArgs) > 0 {
		for _, v := range options.ExtraArgs {
			tmp := strings.Split(v, "=")
			args = append(args, "-"+tmp[0])
			if len(tmp) > 1 {
				args = append(args, tmp[1])
			}
		}
	}

	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, nil, logfile)
	return err
}

func (p *ClusterSvc) StopLocal() {
	StopLocal(p.cmd)
}

func (p *ClusterSvc) GetExeName() string { return "cluster-svc" }

func (p *ClusterSvc) LookupArgs() string { return p.Name }

// Vault

// In dev mode, Vault is locked to below address
var VaultAddress = "http://127.0.0.1:8200"

type VaultRoles struct {
	DmeRoleID       string `json:"dmeroleid"`
	DmeSecretID     string `json:"dmesecretid"`
	CRMRoleID       string `json:"crmroleid"`
	CRMSecretID     string `json:"crmsecretid"`
	RotatorRoleID   string `json:"rotatorroleid"`
	RotatorSecretID string `json:"rotatorsecretid"`
	CtrlRoleID      string `json:"controllerroleid"`
	CtrlSecretID    string `json:"controllersecretid"`
}

func (p *Vault) StartLocal(logfile string, opts ...StartOp) error {
	// Note: for e2e tests, vault is started in dev mode.
	// In dev mode, vault is automatically unsealed, TLS is disabled,
	// data is in-memory only, and root key is printed during startup.
	// DO NOT run Vault in dev mode for production setups.
	if p.DmeSecret == "" {
		p.DmeSecret = "dme-secret"
	}

	args := []string{"server", "-dev"}
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, nil, logfile)
	if err != nil {
		return err
	}
	// wait until vault is online and ready
	for ii := 0; ii < 10; ii++ {
		var serr error
		p.Run("vault", "status", &serr)
		if serr == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// run setup script
	gopath := os.Getenv("GOPATH")
	region := "local"
	setup := gopath + "/src/github.com/mobiledgex/edge-cloud/vault/setup.sh " + region
	out := p.Run("/bin/sh", setup, &err)
	if err != nil {
		fmt.Println(out)
	}
	// get roleIDs and secretIDs
	roles := VaultRoles{}
	p.GetAppRole(region, "dme", &roles.DmeRoleID, &roles.DmeSecretID, &err)
	p.GetAppRole(region, "crm", &roles.CRMRoleID, &roles.CRMSecretID, &err)
	p.GetAppRole(region, "rotator", &roles.RotatorRoleID, &roles.RotatorSecretID, &err)
	p.GetAppRole(region, "controller", &roles.CtrlRoleID, &roles.CtrlSecretID, &err)
	p.PutSecret(region, "dme", p.DmeSecret+"-old", &err)
	p.PutSecret(region, "dme", p.DmeSecret, &err)
	// Get the directory where the influx.json file is
	if _, serr := os.Stat(InfluxCredsFile); !os.IsNotExist(serr) {
		path := "secret/" + region + "/accounts/influxdb"
		p.PutSecretsJson(path, InfluxCredsFile, &err)
	}
	if err != nil {
		p.StopLocal()
		return err
	}
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.RolesFile != "" {
		roleYaml, err := yaml.Marshal(&roles)
		if err != nil {
			p.StopLocal()
			return err
		}
		err = ioutil.WriteFile(options.RolesFile, roleYaml, 0644)
		if err != nil {
			p.StopLocal()
			return err
		}
	}
	return err
}

func (p *Vault) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Vault) GetExeName() string { return "vault" }

func (p *Vault) LookupArgs() string { return "" }

func (p *Vault) GetAppRole(region, name string, roleID, secretID *string, err *error) {
	if *err != nil {
		return
	}
	if region != "" {
		name = region + "." + name
	}
	out := p.Run("vault", fmt.Sprintf("read auth/approle/role/%s/role-id", name), err)
	vals := p.mapVals(out)
	if val, ok := vals["role_id"]; ok {
		*roleID = val
	}
	out = p.Run("vault", fmt.Sprintf("write -f auth/approle/role/%s/secret-id", name), err)
	vals = p.mapVals(out)
	if val, ok := vals["secret_id"]; ok {
		*secretID = val
	}
}

func (p *Vault) PutSecretsJson(SecretsPath, jsonFile string, err *error) {
	p.Run("vault", fmt.Sprintf("kv put %s @%s", SecretsPath, jsonFile), err)
}

func (p *Vault) PutSecret(region, name, secret string, err *error) {
	if region != "" {
		region += "/"
	}
	p.Run("vault", fmt.Sprintf("kv put %sjwtkeys/%s secret=%s", region, name, secret), err)
}

func (p *Vault) Run(bin, args string, err *error) string {
	if *err != nil {
		return ""
	}
	cmd := exec.Command(bin, strings.Split(args, " ")...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("VAULT_ADDR=%s", VaultAddress))
	out, cerr := cmd.CombinedOutput()
	if cerr != nil {
		*err = fmt.Errorf("cmd '%s %s' failed, %s, %v", bin, args, string(out), cerr.Error())
		return string(out)
	}
	return string(out)
}

func (p *Vault) mapVals(resp string) map[string]string {
	vals := make(map[string]string)
	for _, line := range strings.Split(resp, "\n") {
		pair := strings.Fields(strings.TrimSpace(line))
		if len(pair) != 2 {
			continue
		}
		vals[pair[0]] = pair[1]
	}
	return vals
}

func (p *Vault) StartLocalRoles() (*VaultRoles, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	rolesfile := dir + "/roles.yaml"
	err = p.StartLocal(dir+"/vault.log", WithRolesFile(rolesfile))
	if err != nil {
		return nil, err
	}

	// rolesfile contains the roleIDs/secretIDs needed to access vault
	dat, err := ioutil.ReadFile(rolesfile)
	if err != nil {
		p.StopLocal()
		return nil, err
	}
	roles := VaultRoles{}
	err = yaml.Unmarshal(dat, &roles)
	if err != nil {
		p.StopLocal()
		return nil, err
	}
	return &roles, nil
}

func (p *Traefik) StartLocal(logfile string, opts ...StartOp) error {
	configDir := path.Dir(logfile) + "/traefik"
	if err := os.MkdirAll(configDir, 0777); err != nil {
		return err
	}
	certsDir := ""
	if p.TLS.ServerCert != "" && p.TLS.ServerKey != "" && p.TLS.CACert != "" {
		certsDir = path.Dir(p.TLS.ServerCert)
	}

	args := []string{
		"run", "--rm", "--name", p.Name,
		"-p", "8080:8080", // web UI
		"-p", "14268:14268", // jaeger collector
		"-p", "16686:16686", // jeager UI
		"-p", "16687:16687", // jeager UI insecure (for local debugging)
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", fmt.Sprintf("%s:/etc/traefik", configDir),
	}
	if certsDir != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/certs", certsDir))
	}
	args = append(args, "traefik:v2.0")

	staticArgs := TraefikStaticArgs{}

	// Traefik consists of a Static Config file, and zero or more
	// dynamic config files. Dynamic config files can be hot-reloaded.
	// The allowed contents of each type are different.
	// Entry points are configured statically, while routers, services,
	// etc are configured dynmically, either through a file provider
	// or docker provider (snooping on docker events).

	if p.TLS.ServerCert != "" && p.TLS.ServerKey != "" && p.TLS.CACert != "" {
		certsDir = path.Dir(p.TLS.ServerCert)
		args = append(args, "-v", fmt.Sprintf("%s:/certs", certsDir))
		dynArgs := TraefikDynArgs{
			ServerCert: path.Base(p.TLS.ServerCert),
			ServerKey:  path.Base(p.TLS.ServerKey),
			CACert:     path.Base(p.TLS.CACert),
		}
		dynFile := "dyn.yml"
		tmpl := template.Must(template.New("dyn").Parse(TraefikDynFile))
		f, err := os.Create(configDir + "/" + dynFile)
		if err != nil {
			return err
		}
		defer f.Close()

		out := bufio.NewWriter(f)
		err = tmpl.Execute(out, dynArgs)
		if err != nil {
			return err
		}
		out.Flush()
		staticArgs.DynFile = dynFile
	}

	tmpl := template.Must(template.New("st").Parse(TraefikStaticFile))
	f, err := os.Create(configDir + "/traefik.yml")
	if err != nil {
		return err
	}
	defer f.Close()

	out := bufio.NewWriter(f)
	err = tmpl.Execute(out, staticArgs)
	if err != nil {
		return err
	}
	out.Flush()

	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, nil, logfile)
	return err
}

func (p *Traefik) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Traefik) GetExeName() string { return "docker" }

func (p *Traefik) LookupArgs() string { return p.Name }

type TraefikStaticArgs struct {
	DynFile string
}

var TraefikStaticFile = `
providers:
  docker: {}
{{- if ne .DynFile ""}}
  file:
    watch: true
    filename: /etc/traefik/{{.DynFile}}
{{- end}}
log:
  level: debug
api:
  dashboard: true
  debug: true
entryPoints:
  jaeger-collector:
    address: :14268
  jaeger-ui:
    address: :16686
  jaeger-ui-insecure:
    address: :16687
`

type TraefikDynArgs struct {
	ServerCert string
	ServerKey  string
	CACert     string
}

var TraefikDynFile = `
tls:
  certificates:
  - certFile: /certs/{{.ServerCert}}
    keyFile: /certs/{{.ServerKey}}
  options:
    default:
      clientAuth:
        caFiles:
        - /certs/{{.CACert}}
        clientAuthType: RequireAndVerifyClientCert
  stores:
    default:
      defaultCertificate:
        certFile: /certs/{{.ServerCert}}
        keyFile: /certs/{{.ServerKey}}
`

func (p *Jaeger) StartLocal(logfile string, opts ...StartOp) error {
	// Jaeger does not support TLS, so we use traefik
	// as a sidecar reverse proxy to implement mTLS.
	// No Jaeger ports are exposed because traefik proxies requests
	// to Jaeger on the internal docker network.
	// However, in order for traefik to understand how to do so,
	// it checks the labels set on the Jaeger docker process.
	labels := []string{
		"traefik.http.routers.jaeger-ui.entrypoints=jaeger-ui",
		"traefik.http.routers.jaeger-ui.rule=PathPrefix(`/`)",
		"traefik.http.routers.jaeger-ui.service=jaeger-ui",
		"traefik.http.routers.jaeger-ui.tls=true",
		"traefik.http.routers.jaeger-c.entrypoints=jaeger-collector",
		"traefik.http.routers.jaeger-c.rule=PathPrefix(`/`)",
		"traefik.http.routers.jaeger-c.service=jaeger-c",
		"traefik.http.routers.jaeger-c.tls=true",
		"traefik.http.routers.jaeger-ui-notls.entrypoints=jaeger-ui-insecure",
		"traefik.http.routers.jaeger-ui-notls.rule=PathPrefix(`/`)",
		"traefik.http.routers.jaeger-ui-notls.service=jaeger-ui-notls",
		"traefik.http.services.jaeger-ui.loadbalancer.server.port=16686",
		"traefik.http.services.jaeger-c.loadbalancer.server.port=14268",
		"traefik.http.services.jaeger-ui-notls.loadbalancer.server.port=16686",
	}
	args := []string{
		"run", "--rm", "--name", p.Name,
	}
	for _, l := range labels {
		args = append(args, "-l", l)
	}
	args = append(args, "jaegertracing/all-in-one:1.13")

	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, nil, logfile)
	return err
}

func (p *Jaeger) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Jaeger) GetExeName() string { return "docker" }

func (p *Jaeger) LookupArgs() string { return p.Name }

func (p *NotifyRoot) StartLocal(logfile string, opts ...StartOp) error {
	args := []string{}
	if p.TLS.ServerCert != "" {
		args = append(args, "--tls")
		args = append(args, p.TLS.ServerCert)
	}
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, nil, logfile)
	return err
}

func (p *NotifyRoot) StopLocal() {
	StopLocal(p.cmd)
}

func (p *NotifyRoot) GetExeName() string { return "notifyroot" }

func (p *NotifyRoot) LookupArgs() string { return "" }

// Support funcs

func StartLocal(name, bin string, args, envs []string, logfile string) (*exec.Cmd, error) {
	cmd := exec.Command(bin, args...)
	if envs != nil {
		// Append to the current process's env
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, envs...)
	}
	if logfile == "" {
		writer := NewColorWriter(name)
		cmd.Stdout = writer
		cmd.Stderr = writer
	} else {
		fmt.Printf("Creating logfile %v\n", logfile)
		// open the out file for writing
		outfile, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Printf("ERROR Creating logfile %v -- %v\n", logfile, err)
			panic(err)
		}
		cmd.Stdout = outfile
		cmd.Stderr = outfile
	}

	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func StopLocal(cmd *exec.Cmd) {
	if cmd != nil {
		cmd.Process.Kill()
		cmd.Process.Wait()
	}
}

//Location API simulator

func (p *LocApiSim) StartLocal(logfile string, opts ...StartOp) error {
	if p.Locfile != "" {

	}

	args := []string{"-port", fmt.Sprintf("%d", p.Port), "-file", p.Locfile}
	if p.Geofile != "" {
		args = append(args, "-geo", p.Geofile)
	}
	if p.Country != "" {
		args = append(args, "-country", p.Country)
	}
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, nil, logfile)
	return err
}

func (p *LocApiSim) StopLocal() {
	StopLocal(p.cmd)
}

func (p *LocApiSim) GetExeName() string { return "loc-api-sim" }

func (p *LocApiSim) LookupArgs() string {
	return fmt.Sprintf("-port %d", p.Port)
}

//Token service simulator

func (p *TokSrvSim) StartLocal(logfile string, opts ...StartOp) error {
	args := []string{"-port", fmt.Sprintf("%d", p.Port)}
	if p.Token != "" {
		args = append(args, "-token")
		args = append(args, p.Token)
	}
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, nil, logfile)
	return err
}

func (p *TokSrvSim) StopLocal() {
	StopLocal(p.cmd)
}

func (p *TokSrvSim) GetExeName() string { return "tok-srv-sim" }

func (p *TokSrvSim) LookupArgs() string {
	return fmt.Sprintf("-port %d", p.Port)
}

//Generic sample app for use in test

func (p *SampleApp) StartLocal(logfile string, opts ...StartOp) error {
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), p.Args, nil, logfile)
	return err
}

func (p *SampleApp) StopLocal() {
	StopLocal(p.cmd)
}

func (p *SampleApp) GetExeName() string { return p.Exename }

func (p *SampleApp) LookupArgs() string {
	return strings.Join(p.Args, " ")
}

type ColorWriter struct {
	Name  string
	Color ct.Color
}

func (c *ColorWriter) Write(p []byte) (int, error) {
	buf := bytes.NewBuffer(p)
	printed := 0
	for {
		line, err := buf.ReadBytes('\n')
		if len(line) > 0 {
			ct.ChangeColor(c.Color, false, ct.None, false)
			fmt.Printf("%s : %s", c.Name, string(line))
			ct.ResetColor()
			printed += len(line)
		}
		if err != nil {
			if err != io.EOF {
				return printed, err
			}
			break
		}
	}
	return printed, nil
}

var nextColorIdx = 0
var nextColorMux sync.Mutex

var colors = []ct.Color{
	ct.Green,
	ct.Cyan,
	ct.Magenta,
	ct.Blue,
	ct.Red,
	ct.Yellow,
}

func NewColorWriter(name string) *ColorWriter {
	nextColorMux.Lock()
	color := colors[nextColorIdx]
	nextColorIdx++
	if nextColorIdx >= len(colors) {
		nextColorIdx = 0
	}
	nextColorMux.Unlock()

	writer := ColorWriter{
		Name:  name,
		Color: color,
	}
	return &writer
}
