package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	dmeproto "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"google.golang.org/grpc"

	yaml "gopkg.in/yaml.v2"
)

var (
	commandName = "setup-mex"
	actions     = flag.String("actions", "", "one or more of: "+actionList+" separated by ,")
	deployment  = flag.String("deployment", "process", deploymentList)
	appFile     = flag.String("appfile", "", "optional controller application data file")
	setupFile   = flag.String("setupfile", "", "mandatory yml topology file")
	outputDir   = flag.String("outputdir", "", "option directory to store output and logs")
	compareYaml = flag.String("compareyaml", "", "comma separated list of yamls to compare")
	merFile     = flag.String("merfile", "", "match engine request input file")
	dataDir     = flag.String("datadir", "", "optional path of data files")
	procs       processData
)

type yamlReplacementVariables struct {
	Vars []map[string]string
}

type returnCodeWithText struct {
	success bool
	text    string
}

type etcdProcess struct {
	process.EtcdLocal
	Hostname string
}
type controllerProcess struct {
	process.ControllerLocal
	Hostname string
}
type dmeProcess struct {
	process.DmeLocal
	Hostname string
}
type crmProcess struct {
	process.CrmLocal
	Hostname string
}
type locSimProcess struct {
	process.LocApiSimLocal
	Hostname string
}

type processData struct {
	Locsims     []locSimProcess     `yaml:"locsims"`
	Etcds       []etcdProcess       `yaml:"etcds"`
	Controllers []controllerProcess `yaml:"controllers"`
	Dmes        []dmeProcess        `yaml:"dmes"`
	Crms        []crmProcess        `yaml:"crms"`
}

//this is possible actions and optional parameters
var actionChoices = map[string]string{
	"start":          "procname",
	"stop":           "procname",
	"status":         "procname",
	"show":           "procname",
	"update":         "procname",
	"delete":         "procname",
	"create":         "procname",
	"deploy":         "",
	"cleanup":        "",
	"fetchlogs":      "",
	"findcloudlet":   "procname",
	"verifylocation": "procname",
}

//these are strings which may be present in the yaml but not in the corresponding data structures.
//These are the only allowed exceptions to the strict yaml unmarshalling
var yamlExceptions = map[string]map[string]bool{
	"setup": {
		"vars": true,
	},
	"data": {
		"ip_str": true, // ansible workaround
	},
}

var data edgeproto.ApplicationData

var matchEngineRequest dmeproto.Match_Engine_Request

var apiConnectTimeout = 5 * time.Second

var actionSlice = make([]string, 0)

var apiAddrsUpdated = false

var actionList = fmt.Sprintf("%v", reflect.ValueOf(actionChoices).MapKeys())

var deploymentChoices = map[string]bool{"process": true,
	"container": true}
var deploymentList = fmt.Sprintf("%v", reflect.ValueOf(deploymentChoices).MapKeys())

func printUsage() {
	fmt.Println("\nUsage: \n" + commandName + " [options]\n\noptions:")
	buf := new(bytes.Buffer)
	flag.CommandLine.SetOutput(buf)

	flag.PrintDefaults()
	opts := string(buf.Bytes())
	ss := strings.Split(opts, "\n")

	skipNextLine := false

	// skip the test options that get imported
	for _, line := range ss {
		if skipNextLine {
			skipNextLine = false
			continue
		}
		if strings.Contains(line, "test") {
			skipNextLine = true
			continue
		}
		fmt.Println(line)
	}
}

func isYamlOk(e error, yamltype string) bool {
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
			log.Printf("Fatal Unmarshal Error: %v\n", err1)
			rc = false
		}
	}
	return rc
}

// TODO, would be nice to figure how to do these with the same implementation
func connectController(p *process.ControllerLocal, c chan returnCodeWithText) {
	log.Printf("attempt to connect to process %v at %v\n", (*p).Name, (*p).ApiAddr)
	api, err := (*p).ConnectAPI(10 * time.Second)
	if err != nil {
		c <- returnCodeWithText{false, "Failed to connect to " + (*p).Name}
	} else {
		c <- returnCodeWithText{true, "OK connect to " + (*p).Name}
		api.Close()
	}
}

