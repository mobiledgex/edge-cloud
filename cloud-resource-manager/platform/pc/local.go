package pc

import (
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/kr/pty"
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
func (s *LocalClient) Shell(sin io.Reader, sout, serr io.Writer, args ...string) error {
	if len(args) == 0 {
		args = []string{"/bin/sh"}
	}
	cmd := exec.Command(args[0], args[1:]...)
	tty, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer tty.Close()

	// wait until all data has been written to avoid
	// race conditions between write back and caller closing
	// the webrtc data channel.
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		io.Copy(sout, tty)
		wg.Done()
	}()
	go func() {
		io.Copy(tty, sin)
	}()
	cmd.Wait()
	wg.Wait()
	return nil
}

// Start starts the specified command without waiting for it to finish. You
// have to call the Wait function for that.
//
// The first two io.ReadCloser are the standard output and the standard
// error of the executing command respectively. The returned error follows
// the same logic as in the exec.Cmd.Start function.
func (s *LocalClient) Start(command string) (io.ReadCloser, io.ReadCloser, io.WriteCloser, error) {
	cmd := exec.Command("sh", "-c", command)
	sout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	serr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	sin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, nil, nil, err
	}
	s.cmd = cmd
	return sout, serr, sin, nil
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
