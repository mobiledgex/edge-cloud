package main

// exec-container is used for running k8s or docker exec commands within a developer container. Its purpose
// is to provide a wrapper for the CRM to run developer-specified commands as an additional layer of security

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"os/exec"
)

var (
	containerType *string
	userArgs      *string
	namespace     *string
	container     *string
	podName       *string
	debug         *bool
)

const (
	dockerContainerType string = "docker"
	k8sContainerType    string = "k8s"
)

func init() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	containerType = flag.String("containerType", "", "docker or k8s")
	namespace = flag.String("namespace", "default", "namespace for k8s")
	container = flag.String("containerName", "", "container name or id")
	podName = flag.String("podName", "", "pod name for k8s")
	debug = flag.Bool("debug", false, "debug logging")
	userArgs = flag.String("userArgs", "", "base64 encoded user provided args")
	flag.Parse()
}

func validateArgs() error {
	if *containerType == "docker" {
		if *container == "" {
			return fmt.Errorf("must specify --container for docker exec")
		}
	} else if *containerType == "k8s" {
		if *podName == "" {
			return fmt.Errorf("must specify --pod for k8s exec")
		}
	} else {
		return fmt.Errorf("must specify --containerType of \"docker\" or \"k8s\"")
	}
	if len(*userArgs) == 0 {
		return fmt.Errorf("must specify --userArgs")
	}
	return nil
}

func debugPrintf(format string, v ...interface{}) {
	if *debug {
		fmt.Printf(format, v...)
	}
}

func decodeBase64(str string) (string, error) {
	v, err := base64.StdEncoding.DecodeString(*userArgs)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func runKubectlCommand() {
	debugPrintf("runKubectlCommand args %s\n", *userArgs)
	decodedArgs, err := decodeBase64(*userArgs)
	if err != nil {
		fmt.Printf("error decoding base64 user args - %v", err)
		os.Exit(1)
	}
	cmdArgs := []string{"exec", "-n", *namespace, "-it", *podName}
	if *container != "" {
		cmdArgs = append(cmdArgs, "-c", *container)
	}
	userArgs := []string{"--", decodedArgs}
	cmdArgs = append(cmdArgs, userArgs...)
	cmd := exec.Command("kubectl", cmdArgs...)
	debugPrintf("command args %v\n", cmd)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runDockerCommand() {
	debugPrintf("runDockerCommand args %s\n", *userArgs)
	decodedArgs, err := decodeBase64(*userArgs)
	if err != nil {
		fmt.Printf("error decoding base64 user args - %v", err)
		os.Exit(1)
	}
	cmdArgs := []string{"exec", "-it", *container, decodedArgs}
	cmd := exec.Command("docker", cmdArgs...)
	debugPrintf("command args %v\n", cmd)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func main() {
	err := validateArgs()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if *containerType == k8sContainerType {
		runKubectlCommand()
	} else if *containerType == dockerContainerType {
		runDockerCommand()
	}
}