func connectCrm(p *process.CrmLocal, c chan returnCodeWithText) {
	log.Printf("attempt to connect to process %v at %v\n", (*p).Name, (*p).ApiAddr)
	api, err := (*p).ConnectAPI(10 * time.Second)
	if err != nil {
		c <- returnCodeWithText{false, "Failed to connect to " + (*p).Name}
	} else {
		c <- returnCodeWithText{true, "OK connect to " + (*p).Name}
		api.Close()
	}
}

func connectDme(p *process.DmeLocal, c chan returnCodeWithText) {
	log.Printf("attempt to connect to process %v at %v\n", (*p).Name, (*p).ApiAddr)
	api, err := (*p).ConnectAPI(10 * time.Second)
	if err != nil {
		c <- returnCodeWithText{false, "Failed to connect to " + (*p).Name}
	} else {
		c <- returnCodeWithText{true, "OK connect to " + (*p).Name}
		api.Close()
	}
}

func waitForProcesses(processName string) bool {
	log.Println("Wait for processes to respond to APIs")
	c := make(chan returnCodeWithText)
	var numProcs = 0 //len(procs.Controllers) + len(procs.Crms) + len(procs.Dmes)
	for _, ctrl := range procs.Controllers {
		if processName != "" && processName != ctrl.Name {
			continue
		}
		numProcs += 1
		ctrlp := ctrl.ControllerLocal
		go connectController(&ctrlp, c)
	}
	for _, dme := range procs.Dmes {
		if processName != "" && processName != dme.Name {
			continue
		}
		numProcs += 1
		dmep := dme.DmeLocal
		go connectDme(&dmep, c)
	}
	for _, crm := range procs.Crms {
		if processName != "" && processName != crm.Name {
			continue
		}
		numProcs += 1
		crmp := crm.CrmLocal
		go connectCrm(&crmp, c)
	}
	allpass := true
	for i := 0; i < numProcs; i++ {
		rc := <-c
		log.Println(rc.text)
		if !rc.success {
			allpass = false
		}
	}
	return allpass
}

//to identify running processes, we need to know where they are running
//and some unique arguments to look for for pgrep. Returns info about
// the hostname, process executable and arguments
func findProcess(processName string) (string, string, string) {
	for _, etcd := range procs.Etcds {
		if etcd.Name == processName {
			return etcd.Hostname, "etcd", "etcd .*-name " + etcd.Name
		}
	}
	for _, ctrl := range procs.Controllers {
		if ctrl.Name == processName {
			return ctrl.Hostname, "controller", "-apiAddr " + ctrl.ApiAddr
		}
	}
	for _, crm := range procs.Crms {
		if crm.Name == processName {

			return crm.Hostname, "crmserver", "-apiAddr " + crm.ApiAddr
		}
	}
	for _, dme := range procs.Dmes {
		if dme.Name == processName {
			return dme.Hostname, "dme-server", "-apiAddr " + dme.ApiAddr
		}
	}
	for _, loc := range procs.Locsims {
		if loc.Name == processName {
			return loc.Hostname, "loc-api-sim", fmt.Sprintf("port -%d", loc.Port)
		}
	}

	return "", "", ""
}

func getExternalApiAddress(internalApiAddr string, externalHost string) string {
	//in cloud deployments, the internal address the controller listens to may be different than the
	//external address which clients need to use.   So use the external hostname and api port
	return externalHost + ":" + strings.Split(internalApiAddr, ":")[1]
}

