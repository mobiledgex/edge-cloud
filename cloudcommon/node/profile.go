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

package node

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
)

var cpuprofile *os.File
var cpufilelock sync.Mutex

func StartCpuProfile() string {
	cpufilelock.Lock()
	defer cpufilelock.Unlock()

	if cpuprofile != nil {
		return "cpu profiling already in progress"
	}
	tmpfile, err := ioutil.TempFile("", "cpuprofile")
	if err != nil {
		return err.Error()
	}
	err = pprof.StartCPUProfile(tmpfile)
	if err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return err.Error()
	}
	cpuprofile = tmpfile
	return "started. output will be base64 encoded go tool pprof file contents"
}

func StopCpuProfile() string {
	cpufilelock.Lock()
	defer cpufilelock.Unlock()

	if cpuprofile == nil {
		return "no cpu profiling in progress"
	}
	pprof.StopCPUProfile()
	cpuprofile.Close()

	dat, err := ioutil.ReadFile(cpuprofile.Name())
	os.Remove(cpuprofile.Name())
	cpuprofile = nil
	if err != nil {
		return err.Error()
	}
	return base64.StdEncoding.EncodeToString(dat)
}

func GetMemProfile() string {
	buf := &bytes.Buffer{}
	runtime.GC() // get up-to-date statistics
	err := pprof.WriteHeapProfile(buf)
	if err != nil {
		return err.Error()
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
