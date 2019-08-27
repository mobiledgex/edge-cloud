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

func handleWWWAuth(method, regUrl, authHeader string, auth *RegistryAuth) (*http.Response, error) {
	log.DebugLog(
		log.DebugLevelApi, "handling www-auth for Docker Registry v2 Authentication",
		"regUrl", regUrl,
		"authHeader", authHeader,
	)
	authURL := ""
	if strings.HasPrefix(authHeader, "Bearer") {
		parts := strings.Split(strings.TrimPrefix(authHeader, "Bearer "), ",")

		m := map[string]string{}
		for _, part := range parts {
			if splits := strings.Split(part, "="); len(splits) == 2 {
				m[splits[0]] = strings.Replace(splits[1], "\"", "", 2)
			}
		}
		if _, ok := m["realm"]; !ok {
			return nil, fmt.Errorf("unable to find realm")
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
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			authTok := RegistryAuth{}
			json.NewDecoder(resp.Body).Decode(&authTok)
			authTok.AuthType = TokenAuth

			log.DebugLog(log.DebugLevelApi, "retrying request with auth-token")
			resp, err = SendHTTPReqAuth(method, regUrl, &authTok)
			if err != nil {
				return nil, err
			}
			if resp.StatusCode != http.StatusOK {
				if resp != nil {
					resp.Body.Close()
				}
				return nil, fmt.Errorf(http.StatusText(resp.StatusCode))
			}
			return resp, nil
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			if auth == nil {
				return nil, fmt.Errorf("Unable to find hostname in Vault")
			}
		}
		return nil, fmt.Errorf(http.StatusText(resp.StatusCode))
	}
	return nil, fmt.Errorf("unable to find bearer token")
}

/*
 * Sends HTTP request to regUrl
 * Checks if any Auth Credentials is needed by doing a lookup to Vault path
 *  - If it finds auth details, then HTTP request is sent with auth details set in HTTP Header
 *  - else, we assume it to be a public registry which requires no authentication
 * Following is the flow for Docker Registry v2 authentication:
 * - Send HTTP request to regUrl with auth (if found in Vault) or else without auth
 * - If the registry requires authorization, it will return a 401 Unauthorized response with a
 *   WWW-Authenticate header detailing how to authenticate to this registry
 * - We then make a request to the authorization service for a Bearer token
 * - The authorization service returns an opaque Bearer token representing the client’s authorized access
 * - Retry the original request with the Bearer token embedded in the request’s Authorization header
 * - The Registry authorizes the client by validating the Bearer token and the claim set embedded within
 *   it and begins the session as usual
 */
func SendHTTPReq(method, regUrl string, vaultAddr string) (*http.Response, error) {
	auth, err := GetRegistryAuth(regUrl, vaultAddr)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "warning, cannot get registry credentials from vault - assume public registry", "err", err)
	}
	resp, err := SendHTTPReqAuth(method, regUrl, auth)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		// Following is valid only for Docker Registry v2 Authentication
		// close response body as we will retry with authtoken
		resp.Body.Close()
		authHeader := resp.Header.Get("Www-Authenticate")
		if authHeader != "" {
			// fetch authorization token to access tags
			resp, err = handleWWWAuth(method, regUrl, authHeader, auth)
			if err == nil {
				return resp, nil
			}
			log.DebugLog(log.DebugLevelApi, "unable to handle www-auth", "err", err)
		} else {
			err = fmt.Errorf("Access denied to registry path")
		}
		return nil, err
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
		return err
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
	return fmt.Errorf("Invalid registry path: %s", http.StatusText(resp.StatusCode))
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
	return fmt.Errorf("Invalid image path: %s", http.StatusText(resp.StatusCode))
}
