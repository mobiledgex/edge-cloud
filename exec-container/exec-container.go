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

package main

// exec-container is used for running k8s or docker exec commands within a developer container. Its purpose
// is to provide a wrapper for the CRM to run developer-specified commands as an additional layer of security

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

var (
	containerType   *string
	encodedUserArgs *string
	namespace       *string
	container       *string
	podName         *string
	kubeConfig      *string
	debug           *bool
)

const (
	dockerContainerType string = "docker"
	k8sContainerType    string = "k8s"
)

func init() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	containerType = flag.String("containerType", "", "docker or k8s")
	namespace = flag.String("namespace", "default", "namespace for k8s")
	kubeConfig = flag.String("kubeconfig", "", "kubeconfig file for k8s")
	container = flag.String("container", "", "container name or id")
	podName = flag.String("podName", "", "pod name for k8s")
	debug = flag.Bool("debug", false, "debug logging")
	encodedUserArgs = flag.String("encodedUserArgs", "", "base64 encoded user provided args")
	flag.Parse()
}

func validateArgs() error {
	if *containerType == dockerContainerType {
		if *container == "" {
			return fmt.Errorf("must specify --container for docker exec")
		}
	} else if *containerType == k8sContainerType {
		if *podName == "" {
			return fmt.Errorf("must specify --podName for k8s exec")
		}
		if *kubeConfig == "" {
			return fmt.Errorf("must specify --kubeconfig for k8s exec")

		}
	} else {
		return fmt.Errorf("must specify --containerType of \"docker\" or \"k8s\"")
	}
	if *encodedUserArgs == "" {
		return fmt.Errorf("must specify --encodedUserArgs")
	}
	return nil
}

func debugLog(format string, v ...interface{}) {
	if *debug {
		log.Printf(format, v...)
	}
}

func decodeBase64(str string) (string, error) {
	v, err := base64.StdEncoding.DecodeString(*encodedUserArgs)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func runCommand(command string, cmdArgs []string) {
	cmd := exec.Command(command, cmdArgs...)
	debugLog("command args %v\n", cmd)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runKubectlCommand() {
	debugLog("runKubectlCommand args %s\n", *encodedUserArgs)
	decodedArgs, err := decodeBase64(*encodedUserArgs)
	if err != nil {
		log.Fatalf("error decoding base64 user args - %v", err)
		os.Exit(1)
	}
	cmdArgs := []string{"exec", "--kubeconfig=" + *kubeConfig, "-n", *namespace, "-it", *podName}
	if *container != "" {
		cmdArgs = append(cmdArgs, "-c", *container)
	}
	encodedUserArgs := []string{"--", decodedArgs}
	cmdArgs = append(cmdArgs, encodedUserArgs...)
	runCommand("kubectl", cmdArgs)
}

func runDockerCommand() {
	debugLog("runDockerCommand args %s\n", *encodedUserArgs)
	decodedArgs, err := decodeBase64(*encodedUserArgs)
	if err != nil {
		fmt.Printf("error decoding base64 user args - %v", err)
		os.Exit(1)
	}
	cmdArgs := []string{"exec", "-it", *container, decodedArgs}
	runCommand("docker", cmdArgs)
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
