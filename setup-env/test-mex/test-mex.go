package main

/* runs a single test case which consists of one or more actions.  Each action will call either a
   controller or DME api, or a setup-mex function to deploy, start, or stop a process */

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/mobiledgex/edge-cloud/setup-env/apis"
	setupmex "github.com/mobiledgex/edge-cloud/setup-env/setup-mex"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

func init() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	actions = flag.String("actions", "", "one or more of: "+actionList+" separated by ,")
}

var deploymentChoices = map[string]bool{"process": true,
	"container": true}
var deploymentList = fmt.Sprintf("%v", reflect.ValueOf(deploymentChoices).MapKeys())

var actionSlice = make([]string, 0)
var actionList = fmt.Sprintf("%v", reflect.ValueOf(actionChoices).MapKeys())

var (
	commandName = "test-mex"
	actions     *string
	deployment  *string
	apiFile     *string
	apiName     *string
	setupFile   *string
	outputDir   *string
	compareYaml *string
	dataDir     *string
)

//re-init the flags because otherwise we inherit a bunch of flags from the testing
//package which get inserted into the usage.
func init() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	actions = flag.String("actions", "", "one or more of: "+actionList+" separated by ,")
	deployment = flag.String("deployment", "process", deploymentList)
	apiFile = flag.String("apifile", "", "optional input yaml file for APIs")
	apiName = flag.String("apiname", "", "name of controller or DME API")
	setupFile = flag.String("setupfile", "", "mandatory yml topology file")
	outputDir = flag.String("outputdir", "", "option directory to store output and logs")
	compareYaml = flag.String("compareyaml", "", "comma separated list of yamls to compare")
	dataDir = flag.String("datadir", "", "optional path of data files")
}

//this is possible actions and optional parameters
var actionChoices = map[string]string{
	"start":     "procname",
	"stop":      "procname",
	"status":    "procname",
	"ctrlapi":   "procname",
	"dmeapi":    "procname",
	"deploy":    "",
	"cleanup":   "",
	"fetchlogs": "",
}

func printUsage() {
	fmt.Println("\nUsage: \n" + commandName + " [options]\n\noptions:")
	flag.PrintDefaults()
}

//some actions have sub arguments associated after equal sign, e.g.--actions stop=ctrl1
func getActionParam(a string) (string, string) {
	argslice := strings.Split(a, "=")
	action := argslice[0]
	actionParam := ""
	if len(argslice) > 1 {
		actionParam = argslice[1]
	}
	return action, actionParam
}

//actions can be split with a dash like ctrlapi-show
func getActionSubtype(a string) (string, string) {
	argslice := strings.Split(a, "-")
	action := argslice[0]
	actionSubtype := ""
	if len(argslice) > 1 {
		actionSubtype = argslice[1]
	}
	return action, actionSubtype
}

func validateArgs() {
	flag.Parse()
	_, validDeployment := deploymentChoices[*deployment]
	errFound := false

	if *actions != "" {
		actionSlice = strings.Split(*actions, ",")
	}
	for _, a := range actionSlice {
		act, actionParam := getActionParam(a)
		action, _ := getActionSubtype(act)

		optionalParam, validAction := actionChoices[action]
		if !validAction {
			fmt.Printf("ERROR: -actions must be one of: %v, received: %s\n", actionList, action)
			errFound = true
		} else if action == "fetchlogs" && *outputDir == "" {
			fmt.Printf("ERROR: cannot use action=fetchlogs option without -outputdir\n")
			errFound = true
		}
		if optionalParam == "" && actionParam != "" {
			fmt.Printf("ERROR: action %v does not take a parameter\n", action)
			errFound = true
		}
	}

	if !validDeployment {
		fmt.Printf("ERROR: -deployment must be one of: %v\n", deploymentList)
		errFound = true
	}
	if *apiFile != "" {
		if _, err := os.Stat(*apiFile); err != nil {
			fmt.Printf("ERROR: file " + *apiFile + " does not exist\n")
			errFound = true
		}
	}

	if *setupFile == "" {
		fmt.Printf("ERROR -setupfile is mandatory\n")
		errFound = true
	} else {
		if _, err := os.Stat(*setupFile); err != nil {
			fmt.Printf("ERROR: file " + *setupFile + " does not exist\n")
			errFound = true
		}
	}

	if *compareYaml != "" {
		yarray := strings.Split(*compareYaml, ",")
		if len(yarray) != 3 {
			fmt.Printf("ERROR: compareyaml must be a string with 2 yaml files and a filetype separated by comma\n")
			errFound = true
		}
	}
	if errFound {
		printUsage()
		os.Exit(1)
	}
}

func main() {
	validateArgs()

	errorsFound := 0
	if *outputDir != "" {
		*outputDir = util.CreateOutputDir(false, *outputDir, commandName+".log")
	}

	if *setupFile != "" {
		setupmex.ReadSetupFile(*setupFile, *dataDir)
	}

	for _, a := range actionSlice {
		act, actionParam := getActionParam(a)
		action, actionSubtype := getActionSubtype(act)

		util.PrintStepBanner("running action: " + a)
		switch action {
		case "deploy":
			if !setupmex.DeployProcesses() {
				errorsFound += 1
			}
		case "start":
			startFailed := false
			if !setupmex.StartProcesses(actionParam, *outputDir) {
				startFailed = true
				errorsFound += 1
			} else {
				if !setupmex.StartRemoteProcesses(actionParam) {
					startFailed = true
					errorsFound += 1
				}
			}
			if startFailed {
				setupmex.StopProcesses(actionParam)
				setupmex.StopRemoteProcesses(actionParam)
				errorsFound += 1
				break

			}
			setupmex.UpdateApiAddrs()
			if !setupmex.WaitForProcesses(actionParam) {
				errorsFound += 1
			}
		case "status":
			setupmex.UpdateApiAddrs()
			if !setupmex.WaitForProcesses(actionParam) {
				errorsFound += 1
			}
		case "stop":
			setupmex.StopProcesses(actionParam)
			if !setupmex.StopRemoteProcesses(actionParam) {
				errorsFound += 1
			}
		case "ctrlapi":
			setupmex.UpdateApiAddrs()
			if !apis.RunControllerAPI(actionSubtype, actionParam, *apiFile, *outputDir) {
				log.Printf("Unable to run api for %s. Check connectivity to controller API\n", action)
				errorsFound += 1
			}
		case "dmeapi":
			setupmex.UpdateApiAddrs()
			if !apis.RunDmeAPI(actionSubtype, actionParam, *apiFile, *outputDir) {
				log.Printf("Unable to run api for %s. Check connectivity to DME API\n", action)
				errorsFound += 1
			}
		case "cleanup":
			if !setupmex.CleanupRemoteProcesses() {
				errorsFound += 1
			}
		case "fetchlogs":
			if !setupmex.FetchRemoteLogs(*outputDir) {
				errorsFound += 1
			}
		default:
			log.Fatal("unexpected action: " + action)
		}
	}
	if *compareYaml != "" {
		//separate the arg into two files and then replace variables
		//	yamlType := strings.Split(*compareYaml, ",")[0]
		s := strings.Split(*compareYaml, ",")
		firstYamlFile := s[0]
		secondYamlFile := s[1]
		fileType := s[2]

		if !util.CompareYamlFiles(firstYamlFile, secondYamlFile, fileType) {
			errorsFound += 1
		}

	}
	if *outputDir != "" {
		fmt.Printf("\nNum Errors found: %d, Results in: %s\n", errorsFound, *outputDir)
		os.Exit(errorsFound)
	}
}
