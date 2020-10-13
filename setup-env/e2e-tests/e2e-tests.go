package main

// executes end-to-end MEX tests by calling test-mex multiple times as directed by the input test file.

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/setup-env/e2e-tests/e2eapi"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

var (
	commandName = "e2e-tests"
	testFile    *string
	outputDir   *string
	setupFile   *string
	varsFile    *string
	stopOnFail  *bool
	verbose     *bool
	notimestamp *bool
	failedTests = make(map[string]int)
)

//re-init the flags because otherwise we inherit a bunch of flags from the testing
//package which get inserted into the usage.
func init() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	testFile = flag.String("testfile", "", "input file with tests")
	outputDir = flag.String("outputdir", "/tmp/e2e_test_out", "output directory, timestamp will be appended")
	setupFile = flag.String("setupfile", "", "network config setup file")
	varsFile = flag.String("varsfile", "", "yaml file containing vars, key: value definitions")
	stopOnFail = flag.Bool("stop", false, "stop on failures")
	verbose = flag.Bool("verbose", false, "prints full output screen")
	notimestamp = flag.Bool("notimestamp", false, "no timestamp on outputdir, logs will be appended to by subsequent runs")
}

// a list of tests, which may include another file which has tests.  Looping can
//be done at either test level or the included file level.
type e2e_test struct {
	Name        string   `yaml:"name"`
	IncludeFile string   `yaml:"includefile"`
	Mods        []string `yaml:"mods"`
	Loops       int      `yaml:"loops"`
}

type e2e_tests struct {
	Description string                   `yaml:"description"`
	Program     string                   `yaml:"program"`
	Tests       []map[string]interface{} `yaml:"tests"`
}

var testsToRun e2e_tests
var e2eHome string
var configStr string
var testConfig e2eapi.TestConfig
var defaultProgram string

func printUsage() {
	fmt.Println("\nUsage: \n" + commandName + " [options]\n\noptions:")
	flag.PrintDefaults()
}

func validateArgs() {
	//re-init the flags so we don't get a bunch of test flags in the usage
	flag.Parse()
	testConfig.Vars = make(map[string]string)

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
	if err := e2eapi.ReadVarsFile(*varsFile, testConfig.Vars); err != nil {
		fmt.Printf("failed to read yaml vars file %s, %v\n", *varsFile, err)
		errorFound = true
	}
	testConfig.SetupFile = *setupFile
	*outputDir = util.CreateOutputDir(!*notimestamp, *outputDir, commandName+".log")
	testConfig.Vars["outputdir"] = *outputDir
	dataDir, found := testConfig.Vars["datadir"]
	if !found {
		dataDir = "$GOPATH/src/github.com/mobiledgex/edge-cloud/setup-env/e2e-tests/data"
		testConfig.Vars["datadir"] = dataDir
	}

	// expand any environment variables in path (like $GOPATH)
	for key, val := range testConfig.Vars {
		testConfig.Vars[key] = os.ExpandEnv(val)
	}

	defaultProgram = testConfig.Vars["default-program"]

	configBytes, err := json.Marshal(&testConfig)
	if err != nil {
		fmt.Printf("failed to marshal TestConfig, %v\n", err)
		errorFound = true
	}
	configStr = string(configBytes)

	if errorFound {
		printUsage()
		os.Exit(1)
	}
}

func readYamlFile(fileName string, tests interface{}) bool {
	err := util.ReadYamlFile(fileName, tests, util.WithVars(testConfig.Vars), util.ValidateReplacedVars())
	if err != nil {
		log.Fatalf("*** Error in reading test file: %v - err: %v\n", *testFile, err)
	}
	return true
}

func parseTest(testinfo map[string]interface{}, test *e2e_test) error {
	// we could use mapstructure here but it's easy to just
	// convert map to json and then unmarshal json.
	spec, err := json.Marshal(testinfo)
	if err != nil {
		return err
	}
	return json.Unmarshal(spec, test)
}

