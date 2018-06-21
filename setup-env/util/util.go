package util

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
)

func PrintStartBanner(label string) {
	log.Printf("\n\n************ Begin: %s ************\n\n ", label)
}

//creates an output directory with an optional timestamp.  Server log files, output from APIs, and
//output from the script itself will all go there if specified
func CreateOutputDir(useTimestamp bool, outputDir string, logFileName string) string {
	if useTimestamp {
		startTimestamp := time.Now().Format("2006-01-02T15:04:05")
		outputDir = outputDir + "/" + startTimestamp
	}
	fmt.Printf("Creating output dir: %s\n", outputDir)
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("Error trying to create directory %v: %v\n", outputDir, err)
	}

	logName := outputDir + "/" + logFileName
	logFile, err := os.OpenFile(logName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		log.Fatalf("Error creating logfile %s\n", logName)
	}
	//log to both stdout and logfile
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	return outputDir
}

func ReadYamlFile(filename string, iface interface{}, varlist string) error {
	if strings.HasPrefix(filename, "~") {
		filename = strings.Replace(filename, "~", os.Getenv("HOME"), 1)
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.New(fmt.Sprintf("error reading yaml file: %v err: %v\n", filename, err))

	}
	if varlist != "" {
		//replace variables denoted as {{variablename}}
		yamlstr := string(yamlFile)
		vars := strings.Split(varlist, ",")
		for _, va := range vars {
			k := strings.Split(va, "=")[0]
			v := strings.Split(va, "=")[1]
			yamlstr = strings.Replace(yamlstr, "{{"+k+"}}", v, -1)
		}
		yamlFile = []byte(yamlstr)
	}
	//make sure there are no unreplaced variables left and inform the user if so
	re := regexp.MustCompile("{{(\\S+)}}")
	matches := re.FindAllStringSubmatch(string(yamlFile), 1)
	if len(matches) > 0 {
		return errors.New(fmt.Sprintf("Unreplaced variables in yaml: %v", matches))
	}

	err = yaml.UnmarshalStrict(yamlFile, iface)
	if err != nil {
		return err
	}

	return nil
}

//compares two yaml files for equivalence
func CompareYamlFiles(firstYamlFile string, secondYamlFile string, replaceVars string) bool {

	PrintStartBanner("compareYamlFiles")

	log.Printf("Comparing %v to %v\n", firstYamlFile, secondYamlFile)
	var y1 interface{}
	var y2 interface{}

	err1 := ReadYamlFile(firstYamlFile, &y1, replaceVars)
	if err1 != nil {
		log.Printf("error reading yaml file: %s\n", firstYamlFile)
		return false
	}
	err2 := ReadYamlFile(secondYamlFile, &y2, replaceVars)
	if err2 != nil {
		log.Printf("error reading yaml file: %s\n", secondYamlFile)
		return false
	}

	if !cmp.Equal(y1, y2) {
		log.Println("Comparison fail")
		log.Printf(cmp.Diff(y1, y2))
		return false
	}
	log.Println("Comparison success")
	return true
}
