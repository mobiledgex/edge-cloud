// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

/* runs a single test case which consists of one or more actions.  Each action will call either a
   controller or DME api, or a setup-mex function to deploy, start, or stop a process */

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	log "github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/setup-env/e2e-tests/e2eapi"
	setupmex "github.com/mobiledgex/edge-cloud/setup-env/setup-mex"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

var actionList = fmt.Sprintf("%v", reflect.ValueOf(actionChoices).MapKeys())

var (
	commandName = "test-mex"
	configStr   *string
	specStr     *string
	modsStr     *string
	outputDir   string
	stopOnFail  *bool
)

//re-init the flags because otherwise we inherit a bunch of flags from the testing
//package which get inserted into the usage.
func init() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	configStr = flag.String("testConfig", "", "json formatted TestConfig")
	specStr = flag.String("testSpec", "", "json formatted TestSpec")
	modsStr = flag.String("mods", "", "json formatted mods")
	stopOnFail = flag.Bool("stop", false, "stop on failures")
}

//this is possible actions and optional parameters
var actionChoices = map[string]string{
	"start":      "procname",
	"stop":       "procname",
	"status":     "procname",
	"ctrlapi":    "procname",
	"ctrlinfo":   "procname",
	"dmeapi":     "procname",
	"influxapi":  "procname",
	"exec":       "",
	"cleanup":    "",
	"gencerts":   "",
	"cleancerts": "",
	"cmds":       "",
	"sleep":      "seconds",
	"clientshow": "workerId",
}

func printUsage() {
	fmt.Println("\nUsage: \n" + commandName + " [options]\n\noptions:")
	flag.PrintDefaults()
}

func validateArgs(config *e2eapi.TestConfig, spec *util.TestSpec) {
	errFound := false

	errs := config.Validate()
	for _, err := range errs {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		errFound = true
	}
	outputDir = config.Vars["outputdir"]

	for _, a := range spec.Actions {
		act, actionParam := setupmex.GetActionParam(a)
		action, _ := setupmex.GetActionSubtype(act)

		optionalParam, validAction := actionChoices[action]
		if !validAction {
			fmt.Fprintf(os.Stderr, "ERROR: -actions must be one of: %v, received: %s\n", actionList, action)
			errFound = true
		} else if action == "fetchlogs" && outputDir == "" {
			fmt.Fprintf(os.Stderr, "ERROR: cannot use action=fetchlogs option without -outputdir\n")
			errFound = true
		}
		if optionalParam == "" && actionParam != "" {
			fmt.Fprintf(os.Stderr, "ERROR: action %v does not take a parameter\n", action)
			errFound = true
		}
	}

	if spec.ApiFile != "" {
		files := strings.Split(spec.ApiFile, ",")
		for _, file := range files {
			if _, err := os.Stat(file); err != nil {
				fmt.Fprint(os.Stderr, "ERROR: file "+file+" does not exist\n")
				errFound = true
			}
		}
	}
	if spec.ApiType != "" {
		if spec.ApiType != "rest" && spec.ApiType != "grpc" {
			fmt.Fprintf(os.Stderr, "ERROR - apitype invalid")
			errFound = true
		}
	}

	if errFound {
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	util.SetLogFormat()

	config := e2eapi.TestConfig{}
	spec := util.TestSpec{}
	mods := []string{}

	err := json.Unmarshal([]byte(*configStr), &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unmarshaling TestConfg: %v", err)
		os.Exit(1)
	}
	err = json.Unmarshal([]byte(*specStr), &spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unmarshaling TestSpec: %v", err)
		os.Exit(1)
	}
	err = json.Unmarshal([]byte(*modsStr), &mods)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unmarshaling mods: %v", err)
		os.Exit(1)
	}
	validateArgs(&config, &spec)

	errors := []string{}
	if outputDir != "" {
		outputDir = util.CreateOutputDir(false, outputDir, commandName+".log")
	}

	if config.SetupFile != "" {
		if !setupmex.ReadSetupFile(config.SetupFile, &util.Deployment, config.Vars) {
			os.Exit(1)
		}
		util.DeploymentReplacementVars = config.Vars
	}

	retry := setupmex.NewRetry(spec.RetryCount, spec.RetryIntervalSec, len(spec.Actions))
	ranTest := false
	for {
		tryErrs := []string{}
		for ii, a := range spec.Actions {
			if !retry.ShouldRunAction(ii) {
				continue
			}
			util.PrintStepBanner("name: " + spec.Name)
			util.PrintStepBanner("running action: " + a + retry.Tries())
			actionretry := false
			errs := setupmex.RunAction(ctx, a, outputDir, &spec, mods, config.Vars, &actionretry)
			ranTest = true
			if len(errs) > 0 {
				if actionretry {
					// potential errors that may be ignored after retry
					tryErrs = append(tryErrs, errs...)
				} else {
					// no retry for action, so register errs as final errors
					errors = append(errors, errs...)
					if *stopOnFail {
						break
					}
				}
			}
			retry.SetActionRetry(ii, actionretry)
		}
		if *stopOnFail && len(errors) > 0 {
			// stopOnFail case
			break
		}
		if spec.CompareYaml.Yaml1 != "" && spec.CompareYaml.Yaml2 != "" {
			pass := util.CompareYamlFiles(spec.Name, spec.Actions, &spec.CompareYaml)
			if !pass {
				tryErrs = append(tryErrs, "compare yaml failed")
			}
			ranTest = true
		}
		if len(tryErrs) == 0 || retry.Done() {
			errors = append(errors, tryErrs...)
			break
		}
		fmt.Printf("encountered failures, will retry:\n")
		for _, e := range tryErrs {
			fmt.Printf("- %s\n", e)
		}
		fmt.Printf("")
	}
	if !ranTest {
		errors = append(errors, "no test content")
	}

	fmt.Printf("\nNum Errors found: %d, Results in: %s\n", len(errors), outputDir)
	if len(errors) > 0 {
		errstring := strings.Join(errors, ",")
		fmt.Fprint(os.Stderr, errstring)
		os.Exit(len(errors))
	}
	os.Exit(0)
}
