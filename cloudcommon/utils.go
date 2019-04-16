package cloudcommon

import (
	"fmt"
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

	path := fileUrl.Path
	segments := strings.Split(path, "/")
	if len(segments) <= 0 {
		return "", fmt.Errorf("Error parsing file URL %s, path %s", fileUrlPath, path)
	}
	return segments[len(segments)-1], nil
}

func GetFileName(fileUrlPath string) (string, error) {
	log.DebugLog(log.DebugLevelMexos, "get file name from url", "file-url", fileUrlPath)
	fileName, err := GetFileNameWithExt(fileUrlPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(fileName, filepath.Ext(fileName)), nil
}
