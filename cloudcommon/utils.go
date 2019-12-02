package cloudcommon

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"
)

func GetFileNameWithExt(fileUrlPath string) (string, error) {
	log.DebugLog(log.DebugLevelMexos, "get file name with extension from url", "file-url", fileUrlPath)
	fileUrl, err := url.Parse(fileUrlPath)
	if err != nil {
		return "", fmt.Errorf("Error parsing file URL %s, %v", fileUrlPath, err)
	}

	_, file := filepath.Split(fileUrl.Path)
	return file, nil
}

func GetFileName(fileUrlPath string) (string, error) {
	log.DebugLog(log.DebugLevelMexos, "get file name from url", "file-url", fileUrlPath)
	fileName, err := GetFileNameWithExt(fileUrlPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(fileName, filepath.Ext(fileName)), nil
}

func GetDockerBaseImageVersion() (string, error) {
	dat, err := ioutil.ReadFile("/version.txt")
	if err != nil {
		return "", err
	}
	out := strings.Fields(string(dat))
	if len(out) != 2 {
		return "", fmt.Errorf("invalid version details: %s", out)
	}
	return out[1], nil
}

func GetAvailablePort(ipaddr string) (string, error) {
	// Get non-conflicting port only if actual port is 0
	ipobj := strings.Split(ipaddr, ":")
	if len(ipobj) != 2 {
		return "", fmt.Errorf("invalid address format")
	}
	if ipobj[1] != "0" {
		return ipaddr, nil
	}
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", fmt.Errorf("unable to get TCP port: %v", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("unable to get TCP port: %v", err)
	}
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port

	return fmt.Sprintf("%s:%d", ipobj[0], port), nil
}

func MapParse(vars string) (map[string]string, error) {
	out := strings.Split(vars, "\n")
	if len(out) <= 1 {
		return nil, fmt.Errorf("invalid vars: %v", out)
	}
	varsmap := make(map[string]string)
	for _, v := range out {
		out1 := strings.Split(v, "=")
		if len(out1) != 2 {
			return nil, fmt.Errorf("invalid separator for key-value pair: %v", out1)
		}
		key := strings.TrimSpace(out1[0])
		value := strings.TrimSpace(out1[1])
		varsmap[key] = value
	}
	return varsmap, nil
}
