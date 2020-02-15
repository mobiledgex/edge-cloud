package pc

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"runtime"

	"github.com/mobiledgex/edge-cloud/log"
)

// PlatformClient provides a way to run commands on the
// underlying platform. It is currently implemented by
// golang-ssh's ssh.Client, and LocalClient.
type PlatformClient interface {
	// Output returns the output of the command.
	Output(command string) (string, error)
	// Shell requests a shell. If an arg is passed, it tries
	// to exec them on the remote shell.
	Shell(sin io.Reader, sout, serr io.Writer, args ...string) error
	// Start stars the specified command without waiting for
	// it to finish. Wait should be called to wait for it to finish.
	// The first two io.ReadCloser are the standard output and the standard
	// error of the executing command respectively. The returned error
	// follows the same logic as in the exec.Cmd.Start function.
	Start(command string) (io.ReadCloser, io.ReadCloser, io.WriteCloser, error)
	// Wait waits for the command started by the Start function.
	// The returned error follows the same logic as exec.Cmd.Wait.
	Wait() error
	// AddHpp adds a new host to the end of the list and returns a new client.
	// The original client is unchanged
	AddHop(host string, port int) (interface{}, error)
}

// Sudo is a toggle for executing as superuser
type Sudo bool

// SudoOn means run in sudo mode
var SudoOn Sudo = true

// NoSudo means dont run in sudo mode
var NoSudo Sudo = false

// Some utility functions

// WriteFile writes the file contents optionally in sudo mode
func WriteFile(client PlatformClient, file string, contents string, kind string, sudo Sudo) error {
	log.DebugLog(log.DebugLevelMexos, "write file", "kind", kind, "sudo", sudo)

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
	log.DebugLog(log.DebugLevelMexos, "wrote file", "kind", kind)
	return nil
}

func DeleteDir(ctx context.Context, client PlatformClient, dir string) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "deleting directory", "dir", dir)
	cmd := fmt.Sprintf("rm -rf %s", dir)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deleting dir %s, %s, %v", cmd, out, err)
	}
	return nil
}

func DeleteFile(client PlatformClient, file string) error {
	log.DebugLog(log.DebugLevelMexos, "delete file")
	cmd := fmt.Sprintf("rm -f %s", file)
	out, err := client.Output(cmd)
	if err != nil {
		return fmt.Errorf("error deleting  %s, %s, %v", cmd, out, err)
	}
	log.DebugLog(log.DebugLevelMexos, "deleted file", "file", file)
	return nil
}

func CopyFile(client PlatformClient, src, dst string) error {
	cmd := fmt.Sprintf("cp %s %s", src, dst)
	out, err := client.Output(cmd)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "copy failed", "src", src, "dst", dst, "err", err, "out", out)
	}
	return err
}

func Run(client PlatformClient, cmd string) error {
	out, err := client.Output(cmd)
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "cmd failed", "cmd", cmd, "err", err, "out", out)
		return fmt.Errorf("command \"%s\" failed, %v", cmd, err)
	}
	return nil
}
