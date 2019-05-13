package process

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/mobiledgex/edge-cloud/cloudcommon/influxsup"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"
)

// Local processes all run in the same global namespace, using different
// tcp ports to communicate with each other.

type TLSCerts struct {
	ServerCert string
	ServerKey  string
	ClientCert string
}

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
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	if p.ShortTimeouts {
		args = append(args, "-shortTimeouts")
	}

	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, nil, logfile)
	return err
}

func (p *Controller) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Controller) GetExeName() string { return "controller" }

func (p *Controller) LookupArgs() string { return "--apiAddr " + p.ApiAddr }

func getRestClientImpl(timeout time.Duration, addr string, tlsCertFile string) (*http.Client, error) {
	tlsConfig, err := tls.GetTLSClientConfig(addr, tlsCertFile)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}
	return client, nil
}

func connectAPIImpl(timeout time.Duration, apiaddr string, tlsCertFile string) (*grpc.ClientConn, error) {
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
	dialOption, err := tls.GetTLSClientDialOption(apiaddr, tlsCertFile)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(apiaddr, dialOption)
	return conn, err
}

func (p *Controller) ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error) {
	return connectAPIImpl(timeout, p.ApiAddr, p.TLS.ClientCert)
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
	return connectAPIImpl(timeout, p.ApiAddr, p.TLS.ClientCert)
}

func (p *Dme) GetRestClient(timeout time.Duration) (*http.Client, error) {
	return getRestClientImpl(timeout, p.HttpAddr, p.TLS.ClientCert)
}

// CrmLocal

func (p *Crm) StartLocal(logfile string, opts ...StartOp) error {
	args := []string{"--notifyAddrs", p.NotifyAddrs}
	if p.ApiAddr != "" {
		args = append(args, "--apiAddr")
		args = append(args, p.ApiAddr)
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

func (p *Crm) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Crm) GetExeName() string { return "crmserver" }

func (p *Crm) LookupArgs() string { return "--apiAddr " + p.ApiAddr }

func (p *Crm) ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error) {
	return connectAPIImpl(timeout, p.ApiAddr, p.TLS.ClientCert)
}

// InfluxLocal

func (p *Influx) StartLocal(logfile string, opts ...StartOp) error {
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.CleanStartup {
		if err := p.ResetData(); err != nil {
			return err
		}
	}

	configFile, err := influxsup.SetupInflux(p.DataDir)
	if err != nil {
		return err
	}
	p.Config = configFile
	args := []string{"-config", configFile}
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, nil, logfile)
	return err
}

func (p *Influx) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Influx) GetExeName() string { return "influxd" }

func (p *Influx) LookupArgs() string { return "" }

func (p *Influx) ResetData() error {
	return os.RemoveAll(p.DataDir)
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

func (p *ClusterSvc) StopLocal() {
	StopLocal(p.cmd)
}

func (p *ClusterSvc) GetExeName() string { return "cluster-svc" }

func (p *ClusterSvc) LookupArgs() string { return "" }

// Vault

// In dev mode, Vault is locked to below address
var VaultAddress = "http://127.0.0.1:8200"

type VaultRoles struct {
	DmeRoleID       string `json:"dmeroleid"`
	DmeSecretID     string `json:"dmesecretid"`
	MCORMRoleID     string `json:"mcormroleid"`
	MCORMSecretID   string `json:"mcormsecretid"`
	RotatorRoleID   string `json:"rotatorroleid"`
	RotatorSecretID string `json:"rotatorsecretid"`
}

func (p *Vault) StartLocal(logfile string, opts ...StartOp) error {
	// Note: for e2e tests, vault is started in dev mode.
	// In dev mode, vault is automatically unsealed, TLS is disabled,
	// data is in-memory only, and root key is printed during startup.
	// DO NOT run Vault in dev mode for production setups.
	if p.DmeSecret == "" {
		p.DmeSecret = "dme-secret"
	}
	if p.McormSecret == "" {
		p.McormSecret = "mcorm-secret"
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
		p.run("vault", "status", &serr)
		if serr == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// run setup script
	gopath := os.Getenv("GOPATH")
	setup := gopath + "/src/github.com/mobiledgex/edge-cloud/vault/setup.sh"
	out := p.run("/bin/sh", setup, &err)
	if err != nil {
		fmt.Println(out)
	}
	// get roleIDs and secretIDs
	roles := VaultRoles{}
	p.getAppRole("dme", &roles.DmeRoleID, &roles.DmeSecretID, &err)
	p.getAppRole("mcorm", &roles.MCORMRoleID, &roles.MCORMSecretID, &err)
	p.getAppRole("rotator", &roles.RotatorRoleID, &roles.RotatorSecretID, &err)
	p.putSecret("dme", p.DmeSecret+"-old", &err)
	p.putSecret("dme", p.DmeSecret, &err)
	p.putSecret("mcorm", p.McormSecret+"-old", &err)
	p.putSecret("mcorm", p.McormSecret, &err)
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

func (p *Vault) getAppRole(name string, roleID, secretID *string, err *error) {
	if *err != nil {
		return
	}
	out := p.run("vault", fmt.Sprintf("read auth/approle/role/%s/role-id", name), err)
	vals := p.mapVals(out)
	if val, ok := vals["role_id"]; ok {
		*roleID = val
	}
	out = p.run("vault", fmt.Sprintf("write -f auth/approle/role/%s/secret-id", name), err)
	vals = p.mapVals(out)
	if val, ok := vals["secret_id"]; ok {
		*secretID = val
	}
}

func (p *Vault) putSecret(name, secret string, err *error) {
	p.run("vault", fmt.Sprintf("kv put jwtkeys/%s secret=%s", name, secret), err)
}

func (p *Vault) run(bin, args string, err *error) string {
	if *err != nil {
		return ""
	}
	cmd := exec.Command(bin, strings.Split(args, " ")...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("VAULT_ADDR=%s", VaultAddress))
	out, cerr := cmd.CombinedOutput()
	if cerr != nil {
		*err = fmt.Errorf("cmd '%s %s' failed, %s", bin, args, cerr.Error())
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

// Support funcs

func StartLocal(name, bin string, args, envs []string, logfile string) (*exec.Cmd, error) {
	cmd := exec.Command(bin, args...)
	if envs != nil {
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
