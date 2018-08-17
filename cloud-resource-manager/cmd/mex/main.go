package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/log"
)

const apiversion = "v1"

var resourceMap = map[string]func(args []string){
	"cluster":     func(args []string) { manifestHandler("cluster", args) },
	"platform":    func(args []string) { manifestHandler("platform", args) },
	"application": func(args []string) { manifestHandler("application", args) },
}

var clusterOps = map[string]func(*crmutil.Manifest) error{
	"create": crmutil.MEXClusterCreateManifest,
	"remove": crmutil.MEXClusterRemoveManifest,
}

var platformOps = map[string]func(*crmutil.Manifest) error{
	"init":  crmutil.MEXPlatformInitManifest,
	"clean": crmutil.MEXPlatformCleanManifest,
}

var applicationOps = map[string]func(*crmutil.Manifest) error{
	"run":  crmutil.MEXAppCreateAppManifest,
	"kill": crmutil.MEXAppDeleteAppManifest,
}

var categories = map[string]map[string]func(*crmutil.Manifest) error{
	"cluster":     clusterOps,
	"platform":    platformOps,
	"application": applicationOps,
}

var mainflag = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

func printUsage() {
	mainflag.Usage()
	fmt.Println("e.g. mex platform {init|clean} -manifest my.yaml")
	fmt.Println("e.g. mex [-platform pl.yaml] cluster {create|remove} -manifest my.yaml")
	fmt.Println("e.g. mex [-platform pl.yaml] application {kill|run} -manifest my.yaml")
}

func main() {
	var err error
	help := mainflag.Bool("help", false, "help")
	debugLevels := mainflag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
	platform := mainflag.String("platform", "", "platform data")
	if err = mainflag.Parse(os.Args[1:]); err != nil {
		log.FatalLog("parse error", "error", err)
	}
	if *help {
		printUsage()
		os.Exit(0)
	}
	log.SetDebugLevelStrs(*debugLevels)
	//XXX TODO make log to a remote server / aggregator
	args := mainflag.Args()
	if len(args) < 2 {
		printUsage()
		fmt.Println("insufficient args")
		os.Exit(1)
	}
	_, ok := resourceMap[args[0]]
	if !ok {
		printUsage()
		fmt.Println("valid resources are", "resources", reflect.ValueOf(resourceMap).MapKeys())
		os.Exit(1)
	}
	log.DebugLog(log.DebugLevelMexos, "platform init should be not done too often. letsencrypt api has 20 per week limit")
	crmutil.MEXInit()
	if *platform != "" {
		mf := &crmutil.Manifest{}
		dat, err := ioutil.ReadFile(*platform)
		if err != nil {
			log.FatalLog("can't read platform", "platform", *platform)
		}
		//TODO allow reading manifest data file from https://
		err = yaml.Unmarshal(dat, mf)
		if err != nil {
			log.FatalLog("can't unmarshal", "error", err)
		}
		rootLB, err := crmutil.NewRootLBManifest(mf)
		if err != nil {
			log.FatalLog("can't get new rootLB", "error", err)
		}
		log.DebugLog(log.DebugLevelMexos, "got rootLB", "rootLB", rootLB)
	}
	resourceMap[args[0]](args[1:])
}

func validateCommand(rsrc string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing command")
	}
	commands, ok := categories[rsrc]
	if !ok {
		log.FatalLog("resource not avail", "resource", rsrc)
	}
	_, ok = commands[args[0]]
	if !ok {
		return fmt.Errorf("valid commands are %v", reflect.ValueOf(commands).MapKeys())
	}
	return nil
}

func manifestHandler(kind string, args []string) {
	subflags := flag.NewFlagSet(kind, flag.ExitOnError)
	manifest := subflags.String("manifest", "", "manifest for "+kind)
	if err := validateCommand(kind, args); err != nil {
		printUsage()
		log.FatalLog("can't validate command", "error", err)
	}
	cmd := args[0]
	args = args[1:]
	if err := subflags.Parse(args); err != nil {
		log.FatalLog("parse error", "error", err)
	}
	log.DebugLog(log.DebugLevelMexos, "We have", "kind", kind, "cmd", cmd, "args", args)
	if *manifest == "" {
		printUsage()
		log.FatalLog("no manifest file")
	}
	mf := &crmutil.Manifest{}
	dat, err := ioutil.ReadFile(*manifest)
	if err != nil {
		log.FatalLog("can't read", "manifest", *manifest, "error", err)
	}
	err = yaml.Unmarshal(dat, mf)
	if err != nil {
		log.FatalLog("can't unmarshal", "error", err)
	}
	if mf.APIVersion != apiversion {
		log.FatalLog("invalid api version")
	}
	if !strings.Contains(mf.Kind, kind) {
		log.FatalLog("not a resource", "mf", mf, "mf resource", mf.Resource, "kind", kind)
	}
	err = categories[kind][cmd](mf)
	if err != nil {
		log.FatalLog("failure", "error", err, "cmd", cmd, "kind", kind)
	}
}
