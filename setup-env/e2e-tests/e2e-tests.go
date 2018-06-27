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
	setupFile   = flag.String("setupfile", "", "network config setup file")
	dataDir     = flag.String("datadir", "", "directory where app data files exist")
	stopOnFail  = flag.Bool("stop", false, "stop on failures")
)

type e2e_test struct {
	Name        string
	Appfile     string
	Merfile     string
	Actions     []string
	Compareyaml struct {
		Yaml1    string
		Yaml2    string
		Filetype string
	}
}

type e2e_tests struct {
	Description string     `yaml:"description"`
	Tests       []e2e_test `yaml:"tests"`
}

var testsToRun e2e_tests
var e2eHome string

func validateArgs() {

	flag.Parse()
	if *testFile == "" {
		log.Fatal("Argument -testfile <file> is required")
	}
	if *outputDir == "" {
		log.Fatal("Argument -outputdir <dir> is required")
	}
	if *setupFile == "" {
		log.Fatal("Argument -setupfile <file> is required")
	}
	if *dataDir == "" {
		log.Fatal("Argument -datadir <dir> is required")
	}

}

func readTestFile() bool {
	//outputdir is always appended as a variable
	varstr := "outputdir=" + *outputDir
	varstr += ",datadir=" + *dataDir

	err := util.ReadYamlFile(*testFile, &testsToRun, varstr, false)
	if err != nil {
		log.Printf("*** Error in reading test file: %v - err: %v\n", *testFile, err)
		if strings.Contains(string(err.Error()), "Unreplaced variables") {
			log.Printf("\n** re-run with -vars varname=value\n")
		}
	}
	return true
}

func runTests() {
	numPassed := 0
	numFailed := 0
	numTestsRun := 0

	for _, t := range testsToRun.Tests {
		util.PrintStartBanner("Starting test: " + t.Name)
		cmdstr := fmt.Sprintf("setup-mex -outputdir %s -setupfile %s ", *outputDir, *setupFile)
		if len(t.Actions) > 0 {
			cmdstr += fmt.Sprintf("-actions %s ", strings.Join(t.Actions, ","))
		}
		if t.Appfile != "" {
			cmdstr += fmt.Sprintf("-appfile %s ", t.Appfile)
		}
		if t.Merfile != "" {
			cmdstr += fmt.Sprintf("-merfile %s ", t.Merfile)
		}
		if t.Compareyaml.Yaml1 != "" {
			cmdstr += fmt.Sprintf("-compareyaml %s,%s,%s ", t.Compareyaml.Yaml1, t.Compareyaml.Yaml2, t.Compareyaml.Filetype)
		}

		fmt.Printf("executing: %s\n", cmdstr)
		cmd := exec.Command("sh", "-c", cmdstr)
		out, err := cmd.CombinedOutput()
		fmt.Println(string(out))
		if err == nil {
			log.Printf("\n        *** PASS: %v\n", t.Name)
			numPassed += 1
		} else {
			log.Printf("\n        *** FAIL: %v -- %v\n", t.Name, err)
			numFailed += 1
			if *stopOnFail {
				log.Printf("*** STOPPING ON FAILURE due to --stop option\n")
				break
			}
		}
		numTestsRun++
	}
	log.Printf("\n\n*** Summary of testfile %s Tests Run: %d Passed: %d Failed: %d -- Logs in %s\n", *testFile, numTestsRun, numPassed, numFailed, *outputDir)

}

func main() {
	validateArgs()
	*outputDir = util.CreateOutputDir(true, *outputDir, commandName+".log")
	readTestFile()
	runTests()
}
