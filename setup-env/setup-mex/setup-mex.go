package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	yaml "gopkg.in/yaml.v2"
)

var (
	commandName  = "setup-mex"
	actions      = flag.String("actions", "", "one or more of: "+actionList+" separated by ,")
	deployment   = flag.String("deployment", "process", deploymentList)
	dataFile     = flag.String("datafile", "", "optional yml data file")
	setupFile    = flag.String("setupfile", "", "mandatory yml topology file")
	outputDir    = flag.String("outputdir", "", "option directory to store output and logs, TS suffix will be replaced with timestamp")
	useTimestamp = flag.Bool("timestamp", false, "append current timestamp to outputdir")

	procs processData
)

type applicationData struct {
	Operators    []edgeproto.Operator  `yaml:"operators"`
	Cloudlets    []edgeproto.Cloudlet  `yaml:"cloudlets"`
	Developers   []edgeproto.Developer `yaml:"developers"`
	Applications []edgeproto.App       `yaml:"apps"`
	AppInstances []edgeproto.AppInst   `yaml:"appinstances"`
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

type processData struct {
	Etcds       []etcdProcess       `yaml:"etcds"`
	Controllers []controllerProcess `yaml:"controllers"`
	Dmes        []dmeProcess        `yaml:"dmes"`
	Crms        []crmProcess        `yaml:"crms"`
}

var actionChoices = map[string]bool{
	"start":     true,
	"stop":      true,
	"status":    true,
	"show":      true,
	"update":    true,
	"delete":    true,
	"create":    true,
	"deploy":    true,
	"cleanup":   true,
	"fetchlogs": true,
}

//these are strings which may be present in the yaml but not in the corresponding data structures.
//These are the only allowed exceptions to the strict yaml unmarshalling
var yamlExceptions = map[string]map[string]bool{
	"setup": {},
	"data": {
		"ip_str": true, // ansible workaround
	},
}

var data applicationData

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
		if allowedException {
			log.Printf("notice: unmarshal: %v ignored\n", err1)
		} else if strings.Contains(err1, "yaml: unmarshal errors") {
			// ignore this summary error
		} else {
			//all other errors are unexpected and mean something is wrong in the yaml
			log.Printf("Fatal Unmarshal Error: %v\n", err1)
			rc = false
		}
	}
	return rc
}

func readDataFile(datafile string) {
	yamlFile, err := ioutil.ReadFile(datafile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		os.Exit(1)
	}
	err = yaml.UnmarshalStrict(yamlFile, &data)
	if err != nil {
		if !isYamlOk(err, "data") {
			log.Fatal("One or more fatal unmarshal errors, exiting")
		}
	}
}

// TODO, would be nice to figure how to do these with the same implementation
func connectController(p *process.ControllerLocal, c chan string) {
	log.Printf("attempt to connect to process %v at %v\n", (*p).Name, (*p).ApiAddr)
	api, err := (*p).ConnectAPI(10 * time.Second)
	if err != nil {
		c <- "Failed to connect to " + (*p).Name
	} else {
		c <- "OK connect to " + (*p).Name
		api.Close()
	}
}

func connectCrm(p *process.CrmLocal, c chan string) {
	log.Printf("attempt to connect to process %v at %v\n", (*p).Name, (*p).ApiAddr)
	api, err := (*p).ConnectAPI(10 * time.Second)
	if err != nil {
		c <- "Failed to connect to " + (*p).Name
	} else {
		c <- "OK connect to " + (*p).Name
		api.Close()
	}
}

func connectDme(p *process.DmeLocal, c chan string) {
	log.Printf("attempt to connect to process %v at %v\n", (*p).Name, (*p).ApiAddr)
	api, err := (*p).ConnectAPI(10 * time.Second)
	if err != nil {
		c <- "Failed to connect to " + (*p).Name
	} else {
		c <- "OK connect to " + (*p).Name
		api.Close()
	}
}

