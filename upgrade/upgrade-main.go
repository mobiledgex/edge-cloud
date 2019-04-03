package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	mexlog "github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	upgrade "github.com/mobiledgex/edge-cloud/upgrade/upgradeutil"
)

var yamlFile = flag.String("yaml", "", "File to be upgraded")
var outputFile = flag.String("output", "upgrade.yml", "Output yaml file if the input is a yaml file")
var etcdUrls = flag.String("etcdUrls", "http://127.0.0.1:2380", "etcd client listener URLs")
var upgradeFuncName = flag.String("upgrade", "", "Upgrade function to run")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", mexlog.DebugLevelStrings))

func main() {
	flag.Parse()
	mexlog.SetDebugLevelStrs(*debugLevels)
	// Upgrading an input yaml file
	if *yamlFile != "" {
		uf, found := upgrade.UpgradeYamlFuncs[*upgradeFuncName]
		if !found {
			log.Fatalf("Upgrade function %s not found.", *upgradeFuncName)
		}

		upgradeFunc, ok := uf.(func(string) (*edgeproto.ApplicationData, error))
		if !ok {
			log.Fatalf("Upgrade function if invalid %v\n", uf)
		}
		fmt.Printf("Using function %s to upgrade\n", *upgradeFuncName)

		appData, err := upgradeFunc(*yamlFile)
		if err != nil {
			log.Fatalf("Failed to do upgrade a file(%s) err:%v\n", *upgradeFuncName, err)
		}
		log.Printf("writing to file: %s\n", *outputFile)
		util.PrintToYamlFile(*outputFile, "", appData, true)
	}
	// Upgrading a running etcd database
	//TODO
}
