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
	s.HttpAddr = "127.0.0.1:8086"
	s.BindAddr = "127.0.0.1:8088"
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
