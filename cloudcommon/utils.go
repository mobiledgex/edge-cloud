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

func GetAvailablePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
