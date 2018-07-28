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
	"cluster":  clusterHandler,
	"platform": platformHandler,
}

var clusterCommands = map[string]func(*crmutil.Manifest) error{
	"create": crmutil.MEXClusterCreate,
	"remove": crmutil.MEXClusterRemove,
}

var platformCommands = map[string]func(*crmutil.Manifest) error{
	"init":  crmutil.MEXPlatformInit,
	"clean": crmutil.MEXPlatformClean,
}

var resourceCommands = map[string]map[string]func(*crmutil.Manifest) error{
	"cluster":  clusterCommands,
	"platform": platformCommands,
}

var mainflag = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

func printUsage() {
	mainflag.Usage()
	log.Infoln("e.g. mex [-debug] cluster {create|remove} -manifest my.yaml")
	log.Infoln("e.g. mex [-debug] platform {init|clean} -manifest my.yaml")
}

func main() {
	debug := mainflag.Bool("debug", false, "debugging")
	help := mainflag.Bool("help", false, "help")

	if err := mainflag.Parse(os.Args[1:]); err != nil {
		log.Fatalln("parse error", err)
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
	}
	resourceMap[args[0]](args[1:])
}

func validateCommand(rsrc string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing command")
	}

	commands, ok := resourceCommands[rsrc]
	if !ok {
		log.Fatalln("resource %s not avail", rsrc)
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
	}

	log.Debugln(kind, cmd, args)

	if *manifest == "" {
		printUsage()
		log.Fatalln("no manifest file")
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
	err = resourceCommands[kind][cmd](mf)
	if err != nil {
		log.Fatal(err)
	}
}

func clusterHandler(args []string) {
	manifestHandler("cluster", args)
}

func platformHandler(args []string) {
	manifestHandler("platform", args)
}