func runTests(dirName, fileName, progName string, depth int, mods []string) (int, int, int) {
	numPassed := 0
	numFailed := 0
	numTestsRun := 0

	if fileName[0] == '/' {
		// absolute path
		dirName = path.Dir(fileName)
		fileName = path.Base(fileName)
	}

	indentstr := ""
	for i := 0; i < depth; i++ {
		indentstr = indentstr + " - "
	}
	var testsToRun e2e_tests
	if !readYamlFile(dirName+"/"+fileName, &testsToRun) {
		log.Printf("\n** unable to read yaml file %s\n", fileName)
		return 0, 0, 0
	}
	if testsToRun.Program != "" {
		progName = testsToRun.Program
	}
	if progName == "" {
		progName = "test-mex"
	}

	//if no loop count specified, run once

	for _, testinfo := range testsToRun.Tests {
		t := e2e_test{}
		err := parseTest(testinfo, &t)
		if err != nil {
			log.Printf("\nfailed to parse test %v, %v\n", testinfo, err)
			numTestsRun++
			numFailed++
			if *stopOnFail {
				return numTestsRun, numPassed, numFailed
			}
			continue
		}
		loopCount := 1
		loopStr := ""

		if t.Loops > loopCount {
			loopCount = t.Loops
		}
		for i := 1; i <= loopCount; i++ {
			if i > 1 {
				loopStr = fmt.Sprintf("(loop %d)", i)
			}
			namestr := t.Name
			if namestr == "" && t.IncludeFile != "" {
				if len(t.IncludeFile) > 58 {
					ilen := len(t.IncludeFile)
					namestr = "include: ..." +
						t.IncludeFile[ilen-58:ilen]
				} else {
					namestr = "include: " + t.IncludeFile
				}
			}
			f := indentstr + fileName
			if len(mods) > 0 {
				f += " " + strings.Join(mods, ",")
			}
			if len(f) > 30 {
				f = f[0:27] + "..."
			}
			fmt.Printf("%-30s %-60s ", f, namestr+loopStr)
			if t.IncludeFile != "" {
				if depth >= 10 {
					//avoid an infinite recusive loop in which a testfile contains itself
					log.Fatalf("excessive include depth %d, possible loop: %s", depth, fileName)
				}
				fmt.Println()
				nr, np, nf := runTests(dirName, t.IncludeFile, progName, depth+1, append(mods, t.Mods...))
				numTestsRun += nr
				numPassed += np
				numFailed += nf
				if *stopOnFail && nf > 0 {
					return numTestsRun, numPassed, numFailed
				}
				continue
			}
			testSpec, err := json.Marshal(testinfo)
			if err != nil {
				fmt.Printf("FAIL: cannot marshal test info %v, %v\n", err, testinfo)
				numTestsRun++
				numFailed++
				if *stopOnFail {
					return numTestsRun, numPassed, numFailed
				}
				continue
			}
			modsSpec, err := json.Marshal(mods)
			if err != nil {
				fmt.Printf("FAIL: cannot marshal mods %v, %v\n", err, mods)
				numTestsRun++
				numFailed++
				if *stopOnFail {
					return numTestsRun, numPassed, numFailed
				}
				continue
			}
			sof := ""
			if *stopOnFail {
				sof = "-stop"
			}
			cmdstr := fmt.Sprintf("%s -testConfig '%s' -testSpec '%s' -mods '%s' %s", progName, configStr, string(testSpec), string(modsSpec), sof)
			cmd := exec.Command("sh", "-c", cmdstr)
			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			err = cmd.Run()
			if *verbose {
				fmt.Println(out.String())
			}
			if err == nil {
				fmt.Println("PASS")
				numPassed += 1
			} else {
				fmt.Printf("FAIL: %s\n", stderr.String())
				numFailed += 1
				_, ok := failedTests[fileName+":"+t.Name]
				if !ok {
					failedTests[fileName+":"+t.Name] = 0
				}
				failedTests[fileName+":"+t.Name] += 1

				if *stopOnFail {
					fmt.Printf("*** STOPPING ON FAILURE due to --stop option\n")
					return numTestsRun, numPassed, numFailed
				}
			}
			numTestsRun++
		}
	}
	if *verbose {
		fmt.Printf("\n\n*** Summary of testfile %s Tests Run: %d Passed: %d Failed: %d -- Logs in %s\n", fileName, numTestsRun, numPassed, numFailed, *outputDir)
	}
	return numTestsRun, numPassed, numFailed

}

func main() {
	validateArgs()

	fmt.Printf("\n%-30s %-60s Result\n", "Testfile", "Test")
	fmt.Printf("-----------------------------------------------------------------------------------------------------\n")
	if *testFile != "" {
		dirName := path.Dir(*testFile)
		fileName := path.Base(*testFile)
		start := time.Now()
		totalRun, totalPassed, totalFailed := runTests(dirName, fileName, defaultProgram, 0, []string{})
		fmt.Printf("\nTotal Run: %d, passed: %d, failed: %d, took: %s\n", totalRun, totalPassed, totalFailed, time.Since(start).String())
		if totalFailed > 0 {
			fmt.Printf("Failed Tests: ")
			for t, f := range failedTests {
				fmt.Printf("  %s: failures %d\n", t, f)
			}
			fmt.Printf("Logs in %s\n", *outputDir)
			os.Exit(1)
		}
	}

}
