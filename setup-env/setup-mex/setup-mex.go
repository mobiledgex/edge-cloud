package setupmex

// consists of utilities used to deploy, start, stop MEX processes either locally or remotely via Ansible.

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	yaml "gopkg.in/yaml.v2"
)

var apiAddrsUpdated = false

//root TLS Dir
var tlsDir = ""

//outout TLS cert dir
var tlsOutDir = ""

//when first creating a cluster, it may take a while for the load balancer to get an IP. Usually
// this happens much faster, but occasionally it takes longer
var maxWaitForServiceSeconds = 900 //15 min

func isLocalIP(hostname string) bool {
	return hostname == "localhost" || hostname == "127.0.0.1"
}

func WaitForProcesses(processName string) bool {
	log.Println("Wait for processes to respond to APIs")
	c := make(chan util.ReturnCodeWithText)
	procs := make([]process.Process, 0)
	for _, ctrl := range util.Deployment.Controllers {
		if processName != "" && processName != ctrl.Name {
			continue
		}
		procs = append(procs, ctrl)
		go util.ConnectController(ctrl, c)
	}
	for _, dme := range util.Deployment.Dmes {
		if processName != "" && processName != dme.Name {
			continue
		}
		procs = append(procs, dme)
		go util.ConnectDme(dme, c)
	}
	for _, crm := range util.Deployment.Crms {
		if processName != "" && processName != crm.Name {
			continue
		}
		procs = append(procs, crm)
		go util.ConnectCrm(crm, c)
	}
	allpass := true
	for i := 0; i < len(procs); i++ {
		rc := <-c
		log.Println(rc.Text)
		if !rc.Success {
			allpass = false
		}
	}
	if !ensureProcesses(procs) {
		allpass = false
	}
	return allpass
}

// This uses the same methods as kill processes to look for local processes,
// to ensure that the lookup method for finding local processes is valid.
func ensureProcesses(procs []process.Process) bool {
	if util.IsK8sDeployment() {
		return true
	}
	ensured := true
	for _, p := range procs {
		if !isLocalIP(p.GetHostname()) {
			continue
		}
		exeName := p.GetExeName()
		args := p.LookupArgs()
		log.Printf("Looking for host %v processexe %v processargs %v\n", p.GetHostname(), exeName, args)
		if !util.EnsureProcessesByName(exeName, args) {
			ensured = false
			break
		}
	}
	return ensured
}

func getExternalApiAddress(internalApiAddr string, externalHost string) string {
	//in cloud deployments, the internal address the controller listens to may be different than the
	//external address which clients need to use.   So use the external hostname and api port
	if externalHost == "0.0.0.0" || externalHost == "127.0.0.1" {
		// local host: prevent swapping around these two addresses
		// because they are used interchangably between the host and
		// api addr fields, and they are also used by pgrep to search
		// for the process, which can cause pgrep to fail to find the
		// process.
		return internalApiAddr
	}
	return externalHost + ":" + strings.Split(internalApiAddr, ":")[1]
}

// if there is a DNS address configured we will use that.  Required because TLS certs
// are generated against the DNS name if one is available.
func getDNSNameForAddr(addr string) string {
	//split the port off

	ss := strings.Split(addr, ":")
	if len(ss) < 2 {
		return addr
	}
	a := ss[0]
	p := ss[1]

	for _, r := range util.Deployment.Cloudflare.Records {
		if r.Content == a {
			return r.Name + ":" + p
		}
	}
	// no record found, just use the add
	return addr
}

