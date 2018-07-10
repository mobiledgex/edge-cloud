package main

// executes end-to-end MEX tests by calling test-mex multiple times as directed by the input test file.

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

var (
	commandName = "e2e-tests"
	testFile    *string
	testGroup   *string
	outputDir   *string
	setupFile   *string
	dataDir     *string
	stopOnFail  *bool
	verbose     *bool
)

//re-init the flags because otherwise we inherit a bunch of flags from the testing
//package which get inserted into the usage.
func init() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	testFile = flag.String("testfile", "", "input file with tests")
	testGroup = flag.String("testgroup", "", "input file with multiple test files")
	outputDir = flag.String("outputdir", "/tmp/e2e_test_out", "output directory, timestamp will be appended")
	setupFile = flag.String("setupfile", "", "network config setup file")
	dataDir = flag.String("datadir", "$GOPATH/src/github.com/mobiledgex/edge-cloud/setup-env/e2e-tests/data", "directory where app data files exist")
	stopOnFail = flag.Bool("stop", false, "stop on failures")
	verbose = flag.Bool("verbose", false, "prints full output screen")

}

// an individual test
// name of an individual test file and how many times to run it
type e2e_test_file struct {
	Filename string `yaml:"filename"`
	Loops    int    `yaml:"loops"`
}

type e2e_test_files struct {
	Testfiles []e2e_test_file `yaml:"testfiles"`
}

// group of test files, each of which may have multiple tests
type e2e_test_group struct {
	Name      string
	Testfiles []e2e_test_file `yaml:"testfiles"`
}

// a list of tests

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

	if *testFile == "" && *testGroup == "" {
		fmt.Println("Argument -testfile <file> or -testgroup <file> is required")
		errorFound = true
	} else if *testFile != "" && *testGroup != "" {
		fmt.Println("Only one of -testfile <file> or -testgroup <file> is allowed")
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

func readYamlFile(fileName string, tests interface{}) bool {
	//outputdir is always appended as a variable
	varstr := "outputdir=" + *outputDir
	varstr += ",datadir=" + *dataDir

	err := util.ReadYamlFile(fileName, tests, varstr, true)
	if err != nil {
		log.Fatalf("*** Error in reading test file: %v - err: %v\n", *testFile, err)
	}
	return true
}

func runTests(fileName string) (int, int, int) {
	numPassed := 0
	numFailed := 0
	numTestsRun := 0

	basename := path.Base(fileName)

	var testsToRun e2e_tests
	if !readYamlFile(fileName, &testsToRun) {
		return 0, 0, 0
	}

	for _, t := range testsToRun.Tests {
		fmt.Printf("%-25s %-50s ", basename, t.Name)
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

		//		log.Printf("executing: %s\n", cmdstr)
		cmd := exec.Command("sh", "-c", cmdstr)

		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if *verbose {
			fmt.Println(out.String())
		}
		if err == nil {
			fmt.Println("PASS")
			numPassed += 1
		} else {
			fmt.Printf("FAIL: %s\n", stderr.String())
			numFailed += 1
			if *stopOnFail {
				fmt.Printf("*** STOPPING ON FAILURE due to --stop option\n")
				break
			}
		}

		numTestsRun++
	}
	if *verbose {
		fmt.Printf("\n\n*** Summary of testfile %s Tests Run: %d Passed: %d Failed: %d -- Logs in %s\n", fileName, numTestsRun, numPassed, numFailed, *outputDir)
	}
	return numTestsRun, numPassed, numFailed
}

func runTestGroup(fileName string) {
	var testGroup e2e_test_group
	if !readYamlFile(fileName, &testGroup) {
		log.Fatalf("unable to read test group file %s\n", fileName)
	}
	totalRun := 0
	totalPassed := 0
	totalFailed := 0
	//look in same dir for the included test files
	dir := path.Dir(fileName)

	var failedTcs = make(map[string]int)
	for _, tg := range testGroup.Testfiles {
		for i := 0; i <= tg.Loops; i++ {
			//		fmt.Printf("Running testfile %s\n", tg.Filename)
			run, pass, fail := runTests(dir + "/" + tg.Filename)
			totalRun += run
			totalPassed += pass
			totalFailed += fail

			if fail > 0 {
				_, ok := failedTcs[tg.Filename]
				if !ok {
					failedTcs[tg.Filename] = 0
				}
				failedTcs[tg.Filename] += fail
			}
		}
	}
	fmt.Printf("\nTotal Run: %d passed: %d failed: %d\n", totalRun, totalPassed, totalFailed)
	if totalFailed > 0 {
		fmt.Printf("Failed Tests: ")
		for t, f := range failedTcs {
			fmt.Printf("  %s: failures %d\n", t, f)
		}
		fmt.Printf("Logs in %s\n", *outputDir)
	}

}

func main() {
	validateArgs()
	*outputDir = util.CreateOutputDir(true, *outputDir, commandName+".log")

	fmt.Printf("\n%-25s %-50s Result\n", "Testfile", "Test")
	fmt.Printf("-------------------------------------------------------------------------------------\n")
	if *testFile != "" {
		runTests(*testFile)
	}
	if *testGroup != "" {
		runTestGroup(*testGroup)
	}

}
