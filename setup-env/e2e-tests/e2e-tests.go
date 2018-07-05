package main

// executes end-to-end MEX tests by calling test-mex multiple times as directed by the input test file.

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

var (
	commandName = "e2e-tests"
	testFile    *string
	outputDir   *string
	setupFile   *string
	dataDir     *string
	stopOnFail  *bool
)

//re-init the flags because otherwise we inherit a bunch of flags from the testing
//package which get inserted into the usage.
func init() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	testFile = flag.String("testfile", "", "input file with tests")
	outputDir = flag.String("outputdir", "/tmp/e2e_test_out", "output directory, timestamp will be appended")
	setupFile = flag.String("setupfile", "", "network config setup file")
	dataDir = flag.String("datadir", "$GOPATH/src/github.com/mobiledgex/edge-cloud/setup-env/e2e-tests/data", "directory where app data files exist")
	stopOnFail = flag.Bool("stop", false, "stop on failures")
}

type e2e_test struct {
	Name        string
	Apifile     string
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

func printUsage() {
	fmt.Println("\nUsage: \n" + commandName + " [options]\n\noptions:")
	flag.PrintDefaults()
}

func validateArgs() {

	//re-init the flags so we don't get a bunch of test flags in the usage
	flag.Parse()

	errorFound := false
	if *testFile == "" {
		fmt.Println("Argument -testfile <file> is required")
		errorFound = true
	}
	if *outputDir == "" {
		fmt.Println("Argument -outputdir <dir> is required")
		errorFound = true
	}
	if *setupFile == "" {
		fmt.Println("Argument -setupfile <file> is required")
		errorFound = true
	}
	if *dataDir == "" {
		fmt.Println("Argument -datadir <dir> is required")
		errorFound = true
	} else if strings.Contains(*dataDir, "$GOPATH") && os.Getenv("GOPATH") == "" {
		//note $GOPATH is not evaluated until setup-mex is called
		fmt.Println("Argument -datadir <dir> is required, or set GOPATH properly to use default data dir")
		errorFound = true
	}
	if errorFound {
		printUsage()
		os.Exit(1)
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
		cmdstr := fmt.Sprintf("test-mex -outputdir %s -setupfile %s -datadir %s ", *outputDir, *setupFile, *dataDir)
		if len(t.Actions) > 0 {
			cmdstr += fmt.Sprintf("-actions %s ", strings.Join(t.Actions, ","))
		}
		if t.Apifile != "" {
			cmdstr += fmt.Sprintf("-apifile %s ", t.Apifile)
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

	os.Exit(numFailed)
}

func main() {
	validateArgs()
	*outputDir = util.CreateOutputDir(true, *outputDir, commandName+".log")
	readTestFile()
	runTests()
}
