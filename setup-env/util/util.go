package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

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

func ReadYamlFile(filename string, iface interface{}) error {
	if strings.HasPrefix(filename, "~") {
		filename = strings.Replace(filename, "~", os.Getenv("HOME"), 1)
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("error reading yaml file: %v err: %v\n", filename, err)
	}
	err = yaml.UnmarshalStrict(yamlFile, iface)
	if err != nil {
		return err
	}

	return nil
}
