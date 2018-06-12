package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var (
	commandName = "simapp"
	port        = flag.Int("port", 0, "listen port")
	action      = flag.String("action", "", "[start stop]")
	appname     = flag.String("name", "", "App Name")
)

func sayAlive(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = *appname + " is Alive\n" + message
	w.Write([]byte(message))
}

func printUsage() {
	fmt.Println("\nUsage: \n" + commandName + " [options]\n\noptions:")
	flag.PrintDefaults()
}

func startApp() {

	path := fmt.Sprintf("./simapp -name %s -port %d -action run", *appname, *port)
	fmt.Println("starting: " + path)
	cmd := exec.Command("sh", "-c", path)
	err := cmd.Start()
	if err != nil {
		log.Printf("Error in command start %v", err)
	} else {
		log.Println("App started successfully")
		log.Printf("Test using: curl http://127.0.0.1:%d%s%s", *port, "/apps/", *appname)
	}

}

func run() {
	log.Printf("running..\n")
	path := fmt.Sprintf("/apps/%s", *appname)
	http.HandleFunc(path, sayAlive)

	portstr := fmt.Sprintf(":%d", *port)

	log.Printf("Listening on http://127.0.0.1:%d%s", *port, path)
	if err := http.ListenAndServe(portstr, nil); err != nil {
		panic(err)
	}
}

func stopApp() {
	exec.Command("sh", "-c", "pkill -SIGINT simapp").Output()
}

func validateArgs() {
	flag.Parse()
	if *appname == "" {
		log.Println("ERROR: need -name")
		printUsage()
		os.Exit(1)
	}

	if *port == 0 {
		log.Println("ERROR: need -port")
		printUsage()
		os.Exit(1)
	}
	if !((*action == "start") || (*action == "stop") || (*action == "run")) {
		log.Println("ERROR: action must be specified as [start stop run] ")
		printUsage()
		os.Exit(1)
	}

}

func main() {
	validateArgs()

	if *action == "start" {
		startApp()
	}
	if *action == "stop" {
		stopApp()
	}
	if *action == "run" {
		run()
	}

}
