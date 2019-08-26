package cloudcommon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/mobiledgex/edge-cloud/vault"
)

const (
	BasicAuth  = "basic"
	TokenAuth  = "token"
	ApiKeyAuth = "apikey"
)

type RegistryAuth struct {
	AuthType string
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	ApiKey   string `json:"apikey"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
}

type RegistryTags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func getVaultRegistryPath(registry, vaultAddr string) string {
	return fmt.Sprintf(
		"%s/v1/secret/data/registry/%s",
		vaultAddr, registry,
	)
}

func GetRegistryAuth(imgUrl, vaultAddr string) (*RegistryAuth, error) {
	if vaultAddr == "" {
		return nil, fmt.Errorf("missing vaultAddr")
	}
	urlObj, err := util.ImagePathParse(imgUrl)
	if err != nil {
		return nil, err
	}
	hostname := strings.Split(urlObj.Host, ":")

	if len(hostname) < 1 {
		return nil, fmt.Errorf("empty hostname")
	}
	vaultPath := getVaultRegistryPath(hostname[0], vaultAddr)
	log.DebugLog(log.DebugLevelApi, "get registry auth", "vault-path", vaultPath)

	data, err := vault.GetVaultData(vaultPath)
	if err != nil {
		return nil, err
	}
	auth := &RegistryAuth{}
	err = mapstructure.WeakDecode(data["data"], auth)
	if err != nil {
		return nil, err
	}
	auth.Hostname = hostname[0]
	if len(hostname) > 1 {
		auth.Port = hostname[1]
	}
	if auth.Username != "" && auth.Password != "" {
		auth.AuthType = BasicAuth
	} else if auth.Token != "" {
		auth.AuthType = TokenAuth
	} else if auth.ApiKey != "" {
		auth.AuthType = ApiKeyAuth
	}
	return auth, nil
}

func SendHTTPReqAuth(method, regUrl string, auth *RegistryAuth) (*http.Response, error) {
	log.DebugLog(log.DebugLevelApi, "send http request", "method", method, "url", regUrl)
	client := &http.Client{
		Transport: &http.Transport{
			// Connection Timeout
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,

			// Response Header Timeout
			ExpectContinueTimeout: 5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
		},
		// Prevent endless redirects
		Timeout: 10 * time.Minute,
	}

	req, err := http.NewRequest(method, regUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed sending request %v", err)
	}

	if auth != nil {
		switch auth.AuthType {
		case BasicAuth:
			req.SetBasicAuth(auth.Username, auth.Password)
		case TokenAuth:
			req.Header.Set("Authorization", "Bearer "+auth.Token)
		case ApiKeyAuth:
			req.Header.Set("X-JFrog-Art-Api", auth.ApiKey)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getAuthToken(regUrl, authHeader string, auth *RegistryAuth) *RegistryAuth {
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
		resp, err := SendHTTPReqAuth("GET", authURL, auth)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			authTok := RegistryAuth{}
			json.NewDecoder(resp.Body).Decode(&authTok)
			authTok.AuthType = TokenAuth
			return &authTok
		}
	}
	return nil
}

func SendHTTPReq(method, regUrl string, vaultAddr string) (*http.Response, error) {
	resp, err := SendHTTPReqAuth(method, regUrl, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		auth, err := GetRegistryAuth(regUrl, vaultAddr)
		if err != nil {
			log.DebugLog(log.DebugLevelMexos, "warning, cannot get registry credentials from vault - assume public registry", "err", err)
		}
		resp, err = SendHTTPReqAuth(method, regUrl, auth)
		if err != nil {
			return nil, fmt.Errorf("Access denied to registry path")
		}
		if resp.StatusCode == http.StatusUnauthorized {
			// close respone body as we will retry with authtoken
			resp.Body.Close()
			authHeader := resp.Header.Get("Www-Authenticate")
			if authHeader != "" {
				// fetch authorization token to access tags
				authTok := getAuthToken(regUrl, authHeader, auth)
				if authTok == nil {
					return nil, fmt.Errorf("Access denied to registry path")
				}
				// retry with token
				resp, err = SendHTTPReqAuth(method, regUrl, authTok)
				if err != nil || resp.StatusCode != http.StatusOK {
					if resp != nil {
						resp.Body.Close()
					}
					return nil, fmt.Errorf("Access denied to registry path")
				}
			} else {
				fmt.Println(">>>>", resp, err)
				return nil, fmt.Errorf("Access denied to registry path")
			}
		}
	}
	return resp, nil
}

func ValidateDockerRegistryPath(regUrl, vaultAddr string) error {
	log.DebugLog(log.DebugLevelApi, "validate registry path", "path", regUrl)

	if regUrl == "" {
		return fmt.Errorf("registry path is empty")
	}

	version := "v2"
	matchTag := "latest"
	regPath := ""

	urlObj, err := util.ImagePathParse(regUrl)
	if err != nil {
		return err
	}
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

	resp, err := SendHTTPReq("GET", regUrl, vaultAddr)
	if err != nil {
		return fmt.Errorf("Invalid registry path: %s", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		tagsList := RegistryTags{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Failed to read response body, %v", err)
		}
		log.DebugLog(log.DebugLevelApi, "list tags", "resp", string(body))
		err = json.Unmarshal(body, &tagsList)
		if err != nil {
			return err
		}
		for _, tag := range tagsList.Tags {
			if tag == matchTag {
				return nil
			}
		}
		return fmt.Errorf("Invalid registry tag: %s does not exist", matchTag)
	}
	return fmt.Errorf("Invalid registry path (%s)", http.StatusText(resp.StatusCode))
}

func ValidateVMRegistryPath(imgUrl, vaultAddr string) error {
	log.DebugLog(log.DebugLevelApi, "validate vm-image path", "path", imgUrl)
	if imgUrl == "" {
		return fmt.Errorf("image path is empty")
	}

	resp, err := SendHTTPReq("GET", imgUrl, vaultAddr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("Invalid image path (%s)", http.StatusText(resp.StatusCode))
}
