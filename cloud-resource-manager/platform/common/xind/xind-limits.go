package xind

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

func getMacLimits(ctx context.Context, client ssh.Client, info *edgeproto.CloudletInfo) error {
	// get everything
	sysout, err := client.Output("sysctl -a")
	if err != nil {
		return err
	}

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
	log.DebugLog(log.DebugLevelInfra, "getMacLimits results", "info", info)
	return nil
}

func getLinuxLimits(ctx context.Context, client ssh.Client, info *edgeproto.CloudletInfo) error {
	// get memory
	memline, err := client.Output("grep MemTotal /proc/meminfo")
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
	cpuline, err := client.Output("grep -c processor /proc/cpuinfo")
	cpuline = strings.TrimSpace(cpuline)
	cpus, err := strconv.Atoi(cpuline)
	if err != nil {
		return err
	}
	info.OsMaxVcores = uint64(cpus)

	// disk space
	fdstr, err := client.Output("fdisk -l")
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
	log.SpanLog(ctx, log.DebugLevelInfra, "getLinuxLimits results", "info", info)
	return nil

}

// GetLimits gets CPU, Memory from the local machine
func GetLimits(ctx context.Context, client ssh.Client, info *edgeproto.CloudletInfo) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetLimits called", "os", runtime.GOOS)
	switch runtime.GOOS {
	case "darwin":
		return getMacLimits(ctx, client, info)
	case "linux":
		return getLinuxLimits(ctx, client, info)
	}
	return fmt.Errorf("Unsupported OS %s for XIND", runtime.GOOS)
}
