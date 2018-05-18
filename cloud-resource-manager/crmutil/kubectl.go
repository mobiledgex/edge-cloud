package crmutil

import (
	"bufio"
	"os/exec"

	"github.com/bobbae/q"
)

func runKubeCtlCreateDeployment(fn string) error {
	cmdArgs := []string{"create", "-f", fn}

	return runCmd("kubectl", cmdArgs)
}

func runKubeCtlDeleteDeployment(fn string) error {
	cmdArgs := []string{"delete", "-f", fn}

	return runCmd("kubectl", cmdArgs)
}

func runCmd(cmdName string, cmdArgs []string) error {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			q.Q(cmdName, scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
