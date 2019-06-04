package cloudcommon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
)

const RequestTimeout = 5 * time.Second

type TokenAuth struct {
	Token string `json:"token"`
}

type BasicAuth struct {
	Username string
	Password string
}

type RegistryTags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func MexRegistry(host string) bool {
	hostname := strings.Split(host, ":")
	if len(hostname) < 1 {
		return false
	}
	if strings.HasSuffix(hostname[0], ".mobiledgex.net") {
		return true
	}
	return false
}

func GetRegistryAuth() *BasicAuth {
	bAuth := BasicAuth{}
	bAuth.Username = os.Getenv("MEX_DOCKER_REG_USER")
	bAuth.Password = os.Getenv("MEX_DOCKER_REG_PASS")
	if bAuth.Username != "" && bAuth.Password != "" {
		return &bAuth
	}
	return nil
}

func SendHTTPReq(method, fileUrlPath string, auth interface{}) (*http.Response, error) {
	log.DebugLog(log.DebugLevelApi, "send http request", "method", method, "url", fileUrlPath)
	client := &http.Client{
		Timeout: RequestTimeout,
	}
	req, err := http.NewRequest(method, fileUrlPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed sending request %v", err)
	}
	if auth != nil {
		if basicAuth, ok := auth.(*BasicAuth); ok && basicAuth != nil {
			req.SetBasicAuth(basicAuth.Username, basicAuth.Password)
		}
		if tokAuth, ok := auth.(*TokenAuth); ok && tokAuth != nil {
			req.Header.Set("Authorization", "Bearer "+tokAuth.Token)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed fetching response %v", err)
	}
	// Check server response
	if resp.StatusCode != http.StatusOK {
		return resp, nil
	}
	return resp, nil
}

func ValidateRegistryPath(regUrl string) error {
	log.DebugLog(log.DebugLevelApi, "validate registry path", "path", regUrl)

	protocol := "https"
	version := "v2"
	matchTag := "latest"
	regPath := ""

	urlObj, err := url.Parse(protocol + "://" + regUrl)
	out := strings.Split(urlObj.Path, ":")
	if len(out) == 1 {
		regPath = urlObj.Path
	} else if len(out) == 2 {
		regPath = out[0]
		matchTag = out[1]
	} else {
		return fmt.Errorf("Invalid tag in registry path")
	}

	regUrl = fmt.Sprintf("%s://%s/%s%s/tags/list", urlObj.Scheme, urlObj.Host, version, regPath)
	log.DebugLog(log.DebugLevelApi, "registry api url", "url", regUrl)

	var basicAuth *BasicAuth
	if MexRegistry(urlObj.Host) {
		basicAuth = GetRegistryAuth()
	}
	resp, err := SendHTTPReq("GET", regUrl, basicAuth)
	if err != nil {
		return fmt.Errorf("Invalid registry path")
	}
	if resp.StatusCode == http.StatusUnauthorized {
		// close respone body as we will retry with authtoken
		resp.Body.Close()
		authHeader := resp.Header.Get("Www-Authenticate")
		if authHeader != "" {
			// fetch authorization token to access tags
			authTok := getAuthToken(regUrl, authHeader, basicAuth)
			if authTok == nil {
				return fmt.Errorf("Access denied to registry path")
			}
			// retry with token
			resp, err = SendHTTPReq("GET", regUrl, authTok)
			if err != nil || resp.StatusCode != http.StatusOK {
				if resp != nil {
					resp.Body.Close()
				}
				return fmt.Errorf("Access denied to registry path")
			}
		} else {
			return fmt.Errorf("Access denied to registry path")
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		tagsList := RegistryTags{}
		json.NewDecoder(resp.Body).Decode(&tagsList)
		for _, tag := range tagsList.Tags {
			if tag == matchTag {
				return nil
			}
		}
		return fmt.Errorf("Invalid registry tag: %s does not exist", matchTag)
	}
	return fmt.Errorf("Invalid registry path: %s", http.StatusText(resp.StatusCode))
}

func getAuthToken(regUrl, authHeader string, basicAuth *BasicAuth) *TokenAuth {
	log.DebugLog(log.DebugLevelApi, "get auth token", "regUrl", regUrl, "authHeader", authHeader)
	authURL := ""
	if strings.HasPrefix(authHeader, "Bearer") {
		parts := strings.Split(strings.Replace(authHeader, "Bearer ", "", 1), ",")

		m := map[string]string{}
		for _, part := range parts {
			if splits := strings.Split(part, "="); len(splits) == 2 {
				m[splits[0]] = strings.Replace(splits[1], "\"", "", 2)
			}
		}
		if _, ok := m["realm"]; !ok {
			return nil
		}

		authURL = m["realm"]
		if v, ok := m["service"]; ok {
			authURL += "?service=" + v
		}
		if v, ok := m["scope"]; ok {
			authURL += "&scope=" + v
		}
		resp, err := SendHTTPReq("GET", authURL, basicAuth)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			authTok := TokenAuth{}
			json.NewDecoder(resp.Body).Decode(&authTok)
			return &authTok
		}
	}
	return nil
}
