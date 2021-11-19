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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/elastic/go-elasticsearch/v7"
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

func (s *TLSCerts) AddInternalPkiArgs(args []string) []string {
	if s.ServerCert != "" {
		args = append(args, "--itlsCert", s.ServerCert)
	}
	if s.ServerKey != "" {
		args = append(args, "--itlsKey", s.ServerKey)
	}
	if s.CACert != "" {
		args = append(args, "--itlsCA", s.CACert)
	}
	return args
}

type LocalAuth struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

var InfluxCredsFile = "/tmp/influx.json"

// EtcdLocal
var EtcdRamDiskSizeVar = "ETCD_RAMDISK_SIZEG"
var RamDisk = "ramdisk"
var MaxRamDiskSizeG = 3.0

func (p *Etcd) StartLocal(logfile string, opts ...StartOp) error {
	etcdRamDiskSizeG := os.Getenv(EtcdRamDiskSizeVar)
	if runtime.GOOS == "darwin" && etcdRamDiskSizeG != "" {
		// macos specific
		dir := "/Volumes/" + RamDisk
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			// create ram disk
			size, err := strconv.ParseFloat(etcdRamDiskSizeG, 32)
			if err != nil {
				return fmt.Errorf("Failed to convert %s value %s to float: %v", EtcdRamDiskSizeVar, etcdRamDiskSizeG, err)
			}
			// prevent the user from killing their machine
			if size > MaxRamDiskSizeG {
				return fmt.Errorf("RAM disk sizes larger than %fG not allowed to avoid killing your machine", MaxRamDiskSizeG)
			}
			// create device
			args := []string{"hdiutil", "attach", "-nomount",
				fmt.Sprintf("ram://%d", uint(size*2097152))}
			log.Printf("Creating ramdisk: %s\n", strings.Join(args, " "))
			out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
			if err != nil {
				return fmt.Errorf("Failed to create ramdisk: %s: %s, %v", strings.Join(args, " "), string(out), err)
			}
			diskID := string(out)
			eraseCmd := fmt.Sprintf("diskutil erasevolume HFS+ %s %s", RamDisk, diskID)
			log.Printf("Formatting ramdisk: %s\n", eraseCmd)
			out, err = exec.Command("bash", "-c", eraseCmd).CombinedOutput()
			if err != nil {
				return fmt.Errorf("Failed to clear ramdisk: %s: %s, %v", strings.Join(args, " "), string(out), err)
			}
		}
		base := filepath.Base(p.DataDir)
		p.DataDir = dir + "/" + base
		log.Printf("Using ramdisk for etcd %s storage: %s\n", p.Name, p.DataDir)
	}

	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.CleanStartup {
		if err := p.ResetData(); err != nil {
			return err
		}
	}

	args := []string{"--name", p.Name, "--data-dir", p.DataDir, "--listen-peer-urls", p.PeerAddrs, "--listen-client-urls", p.ClientAddrs, "--advertise-client-urls", p.ClientAddrs, "--initial-advertise-peer-urls", p.PeerAddrs, "--initial-cluster", p.InitialCluster}

	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
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

// This should be called after etcd processes are stopped
func CleanupEtcdRamDisk() error {
	if runtime.GOOS == "darwin" {
		dir := "/Volumes/" + RamDisk
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			return nil
		}
		log.Printf("Cleaning up RAM disk. Getting device ID...\n")
		args := []string{"bash", "-c", "diskutil list " + RamDisk}
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("Failed to get device ID for RAM disk: %s: %s, %v", strings.Join(args, " "), string(out), err)
		}
		outFields := strings.Fields(string(out))
		if len(outFields) < 1 {
			return fmt.Errorf("diskutil output device ID not found: %s", string(out))
		}
		diskID := outFields[0]
		log.Printf("Unmounting RAM disk %s\n", diskID)
		args = []string{"umount", "-f", diskID}
		out, err = exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("unmount etcd RAM disk failed: %s: %s, %v", strings.Join(args, " "), string(out), err)
		}
		log.Printf("Ejecting RAM disk %s\n", diskID)
		args = []string{"hdiutil", "detach", diskID}
		out, err = exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("delete etcd RAM disk failed: %s: %s, %v", strings.Join(args, " "), string(out), err)
		}
	}
	return nil
}

