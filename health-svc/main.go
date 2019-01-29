package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)

	notifyClient := notify.NewClient(strings.Split(*notifyAddrs, ","),
		*tlsCertFile)
	(&AppInstCheck{}).init(notifyClient)
	notifyClient.Start()
	defer notifyClient.Stop()

	// block forever
	select {}
}
