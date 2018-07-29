package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	log "gitlab.com/bobbae/logrus"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
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
	"run":  crmutil.MEXCreateAppManifest,
	"kill": crmutil.MEXKillAppManifest,
}

var categories = map[string]map[string]func(*crmutil.Manifest) error{
	"cluster":     clusterOps,
	"platform":    platformOps,
	"application": applicationOps,
}

var mainflag = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

func printUsage() {
	mainflag.Usage()
	log.Infoln("e.g. mex [-debug] platform {init|clean} -manifest my.yaml")
	log.Infoln("e.g. mex [-debug] [-platform pl.yaml] cluster {create|remove} -manifest my.yaml")
	log.Infoln("e.g. mex [-debug] [-platform pl.yaml] application {kill|run} -manifest my.yaml")
}

func main() {
	var err error
	debug := mainflag.Bool("debug", false, "debugging")
	quiet := mainflag.Bool("quiet", false, "less verbose")
	help := mainflag.Bool("help", false, "help")
	platform := mainflag.String("platform", "", "platform data")
	if err = mainflag.Parse(os.Args[1:]); err != nil {
		log.Fatalln("parse error", err)
		os.Exit(1)
	}
	if *help {
		printUsage()
		os.Exit(0)
	}
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	args := mainflag.Args()
	log.Debugln("args", args)
	if len(args) < 2 {
		printUsage()
		log.Fatalf("insufficient args")
	}
	_, ok := resourceMap[args[0]]
	if !ok {
		printUsage()
		log.Fatalln("valid resources are", reflect.ValueOf(resourceMap).MapKeys())
		os.Exit(1)
	}
	if !*quiet {
		log.Infoln("platform init should be not done too often. letsencrypt api has 20 per week limit")
	}
	crmutil.MEXInit()
	if *platform != "" {
		mf := &crmutil.Manifest{}
		dat, err := ioutil.ReadFile(*platform)
		if err != nil {
			log.Fatalln("can't read platform from %s", *platform)
			os.Exit(1)
		}
		err = yaml.Unmarshal(dat, mf)
		if err != nil {
			log.Fatalln("can't unmarshal %v", err)
			os.Exit(1)
		}
		rootLB, err := crmutil.NewRootLBManifest(mf)
		if err != nil {
			log.Fatalln("can't get new rootLB, %v", err)
			os.Exit(1)
		}
		log.Debugln("got rootLB", rootLB)
	}
	resourceMap[args[0]](args[1:])
}

func validateCommand(rsrc string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing command")
	}
	commands, ok := categories[rsrc]
	if !ok {
		log.Fatalln("resource %s not avail", rsrc)
		os.Exit(1)
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
		log.Fatal(err)
	}
	cmd := args[0]
	args = args[1:]
	if err := subflags.Parse(args); err != nil {
		log.Fatalln("parse error", err)
		os.Exit(1)
	}
	log.Debugln(kind, cmd, args)
	if *manifest == "" {
		printUsage()
		log.Fatalln("no manifest file")
		os.Exit(1)
	}
	mf := &crmutil.Manifest{}
	dat, err := ioutil.ReadFile(*manifest)
	if err != nil {
		log.Fatalf("can't read %s, %v", *manifest, err)
	}
	err = yaml.Unmarshal(dat, mf)
	if err != nil {
		log.Fatalf("can't unmarshal, %v", err)
	}
	if mf.APIVersion != apiversion {
		log.Fatalf("invalid api version")
	}
	if !strings.Contains(mf.Resource, kind) {
		log.Fatalf("not %s resource", kind)
	}
	err = categories[kind][cmd](mf)
	if err != nil {
		log.Fatal(err)
	}
}
