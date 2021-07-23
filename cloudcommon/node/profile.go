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
