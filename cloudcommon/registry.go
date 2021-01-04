package cloudcommon

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/mobiledgex/edge-cloud/vault"
)

const (
	NoAuth            = "noauth"
	BasicAuth         = "basic"
	TokenAuth         = "token"
	ApiKeyAuth        = "apikey"
	DockerHub         = "docker.io"
	DockerHubRegistry = "registry-1.docker.io"
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

type RequestConfig struct {
	Timeout time.Duration
}

func getVaultRegistryPath(registry string) string {
	return fmt.Sprintf("/secret/data/registry/%s", registry)
}

type RegistryAuthApi interface {
	GetRegistryAuth(ctx context.Context, imgUrl string) (*RegistryAuth, error)
}

func GetRegistryAuth(ctx context.Context, imgUrl string, vaultConfig *vault.Config) (*RegistryAuth, error) {
	if vaultConfig == nil || vaultConfig.Addr == "" {
		return nil, fmt.Errorf("no vault specified")
	}
	urlObj, err := util.ImagePathParse(imgUrl)
	if err != nil {
		return nil, err
	}
	hostname := strings.Split(urlObj.Host, ":")

	if len(hostname) < 1 {
		return nil, fmt.Errorf("empty hostname")
	}
	vaultPath := getVaultRegistryPath(hostname[0])
	log.SpanLog(ctx, log.DebugLevelApi, "get registry auth", "vault-path", vaultPath)
	auth := &RegistryAuth{}
	err = vault.GetData(vaultConfig, vaultPath, 0, auth)
	if err != nil && strings.Contains(err.Error(), "no secrets") {
		// no secrets found, assume public registry
		log.SpanLog(ctx, log.DebugLevelApi, "warning, no registry credentials in vault, assume public registry", "err", err)
		auth.AuthType = NoAuth
		err = nil
	}
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

func SendHTTPReqAuth(ctx context.Context, method, regUrl string, auth *RegistryAuth, reqConfig *RequestConfig) (*http.Response, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "send http request", "method", method, "url", regUrl, "reqConfig", reqConfig)
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
			// use proxy if env vars set
			Proxy: http.ProxyFromEnvironment,
		},
		// Prevent endless redirects
		Timeout: 10 * time.Minute,
	}
	if reqConfig != nil && reqConfig.Timeout > 10*time.Minute {
		client.Timeout = reqConfig.Timeout
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

func handleWWWAuth(ctx context.Context, method, regUrl, authHeader string, auth *RegistryAuth) (*http.Response, error) {
	log.SpanLog(ctx,
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
		resp, err := SendHTTPReqAuth(ctx, "GET", authURL, auth, nil)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			authTok := RegistryAuth{}
			json.NewDecoder(resp.Body).Decode(&authTok)
			authTok.AuthType = TokenAuth

			log.SpanLog(ctx, log.DebugLevelApi, "retrying request with auth-token")
			resp, err = SendHTTPReqAuth(ctx, method, regUrl, &authTok, nil)
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
func SendHTTPReq(ctx context.Context, method, regUrl string, authApi RegistryAuthApi, reqConfig *RequestConfig) (*http.Response, error) {
	auth, err := authApi.GetRegistryAuth(ctx, regUrl)
	if err != nil {
		return nil, err
	}
	resp, err := SendHTTPReqAuth(ctx, method, regUrl, auth, reqConfig)
	if err != nil {
		return nil, err
	}
	// regUrl will be of this format: `https://..../imgName/tags/list`
	regUrlParts := strings.Split(regUrl, "/")
	imgName := regUrlParts[len(regUrlParts)-3]
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		// Following is valid only for Docker Registry v2 Authentication
		// close response body as we will retry with authtoken
		resp.Body.Close()
		authHeader := resp.Header.Get("Www-Authenticate")
		if authHeader != "" {
			// fetch authorization token to access tags
			resp, err = handleWWWAuth(ctx, method, regUrl, authHeader, auth)
			if err == nil {
				return resp, nil
			}
			log.SpanLog(ctx, log.DebugLevelApi, "unable to handle www-auth", "err", err)
			if err.Error() == http.StatusText(http.StatusNotFound) {
				return nil, fmt.Errorf("Please confirm that %s has been uploaded to registry", imgName)
			}
		}
		return nil, fmt.Errorf("Access denied to registry path")
	case http.StatusForbidden:
		resp.Body.Close()
		return nil, fmt.Errorf("Invalid credentials to access URL: %s", regUrl)
	case http.StatusOK:
		return resp, nil
	default:
		resp.Body.Close()
		return nil, fmt.Errorf("Invalid URL: %s, %s", regUrl, http.StatusText(resp.StatusCode))
	}
}

func ValidateDockerRegistryPath(ctx context.Context, regUrl string, authApi RegistryAuthApi) error {
	log.SpanLog(ctx, log.DebugLevelApi, "validate registry path", "path", regUrl)

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
	if urlObj.Host == DockerHub {
		// Even though public images are typically pulled from docker.io, the API v2 calls must be made to registry-1.docker.io
		urlObj.Host = DockerHubRegistry
		log.SpanLog(ctx, log.DebugLevelApi, "substituting docker hub registry for docker hub", "host", urlObj.Host)
	}
	regUrl = fmt.Sprintf("%s://%s/%s%s/tags/list", urlObj.Scheme, urlObj.Host, version, regPath)
	log.SpanLog(ctx, log.DebugLevelApi, "registry api url", "url", regUrl)

	resp, err := SendHTTPReq(ctx, "GET", regUrl, authApi, nil)
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
		log.SpanLog(ctx, log.DebugLevelApi, "list tags", "resp", string(body))
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

func ValidateVMRegistryPath(ctx context.Context, imgUrl string, authApi RegistryAuthApi) error {
	log.SpanLog(ctx, log.DebugLevelApi, "validate vm-image path", "path", imgUrl)
	if imgUrl == "" {
		return fmt.Errorf("image path is empty")
	}

	resp, err := SendHTTPReq(ctx, "GET", imgUrl, authApi, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("Invalid image path: %s", http.StatusText(resp.StatusCode))
}

type VaultRegistryAuthApi struct {
	VaultConfig *vault.Config
}

func (s *VaultRegistryAuthApi) GetRegistryAuth(ctx context.Context, imgUrl string) (*RegistryAuth, error) {
	return GetRegistryAuth(ctx, imgUrl, s.VaultConfig)
}

// For unit tests
type DummyRegistryAuthApi struct{}

func (s *DummyRegistryAuthApi) GetRegistryAuth(ctx context.Context, imgUrl string) (*RegistryAuth, error) {
	return &RegistryAuth{}, nil
}
