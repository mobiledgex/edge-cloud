package setupmex

// consists of utilities used to deploy, start, stop MEX processes either locally or remotely via Ansible.

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/setup-env/util"

	yaml "gopkg.in/yaml.v2"
)

var apiAddrsUpdated = false

//when first creating a cluster, it may take a while for the load balancer to get an IP. Usually
// this happens much faster, but occasionally it takes longer
var maxWaitForServiceSeconds = 300

func WaitForProcesses(processName string) bool {
	log.Println("Wait for processes to respond to APIs")
	c := make(chan util.ReturnCodeWithText)
	var numProcs = 0 //len(procs Controllers) + len(procs.Crms) + len(procs.Dmes)
	for _, ctrl := range util.Deployment.Controllers {
		if processName != "" && processName != ctrl.Name {
			continue
		}
		numProcs += 1
		ctrlp := ctrl.ControllerLocal
		go util.ConnectController(&ctrlp, c)
	}
	for _, dme := range util.Deployment.Dmes {
		if processName != "" && processName != dme.Name {
			continue
		}
		numProcs += 1
		dmep := dme.DmeLocal
		go util.ConnectDme(&dmep, c)
	}
	for _, crm := range util.Deployment.Crms {
		if processName != "" && processName != crm.Name {
			continue
		}
		numProcs += 1
		crmp := crm.CrmLocal
		go util.ConnectCrm(&crmp, c)
	}
	allpass := true
	for i := 0; i < numProcs; i++ {
		rc := <-c
		log.Println(rc.Text)
		if !rc.Success {
			allpass = false
		}
	}
	return allpass
}

