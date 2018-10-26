package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
)

var (
	appname           *string
	appvers           *string
	devname           *string
	privKeyFile       *string
	expirationSeconds *int
)

func printUsage() {
	fmt.Println("\nUsage: genauthtoken: [options]")
	flag.PrintDefaults()
}

func init() {
	appname = flag.String("appname", "", "application name")
	appvers = flag.String("appvers", "", "application version")
	devname = flag.String("devname", "", "developer name")
	privKeyFile = flag.String("privkeyfile", "", "private key file")

	expirationSeconds = flag.Int("expSeconds", 60, "expiration seconds")
}

func main() {
	flag.Parse()

	if *privKeyFile == "" {
		fmt.Println("no private key file")
		printUsage()
		os.Exit(1)
	}
	expTime := time.Now().Add(time.Duration(*expirationSeconds) * time.Second).Unix()
	token, err := dmecommon.GenerateAuthToken(*privKeyFile, *devname, *appname, *appvers, expTime)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Token: \n%s\n", token)
}
