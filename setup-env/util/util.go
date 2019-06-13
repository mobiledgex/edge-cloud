package util

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dmeproto "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/testutil"
	yaml "github.com/mobiledgex/yaml/v2"
	"google.golang.org/grpc"
)

var Deployment DeploymentData
var ApiAddrNone = "NONE"

type yamlFileType int

const (
	YamlAppdata yamlFileType = 0
	YamlOther   yamlFileType = 1
)

type SetupVariables struct {
	Vars     []map[string]string
	Includes []string
}

// replacement variables taken from the setup
var DeploymentReplacementVars map[string]string

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

type ClusterInfo struct {
	MexManifest string
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
	CopyFiles   []K8CopyFile
}

type TLSCertInfo struct {
	CommonName string
	IPs        []string
	DNSNames   []string
}

type DeploymentData struct {
	TLSCerts    []*TLSCertInfo        `yaml:"tlscerts"`
	Locsims     []*process.LocApiSim  `yaml:"locsims"`
	Toksims     []*process.TokSrvSim  `yaml:"toksims"`
	Vaults      []*process.Vault      `yaml:"vaults"`
	Etcds       []*process.Etcd       `yaml:"etcds"`
	Controllers []*process.Controller `yaml:"controllers"`
	Dmes        []*process.Dme        `yaml:"dmes"`
	SampleApps  []*process.SampleApp  `yaml:"sampleapps"`
	Influxs     []*process.Influx     `yaml:"influxs"`
	ClusterSvcs []*process.ClusterSvc `yaml:"clustersvcs"`
}

type errorReply struct {
	Code    int
	Message string
	Details []string
}

func GetAllProcesses() []process.Process {
	all := make([]process.Process, 0)
	for _, p := range Deployment.Locsims {
		all = append(all, p)
	}
	for _, p := range Deployment.Toksims {
		all = append(all, p)
	}
	for _, p := range Deployment.Vaults {
		all = append(all, p)
	}
	for _, p := range Deployment.Etcds {
		all = append(all, p)
	}
	for _, p := range Deployment.Controllers {
		all = append(all, p)
	}
	for _, p := range Deployment.Dmes {
		all = append(all, p)
	}
	for _, p := range Deployment.SampleApps {
		all = append(all, p)
	}
	for _, p := range Deployment.Influxs {
		all = append(all, p)
	}
	for _, p := range Deployment.ClusterSvcs {
		all = append(all, p)
	}
	return all
}

