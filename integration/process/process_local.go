package process

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	"google.golang.org/grpc"
)

// Local processes all run in the same global namespace, using different
// tcp ports to communicate with each other.

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
	Name       string
	EtcdAddrs  string
	ApiAddr    string
	HttpAddr   string
	NotifyAddr string
	cmd        *exec.Cmd
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
	options := StartOptions{}
	options.ApplyStartOptions(opts...)
	if options.Debug != "" {
		args = append(args, "-d")
		args = append(args, options.Debug)
	}

	var err error
	p.cmd, err = StartLocal(p.Name, "controller", args, logfile)
	return err
}

func (p *ControllerLocal) Stop() {
	StopLocal(p.cmd)
}

func connectAPIImpl(timeout time.Duration, apiaddr string) (*grpc.ClientConn, error) {
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
			log.Printf("Timeout in connection to %v\n", apiaddr)
			return nil, err
		}
		if err == nil {
			break
		}
		timeout -= wait
		time.Sleep(wait)
	}

	conn, err := grpc.Dial(apiaddr, grpc.WithInsecure())
	return conn, err
}

func (p *ControllerLocal) ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error) {
	return connectAPIImpl(timeout, p.ApiAddr)
}

// DmeLocal

type DmeLocal struct {
	Name        string
	ApiAddr     string
	NotifyAddrs string
	LocVerUrl   string
	Carrier     string
	cmd         *exec.Cmd
}

func (p *DmeLocal) Start(logfile string, opts ...StartOp) error {
	args := []string{"--notifyAddrs", p.NotifyAddrs}
	if p.ApiAddr != "" {
		args = append(args, "--apiAddr")
		args = append(args, p.ApiAddr)
	}
	if p.LocVerUrl != "" {
		args = append(args, "--locverurl")
		args = append(args, p.LocVerUrl)
	}
	if p.Carrier != "" {
		args = append(args, "--carrier")
		args = append(args, p.Carrier)
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
	return connectAPIImpl(timeout, p.ApiAddr)
}

// CrmLocal

type CrmLocal struct {
	Name        string
	ApiAddr     string
	NotifyAddrs string
	cmd         *exec.Cmd
}

func (p *CrmLocal) Start(logfile string) error {
	args := []string{"--notifyAddrs", p.NotifyAddrs}
	if p.ApiAddr != "" {
		args = append(args, "--apiAddr")
		args = append(args, p.ApiAddr)
	}
	var err error
	p.cmd, err = StartLocal(p.Name, "crmserver", args, logfile)
	return err
}

func (p *CrmLocal) Stop() {
	StopLocal(p.cmd)
}

func (p *CrmLocal) ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error) {
	return connectAPIImpl(timeout, p.ApiAddr)
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
		outfile, err := os.Create(logfile)
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
	cmd     *exec.Cmd
}

func (p *LocApiSimLocal) Start(logfile string) error {
	args := []string{"-port", fmt.Sprintf("%d", p.Port), "-file", p.Locfile}
	var err error
	p.cmd, err = StartLocal(p.Name, "loc-api-sim", args, logfile)
	return err
}

func (p *LocApiSimLocal) Stop() {
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
