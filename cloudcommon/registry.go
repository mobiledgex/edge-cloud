package cloudcommon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

const RequestTimeout = 5 * time.Second

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

func GetRegistryAuth(registry, vaultAddr string) *RegistryAuth {
	hostname := strings.Split(registry, ":")

	if len(hostname) < 1 {
		return nil
	}
	vaultPath := getVaultRegistryPath(hostname[0], vaultAddr)
	log.DebugLog(log.DebugLevelApi, "get registry auth", "vault-path", vaultPath)

	data, err := vault.GetVaultData(vaultPath)
	if err != nil {
		return nil
	}
	auth := &RegistryAuth{}
	err = mapstructure.WeakDecode(data["data"], auth)
	if err != nil {
		return nil
	}
	if auth.Username != "" && auth.Password != "" {
		auth.AuthType = BasicAuth
		return auth
	}
	if auth.Token != "" {
		auth.AuthType = TokenAuth
		return auth
	}
	if auth.ApiKey != "" {
		auth.AuthType = ApiKeyAuth
		return auth
	}
	return nil
}

func SendHTTPReq(method, fileUrlPath string, auth *RegistryAuth) (*http.Response, error) {
	log.DebugLog(log.DebugLevelApi, "send http request", "method", method, "url", fileUrlPath)
	client := &http.Client{
		Timeout: RequestTimeout,
	}
	req, err := http.NewRequest(method, fileUrlPath, nil)
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
		return nil, fmt.Errorf("failed fetching response %v", err)
	}
	return resp, nil
}

func ValidateDockerRegistryPath(regUrl, vaultAddr string) error {
	log.DebugLog(log.DebugLevelApi, "validate registry path", "path", regUrl)

	if regUrl == "" {
		return fmt.Errorf("registry path is empty")
	}

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

	auth := GetRegistryAuth(urlObj.Host, vaultAddr)

	resp, err := SendHTTPReq("GET", regUrl, auth)
	if err != nil {
		return fmt.Errorf("Invalid registry path")
	}
	if resp.StatusCode == http.StatusUnauthorized {
		// close respone body as we will retry with authtoken
		resp.Body.Close()
		authHeader := resp.Header.Get("Www-Authenticate")
		if authHeader != "" {
			// fetch authorization token to access tags
			authTok := getAuthToken(regUrl, authHeader, auth)
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
		resp, err := SendHTTPReq("GET", authURL, auth)
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

func ValidateVMRegistryPath(imgUrl, vaultAddr string) error {
	log.DebugLog(log.DebugLevelApi, "validate vm-image path", "path", imgUrl)
	if imgUrl == "" {
		return fmt.Errorf("image path is empty")
	}

	urlObj, err := url.Parse(imgUrl)
	if err != nil {
		return nil
	}
	auth := GetRegistryAuth(urlObj.Host, vaultAddr)

	resp, err := SendHTTPReq("GET", imgUrl, auth)
	if err != nil {
		return fmt.Errorf("Invalid image path")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("Invalid image path: %s", http.StatusText(resp.StatusCode))
}
