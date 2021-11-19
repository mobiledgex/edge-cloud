package process

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Process interface {
	// Get the name of the process
	GetName() string
	// Get the hostname of the process
	GetHostname() string
	// Get EnvVars
	GetEnv() []string
	// Start the process
	StartLocal(logfile string, opts ...StartOp) error
	// Stop the process
	StopLocal()
	// Get the exe name of the process binary
	GetExeName() string
	// Get lookup args that can be used to find the local process using pgrep
	LookupArgs() string
}

type ProcessInfo struct {
	pid   int
	alive bool
}

type Common struct {
	Kind        string
	Name        string
	Hostname    string
	DockerImage string
	EnvVars     map[string]string
}

type HARole string

var HARoleAll HARole = "all"
var HARolePrimary HARole = "primary"
var HARoleSecondary HARole = "secondary"

var LocalRedisPort = "6379"
var LocalRedisAddr = "127.0.0.1:" + LocalRedisPort
var NoRedisAddr = ""

func (c *Common) GetName() string {
	return c.Name
}

func (c *Common) GetHostname() string {
	return c.Hostname
}

func (c *Common) GetEnv() []string {
	envs := []string{}
	for k, v := range c.EnvVars {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return envs
}

// Common args for all nodeMgr processes
type NodeCommon struct {
	TLS           TLSCerts
	VaultAddr     string
	UseVaultPki   bool
	DeploymentTag string
	AccessApiAddr string
	AccessKeyFile string
}

func (p *NodeCommon) GetNodeMgrArgs() []string {
	args := []string{}
	if p.VaultAddr != "" {
		args = append(args, "--vaultAddr", p.VaultAddr)
	}
	if p.UseVaultPki {
		args = append(args, "--useVaultPki")
	}
	if p.DeploymentTag != "" {
		args = append(args, "--deploymentTag", p.DeploymentTag)
	}
	if p.AccessApiAddr != "" {
		args = append(args, "--accessApiAddr", p.AccessApiAddr)
	}
	if p.AccessKeyFile != "" {
		args = append(args, "--accessKeyFile", p.AccessKeyFile)
	}
	return p.TLS.AddInternalPkiArgs(args)
}

// options

type StartOptions struct {
	Debug        string
	RolesFile    string
	CleanStartup bool
	ExtraArgs    []string
}

type StartOp func(op *StartOptions)

func WithDebug(debug string) StartOp {
	return func(op *StartOptions) { op.Debug = debug }
}

func WithRolesFile(rolesfile string) StartOp {
	return func(op *StartOptions) { op.RolesFile = rolesfile }
}

func WithCleanStartup() StartOp {
	return func(op *StartOptions) { op.CleanStartup = true }
}

func WithExtraArgs(params []string) StartOp {
	return func(op *StartOptions) { op.ExtraArgs = params }
}

func (s *StartOptions) ApplyStartOptions(opts ...StartOp) {
	for _, fn := range opts {
		fn(s)
	}
}

func GetTypeString(p interface{}) string {
	t := reflect.TypeOf(p)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
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
		pgrepCommand = "pgrep -f '" + processName + " .*" + processArgs + ".*'"
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

func StopProcess(p Process, maxwait time.Duration, c chan string) {
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
			//try to kill gracefully
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
