package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	upgrade "github.com/mobiledgex/edge-cloud/upgrade/upgradeutil"
)

var region = flag.Uint("region", 1, "Region")
var yamlFile = flag.String("yaml", "", "File to be upgraded")
var outputFile = flag.String("output", "upgrade.yml", "Output yaml file(if theinput is a yaml file)")
var etcdUrls = flag.String("etcdUrls", "http://127.0.0.1:2380", "etcd client listener URLs")
var upgradeFuncName = flag.String("upgrade", "", "Upgrade function to run")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))

func main() {
	// Running count of upgraded entries
	var upgCnt uint
	var err error
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	fmt.Printf("Using function %s to upgrade\n", *upgradeFuncName)
	// Upgrading an input yaml file
	if *yamlFile != "" {
		var appData *edgeproto.ApplicationData
		// Fix up the yaml file paths
		if strings.HasPrefix(*yamlFile, "~") {
			*yamlFile = strings.Replace(*yamlFile, "~", os.Getenv("HOME"), 1)
		}
		if strings.HasPrefix(*outputFile, "~") {
			*outputFile = strings.Replace(*outputFile, "~", os.Getenv("HOME"), 1)
		}

		uf, found := upgrade.UpgradeYamlFuncs[*upgradeFuncName]
		if !found {
			panic(fmt.Errorf("Upgrade function %s not found.", *upgradeFuncName))
		}

		upgradeFunc, ok := uf.(func(string) (*edgeproto.ApplicationData, uint, error))
		if !ok {
			panic(fmt.Errorf("Upgrade function if invalid %v\n", uf))
		}
		appData, upgCnt, err = upgradeFunc(*yamlFile)
		if err != nil {
			panic(fmt.Errorf("Failed to do upgrade a file(%s) err:%v\n", *upgradeFuncName, err))
		}
		outdir := filepath.Dir(*outputFile)
		outfile := filepath.Base(*outputFile)
		fmt.Printf("DIrectory: <%s> and file: <%s>\n", outdir, outfile)
		util.PrintToYamlFile(outfile, outdir, appData, true)
	} else {
		// Upgrading a running etcd database
		uf, found := upgrade.UpgradeFuncs[*upgradeFuncName]
		if !found {
			panic(fmt.Errorf("Upgrade function %s not found.", *upgradeFuncName))
		}
		upgradeFunc, ok := uf.(func(string, uint) (uint, error))
		if !ok {
			panic(fmt.Errorf("Upgrade function if invalid %v\n", uf))
		}

		if upgCnt, err = upgradeFunc(*etcdUrls, *region); err != nil {
			panic(fmt.Errorf("Failed to upgrade etcd database, err: %v\n", err))
		}
	}
	fmt.Printf("Object upgrade complete - Updated %d entries\n", upgCnt)
}
