package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/mc/orm"
)

var addr = flag.String("addr", "127.0.0.1:9900", "REST listener address")
var sqlAddr = flag.String("sqlAddr", "127.0.0.1:5432", "Postgresql address")
var localSql = flag.Bool("localSql", false, "Run local postgres db")
var initSql = flag.Bool("initSql", false, "Init db when using localSql")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file")
var tlsKeyFile = flag.String("tlskey", "", "server tls key file")
var clientCert = flag.String("clientCert", "", "internal tls client cert file")
var vaultAddr = flag.String("vaultAddr", "http://127.0.0.1:8200", "Vault address")
var localVault = flag.Bool("localVault", false, "Run local Vault")
var ldapAddr = flag.String("ldapAddr", "127.0.0.1:9389", "LDAP listener address")
var gitlabAddr = flag.String("gitlabAddr", "http://127.0.0.1:80", "Gitlab server address")
var pingInterval = flag.Duration("pingInterval", 20*time.Second, "SQL database ping keep-alive interval")

var sigChan chan os.Signal

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	sigChan = make(chan os.Signal, 1)

	config := orm.ServerConfig{
		ServAddr:     *addr,
		SqlAddr:      *sqlAddr,
		VaultAddr:    *vaultAddr,
		RunLocal:     *localSql,
		InitLocal:    *initSql,
		LocalVault:   *localVault,
		TlsCertFile:  *tlsCertFile,
		TlsKeyFile:   *tlsKeyFile,
		LDAPAddr:     *ldapAddr,
		GitlabAddr:   *gitlabAddr,
		ClientCert:   *clientCert,
		PingInterval: *pingInterval,
	}
	server, err := orm.RunServer(&config)
	if err != nil {
		log.FatalLog("Failed to run orm server", "err", err)
	}
	defer server.Stop()

	// wait until process is killed/interrupted
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
