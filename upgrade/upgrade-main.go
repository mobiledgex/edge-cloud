package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	mexlog "github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	upgrade "github.com/mobiledgex/edge-cloud/upgrade/upgradeutil"
)

var yamlFile = flag.String("yaml", "", "File to be upgraded")
var outputFile = flag.String("output", "upgrade.yml", "Output yaml file(if theinput is a yaml file)")
var etcdUrls = flag.String("etcdUrls", "http://127.0.0.1:2380", "etcd client listener URLs")
var upgradeFuncName = flag.String("upgrade", "", "Upgrade function to run")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", mexlog.DebugLevelStrings))

func main() {
	flag.Parse()
	mexlog.SetDebugLevelStrs(*debugLevels)
	// Upgrading an input yaml file
	if *yamlFile != "" {
		// Fix up the yaml file paths
		if strings.HasPrefix(*yamlFile, "~") {
			*yamlFile = strings.Replace(*yamlFile, "~", os.Getenv("HOME"), 1)
		}
		if strings.HasPrefix(*outputFile, "~") {
			*outputFile = strings.Replace(*outputFile, "~", os.Getenv("HOME"), 1)
		}

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
		outdir := filepath.Dir(*outputFile)
		outfile := filepath.Base(*outputFile)
		fmt.Printf("DIrectory: <%s> and file: <%s>\n", outdir, outfile)
		util.PrintToYamlFile(outfile, outdir, appData, true)
	}
	// Upgrading a running etcd database
	//TODO
}
