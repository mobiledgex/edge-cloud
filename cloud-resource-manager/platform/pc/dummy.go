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