func GetProcessByName(processName string) process.Process {
	for _, p := range GetAllProcesses() {
		if processName == p.GetName() {
			return p
		}
	}
	return nil
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
		pgrepCommand = "pgrep -f \"" + processName + " .*" + processArgs + ".*\""
	}
	log.Printf("Running pgrep %v\n", pgrepCommand)
	out, perr := exec.Command("sh", "-c", pgrepCommand).Output()
	if perr != nil {
		log.Printf("Process not found for: %s\n", pgrepCommand)
		pinfo := ProcessInfo{alive: false}
		processes = append(processes, pinfo)
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

func ConnectController(p *process.Controller, c chan ReturnCodeWithText) {
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
func GetController(ctrlname string) *process.Controller {
	if ctrlname == "" {
		return Deployment.Controllers[0]
	}
	for _, ctrl := range Deployment.Controllers {
		if ctrl.Name == ctrlname {
			return ctrl
		}
	}
	log.Fatalf("Error: could not find specified controller: %v\n", ctrlname)
	return nil //unreachable
}

func GetDme(dmename string) *process.Dme {
	if dmename == "" {
		return Deployment.Dmes[0]
	}
	for _, dme := range Deployment.Dmes {
		if dme.Name == dmename {
			return dme
		}
	}
	log.Fatalf("Error: could not find specified dme: %v\n", dmename)
	return nil //unreachable
}

func ConnectCrm(p *process.Crm, c chan ReturnCodeWithText) {
	log.Printf("attempt to connect to process %v at %v\n", p.Name, p.ApiAddr)
	if p.ApiAddr == ApiAddrNone {
		c <- ReturnCodeWithText{true, "skipped nonexistent addr " + p.Name}
	}
	api, err := p.ConnectAPI(10 * time.Second)
	if err != nil {
		c <- ReturnCodeWithText{false, "Failed to connect to " + p.Name}
		return
	}
	api.Close()
	// check that controller sees crm online (has received cloudletinfo),
	// which is required before create clusterinst/appinst will work.
	if len(Deployment.Controllers) > 0 {
		err = checkCloudletState(p, 10*time.Second)
		if err != nil {
			c <- ReturnCodeWithText{false, "Ok connect to " + p.Name + " but " + err.Error()}
		} else {
			c <- ReturnCodeWithText{true, "OK connect to " + p.Name + " with CloudletInfo"}
		}
	} else {
		// this is a CRM only test
		c <- ReturnCodeWithText{true, "OK connect to " + p.Name}
	}
}

func ConnectDme(p *process.Dme, c chan ReturnCodeWithText) {
	log.Printf("attempt to connect to process %v at %v\n", p.Name, p.ApiAddr)
	api, err := p.ConnectAPI(10 * time.Second)
	if err != nil {
		c <- ReturnCodeWithText{false, "Failed to connect to " + p.Name}
	} else {
		c <- ReturnCodeWithText{true, "OK connect to " + p.Name}
		api.Close()
	}
}

func checkCloudletState(p *process.Crm, timeout time.Duration) error {
	filter := edgeproto.CloudletInfo{}
	err := json.Unmarshal([]byte(p.CloudletKey), &filter.Key)
	if err != nil {
		return fmt.Errorf("unable to parse CloudletKey")
	}

	conn := connectOnlineController(timeout)
	if conn == nil {
		return fmt.Errorf("unable to connect to online controller")
	}

	infoapi := edgeproto.NewCloudletInfoApiClient(conn)
	show := testutil.ShowCloudletInfo{}
	startTimeMs := time.Now().UnixNano() / int64(time.Millisecond)
	maxTimeMs := int64(timeout/time.Millisecond) + startTimeMs
	wait := 20 * time.Millisecond
	err = fmt.Errorf("unable to check CloudletInfo")
	for {
		timeout -= wait
		time.Sleep(wait)
		currTimeMs := time.Now().UnixNano() / int64(time.Millisecond)
		if currTimeMs > maxTimeMs {
			err = fmt.Errorf("timed out, last error was %s", err.Error())
			break
		}
		show.Init()
		stream, showErr := infoapi.ShowCloudletInfo(context.Background(), &filter)
		show.ReadStream(stream, showErr)
		if showErr != nil {
			err = fmt.Errorf("show CloudletInfo failed: %s", showErr.Error())
			continue
		}
		info, found := show.Data[filter.Key.GetKeyString()]
		if !found {
			err = fmt.Errorf("CloudletInfo not found")
			continue
		}
		if info.State != edgeproto.CloudletState_CLOUDLET_STATE_READY && info.State != edgeproto.CloudletState_CLOUDLET_STATE_ERRORS {
			err = fmt.Errorf("CloudletInfo bad state %s", edgeproto.CloudletState_name[int32(info.State)])
			continue
		}
		err = nil
		break
	}
	return err
}

func connectOnlineController(delay time.Duration) *grpc.ClientConn {
	for _, ctrl := range Deployment.Controllers {
		conn, err := ctrl.ConnectAPI(delay)
		if err == nil {
			return conn
		}
	}
	return nil
}

func StopProcess(p process.Process, maxwait time.Duration, c chan string) {
	// first attempt graceful stop
	p.StopLocal()
	// make sure process is dead or kill it
	KillProcessesByName(p.GetExeName(), maxwait, p.LookupArgs(), c)
}

//first tries to kill process with SIGINT, then waits up to maxwait time
//for it to die.  After that point it kills with SIGKILL
func KillProcessesByName(processName string, maxwait time.Duration, processArgs string, c chan string) {
	processes := getPidsByName(processName, processArgs)
	waitInterval := 100 * time.Millisecond

	for _, p := range processes {
		if !p.alive {
			continue
		}
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

func EnsureProcessesByName(processName string, processArgs string) bool {
	processes := getPidsByName(processName, processArgs)
	ensured := true
	for _, p := range processes {
		if !p.alive {
			log.Printf("Process not alive: %s args %s\n", processName, processArgs)
			ensured = false
		}
	}
	return ensured
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

func PrintToYamlFile(fname, outputDir string, data interface{}, truncate bool) {
	out, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalf("yaml marshal data failed, %v, %+v\n", err, data)
	}
	PrintToFile(fname, outputDir, string(out), truncate)
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

type ReadYamlOptions struct {
	vars                 map[string]string
	validateReplacedVars bool
}

type ReadYamlOp func(opts *ReadYamlOptions)

func ReadYamlFile(filename string, iface interface{}, ops ...ReadYamlOp) error {
	opts := ReadYamlOptions{}
	for _, op := range ops {
		op(&opts)
	}

	if strings.HasPrefix(filename, "~") {
		filename = strings.Replace(filename, "~", os.Getenv("HOME"), 1)
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.New(fmt.Sprintf("error reading yaml file: %v err: %v\n", filename, err))
	}
	if opts.vars != nil {
		//replace variables denoted as {{variablename}}
		yamlstr := string(yamlFile)
		for k, v := range opts.vars {
			yamlstr = strings.Replace(yamlstr, "{{"+k+"}}", v, -1)
		}
		yamlFile = []byte(yamlstr)
	}
	if opts.validateReplacedVars {
		//make sure there are no unreplaced variables left and inform the user if so
		re := regexp.MustCompile("{{(\\S+)}}")
		matches := re.FindAllStringSubmatch(string(yamlFile), 1)
		if len(matches) > 0 {
			return errors.New(fmt.Sprintf("Unreplaced variables in yaml: %v", matches))
		}
	}

	err = yaml.Unmarshal(yamlFile, iface)
	if err != nil {
		return err
	}

	return nil
}

func WithVars(vars map[string]string) ReadYamlOp {
	return func(opts *ReadYamlOptions) {
		opts.vars = vars
	}
}

func ValidateReplacedVars() ReadYamlOp {
	return func(opts *ReadYamlOptions) {
		opts.validateReplacedVars = true
	}
}

func removeAppinstUris(appdata *edgeproto.ApplicationData) {
	for i, ai := range appdata.AppInstances {
		if ai.Key.ClusterInstKey.CloudletKey != cloudcommon.DefaultCloudletKey {
			appdata.AppInstances[i].Uri = ""
		}
	}

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
	copts := []cmp.Option{}

	if fileType == "appdata" {
		//for appdata, use the ApplicationData type so we can sort it
		var a1 edgeproto.ApplicationData
		var a2 edgeproto.ApplicationData

		err1 = ReadYamlFile(firstYamlFile, &a1)
		err2 = ReadYamlFile(secondYamlFile, &a2)
		a1.Sort()
		a2.Sort()
		// Appinstance URIs usually not provisioned, as they are inherited from the cloudlet. However
		// they are provioned for the default appinst.  So we cannot use "nocmp".  Loop through and remove the URIs
		// for non-defaultCloudlets so that the comparison will succeeed
		removeAppinstUris(&a1)
		removeAppinstUris(&a2)
		y1 = a1
		y2 = a2
	} else if fileType == "findcloudlet" {
		var f1 dmeproto.FindCloudletReply
		var f2 dmeproto.FindCloudletReply

		err1 = ReadYamlFile(firstYamlFile, &f1)
		err2 = ReadYamlFile(secondYamlFile, &f2)

		//publicport is variable so we nil it out for comparison purposes.
		for _, p := range f1.Ports {
			p.PublicPort = 0
		}
		for _, p := range f2.Ports {
			p.PublicPort = 0
		}
		y1 = f1
		y2 = f2
	} else {
		err1 = ReadYamlFile(firstYamlFile, &y1)
		err2 = ReadYamlFile(secondYamlFile, &y2)
	}
	if err1 != nil {
		log.Printf("Error in reading yaml file %v -- %v\n", firstYamlFile, err1)
		return false
	}
	if err2 != nil {
		log.Printf("Error in reading yaml file %v -- %v\n", secondYamlFile, err2)
		return false
	}

	if !cmp.Equal(y1, y2, copts...) {
		log.Println("Comparison fail")
		log.Printf(cmp.Diff(y1, y2, copts...))
		return false
	}
	log.Println("Comparison success")
	return true
}

func ControllerCLI(ctrl *process.Controller, args ...string) ([]byte, error) {
	cmdargs := []string{"--addr", ctrl.ApiAddr, "controller"}
	if ctrl.TLS.ClientCert != "" {
		cmdargs = append(cmdargs, "--tls", ctrl.TLS.ClientCert)
	}
	cmdargs = append(cmdargs, args...)
	cmd := exec.Command("edgectl", cmdargs...)
	return cmd.CombinedOutput()
}

func CallRESTPost(httpAddr string, client *http.Client, pb proto.Message, reply proto.Message) error {
	str, err := new(jsonpb.Marshaler).MarshalToString(pb)
	if err != nil {
		log.Printf("Could not marshal request\n")
		return err
	}
	bytesRep := []byte(str)
	req, err := http.NewRequest("POST", httpAddr, bytes.NewBuffer(bytesRep))
	if err != nil {
		log.Printf("Failed to create a request\n")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to HTTP <%s>\n", httpAddr)
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	reader := bytes.NewReader(body)
	err = jsonpb.Unmarshal(reader, reply)

	if err != nil {
		log.Printf("Failed to unmarshal reply : %s %v\n", body, err)
		//try to unmarshal it as an error yaml reply
		var ereply errorReply
		err2 := json.Unmarshal(body, &ereply)
		if err2 == nil {
			log.Printf("Reply is an error response, message: %+v\n", ereply.Message)
			return fmt.Errorf("Error reply message: %s", ereply.Message)
		}
		// not an error reply either
		log.Printf("Failed to unmarshal as an error reply : %s %v\n", body, err2)

		// return the original error
		return err
	}
	return nil
}
