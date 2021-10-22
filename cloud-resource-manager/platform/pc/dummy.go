package pc

import (
	"io"
	"strings"
	"time"

	ssh "github.com/mobiledgex/golang-ssh"
)

// Dummy SSH Client for unit tests to mock SSH Client.
// Ignores input and delivers specified output and error.

type DummyClient struct {
	Out string
	Err error
}

func (s *DummyClient) Output(command string) (string, error) {
	return s.Out, s.Err
}

func (s *DummyClient) OutputWithTimeout(command string, timeout time.Duration) (string, error) {
	return s.Out, s.Err
}

func (s *DummyClient) Shell(sin io.Reader, sout, serr io.Writer, args ...string) error {
	return nil
}

func (s *DummyClient) Start(command string) (io.ReadCloser, io.ReadCloser, io.WriteCloser, error) {
	return NopReadCloser{strings.NewReader(s.Out)}, NopReadCloser{strings.NewReader("")}, NopWriteCloser{NullWriter{}}, s.Err
}

func (s *DummyClient) Wait() error {
	return nil
}

func (s *DummyClient) AddHop(host string, port int) (ssh.Client, error) {
	return s, nil
}

func (s *DummyClient) StartPersistentConn(timeout time.Duration) error {
	return nil
}

func (s *DummyClient) StopPersistentConn() {}

// For use with DummyClient Start() method
type NopReadCloser struct {
	io.Reader
}

func (NopReadCloser) Close() error { return nil }

type NopWriteCloser struct {
	io.Writer
}

func (NopWriteCloser) Close() error { return nil }

type NullWriter struct{}

func (NullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