//to identify running processes, we need to know where they are running
//and some unique arguments to look for for pgrep. Returns info about
// the hostname, process executable and arguments
func findProcess(processName string) (string, string, string) {
	for _, etcd := range util.Deployment.Etcds {
		if etcd.Name == processName {
			return etcd.Hostname, "etcd", "-name " + etcd.Name
		}
	}
	for _, ctrl := range util.Deployment.Controllers {
		if ctrl.Name == processName {
			return ctrl.Hostname, "controller", "-apiAddr " + ctrl.ApiAddr
		}
	}
	for _, crm := range util.Deployment.Crms {
		if crm.Name == processName {

			return crm.Hostname, "crmserver", "-apiAddr " + crm.ApiAddr
		}
	}
	for _, dme := range util.Deployment.Dmes {
		if dme.Name == processName {
			return dme.Hostname, "dme-server", "-apiAddr " + dme.ApiAddr
		}
	}
	for _, tok := range util.Deployment.Toksims {
		if tok.Name == processName {
			return tok.Hostname, "tok-srv-sim", fmt.Sprintf("port -%d", tok.Port)
		}
	}
	for _, loc := range util.Deployment.Locsims {
		if loc.Name == processName {
			return loc.Hostname, "loc-api-sim", fmt.Sprintf("port -%d", loc.Port)
		}
	}
	for _, sam := range util.Deployment.SampleApps {
		if sam.Name == processName {
			argstr := strings.Join(sam.Args, " ")
			return sam.Hostname, sam.Exename, argstr
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
func UpdateApiAddrs() {
	if apiAddrsUpdated {
		//no need to do this more than once
		return
	}
	//for k8s deployments, get the ip from the service
	if util.IsK8sDeployment() {
		if len(util.Deployment.Controllers) > 0 {
			addr, err := util.GetK8sServiceAddr("controller", maxWaitForServiceSeconds)
			if err != nil {
				log.Fatalf("unable to get controller service")
			}
			util.Deployment.Controllers[0].ApiAddr = addr
		}
		if len(util.Deployment.Dmes) > 0 {
			addr, err := util.GetK8sServiceAddr("dme", maxWaitForServiceSeconds)
			if err != nil {
				log.Fatalf("unable to get dme service")
			}
			util.Deployment.Dmes[0].ApiAddr = addr
		}
		if len(util.Deployment.Crms) > 0 {
			addr, err := util.GetK8sServiceAddr("crm", maxWaitForServiceSeconds)
			if err != nil {
				log.Fatalf("unable to get crm service")
			}
			util.Deployment.Crms[0].ApiAddr = addr
		}
	} else {
		for i, ctrl := range util.Deployment.Controllers {
			util.Deployment.Controllers[i].ApiAddr = getExternalApiAddress(ctrl.ApiAddr, ctrl.Hostname)
		}
		for i, dme := range util.Deployment.Dmes {
			util.Deployment.Dmes[i].ApiAddr = getExternalApiAddress(dme.ApiAddr, dme.Hostname)
		}
		for i, crm := range util.Deployment.Crms {
			util.Deployment.Crms[i].ApiAddr = getExternalApiAddress(crm.ApiAddr, crm.Hostname)
		}
	}
	apiAddrsUpdated = true
}

func getLogFile(procname string, outputDir string) string {
	if outputDir == "" {
		return "./" + procname + ".log"
	} else {
		return outputDir + "/" + procname + ".log"
	}
}

func ReadSetupFile(setupfile string, dataDir string) bool {
	//the setup file has a vars section with replacement variables.  ingest the file once
	//to get these variables, and then ingest again to parse the setup data with the variables
	var varlist []string
	var replacementVars util.YamlReplacementVariables
	util.DeploymentReplacementVars = ""

	if dataDir != "" {
		varlist = append(varlist, "datadir="+dataDir)
	}
	util.ReadYamlFile(setupfile, &replacementVars, "", false)

	for _, repl := range replacementVars.Vars {
		for varname, value := range repl {
			varlist = append(varlist, varname+"="+value)
		}
	}
	if len(varlist) > 0 {
		util.DeploymentReplacementVars = strings.Join(varlist, ",")
	}
	err := util.ReadYamlFile(setupfile, &util.Deployment, util.DeploymentReplacementVars, true)
	if err != nil {
		if !util.IsYamlOk(err, "setup") {
			fmt.Fprintf(os.Stderr, "One or more fatal unmarshal errors in %s", setupfile)
			return false
		}
	}
	//equals sign is not well handled in yaml so it is url encoded and changed after loading
	//for some reason, this only happens when the yaml is read as ProcessData and not
	//as a generic interface.  TODO: further study on this.
	for i, _ := range util.Deployment.Dmes {
		util.Deployment.Dmes[i].TokSrvUrl = strings.Replace(util.Deployment.Dmes[i].TokSrvUrl, "%3D", "=", -1)
	}
	return true
}

func StopProcesses(processName string) bool {
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
	for _, p := range util.Deployment.Etcds {
		log.Printf("cleaning etcd %+v", p)
		p.ResetData()
	}

	processExeNames := []string{"etcd", "controller", "dme-server", "crmserver", "loc-api-sim", "tok-srv-sim"}
	for _, a := range util.Deployment.SampleApps {
		processExeNames = append(processExeNames, a.Exename)
	}

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

	if !stageYamlFile("setup.yml", ansHome+"/playbooks", &util.Deployment) {
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
		fmt.Fprintf(os.Stderr, "Ansible playbook failed: %v ", err)
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
		fmt.Fprintf(os.Stderr, "Error reading location file %s -- %v\n", srcFile, yerr)
	}
	if !stageYamlFile("locsim.yml", destDir, locdb) {
		fmt.Fprintf(os.Stderr, "Error staging location db file %s to %s\n", srcFile, destDir)
	}
}

func createAnsibleInventoryFile(procNameFilter string) (string, bool) {
	ansHome := os.Getenv("ANSIBLE_DIR")
	log.Printf("Creating inventory file in ANSIBLE_DIR:%s using procname filter: %s\n", ansHome, procNameFilter)

	foundServer := false
	if ansHome == "" {
		fmt.Fprint(os.Stderr, "Need to set ANSIBLE_DIR environment variable for deployment")
	}

	invfile, err := os.Create(ansHome + "/mex_inventory")
	log.Printf("Creating inventory file: %v", invfile.Name())
	if err != nil {
		fmt.Fprint(os.Stderr, "Cannot create file", err)
	}
	defer invfile.Close()

	allRemoteServers := make(map[string]string)
	etcdRemoteServers := make(map[string]string)
	ctrlRemoteServers := make(map[string]string)
	crmRemoteServers := make(map[string]string)
	dmeRemoteServers := make(map[string]string)
	locApiSimulators := make(map[string]string)
	tokSrvSimulators := make(map[string]string)
	sampleApps := make(map[string]string)

	for _, p := range util.Deployment.Etcds {
		if procNameFilter != "" && procNameFilter != p.Name {
			continue
		}
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			etcdRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range util.Deployment.Controllers {
		if procNameFilter != "" && procNameFilter != p.Name {
			continue
		}
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			ctrlRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range util.Deployment.Crms {
		if procNameFilter != "" && procNameFilter != p.Name {
			continue
		}
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			crmRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range util.Deployment.Dmes {
		if procNameFilter != "" && procNameFilter != p.Name {
			continue
		}
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			dmeRemoteServers[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range util.Deployment.Locsims {
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
	for _, p := range util.Deployment.Toksims {
		if procNameFilter != "" && procNameFilter != p.Name {
			continue
		}
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			tokSrvSimulators[p.Hostname] = p.Name
			foundServer = true
		}
	}
	for _, p := range util.Deployment.SampleApps {
		if procNameFilter != "" && procNameFilter != p.Name {
			continue
		}
		if p.Hostname != "" && p.Hostname != "localhost" && p.Hostname != "127.0.0.1" {
			allRemoteServers[p.Hostname] = p.Name
			sampleApps[p.Hostname] = p.Name
			foundServer = true
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
	if len(tokSrvSimulators) > 0 {
		fmt.Fprintln(invfile, "")
		fmt.Fprintln(invfile, "[toksims]")
		for s := range tokSrvSimulators {
			fmt.Fprintln(invfile, s)
		}
	}
	if len(sampleApps) > 0 {
		fmt.Fprintln(invfile, "")
		fmt.Fprintln(invfile, "[sampleapps]")
		for s := range sampleApps {
			fmt.Fprintln(invfile, s)
		}
	}
	fmt.Fprintln(invfile, "")
	return invfile.Name(), foundServer
}

func getCloudflareUserAndKey() (string, string) {
	user := os.Getenv("MEX_CF_USER")
	apikey := os.Getenv("MEX_CF_KEY")
	return user, apikey
}

func createCloudfareRecords() error {
	log.Printf("createCloudfareRecords\n")

	ttl := 300
	if util.Deployment.Cloudflare.Zone == "" {
		return nil
	}
	user, apiKey := getCloudflareUserAndKey()
	if user == "" || apiKey == "" {
		log.Printf("Unable to get Cloudflare settings\n")
		return fmt.Errorf("need to set MEX_CF_USER and MEX_CF_KEY for cloudflare")
	}

	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Printf("Error in getting Cloudflare API %v\n", err)
		return err
	}
	zoneID, err := api.ZoneIDByName(util.Deployment.Cloudflare.Zone)
	if err != nil {
		log.Printf("Cloudflare zone error: %v\n", err)
		return err
	}
	for _, r := range util.Deployment.Cloudflare.Records {
		log.Printf("adding dns entry: %s content: %s \n", r.Name, r.Content)

		record := cloudflare.DNSRecord{
			Name:    strings.ToLower(r.Name),
			Type:    strings.ToUpper(r.Type),
			Content: r.Content,
			TTL:     ttl,
			Proxied: false,
		}

		resp, err := api.CreateDNSRecord(zoneID, record)
		if err != nil {
			log.Printf("Error, cannot create DNS record for zone %s, %v", zoneID, err)
			return err
		}
		log.Printf("Cloudflare DNS Response %+v\n", resp)
	}
	return nil
}

func deleteCloudfareRecords() error {
	log.Printf("deleteCloudfareRecords\n")

	if util.Deployment.Cloudflare.Zone == "" {
		return nil
	}
	user, apiKey := getCloudflareUserAndKey()
	if user == "" || apiKey == "" {
		log.Printf("Unable to get Cloudflare settings\n")
		return fmt.Errorf("need to set CF_USER and CF_KEY for cloudflare")
	}
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Printf("Error in getting Cloudflare API %v\n", err)
		return err
	}
	zoneID, err := api.ZoneIDByName(util.Deployment.Cloudflare.Zone)
	if err != nil {
		log.Printf("Cloudflare zone error: %v\n", err)
		return err
	}

	//make a hash of the records we are looking for so we don't have to iterate thru the
	//list many times
	recordsToClean := make(map[string]bool)
	for _, d := range util.Deployment.Cloudflare.Records {
		recordsToClean[strings.ToLower(d.Name+d.Type+d.Content)] = true
		log.Printf("cloudflare recordsToClean: %v", d.Name+d.Type+d.Content)
	}

	//find all the records for the zone and delete ours.  Alternately we could apply a filter when doing the query
	//but there could be multiple records and building that filter could be hard
	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{})
	for _, r := range records {
		_, exists := recordsToClean[strings.ToLower(r.Name+r.Type+r.Content)]
		if exists {
			log.Printf("Found a DNS record to delete %v\n", r)
			err := api.DeleteDNSRecord(zoneID, r.ID)
			if err != nil {
				log.Printf("Error in deleting DNS record for %s - %v\n", r.Name, err)
				return err
			}
		}
	}

	return nil
}

func DeployProcesses() bool {
	if util.IsK8sDeployment() {
		return true //nothing to do for k8s
	}
	err := createCloudfareRecords()
	if err != nil {
		return false
	}
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_deploy.yml"
	return runPlaybook(playbook, []string{}, "")
}

func StartRemoteProcesses(processName string) bool {
	if util.IsK8sDeployment() {
		return true //nothing to do for k8s
	}
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_start.yml"

	return runPlaybook(playbook, []string{}, processName)
}

func StopRemoteProcesses(processName string) bool {

	if util.IsK8sDeployment() {
		return true //nothing to do for k8s
	}

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

func CleanupRemoteProcesses() bool {
	cfrc := true
	err := deleteCloudfareRecords()
	if err != nil {
		cfrc = false
	}
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_cleanup.yml"
	return cfrc && runPlaybook(playbook, []string{}, "")
}

func FetchRemoteLogs(outputDir string) bool {
	if util.IsK8sDeployment() {
		//TODO: need to get the logs from K8s
		return true
	}
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_fetch_logs.yml"
	return runPlaybook(playbook, []string{"local_log_path=" + outputDir}, "")
}

func StartProcesses(processName string, outputDir string) bool {
	if util.IsK8sDeployment() {
		return true //nothing to do for k8s
	}
	util.PrintStepBanner("starting local processes")
	for _, etcd := range util.Deployment.Etcds {
		if processName != "" && processName != etcd.Name {
			continue
		}
		if etcd.Hostname == "localhost" || etcd.Hostname == "127.0.0.1" {
			log.Printf("Starting Etcd %+v", etcd)
			if processName == "" {
				//only reset the data if this is a full start of all etcds
				etcd.ResetData()
			}
			logfile := getLogFile(etcd.Name, outputDir)
			err := etcd.Start(logfile)
			if err != nil {
				log.Printf("Error on Etcd startup: %v", err)
				return false
			}
		}
	}
	for _, ctrl := range util.Deployment.Controllers {
		if processName != "" && processName != ctrl.Name {
			continue
		}
		if ctrl.Hostname == "localhost" || ctrl.Hostname == "127.0.0.1" {
			log.Printf("Starting Controller %+v\n", ctrl)
			logfile := getLogFile(ctrl.Name, outputDir)
			err := ctrl.Start(logfile, process.WithDebug("etcd,api,notify"))
			if err != nil {
				log.Printf("Error on controller startup: %v", err)
				return false
			}
		}
	}
	for _, crm := range util.Deployment.Crms {
		if processName != "" && processName != crm.Name {
			continue
		}
		if crm.Hostname == "localhost" || crm.Hostname == "127.0.0.1" {
			log.Printf("Starting CRM %+v\n", crm)
			logfile := getLogFile(crm.Name, outputDir)
			err := crm.Start(logfile, process.WithDebug("api,notify,mexos"))
			if err != nil {
				log.Printf("Error on CRM startup: %v", err)
				return false
			}
		}
	}
	for _, dme := range util.Deployment.Dmes {
		if processName != "" && processName != dme.Name {
			continue
		}
		if dme.Hostname == "localhost" || dme.Hostname == "127.0.0.1" {
			log.Printf("Starting DME %+v\n", dme)
			logfile := getLogFile(dme.Name, outputDir)
			err := dme.Start(logfile, process.WithDebug("dmelocapi,dmedb,dmereq"))
			if err != nil {
				log.Printf("Error on DME startup: %v", err)
				return false
			}
		}
	}
	for _, loc := range util.Deployment.Locsims {
		if processName != "" && processName != loc.Name {
			continue
		}
		if loc.Hostname == "localhost" || loc.Hostname == "127.0.0.1" {
			log.Printf("Starting LocSim %+v\n", loc)
			if loc.Locfile != "" {
				stageLocDbFile(loc.Locfile, "/var/tmp")
			}
			logfile := getLogFile(loc.Name, outputDir)
			err := loc.Start(logfile)
			if err != nil {
				log.Printf("Error on LocSim startup: %v", err)
				return false
			}

		}
	}
	for _, tok := range util.Deployment.Toksims {
		if processName != "" && processName != tok.Name {
			continue
		}
		if tok.Hostname == "localhost" || tok.Hostname == "127.0.0.1" {
			log.Printf("Starting TokSim %+v\n", tok)
			logfile := getLogFile(tok.Name, outputDir)
			err := tok.Start(logfile)
			if err != nil {
				log.Printf("Error on TokSim startup: %v", err)
				return false
			}
		}
	}
	for _, sam := range util.Deployment.SampleApps {
		if processName != "" && processName != sam.Name {
			continue
		}
		if sam.Hostname == "localhost" || sam.Hostname == "127.0.0.1" {
			log.Printf("Starting Sample app %+v\n", sam)
			logfile := getLogFile(sam.Name, outputDir)
			err := sam.Start(logfile)
			if err != nil {
				log.Printf("Error on Sample App startup: %v", err)
				return false
			}
		}
	}
	return true
}
