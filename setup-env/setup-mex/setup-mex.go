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
	"strconv"
	"strings"
	"time"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/setup-env/apis"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	yaml "gopkg.in/yaml.v2"
)

type TestSpec struct {
	ApiType     string      `json:"api" yaml:"api"`
	ApiFile     string      `json:"apifile" yaml:"apifile"`
	Actions     []string    `json:"actions" yaml:"actions"`
	CompareYaml CompareYaml `json:"compareyaml" yaml:"compareyaml"`
}

type CompareYaml struct {
	Yaml1    string `json:"yaml1" yaml:"yaml1"`
	Yaml2    string `json:"yaml2" yaml:"yaml2"`
	FileType string `json:"filetype" yaml:"filetype"`
}

//root TLS Dir
var tlsDir = ""

//outout TLS cert dir
var tlsOutDir = ""

// some actions have sub arguments associated after equal sign,
// e.g.--actions stop=ctrl1
func GetActionParam(a string) (string, string) {
	argslice := strings.Split(a, "=")
	action := argslice[0]
	actionParam := ""
	if len(argslice) > 1 {
		actionParam = argslice[1]
	}
	return action, actionParam
}

// actions can be split with a dash like ctrlapi-show
func GetActionSubtype(a string) (string, string) {
	argslice := strings.Split(a, "-")
	action := argslice[0]
	actionSubtype := ""
	if len(argslice) > 1 {
		actionSubtype = argslice[1]
	}
	return action, actionSubtype
}

func IsLocalIP(hostname string) bool {
	return hostname == "localhost" || hostname == "127.0.0.1"
}

func WaitForProcesses(processName string, procs []process.Process) bool {
	if !ensureProcesses(processName, procs) {
		return false
	}
	log.Println("Wait for processes to respond to APIs")
	c := make(chan util.ReturnCodeWithText)
	count := 0
	for _, ctrl := range util.Deployment.Controllers {
		if processName != "" && processName != ctrl.Name {
			continue
		}
		count++
		go util.ConnectController(ctrl, c)
	}
	for _, dme := range util.Deployment.Dmes {
		if processName != "" && processName != dme.Name {
			continue
		}
		count++
		go util.ConnectDme(dme, c)
	}
	for _, crm := range util.Deployment.Crms {
		if processName != "" && processName != crm.Name && processName != "allcrms" {
			continue
		}
		count++
		go util.ConnectCrm(crm, c)
	}
	allpass := true
	for i := 0; i < count; i++ {
		rc := <-c
		log.Println(rc.Text)
		if !rc.Success {
			allpass = false
		}
	}
	return allpass
}

// This uses the same methods as kill processes to look for local processes,
// to ensure that the lookup method for finding local processes is valid.
func ensureProcesses(processName string, procs []process.Process) bool {
	log.Println("Check processes are running")
	ensured := true
	for _, p := range procs {
		if processName != "" && processName != p.GetName() {
			continue
		}
		if !IsLocalIP(p.GetHostname()) {
			continue
		}

		exeName := p.GetExeName()
		args := p.LookupArgs()
		log.Printf("Looking for host %v processexe %v processargs %v\n", p.GetHostname(), exeName, args)
		if !util.EnsureProcessesByName(exeName, args) {
			ensured = false
		}
	}
	return ensured
}

func getLogFile(procname string, outputDir string) string {
	if outputDir == "" {
		return "./" + procname + ".log"
	} else {
		return outputDir + "/" + procname + ".log"
	}
}

