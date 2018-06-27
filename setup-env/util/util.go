package util

import (
	"context"
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
	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
	"google.golang.org/grpc"
)

type yamlFileType int

const (
	YamlAppdata yamlFileType = 0
	YamlOther   yamlFileType = 1
)

type processinfo struct {
	pid   int
	alive bool
}

//get list of pids for a process name
func getPidsByName(processName string, processArgs string) []processinfo {
	//pidlist is a set of pids and alive bool
	var processes []processinfo
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
			pinfo := processinfo{pid: p, alive: true}
			processes = append(processes, pinfo)
		}
	}
	return processes
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
		log.Printf("Error in reading yaml file %v\n", firstYamlFile)
		return false
	}
	if err2 != nil {
		log.Printf("Error in reading yaml file %v\n", secondYamlFile)
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

func RunOperatorApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	opAPI := edgeproto.NewOperatorApiClient(conn)
	var err error = nil
	for _, o := range appdata.Operators {
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
	return err
}

func RunDeveloperApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	devApi := edgeproto.NewDeveloperApiClient(conn)
	for _, d := range appdata.Developers {
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
	return err
}

func RunCloudletApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	clAPI := edgeproto.NewCloudletApiClient(conn)
	for _, c := range appdata.Cloudlets {
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
	return err
}

func RunAppApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	appAPI := edgeproto.NewAppApiClient(conn)
	for _, a := range appdata.Applications {
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
	return err
}

func RunAppinstApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	appinAPI := edgeproto.NewAppInstApiClient(conn)
	for _, a := range appdata.AppInstances {
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
	return err
}