//in cloud deployments, the internal address the controller listens to may be different than the
//external address which clients need to use as floating IPs are used.  So use the external
//hostname and api port when connecting to the API.  This needs to be done after startup
//but before trying to connect to the APIs remotely
func UpdateAPIAddrs() bool {
	if apiAddrsUpdated {
		//no need to do this more than once
		return true
	}
	//for k8s deployments, get the ip from the service
	if util.IsK8sDeployment() {
		if len(util.Deployment.Controllers) > 0 {
			if util.Deployment.Controllers[0].ApiAddr != "" {
				for i, ctrl := range util.Deployment.Controllers {
					util.Deployment.Controllers[i].ApiAddr = getExternalApiAddress(ctrl.ApiAddr, ctrl.Hostname)
					log.Printf("set controller API addr to %s\n", util.Deployment.Controllers[i].ApiAddr)
				}
			} else {
				addr, err := util.GetK8sServiceAddr("controller", maxWaitForServiceSeconds)
				if err != nil {
					fmt.Fprintf(os.Stderr, "unable to get controller service ")
					return false
				}
				util.Deployment.Controllers[0].ApiAddr = addr
				log.Printf("set controller API addr from k8s service to %s\n", addr)

			}
		}
		if len(util.Deployment.Dmes) > 0 {
			if util.Deployment.Dmes[0].ApiAddr != "" {
				for i, dme := range util.Deployment.Dmes {
					util.Deployment.Dmes[i].ApiAddr = getExternalApiAddress(dme.ApiAddr, dme.Hostname)
				}
			} else {
				addr, err := util.GetK8sServiceAddr("dme", maxWaitForServiceSeconds)
				if err != nil {
					fmt.Fprintf(os.Stderr, "unable to get dme service ")
					return false
				}
				util.Deployment.Dmes[0].ApiAddr = addr
			}
		}
		if len(util.Deployment.Crms) > 0 {
			addr, err := util.GetK8sServiceAddr("crm", maxWaitForServiceSeconds)
			if err != nil {
				//we may not always deploy CRM with service addresses if it is in
				//the same cluster as the controller and we don't need the direct api
				if strings.HasSuffix(err.Error(), "not found") {
					log.Printf("No CRM service")
					addr = util.ApiAddrNone
				} else {
					fmt.Fprintf(os.Stderr, "unable to get crm service ")
					return false
				}
			}
			addr = getDNSNameForAddr(addr)
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
	return true
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

	//{{tlsoutdir}} is relative to the GO dir.
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		fmt.Fprintf(os.Stderr, "GOPATH not set, cannot calculate tlsoutdir")
		return false
	}
	tlsDir = goPath + "/src/github.com/mobiledgex/edge-cloud/tls"
	tlsOutDir = tlsDir + "/out"
	varlist = append(varlist, "tlsoutdir="+tlsOutDir)

	setupdir := filepath.Dir(setupfile)
	varlist = append(varlist, "setupdir="+setupdir)

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

// CleanupDIND kills all containers on the kubeadm-dind-net-xxx network and then cleans up DIND
func CleanupDIND() error {
	// find docker networks
	log.Printf("Running CleanupDIND, getting docker networks\n")
	r, _ := regexp.Compile("kubeadm-dind-net(-(\\S+)-(\\d+))?")

	lscmd := exec.Command("docker", "network", "ls", "--format='{{.Name}}'")
	output, err := lscmd.Output()
	if err != nil {
		log.Printf("Error running docker network ls: %v", err)
		return err
	}
	netnames := strings.Split(string(output), "\n")
	for _, n := range netnames {
		n := strings.Trim(n, "'") //remove quotes
		if r.MatchString(n) {
			matches := r.FindStringSubmatch(n)
			clusterName := matches[2]
			clusterID := matches[3]

			log.Printf("found DIND net: %s clusterName: %s clusterID: %s\n", n, clusterName, clusterID)
			inscmd := exec.Command("docker", "network", "inspect", n, "--format={{range .Containers}}{{.Name}},{{end}}")
			output, err = inscmd.CombinedOutput()
			if err != nil {
				log.Printf("Error running docker network inspect: %s - %v - %v\n", n, string(output), err)
				return fmt.Errorf("error in docker inspect %v", err)
			}
			ostr := strings.TrimRight(string(output), ",") //trailing comma
			ostr = strings.TrimSpace(ostr)
			containers := strings.Split(ostr, ",")
			// first we need to kill all containers using the network as the dind script will
			// not clean these up, and cannot delete the network if they are present
			for _, c := range containers {
				if c == "" {
					continue
				}
				if strings.HasPrefix(c, "kube-node") || strings.HasPrefix(c, "kube-master") {
					// dind clean should handle this
					log.Printf("skipping kill of kube container: %s\n", c)
					continue
				}
				log.Printf("killing container: [%s]\n", c)
				killcmd := exec.Command("docker", "kill", c)
				output, err = killcmd.CombinedOutput()
				if err != nil {
					log.Printf("Error killing docker container: %s - %v - %v\n", c, string(output), err)
					return fmt.Errorf("error in docker kill %v", err)
				}
			}
			// now cleanup DIND cluster
			if clusterName != "" {
				os.Setenv("DIND_LABEL", clusterName)
				os.Setenv("CLUSTER_ID", clusterID)
			} else {
				log.Printf("no clustername, doing clean for default cluster")
				os.Unsetenv("DIND_LABEL")
				os.Unsetenv("CLUSTER_ID")
			}
			log.Printf("running dind-cluster-v1.13.sh clean clusterName: %s clusterID: %s\n", clusterName, clusterID)
			out, err := sh.Command("dind-cluster-v1.13.sh", "clean").CombinedOutput()
			if err != nil {
				log.Printf("Error in dind-cluster-v1.13.sh clean: %v - %v\n", out, err)
				return fmt.Errorf("ERROR in cleanup Dind Cluster: %s", clusterName)
			}
			log.Printf("done dind-cluster-v1.13.sh clean for: %s out: %s\n", clusterName, out)
		}
	}
	log.Println("done CleanupDIND")
	return nil

}

func StopProcesses(processName string, allprocs []process.Process) bool {
	util.PrintStepBanner("stopping processes " + processName)
	maxWait := time.Second * 15
	c := make(chan string)
	count := 0

	for ii, p := range allprocs {
		if !isLocalIP(p.GetHostname()) {
			continue
		}
		if processName != "" && processName != p.GetName() {
			// If a process name is specified, we stop just that one.
			// The name here identifies the specific instance of the
			// application, e.g. 'ctrl1'.
			continue
		}
		log.Println("stopping/killing processes " + p.GetName())
		go util.StopProcess(allprocs[ii], maxWait, c)
		count++
	}
	if processName != "" && count == 0 {
		log.Printf("Error: unable to find process name %v in setup\n", processName)
		return false
	}

	for i := 0; i < count; i++ {
		log.Printf("stop/kill result: %v\n", <-c)
	}

	if processName == "" {
		// doing full clean up
		for _, p := range util.Deployment.Etcds {
			log.Printf("cleaning etcd %+v", p)
			p.ResetData()
		}
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

// for ansible we need to ssh to the ip address if available as the DNS record may not yet exist
func hostNameToAnsible(hostname string) string {
	for _, r := range util.Deployment.Cloudflare.Records {
		if r.Name == hostname {
			return hostname + " ansible_ssh_host=" + r.Content
		}
	}
	return hostname
}

func createAnsibleInventoryFile(procNameFilter string) (string, bool) {
	ansHome := os.Getenv("ANSIBLE_DIR")
	log.Printf("Creating inventory file in ANSIBLE_DIR:%s using procname filter: %s\n", ansHome, procNameFilter)

	if ansHome == "" {
		fmt.Fprint(os.Stderr, "Need to set ANSIBLE_DIR environment variable for deployment")
	}

	invfile, err := os.Create(ansHome + "/mex_inventory")
	log.Printf("Creating inventory file: %v", invfile.Name())
	if err != nil {
		fmt.Fprint(os.Stderr, "Cannot create file", err)
	}
	defer invfile.Close()

	//use the mobiledgex ssh key
	fmt.Fprintln(invfile, "[all:vars]")
	fmt.Fprintln(invfile, "ansible_ssh_private_key_file=~/.mobiledgex/id_rsa_mex")
	allservers := make(map[string]map[string]string)

	allprocs := util.GetAllProcesses()
	for _, p := range allprocs {
		if procNameFilter != "" && procNameFilter != p.GetName() {
			continue
		}
		if p.GetHostname() == "" || isLocalIP(p.GetHostname()) {
			continue
		}

		i := hostNameToAnsible(p.GetHostname())
		typ := process.GetTypeString(p)
		alltyps, found := allservers[typ]
		if !found {
			alltyps = make(map[string]string)
			allservers[typ] = alltyps
		}
		alltyps[i] = p.GetName()

		// type-specific stuff
		if locsim, ok := p.(*process.LocApiSim); ok {
			if locsim.Locfile != "" {
				stageLocDbFile(locsim.Locfile, ansHome+"/playbooks")
			}
		}
	}

	//create ansible inventory
	fmt.Fprintln(invfile, "[mexservers]")
	for _, alltyps := range allservers {
		for s := range alltyps {
			fmt.Fprintln(invfile, s)
		}
	}
	for typ, alltyps := range allservers {
		fmt.Fprintln(invfile, "")
		fmt.Fprintln(invfile, "["+strings.ToLower(typ)+"]")
		for s := range alltyps {
			fmt.Fprintln(invfile, s)
		}
	}
	fmt.Fprintln(invfile, "")
	return invfile.Name(), len(allservers) > 0
}

// CleanupTLSCerts . Deletes certs for a CN
func CleanupTLSCerts() error {
	if len(util.Deployment.TLSCerts) == 0 {
		//nothing to do
		return nil
	}
	for _, t := range util.Deployment.TLSCerts {
		patt := tlsOutDir + "/" + t.CommonName + ".*"
		log.Printf("Removing [%s]\n", patt)

		files, err := filepath.Glob(patt)
		if err != nil {
			return err
		}
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				return err
			}
		}
	}
	return nil
}

// GenerateTLSCerts . create tls certs using certstrap.  This requires certstrap binary to be installed.  We can eventually
// do this programmatically but certstrap has some dependency problems that require manual package workarounds
// and so will use the command for now so as not to break builds.
func GenerateTLSCerts() error {
	if len(util.Deployment.TLSCerts) == 0 {
		//nothing to do
		return nil
	}
	for _, t := range util.Deployment.TLSCerts {

		var cmdargs = []string{"--depot-path", tlsOutDir, "request-cert", "--passphrase", "", "--common-name", t.CommonName}
		if len(t.DNSNames) > 0 {
			cmdargs = append(cmdargs, "--domain", strings.Join(t.DNSNames, ","))
		}
		if len(t.IPs) > 0 {
			cmdargs = append(cmdargs, "--ip", strings.Join(t.IPs, ","))
		}

		cmd := exec.Command("certstrap", cmdargs[0:]...)
		output, err := cmd.CombinedOutput()
		log.Printf("Certstrap Request Cert cmdargs: %v output:\n%v\n", cmdargs, string(output))
		if err != nil {
			if strings.HasPrefix(string(output), "Certificate request has existed") {
				// this is ok
			} else {
				return err
			}
		}

		cmd = exec.Command("certstrap", "--depot-path", tlsOutDir, "sign", "--CA", "mex-ca", t.CommonName)
		output, err = cmd.CombinedOutput()
		log.Printf("Certstrap Sign Cert cmdargs: %v output:\n%v\n", cmdargs, string(output))
		if strings.HasPrefix(string(output), "Certificate has existed") {
			// this is ok
		} else {
			return err
		}
	}
	return nil
}

func getCloudflareUserAndKey() (string, string) {
	user := os.Getenv("MEX_CF_USER")
	apikey := os.Getenv("MEX_CF_KEY")
	return user, apikey
}

func CreateCloudflareRecords() error {
	log.Printf("createCloudflareRecords\n")

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

		addRecord := cloudflare.DNSRecord{
			Name:    strings.ToLower(r.Name),
			Type:    strings.ToUpper(r.Type),
			Content: r.Content,
			TTL:     ttl,
			Proxied: false,
		}
		queryRecord := cloudflare.DNSRecord{
			Name:    strings.ToLower(r.Name),
			Type:    strings.ToUpper(r.Type),
			Proxied: false,
		}

		records, err := api.DNSRecords(zoneID, queryRecord)
		if err != nil {
			log.Printf("Error querying dns %s, %v", zoneID, err)
			return err
		}
		for _, r := range records {
			log.Printf("Found a DNS record to delete %v\n", r)
			//we could try updating instead, but that is problematic if there
			//are multiple.  We are going to add it back anyway
			err := api.DeleteDNSRecord(zoneID, r.ID)
			if err != nil {
				log.Printf("Error in deleting DNS record for %s - %v\n", r.Name, err)
				return err
			}
		}

		resp, err := api.CreateDNSRecord(zoneID, addRecord)
		if err != nil {
			log.Printf("Error, cannot create DNS record for zone %s, %v", zoneID, err)
			return err
		}
		log.Printf("Cloudflare Create DNS Response %+v\n", resp)

	}
	return nil
}

//delete provioned records from DNS
func DeleteCloudfareRecords() error {
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
		//	recordsToClean[strings.ToLower(d.Name+d.Type+d.Content)] = true
		//delete records with the same name even if they point to a different ip
		recordsToClean[strings.ToLower(d.Name+d.Type)] = true
		log.Printf("cloudflare recordsToClean: %v", d.Name+d.Type)
	}

	//find all the records for the zone and delete ours.  Alternately we could apply a filter when doing the query
	//but there could be multiple records and building that filter could be hard
	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{})
	for _, r := range records {
		_, exists := recordsToClean[strings.ToLower(r.Name+r.Type)]
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
		p := util.GetProcessByName(processName)
		if isLocalIP(p.GetHostname()) {
			log.Printf("process %v is not remote\n", processName)
			return true
		}
		vars := []string{"processbin=" + p.GetExeName(), "processargs=\"" + p.LookupArgs() + "\""}
		playbook := ansHome + "/playbooks/mex_stop_matching_process.yml"
		return runPlaybook(playbook, vars, processName)

	}
	playbook := ansHome + "/playbooks/mex_stop.yml"
	return runPlaybook(playbook, []string{}, "")
}