func ReadSetupFile(setupfile string, deployment interface{}, vars map[string]string) bool {
	//the setup file has a vars section with replacement variables.  ingest the file once
	//to get these variables, and then ingest again to parse the setup data with the variables
	var setupVars util.SetupVariables

	//{{tlsoutdir}} is relative to the GO dir.
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		fmt.Fprintf(os.Stderr, "GOPATH not set, cannot calculate tlsoutdir")
		return false
	}
	tlsDir = goPath + "/src/github.com/mobiledgex/edge-cloud/tls"
	tlsOutDir = tlsDir + "/out"
	vars["tlsoutdir"] = tlsOutDir

	setupdir := filepath.Dir(setupfile)
	vars["setupdir"] = setupdir

	util.ReadYamlFile(setupfile, &setupVars)

	for _, repl := range setupVars.Vars {
		for varname, value := range repl {
			vars[varname] = value
		}
	}
	files := []string{setupfile}
	files = append(files, setupVars.Includes...)

	for _, filename := range files {
		err := util.ReadYamlFile(filename, deployment,
			util.WithVars(vars),
			util.ValidateReplacedVars())
		if err != nil {
			if !util.IsYamlOk(err, "setup") {
				fmt.Fprintf(os.Stderr, "One or more fatal unmarshal errors in %s", setupfile)
				return false
			}
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
	// cleanup nginxL7.
	pscmd := exec.Command("docker", "ps", "-a", "-q", "--filter", "ancestor=nginx")
	output, err = pscmd.Output()
	if err != nil {
		log.Printf("Error running docker ps: %v", err)
		return err
	}
	nginxContainers := strings.Split(string(output), "\n")
	nginxCmds := []string{"kill", "rm"}
	for _, nginx := range nginxContainers {
		if nginx == "" {
			continue
		}
		for _, cmd := range nginxCmds {
			killcmd := exec.Command("docker", cmd, nginx)
			output, err = killcmd.CombinedOutput()
			if err != nil {
				// not fatal as it may not have been running
				log.Printf("Error running command: %s on nginx container: %s - %v - %v\n", cmd, nginx, string(output), err)
			}
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
		if !IsLocalIP(p.GetHostname()) {
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

func StageYamlFile(filename string, directory string, contents interface{}) bool {

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

func StageLocDbFile(srcFile string, destDir string) {
	var locdb interface{}
	yerr := util.ReadYamlFile(srcFile, &locdb)
	if yerr != nil {
		fmt.Fprintf(os.Stderr, "Error reading location file %s -- %v\n", srcFile, yerr)
	}
	if !StageYamlFile("locsim.yml", destDir, locdb) {
		fmt.Fprintf(os.Stderr, "Error staging location db file %s to %s\n", srcFile, destDir)
	}
}

// CleanupTLSCerts . Deletes certs for a CN
func CleanupTLSCerts() error {
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

func StartLocal(processName, outputDir string, p process.Process, opts ...process.StartOp) bool {
	if processName != "" && processName != p.GetName() {
		return true
	}
	if !IsLocalIP(p.GetHostname()) {
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
	if outputDir == "" {
		outputDir = "."
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
	for _, p := range util.Deployment.Locsims {
		if processName != "" && processName != p.Name {
			continue
		}
		if IsLocalIP(p.Hostname) {
			log.Printf("Starting LocSim %+v\n", p)
			if p.Locfile != "" {
				StageLocDbFile(p.Locfile, "/var/tmp")
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

func Cleanup() error {
	return CleanupDIND()
}

func RunAction(actionSpec, outputDir string, spec *TestSpec, mods []string) []string {
	act, actionParam := GetActionParam(actionSpec)
	action, actionSubtype := GetActionSubtype(act)

	errors := []string{}
	switch action {
	case "gencerts":
		err := GenerateTLSCerts()
		if err != nil {
			errors = append(errors, err.Error())
		}
	case "cleancerts":
		err := CleanupTLSCerts()
		if err != nil {
			errors = append(errors, err.Error())
		}
	case "start":
		startFailed := false
		allprocs := util.GetAllProcesses()
		if !StartProcesses(actionParam, outputDir) {
			startFailed = true
			errors = append(errors, "start failed")
		}
		if startFailed {
			if !StopProcesses(actionParam, allprocs) {
				errors = append(errors, "stop failed")
			}
			break

		}
		if !WaitForProcesses(actionParam, allprocs) {
			errors = append(errors, "wait for process failed")
		}
	case "status":
		if !WaitForProcesses(actionParam, util.GetAllProcesses()) {
			errors = append(errors, "wait for process failed")
		}
	case "stop":
		allprocs := util.GetAllProcesses()
		if !StopProcesses(actionParam, allprocs) {
			errors = append(errors, "stop failed")
		}
	case "ctrlapi":
		if !apis.RunControllerAPI(actionSubtype, actionParam, spec.ApiFile, outputDir, mods) {
			log.Printf("Unable to run api for %s, %v\n", action, mods)
			errors = append(errors, "controller api failed")
		}
	case "dmeapi":
		if !apis.RunDmeAPI(actionSubtype, actionParam, spec.ApiFile, spec.ApiType, outputDir) {
			log.Printf("Unable to run api for %s\n", action)
			errors = append(errors, "dme api failed")
		}
	case "cleanup":
		err := Cleanup()
		if err != nil {
			errors = append(errors, err.Error())
		}
	case "sleep":
		t, err := strconv.ParseUint(actionParam, 10, 32)
		if err == nil {
			time.Sleep(time.Second * time.Duration(t))
		} else {
			errors = append(errors, "Error in parsing sleeptime")
		}
	default:
		errors = append(errors, fmt.Sprintf("invalid action %s", action))
	}
	return errors
}
