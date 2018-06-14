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

	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	yaml "gopkg.in/yaml.v2"
)

var (
	commandName = "setup-mex"
	action      = flag.String("action", "", actionList)
	deployment  = flag.String("deployment", "process", deploymentList)
	dataFile    = flag.String("datafile", "", "optional yml data file")
	setupFile   = flag.String("setupfile", "", "mandatory yml topology file")
	procs       processData
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
	"start":   true,
	"stop":    true,
	"status":  true,
	"update":  true,
	"create":  true,
	"deploy":  true,
	"cleanup": true,
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

// TODO, would be nice to figure how to do these 3 with the same implementation
func connectController(p *process.ControllerLocal, c chan string) {
	log.Printf("attempt to connect to process %v", (*p).Name)
	api, err := (*p).ConnectAPI(10 * time.Second)
	if err != nil {
		c <- "Failed to connect to " + (*p).Name
	} else {
		c <- "OK connect to " + (*p).Name
		api.Close()
	}
}

func connectCrm(p *process.CrmLocal, c chan string) {
	log.Printf("attempt to connect to process %v", (*p).Name)
	api, err := (*p).ConnectAPI(10 * time.Second)
	if err != nil {
		c <- "Failed to connect to " + (*p).Name
	} else {
		c <- "OK connect to " + (*p).Name
		api.Close()
	}
}

func connectDme(p *process.DmeLocal, c chan string) {
	log.Printf("attempt to connect to process %v", (*p).Name)
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

func applyApplicationData(update bool) bool {
	log.Printf("Applying data via APIs\n")

	//just connect to the first controller, it should sync
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
		for _, op := range data.Operators {
			log.Printf("API for operator: %v", op.Key.Name)
			if update {
				_, err = opAPI.UpdateOperator(ctx, &op)
			} else {
				_, err = opAPI.CreateOperator(ctx, &op)
			}
			if err != nil {
				log.Printf("Error api operator: %v", err)
			}
		}
		devAPI := edgeproto.NewDeveloperApiClient(ctrlapi)
		for _, dev := range data.Developers {
			log.Printf("API for developer: %v", dev.Key.Name)
			if update {
				_, err = devAPI.UpdateDeveloper(ctx, &dev)
			} else {
				_, err = devAPI.CreateDeveloper(ctx, &dev)
			}
			if err != nil {
				log.Printf("Error api developer: %v", err)
			}
		}
		clAPI := edgeproto.NewCloudletApiClient(ctrlapi)
		for _, cl := range data.Cloudlets {
			log.Printf("API for cloudlet: %v", cl.Key.Name)
			if update {
				_, err = clAPI.UpdateCloudlet(ctx, &cl)
			} else {
				_, err = clAPI.CreateCloudlet(ctx, &cl)
			}
			if err != nil {
				log.Printf("Error api cloudlet: %v", err)
			}
		}
		appAPI := edgeproto.NewAppApiClient(ctrlapi)
		for _, app := range data.Applications {
			log.Printf("API for app: %v", app.Key.Name)
			if update {
				_, err = appAPI.UpdateApp(ctx, &app)
			} else {
				_, err = appAPI.CreateApp(ctx, &app)
			}
			if err != nil {
				log.Printf("Error api app: %v", err)
			}
		}
		appinAPI := edgeproto.NewAppInstApiClient(ctrlapi)
		for _, appin := range data.AppInstances {
			log.Printf("API for appInst: %v", appin.Key.AppKey.Name)
			if update {
				_, err = appinAPI.UpdateAppInst(ctx, &appin)
			} else {
				_, err = appinAPI.CreateAppInst(ctx, &appin)
			}
			if err != nil {
				log.Printf("Error api appinst: %v", err)
			}

		}
		cancel()
	}
	ctrlapi.Close()
	log.Printf("Done applying data\n")
	return true
}

func getLogFile(procname string) string {
	//hardcoding to current dir for now, make this an option maybe
	return "./" + procname + ".log"
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

func cleanup() {
	for _, p := range procs.Etcds {
		logfile := getLogFile(p.Name)
		log.Printf("Cleaning logfile %v", logfile)
		os.Remove(logfile)
	}
	for _, p := range procs.Controllers {
		logfile := getLogFile(p.Name)
		log.Printf("Cleaning logfile %v", logfile)
		os.Remove(logfile)
	}
	for _, p := range procs.Crms {
		logfile := getLogFile(p.Name)
		log.Printf("Cleaning logfile %v", logfile)
		os.Remove(logfile)
	}
	for _, p := range procs.Dmes {
		logfile := getLogFile(p.Name)
		log.Printf("Cleaning logfile %v", logfile)
		os.Remove(logfile)
	}
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
	log.Printf("Running ansible-playbook %s -i %s -e %s", playbook, invFile, "'"+argstr+"'")
	cmd := exec.Command("ansible-playbook", "-i", invFile, "-e", argstr, playbook)

	output, err := cmd.CombinedOutput()
	fmt.Printf("Ansible Output:\n%v\n", string(output))

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
	log.Printf("\n\n************ Starting Remote Processes ************\n\n ")

	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_start.yml"

	return runPlaybook(playbook, []string{})
}

func stopRemoteProcesses() bool {
	log.Printf("\n\n************ Stopping Remote Processes ************\n\n ")

	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_stop.yml"
	return runPlaybook(playbook, []string{})
}

func cleanupRemoteProcesses() bool {
	log.Printf("\n\n************ Cleanup Remote Processes ************\n\n ")

	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_cleanup.yml"
	return runPlaybook(playbook, []string{})
}

func startProcesses() bool {
	log.Printf("\n\n************ Starting Local Processes ************\n\n ")
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
	_, validAction := actionChoices[*action]
	_, validDeployment := deploymentChoices[*deployment]

	errFound := false

	if !validAction {
		fmt.Printf("ERROR: -action must be one of: %v\n", actionList)
		errFound = true
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
	} else if *action == "update" || *action == "create" {
		fmt.Printf("ERROR: if -action=update or create, -datafile must be specified\n")
		errFound = true
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
	if errFound {
		printUsage()
		os.Exit(1)
	}
}

func main() {
	validateArgs()

	if *dataFile != "" {
		readDataFile(*dataFile)
	}
	if *setupFile != "" {
		readSetupFile(*setupFile)
	}
	switch *action {
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
		waitForProcesses()
		if *dataFile != "" {
			if !applyApplicationData(false) {
				log.Println("Unable to apply application data. Check connectivity to controller APIs")
			}
		}
	case "status":
		waitForProcesses()
	case "stop":
		stopProcesses()
		stopRemoteProcesses()
	case "create":
		if !applyApplicationData(false) {
			log.Println("Unable to apply application data for create. Check connectivity to controller APIs")
		}
	case "update":
		if !applyApplicationData(true) {
			log.Println("Unable to apply application data for update. Check connectivity to controller APIs")
		}
	case "cleanup":
		cleanup()
		cleanupRemoteProcesses()
	default:
		log.Fatal("unexpected action: " + *action)
	}
	fmt.Println("Done")

}
