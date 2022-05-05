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

package apis

// interacts with the controller APIs for use by the e2e test tool

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/edgexr/edge-cloud/setup-env/util"
)

// This starts and stops ShowAppInstClient apis
func RunAppInstClientAPI(api string, workerId string, apiFile string, outputDir string) bool {
	outputFile := outputDir + "/" + "show-appinstclients-" + workerId + ".yml"
	pidFile := "showAppInstClient" + workerId + ".pid"
	if api == "start" {
		err := ioutil.WriteFile(outputFile, []byte(util.LicenseTxt), 0644)
		if err != nil {
			log.Printf("Error - failed to write file: %s, %v\n", outputFile, err)
			return false
		}

		ctrl := util.GetController("")
		args := []string{"edgectl"}
		tlsFile := ctrl.GetTlsFile()
		if tlsFile != "" {
			args = append(args, "--tls", tlsFile)
		}
		if ctrl.ApiAddr != "" {
			args = append(args, "--addr", ctrl.ApiAddr)
		}
		args = append(args, "controller", "ShowAppInstClient", "--datafile", apiFile)
		// hide Location, since timestamps are always changing
		args = append(args, "--hidetags", "nocmp")
		args = append(args, ">>", outputFile)
		cmdStr := strings.Join(args, " ")
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		err = cmd.Start()
		if err != nil || cmd.Process == nil {
			log.Printf("Error - failed to start command: %v\n", cmd)
			return false
		}
		// record pid in a file
		pid := fmt.Sprintf("%d", cmd.Process.Pid)
		util.PrintToFile(pidFile, outputDir, pid, true)
	} else if api == "stop" {
		dat, err := ioutil.ReadFile(outputDir + "/" + pidFile)
		if err != nil {
			log.Printf("Failed to read file %s: %s\n", pidFile, err.Error())
			return false
		}
		pid, err := strconv.Atoi(string(dat))
		if err != nil {
			log.Printf("Failed to parse pid in file %s: %s\n", pidFile, err.Error())
			return false
		}
		err = syscall.Kill(-pid, syscall.SIGKILL)
		if err != nil {
			log.Printf("Failed to kill process %d: %s\n", pid, err.Error())
			return false
		}
	} else {
		log.Printf("Error: unsupported controller API %s\n", api)
		return false
	}
	return true
}
