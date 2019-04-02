package dind

import (
	"fmt"
	"math"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func getMacLimits(info *edgeproto.CloudletInfo) error {
	// get everything
	s, err := sh.Command("sysctl", "-a").Output()
	if err != nil {
		return err
	}
	sysout := string(s)

	rmem, _ := regexp.Compile("hw.memsize:\\s+(\\d+)")
	if rmem.MatchString(sysout) {
		matches := rmem.FindStringSubmatch(sysout)
		memoryB, err := strconv.Atoi(matches[1])
		if err != nil {
			return err
		}
		memoryMb := math.Round((float64(memoryB) / 1024 / 1024))
		info.OsMaxRam = uint64(memoryMb)
	}
	rcpu, _ := regexp.Compile("hw.ncpu:\\s+(\\d+)")
	if rcpu.MatchString(sysout) {
		matches := rcpu.FindStringSubmatch(sysout)
		cpus, err := strconv.Atoi(matches[1])
		if err != nil {
			return err
		}
		info.OsMaxVcores = uint64(cpus)
	}
	// hardcoding disk size for now, TODO: consider changing this but we need to consider that the
	// whole disk is not available for DIND.
	info.OsMaxVolGb = 500
	log.DebugLog(log.DebugLevelMexos, "getMacLimits results", "info", info)
	return nil
}

func getLinuxLimits(info *edgeproto.CloudletInfo) error {
	// get memory
	m, err := sh.Command("grep", "MemTotal", "/proc/meminfo").Output()
	memline := string(m)
	if err != nil {
		return err
	}
	rmem, _ := regexp.Compile("MemTotal:\\s+(\\d+)\\s+kB")

	if rmem.MatchString(string(memline)) {
		matches := rmem.FindStringSubmatch(memline)
		memoryKb, err := strconv.Atoi(matches[1])
		if err != nil {
			return err
		}
		memoryMb := math.Round((float64(memoryKb) / 1024))
		info.OsMaxRam = uint64(memoryMb)
	}
	c, err := sh.Command("grep", "-c", "processor", "/proc/cpuinfo").Output()
	cpuline := string(c)
	cpuline = strings.TrimSpace(cpuline)
	cpus, err := strconv.Atoi(cpuline)
	if err != nil {
		return err
	}
	info.OsMaxVcores = uint64(cpus)

	// disk space
	fd, err := sh.Command("fdisk", "-l").Output()
	fdstr := string(fd)
	rdisk, err := regexp.Compile("Disk\\s+\\S+:\\s+(\\d+)\\s+GiB")
	if err != nil {
		return err
	}
	matches := rdisk.FindStringSubmatch(fdstr)
	if matches != nil {
		//for now just looking for one disk
		diskGb, err := strconv.Atoi(matches[1])
		if err != nil {
			return err
		}
		info.OsMaxVolGb = uint64(diskGb)
	}
	log.DebugLog(log.DebugLevelMexos, "getLinuxLimits results", "info", info)
	return nil

}

// DINDGetLimits gets CPU, Memory from the local machine
func GetLimits(info *edgeproto.CloudletInfo) error {
	log.DebugLog(log.DebugLevelMexos, "DINDGetLimits called", "os", runtime.GOOS)
	switch runtime.GOOS {
	case "darwin":
		return getMacLimits(info)
	case "linux":
		return getLinuxLimits(info)
	}
	return fmt.Errorf("Unsupported OS %s for DIND", runtime.GOOS)
}