// ControllerLocal

func (p *Controller) StartLocal(logfile string, opts ...StartOp) error {
	args := []string{"--etcdUrls", p.EtcdAddrs, "--notifyAddr", p.NotifyAddr}
	args = append(args, p.GetNodeMgrArgs()...)
	if p.ApiAddr != "" {
		args = append(args, "--apiAddr")
		args = append(args, p.ApiAddr)
	}
	if p.HttpAddr != "" {
		args = append(args, "--httpAddr")
		args = append(args, p.HttpAddr)
	}
	if p.InfluxAddr != "" {
		args = append(args, "--influxAddr")
		args = append(args, p.InfluxAddr)
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
	if p.NotifyRootAddrs != "" {
		args = append(args, "--notifyRootAddrs")
		args = append(args, p.NotifyRootAddrs)
	}
	if p.NotifyParentAddrs != "" {
		args = append(args, "--notifyParentAddrs")
		args = append(args, p.NotifyParentAddrs)
	}
	if p.Region != "" {
		args = append(args, "--region", p.Region)
	}
	if p.EdgeTurnAddr != "" {
		args = append(args, "--edgeTurnAddr")
		args = append(args, p.EdgeTurnAddr)
	}
	if p.AppDNSRoot != "" {
		args = append(args, "--appDNSRoot")
		args = append(args, p.AppDNSRoot)
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
	if p.CheckpointInterval != "" {
		args = append(args, "--checkpointInterval")
		args = append(args, p.CheckpointInterval)
	}
	if p.ChefServerPath != "" {
		args = append(args, "--chefServerPath")
		args = append(args, p.ChefServerPath)
	}

	envs := p.GetEnv()
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
		rr := roles.GetRegionRoles(p.Region)
		envs = append(envs,
			fmt.Sprintf("VAULT_ROLE_ID=%s", rr.CtrlRoleID),
			fmt.Sprintf("VAULT_SECRET_ID=%s", rr.CtrlSecretID),
		)
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

func (p *Controller) GetTlsFile() string {
	if p.UseVaultPki && p.VaultAddr != "" {
		region := p.Region
		if region == "" {
			region = "local"
		}
		return "/tmp/edgectl." + region + "/mex.crt"
	}
	return p.TLS.ClientCert
}

func (p *Controller) ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error) {
	tlsMode := mextls.MutualAuthTLS
	skipVerify := false
	if p.TestMode {
		skipVerify = true
	}
	tlsConfig, err := mextls.GetTLSClientConfig(tlsMode, p.ApiAddr, nil, p.GetTlsFile(), "", skipVerify)
	if err != nil {
		return nil, err
	}
	return connectAPIImpl(timeout, p.ApiAddr, tlsConfig)
}

// DmeLocal

func (p *Dme) StartLocal(logfile string, opts ...StartOp) error {
	args := []string{"--notifyAddrs", p.NotifyAddrs}
	args = append(args, p.GetNodeMgrArgs()...)
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
	if p.CookieExpr != "" {
		args = append(args, "--cookieExpiration")
		args = append(args, p.CookieExpr)
	}
	if p.Region != "" {
		args = append(args, "--region", p.Region)
	}

	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	envs := p.GetEnv()
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
		rr := roles.GetRegionRoles(p.Region)
		envs = append(envs,
			fmt.Sprintf("VAULT_ROLE_ID=%s", rr.DmeRoleID),
			fmt.Sprintf("VAULT_SECRET_ID=%s", rr.DmeSecretID),
		)
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
	return connectAPIImpl(timeout, p.ApiAddr, p.getTlsConfig(p.ApiAddr))
}

func (p *Dme) GetRestClient(timeout time.Duration) (*http.Client, error) {
	return getRestClientImpl(timeout, p.HttpAddr, p.getTlsConfig(p.HttpAddr))
}

func (p *Dme) getTlsConfig(addr string) *tls.Config {
	if p.UseVaultPki && p.VaultAddr != "" {
		return &tls.Config{
			InsecureSkipVerify: true,
		}
	} else if p.TLS.ServerCert != "" && p.TLS.ServerKey != "" {
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
	} else {
		return nil
	}
}

// CrmLocal

func (p *Crm) GetArgs(opts ...StartOp) []string {
	args := []string{"--notifyAddrs", p.NotifyAddrs}
	args = append(args, p.GetNodeMgrArgs()...)
	if p.NotifySrvAddr != "" {
		args = append(args, "--notifySrvAddr")
		args = append(args, p.NotifySrvAddr)
	}
	if p.CloudletKey != "" {
		args = append(args, "--cloudletKey")
		args = append(args, p.CloudletKey)
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
	if p.Region != "" {
		args = append(args, "--region")
		args = append(args, p.Region)
	}
	if p.AppDNSRoot != "" {
		args = append(args, "--appDNSRoot")
		args = append(args, p.AppDNSRoot)
	}
	if p.ChefServerPath != "" {
		args = append(args, "--chefServerPath")
		args = append(args, p.ChefServerPath)
	}
	if p.CacheDir != "" {
		args = append(args, "--cacheDir")
		args = append(args, p.CacheDir)
	}
	args = append(args, "--HARole")
	args = append(args, string(p.HARole))

	if p.RedisAddr != "" {
		args = append(args, "--redisAddr")
		args = append(args, p.RedisAddr)
	}

	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	if p.CommercialCerts {
		args = append(args, "-commercialCerts")
	}
	return args
}

func (p *Crm) StartLocal(logfile string, opts ...StartOp) error {
	var err error

	args := p.GetArgs(opts...)
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
	return err
}

func (p *Crm) StopLocal() {
	StopLocal(p.cmd)
}

func (p *Crm) Wait() error {
	if p.cmd != nil {
		return p.cmd.Wait()
	}
	return nil
}

func (p *Crm) GetExeName() string { return "crmserver" }

func (p *Crm) LookupArgs() string {
	retval := "--cloudletKey " + p.CloudletKey
	return retval
}

func (p *Crm) LookupArgsWithHARole(HARole HARole) string {
	retval := p.LookupArgs() + ".*--HARole " + string(HARole)
	return retval
}

func (p *Crm) String(opts ...StartOp) string {
	cmd_str := p.GetExeName()
	args := p.GetArgs(opts...)
	for _, v := range args {
		if strings.HasPrefix(v, "-") {
			cmd_str += " " + v
		} else {
			cmd_str += " '" + v + "'"
		}
	}
	return cmd_str
}

// InfluxLocal

func (p *Influx) StartLocal(logfile string, opts ...StartOp) (reterr error) {
	var prefix string
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.CleanStartup {
		if err := p.ResetData(); err != nil {
			return err
		}
	}

	// check if another influx already running
	if p.HttpAddr == "" {
		p.HttpAddr = influxsup.DefaultHttpAddr
	}
	if p.BindAddr == "" {
		p.BindAddr = influxsup.DefaultBindAddr
	}
	conn, err := net.DialTimeout("tcp", p.HttpAddr, 100*time.Millisecond)
	if err == nil && conn != nil {
		conn.Close()
		return fmt.Errorf("InfluxDB http addr %s already in use", p.HttpAddr)
	}
	conn, err = net.DialTimeout("tcp", p.BindAddr, 100*time.Millisecond)
	if err == nil && conn != nil {
		conn.Close()
		return fmt.Errorf("InfluxDB bind addr %s already in use", p.BindAddr)
	}

	influxops := []influxsup.InfluxOp{
		influxsup.WithAuth(p.Auth.User != ""),
	}
	if p.TLS.ServerCert != "" {
		influxops = append(influxops, influxsup.WithServerCert(p.TLS.ServerCert))
	}
	if p.TLS.ServerKey != "" {
		influxops = append(influxops, influxsup.WithServerCertKey(p.TLS.ServerKey))
	}
	if p.BindAddr != "" {
		influxops = append(influxops, influxsup.WithBindAddr(p.BindAddr))
	}
	if p.HttpAddr != "" {
		influxops = append(influxops, influxsup.WithHttpAddr(p.HttpAddr))
	}

	configFile, err := influxsup.SetupInflux(p.DataDir, influxops...)
	if err != nil {
		return err
	}
	p.Config = configFile
	args := []string{"-config", configFile}
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
	if err != nil {
		return err
	}
	defer func() {
		if reterr != nil {
			p.StopLocal()
		}
	}()

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
			return err
		}
	}
	// create auth file for Vault
	creds_json, err := json.Marshal(p.Auth)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(InfluxCredsFile, creds_json, 0644)
	if err != nil {
		return err
	}
	// make sure influx is online
	if prefix == "" {
		prefix = "http://" + p.HttpAddr
	}
	client, err := influxsup.GetClient(prefix, p.Auth.User, p.Auth.Pass)
	if err != nil {
		return err
	}
	online := false
	for ii := 0; ii < 50; ii++ {
		if _, _, err := client.Ping(0); err == nil {
			online = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !online {
		return fmt.Errorf("InfluxDB service not online")
	}
	return nil
}

func (p *Influx) StopLocal() {
	StopLocal(p.cmd)
	// make sure influx isn't running anymore
	log.Printf("Check that InfluxDB at %s is not running\n", p.HttpAddr)
	done := false
	for ii := 0; ii < 50; ii++ {
		conn, err := net.DialTimeout("tcp", p.HttpAddr, 100*time.Millisecond)
		if err == nil && conn != nil {
			conn.Close()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		done = true
		break
	}
	log.Printf("InfluxDB at %s is done: %t\n", p.HttpAddr, done)
}

func (p *Influx) GetExeName() string { return "influxd" }

func (p *Influx) LookupArgs() string { return "-config " + p.Config }

func (p *Influx) ResetData() error {
	return os.RemoveAll(p.DataDir)
}

func (p *Influx) GetClient() (influxclient.Client, error) {
	httpaddr := ""
	if p.TLS.ServerCert != "" {
		httpaddr = "https://" + p.HttpAddr
	} else {
		httpaddr = "http://" + p.HttpAddr
	}
	return influxsup.GetClient(httpaddr, p.Auth.User, p.Auth.Pass)
}

// ClusterSvc process

func (p *ClusterSvc) StartLocal(logfile string, opts ...StartOp) error {
	args := []string{"--notifyAddrs", p.NotifyAddrs}
	args = append(args, p.GetNodeMgrArgs()...)
	if p.CtrlAddrs != "" {
		args = append(args, "--ctrlAddrs")
		args = append(args, p.CtrlAddrs)
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
	if p.Region != "" {
		args = append(args, "--region", p.Region)
	}
	args = append(args, "--hostname", p.Name)
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	envs := p.GetEnv()
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
		rr := roles.GetRegionRoles(p.Region)
		envs = append(envs,
			fmt.Sprintf("VAULT_ROLE_ID=%s", rr.ClusterSvcRoleID),
			fmt.Sprintf("VAULT_SECRET_ID=%s", rr.ClusterSvcSecretID),
		)
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
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, envs, logfile)
	return err
}

func (p *ClusterSvc) StopLocal() {
	StopLocal(p.cmd)
}

func (p *ClusterSvc) GetExeName() string { return "cluster-svc" }

func (p *ClusterSvc) LookupArgs() string { return p.Name }

// Vault

// In dev mode, Vault is locked to below address
var defaultVaultAddress = "http://127.0.0.1:8200"

type VaultRoles struct {
	NotifyRootRoleID   string `json:"notifyrootroleid"`
	NotifyRootSecretID string `json:"notifyrootsecretid"`
	RegionRoles        map[string]*VaultRegionRoles
}

type VaultRegionRoles struct {
	DmeRoleID          string `json:"dmeroleid"`
	DmeSecretID        string `json:"dmesecretid"`
	RotatorRoleID      string `json:"rotatorroleid"`
	RotatorSecretID    string `json:"rotatorsecretid"`
	CtrlRoleID         string `json:"controllerroleid"`
	CtrlSecretID       string `json:"controllersecretid"`
	ClusterSvcRoleID   string `json:"clustersvcroleid"`
	ClusterSvcSecretID string `json:"clustersvcsecretid"`
	EdgeTurnRoleID     string `json:"edgeturnroleid"`
	EdgeTurnSecretID   string `json:"edgeturnsecretid"`
}

func (s *VaultRoles) GetRegionRoles(region string) *VaultRegionRoles {
	if region == "" {
		region = "local"
	}
	return s.RegionRoles[region]
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
	if p.ListenAddr == "" {
		p.ListenAddr = defaultVaultAddress
	}
	if !strings.HasPrefix(p.ListenAddr, "http://") {
		return fmt.Errorf("vault listen addr must start with http://")
	}
	// unfortunately arg passed to vault cannot have http
	addr := strings.TrimPrefix(p.ListenAddr, "http://")
	args = append(args, "-dev-listen-address="+addr)

	// Specify the root token. Vault generates one automatically
	// and stores it in ~/.vault-token for the CLI to use, but
	// somehow with unit tests running multiple Vaults, they mess
	// up the ~/.vault-token for each other (even though they're
	// not supposed to be running at the same time).
	p.RootToken = "vault-token"
	args = append(args, "-dev-root-token-id="+p.RootToken)
	if p.CADir == "" {
		p.CADir = "/tmp/vault_pki"
	}
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
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
	if p.cmd.Process == nil {
		return fmt.Errorf("failed to start vault process, see log %s", logfile)
	}
	options := StartOptions{}
	options.ApplyStartOptions(opts...)

	// run setup script
	gopath := os.Getenv("GOPATH")
	setup := gopath + "/src/github.com/mobiledgex/edge-cloud/vault/setup.sh"
	out := p.Run("/bin/sh", setup, &err)
	fmt.Println(out)
	// get roleIDs and secretIDs
	vroles := VaultRoles{}
	vroles.RegionRoles = make(map[string]*VaultRegionRoles)
	p.GetAppRole("", "notifyroot", &vroles.NotifyRootRoleID, &vroles.NotifyRootSecretID, &err)

	if p.Regions == "" {
		p.Regions = "local"
	}
	for _, region := range strings.Split(p.Regions, ",") {
		// run setup script
		setup := gopath + "/src/github.com/mobiledgex/edge-cloud/vault/setup-region.sh " + region
		out := p.Run("/bin/sh", setup, &err)
		if err != nil {
			fmt.Println(out)
		}
		// get roleIDs and secretIDs
		roles := VaultRegionRoles{}
		p.GetAppRole(region, "dme", &roles.DmeRoleID, &roles.DmeSecretID, &err)
		p.GetAppRole(region, "rotator", &roles.RotatorRoleID, &roles.RotatorSecretID, &err)
		p.GetAppRole(region, "controller", &roles.CtrlRoleID, &roles.CtrlSecretID, &err)
		p.GetAppRole(region, "cluster-svc", &roles.ClusterSvcRoleID, &roles.ClusterSvcSecretID, &err)
		p.GetAppRole(region, "edgeturn", &roles.EdgeTurnRoleID, &roles.EdgeTurnSecretID, &err)
		p.PutSecret(region, "dme", p.DmeSecret+"-old", &err)
		p.PutSecret(region, "dme", p.DmeSecret, &err)
		vroles.RegionRoles[region] = &roles
		// Get the directory where the influx.json file is
		if _, serr := os.Stat(InfluxCredsFile); !os.IsNotExist(serr) {
			path := "secret/" + region + "/accounts/influxdb"
			p.PutSecretsJson(path, InfluxCredsFile, &err)
		}
		if err != nil {
			p.StopLocal()
			return err
		}
	}
	if options.RolesFile != "" {
		roleYaml, err := yaml.Marshal(&vroles)
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
	for _, vaultData := range p.VaultDatas {
		data, err := json.Marshal(vaultData.Data)
		if err != nil {
			log.Printf("Failed to marshal vault data - %v[err:%v]\n", vaultData, err)
			continue
		}
		// get a reader for the data
		reader := strings.NewReader(string(data))
		p.RunWithInput("vault", fmt.Sprintf("kv put %s -", vaultData.Path), reader, &err)
		if err != nil {
			log.Printf("Failed to store secret in [%s] - err:%v\n", vaultData.Path, err)
			continue
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
	return p.RunWithInput(bin, args, nil, err)
}

func (p *Vault) RunWithInput(bin, args string, input io.Reader, err *error) string {
	if *err != nil {
		return ""
	}
	cmd := exec.Command(bin, strings.Split(args, " ")...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("VAULT_ADDR=%s", p.ListenAddr),
		fmt.Sprintf("VAULT_TOKEN=%s", p.RootToken),
		fmt.Sprintf("CADIR=%s", p.CADir))
	if input != nil {
		cmd.Stdin = input
	}
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
	if p.CADir == "" {
		p.CADir = dir + "/vault_pki"
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
	if p.DockerNetwork != "" {
		args = append(args, "--network", p.DockerNetwork)
	}
	args = append(args, "traefik:v2.0")

	staticArgs := TraefikStaticArgs{}

	// Traefik consists of a Static Config file, and zero or more
	// dynamic config files. Dynamic config files can be hot-reloaded.
	// The allowed contents of each type are different.
	// Entry points are configured statically, while routers, services,
	// etc are configured dynmically, either through a file provider
	// or docker provider (snooping on docker events).

	err := writeAllCAs(p.TLS.CACert, configDir+"/traefikCAs.pem")
	if err != nil {
		return err
	}
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

	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
	return err
}

func (p *Traefik) StopLocal() {
	StopLocal(p.cmd)
	// if container is from previous aborted run
	cmd := exec.Command("docker", "kill", p.Name)
	cmd.Run()
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
        - traefikCAs.pem
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
	args := p.GetRunArgs()
	for _, l := range labels {
		args = append(args, "-l", l)
	}
	// jaeger version should match "jaeger_version" in
	// ansible/roles/jaeger/defaults/main.yaml
	args = append(args, "jaegertracing/all-in-one:1.17.1",
		"--collector.num-workers=500",
		"--collector.queue-size=10000",
	)
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
	return err
}

func (p *Jaeger) StartLocalNoTraefik(logfile string, opts ...StartOp) error {
	args := p.GetRunArgs()
	// jaeger version should match "jaeger_version" in
	// ansible/roles/jaeger/defaults/main.yaml
	args = append(args,
		"-p", "16686:16686",
		"-p", "14268:14268",
		"jaegertracing/all-in-one:1.17.1")
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
	if err == nil {
		// wait until up
		url := "http://127.0.0.1:16686/"
		var resp *http.Response
		for ii := 0; ii < 30; ii++ {
			client := http.Client{
				Timeout: time.Second,
			}
			resp, err = client.Get(url)
			if err != nil {
				time.Sleep(time.Second)
				log.Printf("jeager %s try %d: err %v\n", url, ii, err)
				continue
			}
			log.Printf("jeager %s try %d: response %d\n", url, ii, resp.StatusCode)
			if resp.StatusCode != http.StatusOK {
				time.Sleep(time.Second)
				continue
			}
			break
		}
	}
	return err
}

func (p *RedisCache) StartLocal(logfile string, opts ...StartOp) error {
	args := p.GetRunArgs()
	redisPort := LocalRedisPort
	args = append(args,
		"-p", redisPort+":"+redisPort,
		"redis",
	)
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
	if err != nil {
		return err
	}
	// wait for redis to become ready
	maxRedisWait := 10 * time.Second
	start := time.Now()
	for {
		conn, err := net.Dial("tcp", LocalRedisAddr)
		if err == nil {
			conn.Close()
			break
		}
		elapsed := time.Since(start)
		if elapsed > maxRedisWait {
			return fmt.Errorf("Timed out try to connect to redis")
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (d *DockerNetwork) Create() error {
	return d.run("create")
}

func (d *DockerNetwork) Delete() error {
	err := d.run("rm")
	if err != nil && strings.Contains(err.Error(), "No such network") {
		err = nil
	}
	return err
}

func (d *DockerNetwork) run(action string) error {
	args := []string{"docker", "network", action, d.Name}
	log.Printf("Running: %s\n", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s, %s", string(out), err)
	}
	log.Printf("%s", string(out))
	return nil
}

func (p *DockerGeneric) GetRunArgs() []string {
	args := []string{
		"run", "--rm", "--name", p.Name,
	}
	if runtime.GOOS != "darwin" {
		// For Linux, "host.docker.internal" host name doesn't work from inside docker container
		// Use "--add-host" to add this mapping, only works if Docker version >= 20.04
		args = append(args, "--add-host", "host.docker.internal:host-gateway")
	}
	for k, v := range p.DockerEnvVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	for _, link := range p.Links {
		args = append(args, "--link", link)
	}
	if p.DockerNetwork != "" {
		args = append(args, "--network", p.DockerNetwork)
	}
	return args
}

func (p *DockerGeneric) StopLocal() {
	StopLocal(p.cmd)
	// if container is from previous aborted run
	cmd := exec.Command("docker", "kill", p.Name)
	cmd.Run()
}

func (p *DockerGeneric) GetExeName() string { return "docker" }

func (p *DockerGeneric) LookupArgs() string { return p.Name }

func (p *DockerGeneric) SetCmd(cmd *exec.Cmd) { p.cmd = cmd }

func (p *DockerGeneric) GetCmd() *exec.Cmd { return p.cmd }

func (p *ElasticSearch) StartLocal(logfile string, opts ...StartOp) error {
	switch p.Type {
	case "kibana":
		return p.StartKibana(logfile, opts...)
	default:
		return p.StartElasticSearch(logfile, opts...)
	}
}

func (p *ElasticSearch) StartElasticSearch(logfile string, opts ...StartOp) error {
	// simple single node cluster
	args := p.GetRunArgs()
	args = append(args,
		"-p", "9200:9200",
		"-p", "9300:9300",
		"-e", "discovery.type=single-node",
		"-e", "xpack.security.enabled=false",
		"docker.elastic.co/elasticsearch/elasticsearch:7.6.2",
	)
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
	if err == nil {
		// wait until up
		addr := "http://127.0.0.1:9200"
		cfg := elasticsearch.Config{
			Addresses: []string{addr},
		}
		client, perr := elasticsearch.NewClient(cfg)
		if perr != nil {
			return perr
		}
		for ii := 0; ii < 30; ii++ {
			res, perr := client.Info()
			log.Printf("elasticsearch info %s try %d: res %v, perr %v\n", addr, ii, res, perr)
			if perr == nil {
				res.Body.Close()
			}
			if perr == nil && res.StatusCode == http.StatusOK {
				break
			}
			time.Sleep(2 * time.Second)
		}
		if perr != nil {
			return perr
		}
	}
	return err
}

func (p *ElasticSearch) StartKibana(logfile string, opts ...StartOp) error {
	args := p.GetRunArgs()
	args = append(args,
		"-p", "5601:5601",
		"docker.elastic.co/kibana/kibana:7.6.2",
	)
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
	return err
}

func writeAllCAs(inputCAFile, outputCAFile string) error {
	// Combine all CAs into one for nginx or other TLS-terminating proxies.
	// Note that nginx requires the full CA chain, so must include
	// the root's public CA cert as well (not just intermediates).
	certs := "/tmp/vault_pki/*.pem"
	if inputCAFile != "" {
		certs += " " + inputCAFile
	}
	cmd := exec.Command("bash", "-c", "cat "+certs+" > "+outputCAFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s, %s", string(out), err)
	}
	return nil
}

func (p *NginxProxy) StartLocal(logfile string, opts ...StartOp) error {
	configDir := path.Dir(logfile) + "/" + p.Name
	if err := os.MkdirAll(configDir, 0777); err != nil {
		return err
	}

	// make a copy of process to remap certs files
	pArgs := *p

	// Terminate TLS using mex-ca.crt and vault CAs.
	if p.TLS.ServerCert != "" {
		if p.TLS.ServerKey == "" {
			err := fmt.Errorf("NginxProxy with ServerCert requires ServerKey")
			log.Printf("%v\n", err)
			return err
		}
		err := writeAllCAs("", configDir+"/allcas.pem")
		if err != nil {
			return err
		}
		pArgs.TLS.ServerCert = path.Base(p.TLS.ServerCert)
		pArgs.TLS.ServerKey = path.Base(p.TLS.ServerKey)
	}

	tmpl := template.Must(template.New("nginxProxy").Parse(nginxProxyConfig))
	f, err := os.Create(configDir + "/nginx.conf")
	if err != nil {
		return err
	}
	defer f.Close()

	wr := bufio.NewWriter(f)
	err = tmpl.Execute(wr, &pArgs)
	if err != nil {
		return err
	}
	wr.Flush()

	args := p.GetRunArgs()
	for _, server := range p.Servers {
		if server.Port != "" {
			args = append(args, "-p", server.Port+":"+server.Port)
		}
		if server.TlsPort != "" {
			args = append(args, "-p", server.TlsPort+":"+server.TlsPort)
		}
	}
	if p.TLS.ServerCert != "" {
		certsDir := path.Dir(p.TLS.ServerCert)
		args = append(args, "-v", fmt.Sprintf("%s:/certs", certsDir))
	}
	args = append(args,
		"-v", fmt.Sprintf("%s:/etc/nginx", configDir),
		"nginx:latest",
	)
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
	return err
}

var nginxProxyConfig = `
events {
  worker_connections 128;
}
http {
  tcp_nopush on;
  tcp_nodelay on;
  default_type application/octet-stream;

  access_log /etc/nginx/access.log;
  error_log /etc/nginx/error.log;

{{- range .Servers}}
  server {
{{- if .TlsPort}}
    listen {{.TlsPort}} ssl;
{{- end}}
{{- if .Port}}
    listen {{.Port}};
{{- end}}
{{- if $.TLS.ServerCert}}

    ssl_certificate /certs/{{$.TLS.ServerCert}};
    ssl_certificate_key /certs/{{$.TLS.ServerKey}};
    ssl_client_certificate /etc/nginx/allcas.pem;
    ssl_verify_client on;
    ssl_verify_depth 2;
    ssl_session_cache shared:le_nginx_SSL:1m;
    ssl_session_cache shared:le_nginx_SSL:1m;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers "ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA:ECDHE-RSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-RSA-AES256-SHA256:DHE-RSA-AES256-SHA:ECDHE-ECDSA-DES-CBC3-SHA:ECDHE-RSA-DES-CBC3-SHA:EDH-RSA-DES-CBC3-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:DES-CBC3-SHA:!DSS";
{{- end}}

    server_name {{$.Name}} localhost;

    proxy_buffering off;

    location / {
      proxy_pass         {{.Target}};
      proxy_set_header   Host $host;
      proxy_set_header   X-Real-IP $remote_addr;
      proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
    }
  }
{{- end}}
}
`

func (p *NotifyRoot) StartLocal(logfile string, opts ...StartOp) error {
	args := p.GetNodeMgrArgs()

	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	envs := p.GetEnv()
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
		envs = append(envs,
			fmt.Sprintf("VAULT_ROLE_ID=%s", roles.NotifyRootRoleID),
			fmt.Sprintf("VAULT_SECRET_ID=%s", roles.NotifyRootSecretID),
		)
	}

	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, envs, logfile)
	return err
}

func (p *NotifyRoot) StopLocal() {
	StopLocal(p.cmd)
}

func (p *NotifyRoot) GetExeName() string { return "notifyroot" }

func (p *NotifyRoot) LookupArgs() string { return "" }

func (p *EdgeTurn) StartLocal(logfile string, opts ...StartOp) error {
	args := p.GetNodeMgrArgs()
	if p.ListenAddr != "" {
		args = append(args, "--listenAddr")
		args = append(args, p.ListenAddr)
	}
	if p.ProxyAddr != "" {
		args = append(args, "--proxyAddr")
		args = append(args, p.ProxyAddr)
	}
	if p.Region != "" {
		args = append(args, "--region", p.Region)
	}
	if p.TestMode {
		args = append(args, "--testMode")
	}
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	envs := p.GetEnv()
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
		rr := roles.GetRegionRoles(p.Region)
		envs = append(envs,
			fmt.Sprintf("VAULT_ROLE_ID=%s", rr.EdgeTurnRoleID),
			fmt.Sprintf("VAULT_SECRET_ID=%s", rr.EdgeTurnSecretID),
		)
	}
	var err error
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, envs, logfile)
	return err
}

func (p *EdgeTurn) StopLocal() {
	StopLocal(p.cmd)
}

func (p *EdgeTurn) GetExeName() string { return "edgeturn" }

func (p *EdgeTurn) LookupArgs() string { return "" }

// Support funcs

func StartLocal(name, bin string, args, envs []string, logfile string) (*exec.Cmd, error) {
	log.Printf("StartLocal:\n%s %s\n", bin, strings.Join(args, " "))
	cmd := exec.Command(bin, args...)
	if len(envs) > 0 {
		log.Printf("%s env: %v\n", name, envs)
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
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
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
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), args, p.GetEnv(), logfile)
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
	p.cmd, err = StartLocal(p.Name, p.GetExeName(), p.Args, p.GetEnv(), logfile)
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
