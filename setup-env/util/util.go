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
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
)

var Deployment DeploymentData

type yamlFileType int

const (
	YamlAppdata yamlFileType = 0
	YamlOther   yamlFileType = 1
)

type YamlReplacementVariables struct {
	Vars []map[string]string
}

// replacement variables taken from the setup
var DeploymentReplacementVars string

type ProcessInfo struct {
	pid   int
	alive bool
}

type ReturnCodeWithText struct {
	Success bool
	Text    string
}

type GoogleCloudInfo struct {
	Cluster     string
	Zone        string
	MachineType string
}

type K8sPod struct {
	PodName  string
	PodCount int
	MaxWait  int
}

type K8CopyFile struct {
	PodName string
	Src     string
	Dest    string
}

type K8sDeploymentStep struct {
	File        string
	Description string
	WaitForPods []K8sPod
	CopyFile    K8CopyFile
}

type EtcdProcess struct {
	process.EtcdLocal
	Hostname string
}
type ControllerProcess struct {
	process.ControllerLocal
	Hostname string
}
type DmeProcess struct {
	process.DmeLocal
	Hostname string
}
type CrmProcess struct {
	process.CrmLocal
	Hostname string
}
type LocSimProcess struct {
	process.LocApiSimLocal
	Hostname string
}
type TokSimProcess struct {
	process.TokSrvSimLocal
	Hostname string
}
type SampleAppProcess struct {
	process.SampleAppLocal
	Args     []string
	Hostname string
}

type DnsRecord struct {
	Name    string
	Type    string
	Content string
}

//cloudflare dns records
type CloudflareDNS struct {
	Zone    string
	Records []DnsRecord
}

type DeploymentData struct {
	GCloud        GoogleCloudInfo     `yaml:"gcloud"`
	K8sDeployment []K8sDeploymentStep `yaml:"k8s-deployment"`
	Locsims       []LocSimProcess     `yaml:"locsims"`
	Toksims       []TokSimProcess     `yaml:"toksims"`
	Etcds         []EtcdProcess       `yaml:"etcds"`
	Controllers   []ControllerProcess `yaml:"controllers"`
	Dmes          []DmeProcess        `yaml:"dmes"`
	Crms          []CrmProcess        `yaml:"crms"`
	SampleApps    []SampleAppProcess  `yaml:"sampleapps"`
	Cloudflare    CloudflareDNS       `yaml:"cloudflare"`
}

//these are strings which may be present in the yaml but not in the corresponding data structures.
//These are the only allowed exceptions to the strict yaml unmarshalling
var yamlExceptions = map[string]map[string]bool{
	"setup": {
		"vars": true,
	},
	"appdata": {
		"ip_str": true, // ansible workaround
	},
}

func IsK8sDeployment() bool {
	return Deployment.GCloud.Cluster != "" //TODO Azure
}

func IsYamlOk(e error, yamltype string) bool {
	rc := true
	errstr := e.Error()
	for _, err1 := range strings.Split(errstr, "\n") {
		allowedException := false
		for ye := range yamlExceptions[yamltype] {
			if strings.Contains(err1, ye) {
				allowedException = true
			}
		}

		if allowedException || strings.Contains(err1, "yaml: unmarshal errors") {
			// ignore this summary error
		} else {
			//all other errors are unexpected and mean something is wrong in the yaml
			log.Printf("Fatal Unmarshal Error in: %v\n", err1)
			rc = false
		}
	}
	return rc
}

//get list of pids for a process name
func getPidsByName(processName string, processArgs string) []ProcessInfo {
	//pidlist is a set of pids and alive bool
	var processes []ProcessInfo
	var pgrepCommand string
	if processArgs == "" {
		//look for any instance of this process name
		pgrepCommand = "pgrep -x " + processName
	} else {
		//look for a process running with particular arguments
		pgrepCommand = "pgrep -f \"" + processName + " .*" + processArgs + "\""
	}
	log.Printf("Running pgrep %v\n", pgrepCommand)
	out, perr := exec.Command("sh", "-c", pgrepCommand).Output()
	if perr != nil {
		return processes
	}

	for _, pid := range strings.Split(string(out), "\n") {
		if pid == "" {
			continue
		}
		p, err := strconv.Atoi(pid)
		if err != nil {
			fmt.Printf("Error in finding pid from process: %v -- %v", processName, err)
		} else {
			pinfo := ProcessInfo{pid: p, alive: true}
			processes = append(processes, pinfo)
		}
	}
	return processes
}

func ConnectController(p *process.ControllerLocal, c chan ReturnCodeWithText) {
	log.Printf("attempt to connect to process %v at %v\n", p.Name, p.ApiAddr)
	api, err := p.ConnectAPI(10 * time.Second)
	if err != nil {
		c <- ReturnCodeWithText{false, "Failed to connect to " + p.Name}
	} else {
		c <- ReturnCodeWithText{true, "OK connect to " + p.Name}
		api.Close()
	}
}

