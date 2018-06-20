package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

var (
	commandName = "e2e-tests"
	testFile    = flag.String("testfile", "", "input file with tests")
	outputDir   = flag.String("outputdir", "", "output directory, timestamp will be appended")
	stopOnFail  = flag.Bool("stop", false, "stop on failures")
)

type e2e_test struct {
	Name        string
	Datafile    string
	Setupfile   string
	Actions     []string
	Compareyaml struct {
		Yaml1 string
		Yaml2 string
	}
}

type e2e_tests struct {
	Tests []e2e_test `yaml:"tests"`
}

var testsToRun e2e_tests

func validateArgs() {
	flag.Parse()
	if *testFile == "" {
		log.Fatal("Argument -testfile <file> is required")
	}
	if *outputDir == "" {
		log.Fatal("Argument -outputdir <dir> is required")
	}
}

func readTestFile() {
	err := util.ReadYamlFile(*testFile, &testsToRun)
	if err != nil {
		log.Fatalf("Error in reading test file: %v - err: %v\n", *testFile, err)
	}
}

func runTests() {
	numPassed := 0
	numFailed := 0
	numTestsRun := 0
	for _, t := range testsToRun.Tests {
		util.PrintStartBanner(t.Name)
		cmdstr := fmt.Sprintf("setup-mex -outputdir %s ", *outputDir)
		if len(t.Actions) > 0 {
			cmdstr += fmt.Sprintf("-actions %s ", strings.Join(t.Actions, ","))
		}
		if t.Setupfile != "" {
			cmdstr += fmt.Sprintf("-setupfile %s ", t.Setupfile)
		}
		if t.Datafile != "" {
			cmdstr += fmt.Sprintf("-datafile %s ", t.Datafile)
		}
		if t.Compareyaml.Yaml1 != "" {
			cmdstr += fmt.Sprintf("-compareyaml %s,%s ", t.Compareyaml.Yaml1, t.Compareyaml.Yaml2)
		}

		fmt.Printf("executing: %s\n", cmdstr)
		cmd := exec.Command("sh", "-c", cmdstr)
		out, err := cmd.CombinedOutput()
		fmt.Println(string(out))
		if err == nil {
			log.Printf("-- PASS: %v\n", t.Name)
			numPassed += 1
		} else {
			log.Printf("-- FAIL: %v -- %v\n", t.Name, err)
			numFailed += 1
			if *stopOnFail {
				log.Printf("*** STOPPING ON FAILURE due to --stop option\n")
				break
			}
		}
		numTestsRun++
	}
	log.Printf("\n\n*** Summary: Tests Run: %d Passed: %d Failed: %d\n", numTestsRun, numPassed, numFailed)

}

func main() {
	validateArgs()
	*outputDir = util.CreateOutputDir(true, *outputDir, commandName+".log")
	readTestFile()
	runTests()
}
