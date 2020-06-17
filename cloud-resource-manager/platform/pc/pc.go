package pc

import (
	"context"
	"encoding/base64"
	"fmt"
	"runtime"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

// Sudo is a toggle for executing as superuser
type Sudo bool

// SudoOn means run in sudo mode
var SudoOn Sudo = true

// NoSudo means dont run in sudo mode
var NoSudo Sudo = false

// OverwriteDir is a toggle to indicate overwrite during dir creation
type OverwriteDir bool

var Overwrite OverwriteDir = true
var NoOverwrite OverwriteDir = false

// Some utility functions

// WriteFile writes the file contents optionally in sudo mode
func WriteFile(client ssh.Client, file string, contents string, kind string, sudo Sudo) error {
	log.DebugLog(log.DebugLevelInfra, "write file", "kind", kind, "sudo", sudo)

	// encode to avoid issues with quotes, special characters, and shell
	// evaluation of $vars.
	dat := base64.StdEncoding.EncodeToString([]byte(contents))

	// On a mac base64 command "-d" option is "-D"
	// If we are running on a mac and we are trying to run base64 decode replace "-d" with "-D"
	decodeCmd := "base64 -d"
	if runtime.GOOS == "darwin" {
		if _, isLocalClient := client.(*LocalClient); isLocalClient {
			decodeCmd = "base64 -D"
		}
	}
	cmd := fmt.Sprintf("%s <<< %s > %s", decodeCmd, dat, file)
	if sudo {
		cmd = fmt.Sprintf("sudo bash -c '%s'", cmd)
	}
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error writing %s, %s, %s, %v", kind, cmd, out, err)
	}
	log.DebugLog(log.DebugLevelInfra, "wrote file", "kind", kind)
	return nil
}

func CreateDir(ctx context.Context, client ssh.Client, dir string, ow OverwriteDir) error {
	output, err := client.Output("mkdir " + dir)
	if err == nil {
		return nil
	}
	if !strings.Contains(output, "File exists") {
		log.SpanLog(ctx, log.DebugLevelInfra, "mkdir err", "out", output, "err", err)
		return err
	}

	if !ow {
		return nil
	}

	// If overwrite, then try deleting the directory and recreate it
	err = DeleteDir(ctx, client, dir, NoSudo)
	if err != nil {
		delerr := fmt.Errorf("unable to delete already existing directory: %v", err)
		log.SpanLog(ctx, log.DebugLevelInfra, "mkdir err", "err", delerr)
		return err
	}
	output, err = client.Output("mkdir " + dir)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "mkdir err", "out", output, "err", err)
		return err
	}
	return nil
}

func DeleteDir(ctx context.Context, client ssh.Client, dir string, sudo Sudo) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "deleting directory", "dir", dir)
	cmd := fmt.Sprintf("rm -rf %s", dir)
	if sudo == SudoOn {
		cmd = fmt.Sprintf("sudo rm -rf %s", dir)
	}
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deleting dir %s, %s, %v", cmd, out, err)
	}
	return nil
}

func DeleteFile(client ssh.Client, file string) error {
	log.DebugLog(log.DebugLevelInfra, "delete file")
	cmd := fmt.Sprintf("rm -f %s", file)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deleting  %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelInfra, "deleted file", "file", file)
	return nil
}

func CopyFile(client ssh.Client, src, dst string) error {
	cmd := fmt.Sprintf("cp %s %s", src, dst)
	out, err := client.Output(cmd)
	if err != nil {
		log.DebugLog(log.DebugLevelInfra, "copy failed", "src", src, "dst", dst, "err", err, "out", out)
	}
	return err
}

func Run(client ssh.Client, cmd string) error {
	out, err := client.Output(cmd)
	if err != nil {
		log.DebugLog(log.DebugLevelInfra, "cmd failed", "cmd", cmd, "err", err, "out", out)
		return fmt.Errorf("command \"%s\" failed, %v", cmd, err)
	}
	return nil
}