//default is to connect to the first controller, unless we specified otherwise
func GetController(ctrlname string) *ControllerProcess {
	if ctrlname == "" {
		return &Deployment.Controllers[0]
	}
	for _, ctrl := range Deployment.Controllers {
		if ctrl.Name == ctrlname {
			return &ctrl
		}
	}
	log.Fatalf("Error: could not find specified controller: %v\n", ctrlname)
	return nil //unreachable
}

func GetDme(dmename string) *DmeProcess {
	if dmename == "" {
		return &Deployment.Dmes[0]
	}
	for _, dme := range Deployment.Dmes {
		if dme.Name == dmename {
			return &dme
		}
	}
	log.Fatalf("Error: could not find specified dme: %v\n", dmename)
	return nil //unreachable
}

func ConnectCrm(p *process.CrmLocal, c chan ReturnCodeWithText) {
	log.Printf("attempt to connect to process %v at %v\n", p.Name, p.ApiAddr)
	api, err := p.ConnectAPI(10 * time.Second)
	if err != nil {
		c <- ReturnCodeWithText{false, "Failed to connect to " + p.Name}
	} else {
		c <- ReturnCodeWithText{true, "OK connect to " + p.Name}
		api.Close()
	}
}

func ConnectDme(p *process.DmeLocal, c chan ReturnCodeWithText) {
	log.Printf("attempt to connect to process %v at %v\n", p.Name, p.ApiAddr)
	api, err := p.ConnectAPI(10 * time.Second)
	if err != nil {
		c <- ReturnCodeWithText{false, "Failed to connect to " + p.Name}
	} else {
		c <- ReturnCodeWithText{true, "OK connect to " + p.Name}
		api.Close()
	}
}

//first tries to kill process with SIGINT, then waits up to maxwait time
//for it to die.  After that point it kills with SIGKILL
func KillProcessesByName(processName string, maxwait time.Duration, processArgs string, c chan string) {
	processes := getPidsByName(processName, processArgs)
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
		//we run out of waiting time. Passing maxwait of zero duration means kill
		//forcefully no matter what, which we want in some disruptive tests
		if maxwait <= 0 {
			break
		}
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

		time.Sleep(waitInterval)
		maxwait -= waitInterval

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
	log.Printf("\n\n   *** %s\n", label)
}

func PrintStepBanner(label string) {
	log.Printf("\n\n      --- %s\n", label)
}

//for specific output that we want to put in a separate file.  If no
//output dir, just  print to the stdout
func PrintToFile(fname string, outputDir string, out string, truncate bool) {
	if outputDir == "" {
		fmt.Print(out)
	} else {
		outfile := outputDir + "/" + fname
		mode := os.O_APPEND
		if truncate {
			mode = os.O_TRUNC
		}
		ofile, err := os.OpenFile(outfile, mode|os.O_CREATE|os.O_WRONLY, 0666)
		defer ofile.Close()
		if err != nil {
			log.Fatalf("unable to append output file: %s, err: %v\n", outfile, err)
		}
		log.Printf("writing file: %s\n%s\n", fname, out)
		fmt.Fprintf(ofile, out)
	}
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

func ReadYamlFile(filename string, iface interface{}, varlist string, validateReplacedVars bool) error {
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
	if validateReplacedVars {
		//make sure there are no unreplaced variables left and inform the user if so
		re := regexp.MustCompile("{{(\\S+)}}")
		matches := re.FindAllStringSubmatch(string(yamlFile), 1)
		if len(matches) > 0 {
			return errors.New(fmt.Sprintf("Unreplaced variables in yaml: %v", matches))
		}
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
func CompareYamlFiles(firstYamlFile string, secondYamlFile string, fileType string) bool {

	PrintStepBanner("running compareYamlFiles")

	log.Printf("Comparing yamls: %v  %v\n", firstYamlFile, secondYamlFile)

	var err1 error
	var err2 error
	var y1 interface{}
	var y2 interface{}

	if fileType == "appdata" {
		//for appdata, use the ApplicationData type so we can sort it
		var a1 edgeproto.ApplicationData
		var a2 edgeproto.ApplicationData

		err1 = ReadYamlFile(firstYamlFile, &a1, "", false)
		err2 = ReadYamlFile(secondYamlFile, &a2, "", false)
		a1.Sort()
		a2.Sort()
		y1 = a1
		y2 = a2

	} else {
		err1 = ReadYamlFile(firstYamlFile, &y1, "", false)
		err2 = ReadYamlFile(secondYamlFile, &y2, "", false)
	}
	if err1 != nil {
		log.Printf("Error in reading yaml file %v -- %v\n", firstYamlFile, err1)
		return false
	}
	if err2 != nil {
		log.Printf("Error in reading yaml file %v -- %v\n", secondYamlFile, err2)
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
