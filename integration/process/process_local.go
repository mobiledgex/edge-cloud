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
	ClientCert string
}

// EtcdLocal
type EtcdLocal struct {
	Name           string
	DataDir        string
	PeerAddrs      string
	ClientAddrs    string
	InitialCluster string
	cmd            *exec.Cmd
}

func (p *EtcdLocal) Start(logfile string) error {
	args := []string{"--name", p.Name, "--data-dir", p.DataDir, "--listen-peer-urls", p.PeerAddrs, "--listen-client-urls", p.ClientAddrs, "--advertise-client-urls", p.ClientAddrs, "--initial-advertise-peer-urls", p.PeerAddrs, "--initial-cluster", p.InitialCluster}
	var err error
	p.cmd, err = StartLocal(p.Name, "etcd", args, nil, logfile)
	return err
}

func (p *EtcdLocal) Stop() {
	StopLocal(p.cmd)
}

func (p *EtcdLocal) ResetData() error {
	return os.RemoveAll(p.DataDir)
}

// ControllerLocal

type ControllerLocal struct {
	Name          string
	EtcdAddrs     string
	ApiAddr       string
	HttpAddr      string
	NotifyAddr    string
	TLS           TLSCerts
	ShortTimeouts bool
	cmd           *exec.Cmd
}

func (p *ControllerLocal) Start(logfile string, opts ...StartOp) error {
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
	p.cmd, err = StartLocal(p.Name, "controller", args, nil, logfile)
	return err
}

func (p *ControllerLocal) Stop() {
	StopLocal(p.cmd)
}

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

func (p *ControllerLocal) ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error) {
	return connectAPIImpl(timeout, p.ApiAddr, p.TLS.ClientCert)
}

// DmeLocal
type DmeLocal struct {
	Name        string
	ApiAddr     string
	HttpAddr    string
	NotifyAddrs string
	LocVerUrl   string
	TokSrvUrl   string
	Carrier     string
	CloudletKey string
	VaultAddr   string
	TLS         TLSCerts
	cmd         *exec.Cmd
}

func (p *DmeLocal) Start(logfile string, opts ...StartOp) error {
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
	p.cmd, err = StartLocal(p.Name, "dme-server", args, envs, logfile)
	return err
}

func (p *DmeLocal) Stop() {
	StopLocal(p.cmd)
}

func (p *DmeLocal) ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error) {
	return connectAPIImpl(timeout, p.ApiAddr, p.TLS.ClientCert)
}

func (p *DmeLocal) GetRestClient(timeout time.Duration) (*http.Client, error) {
	return getRestClientImpl(timeout, p.HttpAddr, p.TLS.ClientCert)
}

// CrmLocal

type CrmLocal struct {
	Name        string
	ApiAddr     string
	NotifyAddrs string
	CloudletKey string
	TLS         TLSCerts
	cmd         *exec.Cmd
}

func (p *CrmLocal) Start(logfile string, opts ...StartOp) error {
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
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}

	var err error
	p.cmd, err = StartLocal(p.Name, "crmserver", args, nil, logfile)
	return err
}

func (p *CrmLocal) Stop() {
	StopLocal(p.cmd)
}

func (p *CrmLocal) ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error) {
	return connectAPIImpl(timeout, p.ApiAddr, p.TLS.ClientCert)
}

// InfluxLocal

type InfluxLocal struct {
	Name     string
	DataDir  string
	HttpAddr string
	Config   string // set during Start
	cmd      *exec.Cmd
}

func (p *InfluxLocal) Start(logfile string) error {
	configFile, err := influxsup.SetupInflux(p.DataDir)
	if err != nil {
		return err
	}
	p.Config = configFile
	args := []string{"-config", configFile}
	p.cmd, err = StartLocal(p.Name, "influxd", args, nil, logfile)
	return err
}

func (p *InfluxLocal) Stop() {
	StopLocal(p.cmd)
}

func (p *InfluxLocal) ResetData() error {
	return os.RemoveAll(p.DataDir)
}

// Postgres Sql
type SqlLocal struct {
	Name     string
	DataDir  string
	HttpAddr string
	Username string
	Dbname   string
	TLS      TLSCerts
	cmd      *exec.Cmd
}

