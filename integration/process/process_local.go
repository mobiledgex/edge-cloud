package process

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

func (p *EtcdLocal) Start() error {
	args := []string{"--name", p.Name, "--data-dir", p.DataDir, "--listen-peer-urls", p.PeerAddrs, "--listen-client-urls", p.ClientAddrs, "--advertise-client-urls", p.ClientAddrs, "--initial-advertise-peer-urls", p.PeerAddrs, "--initial-cluster", p.InitialCluster}
	var err error
	p.cmd, err = StartLocal(p.Name, "etcd", args)
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

func (p *ControllerLocal) Start(opts ...StartOp) error {
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
	p.cmd, err = StartLocal(p.Name, "controller", args)
	return err
}

func (p *ControllerLocal) Stop() {
	StopLocal(p.cmd)
}

func (p *ControllerLocal) ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error) {
	// Wait for controller to be ready to connect.
	// Note: using grpc WithBlock() takes about a second longer
	// than doing the retry connect below so requires a larger timeout.
	wait := 20 * time.Millisecond
	for {
		_, err := net.Dial("tcp", p.ApiAddr)
		if err == nil || timeout < wait {
			break
		}
		timeout -= wait
		time.Sleep(wait)
	}
	conn, err := grpc.Dial(p.ApiAddr, grpc.WithInsecure())
	return conn, err
}

// DmeLocal

type DmeLocal struct {
	Name        string
	NotifyAddrs string
	cmd         *exec.Cmd
}

func (p *DmeLocal) Start() error {
	args := []string{"--notifyAddrs", p.NotifyAddrs}
	var err error
	p.cmd, err = StartLocal(p.Name, "dme-server", args)
	return err
}

func (p *DmeLocal) Stop() {
	StopLocal(p.cmd)
}

func (p *DmeLocal) ConnectAPI() (*grpc.ClientConn, error) {
	// TODO
	return nil, errors.New("TODO")
}

// CrmLocal

type CrmLocal struct {
	Name        string
	NotifyAddrs string
	cmd         *exec.Cmd
}

func (p *CrmLocal) Start() error {
	args := []string{"--notifyAddrs", p.NotifyAddrs}
	var err error
	p.cmd, err = StartLocal(p.Name, "crmserver", args)
	return err
}

func (p *CrmLocal) Stop() {
	StopLocal(p.cmd)
}

func (p *CrmLocal) ConnectAPI() (*grpc.ClientConn, error) {
	// TODO
	return nil, errors.New("TODO")
}

// Support funcs

func StartLocal(name, bin string, args []string) (*exec.Cmd, error) {
	cmd := exec.Command(bin, args...)
	writer := NewColorWriter(name)
	cmd.Stdout = writer
	cmd.Stderr = writer
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