//in cloud deployments, the internal address the controller listens to may be different than the
//external address which clients need to use as floating IPs are used.  So use the external
//hostname and api port when connecting to the API.  This needs to be done after startup
//but before trying to connect to the APIs remotely
func updateApiAddrs() {
	if apiAddrsUpdated {
		//no need to do this more than once
		return
	}
	for i, ctrl := range procs.Controllers {
		procs.Controllers[i].ApiAddr = getExternalApiAddress(ctrl.ApiAddr, ctrl.Hostname)
	}
	for i, dme := range procs.Dmes {
		procs.Dmes[i].ApiAddr = getExternalApiAddress(dme.ApiAddr, dme.Hostname)
	}
	for i, crm := range procs.Crms {
		procs.Crms[i].ApiAddr = getExternalApiAddress(crm.ApiAddr, crm.Hostname)
	}
	apiAddrsUpdated = true
}

func printYaml(i interface{}) {
	out, err := yaml.Marshal(i)
	if err != nil {
		log.Fatalf("Error encoding yaml for %+v: %v", i, err)
	}
	s := string(out[:])
	log.Printf("YAML: %s %s\n", s, out)
}

//for specific output that we want to put in a separate file.  If no
//output dir, just  print to the stdout
func printToFile(fname string, out string, truncate bool) {
	if *outputDir == "" {
		fmt.Print(out)
	} else {
		outfile := *outputDir + "/" + fname
		mode := os.O_APPEND
		if truncate {
			mode = os.O_TRUNC
		}
		ofile, err := os.OpenFile(outfile, mode|os.O_CREATE|os.O_WRONLY, 0666)
		defer ofile.Close()
		if err != nil {
			log.Fatalf("unable to append output file: %s, err: %v\n", outfile, err)
		}
		defer ofile.Close()
		log.Printf("writing file: %s\n%s\n", fname, out)
		fmt.Fprintf(ofile, out)
	}
}

//default is to connect to the first controller, unless we specified otherwise
func getController(ctrlname string) *controllerProcess {
	if ctrlname == "" {
		return &procs.Controllers[0]
	}
	for _, ctrl := range procs.Controllers {
		if ctrl.Name == ctrlname {
			return &ctrl
		}
	}
	log.Fatalf("Error: could not find specified controller: %v\n", ctrlname)
	return nil //unreachable
}

func getDme(dmename string) *dmeProcess {
	if dmename == "" {
		return &procs.Dmes[0]
	}
	for _, dme := range procs.Dmes {
		if dme.Name == dmename {
			return &dme
		}
	}
	log.Fatalf("Error: could not find specified dme: %v\n", dmename)
	return nil //unreachable
}

func runShowCommands(ctrlname string) bool {
	errFound := false
	var showCmds = []string{
		"operators: ShowOperator",
		"developers: ShowDeveloper",
		"cloudlets: ShowCloudlet",
		"apps: ShowApp",
		"appinstances: ShowAppInst",
	}
	ctrl := getController(ctrlname)

	for i, c := range showCmds {
		label := strings.Split(c, " ")[0]
		cmdstr := strings.Split(c, " ")[1]
		cmd := exec.Command("edgectl", "--addr", ctrl.ApiAddr, "controller", cmdstr)
		log.Printf("generating output for %s\n", label)
		out, _ := cmd.CombinedOutput()
		truncate := false
		//truncate the file for the first command output, afterwards append
		if i == 0 {
			truncate = true
		}
		//edgectl returns exitcode 0 even if it cannot connect, so look for the error
		if strings.Contains(string(out), cmdstr+" failed") {
			log.Printf("Found failure in show output\n")
			errFound = true
		}
		printToFile("show-commands.yml", label+"\n"+string(out)+"\n", truncate)
	}
	return !errFound
}