func waitForProcesses() {
	log.Println("Wait for processes to respond to APIs")
	c := make(chan string)
	var numProcs = 0 //len(procs.Controllers) + len(procs.Crms) + len(procs.Dmes)
	for _, ctrl := range procs.Controllers {
		numProcs += 1
		ctrlp := ctrl.ControllerLocal
		go connectController(&ctrlp, c)
	}
	for _, dme := range procs.Dmes {
		numProcs += 1
		dmep := dme.DmeLocal
		go connectDme(&dmep, c)
	}
	for _, crm := range procs.Crms {
		numProcs += 1
		crmp := crm.CrmLocal
		go connectCrm(&crmp, c)
	}
	for i := 0; i < numProcs; i++ {
		log.Println(<-c)
	}
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

//creates an output directory with an optional timestamp.  Server log files, output from APIs, and
//output from the script itself will all go there if specified
func createOutputDir() {
	if *useTimestamp {
		startTimestamp := time.Now().Format("2006-01-02T15:04:05")
		*outputDir = *outputDir + "/" + startTimestamp
	}
	fmt.Printf("Creating output dir: %s\n", *outputDir)
	err := os.MkdirAll(*outputDir, 0755)
	if err != nil {
		log.Fatalf("Error trying to create directory %v: %v\n", *outputDir, err)
	}

	logName := *outputDir + "/" + commandName + ".log"
	logFile, err := os.OpenFile(logName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		log.Fatalf("Error creating logfile %s\n", logName)
	}
	//log to both stdout and logfile
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

//for specific output that we want to put in a separate file.  If no
//output dir, just  print to the stdout
func printToFile(fname string, out string) {
	if *outputDir == "" {
		fmt.Print(out)
	} else {
		outfile := *outputDir + "/" + fname
		ofile, err := os.OpenFile(outfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		defer ofile.Close()
		if err != nil {
			log.Fatalf("unable to append output file: %s, err: %v\n", outfile, err)
		}
		defer ofile.Close()
		fmt.Fprintf(ofile, out)

	}
}

func runShowCommands() {
	var showCmds = map[string]string{
		"operators:":    "ShowOperator",
		"developers:":   "ShowDeveloper",
		"cloudlets:":    "ShowCloudlet",
		"apps:":         "ShowApp",
		"appinstances:": "ShowAppInst",
	}
	for label, cmd := range showCmds {
		cmd := exec.Command("edgectl", "--addr", procs.Controllers[0].ApiAddr, "controller", cmd)
		log.Printf("generating output for %s\n", label)
		out, _ := cmd.CombinedOutput()
		printToFile("show-commands.yml", label+"\n"+string(out)+"\n")
	}
}

func runApis(mode string) bool {
	log.Printf("Applying data via APIs\n")

	ctrl := procs.Controllers[0]
	log.Printf("Connecting to controller %v at address %v", ctrl.Name, ctrl.ApiAddr)
	ctrlapi, err := ctrl.ConnectAPI(apiConnectTimeout)

	if err != nil {
		log.Printf("Error connecting to controller api: %v\n", ctrl.ApiAddr)
		return false
	} else {
		log.Printf("Connected to controller %v success", ctrl.Name)
		ctx, cancel := context.WithTimeout(context.Background(), apiConnectTimeout)
		opAPI := edgeproto.NewOperatorApiClient(ctrlapi)
		for _, o := range data.Operators {
			log.Printf("API %v for operator: %v", mode, o.Key.Name)
			switch mode {
			case "create":
				_, err = opAPI.CreateOperator(ctx, &o)
			case "update":
				_, err = opAPI.UpdateOperator(ctx, &o)
			case "delete":
				_, err = opAPI.DeleteOperator(ctx, &o)
			}
		}
		devApi := edgeproto.NewDeveloperApiClient(ctrlapi)
		for _, d := range data.Developers {
			log.Printf("API %v for developer: %v", mode, d.Key.Name)
			switch mode {
			case "create":
				_, err = devApi.CreateDeveloper(ctx, &d)
			case "update":
				_, err = devApi.UpdateDeveloper(ctx, &d)
			case "delete":
				_, err = devApi.DeleteDeveloper(ctx, &d)
			}
		}
		clAPI := edgeproto.NewCloudletApiClient(ctrlapi)
		for _, c := range data.Cloudlets {
			log.Printf("API %v for cloudlet: %v", mode, c.Key.Name)
			switch mode {
			case "create":
				_, err = clAPI.CreateCloudlet(ctx, &c)
			case "update":
				_, err = clAPI.UpdateCloudlet(ctx, &c)
			case "delete":
				_, err = clAPI.DeleteCloudlet(ctx, &c)
			}
		}
		appAPI := edgeproto.NewAppApiClient(ctrlapi)
		for _, a := range data.Applications {
			log.Printf("API %v for app: %v", mode, a.Key.Name)
			switch mode {
			case "create":
				_, err = appAPI.CreateApp(ctx, &a)
			case "update":
				_, err = appAPI.UpdateApp(ctx, &a)
			case "delete":
				_, err = appAPI.DeleteApp(ctx, &a)
			}
		}
		appinAPI := edgeproto.NewAppInstApiClient(ctrlapi)
		for _, a := range data.AppInstances {
			log.Printf("API %v for appinstance: %v", mode, a.Key.AppKey.Name)
			switch mode {
			case "create":
				_, err = appinAPI.CreateAppInst(ctx, &a)
			case "update":
				_, err = appinAPI.UpdateAppInst(ctx, &a)
			case "delete":
				_, err = appinAPI.DeleteAppInst(ctx, &a)
			}
		}
		cancel()
	}
	ctrlapi.Close()
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
	yamlFile, err := ioutil.ReadFile(setupfile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		os.Exit(1)
	}
	err = yaml.UnmarshalStrict(yamlFile, &procs)
	if err != nil {
		if !isYamlOk(err, "setup") {
			log.Fatal("One or more fatal unmarshal errors, exiting")
		}
	}
}

func stopProcesses() {
	for _, p := range procs.Etcds {
		log.Printf("cleaning etcd %+v", p)
		p.ResetData()
	}
	log.Println("killing processes")
	exec.Command("sh", "-c", "pkill -SIGINT etcd").Output()
	exec.Command("sh", "-c", "pkill -SIGINT controller").Output()
	exec.Command("sh", "-c", "pkill -SIGINT crmserver").Output()
	exec.Command("sh", "-c", "pkill -SIGINT dme-server").Output()
}

func runPlaybook(playbook string, evars []string) bool {
	invFile, found := createAnsibleInventoryFile()
	copySetupFileToAnsible()

	if !found {
		log.Println("Notice: No remote servers found")
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

func copySetupFileToAnsible() {
	ansHome := os.Getenv("ANSIBLE_DIR")
	if ansHome == "" {
		log.Fatal("Need to set ANSIBLE_DIR environment variable for deployment")
	}
	dstFile := ansHome + "/playbooks/setup.yml"

	log.Printf("copying %s to %s\n", *setupFile, dstFile)
	// Read all content of src to data
	data, err := ioutil.ReadFile(*setupFile)
	if err != nil {
		log.Fatal("Error reading source process file for copy")
	}
	// Write data to dst
	err = ioutil.WriteFile(dstFile, data, 0644)
	if err != nil {
		log.Fatal("Error reading destination process file for copy")
	}
}

func createAnsibleInventoryFile() (string, bool) {
	ansHome := os.Getenv("ANSIBLE_DIR")
	log.Printf("Using ANSIBLE_DIR:[%s]", ansHome)

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

	for _, p := range procs.Etcds {
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			etcdRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range procs.Controllers {
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			ctrlRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range procs.Crms {
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			crmRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range procs.Dmes {
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			dmeRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}

	//create ansible inventory
	fmt.Fprintln(invfile, "[mexservers]")
	for s := range allRemoteServers {
		fmt.Fprintln(invfile, s)
	}
	fmt.Fprintln(invfile, "")
	fmt.Fprintln(invfile, "[etcds]")
	for s := range etcdRemoteServers {
		fmt.Fprintln(invfile, s)
	}
	fmt.Fprintln(invfile, "")
	fmt.Fprintln(invfile, "[controllers]")
	for s := range ctrlRemoteServers {
		fmt.Fprintln(invfile, s)
	}
	fmt.Fprintln(invfile, "")
	fmt.Fprintln(invfile, "[crms]")
	for s := range crmRemoteServers {
		fmt.Fprintln(invfile, s)
	}
	fmt.Fprintln(invfile, "")
	fmt.Fprintln(invfile, "[dmes]")
	for s := range dmeRemoteServers {
		fmt.Fprintln(invfile, s)
	}
	fmt.Fprintln(invfile, "")
	return invfile.Name(), foundServer
}

func deployProcesses() bool {
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_deploy.yml"
	return runPlaybook(playbook, []string{})
}

func startRemoteProcesses() bool {
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_start.yml"

	return runPlaybook(playbook, []string{})
}

func stopRemoteProcesses() bool {
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_stop.yml"
	return runPlaybook(playbook, []string{})
}

func cleanupRemoteProcesses() bool {
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_cleanup.yml"
	return runPlaybook(playbook, []string{})
}

func fetchRemoteLogs() bool {
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_fetch_logs.yml"
	return runPlaybook(playbook, []string{"local_log_path=" + *outputDir})
}

func startProcesses() bool {
	log.Printf("*** Starting Local Processes")
	for _, etcd := range procs.Etcds {
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
	return true
}

func validateArgs() {
	flag.Parse()
	_, validDeployment := deploymentChoices[*deployment]
	errFound := false

	actionSlice = strings.Split(*actions, ",")
	for _, action := range actionSlice {
		_, validAction := actionChoices[action]
		if !validAction {
			fmt.Printf("ERROR: -actions must be one of: %v, received: %s\n", actionList, action)
			errFound = true
		} else if (action == "update" || action == "create") && *dataFile == "" {
			if *dataFile == "" {
				fmt.Printf("ERROR: if action=update or create, -datafile must be specified\n")
				errFound = true
			}
		} else if action == "fetchlogs" && *outputDir == "" {
			fmt.Printf("ERROR: cannot use action=fetchlogs option without -outputdir\n")
			errFound = true
		}
	}

	if !validDeployment {
		fmt.Printf("ERROR: -deployment must be one of: %v\n", deploymentList)
		errFound = true
	}
	if *dataFile != "" {
		if _, err := os.Stat(*dataFile); err != nil {
			fmt.Printf("ERROR: file " + *dataFile + " does not exist\n")
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
	if *useTimestamp && *outputDir == "" {
		fmt.Printf("ERROR: cannot use -timestamp option without -outputdir\n")
		errFound = true
	}

	if errFound {
		printUsage()
		os.Exit(1)
	}
}

func main() {
	validateArgs()

	if *outputDir != "" {
		createOutputDir()
	}
	if *dataFile != "" {
		readDataFile(*dataFile)
	}
	if *setupFile != "" {
		readSetupFile(*setupFile)
	}

	for _, action := range actionSlice {
		log.Printf("\n\n************ Begin: %s ************\n\n ", action)
		switch action {
		case "deploy":
			deployProcesses()
		case "start":
			startFailed := false
			if !startProcesses() {
				startFailed = true
			} else {
				if !startRemoteProcesses() {
					startFailed = true
				}
			}
			if startFailed {
				stopProcesses()
				stopRemoteProcesses()
				log.Fatal("Startup failed, exiting")
				os.Exit(1)
			}
			updateApiAddrs()
			waitForProcesses()

		case "status":
			updateApiAddrs()
			waitForProcesses()
		case "stop":
			stopProcesses()
			stopRemoteProcesses()
		case "create":
			fallthrough
		case "update":
			fallthrough
		case "delete":
			updateApiAddrs()
			if !runApis(action) {
				log.Printf("Unable to run apis for %s. Check connectivity to controller API\n", action)
			}
		case "show":
			updateApiAddrs()
			runShowCommands()
		case "cleanup":
			cleanupRemoteProcesses()
		case "fetchlogs":
			fetchRemoteLogs()
		default:
			log.Fatal("unexpected action: " + action)
		}
	}
	if *outputDir != "" {
		fmt.Printf("\nResults in: %s\n", *outputDir)
	}
}
