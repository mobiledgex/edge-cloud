package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/log"
)

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
	"create": crmutil.MEXPlatformInitManifest,
	"remove": crmutil.MEXPlatformCleanManifest,
}

var applicationOps = map[string]func(*crmutil.Manifest) error{
	"create": crmutil.MEXAppCreateAppManifest,
	"remove": crmutil.MEXAppDeleteAppManifest,
}

var categories = map[string]map[string]func(*crmutil.Manifest) error{
	"cluster":     clusterOps,
	"platform":    platformOps,
	"application": applicationOps,
}

var mainflag = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

func printUsage() {
	originalUsage()
	fmt.Println("mex -stack myvals.yaml {platform|cluster|application} {create|remove}")
	fmt.Println("mex [-platform pl.yaml] platform {create|remove} -manifest my.yaml")
	fmt.Println("mex [-platform pl.yaml] cluster {create|remove} -manifest my.yaml")
	fmt.Println("mex [-platform pl.yaml] application {create|remove} -manifest my.yaml")
}

var originalUsage func()

func main() {
	var err error
	help := mainflag.Bool("help", false, "help")
	debugLevels := mainflag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
	platform := mainflag.String("platform", "", "platform data")
	tdir := mainflag.String("dir", ".", "base directory containing templates")
	stack := mainflag.String("stack", "", "stack values")
	originalUsage = mainflag.Usage
	mainflag.Usage = printUsage
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
	crmutil.MEXInit()
	mf := &crmutil.Manifest{}
	if *stack != "" {
		if len(args) < 1 {
			printUsage()
			fmt.Println("insufficient args")
			os.Exit(1)
		}
		log.DebugLog(log.DebugLevelMexos, "mf from stack", "file", *stack)
		dat, err := ioutil.ReadFile(*stack)
		if err != nil {
			log.FatalLog("can't read stack", "file", *stack, "error", err)
		}
		err = yaml.Unmarshal(dat, mf)
		if err != nil {
			log.FatalLog("can't unmarshal stack", "error", err)
		}
		mfVals := &mf.Values
		if err := internEnv(mfVals); err != nil {
			log.FatalLog("cannot intern env", "error", err, "mfVals", mfVals)
		}
		if err := crmutil.CheckPlatformEnv(mfVals.OperatorKind); err != nil {
			log.FatalLog("platform env check failed", "error", err, "mfVals", mfVals)
		}
		kind := args[0]
		fnfmt := "%s/%s/%s/" + mfVals.Base + ".yaml"
		fn := fmt.Sprintf(fnfmt, *tdir, kind, mfVals.Operator)
		if kind == "application" {
			fn = fmt.Sprintf(fnfmt, tdir, kind, mfVals.AppKind)
		}
		err = tmplUnmarshal(kind, fn, mf, mfVals)
		if err != nil {
			log.FatalLog("cannot unmarshal", "error", err, "fn", fn, "kind", kind)
		}
		if err := getNewRootLB(mf); err != nil {
			log.FatalLog("can't instantiate new rootLB", "error", err)
		}
		op := args[1]
		if err := crmutil.CheckManifest(mf); err != nil {
			log.FatalLog("incorrect manifest", "error", err)
		}
		log.DebugLog(log.DebugLevelMexos, "call", "kind", kind, "op", op, "mf", mf)
		err = callOps(kind, op, mf)
		if err != nil {
			log.FatalLog("ops failure", "op", op, "kind", kind, "error", err)
		}
		os.Exit(0)
	}
	//log.DebugLog(log.DebugLevelMexos, "platform init should be not done too often. letsencrypt api has 20 per week limit")
	if *platform != "" { // XXX TODO This should be deprecated
		log.DebugLog(log.DebugLevelMexos, "mf platform from", "file", *platform)
		dat, err := ioutil.ReadFile(*platform)
		if err != nil {
			log.FatalLog("can't read platform", "platform", *platform)
		}
		//TODO allow reading manifest data file from https://
		err = yaml.Unmarshal(dat, mf)
		if err != nil {
			log.FatalLog("can't unmarshal", "error", err)
		}
		if err := getNewRootLB(mf); err != nil {
			log.FatalLog("can't instantiate new rootLB", "error", err)
		}
	}
	resourceMap[args[0]](args[1:])
}

func getNewRootLB(mf *crmutil.Manifest) error {
	_, err := crmutil.NewRootLBManifest(mf)
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "got rootLB")
	return nil
}

func internEnv(mfv *crmutil.ValueDetail) error {
	for _, u := range []string{mfv.OpenRC, mfv.MexEnv} {
		dat, err := crmutil.GetVaultData(u)
		if err != nil {
			return err
		}
		vr, err := crmutil.GetVaultEnvResponse(dat)
		if err != nil {
			return err
		}
		//log.DebugLog(log.DebugLevelMexos, "interning vault data", "data", vr)
		err = crmutil.InternVaultEnv(vr.Data.Data.Env)
		if err != nil {
			return err
		}
	}
	return nil
}

func callOps(kind, op string, mf *crmutil.Manifest) error {
	_, ok := categories[kind]
	if !ok {
		return fmt.Errorf("invalid category %s", kind)
	}
	err := categories[kind][op](mf)
	if err != nil {
		return err
	}
	return nil
}

func tmplUnmarshal(tn, tf string, mf *crmutil.Manifest, mfv *crmutil.ValueDetail) error {
	dat, err := ioutil.ReadFile(tf)
	if err != nil {
		return err
	}
	tmpl, err := template.New(tn).Parse(string(dat))
	if err != nil {
		return err
	}
	var outbuffer bytes.Buffer
	err = tmpl.Execute(&outbuffer, mfv)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(outbuffer.Bytes(), mf)
	if err != nil {
		return err
	}
	return nil
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
	if !strings.Contains(mf.Kind, kind) {
		log.FatalLog("mismatch kind", "kind", kind)
	}
	err = categories[kind][cmd](mf)
	if err != nil {
		log.FatalLog("failure", "error", err, "cmd", cmd, "kind", kind)
	}
}
