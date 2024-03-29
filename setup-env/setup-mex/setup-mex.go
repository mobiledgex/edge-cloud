// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package setupmex

// consists of utilities used to deploy, start, stop MEX processes either locally or remotely via Ansible.

import (
	"context"
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
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/common/xind"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/kind"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/setup-env/apis"
	"github.com/mobiledgex/edge-cloud/setup-env/util"

	uutil "github.com/mobiledgex/edge-cloud/util"
	yaml "gopkg.in/yaml.v2"
)

//root TLS Dir
var tlsDir = ""

//outout TLS cert dir
var tlsOutDir = ""

// some actions have sub arguments associated after equal sign,
// e.g.--actions stop=ctrl1
func GetActionParam(a string) (string, string) {
	argslice := strings.SplitN(a, "=", 2)
	action := argslice[0]
	actionParam := ""
	if len(argslice) > 1 {
		actionParam = argslice[1]
	}
	return action, actionParam
}

func GetCtrlNameFromCrmStartArgs(args []string) string {
	for ii := range args {
		act, param := GetActionParam(args[ii])
		if act == "ctrl" {
			return param
		}
	}
	return ""
}

func GetHARoleFromActionArgs(args []string) string {
	for ii := range args {
		act, param := GetActionParam(args[ii])
		if act == "harole" {
			return param
		}
	}
	return ""
}

// Change "cluster-svc1 scrapeInterval=30s updateAll" int []{"cluster-svc1", "scrapeInterval=30s", "updateApp"}
func GetActionArgs(a string) []string {
	argSlice := strings.Fields(a)
	return argSlice
}