func (p *SqlLocal) Start(logfile string) error {
	args := []string{"-D", p.DataDir, "start"}
	options := []string{}
	addr := []string{}
	if p.HttpAddr != "" {
		addr = strings.Split(p.HttpAddr, ":")
		if len(addr) == 2 {
			options = append(options, "-p")
			options = append(options, addr[1])
		}
	}
	if p.TLS.ServerCert != "" {
		// files server.crt and server.key must exist
		// in server's data directory (note: this is untested).
		options = append(options, "-l")
	}
	if len(options) > 0 {
		args = append(args, "-o")
		args = append(args, strings.Join(options, " "))
	}
	var err error
	p.cmd, err = StartLocal(p.Name, "pg_ctl", args, nil, logfile)
	if err != nil {
		return err
	}
	// wait until pg_ctl script exits (means postgres service is ready)
	state, err := p.cmd.Process.Wait()
	if err != nil {
		return fmt.Errorf("failed wait for pg_ctl, %s", err.Error())
	}
	if !state.Exited() {
		return fmt.Errorf("pg_ctl not exited")
	}
	if !state.Success() {
		return fmt.Errorf("pg_ctl failed, see script output")
	}

	// create primary user
	out, err := p.runPsql([]string{"-c", "select rolname from pg_roles",
		"postgres"})
	if err != nil {
		p.Stop()
		return fmt.Errorf("sql: failed to list postgres roles, %s", err.Error())
	}
	if !strings.Contains(string(out), p.Username) {
		out, err = p.runPsql([]string{"-c",
			fmt.Sprintf("create user %s", p.Username), "postgres"})
		fmt.Println(string(out))
		if err != nil {
			p.Stop()
			return fmt.Errorf("sql: failed to create user %s, %s",
				p.Username, err.Error())
		}
	}

	// create user database
	out, err = p.runPsql([]string{"-c", "select datname from pg_database",
		"postgres"})
	if err != nil {
		p.Stop()
		return fmt.Errorf("sql: failed to list databases, %s", err.Error())
	}
	if !strings.Contains(string(out), p.Dbname) {
		out, err = p.runPsql([]string{"-c",
			fmt.Sprintf("create database %s", p.Dbname), "postgres"})
		fmt.Println(string(out))
		if err != nil {
			p.Stop()
			return fmt.Errorf("sql: failed to create user %s, %s",
				p.Username, err.Error())
		}
	}
	return nil
}
func (p *SqlLocal) Stop() {
	exec.Command("pg_ctl", "-D", p.DataDir, "stop").CombinedOutput()
}
func (p *SqlLocal) InitDataDir() error {
	err := os.RemoveAll(p.DataDir)
	if err != nil {
		return err
	}
	_, err = exec.Command("initdb", p.DataDir).CombinedOutput()
	return err
}
func (p *SqlLocal) runPsql(args []string) ([]byte, error) {
	if p.HttpAddr != "" {
		addr := strings.Split(p.HttpAddr, ":")
		if len(addr) == 2 {
			args = append([]string{"-h", addr[0], "-p", addr[1]}, args...)
		}
	}
	return exec.Command("psql", args...).CombinedOutput()
}

// Vault
type Vault struct {
	Name        string
	DmeSecret   string
	McormSecret string
	cmd         *exec.Cmd
}

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

func (p *Vault) Start(logfile string, opts ...StartOp) error {
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
	p.cmd, err = StartLocal(p.Name, "vault", args, nil, logfile)
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
		p.Stop()
		return err
	}
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.RolesFile != "" {
		roleYaml, err := yaml.Marshal(&roles)
		if err != nil {
			p.Stop()
			return err
		}
		err = ioutil.WriteFile(options.RolesFile, roleYaml, 0644)
		if err != nil {
			p.Stop()
			return err
		}
	}
	return err
}

func (p *Vault) Stop() {
	StopLocal(p.cmd)
}

func (p *Vault) getAppRole(name string, roleID, secretID *string, err *error) {
	if *err != nil {
		return
	}
	out := p.run("vault", fmt.Sprintf("read auth/approle/role/%s/role-id", name), err)
	log.Printf("VAULT getAppRole %s", fmt.Sprintf("vault read auth/approle/role/%s/role-id", name))

	vals := p.mapVals(out)
	if val, ok := vals["role_id"]; ok {
		*roleID = val
	}
	out = p.run("vault", fmt.Sprintf("write -f auth/approle/role/%s/secret-id", name), err)
	log.Printf("VAULT writeAppRole:  %s", fmt.Sprintf("vault write -f auth/approle/role/%s/secret-id", name))

	vals = p.mapVals(out)
	if val, ok := vals["secret_id"]; ok {
		*secretID = val
		log.Printf("VAULT writeAppRole found secret:  %s", val)

	}
}

func (p *Vault) putSecret(name, secret string, err *error) {
	log.Printf("VAULT putSecret %s", fmt.Sprintf("vault kv put jwtkeys/%s secret=%s", name, secret))
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
	}
}

//Location API simulator
type LocApiSimLocal struct {
	Name    string
	Port    int
	Locfile string
	Geofile string
	Country string
	cmd     *exec.Cmd
}

func (p *LocApiSimLocal) Start(logfile string) error {
	args := []string{"-port", fmt.Sprintf("%d", p.Port), "-file", p.Locfile}
	if p.Geofile != "" {
		args = append(args, "-geo", p.Geofile)
	}
	if p.Country != "" {
		args = append(args, "-country", p.Country)
	}
	var err error
	p.cmd, err = StartLocal(p.Name, "loc-api-sim", args, nil, logfile)
	return err
}

func (p *LocApiSimLocal) Stop() {
	StopLocal(p.cmd)
}

//Token service simulator
type TokSrvSimLocal struct {
	Name  string
	Port  int
	Token string
	cmd   *exec.Cmd
}

func (p *TokSrvSimLocal) Start(logfile string) error {
	args := []string{"-port", fmt.Sprintf("%d", p.Port)}
	if p.Token != "" {
		args = append(args, "-token")
		args = append(args, p.Token)
	}
	var err error
	p.cmd, err = StartLocal(p.Name, "tok-srv-sim", args, nil, logfile)
	return err
}

func (p *TokSrvSimLocal) Stop() {
	StopLocal(p.cmd)
}

//Generic sample app for use in test
type SampleAppLocal struct {
	Name    string
	Exename string
	Args    []string
	cmd     *exec.Cmd
}

func (p *SampleAppLocal) Start(logfile string) error {
	var err error
	p.cmd, err = StartLocal(p.Name, p.Exename, p.Args, nil, logfile)
	return err
}

func (p *SampleAppLocal) Stop() {
	StopLocal(p.cmd)
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
