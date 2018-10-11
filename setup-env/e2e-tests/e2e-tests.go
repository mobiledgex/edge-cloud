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
	outputDir   *string
	setupFile   *string
	dataDir     *string
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
	dataDir = flag.String("datadir", "$GOPATH/src/github.com/mobiledgex/edge-cloud/setup-env/e2e-tests/data", "directory where app data files exist")
	stopOnFail = flag.Bool("stop", false, "stop on failures")
	verbose = flag.Bool("verbose", false, "prints full output screen")
	notimestamp = flag.Bool("notimestamp", false, "no timestamp on outputdir, logs will be overwritten by subsequent runs")
}

// a list of tests, which may include another file which has tests.  Looping can
//be done at either test level or the included file level.
type e2e_test struct {
	Name        string   `yaml:"name"`
	IncludeFile string   `yaml:"includefile"`
	Loops       int      `yaml:"loops"`
	Apifile     string   `yaml:"apifile"`
	Actions     []string `yaml:"actions"`
	Compareyaml struct {
		Yaml1    string `yaml:"yaml1"`
		Yaml2    string `yaml:"yaml2"`
		Filetype string ` yaml:"filetype"`
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

func runTests(dirName string, fileName string, depth int) (int, int, int) {
	numPassed := 0
	numFailed := 0
	numTestsRun := 0

	indentstr := ""
	for i := 0; i < depth; i++ {
		indentstr = indentstr + " - "
	}
	var testsToRun e2e_tests
	if !readYamlFile(dirName+"/"+fileName, &testsToRun) {
		log.Printf("\n** unable to read yaml file %s\n", fileName)
		return 0, 0, 0
	}

	//if no loop count specified, run once

	for _, t := range testsToRun.Tests {
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
				namestr = "include: " + t.IncludeFile
			}
			f := indentstr + fileName
			if len(f) > 30 {
				f = f[0:27] + "..."
			}
			fmt.Printf("%-30s %-60s ", f, namestr+loopStr)
			if t.IncludeFile != "" {
				if len(t.Actions) > 0 || t.Compareyaml.Yaml1 != "" {
					log.Fatalf("Test %s cannot have both included files and actions or yaml compares\n", t.Name)
				}
				if depth >= 10 {
					//avoid an infinite recusive loop in which a testfile contains itself
					log.Fatalf("excessive include depth %d, possible loop: %s", depth, fileName)
				}
				fmt.Println()
				nr, np, nf := runTests(dirName, t.IncludeFile, depth+1)
				numTestsRun += nr
				numPassed += np
				numFailed += nf
				if *stopOnFail && nf > 0 {
					return numTestsRun, numPassed, numFailed
				}
				continue
			}

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
	*outputDir = util.CreateOutputDir(!*notimestamp, *outputDir, commandName+".log")

	fmt.Printf("\n%-30s %-60s Result\n", "Testfile", "Test")
	fmt.Printf("-----------------------------------------------------------------------------------------------------\n")
	if *testFile != "" {
		dirName := path.Dir(*testFile)
		fileName := path.Base(*testFile)
		totalRun, totalPassed, totalFailed := runTests(dirName, fileName, 0)
		fmt.Printf("\nTotal Run: %d passed: %d failed: %d\n", totalRun, totalPassed, totalFailed)
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
