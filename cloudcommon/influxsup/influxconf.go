// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package influxsup

import (
	"bufio"
	"fmt"
	"html/template"
	"os"
)

// SsetupInflux sets up the InfluxDB dir and config file.
// The config file name is returned to be used when running InfluxDB.
// The passed in directory is where all the data and config file
// will be placed.
// Note that InfluxDB requires a config file, and does not accept
// configuration options as command line arguments.
var DefaultHttpAddr = "127.0.0.1:8086"
var DefaultBindAddr = "127.0.0.1:8088"

func SetupInflux(dir string, ops ...InfluxOp) (string, error) {
	err := os.MkdirAll(dir, 0744)
	if err != nil {
		return "", err
	}

	configFileName := fmt.Sprintf("%s/influxdb.conf", dir)
	configFile, err := os.OpenFile(configFileName,
		os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer configFile.Close()

	wr := bufio.NewWriter(configFile)
	tmpl := template.Must(template.New("conf").Parse(influxConfTemplate))
	opts := InfluxOptions{}
	opts.SetDefaults(dir)
	opts.Apply(ops...)
	err = tmpl.Execute(wr, opts)
	if err != nil {
		return "", err
	}
	wr.Flush()
	return configFileName, nil
}

type InfluxOptions struct {
	// DataDir is where all the data will be stored (plus config file)
	DataDir string
	// HttpAddr is the listener addr for clients
	HttpAddr string
	// BindAddr is the listener addr for clustering multiple instances
	BindAddr string
	// TLS server certificate
	TlsCertfile string
	// TLS server cert key
	TlsCertKey string
	// Authentication
	Auth bool
}

type InfluxOp func(opts *InfluxOptions)

func WithHttpAddr(addr string) InfluxOp {
	return func(opts *InfluxOptions) { opts.HttpAddr = addr }
}

func WithBindAddr(addr string) InfluxOp {
	return func(opts *InfluxOptions) { opts.BindAddr = addr }
}

func WithServerCert(cert string) InfluxOp {
	return func(opts *InfluxOptions) { opts.TlsCertfile = cert }
}

func WithServerCertKey(key string) InfluxOp {
	return func(opts *InfluxOptions) { opts.TlsCertKey = key }
}

func WithAuth(auth bool) InfluxOp {
	return func(opts *InfluxOptions) { opts.Auth = auth }
}

func (s *InfluxOptions) Apply(ops ...InfluxOp) {
	for _, fn := range ops {
		fn(s)
	}
}

func (s *InfluxOptions) SetDefaults(dir string) {
	s.DataDir = dir
	s.HttpAddr = DefaultHttpAddr
	s.BindAddr = DefaultBindAddr
}

var influxConfTemplate = `
bind-address = "{{.BindAddr}}"

[meta]
  dir = "{{.DataDir}}/meta"

[data]
  dir = "{{.DataDir}}/data"
  wal-dir = "{{.DataDir}}/wal"

[http]
  enabled = true
  bind-address = "{{.HttpAddr}}"
{{if .Auth}}
  auth-enabled = true
{{end}}
{{if .TlsCertfile}}
  https-enabled = true
  https-certificate = "{{.TlsCertfile}}"
  https-private-key = "{{.TlsCertKey}}"
{{end}}
`