// actions can be split with a dash like ctrlapi-show
func GetActionSubtype(a string) (string, string) {
	argslice := strings.SplitN(a, "-", 2)
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
	allpass := true
	for i := 0; i < count; i++ {
		rc := <-c
		log.Println(rc.Text)
		if !rc.Success {
			log.Printf("Error: connect failed: %s", rc.Text)
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
		if !process.EnsureProcessesByName(exeName, args) {
			log.Printf("Error: ensure process failed: %s", exeName)
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

	_, exist := vars["tlsoutdir"]
	if !exist {
		//{{tlsoutdir}} is relative to the GO dir.
		goPath := os.Getenv("GOPATH")
		if goPath == "" {
			fmt.Fprintf(os.Stderr, "GOPATH not set, cannot calculate tlsoutdir")
			return false
		}
		tlsDir = goPath + "/src/github.com/mobiledgex/edge-cloud/tls"
		tlsOutDir = tlsDir + "/out"
		vars["tlsoutdir"] = tlsOutDir
	}

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
			log.Printf("running %s clean clusterName: %s clusterID: %s\n", cloudcommon.DindScriptName, clusterName, clusterID)
			out, err := sh.Command(cloudcommon.DindScriptName, "clean").CombinedOutput()
			if err != nil {
				log.Printf("Error in dind clean: %v - %v\n", out, err)
				return fmt.Errorf("ERROR in cleanup Dind Cluster: %s", clusterName)
			}
			log.Printf("done dind clean for: %s out: %s\n", clusterName, out)
		}
	}
	log.Println("done CleanupDIND")
	return nil
}

func CleanupLocalProxies() error {
	// cleanup nginx and other docker containers common to both DIND and KIND
	pscmd := exec.Command("docker", "ps", "-a", "-q", "--filter", "label=edge-cloud")
	output, err := pscmd.Output()
	if err != nil {
		log.Printf("Error running docker ps: %v", err)
		return err
	}
	mexContainers := strings.Split(string(output), "\n")
	cmds := []string{"kill", "rm"}
	for _, c := range mexContainers {
		if c == "" {
			continue
		}
		for _, cmd := range cmds {
			killcmd := exec.Command("docker", cmd, c)
			output, err = killcmd.CombinedOutput()
			if err != nil {
				// not fatal as it may not have been running
				log.Printf("Error running command: %s on container: %s - %v - %v\n", cmd, c, string(output), err)
			}
		}
	}
	log.Println("done Cleanup local proxies")
	return nil
}

func CleanupKIND(ctx context.Context) error {
	log.Printf("Running CleanupKIND\n")
	vercmd := exec.Command("kind", "version")
	_, err := vercmd.CombinedOutput()
	if err != nil {
		// no kind installed
		log.Printf("No kind installed\n")
		return nil
	}
	client := &pc.LocalClient{
		WorkingDir: "/tmp",
	}

	clusters, err := kind.GetClusters(ctx, client)
	if err != nil {
		return err
	}
	for _, name := range clusters {
		log.Printf("pausing KIND cluster %s\n", name)
		nodes, err := kind.GetClusterContainerNames(ctx, client, name)
		if err != nil {
			log.Printf("Failed to get KIND cluster %s container names, %s", name, err)
			return err
		}
		err = xind.PauseContainers(ctx, client, nodes)
		if err != nil {
			log.Printf("Failed to pause KIND cluster %s, %s\n", name, err)
			return err
		}
	}
	log.Printf("done Cleanup KIND\n")
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
		go process.StopProcess(allprocs[ii], maxWait, c)
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
		for _, dn := range util.Deployment.DockerNetworks {
			log.Printf("Removing docker network %+v\n", dn)
			if err := dn.Delete(); err != nil {
				log.Printf("%s\n", err)
			}
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

func StartProcesses(processName string, args []string, outputDir string) bool {
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
	if len(args) > 0 {
		opts = append(opts, process.WithExtraArgs(args))
	}

	for _, dn := range util.Deployment.DockerNetworks {
		if processName != "" && dn.Name != processName {
			continue
		}
		if !IsLocalIP(dn.Hostname) {
			continue
		}
		if err := dn.Create(); err != nil {
			log.Printf("%s\n", err)
			return false
		}
	}
	for _, p := range util.Deployment.Influxs {
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
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
	for _, p := range util.Deployment.ElasticSearchs {
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Jaegers {
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Traefiks {
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.NginxProxys {
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.NotifyRoots {
		opts = append(opts, process.WithDebug("api,notify,events"))
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.EdgeTurns {
		opts = append(opts, process.WithRolesFile(rolesfile))
		opts = append(opts, process.WithDebug("api,notify"))
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Controllers {
		opts = append(opts, process.WithDebug("etcd,api,notify,metrics,infra,events"))
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Dmes {
		opts = append(opts, process.WithRolesFile(rolesfile))
		opts = append(opts, process.WithDebug("locapi,dmedb,dmereq,notify,metrics,events"))
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.ClusterSvcs {
		opts = append(opts, process.WithRolesFile(rolesfile))
		opts = append(opts, process.WithDebug("notify,infra,api,events"))
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	for _, p := range util.Deployment.Crms {
		opts = append(opts, process.WithDebug("notify,infra,api,events"))
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
	for _, p := range util.Deployment.RedisCaches {
		if !StartLocal(processName, outputDir, p, opts...) {
			return false
		}
	}
	return true
}

func Cleanup(ctx context.Context) error {
	err := cloudcommon.StopCRMService(ctx, nil, process.HARolePrimary)
	if err != nil {
		return err
	}
	err = CleanupKIND(ctx)
	if err != nil {
		return err
	}
	err = CleanupDIND()
	if err != nil {
		return err
	}
	err = process.CleanupEtcdRamDisk()
	if err != nil {
		return err
	}
	return CleanupLocalProxies()
}

func RunAction(ctx context.Context, actionSpec, outputDir string, spec *util.TestSpec, mods []string, vars map[string]string, retry *bool) []string {
	var actionArgs []string

	act, actionParam := GetActionParam(actionSpec)
	action, actionSubtype := GetActionSubtype(act)
	vars = uutil.AddMaps(vars, spec.ApiFileVars)

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
		if actionSubtype == "argument" {
			// extract the action param and action args
			actionArgs = GetActionArgs(actionParam)
			actionParam = actionArgs[0]
			actionArgs = actionArgs[1:]
		}
		if actionSubtype == "crm" {
			// extract the action param and action args
			actionArgs = GetActionArgs(actionParam)
			ctrlName := ""
			if len(actionArgs) > 0 {
				actionParam = actionArgs[0]
				actionArgs = actionArgs[1:]
				ctrlName = GetCtrlNameFromCrmStartArgs(actionArgs)
			}
			log.Printf("Starting CRM %s connected to ctrl %s\n", actionParam, ctrlName)
			// read the apifile and start crm with the details
			err := apis.StartCrmsLocal(ctx, actionParam, ctrlName, spec.ApiFile, spec.ApiFileVars, outputDir)
			if err != nil {
				errors = append(errors, err.Error())
			}
			break
		}
		if !StartProcesses(actionParam, actionArgs, outputDir) {
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
		if actionSubtype == "argument" {
			// extract the action param and action args
			actionArgs = GetActionArgs(actionParam)
			actionParam = actionArgs[0]
			actionArgs = actionArgs[1:]
		}
		if actionSubtype == "crm" || actionParam == "crm" {
			haRole := process.HARoleAll
			rolearg := GetHARoleFromActionArgs(actionArgs)
			if rolearg != "" {
				haRole = process.HARole(rolearg)
			}
			if err := apis.StopCrmsLocal(ctx, actionParam, spec.ApiFile, spec.ApiFileVars, haRole); err != nil {
				errors = append(errors, err.Error())
			}
		} else {
			allprocs := util.GetAllProcesses()
			if !StopProcesses(actionParam, allprocs) {
				errors = append(errors, "stop failed")
			}
		}
	case "ctrlapi":
		if !apis.RunControllerAPI(actionSubtype, actionParam, spec.ApiFile, spec.ApiFileVars, outputDir, mods, retry) {
			log.Printf("Unable to run api for %s-%s, %v\n", action, actionSubtype, mods)
			errors = append(errors, "controller api failed")
		}
	case "clientshow":
		if !apis.RunAppInstClientAPI(actionSubtype, actionParam, spec.ApiFile, outputDir) {
			log.Printf("Unable to run ShowAppInstClient api for %s, %v\n", action, mods)
			errors = append(errors, "ShowAppInstClient api failed")
		}
	case "exec":
		if !apis.RunCommandAPI(actionSubtype, actionParam, spec.ApiFile, spec.ApiFileVars, outputDir) {
			log.Printf("Unable to run RunCommand api for %s, %v\n", action, mods)
			errors = append(errors, "controller RunCommand api failed")
		}
	case "dmeapi":
		if !apis.RunDmeAPI(actionSubtype, actionParam, spec.ApiFile, spec.ApiFileVars, spec.ApiType, outputDir) {
			log.Printf("Unable to run api for %s\n", action)
			errors = append(errors, "dme api failed")
		}
	case "influxapi":
		if !apis.RunInfluxAPI(actionSubtype, actionParam, spec.ApiFile, spec.ApiFileVars, outputDir) {
			log.Printf("Unable to run influx api for %s\n", action)
			errors = append(errors, "influx api failed")
		}
	case "cmds":
		if !apis.RunCommands(spec.ApiFile, spec.ApiFileVars, outputDir, retry) {
			log.Printf("Unable to run commands for %s\n", action)
			errors = append(errors, "commands failed")
		}
	case "script":
		if !apis.RunScript(spec.ApiFile, outputDir, retry) {
			log.Printf("Unable to run script for %s\n", action)
			errors = append(errors, "script failed")
		}
	case "cleanup":
		err := Cleanup(ctx)
		if err != nil {
			errors = append(errors, err.Error())
		}
	case "sleep":
		t, err := strconv.ParseFloat(actionParam, 64)
		if err == nil {
			time.Sleep(time.Second * time.Duration(t))
		} else {
			errors = append(errors, fmt.Sprintf("Error in parsing sleeptime: %v", err))
		}
	default:
		errors = append(errors, fmt.Sprintf("invalid action %s", action))
	}
	return errors
}

type Retry struct {
	Enable    bool
	Count     int // number of retries (does not include first try)
	Interval  time.Duration
	Try       int
	runAction []bool
}

func NewRetry(count int, intervalSec float64, numActions int) *Retry {
	r := Retry{}
	r.Try = 1
	r.Count = count
	r.Interval = time.Duration(float64(time.Second) * intervalSec)
	r.runAction = make([]bool, numActions, numActions)
	if r.Count > 0 {
		r.Enable = true
	}
	if r.Enable && r.Interval == 0 {
		log.Fatal("Retry interval cannot be zero")
	}
	return &r
}

func (r *Retry) Tries() string {
	return fmt.Sprintf(" (try %d of %d)", r.Try, r.Try+r.Count)
}

func (r *Retry) SetActionRetry(ii int, retry bool) {
	// set whether or not to run the specific action on retries
	r.runAction[ii] = retry
	if !retry {
		return
	}
	// enable retries
	if r.Enable {
		return
	}
	r.Enable = true
	// set defaults
	r.Count = 5
	r.Interval = 500 * time.Millisecond
}

func (r *Retry) ShouldRunAction(ii int) bool {
	if r.Try == 1 {
		// always run actions the first iteration
		return true
	}
	return r.runAction[ii]
}

func (r *Retry) WillRetry() bool {
	return r.Count > 0
}

func (r *Retry) Done() bool {
	if r.Count == 0 {
		return true
	}
	r.Count--
	r.Try++
	time.Sleep(r.Interval)
	return false
}