func CleanupRemoteProcesses() bool {
	ansHome := os.Getenv("ANSIBLE_DIR")
	playbook := ansHome + "/playbooks/mex_cleanup.yml"
	return runPlaybook(playbook, []string{}, "")
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

func StartLocal(processName, outputDir string, p process.Process, opts ...process.StartOp) bool {
	if processName != "" && processName != p.GetName() {
		return true
	}
	if !isLocalIP(p.GetHostname()) {
		return true
	}
	envvars := p.GetEnvVars()
	if envvars != nil {
		for k, v := range envvars {
			// doing it this way means the variable is set for
			// other commands too. Not ideal but not
			// problematic, and only will happen on local
			// process deploy
			os.Setenv(k, v)
		}
	}
	typ := process.GetTypeString(p)
	log.Printf("Starting %s %s+v\n", typ, p)
	logfile := getLogFile(p.GetName(), outputDir)

	err := p.StartLocal(logfile, opts...)
	if err != nil {
		log.Printf("Error on %s startup: %v\n", typ, err)
		return false
	}
	return true
}

func StartProcesses(processName string, outputDir string) bool {
	if util.IsK8sDeployment() {
		return true //nothing to do for k8s
	}
	rolesfile := outputDir + "/roles.yaml"
	util.PrintStepBanner("starting local processes")

	opts := []process.StartOp{}
	if processName == "" {
		// full start of all processes, do clean start
		opts = append(opts, process.WithCleanStartup())
	}

	for _, p := range util.Deployment.Vaults {
		opts = append(opts, process.WithRolesFile(rolesfile))
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Etcds {
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Influxs {
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Controllers {
		opts = append(opts, process.WithDebug("etcd,api,notify"))
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Crms {
		opts = append(opts, process.WithDebug("api,notify,mexos"))
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Dmes {
		opts = append(opts, process.WithRolesFile(rolesfile))
		opts = append(opts, process.WithDebug("locapi,dmedb,dmereq,notify"))
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.ClusterSvcs {
		opts = append(opts, process.WithDebug("notify,mexos"))
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Sqls {
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Mcs {
		opts = append(opts, process.WithRolesFile(rolesfile))
		opts = append(opts, process.WithDebug("api"))
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Locsims {
		if processName != "" && processName != p.Name {
			continue
		}
		if isLocalIP(p.Hostname) {
			log.Printf("Starting LocSim %+v\n", p)
			if p.Locfile != "" {
				stageLocDbFile(p.Locfile, "/var/tmp")
			}
			logfile := getLogFile(p.Name, outputDir)
			err := p.StartLocal(logfile)
			if err != nil {
				log.Printf("Error on LocSim startup: %v", err)
				return false
			}

		}
	}
	for _, p := range util.Deployment.Toksims {
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.SampleApps {
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	return true
}