func runApis(mode string, ctrlname string) bool {
	log.Printf("Applying data via APIs\n")
	ctrl := getController(ctrlname)
	log.Printf("Connecting to controller %v at address %v", ctrl.Name, ctrl.ApiAddr)
	ctrlapi, err := ctrl.ConnectAPI(apiConnectTimeout)

	if err != nil {
		log.Printf("Error connecting to controller api: %v\n", ctrl.ApiAddr)
		return false
	} else {
		log.Printf("Connected to controller %v success", ctrl.Name)
		ctx, cancel := context.WithTimeout(context.Background(), apiConnectTimeout)

		if mode == "delete" {
			//run in reverse order to delete child keys
			util.RunAppinstApi(ctrlapi, ctx, &data, mode)
			util.RunAppApi(ctrlapi, ctx, &data, mode)
			util.RunCloudletApi(ctrlapi, ctx, &data, mode)
			util.RunDeveloperApi(ctrlapi, ctx, &data, mode)
			util.RunOperatorApi(ctrlapi, ctx, &data, mode)
		} else {
			util.RunOperatorApi(ctrlapi, ctx, &data, mode)
			util.RunDeveloperApi(ctrlapi, ctx, &data, mode)
			util.RunCloudletApi(ctrlapi, ctx, &data, mode)
			util.RunAppApi(ctrlapi, ctx, &data, mode)
			util.RunAppinstApi(ctrlapi, ctx, &data, mode)
		}

		cancel()
	}
	ctrlapi.Close()
	return true
}

func runDmeApi(api string, procname string) bool {
	dme := getDme(procname)
	conn, err := grpc.Dial(dme.ApiAddr, grpc.WithInsecure())

	if err != nil {
		log.Printf("Error: unable to connect to dme addr %v\n", dme.ApiAddr)
		return false
	}
	defer conn.Close()
	client := dmeproto.NewMatch_Engine_ApiClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	//generic struct so we can do the marshal in one place even though return types are different
	var dmereply interface{}
	var dmeerror error

	switch api {
	case "findcloudlet":
		dmereply, dmeerror = client.FindCloudlet(ctx, &matchEngineRequest)
	case "verifylocation":
		dmereply, dmeerror = client.VerifyLocation(ctx, &matchEngineRequest)
	default:
		log.Printf("Unsupported dme api %s\n", api)
		return false
	}
	if dmeerror != nil {
		log.Printf("Error in find api %s -- %v\n", api, dmeerror)
		return false
	}
	out, ymlerror := yaml.Marshal(dmereply)
	if ymlerror != nil {
		fmt.Printf("Error: Unable to marshal %s reply: %v\n", api, ymlerror)
		return false
	}

	printToFile(api+".yml", string(out), true)
	return true
}

func getLogFile(procname string) string {
	if *outputDir == "" {
		return "./" + procname + ".log"
	} else {
		return *outputDir + "/" + procname + ".log"
	}
}

func readSetupFile(setupfile string) {
	//the setup file has a vars section with replacement variables.  ingest the file once
	//to get these variables, and then ingest again to parse the setup data with the variables
	var replacementVars yamlReplacementVariables
	var varlist []string
	varstring := ""

	if *dataDir != "" {
		varlist = append(varlist, "datadir="+*dataDir)
	}
	util.ReadYamlFile(setupfile, &replacementVars, "", false)
	for _, repl := range replacementVars.Vars {
		for varname, value := range repl {
			varlist = append(varlist, varname+"="+value)
		}
	}
	if len(varlist) > 0 {
		varstring = strings.Join(varlist, ",")
	}
	err := util.ReadYamlFile(setupfile, &procs, varstring, true)
	if err != nil {
		if !isYamlOk(err, "setup") {
			log.Fatal("One or more fatal unmarshal errors, exiting")
		}
	}
}
func readAppDataFile(datafile string) {
	err := util.ReadYamlFile(datafile, &data, "", true)
	if err != nil {
		if !isYamlOk(err, "data") {
			log.Fatal("One or more fatal unmarshal errors, exiting")
		}
	}
}
func readMerFile(merfile string) {
	err := util.ReadYamlFile(merfile, &matchEngineRequest, "", true)
	if err != nil {
		if !isYamlOk(err, "mer") {
			log.Fatal("One or more fatal unmarshal errors, exiting")
		}
	}
}

