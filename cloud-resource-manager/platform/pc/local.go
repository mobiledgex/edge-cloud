package pc

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Implements nanobox-io's ssh.Client interface, but runs commands locally.
// This is used for kubernetes DIND or other local testing.
type LocalClient struct {
	cmd *exec.Cmd
}

// Output returns the output of the command run on the remote host.
func (s *LocalClient) Output(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Shell requests a shell from the remote. If an arg is passed, it tries to
// exec them on the server.
func (s *LocalClient) Shell(args ...string) error {
	cmd := exec.Command("sh", "-c", strings.Join(args, " "))
	return cmd.Run()
}

// Start starts the specified command without waiting for it to finish. You
// have to call the Wait function for that.
//
// The first two io.ReadCloser are the standard output and the standard
// error of the executing command respectively. The returned error follows
// the same logic as in the exec.Cmd.Start function.
func (s *LocalClient) Start(command string) (io.ReadCloser, io.ReadCloser, error) {
	cmd := exec.Command("sh", "-c", command)
	sout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	errout, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, nil, err
	}
	s.cmd = cmd
	return sout, errout, nil
}

// Wait waits for the command started by the Start function to exit. The
// returned error follows the same logic as in the exec.Cmd.Wait function.
func (s *LocalClient) Wait() error {
	if s.cmd == nil {
		return fmt.Errorf("no command started")
	}
	err := s.cmd.Wait()
	s.cmd = nil
	return err
}
