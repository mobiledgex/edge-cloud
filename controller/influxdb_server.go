package main

import (
	"os"
	"os/exec"

	"github.com/mobiledgex/edge-cloud/cloudcommon/influxsup"
	"github.com/mobiledgex/edge-cloud/log"
)

// Start local InfluxDB server

type InfluxDBServer struct {
	cmd     *exec.Cmd
	logFile *os.File
}

func NewInfluxDBServer() *InfluxDBServer {
	s := InfluxDBServer{}
	return &s
}

func (s *InfluxDBServer) Start(addr string) error {
	// clear out existing database
	datadir := "/var/tmp/.influxdb"
	os.RemoveAll(datadir)

	// set up config file
	configFileName, err := influxsup.SetupInflux(datadir,
		influxsup.WithHttpAddr(addr))
	if err != nil {
		return err
	}
	logFileName := datadir + "/influxdb.log"

	log.InfoLog("Starting local InfluxDB", "config", configFileName)
	s.cmd = exec.Command("influxd", "-config", configFileName)

	logFile, err := os.OpenFile(logFileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	s.cmd.Stdout = logFile
	s.cmd.Stderr = logFile
	err = s.cmd.Start()
	if err != nil {
		logFile.Close()
		return err
	}
	s.logFile = logFile
	return nil
}

func (s *InfluxDBServer) Stop() {
	s.cmd.Process.Kill()
	s.logFile.Close()
	err := s.cmd.Wait()
	if err != nil {
		log.InfoLog("Wait for InfluxDB process failed", "pid", s.cmd.Process.Pid, "err", err)
	}
}
