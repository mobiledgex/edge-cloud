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

	"github.com/mobiledgex/edge-cloud/integration/process"

	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	yaml "gopkg.in/yaml.v2"
)

var (
	commandName = "setup-env"
	action      = flag.String("action", "", actionList)
	deployment  = flag.String("deployment", "", deploymentList)
	dataFile    = flag.String("datafile", "", "optional yml data file")
	processFile = flag.String("processfile", "", "mandatory yml application startup file")
	procs       processData
)

type applicationData struct {
	Operators    []edgeproto.Operator  `yaml:"operators"`
	Cloudlets    []edgeproto.Cloudlet  `yaml:"cloudlets"`
	Developers   []edgeproto.Developer `yaml:"developers"`
	Applications []edgeproto.App       `yaml:"apps"`
	AppInstances []edgeproto.AppInst   `yaml:"appinstances"`
}

type processData struct {
	Etcds       []process.EtcdLocal       `yaml:"etcds"`
	Controllers []process.ControllerLocal `yaml:"controllers"`
	Dmes        []process.DmeLocal        `yaml:"dmes"`
	Crms        []process.CrmLocal        `yaml:"crms"`
}

var actionChoices = map[string]bool{"start": true,
	"stop":    true,
	"status":  true,
	"cleanup": true}

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

func readApplicationData(datafile string) {
	yamlFile, err := ioutil.ReadFile(datafile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		os.Exit(1)
	}

	err = yaml.UnmarshalStrict(yamlFile, &data)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	yaml.Marshal(data)
	/*
		log.Printf("OPERATORS %+v", data.Operators)
		log.Printf("CLOUDLETS %+v", data.Cloudlets)
		log.Printf("DEVELOPERS %+v", data.Developers)
		log.Printf("APPLICATIONS %+v", data.Applications)
		log.Printf("APPINSTS %+v", data.AppInstances)
	*/
}

func applyApplicationData() {
	log.Printf("Applying data via APIs\n")
	for _, ctrl := range procs.Controllers {
		log.Printf("Connecting to controller %v", ctrl.Name)
		ctrlapi, err := ctrl.ConnectAPI(apiConnectTimeout)
		if err != nil {
			log.Printf("Error connecting to controller api: %v", ctrl)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), apiConnectTimeout)
			opAPI := edgeproto.NewOperatorApiClient(ctrlapi)
			for _, op := range data.Operators {
				log.Printf("Creating operator: %v", op.Key.Name)
				_, err = opAPI.CreateOperator(ctx, &op)
				if err != nil {
					log.Printf("Error creating operator: %v", err)
				}
			}
			devAPI := edgeproto.NewDeveloperApiClient(ctrlapi)
			for _, dev := range data.Developers {
				log.Printf("Creating  developer: %v", dev.Key.Name)
				_, err = devAPI.CreateDeveloper(ctx, &dev)
				if err != nil {
					log.Printf("Error creating developer: %v", err)
				}
			}
			clAPI := edgeproto.NewCloudletApiClient(ctrlapi)
			for _, cl := range data.Cloudlets {
				log.Printf("Creating  cloudlet: %v", cl.Key.Name)
				_, err = clAPI.CreateCloudlet(ctx, &cl)
				if err != nil {
					log.Printf("Error creating cloudlet: %v", err)
				}
			}
			appAPI := edgeproto.NewAppApiClient(ctrlapi)
			for _, app := range data.Applications {
				log.Printf("Creating  app: %v", app.Key.Name)
				_, err = appAPI.CreateApp(ctx, &app)
				if err != nil {
					log.Printf("Error creating app: %v", err)
				}
			}
			appinAPI := edgeproto.NewAppInstApiClient(ctrlapi)
			for _, appin := range data.AppInstances {
				log.Printf("Creating  app instance: %v", appin.Key.AppKey.Name)
				_, err = appinAPI.CreateAppInst(ctx, &appin)
				if err != nil {
					log.Printf("Error creating appinst: %v", err)
				}
			}
			cancel()

		}
		ctrlapi.Close()
	}
}

func getLogFile(procname string) string {
	//hardcoding to current dir for now, make this an option maybe
	return "./" + procname + ".log"
}

func readProcessData(datafile string) {
	yamlFile, err := ioutil.ReadFile(datafile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		os.Exit(1)
	}

	err = yaml.UnmarshalStrict(yamlFile, &procs)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	yaml.Marshal(procs)

	/*
		log.Printf("ETCDS %+v", procs.Etcds)
		log.Printf("CONTROLLERS %+v", procs.Controllers)
		log.Printf("CRMS %+v", procs.Crms)
		log.Printf("DMES %+v", procs.Dmes)
	*/
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

func startProcesses() bool {
	for _, etcd := range procs.Etcds {
		log.Printf("Starting Etcd %+v", etcd)
		etcd.ResetData()
		logfile := getLogFile(etcd.Name)
		err := etcd.Start(&logfile)
		if err != nil {
			log.Printf("Error on Etcd startup: %v", err)
			return false
		}

	}
	for _, ctrl := range procs.Controllers {
		log.Printf("Starting Controller %+v\n", ctrl)
		logfile := getLogFile(ctrl.Name)
		err := ctrl.Start(&logfile)
		if err != nil {
			log.Printf("Error on controller startup: %v", err)
			return false
		}
	}
	for _, crm := range procs.Crms {
		log.Printf("Starting CRM %+v\n", crm)
		logfile := getLogFile(crm.Name)
		err := crm.Start(&logfile)
		if err != nil {
			log.Printf("Error on CRM startup: %v", err)
			return false
		}
	}
	for _, dme := range procs.Dmes {
		log.Printf("Starting DME %+v\n", dme)
		logfile := getLogFile(dme.Name)
		err := dme.Start(&logfile)
		if err != nil {
			log.Printf("Error on DME startup: %v", err)
			return false
		}
	}
	return true
}

func validateArgs() {
	flag.Parse()
	_, validAction := actionChoices[*action]
	_, validDeployment := deploymentChoices[*deployment]

	errFound := false

	if !(validAction) {
		fmt.Printf("ERROR: --action must be one of: %v\n", actionList)
		errFound = true
	}
	if !(validDeployment) {
		fmt.Printf("ERROR: --deployment must be one of: %v\n", deploymentList)
		errFound = true
	}
	if *dataFile != "" {
		if _, err := os.Stat(*dataFile); err != nil {
			fmt.Printf("ERROR: file " + *dataFile + " does not exist\n")
			errFound = true
		}
	}
	if *processFile == "" {
		fmt.Printf("ERROR -processfile is mandatory\n")
		errFound = true
	} else {
		if _, err := os.Stat(*processFile); err != nil {
			fmt.Printf("ERROR: file " + *processFile + " does not exist\n")
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
		readApplicationData(*dataFile)
	}
	if *processFile != "" {
		readProcessData(*processFile)
	}
	switch *action {
	case "start":
		if startProcesses() {
			fmt.Println("Sleeping 10 seconds to allow processes to start")
			time.Sleep(10 * time.Second)
		} else {
			stopProcesses()
			log.Fatal("Startup failed, exiting")
			os.Exit(1)
		}
		applyApplicationData()
	case "stop":
		stopProcesses()
	case "cleanup":
		cleanup()
	default:
		log.Fatal("unexpected action: " + *action)
	}

	fmt.Println("Done")

}
