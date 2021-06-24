package cloudcommon

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
)

type SessionInfo struct {
	Token      string
	AccessPort string
}

type ExecReqInfo struct {
	Type    ExecReqType
	InitURL *url.URL
	Cookies []*http.Cookie
}

type ExecReqType int

const (
	ExecReqConsole ExecReqType = 0
	ExecReqShell   ExecReqType = 1
)

func GetFileNameWithExt(fileUrlPath string) (string, error) {
	log.DebugLog(log.DebugLevelInfra, "get file name with extension from url", "file-url", fileUrlPath)
	fileUrl, err := url.Parse(fileUrlPath)
	if err != nil {
		return "", fmt.Errorf("Error parsing file URL %s, %v", fileUrlPath, err)
	}

	_, file := filepath.Split(fileUrl.Path)
	return file, nil
}

func GetFileName(fileUrlPath string) (string, error) {
	log.DebugLog(log.DebugLevelInfra, "get file name from url", "file-url", fileUrlPath)
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
	if len(out) != 3 {
		return "", fmt.Errorf("invalid version details: %s", out)
	}
	return out[2], nil
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

// round the given field denoted by digIdx, we mostly want seconds
// rounded to two digits
func FormatDuration(dur time.Duration, digIdx int) string {

	var divisors = []time.Duration{
		time.Duration(1),
		time.Duration(10),
		time.Duration(100),
		time.Duration(1000),
	}

	if digIdx < 0 {
		digIdx = 0
	}
	if digIdx >= len(divisors) {
		digIdx = len(divisors) - 1
	}

	switch {
	case dur > time.Second:
		dur = dur.Round(time.Second / divisors[digIdx])
	case dur > time.Millisecond:
		dur = dur.Round(time.Millisecond / divisors[digIdx])
	case dur > time.Microsecond:
		dur = dur.Round(time.Microsecond / divisors[digIdx])
	}
	return dur.String()
}

func LookupDNS(name string) (string, error) {
	ips, err := net.LookupIP(name)
	if err != nil {
		return "", fmt.Errorf("DNS lookup error, %s, %v", name, err)
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no DNS records, %s", name)
	}
	for _, ip := range ips {
		return ip.String(), nil //XXX return only first one
	}
	return "", fmt.Errorf("no IP in DNS record for %s", name)
}

func Md5SumStr(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func Md5SumFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Failed to open file %s, %v", filePath, err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("Failed to calculate md5sum of file %s, %v", filePath, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func CreateInfluxMeasurementName(measurement string, interval time.Duration) string {
	return measurement + "-" + interval.String()
}

func DeleteFile(filePath string) error {
	if err := os.Remove(filePath); !os.IsNotExist(err) {
		return err
	}
	return nil
}