func stopProcesses(processName string) bool {
	util.PrintStepBanner("stopping processes " + processName)
	maxWait := time.Second * 15
	c := make(chan string)

	//if a process name is specified, we stop just that one.  The name here identifies the
	//specific instance of the application, e.g. 'ctrl1'
	if processName != "" {
		hostName, processExeName, processArgs := findProcess(processName)
		log.Printf("Found host %v processexe %v processargs %v\n", hostName, processExeName, processArgs)
		if hostName == "" {
			log.Printf("Error: unable to find process name %v in setup\n", processName)
			return false
		}
		if hostName == "localhost" || hostName == "127.0.0.1" {
			//passing zero wait time to kill forcefully
			go util.KillProcessesByName(processExeName, maxWait, processArgs, c)
			log.Printf("kill result: %v\n", <-c)
		}
		return true
	}
	for _, p := range procs.Etcds {
		log.Printf("cleaning etcd %+v", p)
		p.ResetData()
	}
	processExeNames := [5]string{"etcd", "controller", "dme-server", "crmserver", "loc-api-sim"}
	//anything not gracefully exited in maxwait seconds is forcefully killed

	for _, p := range processExeNames {
		log.Println("killing processes " + p)
		go util.KillProcessesByName(p, maxWait, "", c)
	}
	for i := 0; i < len(processExeNames); i++ {
		log.Printf("kill result: %v\n", <-c)
	}
	return true
}

func runPlaybook(playbook string, evars []string, procNamefilter string) bool {
	invFile, found := createAnsibleInventoryFile(procNamefilter)
	ansHome := os.Getenv("ANSIBLE_DIR")

	if !stageYamlFile("setup.yml", ansHome+"/playbooks", procs) {
		return false
	}

	if !found {
		log.Println("No remote servers found, local environment only")
		return true
	}

	argstr := ""
	for _, ev := range evars {
		argstr += ev
		argstr += " "
	}
	log.Printf("Running Playbook: %s with extra-vars: %s\n", playbook, argstr)
	cmd := exec.Command("ansible-playbook", "-i", invFile, "-e", argstr, playbook)

	output, err := cmd.CombinedOutput()
	log.Printf("Ansible Output:\n%v\n", string(output))

	if err != nil {
		log.Fatalf("Ansible playbook failed: %v", err)
		return false
	}
	return true
}

func stageYamlFile(filename string, directory string, contents interface{}) bool {

	dstFile := directory + "/" + filename

	//rather than just copy the file, we unmarshal it because we have done variable replace
	data, err := yaml.Marshal(contents)
	if err != nil {
		log.Printf("Error in marshal of setupfile for ansible %v\n", err)
		return false
	}

	log.Printf("writing setup data to %s\n", dstFile)

	// Write data to dst
	err = ioutil.WriteFile(dstFile, data, 0644)
	if err != nil {
		log.Printf("Error writing file: %v\n", err)
		return false
	}
	return true
}

func stageLocDbFile(srcFile string, destDir string) {
	var locdb interface{}
	yerr := util.ReadYamlFile(srcFile, &locdb, "", false)
	if yerr != nil {
		log.Fatalf("Error reading location file %s -- %v\n", srcFile, yerr)
	}
	if !stageYamlFile("locsim.yml", destDir, locdb) {
		log.Fatalf("Error staging location db file %s to %s\n", srcFile, destDir)
	}
}

