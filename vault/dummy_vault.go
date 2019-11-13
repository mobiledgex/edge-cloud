package vault

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/hashicorp/vault/api"
)

// DummServer for unit testing responds to all requests with empty data.
func DummyServer() (*httptest.Server, *Config) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"data": {}}`)
	}))
	return server, NewConfig(server.URL, &NoAuth{})
}

// NoAuth skips any auth. It is used for unit testing against a fake httptest server.
type NoAuth struct{}

func (s *NoAuth) Login(client *api.Client) error {
	return nil
}

func (s *NoAuth) Type() string {
	return "none"
}
