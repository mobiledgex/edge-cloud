// Run Etcd as a child process.
// May be useful for testing and initial development.

package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mobiledgex/edge-cloud/util"
)

// Configuration for starting the etcd service to be used as the backing store
// for the MEX database
type EtcdConfig struct {
	// etcd binary + path
	EtcdBin string
	// member name
	Name string
	// location of database files
	DataDir string
	// ip and port to listen for peer traffic
	PeerIP   string
	PeerPort uint16
	// etcd client URls
	ClientUrls string
	// more stuff later for certs
	LogFile string
}

type EtcdServer struct {
	Config  EtcdConfig
	testDir string
	logfile *os.File
	cmd     *exec.Cmd
	isLocal bool
	started bool
}

const EtcdLocalData string = "etcdLocal_data"
const EtcdLocalLog string = "etcdLocal.log"

func StartLocalEtcdServer() (*EtcdServer, error) {
	_, filename, _, _ := runtime.Caller(0)
	testdir := filepath.Dir(filename) + "/" + EtcdLocalData
	logfile := testdir + "/" + EtcdLocalLog
	config := EtcdConfig{
		EtcdBin:    "etcd",
		Name:       "test",
		DataDir:    testdir,
		PeerIP:     "127.0.0.1",
		PeerPort:   52379,
		ClientUrls: "http://127.0.0.1:52380",
		LogFile:    logfile,
	}
	util.InfoLog("Starting local etcd", "clientUrls", config.ClientUrls)
	server := EtcdServer{}
	err := server.Start(&config)
	if err != nil {
		return nil, err
	}
	server.isLocal = true
	return &server, nil
}

func (e *EtcdServer) Start(config *EtcdConfig) error {
	if net.ParseIP(config.PeerIP) == nil {
		return errors.New("EtcdConfig: Invalid Peer IP")
	}

	peerUrl := fmt.Sprintf("http://%s:%d", config.PeerIP, config.PeerPort)

	e.cmd = exec.Command(config.EtcdBin, "--name", config.Name,
		"--data-dir", config.DataDir, "--listen-peer-urls", peerUrl,
		"--listen-client-urls", config.ClientUrls, "--advertise-client-urls",
		config.ClientUrls)

	logdir := filepath.Dir(config.LogFile)
	err := os.MkdirAll(logdir, 0744)
	if err != nil {
		return err
	}

	logfile, err := os.OpenFile(config.LogFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	e.cmd.Stdout = logfile
	e.cmd.Stderr = logfile
	err = e.cmd.Start()
	if err != nil {
		logfile.Close()
		return err
	}
	e.Config = *config
	e.logfile = logfile
	e.started = true
	return nil
}

func (e *EtcdServer) Stop() {
	if !e.started {
		return
	}
	e.cmd.Process.Kill()
	e.logfile.Close()
	if e.isLocal {
		// clean up all files
		os.RemoveAll(e.Config.DataDir)
	}
	err := e.cmd.Wait()
	if err != nil {
		util.InfoLog("Wait for etcd process failed", "pid", e.cmd.Process.Pid, "err", err)
	}
}
