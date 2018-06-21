package util

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
)

type processinfo struct {
	pid   int
	alive bool
}

//get list of pids for a process name
func getPidsByName(processName string) []processinfo {
	//pidlist is a set of pids and alive bool
	var processes []processinfo
	out, perr := exec.Command("sh", "-c", "pgrep -x "+processName).Output()
	if perr != nil {
		return processes
	}

	for _, pid := range strings.Split(string(out), "\n") {
		p, err := strconv.Atoi(pid)
		if err != nil {
			fmt.Printf("Error in finding pid from process: %v -- %v", processName, err)
		} else {
			pinfo := processinfo{pid: p, alive: true}
			processes = append(processes, pinfo)
		}
	}
	return processes
}

//first tries to kill process with SIGINT, then waits up to maxwait time
//for it to die.  After that point it kills with SIGKILL
func KillProcessesByName(processName string, maxwait time.Duration, c chan string) {
	processes := getPidsByName(processName)
	waitInterval := 100 * time.Millisecond

	for _, p := range processes {
		process, err := os.FindProcess(p.pid)
		if err == nil {
			//try to kill gracefully
			log.Printf("Sending interrupt to process %v pid %v\n", processName, p.pid)
			process.Signal(os.Interrupt)
		}
	}
	for {
		//loop up to maxwait until either all the processes are gone or
		//we run out of waiting time
		time.Sleep(waitInterval)
		maxwait -= waitInterval

		//loop thru all the processes and see if any are still alive
		foundOneAlive := false
		for i, pinfo := range processes {
			if pinfo.alive {
				process, err := os.FindProcess(pinfo.pid)
				if err != nil {
					log.Printf("Error in FindProcess for pid %v - %v\n", pinfo.pid, err)
				}
				if process == nil {
					//this does not happen in linux
					processes[i].alive = false
				} else {
					err = syscall.Kill(pinfo.pid, 0)
					//if we get an error from kill -0 then the process is gone
					if err != nil {
						//marking it dead so we don't revisit it
						processes[i].alive = false
					} else {
						foundOneAlive = true
					}
				}
			}
		}
		if !foundOneAlive {
			c <- "gracefully shut down " + processName
			return
		}
		if maxwait <= 0 {
			break
		}
	}
	for _, pinfo := range processes {
		if pinfo.alive {
			process, _ := os.FindProcess(pinfo.pid)
			if process != nil {
				process.Kill()
			}
		}
	}

	c <- "forcefully shut down " + processName
}

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
//TODO need to handle different types of interfaces besides appdata, currently using
//that to sort
func CompareYamlFiles(firstYamlFile string, secondYamlFile string, replaceVars string) bool {

	PrintStartBanner("compareYamlFiles")

	log.Printf("Comparing %v to %v\n", firstYamlFile, secondYamlFile)
	var y1 edgeproto.ApplicationData
	var y2 edgeproto.ApplicationData

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

	y1.Sort()
	y2.Sort()

	if !cmp.Equal(y1, y2) {
		log.Println("Comparison fail")
		log.Printf(cmp.Diff(y1, y2))
		return false
	}
	log.Println("Comparison success")
	return true
}