func createAnsibleInventoryFile(procNameFilter string) (string, bool) {
	ansHome := os.Getenv("ANSIBLE_DIR")
	log.Printf("Creating inventory file in ANSIBLE_DIR:%s using procname filter: %s\n", ansHome, procNameFilter)

	foundServer := false
	if ansHome == "" {
		log.Fatal("Need to set ANSIBLE_DIR environment variable for deployment")
	}

	invfile, err := os.Create(ansHome + "/mex_inventory")
	log.Printf("Creating inventory file: %v", invfile.Name())
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer invfile.Close()

	allRemoteServers := make(map[string]string)
	etcdRemoteServers := make(map[string]string)
	ctrlRemoteServers := make(map[string]string)
	crmRemoteServers := make(map[string]string)
	dmeRemoteServers := make(map[string]string)
	locApiSimulators := make(map[string]string)

	for _, p := range procs.Etcds {
		if procNameFilter != "" && procNameFilter != p.Name {
			continue
		}
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			etcdRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range procs.Controllers {
		if procNameFilter != "" && procNameFilter != p.Name {
			continue
		}
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			ctrlRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range procs.Crms {
		if procNameFilter != "" && procNameFilter != p.Name {
			continue
		}
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			crmRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range procs.Dmes {
		if procNameFilter != "" && procNameFilter != p.Name {
			continue
		}
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			dmeRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range procs.Locsims {
		if procNameFilter != "" && procNameFilter != p.Name {
			continue
		}
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			locApiSimulators[p.Hostname] = p.Name
			foundServer = true

			if p.Locfile != "" {
				stageLocDbFile(p.Locfile, ansHome+"/playbooks")
			}
		}
	}

	//create ansible inventory
	fmt.Fprintln(invfile, "[mexservers]")
	for s := range allRemoteServers {
		fmt.Fprintln(invfile, s)
	}
	if len(etcdRemoteServers) > 0 {
		fmt.Fprintln(invfile, "")
		fmt.Fprintln(invfile, "[etcds]")
		for s := range etcdRemoteServers {
			fmt.Fprintln(invfile, s)
		}
	}
	if len(ctrlRemoteServers) > 0 {
		fmt.Fprintln(invfile, "")
		fmt.Fprintln(invfile, "[controllers]")
		for s := range ctrlRemoteServers {
			fmt.Fprintln(invfile, s)
		}
	}
	if len(crmRemoteServers) > 0 {
		fmt.Fprintln(invfile, "")
		fmt.Fprintln(invfile, "[crms]")
		for s := range crmRemoteServers {
			fmt.Fprintln(invfile, s)
		}
	}
	if len(dmeRemoteServers) > 0 {
		fmt.Fprintln(invfile, "")
		fmt.Fprintln(invfile, "[dmes]")
		for s := range dmeRemoteServers {
			fmt.Fprintln(invfile, s)
		}
	}
	if len(locApiSimulators) > 0 {
		fmt.Fprintln(invfile, "")
		fmt.Fprintln(invfile, "[locsims]")
		for s := range locApiSimulators {
			fmt.Fprintln(invfile, s)
		}
	}
	fmt.Fprintln(invfile, "")
	return invfile.Name(), foundServer
}

func deployProcesses() bool {
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_deploy.yml"
	return runPlaybook(playbook, []string{}, "")
}

func startRemoteProcesses(processName string) bool {
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_start.yml"

	return runPlaybook(playbook, []string{}, processName)
}

func stopRemoteProcesses(processName string) bool {
	ansHome := os.Getenv("ANSIBLE_DIR")

	if processName != "" {
		hostName, processExeName, processArgs := findProcess(processName)

		if hostName == "localHost" || hostName == "127.0.0.1" {
			log.Printf("process %v is not remote\n", processName)
			return true
		}
		vars := []string{"processbin=" + processExeName, "processargs=\"" + processArgs + "\""}
		playbook := ansHome + "/playbooks/mex_stop_matching_process.yml"
		return runPlaybook(playbook, vars, processName)

	}
	playbook := ansHome + "/playbooks/mex_stop.yml"
	return runPlaybook(playbook, []string{}, "")
}

func cleanupRemoteProcesses() bool {
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_cleanup.yml"
	return runPlaybook(playbook, []string{}, "")
}

func fetchRemoteLogs() bool {
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_fetch_logs.yml"
	return runPlaybook(playbook, []string{"local_log_path=" + *outputDir}, "")
}

func startProcesses(processName string) bool {
	util.PrintStepBanner("starting local processes")
	for _, etcd := range procs.Etcds {
		if processName != "" && processName != etcd.Name {
			continue
		}
		if etcd.Hostname == "localhost" || etcd.Hostname == "127.0.0.1" {
			log.Printf("Starting Etcd %+v", etcd)
			etcd.ResetData()
			logfile := getLogFile(etcd.Name)
			err := etcd.Start(logfile)
			if err != nil {
				log.Printf("Error on Etcd startup: %v", err)
				return false
			}
		}
	}
	for _, ctrl := range procs.Controllers {
		if processName != "" && processName != ctrl.Name {
			continue
		}
		if ctrl.Hostname == "localhost" || ctrl.Hostname == "127.0.0.1" {
			log.Printf("Starting Controller %+v\n", ctrl)
			logfile := getLogFile(ctrl.Name)
			err := ctrl.Start(logfile)
			if err != nil {
				log.Printf("Error on controller startup: %v", err)
				return false
			}
		}
	}
	for _, crm := range procs.Crms {
		if processName != "" && processName != crm.Name {
			continue
		}
		if crm.Hostname == "localhost" || crm.Hostname == "127.0.0.1" {
			log.Printf("Starting CRM %+v\n", crm)
			logfile := getLogFile(crm.Name)
			err := crm.Start(logfile)
			if err != nil {
				log.Printf("Error on CRM startup: %v", err)
				return false
			}
		}
	}
	for _, dme := range procs.Dmes {
		if processName != "" && processName != dme.Name {
			continue
		}
		if dme.Hostname == "localhost" || dme.Hostname == "127.0.0.1" {
			log.Printf("Starting DME %+v\n", dme)
			logfile := getLogFile(dme.Name)
			err := dme.Start(logfile)
			if err != nil {
				log.Printf("Error on DME startup: %v", err)
				return false
			}
		}
	}
	for _, loc := range procs.Locsims {
		if processName != "" && processName != loc.Name {
			continue
		}
		if loc.Hostname == "localhost" || loc.Hostname == "127.0.0.1" {
			log.Printf("Starting LocSim %+v\n", loc)
			if loc.Locfile != "" {
				stageLocDbFile(loc.Locfile, "/var/tmp")
			}
			logfile := getLogFile(loc.Name)
			err := loc.Start(logfile)
			if err != nil {
				log.Printf("Error on LocSim startup: %v", err)
				return false
			}

		}
	}
	return true
}

//some actions have sub arguments associated after equal sign, e.g.--actions stop=ctrl1
func getActionParam(a string) (string, string) {
	argslice := strings.Split(a, "=")
	action := argslice[0]
	actionParam := ""
	if len(argslice) > 1 {
		actionParam = argslice[1]
	}
	return action, actionParam
}

func validateArgs() {
	flag.Parse()
	_, validDeployment := deploymentChoices[*deployment]
	errFound := false

	if *actions != "" {
		actionSlice = strings.Split(*actions, ",")
	}
	for _, a := range actionSlice {
		action, actionParam := getActionParam(a)

		optionalParam, validAction := actionChoices[action]
		if !validAction {
			fmt.Printf("ERROR: -actions must be one of: %v, received: %s\n", actionList, action)
			errFound = true
		} else if (action == "update" || action == "create" || action == "delete") && *appFile == "" {
			fmt.Printf("ERROR: if action=update, create or delete -appfile must be specified\n")
			errFound = true
		} else if action == "fetchlogs" && *outputDir == "" {
			fmt.Printf("ERROR: cannot use action=fetchlogs option without -outputdir\n")
			errFound = true
		} else if (action == "findcloudlet" || action == "verifylocation") && *merFile == "" {
			fmt.Printf("ERROR: cannot use action=findcloudlet or action=verifylocation option without -mer\n")
			errFound = true
		}
		if optionalParam == "" && actionParam != "" {
			fmt.Printf("ERROR: action %v does not take a parameter\n", action)
			errFound = true
		}
	}

	if !validDeployment {
		fmt.Printf("ERROR: -deployment must be one of: %v\n", deploymentList)
		errFound = true
	}
	if *appFile != "" {
		if _, err := os.Stat(*appFile); err != nil {
			fmt.Printf("ERROR: file " + *appFile + " does not exist\n")
			errFound = true
		}
	}
	if *merFile != "" {
		if _, err := os.Stat(*merFile); err != nil {
			fmt.Printf("ERROR: file " + *merFile + " does not exist\n")
			errFound = true
		}
	}
	if *setupFile == "" {
		fmt.Printf("ERROR -setupfile is mandatory\n")
		errFound = true
	} else {
		if _, err := os.Stat(*setupFile); err != nil {
			fmt.Printf("ERROR: file " + *setupFile + " does not exist\n")
			errFound = true
		}
	}

	if *compareYaml != "" {
		yarray := strings.Split(*compareYaml, ",")
		if len(yarray) != 3 {
			fmt.Printf("ERROR: compareyaml must be a string with 2 yaml files and a filetype separated by comma\n")
			errFound = true
		}
	}
	if errFound {
		printUsage()
		os.Exit(1)
	}
}

func main() {
	validateArgs()

	errorsFound := 0
	if *outputDir != "" {
		*outputDir = util.CreateOutputDir(false, *outputDir, commandName+".log")
	}
	if *appFile != "" {
		readAppDataFile(*appFile)
	}
	if *setupFile != "" {
		readSetupFile(*setupFile)
	}
	if *merFile != "" {
		readMerFile(*merFile)
	}

	for _, a := range actionSlice {
		action, actionParam := getActionParam(a)

		util.PrintStepBanner("running action: " + a)
		switch action {
		case "deploy":
			if !deployProcesses() {
				errorsFound += 1
			}
		case "start":
			startFailed := false
			if !startProcesses(actionParam) {
				startFailed = true
				errorsFound += 1
			} else {
				if !startRemoteProcesses(actionParam) {
					startFailed = true
					errorsFound += 1
				}
			}
			if startFailed {
				stopProcesses(actionParam)
				stopRemoteProcesses(actionParam)
				errorsFound += 1
				break

			}
			updateApiAddrs()
			if !waitForProcesses(actionParam) {
				errorsFound += 1
			}
		case "status":
			updateApiAddrs()
			if !waitForProcesses(actionParam) {
				errorsFound += 1
			}
		case "stop":
			stopProcesses(actionParam)
			if !stopRemoteProcesses(actionParam) {
				errorsFound += 1
			}
		case "create":
			fallthrough
		case "update":
			fallthrough
		case "delete":
			updateApiAddrs()
			if !runApis(action, actionParam) {
				log.Printf("Unable to run apis for %s. Check connectivity to controller API\n", action)
				errorsFound += 1
			}
		case "show":
			updateApiAddrs()
			if !runShowCommands(actionParam) {
				errorsFound += 1
			}
		case "cleanup":
			if !cleanupRemoteProcesses() {
				errorsFound += 1
			}
		case "fetchlogs":
			if !fetchRemoteLogs() {
				errorsFound += 1
			}
		case "findcloudlet":
			fallthrough
		case "verifylocation":
			updateApiAddrs()
			if !runDmeApi(action, actionParam) {
				errorsFound += 1
			}
		default:
			log.Fatal("unexpected action: " + action)
		}
	}
	if *compareYaml != "" {
		//separate the arg into two files and then replace variables
		//	yamlType := strings.Split(*compareYaml, ",")[0]
		s := strings.Split(*compareYaml, ",")
		firstYamlFile := s[0]
		secondYamlFile := s[1]
		fileType := s[2]

		if !util.CompareYamlFiles(firstYamlFile, secondYamlFile, fileType) {
			errorsFound += 1
		}

	}
	if *outputDir != "" {
		fmt.Printf("\nNum Errors found: %d, Results in: %s\n", errorsFound, *outputDir)
		os.Exit(errorsFound)
	}
}
