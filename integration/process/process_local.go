package process

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/mobiledgex/edge-cloud/cloudcommon/influxsup"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
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
	p.cmd, err = StartLocal(p.Name, "etcd", args, logfile)
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
	p.cmd, err = StartLocal(p.Name, "controller", args, logfile)
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
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}
	var err error
	p.cmd, err = StartLocal(p.Name, "dme-server", args, logfile)
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
	p.cmd, err = StartLocal(p.Name, "crmserver", args, logfile)
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
	p.cmd, err = StartLocal(p.Name, "influxd", args, logfile)
	return err
}

func (p *InfluxLocal) Stop() {
	StopLocal(p.cmd)
}

func (p *InfluxLocal) ResetData() error {
	return os.RemoveAll(p.DataDir)
}

// Support funcs

func StartLocal(name, bin string, args []string, logfile string) (*exec.Cmd, error) {
	cmd := exec.Command(bin, args...)
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
	p.cmd, err = StartLocal(p.Name, "loc-api-sim", args, logfile)
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
	p.cmd, err = StartLocal(p.Name, "tok-srv-sim", args, logfile)
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
	p.cmd, err = StartLocal(p.Name, p.Exename, p.Args, logfile)
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
